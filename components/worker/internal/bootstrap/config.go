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
	pkgConstant "github.com/LerianStudio/reporter/pkg/constant"
	reportData "github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/multitenant"
	"github.com/LerianStudio/reporter/pkg/pdf"
	"github.com/LerianStudio/reporter/pkg/pongo"
	reportSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/report"
	templateSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/template"
	"github.com/LerianStudio/reporter/pkg/storage"

	libCommons "github.com/LerianStudio/lib-commons/v3/commons"
	clog "github.com/LerianStudio/lib-commons/v3/commons/log"
	mongoDB "github.com/LerianStudio/lib-commons/v3/commons/mongo"
	libOtel "github.com/LerianStudio/lib-commons/v3/commons/opentelemetry"
	libRabbitMQ "github.com/LerianStudio/lib-commons/v3/commons/rabbitmq"
	tmclient "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/client"
	tmmongo "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/mongo"
	tmrabbitmq "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/rabbitmq"
	libZap "github.com/LerianStudio/lib-commons/v3/commons/zap"
	amqp091 "github.com/rabbitmq/amqp091-go"
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
	// Multi-tenant configuration envs
	MultiTenantEnabled                  bool   `env:"MULTI_TENANT_ENABLED" default:"false"`
	MultiTenantURL                      string `env:"MULTI_TENANT_URL"`
	MultiTenantEnvironment              string `env:"MULTI_TENANT_ENVIRONMENT" default:"staging"`
	MultiTenantMaxTenantPools           int    `env:"MULTI_TENANT_MAX_TENANT_POOLS" default:"100"`
	MultiTenantIdleTimeoutSec           int    `env:"MULTI_TENANT_IDLE_TIMEOUT_SEC" default:"300"`
	MultiTenantCircuitBreakerThreshold  int    `env:"MULTI_TENANT_CIRCUIT_BREAKER_THRESHOLD" default:"5"`
	MultiTenantCircuitBreakerTimeoutSec int    `env:"MULTI_TENANT_CIRCUIT_BREAKER_TIMEOUT_SEC" default:"30"`
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

	if c.MultiTenantEnabled && c.MultiTenantURL == "" {
		errs = append(errs, "MULTI_TENANT_URL is required when MULTI_TENANT_ENABLED=true")
	}

	if c.MultiTenantEnabled {
		if c.MultiTenantCircuitBreakerThreshold == 0 {
			errs = append(errs, "MULTI_TENANT_CIRCUIT_BREAKER_THRESHOLD must be > 0 when MULTI_TENANT_ENABLED=true (default: 5)")
		}

		if c.MultiTenantCircuitBreakerThreshold > 0 && c.MultiTenantCircuitBreakerTimeoutSec == 0 {
			errs = append(errs, "MULTI_TENANT_CIRCUIT_BREAKER_TIMEOUT_SEC must be > 0 when MULTI_TENANT_CIRCUIT_BREAKER_THRESHOLD > 0 (default: 30)")
		}
	}

	errs = c.validateProductionConfig(errs)

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n- %s", strings.Join(errs, "\n- "))
	}

	return nil
}

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
		switch s.value {
		case "":
			errs = append(errs, s.name+" must not be empty in production")
		case pkgConstant.DefaultPasswordPlaceholder:
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

	logger, err := libZap.InitializeLoggerWithError()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	if cfg.MultiTenantEnabled {
		logger.Info("Worker: multi-tenant mode enabled")
	} else {
		logger.Info("Worker: running in SINGLE-TENANT MODE")
	}

	// Build tenant MongoDB manager when multi-tenant mode is active and the
	// Tenant Manager URL is configured. Uses the same construction pattern
	// as the manager component (init_tenant.go) but scoped to ModuleWorker.
	// When either condition is false the manager remains nil and the worker
	// falls back to the static MongoDB connection (single-tenant behaviour).
	var tenantMongoManager *tmmongo.Manager

	if cfg.MultiTenantEnabled && cfg.MultiTenantURL != "" {
		var clientOpts []tmclient.ClientOption

		if cfg.MultiTenantCircuitBreakerThreshold > 0 {
			cbTimeout := time.Duration(cfg.MultiTenantCircuitBreakerTimeoutSec) * time.Second
			clientOpts = append(clientOpts,
				tmclient.WithCircuitBreaker(
					cfg.MultiTenantCircuitBreakerThreshold,
					cbTimeout,
				),
			)
		}

		tmClient := tmclient.NewClient(cfg.MultiTenantURL, logger, clientOpts...)
		tenantMongoManager = tmmongo.NewManager(
			tmClient,
			pkgConstant.ApplicationName,
			tmmongo.WithModule(pkgConstant.ModuleWorker),
			tmmongo.WithLogger(logger),
			tmmongo.WithMaxTenantPools(cfg.MultiTenantMaxTenantPools),
			tmmongo.WithIdleTimeout(time.Duration(cfg.MultiTenantIdleTimeoutSec)*time.Second),
		)
		logger.Info("Worker: tenant MongoDB manager initialized")
	}

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

	telemetry, err := libOtel.InitializeTelemetryWithError(&libOtel.TelemetryConfig{
		LibraryName:               cfg.OtelLibraryName,
		ServiceName:               cfg.OtelServiceName,
		ServiceVersion:            cfg.OtelServiceVersion,
		DeploymentEnv:             cfg.OtelDeploymentEnv,
		CollectorExporterEndpoint: cfg.OtelColExporterEndpoint,
		EnableTelemetry:           cfg.EnableTelemetry,
		Logger:                    logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	cleanups = append(cleanups, func() {
		logger.Info("Cleanup: shutting down telemetry")
		telemetry.ShutdownTelemetry()
	})

	// Init multi-tenant metrics (noop when disabled, real instruments when enabled)
	mtMetrics := initMultiTenantMetrics(cfg, telemetry, logger)

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
	mongoConnection := buildMongoConnection(cfg, logger)

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

	if !cfg.MultiTenantEnabled {
		logger.Info("Ensuring MongoDB indexes exist for reports...")

		if err = reportMongoDBRepository.EnsureIndexes(ctx); err != nil {
			return nil, fmt.Errorf("failed to ensure report indexes: %w", err)
		}
	}

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

	routes, err := initConsumerRoutes(cfg, rabbitMQConnection, logger, telemetry, tenantMongoManager, reportMongoDBRepository)
	if err != nil {
		return nil, err
	}

	cleanups = append(cleanups, closeRabbitMQ(rabbitMQConnection, logger))

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
		mtMetrics:          mtMetrics,
	}, nil
}

// buildMongoConnection creates a MongoConnection with the connection string
// built from configuration, applying default pool size if needed.
func buildMongoConnection(cfg *Config, logger clog.Logger) *mongoDB.MongoConnection {
	escapedPass := url.QueryEscape(cfg.MongoDBPassword)
	mongoSource := fmt.Sprintf("%s://%s:%s@%s:%s",
		cfg.MongoURI, cfg.MongoDBUser, escapedPass, cfg.MongoDBHost, cfg.MongoDBPort)

	if cfg.MongoDBParameters != "" {
		mongoSource += "/?" + cfg.MongoDBParameters
	}

	if cfg.MaxPoolSize <= 0 {
		cfg.MaxPoolSize = int(pkgConstant.MongoDBMaxPoolSize)
	}

	logger.Infof("MongoDB connecting to %s", pkg.RedactConnectionString(mongoSource))

	return &mongoDB.MongoConnection{
		ConnectionStringSource: mongoSource,
		Database:               cfg.MongoDBName,
		Logger:                 logger,
		MaxPoolSize:            uint64(cfg.MaxPoolSize),
	}
}

// initConsumerRoutes creates consumer routes with the appropriate constructor based on multi-tenant mode.
func initConsumerRoutes(
	cfg *Config,
	rabbitMQConnection *libRabbitMQ.RabbitMQConnection,
	logger clog.Logger,
	telemetry *libOtel.Telemetry,
	tenantMongoManager *tmmongo.Manager,
	reportMongoDBRepository *reportData.ReportMongoDBRepository,
) (*rabbitmq.ConsumerRoutes, error) {
	if cfg.MultiTenantEnabled && cfg.MultiTenantURL != "" {
		logger.Info("RabbitMQ Consumer: initializing multi-tenant consumer with vhost isolation")

		var clientOpts []tmclient.ClientOption

		if cfg.MultiTenantCircuitBreakerThreshold > 0 {
			cbTimeout := time.Duration(cfg.MultiTenantCircuitBreakerTimeoutSec) * time.Second
			clientOpts = append(clientOpts,
				tmclient.WithCircuitBreaker(
					cfg.MultiTenantCircuitBreakerThreshold,
					cbTimeout,
				),
			)
		}

		tmClient := tmclient.NewClient(cfg.MultiTenantURL, logger, clientOpts...)

		rabbitMQManager := tmrabbitmq.NewManager(
			tmClient,
			pkgConstant.ApplicationName,
			tmrabbitmq.WithModule(pkgConstant.ModuleWorker),
			tmrabbitmq.WithLogger(logger),
			tmrabbitmq.WithMaxTenantPools(cfg.MultiTenantMaxTenantPools),
			tmrabbitmq.WithIdleTimeout(time.Duration(cfg.MultiTenantIdleTimeoutSec)*time.Second),
		)

		rabbitMQManagerAdapter := newWorkerRabbitMQManagerAdapter(rabbitMQManager)

		routes, err := rabbitmq.NewConsumerRoutesMultiTenant(
			rabbitMQConnection, cfg.RabbitMQNumWorkers, logger, telemetry,
			tenantMongoManager, rabbitMQManagerAdapter, reportMongoDBRepository)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize multi-tenant rabbitmq consumer: %w", err)
		}

		logger.Info("RabbitMQ Consumer: multi-tenant consumer initialized with tmrabbitmq.Manager")

		return routes, nil
	}

	routes, err := rabbitmq.NewConsumerRoutes(rabbitMQConnection, cfg.RabbitMQNumWorkers, logger, telemetry, tenantMongoManager, reportMongoDBRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize rabbitmq consumer: %w", err)
	}

	return routes, nil
}

