package bootstrap

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/LerianStudio/reporter/v4/components/worker/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/v4/components/worker/internal/services"
	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	reportData "github.com/LerianStudio/reporter/v4/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/v4/pkg/pdf"
	"github.com/LerianStudio/reporter/v4/pkg/storage"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	mongoDB "github.com/LerianStudio/lib-commons/v2/commons/mongo"
	libOtel "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	libRabbitMQ "github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
	libZap "github.com/LerianStudio/lib-commons/v2/commons/zap"
	libLicense "github.com/LerianStudio/lib-license-go/v2/middleware"
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
	// SeaweedFS configuration envs
	SeaweedFSHost      string `env:"SEAWEEDFS_HOST"`
	SeaweedFSFilerPort string `env:"SEAWEEDFS_FILER_PORT"`
	SeaweedFSTTL       string `env:"SEAWEEDFS_TTL"`
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
	// PDF Pool configuration envs
	PdfPoolWorkers        int `env:"PDF_POOL_WORKERS" default:"2"`
	PdfPoolTimeoutSeconds int `env:"PDF_TIMEOUT_SECONDS" default:"90"`

	// Storage configuration envs (S3 support - optional, defaults to SeaweedFS)
	StorageProvider   string `env:"STORAGE_PROVIDER"` // "seaweedfs" (default) or "s3"
	S3Region          string `env:"S3_REGION"`
	S3Bucket          string `env:"S3_BUCKET"`
	S3AccessKeyID     string `env:"S3_ACCESS_KEY_ID"`
	S3SecretAccessKey string `env:"S3_SECRET_ACCESS_KEY"`
	S3Endpoint        string `env:"S3_ENDPOINT"`         // For MinIO/LocalStack
	S3ForcePathStyle  bool   `env:"S3_FORCE_PATH_STYLE"` // For MinIO compatibility
	S3TemplateBucket  string `env:"S3_TEMPLATE_BUCKET"`  // Optional: separate bucket for templates
	S3ReportBucket    string `env:"S3_REPORT_BUCKET"`    // Optional: separate bucket for reports
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

	// Create storage repositories (supports both SeaweedFS and S3)
	storageConfig := &storage.Config{
		Provider:           cfg.StorageProvider,
		S3Region:           cfg.S3Region,
		S3Bucket:           cfg.S3Bucket,
		S3AccessKeyID:      cfg.S3AccessKeyID,
		S3SecretAccessKey:  cfg.S3SecretAccessKey,
		S3Endpoint:         cfg.S3Endpoint,
		S3ForcePathStyle:   cfg.S3ForcePathStyle,
		S3TemplateBucket:   cfg.S3TemplateBucket,
		S3ReportBucket:     cfg.S3ReportBucket,
		SeaweedFSHost:      cfg.SeaweedFSHost,
		SeaweedFSFilerPort: cfg.SeaweedFSFilerPort,
	}

	templateStorageRepository, err := storage.CreateTemplateRepository(storageConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create template storage repository: %v", err))
	}

	reportStorageRepository, err := storage.CreateReportRepository(storageConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create report storage repository: %v", err))
	}

	// Log which storage provider is being used
	provider := storageConfig.Provider
	if provider == "" || provider == "seaweedfs" {
		provider = "seaweedfs" // default
	}

	logger.Infof("Using storage provider: %s", provider)

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

	// Initialize MongoDB repositories
	reportMongoDBRepository := reportData.NewReportMongoDBRepository(mongoConnection)

	// Create MongoDB indexes for optimal performance
	// Indexes are created automatically on startup to ensure they exist
	// This is idempotent and safe to run multiple times
	logger.Info("Ensuring MongoDB indexes exist for reports...")
	ctx := pkg.ContextWithLogger(context.Background(), logger)

	if err := reportMongoDBRepository.EnsureIndexes(ctx); err != nil {
		logger.Warnf("Failed to ensure report indexes (non-fatal): %v", err)
	}

	// Initialize circuit breaker manager for datasource resilience
	circuitBreakerManager := pkg.NewCircuitBreakerManager(logger)
	externalDataSources := pkg.ExternalDatasourceConnections(logger)
	healthChecker := pkg.NewHealthChecker(&externalDataSources, circuitBreakerManager, logger)

	// Initialize PDF Pool for PDF generation
	pdfPool := pdf.NewWorkerPool(cfg.PdfPoolWorkers, time.Duration(cfg.PdfPoolTimeoutSeconds)*time.Second, logger)
	logger.Infof("PDF Pool initialized with %d workers and %d seconds timeout", cfg.PdfPoolWorkers, cfg.PdfPoolTimeoutSeconds)

	service := &services.UseCase{
		TemplateStorage:       templateStorageRepository,
		ReportStorage:         reportStorageRepository,
		ExternalDataSources:   externalDataSources,
		ReportDataRepo:        reportMongoDBRepository,
		CircuitBreakerManager: circuitBreakerManager,
		HealthChecker:         healthChecker,
		ReportTTL:             cfg.SeaweedFSTTL,
		PdfPool:               pdfPool,
	}

	if cfg.SeaweedFSTTL != "" {
		logger.Infof("Reports will expire after: %s", cfg.SeaweedFSTTL)
	} else {
		logger.Infof("Reports will be stored permanently (no TTL)")
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
