package bootstrap

import (
	"fmt"
	libCommons "github.com/LerianStudio/lib-commons/commons"
	mongoDB "github.com/LerianStudio/lib-commons/commons/mongo"
	libOtel "github.com/LerianStudio/lib-commons/commons/opentelemetry"
	libRabbitMQ "github.com/LerianStudio/lib-commons/commons/rabbitmq"
	libZap "github.com/LerianStudio/lib-commons/commons/zap"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	reportFile "plugin-template-engine/components/worker/internal/adapters/minio/report"
	templateFile "plugin-template-engine/components/worker/internal/adapters/minio/template"
	reportData "plugin-template-engine/components/worker/internal/adapters/mongodb/report"
	"plugin-template-engine/components/worker/internal/adapters/postgres"
	"plugin-template-engine/components/worker/internal/adapters/rabbitmq"
	"plugin-template-engine/components/worker/internal/services"
	pg "plugin-template-engine/pkg/postgres"
)

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
	// Midaz
	MidazDBHost            string `env:"MIDAZ_DB_HOST"`
	MidazDBPort            string `env:"MIDAZ_DB_PORT"`
	MidazDBUser            string `env:"MIDAZ_DB_USER"`
	MidazDBPass            string `env:"MIDAZ_DB_PASSWORD"`
	MidazOnboardingDBName  string `env:"MIDAZ_ONBOARDING_DB_NAME"`
	MidazTransactionDBName string `env:"MIDAZ_TRANSACTION_DB_NAME"`
}

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

	// TODO: on demand database connection

	externalDataSources := make(map[string]services.DataSource)

	// Midaz Onboarding
	onboardingString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.MidazDBUser, cfg.MidazDBPass, cfg.MidazDBHost, cfg.MidazDBPort, cfg.MidazOnboardingDBName)

	onboardingConnection := &pg.Connection{
		ConnectionString:   onboardingString,
		DBName:             cfg.MidazOnboardingDBName,
		Logger:             logger,
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
	}

	onboardingRepository := postgres.NewRepository(onboardingConnection)

	externalDataSources["onboarding"] = services.DataSource{
		DatabaseType:       "postgres",
		PostgresRepository: onboardingRepository,
	}

	// Midaz Transactions
	transactionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.MidazDBUser, cfg.MidazDBPass, cfg.MidazDBHost, cfg.MidazDBPort, cfg.MidazTransactionDBName)

	transactionConnection := &pg.Connection{
		ConnectionString:   transactionString,
		DBName:             cfg.MidazTransactionDBName,
		Logger:             logger,
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
	}

	transactionRepository := postgres.NewRepository(transactionConnection)

	externalDataSources["transaction"] = services.DataSource{
		DatabaseType:       "postgres",
		PostgresRepository: transactionRepository,
	}

	service := &services.UseCase{
		TemplateFileRepo:    templateFile.NewMinioRepository(minioClient, "templates"),
		ReportFileRepo:      reportFile.NewMinioRepository(minioClient, "reports"),
		ExternalDataSources: externalDataSources,
		ReportDataRepo:      reportData.NewReportMongoDBRepository(mongoConnection),
	}

	multiQueueConsumer := NewMultiQueueConsumer(routes, service)

	return &Service{
		MultiQueueConsumer: multiQueueConsumer,
		Logger:             logger,
	}
}
