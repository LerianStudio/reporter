// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/LerianStudio/reporter/components/manager/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/components/manager/internal/adapters/redis"
	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/pkg/multitenant"
	"github.com/LerianStudio/reporter/pkg/storage"

	libCommons "github.com/LerianStudio/lib-commons/v3/commons"
	"github.com/LerianStudio/lib-commons/v3/commons/log"
	mongoDB "github.com/LerianStudio/lib-commons/v3/commons/mongo"
	libOtel "github.com/LerianStudio/lib-commons/v3/commons/opentelemetry"
	libRabbitmq "github.com/LerianStudio/lib-commons/v3/commons/rabbitmq"
	libRedis "github.com/LerianStudio/lib-commons/v3/commons/redis"
	"github.com/LerianStudio/lib-commons/v3/commons/zap"
)

// mongoResources holds MongoDB-related resources created during initialization.
type mongoResources struct {
	connection   *mongoDB.MongoConnection
	templateRepo *template.TemplateMongoDBRepository
	reportRepo   *report.ReportMongoDBRepository
}

// rabbitResources holds RabbitMQ-related resources created during initialization.
type rabbitResources struct {
	connection *libRabbitmq.RabbitMQConnection
	producer   *rabbitmq.ProducerRabbitMQRepository
	monitor    *RabbitMQMonitor
}

// initConfigAndLogger loads configuration from environment variables, validates it,
// and initializes the structured logger.
func initConfigAndLogger() (*Config, log.Logger, error) {
	cfg := &Config{}
	if err := libCommons.SetConfigFromEnvVars(cfg); err != nil {
		return nil, nil, fmt.Errorf("failed to load config from env vars: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}

	logger, err := zap.InitializeLoggerWithError()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return cfg, logger, nil
}

// initTelemetry initializes OpenTelemetry tracing and returns the telemetry instance
// along with a cleanup function that shuts down the telemetry provider.
func initTelemetry(cfg *Config, logger log.Logger) (*libOtel.Telemetry, func(), error) {
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
		return nil, nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	cleanup := func() {
		logger.Info("Cleanup: shutting down telemetry")
		telemetry.ShutdownTelemetry()
	}

	return telemetry, cleanup, nil
}

// initStorage creates the S3-compatible object storage client used for both
// template and report file storage (differentiated by key prefix).
func initStorage(cfg *Config, logger log.Logger) (storage.ObjectStorage, error) {
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

	return storageClient, nil
}

// initMongoDB establishes the MongoDB connection, creates template and report
// repositories, ensures indexes exist, and returns a cleanup function that
// disconnects the client.
func initMongoDB(cfg *Config, logger log.Logger) (*mongoResources, func(), error) {
	escapedPass := url.QueryEscape(cfg.MongoDBPassword)
	mongoSource := fmt.Sprintf("%s://%s:%s@%s:%s",
		cfg.MongoURI, cfg.MongoDBUser, escapedPass, cfg.MongoDBHost, cfg.MongoDBPort)

	if cfg.MongoDBParameters != "" {
		mongoSource += "/?" + cfg.MongoDBParameters
	}

	mongoMaxPoolSize, _ := strconv.ParseUint(cfg.MongoMaxPoolSize, 10, 64)
	if mongoMaxPoolSize == 0 {
		mongoMaxPoolSize = constant.MongoDefaultMaxPoolSize
	}

	logger.Infof("MongoDB connecting to %s", pkg.RedactConnectionString(mongoSource))

	mongoConnection := &mongoDB.MongoConnection{
		ConnectionStringSource: mongoSource,
		Database:               cfg.MongoDBName,
		Logger:                 logger,
		MaxPoolSize:            mongoMaxPoolSize,
	}

	templateMongoDBRepository, err := template.NewTemplateMongoDBRepository(mongoConnection)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize template mongodb repository: %w", err)
	}

	reportMongoDBRepository, err := report.NewReportMongoDBRepository(mongoConnection)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize report mongodb repository: %w", err)
	}

	// Create MongoDB indexes
	logger.Info("Ensuring MongoDB indexes exist for templates and reports...")

	ctx := pkg.ContextWithLogger(context.Background(), logger)

	if err = templateMongoDBRepository.EnsureIndexes(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to ensure template indexes: %w", err)
	}

	if err = reportMongoDBRepository.EnsureIndexes(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to ensure report indexes: %w", err)
	}

	cleanup := func() {
		if mongoConnection.DB != nil {
			logger.Info("Cleanup: disconnecting MongoDB")

			if disconnectErr := mongoConnection.DB.Disconnect(context.Background()); disconnectErr != nil {
				logger.Errorf("Cleanup: failed to disconnect MongoDB: %v", disconnectErr)
			}
		}
	}

	return &mongoResources{
		connection:   mongoConnection,
		templateRepo: templateMongoDBRepository,
		reportRepo:   reportMongoDBRepository,
	}, cleanup, nil
}

// initRabbitMQ establishes the RabbitMQ connection, creates the producer,
// starts the background connection monitor, and returns cleanup functions for
// the monitor and the connection itself.
func initRabbitMQ(cfg *Config, logger log.Logger) (*rabbitResources, []func()) {
	rabbitSource := fmt.Sprintf("%s://%s:%s@%s:%s",
		cfg.RabbitURI, cfg.RabbitMQUser, cfg.RabbitMQPass, cfg.RabbitMQHost, cfg.RabbitMQPortAMQP)

	logger.Infof("RabbitMQ connecting to %s", pkg.RedactConnectionString(rabbitSource))

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

	producerRabbitMQRepository := rabbitmq.NewProducerRabbitMQ(rabbitMQConnection)

	// Start background RabbitMQ connection monitor.
	// This goroutine periodically checks if the connection is alive and
	// calls EnsureChannel() to reconnect when needed, breaking the deadlock
	// where /ready returns 503 but nothing triggers reconnection.
	rabbitMQMonitor := NewRabbitMQMonitor(rabbitMQConnection, logger)
	rabbitMQMonitor.Start()

	logger.Info("RabbitMQ background connection monitor started")

	cleanups := []func(){
		func() {
			logger.Info("Cleanup: stopping RabbitMQ connection monitor")
			rabbitMQMonitor.Stop()
		},
		func() {
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
		},
	}

	return &rabbitResources{
		connection: rabbitMQConnection,
		producer:   producerRabbitMQRepository,
		monitor:    rabbitMQMonitor,
	}, cleanups
}

// initRedis establishes the Redis/Valkey connection and returns the consumer
// repository along with a cleanup function that closes the connection.
func initRedis(cfg *Config, logger log.Logger) (*redis.RedisConsumerRepository, *libRedis.RedisConnection, func(), error) {
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

	redisConsumerRepository, err := redis.NewConsumerRedis(redisConnection)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize redis connection: %w", err)
	}

	cleanup := func() {
		logger.Info("Cleanup: closing Redis connection")

		if closeErr := redisConnection.Close(); closeErr != nil {
			logger.Errorf("Cleanup: failed to close Redis connection: %v", closeErr)
		}
	}

	return redisConsumerRepository, redisConnection, cleanup, nil
}

// initMultiTenantMetrics creates the multi-tenant OTel metrics instruments.
// When multi-tenant mode is enabled, real OTel instruments are registered on the
// telemetry MeterProvider so they are exported to the configured collector.
// When disabled, no-op instruments are returned with zero runtime overhead.
func initMultiTenantMetrics(cfg *Config, telemetry *libOtel.Telemetry, logger log.Logger) *multitenant.Metrics {
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
