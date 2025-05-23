package bootstrap

import (
	"context"
	"fmt"
	"strings"
	libCommons "github.com/LerianStudio/lib-commons/commons"
	mongoDB "github.com/LerianStudio/lib-commons/commons/mongo"
	libOtel "github.com/LerianStudio/lib-commons/commons/opentelemetry"
	libRabbitMQ "github.com/LerianStudio/lib-commons/commons/rabbitmq"
	libZap "github.com/LerianStudio/lib-commons/commons/zap"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"net/url"
	"plugin-smart-templates/components/worker/internal/adapters/rabbitmq"
	"plugin-smart-templates/components/worker/internal/services"
	"plugin-smart-templates/pkg"
	reportFile "plugin-smart-templates/pkg/minio/report"
	templateFile "plugin-smart-templates/pkg/minio/template"
	reportData "plugin-smart-templates/pkg/mongodb/report"
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

	// Log detalhado das configurações do MinIO antes de tentar conectar
	logger.Infof("Tentando conectar ao MinIO no endpoint: %s", minioEndpoint)
	logger.Infof("MinIO SSL habilitado: %v", cfg.MinioSSLEnabled)
	logger.Infof("MinIO usuário: %s", cfg.MinioAppUsername)
	logger.Infof("MinIO senha: ****** (comprimento: %d caracteres)", len(cfg.MinioAppPassword))

	// Tentativa de conexão com o MinIO
	logger.Info("Iniciando conexão com MinIO...")
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAppUsername, cfg.MinioAppPassword, ""),
		Secure: cfg.MinioSSLEnabled,
	})
	if err != nil {
		// Log detalhado do erro
		logger.Errorf("Falha na conexão com MinIO: %v", err)
		logger.Errorf("Tipo do erro: %T", err)
		
		// Verificar tipos de erro específicos
		if strings.Contains(err.Error(), "authentication") || strings.Contains(err.Error(), "credential") {
			logger.Error("Possível problema de autenticação ou credenciais inválidas")
		} else if strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "dial") {
			logger.Error("Possível problema de conectividade ou endpoint inválido")
		}
		
		logger.Fatalf("Erro fatal ao criar cliente MinIO: %v", err)
	}
	
	// Testar conexão com MinIO
	logger.Info("Cliente MinIO criado, testando conexão...")
	// Listar buckets para verificar se a conexão está funcionando
	buckets, listErr := minioClient.ListBuckets(context.Background())
	if listErr != nil {
		logger.Errorf("Erro ao listar buckets do MinIO: %v", listErr)
		logger.Fatalf("Falha ao verificar conexão com MinIO: %v", listErr)
	}
	
	// Log de sucesso com detalhes dos buckets
	logger.Infof("Conexão com MinIO estabelecida com sucesso ✅")
	logger.Infof("Buckets disponíveis: %d", len(buckets))
	for i, bucket := range buckets {
		logger.Infof("Bucket %d: %s (criado em: %s)", i+1, bucket.Name, bucket.CreationDate)
	}

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

	service := &services.UseCase{
		TemplateFileRepo:    templateFile.NewMinioRepository(minioClient, "templates"),
		ReportFileRepo:      reportFile.NewMinioRepository(minioClient, "reports"),
		ExternalDataSources: pkg.ExternalDatasourceConnections(logger),
		ReportDataRepo:      reportData.NewReportMongoDBRepository(mongoConnection),
	}

	multiQueueConsumer := NewMultiQueueConsumer(routes, service)

	return &Service{
		MultiQueueConsumer: multiQueueConsumer,
		Logger:             logger,
	}
}
