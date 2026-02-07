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

	in2 "github.com/LerianStudio/reporter/components/manager/internal/adapters/http/in"
	"github.com/LerianStudio/reporter/components/manager/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/components/manager/internal/adapters/redis"
	"github.com/LerianStudio/reporter/components/manager/internal/services"
	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	reportSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/report"
	templateSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/template"
	"github.com/LerianStudio/reporter/pkg/storage"

	"github.com/LerianStudio/lib-auth/v2/auth/middleware"
	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	mongoDB "github.com/LerianStudio/lib-commons/v2/commons/mongo"
	libOtel "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	libRabbitmq "github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
	libRedis "github.com/LerianStudio/lib-commons/v2/commons/redis"
	"github.com/LerianStudio/lib-commons/v2/commons/zap"
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
	MongoURI             string        `env:"MONGO_URI"`
	MongoDBHost          string        `env:"MONGO_HOST"`
	MongoDBName          string        `env:"MONGO_NAME"`
	MongoDBUser          string        `env:"MONGO_USER"`
	MongoDBPassword      string        `env:"MONGO_PASSWORD"`
	MongoDBPort          string        `env:"MONGO_PORT"`
	MongoDBParameters    string        `env:"MONGO_PARAMETERS"`
	MongoMaxPoolSize     string `env:"MONGO_MAX_POOL_SIZE" default:"100"`
	MongoMinPoolSize     string `env:"MONGO_MIN_POOL_SIZE" default:"10"`
	MongoMaxConnIdleTime string `env:"MONGO_MAX_CONN_IDLE_TIME" default:"60s"`
	// Storage configuration envs (S3-compatible only)
	ObjectStorageEndpoint     string `env:"OBJECT_STORAGE_ENDPOINT"`
	ObjectStorageRegion       string `env:"OBJECT_STORAGE_REGION" default:"us-east-1"`
	ObjectStorageAccessKeyID  string `env:"OBJECT_STORAGE_ACCESS_KEY_ID"`
	ObjectStorageSecretKey    string `env:"OBJECT_STORAGE_SECRET_KEY"`
	ObjectStorageUsePathStyle bool   `env:"OBJECT_STORAGE_USE_PATH_STYLE" default:"false"`
	ObjectStorageDisableSSL   bool   `env:"OBJECT_STORAGE_DISABLE_SSL" default:"false"`
	ObjectStorageBucket       string `env:"OBJECT_STORAGE_BUCKET" default:"reporter-storage"` // Single bucket for templates/ and reports/ prefixes
	// RabbitMQ configuration envs
	RabbitURI                   string `env:"RABBITMQ_URI"`
	RabbitMQHost                string `env:"RABBITMQ_HOST"`
	RabbitMQHealthCheckURL      string `env:"RABBITMQ_HEALTH_CHECK_URL"`
	RabbitMQPortHost            string `env:"RABBITMQ_PORT_HOST"`
	RabbitMQPortAMQP            string `env:"RABBITMQ_PORT_AMQP"`
	RabbitMQUser                string `env:"RABBITMQ_DEFAULT_USER"`
	RabbitMQPass                string `env:"RABBITMQ_DEFAULT_PASS"`
	RabbitMQGenerateReportQueue string `env:"RABBITMQ_GENERATE_REPORT_QUEUE"`
	RabbitMQExchange            string `env:"RABBITMQ_EXCHANGE"`
	RabbitMQGenerateReportKey   string `env:"RABBITMQ_GENERATE_REPORT_KEY"`
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
}

// Validate checks that all required configuration fields are present
// and that optional numeric bounds are consistent.
// Returns a descriptive multi-error message listing all violations.
func (c *Config) Validate() error {
	var errs []string

	errs = c.validateRequiredFields(errs)
	errs = c.validateMongoPoolBounds(errs)

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n- %s", strings.Join(errs, "\n- "))
	}

	return nil
}

// validateRequiredFields appends an error for each required configuration
// field that is empty and returns the accumulated slice.
func (c *Config) validateRequiredFields(errs []string) []string {
	required := []struct {
		value string
		name  string
	}{
		{c.ServerAddress, "SERVER_ADDRESS"},
		{c.MongoDBHost, "MONGO_HOST"},
		{c.MongoDBName, "MONGO_NAME"},
		{c.RabbitMQHost, "RABBITMQ_HOST"},
		{c.RabbitMQPortAMQP, "RABBITMQ_PORT_AMQP"},
		{c.RabbitMQUser, "RABBITMQ_DEFAULT_USER"},
		{c.RabbitMQPass, "RABBITMQ_DEFAULT_PASS"},
		{c.RabbitMQGenerateReportQueue, "RABBITMQ_GENERATE_REPORT_QUEUE"},
		{c.RabbitMQExchange, "RABBITMQ_EXCHANGE"},
		{c.RabbitMQGenerateReportKey, "RABBITMQ_GENERATE_REPORT_KEY"},
		{c.RedisHost, "REDIS_HOST"},
		{c.ObjectStorageEndpoint, "OBJECT_STORAGE_ENDPOINT"},
	}

	for _, r := range required {
		if r.value == "" {
			errs = append(errs, r.name+" is required")
		}
	}

	return errs
}

