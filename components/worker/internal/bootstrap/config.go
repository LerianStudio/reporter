// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/LerianStudio/reporter/components/worker/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/components/worker/internal/services"
	"github.com/LerianStudio/reporter/pkg"
	reportData "github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/pdf"
	reportSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/report"
	templateSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/template"
	"github.com/LerianStudio/reporter/pkg/storage"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	mongoDB "github.com/LerianStudio/lib-commons/v2/commons/mongo"
	libOtel "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	libRabbitMQ "github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
	libZap "github.com/LerianStudio/lib-commons/v2/commons/zap"
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
	// Storage configuration envs (S3-compatible only)
	ObjectStorageEndpoint     string `env:"OBJECT_STORAGE_ENDPOINT"`
	ObjectStorageRegion       string `env:"OBJECT_STORAGE_REGION" default:"us-east-1"`
	ObjectStorageAccessKeyID  string `env:"OBJECT_STORAGE_ACCESS_KEY_ID"`
	ObjectStorageSecretKey    string `env:"OBJECT_STORAGE_SECRET_KEY"`
	ObjectStorageUsePathStyle bool   `env:"OBJECT_STORAGE_USE_PATH_STYLE" default:"false"`
	ObjectStorageDisableSSL   bool   `env:"OBJECT_STORAGE_DISABLE_SSL" default:"false"`
	ObjectStorageBucket       string `env:"OBJECT_STORAGE_BUCKET" default:"reporter-storage"` // Single bucket for templates/ and reports/ prefixes
	// MongoDB
	MongoURI          string `env:"MONGO_URI"`
	MongoDBHost       string `env:"MONGO_HOST"`
	MongoDBName       string `env:"MONGO_NAME"`
	MongoDBUser       string `env:"MONGO_USER"`
	MongoDBPassword   string `env:"MONGO_PASSWORD"`
	MongoDBPort       string `env:"MONGO_PORT"`
	MongoDBParameters string `env:"MONGO_PARAMETERS"`
	MaxPoolSize       int    `env:"MONGO_MAX_POOL_SIZE"`
	// PDF Pool configuration envs
	PdfPoolWorkers        int `env:"PDF_POOL_WORKERS" default:"2"`
	PdfPoolTimeoutSeconds int `env:"PDF_TIMEOUT_SECONDS" default:"90"`
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

	logger.Infof("RabbitMQ connecting to %s:%s", cfg.RabbitMQHost, cfg.RabbitMQPortAMQP)

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

	// Create single storage client for both templates and reports (using prefixes)
	storageConfig := storage.Config{
		Bucket:            cfg.ObjectStorageBucket,
		S3Endpoint:        cfg.ObjectStorageEndpoint,
		S3Region:          cfg.ObjectStorageRegion,
		S3AccessKeyID:     cfg.ObjectStorageAccessKeyID,
		S3SecretAccessKey: cfg.ObjectStorageSecretKey,
		S3UsePathStyle:    cfg.ObjectStorageUsePathStyle,
		S3DisableSSL:      cfg.ObjectStorageDisableSSL,
	}

	ctx := pkg.ContextWithLogger(context.Background(), logger)

	storageClient, err := storage.NewStorageClient(ctx, storageConfig)
	if err != nil {
		logger.Fatalf("Failed to create storage client: %v", err)
	}

	logger.Infof("Storage initialized with bucket: %s (templates/ and reports/ prefixes)", cfg.ObjectStorageBucket)

	// Use same storage client for both templates and reports (repositories handle prefixes)
	templateStorageClient := storageClient
	reportStorageClient := storageClient

	// Init mongo DB connection
	escapedPass := url.QueryEscape(cfg.MongoDBPassword)
	mongoSource := fmt.Sprintf("%s://%s:%s@%s:%s",
		cfg.MongoURI, cfg.MongoDBUser, escapedPass, cfg.MongoDBHost, cfg.MongoDBPort)

	if cfg.MongoDBParameters != "" {
		mongoSource += "/?" + cfg.MongoDBParameters
	}

	if cfg.MaxPoolSize <= 0 {
		cfg.MaxPoolSize = 100
	}

	mongoConnection := &mongoDB.MongoConnection{
		ConnectionStringSource: mongoSource,
		Database:               cfg.MongoDBName,
		Logger:                 logger,
		MaxPoolSize:            uint64(cfg.MaxPoolSize),
	}

	templateSeaweedFSRepository := templateSeaweedFS.NewStorageRepository(templateStorageClient)
	reportSeaweedFSRepository := reportSeaweedFS.NewStorageRepository(reportStorageClient)

	// Initialize MongoDB repositories
	reportMongoDBRepository := reportData.NewReportMongoDBRepository(mongoConnection)

	// Create MongoDB indexes for optimal performance
	// Indexes are created automatically on startup to ensure they exist
	// This is idempotent and safe to run multiple times
	logger.Info("Ensuring MongoDB indexes exist for reports...")

	if err = reportMongoDBRepository.EnsureIndexes(ctx); err != nil {
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
		TemplateSeaweedFS:     templateSeaweedFSRepository,
		ReportSeaweedFS:       reportSeaweedFSRepository,
		ExternalDataSources:   externalDataSources,
		ReportDataRepo:        reportMongoDBRepository,
		CircuitBreakerManager: circuitBreakerManager,
		HealthChecker:         healthChecker,
		ReportTTL:             "", // TTL not supported in S3 mode - use bucket lifecycle policies
		PdfPool:               pdfPool,
	}

	logger.Infof("Reports will be stored permanently (no TTL - use S3 bucket lifecycle policies for expiration)")

	// Start health checker in background
	healthChecker.Start()

	multiQueueConsumer := NewMultiQueueConsumer(routes, service)

	return &Service{
		MultiQueueConsumer: multiQueueConsumer,
		Logger:             logger,
		healthChecker:      healthChecker,
	}
}
