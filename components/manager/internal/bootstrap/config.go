package bootstrap

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	in2 "github.com/LerianStudio/reporter/v3/components/manager/internal/adapters/http/in"
	"github.com/LerianStudio/reporter/v3/components/manager/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/v3/components/manager/internal/adapters/redis"
	"github.com/LerianStudio/reporter/v3/components/manager/internal/services"
	"github.com/LerianStudio/reporter/v3/pkg"
	"github.com/LerianStudio/reporter/v3/pkg/constant"
	"github.com/LerianStudio/reporter/v3/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/v3/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/v3/pkg/pdf"
	simpleClient "github.com/LerianStudio/reporter/v3/pkg/seaweedfs"
	reportSeaweedFS "github.com/LerianStudio/reporter/v3/pkg/seaweedfs/report"
	templateSeaweedFS "github.com/LerianStudio/reporter/v3/pkg/seaweedfs/template"

	"github.com/LerianStudio/lib-auth/v2/auth/middleware"
	mongoDB "github.com/LerianStudio/lib-commons/v2/commons/mongo"
	libOtel "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	libRabbitmq "github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
	libRedis "github.com/LerianStudio/lib-commons/v2/commons/redis"
	"github.com/LerianStudio/lib-commons/v2/commons/zap"
	libLicense "github.com/LerianStudio/lib-license-go/v2/middleware"
)

// Config is the top-level configuration struct for the entire application.
type Config struct {
	// Service envs
	EnvName       string `env:"ENV_NAME"`
	ServerAddress string `env:"SERVER_ADDRESS"`
	LogLevel      string `env:"LOG_LEVEL"`
	// Otel and telemetry configuration envs
	OtelServiceName         string `env:"OTEL_RESOURCE_SERVICE_NAME"`
	OtelLibraryName         string `env:"OTEL_LIBRARY_NAME"`
	OtelServiceVersion      string `env:"OTEL_RESOURCE_SERVICE_VERSION"`
	OtelDeploymentEnv       string `env:"OTEL_RESOURCE_DEPLOYMENT_ENVIRONMENT"`
	OtelColExporterEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	EnableTelemetry         bool   `env:"ENABLE_TELEMETRY"`
	// Mongo configuration envs
	MongoURI        string `env:"MONGO_URI"`
	MongoDBHost     string `env:"MONGO_HOST"`
	MongoDBName     string `env:"MONGO_NAME"`
	MongoDBUser     string `env:"MONGO_USER"`
	MongoDBPassword string `env:"MONGO_PASSWORD"`
	MongoDBPort     string `env:"MONGO_PORT"`
	// SeaweedFS configuration envs
	SeaweedFSHost      string `env:"SEAWEEDFS_HOST"`
	SeaweedFSFilerPort string `env:"SEAWEEDFS_FILER_PORT"`
	// RabbitMQ configuration envs
	RabbitURI                   string `env:"RABBITMQ_URI"`
	RabbitMQHost                string `env:"RABBITMQ_HOST"`
	RabbitMQHealthCheckURL      string `env:"RABBITMQ_HEALTH_CHECK_URL"`
	RabbitMQPortHost            string `env:"RABBITMQ_PORT_HOST"`
	RabbitMQPortAMQP            string `env:"RABBITMQ_PORT_AMQP"`
	RabbitMQUser                string `env:"RABBITMQ_DEFAULT_USER"`
	RabbitMQPass                string `env:"RABBITMQ_DEFAULT_PASS"`
	RabbitMQGenerateReportQueue string `env:"RABBITMQ_GENERATE_REPORT_QUEUE"`
	// Redis/Valkey configuration envs
	RedisHost                    string `env:"REDIS_HOST"`
	RedisMasterName              string `env:"REDIS_MASTER_NAME" default:""`
	RedisPassword                string `env:"REDIS_PASSWORD"`
	RedisDB                      int    `env:"REDIS_DB" default:"0"`
	RedisProtocol                int    `env:"REDIS_PROTOCOL" default:"3"`
	RedisTLS                     bool   `env:"REDIS_TLS" default:"false"`
	RedisCACert                  string `env:"REDIS_CA_CERT"`
	RedisUseGCPIAM               bool   `env:"REDIS_USE_GCP_IAM" default:"false"`
	RedisServiceAccount          string `env:"REDIS_SERVICE_ACCOUNT" default:""`
	GoogleApplicationCredentials string `env:"GOOGLE_APPLICATION_CREDENTIALS" default:""`
	RedisTokenLifeTime           int    `env:"REDIS_TOKEN_LIFETIME" default:"60"`
	RedisTokenRefreshDuration    int    `env:"REDIS_TOKEN_REFRESH_DURATION" default:"45"`
	// Auth envs
	AuthAddress string `env:"PLUGIN_AUTH_ADDRESS"`
	AuthEnabled bool   `env:"PLUGIN_AUTH_ENABLED"`
	// License configuration envs
	LicenseKey      string `env:"LICENSE_KEY"`
	OrganizationIDs string `env:"ORGANIZATION_IDS"`
	// PDF Pool configuration envs
	PdfPoolWorkers        int `env:"PDF_POOL_WORKERS" default:"5"`
	PdfPoolTimeoutSeconds int `env:"PDF_TIMEOUT_SECONDS" default:"30"`
}

