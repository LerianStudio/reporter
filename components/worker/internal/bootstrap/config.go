package bootstrap

import (
	"fmt"
	"os"
	"plugin-template-engine/pkg"
	"strings"

	libCommons "github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/log"
	mongoDB "github.com/LerianStudio/lib-commons/commons/mongo"
	libOtel "github.com/LerianStudio/lib-commons/commons/opentelemetry"
	libRabbitMQ "github.com/LerianStudio/lib-commons/commons/rabbitmq"
	libZap "github.com/LerianStudio/lib-commons/commons/zap"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"plugin-template-engine/components/worker/internal/adapters/rabbitmq"
	"plugin-template-engine/components/worker/internal/services"
	reportFile "plugin-template-engine/pkg/minio/report"
	templateFile "plugin-template-engine/pkg/minio/template"
	reportData "plugin-template-engine/pkg/mongodb/report"
	pg "plugin-template-engine/pkg/postgres"
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
}

// DataSourceConfig represents the configuration required to establish a connection to a data source.
// Fields include name, connection details, authentication, database, type, and SSL mode.
type DataSourceConfig struct {
	Name     string
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Type     string
	SSLMode  string
}

// InitWorker initializes and configures the application's dependencies and returns the Service instance.
func InitWorker() *Service {
	cfg := &Config{}

	if err := libCommons.SetConfigFromEnvVars(cfg); err != nil {
		panic(err)
	}

	logger := libZap.InitializeLogger()

	telemetry := &libOtel.Telemetry{
		LibraryName:               cfg.OtelLibraryName,
		ServiceName:               cfg.OtelServiceName,
		ServiceVersion:            cfg.OtelServiceVersion,
		DeploymentEnv:             cfg.OtelDeploymentEnv,
		CollectorExporterEndpoint: cfg.OtelColExporterEndpoint,
		EnableTelemetry:           cfg.EnableTelemetry,
	}

	rabbitSource := fmt.Sprintf("%s://%s:%s@%s:%s",
		cfg.RabbitURI, cfg.RabbitMQUser, cfg.RabbitMQPass, cfg.RabbitMQHost, cfg.RabbitMQPortAMQP)

	logger.Infof(rabbitSource)

	rabbitMQConnection := &libRabbitMQ.RabbitMQConnection{
		ConnectionStringSource: rabbitSource,
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
	mongoSource := fmt.Sprintf("%s://%s:%s@%s:%s",
		cfg.MongoURI, cfg.MongoDBUser, cfg.MongoDBPassword, cfg.MongoDBHost, cfg.MongoDBPort)

	if cfg.MaxPoolSize <= 0 {
		cfg.MaxPoolSize = 100
	}

	mongoConnection := &mongoDB.MongoConnection{
		ConnectionStringSource: mongoSource,
		Database:               cfg.MongoDBName,
		Logger:                 logger,
		MaxPoolSize:            uint64(cfg.MaxPoolSize),
	}

	service := &services.UseCase{
		TemplateFileRepo:    templateFile.NewMinioRepository(minioClient, "templates"),
		ReportFileRepo:      reportFile.NewMinioRepository(minioClient, "reports"),
		ExternalDataSources: externalDatasourceConnections(logger),
		ReportDataRepo:      reportData.NewReportMongoDBRepository(mongoConnection),
	}

	multiQueueConsumer := NewMultiQueueConsumer(routes, service)

	return &Service{
		MultiQueueConsumer: multiQueueConsumer,
		Logger:             logger,
	}
}

// externalDatasourceConnections initializes and returns a map of external data source connections.
func externalDatasourceConnections(logger log.Logger) map[string]services.DataSource {
	externalDataSources := make(map[string]services.DataSource)

	dataSourceConfigs := getDataSourceConfigs(logger)

	for _, dataSource := range dataSourceConfigs {
		if !pkg.IsSupportedDatabaseType(strings.ToLower(dataSource.Type)) {
			logger.Warnf("Unsupported database type '%s' for data source '%s'. Only PostgreSQL is currently supported. Skipping this data source.",
				dataSource.Type, dataSource.Name)
			continue
		}

		connectionString := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
			dataSource.Type, dataSource.User, dataSource.Password, dataSource.Host, dataSource.Port, dataSource.Database, dataSource.SSLMode)

		connection := &pg.Connection{
			ConnectionString:   connectionString,
			DBName:             dataSource.Database,
			Logger:             logger,
			MaxOpenConnections: 10,
			MaxIdleConnections: 5,
		}

		externalDataSources[dataSource.Name] = services.DataSource{
			DatabaseType:   dataSource.Type,
			DatabaseConfig: connection,
			Initialized:    false,
		}

		logger.Infof("Configured external data source: %s", dataSource.Name)
	}

	return externalDataSources
}

