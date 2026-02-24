// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	httpIn "github.com/LerianStudio/reporter/components/manager/internal/adapters/http/in"
	"github.com/LerianStudio/reporter/components/manager/internal/services"
	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	reportSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/report"
	templateSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/template"

	"github.com/LerianStudio/lib-auth/v2/auth/middleware"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libRedis "github.com/LerianStudio/lib-commons/v2/commons/redis"
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
	MongoURI             string `env:"MONGO_URI"`
	MongoDBHost          string `env:"MONGO_HOST"`
	MongoDBName          string `env:"MONGO_NAME"`
	MongoDBUser          string `env:"MONGO_USER"`
	MongoDBPassword      string `env:"MONGO_PASSWORD"`
	MongoDBPort          string `env:"MONGO_PORT"`
	MongoDBParameters    string `env:"MONGO_PARAMETERS"`
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
	// CORS configuration envs
	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS"`
	CORSAllowedMethods string `env:"CORS_ALLOWED_METHODS"`
	CORSAllowedHeaders string `env:"CORS_ALLOWED_HEADERS"`
	// Rate limiting configuration envs
	RateLimitEnabled  bool `env:"RATE_LIMIT_ENABLED" default:"true"`
	RateLimitGlobal   int  `env:"RATE_LIMIT_GLOBAL" default:"100"`
	RateLimitExport   int  `env:"RATE_LIMIT_EXPORT" default:"10"`
	RateLimitDispatch int  `env:"RATE_LIMIT_DISPATCH" default:"50"`
	RateLimitWindow   int  `env:"RATE_LIMIT_WINDOW_SECONDS" default:"60"`
	// Trusted proxies configuration
	TrustedProxies string `env:"TRUSTED_PROXIES"`
}

// Validate checks that all required configuration fields are present
// and that optional numeric bounds are consistent.
// Returns a descriptive multi-error message listing all violations.
func (c *Config) Validate() error {
	var errs []string

	errs = c.validateRequiredFields(errs)
	errs = c.validateMongoPoolBounds(errs)
	errs = c.validateRateLimitBounds(errs)
	errs = c.validateProductionConfig(errs)

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

	if maxPool > constant.MongoMaxPoolSizeUpperBound {
		errs = append(errs, fmt.Sprintf("MONGO_MAX_POOL_SIZE must not exceed %d", constant.MongoMaxPoolSizeUpperBound))
	}

	if maxPool > 0 && minPool > maxPool {
		errs = append(errs, "MONGO_MIN_POOL_SIZE must not exceed MONGO_MAX_POOL_SIZE")
	}

	return errs
}

// validateRateLimitBounds checks that rate limit values are within allowed ranges.
// Zero or negative values would disable rate limiting, and excessively large values
// indicate misconfiguration. Both conditions are rejected.
func (c *Config) validateRateLimitBounds(errs []string) []string {
	limits := []struct {
		value int
		name  string
		max   int
	}{
		{c.RateLimitGlobal, "RATE_LIMIT_GLOBAL", constant.RateLimitMaxGlobal},
		{c.RateLimitExport, "RATE_LIMIT_EXPORT", constant.RateLimitMaxExport},
		{c.RateLimitDispatch, "RATE_LIMIT_DISPATCH", constant.RateLimitMaxDispatch},
	}

	for _, l := range limits {
		if l.value <= 0 || l.value > l.max {
			errs = append(errs, fmt.Sprintf("%s must be between 1 and %d", l.name, l.max))
		}
	}

	return errs
}

// validateProductionConfig enforces stricter rules when EnvName is "production".
// Telemetry, authentication, and real credentials are required in production.
func (c *Config) validateProductionConfig(errs []string) []string {
	if c.EnvName != "production" {
		return errs
	}

	if !c.EnableTelemetry {
		errs = append(errs, "ENABLE_TELEMETRY must be true in production")
	}

	if !c.AuthEnabled {
		errs = append(errs, "PLUGIN_AUTH_ENABLED must be true in production")
	}

	if !c.RateLimitEnabled {
		errs = append(errs, "RATE_LIMIT_ENABLED must be true in production")
	}

	secrets := []struct {
		value string
		name  string
	}{
		{c.MongoDBPassword, "MONGO_PASSWORD"},
		{c.RabbitMQPass, "RABBITMQ_DEFAULT_PASS"},
		{c.RedisPassword, "REDIS_PASSWORD"},
		{c.ObjectStorageSecretKey, "OBJECT_STORAGE_SECRET_KEY"},
	}

	for _, s := range secrets {
		if s.value == constant.DefaultPasswordPlaceholder {
			errs = append(errs, s.name+" must not use the default placeholder in production")
		}
	}

	errs = c.validateProductionCORS(errs)

	return errs
}

// validateProductionCORS enforces that CORS origins are explicitly configured
// in production. Wildcard (*) origins and empty origins are forbidden.
func (c *Config) validateProductionCORS(errs []string) []string {
	if c.CORSAllowedOrigins == "" {
		errs = append(errs, "CORS_ALLOWED_ORIGINS must not be empty in production")
		return errs
	}

	if strings.Contains(c.CORSAllowedOrigins, "*") {
		errs = append(errs, "CORS_ALLOWED_ORIGINS must not contain wildcard (*) in production")
	}

	origins := strings.Split(c.CORSAllowedOrigins, ",")
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" || origin == "*" {
			continue
		}

		if strings.HasPrefix(origin, "http://") {
			errs = append(errs, "CORS_ALLOWED_ORIGINS must use HTTPS in production (found: "+origin+")")
		}
	}

	return errs
}

