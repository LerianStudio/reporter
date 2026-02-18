// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/LerianStudio/reporter/components/worker/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/components/worker/internal/services"
	"github.com/LerianStudio/reporter/pkg"
	cn "github.com/LerianStudio/reporter/pkg/constant"
	reportData "github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/pdf"
	"github.com/LerianStudio/reporter/pkg/pongo"
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
	HealthPort                  string `env:"HEALTH_PORT" default:"4006"`
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
	// Crypto configuration envs (for plugin_crm decryption)
	CryptoHashSecretKeyPluginCRM    string `env:"CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM"`
	CryptoEncryptSecretKeyPluginCRM string `env:"CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM"`
	// PDF Pool configuration envs
	PdfPoolWorkers        int `env:"PDF_POOL_WORKERS" default:"2"`
	PdfPoolTimeoutSeconds int `env:"PDF_TIMEOUT_SECONDS" default:"90"`
}

// Validate checks that all required configuration fields are present.
// Returns a descriptive multi-error message listing all missing fields.
func (c *Config) Validate() error {
	var errs []string

	if c.RabbitMQHost == "" {
		errs = append(errs, "RABBITMQ_HOST is required")
	}

	if c.RabbitMQPortAMQP == "" {
		errs = append(errs, "RABBITMQ_PORT_AMQP is required")
	}

	if c.RabbitMQUser == "" {
		errs = append(errs, "RABBITMQ_DEFAULT_USER is required")
	}

	if c.RabbitMQPass == "" {
		errs = append(errs, "RABBITMQ_DEFAULT_PASS is required")
	}

	if c.RabbitMQGenerateReportQueue == "" {
		errs = append(errs, "RABBITMQ_GENERATE_REPORT_QUEUE is required")
	}

	if c.MongoDBHost == "" {
		errs = append(errs, "MONGO_HOST is required")
	}

	if c.MongoDBName == "" {
		errs = append(errs, "MONGO_NAME is required")
	}

	if c.ObjectStorageEndpoint == "" {
		errs = append(errs, "OBJECT_STORAGE_ENDPOINT is required")
	}

	errs = c.validateProductionConfig(errs)

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n- %s", strings.Join(errs, "\n- "))
	}

	return nil
}

// defaultPassword is the placeholder value that must be replaced before
// deploying to production.
const defaultPassword = "CHANGE_ME"

// validateProductionConfig enforces stricter rules when EnvName is "production".
// Telemetry and real credentials are required in production.
func (c *Config) validateProductionConfig(errs []string) []string {
	if c.EnvName != "production" {
		return errs
	}

	if !c.EnableTelemetry {
		errs = append(errs, "ENABLE_TELEMETRY must be true in production")
	}

	secrets := []struct {
		value string
		name  string
	}{
		{c.MongoDBPassword, "MONGO_PASSWORD"},
		{c.RabbitMQPass, "RABBITMQ_DEFAULT_PASS"},
		{c.ObjectStorageSecretKey, "OBJECT_STORAGE_SECRET_KEY"},
		{c.CryptoHashSecretKeyPluginCRM, "CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM"},
		{c.CryptoEncryptSecretKeyPluginCRM, "CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM"},
	}

	for _, s := range secrets {
		if s.value == "" {
			errs = append(errs, s.name+" must not be empty in production")
		} else if s.value == defaultPassword {
			errs = append(errs, s.name+" must not use the default placeholder in production")
		}
	}

	return errs
}

