package bootstrap

import (
	"fmt"
	grcpcin "k8s-golang-addons-boilerplate/internal/adapters/grpc/in"
	"k8s-golang-addons-boilerplate/internal/adapters/http/in"
	"k8s-golang-addons-boilerplate/internal/adapters/postgres/example"
	command "k8s-golang-addons-boilerplate/internal/services/example/command"
	"k8s-golang-addons-boilerplate/internal/services/example/query"
	"k8s-golang-addons-boilerplate/pkg"
	"k8s-golang-addons-boilerplate/pkg/opentelemetry"
	"k8s-golang-addons-boilerplate/pkg/postgres"
	"k8s-golang-addons-boilerplate/pkg/zap"
)

const ApplicationName = "example-boilerplate"

// Config is the top level configuration struct for the entire application.
type Config struct {
	EnvName                 string `env:"ENV_NAME"`
	ProtoAddress            string `env:"PROTO_ADDRESS"`
	ServerAddress           string `env:"SERVER_ADDRESS"`
	LogLevel                string `env:"LOG_LEVEL"`
	OtelServiceName         string `env:"OTEL_RESOURCE_SERVICE_NAME"`
	OtelLibraryName         string `env:"OTEL_LIBRARY_NAME"`
	OtelServiceVersion      string `env:"OTEL_RESOURCE_SERVICE_VERSION"`
	OtelDeploymentEnv       string `env:"OTEL_RESOURCE_DEPLOYMENT_ENVIRONMENT"`
	OtelColExporterEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	PrimaryDBHost           string `env:"DB_HOST"`
	PrimaryDBUser           string `env:"DB_USER"`
	PrimaryDBPassword       string `env:"DB_PASSWORD"`
	PrimaryDBName           string `env:"DB_NAME"`
	PrimaryDBPort           string `env:"DB_PORT"`
	ReplicaDBHost           string `env:"DB_REPLICA_HOST"`
	ReplicaDBUser           string `env:"DB_REPLICA_USER"`
	ReplicaDBPassword       string `env:"DB_REPLICA_PASSWORD"`
	ReplicaDBName           string `env:"DB_REPLICA_NAME"`
	ReplicaDBPort           string `env:"DB_REPLICA_PORT"`
	RedisHost               string `env:"REDIS_HOST"`
	RedisPort               string `env:"REDIS_PORT"`
	RedisUser               string `env:"REDIS_USER"`
	RedisPassword           string `env:"REDIS_PASSWORD"`
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

	// Init database connection
	postgresSQLSourcePrimary := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.PrimaryDBHost, cfg.PrimaryDBUser, cfg.PrimaryDBPassword, cfg.PrimaryDBName, cfg.PrimaryDBPort)

	postgresSQLSourceReplica := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.ReplicaDBHost, cfg.ReplicaDBUser, cfg.ReplicaDBPassword, cfg.ReplicaDBName, cfg.ReplicaDBPort)

	postgresConnection := &postgres.PostgresConnection{
		ConnectionStringPrimary: postgresSQLSourcePrimary,
		ConnectionStringReplica: postgresSQLSourceReplica,
		PrimaryDBName:           cfg.PrimaryDBName,
		ReplicaDBName:           cfg.ReplicaDBName,
		Component:               ApplicationName,
		Logger:                  logger,
	}

	/* Init Redis Cache

	redisSource := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)

	redisConnection := &mredis.RedisConnection{
		Addr:     redisSource,
		User:     cfg.RedisUser,
		Password: cfg.RedisPassword,
		DB:       0,
		Protocol: 3,
		Logger:   logger,
	}

	redisConsumerRepository := redis.NewConsumerRedis(redisConnection)

	*/

	examplePostgreSQLRepository := example.NewExamplePostgresSQLRepository(postgresConnection)

	exampleCommand := &command.ExampleCommand{
		ExampleRepo: examplePostgreSQLRepository,
	}

	exampleQuery := &query.ExampleQuery{
		ExampleRepo: examplePostgreSQLRepository,
	}

	exampleHandler := &in.ExampleHandler{
		ExampleCommand: exampleCommand,
		ExampleQuery:   exampleQuery,
	}

	httpApp := in.NewRoutes(logger, telemetry, exampleHandler)
	serverAPI := NewServer(cfg, httpApp, logger, telemetry)

	grpcApp := grcpcin.NewRouterGRPC(logger, telemetry, exampleQuery, exampleCommand)
	serverGRPC := NewServerGRPC(cfg, grpcApp, logger, telemetry)

	return &Service{
		Server:     serverAPI,
		ServerGRPC: serverGRPC,
		Logger:     logger,
	}
}