// InitServers initiate http and grpc servers.
// Uses a cleanup stack pattern: if any initialization step fails, all previously
// opened connections are closed in reverse order to prevent resource leaks.
func InitServers() (_ *Service, err error) {
	cfg, logger, err := initConfigAndLogger()
	if err != nil {
		return nil, err
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

	// Init OpenTelemetry to control logs and flows
	telemetry, telemetryCleanup, err := initTelemetry(cfg, logger)
	if err != nil {
		return nil, err
	}

	cleanups = append(cleanups, telemetryCleanup)

	// Create single storage client for both templates and reports (using prefixes)
	storageClient, err := initStorage(cfg, logger)
	if err != nil {
		return nil, err
	}

	// Init MongoDB connection and repositories
	mongo, mongoCleanup, err := initMongoDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	cleanups = append(cleanups, mongoCleanup)

	// Init RabbitMQ producer and connection monitor
	rabbit, rabbitCleanups := initRabbitMQ(cfg, logger)
	cleanups = append(cleanups, rabbitCleanups...)

	// Init Redis/Valkey connection
	redisConsumerRepository, redisConnection, redisCleanup, err := initRedis(cfg, logger)
	if err != nil {
		return nil, err
	}

	cleanups = append(cleanups, redisCleanup)

	// Initialize datasources in lazy mode (connect on-demand for faster startup).
	// A single instance is shared across all services that need external data sources.
	externalDataSources := pkg.NewSafeDataSources(pkg.ExternalDatasourceConnectionsLazy(logger))

	// Use same storage client for both templates and reports (repositories handle prefixes)
	templateStorageRepo := templateSeaweedFS.NewStorageRepository(storageClient)
	reportStorageRepo := reportSeaweedFS.NewStorageRepository(storageClient)

	// Build service and handler instances
	templateHandler, err := httpIn.NewTemplateHandler(&services.UseCase{
		TemplateRepo:        mongo.templateRepo,
		TemplateSeaweedFS:   templateStorageRepo,
		ExternalDataSources: externalDataSources,
		RedisRepo:           redisConsumerRepository,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize template handler: %w", err)
	}

	reportHandler, err := httpIn.NewReportHandler(&services.UseCase{
		ReportRepo:                mongo.reportRepo,
		RabbitMQRepo:              rabbit.producer,
		TemplateRepo:              mongo.templateRepo,
		ReportSeaweedFS:           reportStorageRepo,
		ExternalDataSources:       externalDataSources,
		RedisRepo:                 redisConsumerRepository,
		RabbitMQExchange:          cfg.RabbitMQExchange,
		RabbitMQGenerateReportKey: cfg.RabbitMQGenerateReportKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize report handler: %w", err)
	}

	dataSourceHandler, err := httpIn.NewDataSourceHandler(&services.UseCase{
		ExternalDataSources: externalDataSources,
		RedisRepo:           redisConsumerRepository,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize data source handler: %w", err)
	}

	// Build HTTP server with routes, middleware, and readiness probes
	authClient := middleware.NewAuthClient(cfg.AuthAddress, cfg.AuthEnabled, &logger)

	readinessDeps := &httpIn.ReadinessDeps{
		MongoConnection:    mongo.connection,
		RabbitMQConnection: rabbit.connection,
		RedisConnection:    redisConnection,
		StorageClient:      storageClient,
	}

	corsConfig := httpIn.CORSConfig{
		AllowedOrigins: cfg.CORSAllowedOrigins,
		AllowedMethods: cfg.CORSAllowedMethods,
		AllowedHeaders: cfg.CORSAllowedHeaders,
	}

	rateLimitConfig := buildRateLimitConfig(cfg, redisConnection, logger)
	trustedProxies := parseTrustedProxies(cfg.TrustedProxies)

	httpApp := httpIn.NewRoutes(logger, telemetry, templateHandler, reportHandler, dataSourceHandler, authClient, readinessDeps, corsConfig, rateLimitConfig, trustedProxies)
	serverAPI := NewServer(cfg, httpApp, logger, telemetry)

	// Build consolidated shutdown cleanup from the same cleanup stack used for
	// init-failure recovery. Resources are closed in reverse initialization order
	// (Redis -> RabbitMQ -> MongoDB -> Telemetry). Telemetry is flushed last so
	// it captures any shutdown-related spans.
	shutdown := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			func(idx int) {
				defer func() {
					if r := recover(); r != nil {
						logger.Errorf("Cleanup panic at index %d: %v", idx, r)
					}
				}()

				cleanups[idx]()
			}(i)
		}
	}

	return &Service{
		Server:  serverAPI,
		Logger:  logger,
		cleanup: shutdown,
	}, nil
}

// buildRateLimitConfig assembles the rate limit configuration from environment
// settings, using Redis-backed storage when rate limiting is enabled.
func buildRateLimitConfig(cfg *Config, redisConnection *libRedis.RedisConnection, logger log.Logger) httpIn.RateLimitConfig {
	window := constant.RateLimitDefaultWindow
	if cfg.RateLimitWindow > 0 {
		window = time.Duration(cfg.RateLimitWindow) * time.Second
	}

	var rateLimitStorage httpIn.RateLimitStorage
	if cfg.RateLimitEnabled {
		rateLimitStorage = httpIn.NewRedisStorage(redisConnection, logger)
	}

	return httpIn.RateLimitConfig{
		Enabled:     cfg.RateLimitEnabled,
		GlobalMax:   cfg.RateLimitGlobal,
		ExportMax:   cfg.RateLimitExport,
		DispatchMax: cfg.RateLimitDispatch,
		Window:      window,
		Storage:     rateLimitStorage,
	}
}

// parseTrustedProxies splits the comma-separated trusted proxies string into
// a cleaned slice, omitting empty entries.
func parseTrustedProxies(raw string) []string {
	if raw == "" {
		return nil
	}

	var proxies []string

	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			proxies = append(proxies, p)
		}
	}

	return proxies
}