// closeRabbitMQ returns a cleanup function that safely closes
// the RabbitMQ channel and connection.
func closeRabbitMQ(conn *libRabbitMQ.RabbitMQConnection, logger clog.Logger) func() {
	return func() {
		logger.Info("Cleanup: closing RabbitMQ connection")

		if conn.Channel != nil {
			if closeErr := conn.Channel.Close(); closeErr != nil {
				logger.Errorf("Cleanup: failed to close RabbitMQ channel: %v", closeErr)
			}
		}

		if conn.Connection != nil && !conn.Connection.IsClosed() {
			if closeErr := conn.Connection.Close(); closeErr != nil {
				logger.Errorf("Cleanup: failed to close RabbitMQ connection: %v", closeErr)
			}
		}
	}
}

// initMultiTenantMetrics creates the multi-tenant OTel metrics instruments.
// When multi-tenant mode is enabled, real OTel instruments are registered on the
// telemetry MeterProvider so they are exported to the configured collector.
// When disabled, no-op instruments are returned with zero runtime overhead.
func initMultiTenantMetrics(cfg *Config, telemetry *libOtel.Telemetry, logger clog.Logger) *multitenant.Metrics {
	if !cfg.MultiTenantEnabled {
		logger.Info("Multi-tenant metrics: using noop instruments (multi-tenant disabled)")
		return multitenant.NoopMetrics()
	}

	meter := telemetry.MetricProvider.Meter(cfg.OtelLibraryName)

	m, err := multitenant.NewMetrics(meter)
	if err != nil {
		logger.Errorf("Failed to create multi-tenant metrics, falling back to noop: %v", err)
		return multitenant.NoopMetrics()
	}

	logger.Info("Multi-tenant metrics: 4 instruments registered (tenant_connections_total, tenant_connection_errors_total, tenant_consumers_active, tenant_messages_processed_total)")

	return m
}

// workerRabbitMQManagerAdapter wraps tmrabbitmq.Manager to satisfy the RabbitMQManagerConsumerInterface.
type workerRabbitMQManagerAdapter struct {
	manager *tmrabbitmq.Manager
}

func newWorkerRabbitMQManagerAdapter(manager *tmrabbitmq.Manager) *workerRabbitMQManagerAdapter {
	return &workerRabbitMQManagerAdapter{manager: manager}
}

// GetConnection wraps tmrabbitmq.Manager.GetConnection and converts the returned connection
// to our RabbitMQConnectionChannel interface.
func (a *workerRabbitMQManagerAdapter) GetConnection(ctx context.Context, tenantID string) (rabbitmq.RabbitMQConnectionChannel, error) {
	channel, err := a.manager.GetChannel(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return &workerAmqpChannelAdapter{channel: channel}, nil
}

// workerAmqpChannelAdapter wraps *amqp091.Channel to implement RabbitMQConnectionChannel interface.
type workerAmqpChannelAdapter struct {
	channel *amqp091.Channel
}

func (a *workerAmqpChannelAdapter) Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	return a.channel.Publish(exchange, key, mandatory, immediate, msg)
}
