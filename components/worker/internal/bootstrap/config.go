package bootstrap

import (
	"fmt"
	"net/url"
	"plugin-smart-templates/v3/components/worker/internal/adapters/rabbitmq"
	"plugin-smart-templates/v3/components/worker/internal/services"
	"plugin-smart-templates/v3/pkg"
	"plugin-smart-templates/v3/pkg/constant"
	reportFile "plugin-smart-templates/v3/pkg/minio/report"
	templateFile "plugin-smart-templates/v3/pkg/minio/template"
	reportData "plugin-smart-templates/v3/pkg/mongodb/report"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	mongoDB "github.com/LerianStudio/lib-commons/v2/commons/mongo"
	libOtel "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	libRabbitMQ "github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
	libZap "github.com/LerianStudio/lib-commons/v2/commons/zap"
	libLicense "github.com/LerianStudio/lib-license-go/v2/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config holds the application's configurable parameters read from environment variables.
type Config struct {
	EnvName                     string `env:"ENV_NAME"`
	LogLevel                    string `env:"LOG_LEVEL"`
	RabbitURI                   string `env:"RABBITMQ_URI"`
	RabbitMQHost                string `env:"RABBITMQ_HOST"`
	RabbitMQPortHost            string `env:"RABBITMQ_PORT_HOST"`
	RabbitMQPortAMQP            string `env:"RABBITMQ_PORT_AMQP"`
	RabbitMQUser                string `env:"RABBITMQ_DEFAULT_USER"`
	RabbitMQPass                string `env:"RABBITMQ_DEFAULT_PASS"`
	RabbitMQGenerateReportQueue string `env:"RABBITMQ_GENERATE_REPORT_QUEUE"`
	RabbitMQNumWorkers          int    `env:"RABBITMQ_NUMBERS_OF_WORKERS"`
	RabbitMQHealthCheckURL      string `env:"RABBITMQ_HEALTH_CHECK_URL"`
	OtelServiceName             string `env:"OTEL_RESOURCE_SERVICE_NAME"`
	OtelLibraryName             string `env:"OTEL_LIBRARY_NAME"`
	OtelServiceVersion          string `env:"OTEL_RESOURCE_SERVICE_VERSION"`
	OtelDeploymentEnv           string `env:"OTEL_RESOURCE_DEPLOYMENT_ENVIRONMENT"`
	OtelColExporterEndpoint     string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	EnableTelemetry             bool   `env:"ENABLE_TELEMETRY"`
	// MINIO
	MinioAPIHost     string `env:"MINIO_API_HOST"`
	MinioAPIPort     string `env:"MINIO_API_PORT"`
	MinioSSLEnabled  bool   `env:"MINIO_SSL_ENABLED"`
	MinioAppUsername string `env:"MINIO_APP_USER"`
	MinioAppPassword string `env:"MINIO_APP_PASSWORD"`
	// MongoDB
	MongoURI        string `env:"MONGO_URI"`
	MongoDBHost     string `env:"MONGO_HOST"`
	MongoDBName     string `env:"MONGO_NAME"`
	MongoDBUser     string `env:"MONGO_USER"`
	MongoDBPassword string `env:"MONGO_PASSWORD"`
	MongoDBPort     string `env:"MONGO_PORT"`
	MaxPoolSize     int    `env:"MONGO_MAX_POOL_SIZE"`
	// License configuration envs
	LicenseKey      string `env:"LICENSE_KEY"`
	OrganizationIDs string `env:"ORGANIZATION_IDS"`
}

// InitWorker initializes and configures the application's dependencies and returns the Service instance.
func InitWorker() *Service {
	cfg := &Config{}
	if err := libCommons.SetConfigFromEnvVars(cfg); err != nil {
		panic(err)
	}

	logger := libZap.InitializeLogger()

	telemetry := libOtel.InitializeTelemetry(&libOtel.TelemetryConfig{
		LibraryName:               cfg.OtelLibraryName,
		ServiceName:               cfg.OtelServiceName,
		ServiceVersion:            cfg.OtelServiceVersion,
		DeploymentEnv:             cfg.OtelDeploymentEnv,
		CollectorExporterEndpoint: cfg.OtelColExporterEndpoint,
		EnableTelemetry:           cfg.EnableTelemetry,
		Logger:                    logger,
	})

	rabbitSource := fmt.Sprintf("%s://%s:%s@%s:%s",
		cfg.RabbitURI, cfg.RabbitMQUser, cfg.RabbitMQPass, cfg.RabbitMQHost, cfg.RabbitMQPortAMQP)

	logger.Infof(rabbitSource)

	rabbitMQConnection := &libRabbitMQ.RabbitMQConnection{
		ConnectionStringSource: rabbitSource,
		HealthCheckURL:         cfg.RabbitMQHealthCheckURL,
		Host:                   cfg.RabbitMQHost,
		Port:                   cfg.RabbitMQPortHost,
		User:                   cfg.RabbitMQUser,
		Pass:                   cfg.RabbitMQPass,
		Queue:                  cfg.RabbitMQGenerateReportQueue,
		Logger:                 logger,
	}

	routes := rabbitmq.NewConsumerRoutes(rabbitMQConnection, cfg.RabbitMQNumWorkers, logger, telemetry)

	minioEndpoint := fmt.Sprintf("%s:%s", cfg.MinioAPIHost, cfg.MinioAPIPort)

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAppUsername, cfg.MinioAppPassword, ""),
		Secure: cfg.MinioSSLEnabled,
	})
	if err != nil {
		logger.Fatalf("Error creating minio client: %v", err)
	}

	// Init mongo DB connection
	escapedPass := url.QueryEscape(cfg.MongoDBPassword)
	mongoSource := fmt.Sprintf("%s://%s:%s@%s:%s",
		cfg.MongoURI, cfg.MongoDBUser, escapedPass, cfg.MongoDBHost, cfg.MongoDBPort)

	if cfg.MaxPoolSize <= 0 {
		cfg.MaxPoolSize = 100
	}

	mongoConnection := &mongoDB.MongoConnection{
		ConnectionStringSource: mongoSource,
		Database:               cfg.MongoDBName,
		Logger:                 logger,
		MaxPoolSize:            uint64(cfg.MaxPoolSize),
	}

	// Initialize circuit breaker manager for datasource resilience
	circuitBreakerManager := pkg.NewCircuitBreakerManager(logger)
	externalDataSources := pkg.ExternalDatasourceConnections(logger)
	healthChecker := pkg.NewHealthChecker(&externalDataSources, circuitBreakerManager, logger)

	service := &services.UseCase{
		TemplateFileRepo:      templateFile.NewMinioRepository(minioClient, "templates"),
		ReportFileRepo:        reportFile.NewMinioRepository(minioClient, "reports"),
		ExternalDataSources:   externalDataSources,
		ReportDataRepo:        reportData.NewReportMongoDBRepository(mongoConnection),
		CircuitBreakerManager: circuitBreakerManager,
		HealthChecker:         healthChecker,
	}

	// Start health checker in background
	healthChecker.Start()

	licenseClient := libLicense.NewLicenseClient(
		constant.ApplicationName,
		cfg.LicenseKey,
		cfg.OrganizationIDs,
		&logger,
	)
	multiQueueConsumer := NewMultiQueueConsumer(routes, service)

	return &Service{
		MultiQueueConsumer: multiQueueConsumer,
		Logger:             logger,
		licenseShutdown:    licenseClient.GetLicenseManagerShutdown(),
		healthChecker:      healthChecker,
	}
}