// InitWorker initializes and configures the application's dependencies and returns the Service instance.
// Uses a cleanup stack pattern: if any initialization step fails, all previously
// opened connections are closed in reverse order to prevent resource leaks.
func InitWorker() (_ *Service, err error) {
	cfg := &Config{}
	if err := libCommons.SetConfigFromEnvVars(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config from env vars: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Register pongo2 custom filters and tags before any template processing
	if err := pongo.RegisterAll(); err != nil {
		return nil, fmt.Errorf("failed to register pongo2 filters and tags: %w", err)
	}

	logger := libZap.InitializeLogger()

	// Cleanup stack: on failure, close resources in reverse order
	var cleanups []func()

	defer func() {
		if err != nil {
			logger.Infof("Initialization failed, cleaning up %d resources...", len(cleanups))

			for i := len(cleanups) - 1; i >= 0; i-- {
				cleanups[i]()
			}
		}
	}()

	telemetry := libOtel.InitializeTelemetry(&libOtel.TelemetryConfig{
		LibraryName:               cfg.OtelLibraryName,
		ServiceName:               cfg.OtelServiceName,
		ServiceVersion:            cfg.OtelServiceVersion,
		DeploymentEnv:             cfg.OtelDeploymentEnv,
		CollectorExporterEndpoint: cfg.OtelColExporterEndpoint,
		EnableTelemetry:           cfg.EnableTelemetry,
		Logger:                    logger,
	})

	cleanups = append(cleanups, func() {
		logger.Info("Cleanup: shutting down telemetry")
		telemetry.ShutdownTelemetry()
	})

	rabbitSource := fmt.Sprintf("%s://%s:%s@%s:%s",
		cfg.RabbitURI, cfg.RabbitMQUser, cfg.RabbitMQPass, cfg.RabbitMQHost, cfg.RabbitMQPortAMQP)

	logger.Infof("RabbitMQ connecting to %s", pkg.RedactConnectionString(rabbitSource))

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

	routes, err := rabbitmq.NewConsumerRoutes(rabbitMQConnection, cfg.RabbitMQNumWorkers, logger, telemetry)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize rabbitmq consumer: %w", err)
	}

	cleanups = append(cleanups, func() {
		logger.Info("Cleanup: closing RabbitMQ connection")

		if rabbitMQConnection.Channel != nil {
			if closeErr := rabbitMQConnection.Channel.Close(); closeErr != nil {
				logger.Errorf("Cleanup: failed to close RabbitMQ channel: %v", closeErr)
			}
		}

		if rabbitMQConnection.Connection != nil && !rabbitMQConnection.Connection.IsClosed() {
			if closeErr := rabbitMQConnection.Connection.Close(); closeErr != nil {
				logger.Errorf("Cleanup: failed to close RabbitMQ connection: %v", closeErr)
			}
		}
	})

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
		return nil, fmt.Errorf("failed to create storage client: %w", err)
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
		cfg.MaxPoolSize = int(cn.MongoDBMaxPoolSize)
	}

	logger.Infof("MongoDB connecting to %s", pkg.RedactConnectionString(mongoSource))

	mongoConnection := &mongoDB.MongoConnection{
		ConnectionStringSource: mongoSource,
		Database:               cfg.MongoDBName,
		Logger:                 logger,
		MaxPoolSize:            uint64(cfg.MaxPoolSize),
	}

	templateSeaweedFSRepository := templateSeaweedFS.NewStorageRepository(templateStorageClient)
	reportSeaweedFSRepository := reportSeaweedFS.NewStorageRepository(reportStorageClient)

	// Initialize MongoDB repositories
	reportMongoDBRepository, err := reportData.NewReportMongoDBRepository(mongoConnection)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize report mongodb repository: %w", err)
	}

	cleanups = append(cleanups, func() {
		if mongoConnection.DB != nil {
			logger.Info("Cleanup: disconnecting MongoDB")

			if disconnectErr := mongoConnection.DB.Disconnect(context.Background()); disconnectErr != nil {
				logger.Errorf("Cleanup: failed to disconnect MongoDB: %v", disconnectErr)
			}
		}
	})

	// Create MongoDB indexes for optimal performance
	// Indexes are created automatically on startup to ensure they exist
	// This is idempotent and safe to run multiple times
	// Index failure is treated as fatal to match the manager component behavior
	logger.Info("Ensuring MongoDB indexes exist for reports...")

	if err = reportMongoDBRepository.EnsureIndexes(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure report indexes: %w", err)
	}

	// Initialize circuit breaker manager for datasource resilience
	circuitBreakerManager := pkg.NewCircuitBreakerManager(logger)
	externalDataSourcesMap := pkg.ExternalDatasourceConnections(logger)
	externalDataSources := pkg.NewSafeDataSources(externalDataSourcesMap)
	healthChecker := pkg.NewHealthChecker(&externalDataSourcesMap, circuitBreakerManager, logger)

	// Initialize PDF Pool for PDF generation
	pdfPool := pdf.NewWorkerPool(cfg.PdfPoolWorkers, time.Duration(cfg.PdfPoolTimeoutSeconds)*time.Second, logger)
	logger.Infof("PDF Pool initialized with %d workers and %d seconds timeout", cfg.PdfPoolWorkers, cfg.PdfPoolTimeoutSeconds)

	cleanups = append(cleanups, func() {
		logger.Info("Cleanup: closing PDF worker pool")
		pdfPool.Close()
	})

	service := &services.UseCase{
		TemplateSeaweedFS:               templateSeaweedFSRepository,
		ReportSeaweedFS:                 reportSeaweedFSRepository,
		ExternalDataSources:             externalDataSources,
		ReportDataRepo:                  reportMongoDBRepository,
		CircuitBreakerManager:           circuitBreakerManager,
		HealthChecker:                   healthChecker,
		ReportTTL:                       "", // TTL not supported in S3 mode - use bucket lifecycle policies
		PdfPool:                         pdfPool,
		CryptoHashSecretKeyPluginCRM:    cfg.CryptoHashSecretKeyPluginCRM,
		CryptoEncryptSecretKeyPluginCRM: cfg.CryptoEncryptSecretKeyPluginCRM,
	}

	logger.Infof("Reports will be stored permanently (no TTL - use S3 bucket lifecycle policies for expiration)")

	// Start health checker in background
	healthChecker.Start()

	cleanups = append(cleanups, func() {
		logger.Info("Cleanup: stopping health checker")
		healthChecker.Stop()
	})

	multiQueueConsumer := NewMultiQueueConsumer(routes, service, cfg.RabbitMQGenerateReportQueue, logger)

	healthServer := NewHealthServer(cfg.HealthPort, rabbitMQConnection, logger)
	logger.Infof("Health server configured on port %s (/health, /ready)", cfg.HealthPort)

	return &Service{
		MultiQueueConsumer: multiQueueConsumer,
		Logger:             logger,
		healthChecker:      healthChecker,
		healthServer:       healthServer,
		mongoConnection:    mongoConnection,
		rabbitMQConnection: rabbitMQConnection,
		pdfPool:            pdfPool,
		telemetry:          telemetry,
	}, nil
}