// InitServers initiate http and grpc servers.
func InitServers() *Service {
	cfg := &Config{}
	if err := pkg.SetConfigFromEnvVars(cfg); err != nil {
		panic(err)
	}

	logger := zap.InitializeLogger()

	// Init Open telemetry to control logs and flows
	telemetry := libOtel.InitializeTelemetry(&libOtel.TelemetryConfig{
		LibraryName:               cfg.OtelLibraryName,
		ServiceName:               cfg.OtelServiceName,
		ServiceVersion:            cfg.OtelServiceVersion,
		DeploymentEnv:             cfg.OtelDeploymentEnv,
		CollectorExporterEndpoint: cfg.OtelColExporterEndpoint,
		EnableTelemetry:           cfg.EnableTelemetry,
		Logger:                    logger,
	})

	// Config SeaweedFS connection
	seaweedFSEndpoint := fmt.Sprintf("http://%s:%s", cfg.SeaweedFSHost, cfg.SeaweedFSFilerPort)
	seaweedFSClient := simpleClient.NewSeaweedFSClient(seaweedFSEndpoint)

	// Init mongo DB connection
	escapedPass := url.QueryEscape(cfg.MongoDBPassword)
	mongoSource := fmt.Sprintf("%s://%s:%s@%s:%s",
		cfg.MongoURI, cfg.MongoDBUser, escapedPass, cfg.MongoDBHost, cfg.MongoDBPort)

	mongoConnection := &mongoDB.MongoConnection{
		ConnectionStringSource: mongoSource,
		Database:               cfg.MongoDBName,
		Logger:                 logger,
	}

	// Init rabbit MQ for producer
	rabbitSource := fmt.Sprintf("%s://%s:%s@%s:%s",
		cfg.RabbitURI, cfg.RabbitMQUser, cfg.RabbitMQPass, cfg.RabbitMQHost, cfg.RabbitMQPortAMQP)

	rabbitMQConnection := &libRabbitmq.RabbitMQConnection{
		ConnectionStringSource: rabbitSource,
		HealthCheckURL:         cfg.RabbitMQHealthCheckURL,
		Host:                   cfg.RabbitMQHost,
		Port:                   cfg.RabbitMQPortHost,
		User:                   cfg.RabbitMQUser,
		Pass:                   cfg.RabbitMQPass,
		Queue:                  cfg.RabbitMQGenerateReportQueue,
		Logger:                 logger,
	}

	templateMongoDBRepository := template.NewTemplateMongoDBRepository(mongoConnection)
	reportMongoDBRepository := report.NewReportMongoDBRepository(mongoConnection)

	// Create MongoDB indexes for optimal performance
	// Indexes are created automatically on startup to ensure they exist
	// This is idempotent and safe to run multiple times
	logger.Info("Ensuring MongoDB indexes exist for templates and reports...")
	ctx := pkg.ContextWithLogger(context.Background(), logger)

	if err := templateMongoDBRepository.EnsureIndexes(ctx); err != nil {
		logger.Warnf("Failed to ensure template indexes (non-fatal): %v", err)
	}

	if err := reportMongoDBRepository.EnsureIndexes(ctx); err != nil {
		logger.Warnf("Failed to ensure report indexes (non-fatal): %v", err)
	}

	templateSeaweedFSRepository := templateSeaweedFS.NewSimpleRepository(seaweedFSClient, constant.TemplateBucketName)
	reportSeaweedFSRepository := reportSeaweedFS.NewSimpleRepository(seaweedFSClient, constant.ReportBucketName)

	// Init Redis/Valkey connection
	redisConnection := &libRedis.RedisConnection{
		Address:                      strings.Split(cfg.RedisHost, ","),
		Password:                     cfg.RedisPassword,
		DB:                           cfg.RedisDB,
		Protocol:                     cfg.RedisProtocol,
		MasterName:                   cfg.RedisMasterName,
		UseTLS:                       cfg.RedisTLS,
		CACert:                       cfg.RedisCACert,
		UseGCPIAMAuth:                cfg.RedisUseGCPIAM,
		ServiceAccount:               cfg.RedisServiceAccount,
		GoogleApplicationCredentials: cfg.GoogleApplicationCredentials,
		TokenLifeTime:                time.Duration(cfg.RedisTokenLifeTime) * time.Minute,
		RefreshDuration:              time.Duration(cfg.RedisTokenRefreshDuration) * time.Minute,
		Logger:                       logger,
	}

	redisConsumerRepository := redis.NewConsumerRedis(redisConnection)

	templateService := &services.UseCase{
		TemplateRepo:        templateMongoDBRepository,
		TemplateSeaweedFS:   templateSeaweedFSRepository,
		ExternalDataSources: pkg.ExternalDatasourceConnections(logger),
	}

	authClient := middleware.NewAuthClient(cfg.AuthAddress, cfg.AuthEnabled, &logger)

	templateHandler := &in2.TemplateHandler{
		Service: templateService,
	}

	producerRabbitMQRepository := rabbitmq.NewProducerRabbitMQ(rabbitMQConnection)

	// Initialize PDF Pool with logger and timeout
	pdfPool := pdf.NewWorkerPool(cfg.PdfPoolWorkers, time.Duration(cfg.PdfPoolTimeoutSeconds)*time.Second, logger)
	logger.Infof("PDF Pool initialized with %d workers and %d seconds timeout", cfg.PdfPoolWorkers, cfg.PdfPoolTimeoutSeconds)

	// Initialize datasources in lazy mode (connect on-demand for faster startup)
	externalDataSources := pkg.ExternalDatasourceConnectionsLazy(logger)

	reportService := &services.UseCase{
		ReportRepo:          reportMongoDBRepository,
		RabbitMQRepo:        producerRabbitMQRepository,
		TemplateRepo:        templateMongoDBRepository,
		ReportSeaweedFS:     reportSeaweedFSRepository,
		ExternalDataSources: externalDataSources,
		PdfPool:             pdfPool,
	}

	reportHandler := &in2.ReportHandler{
		Service: reportService,
	}

	dataSourceService := &services.UseCase{
		ExternalDataSources: externalDataSources,
		RedisRepo:           redisConsumerRepository,
	}

	dataSourceHandler := &in2.DataSourceHandler{
		Service: dataSourceService,
	}

	licenseClient := libLicense.NewLicenseClient(
		constant.ApplicationName,
		cfg.LicenseKey,
		cfg.OrganizationIDs,
		&logger,
	)

	httpApp := in2.NewRoutes(logger, telemetry, templateHandler, reportHandler, dataSourceHandler, authClient, licenseClient)
	serverAPI := NewServer(cfg, httpApp, logger, telemetry, licenseClient)

	return &Service{
		Server: serverAPI,
		Logger: logger,
	}
}
