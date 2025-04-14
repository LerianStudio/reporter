package bootstrap

import (
	"fmt"
	in2 "plugin-template-engine/components/manager/internal/adapters/http/in"
	templateDB "plugin-template-engine/components/manager/internal/adapters/mongodb/templates"
	"plugin-template-engine/components/manager/internal/services"
	"plugin-template-engine/pkg"
	mongoDB "plugin-template-engine/pkg/mongo"
	"plugin-template-engine/pkg/opentelemetry"
	"plugin-template-engine/pkg/zap"
)

const ApplicationName = "example-boilerplate"

// Config is the top level configuration struct for the entire application.
type Config struct {
	EnvName                 string `env:"ENV_NAME"`
	ServerAddress           string `env:"SERVER_ADDRESS"`
	LogLevel                string `env:"LOG_LEVEL"`
	OtelServiceName         string `env:"OTEL_RESOURCE_SERVICE_NAME"`
	OtelLibraryName         string `env:"OTEL_LIBRARY_NAME"`
	OtelServiceVersion      string `env:"OTEL_RESOURCE_SERVICE_VERSION"`
	OtelDeploymentEnv       string `env:"OTEL_RESOURCE_DEPLOYMENT_ENVIRONMENT"`
	OtelColExporterEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	MongoURI                string `env:"MONGO_URI"`
	MongoDBHost             string `env:"MONGO_HOST"`
	MongoDBName             string `env:"MONGO_NAME"`
	MongoDBUser             string `env:"MONGO_USER"`
	MongoDBPassword         string `env:"MONGO_PASSWORD"`
	MongoDBPort             string `env:"MONGO_PORT"`
}

// InitServers initiate http and grpc servers.
func InitServers() *Service {
	cfg := &Config{}

	if err := pkg.SetConfigFromEnvVars(cfg); err != nil {
		panic(err)
	}

	logger := zap.InitializeLogger()

	// Init Open telemetry to control logs and flows
	telemetry := &opentelemetry.Telemetry{
		LibraryName:               cfg.OtelLibraryName,
		ServiceName:               cfg.OtelServiceName,
		ServiceVersion:            cfg.OtelServiceVersion,
		DeploymentEnv:             cfg.OtelDeploymentEnv,
		CollectorExporterEndpoint: cfg.OtelColExporterEndpoint,
	}

	/*mongoSource := fmt.Sprintf("%s://%s:%s@%s:%s",
	cfg.MongoURI, cfg.MongoDBUser, cfg.MongoDBPassword, cfg.MongoDBHost, cfg.MongoDBPort)*/
	mongoSource := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=%s",
		cfg.MongoDBUser, cfg.MongoDBPassword, cfg.MongoDBHost, cfg.MongoDBPort, cfg.MongoDBName, cfg.MongoDBName)

	mongoConnection := &mongoDB.MongoConnection{
		ConnectionStringSource: mongoSource,
		Database:               cfg.MongoDBName,
		Logger:                 logger,
	}

	templateMongoDBRepository := templateDB.NewTemplateMongoDBRepository(mongoConnection)

	templateService := &services.UseCase{
		TemplateRepo: templateMongoDBRepository,
	}

	templateHandler := &in2.TemplateHandler{
		Service: templateService,
	}

	httpApp := in2.NewRoutes(logger, telemetry, templateHandler)
	serverAPI := NewServer(cfg, httpApp, logger, telemetry)

	return &Service{
		Server: serverAPI,
		Logger: logger,
	}
}