// getDataSourceConfigs retrieves data source configurations from environment variables in the DATASOURCE_[NAME]_* format.
// It validates and returns a slice of DataSourceConfig, logging warnings for incomplete or missing configurations.
func getDataSourceConfigs(logger log.Logger) []DataSourceConfig {
	var dataSources []DataSourceConfig

	dataSourceNames := collectDataSourceNames()

	for name := range dataSourceNames {
		if config, isComplete := buildDataSourceConfig(name, logger); isComplete {
			dataSources = append(dataSources, config)
		}
	}

	if len(dataSources) == 0 {
		logger.Warn("No external data sources found in environment variables. Configure them with DATASOURCE_[NAME]_HOST/PORT/USER/PASSWORD/DATABASE/TYPE/SSLMODE format.")
	}

	return dataSources
}

// collectDataSourceNames identifies all available data source names from environment variables.
func collectDataSourceNames() map[string]bool {
	dataSourceNamesMap := make(map[string]bool)
	prefixPattern := "DATASOURCE_"

	envVars := os.Environ()

	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]

		if strings.HasPrefix(key, prefixPattern) {
			remaining := key[len(prefixPattern):]
			lastUnderscore := strings.LastIndex(remaining, "_")

			if lastUnderscore > 0 {
				name := strings.ToLower(remaining[:lastUnderscore])
				dataSourceNamesMap[name] = true
			}
		}
	}

	return dataSourceNamesMap
}

// buildDataSourceConfig creates a DataSourceConfig for the given name, validating all required fields.
// Returns the config and a boolean indicating if the configuration is complete.
func buildDataSourceConfig(name string, logger log.Logger) (DataSourceConfig, bool) {
	prefixPattern := "DATASOURCE_"
	upperName := strings.ToUpper(name)

	configFields := map[string]string{
		"HOST":     os.Getenv(fmt.Sprintf("%s%s_HOST", prefixPattern, upperName)),
		"PORT":     os.Getenv(fmt.Sprintf("%s%s_PORT", prefixPattern, upperName)),
		"USER":     os.Getenv(fmt.Sprintf("%s%s_USER", prefixPattern, upperName)),
		"PASSWORD": os.Getenv(fmt.Sprintf("%s%s_PASSWORD", prefixPattern, upperName)),
		"DATABASE": os.Getenv(fmt.Sprintf("%s%s_DATABASE", prefixPattern, upperName)),
		"TYPE":     os.Getenv(fmt.Sprintf("%s%s_TYPE", prefixPattern, upperName)),
		"SSLMODE":  os.Getenv(fmt.Sprintf("%s%s_SSLMODE", prefixPattern, upperName)),
	}

	if isConfigComplete(configFields) {
		dataSource := DataSourceConfig{
			Name:     name,
			Host:     configFields["HOST"],
			Port:     configFields["PORT"],
			User:     configFields["USER"],
			Password: configFields["PASSWORD"],
			Database: configFields["DATABASE"],
			Type:     configFields["TYPE"],
			SSLMode:  configFields["SSLMODE"],
		}

		logger.Infof("Found external data source: %s with database: %s (type: %s, sslmode: %s)",
			name, dataSource.Database, dataSource.Type, dataSource.SSLMode)

		return dataSource, true
	}

	logMissingFields(name, configFields, logger)

	return DataSourceConfig{}, false
}

// isConfigComplete checks if all required fields for a data source configuration are present.
func isConfigComplete(fields map[string]string) bool {
	for _, value := range fields {
		if value == "" {
			return false
		}
	}

	return true
}

// logMissingFields logs warnings for any missing configuration fields.
func logMissingFields(name string, fields map[string]string, logger log.Logger) {
	logger.Warnf("Incomplete configuration for data source %s. Skipping.", name)

	for field, value := range fields {
		if value == "" {
			logger.Warnf("Missing %s for data source %s", field, name)
		}
	}
}
