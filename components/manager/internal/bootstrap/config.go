package bootstrap

import (
	"fmt"
	mongoDB "github.com/LerianStudio/lib-commons/commons/mongo"
	libOtel "github.com/LerianStudio/lib-commons/commons/opentelemetry"
	libRabbitmq "github.com/LerianStudio/lib-commons/commons/rabbitmq"
	"github.com/LerianStudio/lib-commons/commons/zap"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	in2 "plugin-template-engine/components/manager/internal/adapters/http/in"
	"plugin-template-engine/components/manager/internal/adapters/rabbitmq"
	"plugin-template-engine/components/manager/internal/services"
	"plugin-template-engine/pkg"
	templateMinio "plugin-template-engine/pkg/minio/template"
	"plugin-template-engine/pkg/mongodb/report"
	"plugin-template-engine/pkg/mongodb/template"
)

// Config is the top level configuration struct for the entire application.
type Config struct {
	EnvName                     string `env:"ENV_NAME"`
	ServerAddress               string `env:"SERVER_ADDRESS"`
	LogLevel                    string `env:"LOG_LEVEL"`
	OtelServiceName             string `env:"OTEL_RESOURCE_SERVICE_NAME"`
	OtelLibraryName             string `env:"OTEL_LIBRARY_NAME"`
	OtelServiceVersion          string `env:"OTEL_RESOURCE_SERVICE_VERSION"`
	OtelDeploymentEnv           string `env:"OTEL_RESOURCE_DEPLOYMENT_ENVIRONMENT"`
	OtelColExporterEndpoint     string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	EnableTelemetry             bool   `env:"ENABLE_TELEMETRY"`
	MongoURI                    string `env:"MONGO_URI"`
	MongoDBHost                 string `env:"MONGO_HOST"`
	MongoDBName                 string `env:"MONGO_NAME"`
	MongoDBUser                 string `env:"MONGO_USER"`
	MongoDBPassword             string `env:"MONGO_PASSWORD"`
	MongoDBPort                 string `env:"MONGO_PORT"`
	MinioAPIHost                string `env:"MINIO_API_HOST"`
	MinioAPIPort                string `env:"MINIO_API_PORT"`
	MinioSSLEnabled             bool   `env:"MINIO_SSL_ENABLED"`
	MinioAppUsername            string `env:"MINIO_APP_USER"`
	MinioAppPassword            string `env:"MINIO_APP_PASSWORD"`
	RabbitURI                   string `env:"RABBITMQ_URI"`
	RabbitMQHost                string `env:"RABBITMQ_HOST"`
	RabbitMQPortHost            string `env:"RABBITMQ_PORT_HOST"`
	RabbitMQPortAMQP            string `env:"RABBITMQ_PORT_AMQP"`
	RabbitMQUser                string `env:"RABBITMQ_DEFAULT_USER"`
	RabbitMQPass                string `env:"RABBITMQ_DEFAULT_PASS"`
	RabbitMQGenerateReportQueue string `env:"RABBITMQ_GENERATE_REPORT_QUEUE"`
}

// InitServers initiate http and grpc servers.
func InitServers() *Service {
	cfg := &Config{}

	if err := pkg.SetConfigFromEnvVars(cfg); err != nil {
		panic(err)
	}

	logger := zap.InitializeLogger()

	// Init Open telemetry to control logs and flows
	telemetry := &libOtel.Telemetry{
		LibraryName:               cfg.OtelLibraryName,
		ServiceName:               cfg.OtelServiceName,
		ServiceVersion:            cfg.OtelServiceVersion,
		DeploymentEnv:             cfg.OtelDeploymentEnv,
		CollectorExporterEndpoint: cfg.OtelColExporterEndpoint,
		EnableTelemetry:           cfg.EnableTelemetry,
	}

	// Config minio connection
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

	mongoConnection := &mongoDB.MongoConnection{
		ConnectionStringSource: mongoSource,
		Database:               cfg.MongoDBName,
		Logger:                 logger,
	}

	// Init rabbit MQ for producer
	rabbitSource := fmt.Sprintf("%s://%s:%s@%s:%s",
		cfg.RabbitURI, cfg.RabbitMQUser, cfg.RabbitMQPass, cfg.RabbitMQHost, cfg.RabbitMQPortAMQP)

	logger.Infof(rabbitSource)

	rabbitMQConnection := &libRabbitmq.RabbitMQConnection{
		ConnectionStringSource: rabbitSource,
		Host:                   cfg.RabbitMQHost,
		Port:                   cfg.RabbitMQPortHost,
		User:                   cfg.RabbitMQUser,
		Pass:                   cfg.RabbitMQPass,
		Queue:                  cfg.RabbitMQGenerateReportQueue,
		Logger:                 logger,
	}

	templateMongoDBRepository := template.NewTemplateMongoDBRepository(mongoConnection)
	reportMongoDBRepository := report.NewReportMongoDBRepository(mongoConnection)

	templateService := &services.UseCase{
		TemplateRepo:  templateMongoDBRepository,
		TemplateMinio: templateMinio.NewMinioRepository(minioClient, "templates"),
	}

	templateHandler := &in2.TemplateHandler{
		Service: templateService,
	}

	producerRabbitMQRepository := rabbitmq.NewProducerRabbitMQ(rabbitMQConnection)
	reportService := &services.UseCase{
		ReportRepo:   reportMongoDBRepository,
		RabbitMQRepo: producerRabbitMQRepository,
		TemplateRepo: templateMongoDBRepository,
	}

	reportHandler := &in2.ReportHandler{
		Service: reportService,
	}

	httpApp := in2.NewRoutes(logger, telemetry, templateHandler, reportHandler)
	serverAPI := NewServer(cfg, httpApp, logger, telemetry)

	return &Service{
		Server: serverAPI,
		Logger: logger,
	}
}