// validateMongoPoolBounds checks that MongoDB connection pool size
// parameters are within allowed ranges and consistent with each other.
func (c *Config) validateMongoPoolBounds(errs []string) []string {
	const mongoMaxPoolSizeUpperBound = 10000

	maxPool, err := strconv.ParseUint(c.MongoMaxPoolSize, 10, 64)
	if err != nil && c.MongoMaxPoolSize != "" {
		errs = append(errs, "MONGO_MAX_POOL_SIZE must be a valid integer")
		return errs
	}

	minPool, err := strconv.ParseUint(c.MongoMinPoolSize, 10, 64)
	if err != nil && c.MongoMinPoolSize != "" {
		errs = append(errs, "MONGO_MIN_POOL_SIZE must be a valid integer")
		return errs
	}

	if maxPool > mongoMaxPoolSizeUpperBound {
		errs = append(errs, "MONGO_MAX_POOL_SIZE must not exceed 10000")
	}

	if maxPool > 0 && minPool > maxPool {
		errs = append(errs, "MONGO_MIN_POOL_SIZE must not exceed MONGO_MAX_POOL_SIZE")
	}

	return errs
}

// InitServers initiate http and grpc servers.
func InitServers() (*Service, error) {
	cfg := &Config{}
	if err := libCommons.SetConfigFromEnvVars(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config from env vars: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
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

	mongoMaxPoolSize, _ := strconv.ParseUint(cfg.MongoMaxPoolSize, 10, 64)
	if mongoMaxPoolSize == 0 {
		mongoMaxPoolSize = 100
	}

	mongoConnection := &mongoDB.MongoConnection{
		ConnectionStringSource: mongoSource,
		Database:               cfg.MongoDBName,
		Logger:                 logger,
		MaxPoolSize:            mongoMaxPoolSize,
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

	templateMongoDBRepository, err := template.NewTemplateMongoDBRepository(mongoConnection)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize template mongodb repository: %w", err)
	}

	reportMongoDBRepository, err := report.NewReportMongoDBRepository(mongoConnection)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize report mongodb repository: %w", err)
	}

	// Create MongoDB indexes
	logger.Info("Ensuring MongoDB indexes exist for templates and reports...")

	if err = templateMongoDBRepository.EnsureIndexes(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure template indexes: %w", err)
	}

	if err = reportMongoDBRepository.EnsureIndexes(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure report indexes: %w", err)
	}

	templateSeaweedFSRepository := templateSeaweedFS.NewStorageRepository(templateStorageClient)
	reportSeaweedFSRepository := reportSeaweedFS.NewStorageRepository(reportStorageClient)

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

	redisConsumerRepository, err := redis.NewConsumerRedis(redisConnection)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize redis connection: %w", err)
	}

	templateService := &services.UseCase{
		TemplateRepo:        templateMongoDBRepository,
		TemplateSeaweedFS:   templateSeaweedFSRepository,
		ExternalDataSources: pkg.NewSafeDataSources(pkg.ExternalDatasourceConnections(logger)),
	}

	authClient := middleware.NewAuthClient(cfg.AuthAddress, cfg.AuthEnabled, &logger)

	templateHandler, err := in2.NewTemplateHandler(templateService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize template handler: %w", err)
	}

	producerRabbitMQRepository := rabbitmq.NewProducerRabbitMQ(rabbitMQConnection)

	// Initialize datasources in lazy mode (connect on-demand for faster startup)
	externalDataSources := pkg.NewSafeDataSources(pkg.ExternalDatasourceConnectionsLazy(logger))

	reportService := &services.UseCase{
		ReportRepo:                reportMongoDBRepository,
		RabbitMQRepo:              producerRabbitMQRepository,
		TemplateRepo:              templateMongoDBRepository,
		ReportSeaweedFS:           reportSeaweedFSRepository,
		ExternalDataSources:       externalDataSources,
		RabbitMQExchange:          cfg.RabbitMQExchange,
		RabbitMQGenerateReportKey: cfg.RabbitMQGenerateReportKey,
	}

	reportHandler, err := in2.NewReportHandler(reportService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize report handler: %w", err)
	}

	dataSourceService := &services.UseCase{
		ExternalDataSources: externalDataSources,
		RedisRepo:           redisConsumerRepository,
	}

	dataSourceHandler, err := in2.NewDataSourceHandler(dataSourceService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize data source handler: %w", err)
	}

	readinessDeps := &in2.ReadinessDeps{
		MongoConnection:    mongoConnection,
		RabbitMQConnection: rabbitMQConnection,
		RedisConnection:    redisConnection,
		StorageClient:      storageClient,
	}

	httpApp := in2.NewRoutes(logger, telemetry, templateHandler, reportHandler, dataSourceHandler, authClient, readinessDeps)
	serverAPI := NewServer(cfg, httpApp, logger, telemetry)

	return &Service{
		Server: serverAPI,
		Logger: logger,
	}, nil
}
