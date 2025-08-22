package pkg

import (
	"fmt"
	"net/url"
	"os"
	"plugin-smart-templates/v2/pkg/mongodb"
	pg "plugin-smart-templates/v2/pkg/postgres"
	"strings"

	"context"
	"time"

	"github.com/LerianStudio/lib-commons/v2/commons/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DataSourceConfig represents the configuration required to establish a connection to a data source.
// Fields include name, connection details, authentication, database, type, and SSL mode.
type DataSourceConfig struct {
	ConfigName  string
	Name        string
	Host        string
	Port        string
	User        string
	Password    string
	Database    string
	Type        string
	SSLMode     string
	SSLCert     string
	SSLRootCert string
	SSL         string
	SSLCA       string
	Options     string
}

// DataSource represents a configuration for an external data source, specifying the database type and repository used.
type DataSource struct {
	// DatabaseType specifies the type of database being used, such as "postgresql" or "mongodb".
	DatabaseType string

	// PostgresRepository is an interface for querying PostgreSQL tables and fields in an external data source.
	PostgresRepository pg.Repository

	// MongoDBRepository is an interface for querying MongoDB collections and fields in an external data source.
	MongoDBRepository mongodb.Repository

	// DatabaseConfig holds the configuration needed to establish a connection
	DatabaseConfig *pg.Connection

	// MongoURI holds the MongoDB connection string
	MongoURI string

	// MongoDBName holds the MongoDB database name
	MongoDBName string

	// Connection holds the actual database connection that can be closed
	Connection *pg.Connection

	// Initialized indicates if the connection has been established
	Initialized bool
}

// ConnectToDataSource establishes a connection to a data source if not already initialized.
func ConnectToDataSource(databaseName string, dataSource *DataSource, logger log.Logger, externalDataSources map[string]DataSource) error {
	switch dataSource.DatabaseType {
	case PostgreSQLType:
		dataSource.PostgresRepository = pg.NewDataSourceRepository(dataSource.DatabaseConfig)

		logger.Infof("Established PostgreSQL connection to %s database", databaseName)

	case MongoDBType:
		dataSource.MongoDBRepository = mongodb.NewDataSourceRepository(dataSource.MongoURI, dataSource.MongoDBName, logger)

		logger.Infof("Established MongoDB connection to %s database", databaseName)

	default:
		return fmt.Errorf("unsupported database type: %s for database: %s", dataSource.DatabaseType, databaseName)
	}

	dataSource.Initialized = true
	externalDataSources[databaseName] = *dataSource

	return nil
}

// ExternalDatasourceConnections initializes and returns a map of external data source connections.
func ExternalDatasourceConnections(logger log.Logger) map[string]DataSource {
	externalDataSources := make(map[string]DataSource)

	dataSourceConfigs := getDataSourceConfigs(logger)
	for _, dataSource := range dataSourceConfigs {
		switch strings.ToLower(dataSource.Type) {
		case MongoDBType:
			ds := initMongoDataSource(dataSource, logger)
			externalDataSources[dataSource.ConfigName] = ds
		case PostgreSQLType:
			ds := initPostgresDataSource(dataSource, logger)
			externalDataSources[dataSource.ConfigName] = ds
		default:
			logger.Errorf("Unsupported database type '%s' for data source '%s'.", dataSource.Type, dataSource.Name)
			continue
		}

		logger.Infof("Configured external data source: %s with config name: %s", dataSource.Name, dataSource.ConfigName)
	}

	return externalDataSources
}

func initMongoDataSource(dataSource DataSourceConfig, logger log.Logger) DataSource {
	mongoURI := fmt.Sprintf("%s://%s:%s@%s:%s/%s",
		dataSource.Type, dataSource.User, dataSource.Password, dataSource.Host, dataSource.Port, dataSource.Database)
	if dataSource.Options != "" {
		mongoURI += "?" + dataSource.Options
	}

	var params []string
	if dataSource.SSL == "true" {
		params = append(params, "ssl=true")
	}

	if dataSource.SSLCA != "" {
		params = append(params, "tlsCAFile="+url.QueryEscape(dataSource.SSLCA))
	}

	if len(params) > 0 {
		if strings.Contains(mongoURI, "?") {
			mongoURI += "&" + strings.Join(params, "&")
		} else {
			mongoURI += "?" + strings.Join(params, "&")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		logger.Errorf("Failed to connect to MongoDB [%s]: %v", dataSource.ConfigName, err)
	} else if err := client.Ping(ctx, nil); err != nil {
		logger.Errorf("Failed to ping MongoDB [%s]: %v", dataSource.ConfigName, err)
	} else {
		logger.Infof("Successfully connected to MongoDB [%s]", dataSource.ConfigName)
	}

	_ = client.Disconnect(ctx)

	return DataSource{
		DatabaseType: MongoDBType,
		MongoURI:     mongoURI,
		MongoDBName:  dataSource.Database,
		Initialized:  false,
	}
}

func initPostgresDataSource(dataSource DataSourceConfig, logger log.Logger) DataSource {
	connectionString := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		dataSource.Type, dataSource.User, url.QueryEscape(dataSource.Password), dataSource.Host, dataSource.Port, dataSource.Database, dataSource.SSLMode)
	if dataSource.SSLMode != "" {
		connectionString += fmt.Sprintf("&sslrootcert=%s", url.QueryEscape(dataSource.SSLRootCert))
	}

	connection := &pg.Connection{
		ConnectionString:   connectionString,
		DBName:             dataSource.Database,
		Logger:             logger,
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
	}
	if err := connection.Connect(); err != nil {
		logger.Errorf("Failed to connect to Postgres [%s]: %v", dataSource.ConfigName, err)
	}

	return DataSource{
		DatabaseType:   dataSource.Type,
		DatabaseConfig: connection,
		Initialized:    false,
	}
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
	prefix := "DATASOURCE_"
	suffix := "_CONFIG_NAME"

	envVars := os.Environ()

	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]

		if strings.HasPrefix(key, prefix) && strings.HasSuffix(key, suffix) {
			remaining := key[len(prefix) : len(key)-len(suffix)]
			dataSourceNamesMap[strings.ToLower(remaining)] = true
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
		"CONFIG_NAME": os.Getenv(fmt.Sprintf("%s%s_CONFIG_NAME", prefixPattern, upperName)),
		"HOST":        os.Getenv(fmt.Sprintf("%s%s_HOST", prefixPattern, upperName)),
		"PORT":        os.Getenv(fmt.Sprintf("%s%s_PORT", prefixPattern, upperName)),
		"USER":        os.Getenv(fmt.Sprintf("%s%s_USER", prefixPattern, upperName)),
		"PASSWORD":    os.Getenv(fmt.Sprintf("%s%s_PASSWORD", prefixPattern, upperName)),
		"DATABASE":    os.Getenv(fmt.Sprintf("%s%s_DATABASE", prefixPattern, upperName)),
		"TYPE":        os.Getenv(fmt.Sprintf("%s%s_TYPE", prefixPattern, upperName)),
		"SSLMODE":     os.Getenv(fmt.Sprintf("%s%s_SSLMODE", prefixPattern, upperName)),
		"SSLROOTCERT": os.Getenv(fmt.Sprintf("%s%s_SSLROOTCERT", prefixPattern, upperName)),
		"SSL":         os.Getenv(fmt.Sprintf("%s%s_SSL", prefixPattern, upperName)),     // For MongoDB SSL
		"SSLCA":       os.Getenv(fmt.Sprintf("%s%s_SSLCA", prefixPattern, upperName)),   // For MongoDB CA file
		"OPTIONS":     os.Getenv(fmt.Sprintf("%s%s_OPTIONS", prefixPattern, upperName)), // For MongoDB URI options
	}

	dataSource := DataSourceConfig{
		Name:        name,
		ConfigName:  configFields["CONFIG_NAME"],
		Host:        configFields["HOST"],
		Port:        configFields["PORT"],
		User:        configFields["USER"],
		Password:    configFields["PASSWORD"],
		Database:    configFields["DATABASE"],
		Type:        configFields["TYPE"],
		SSLMode:     configFields["SSLMODE"],
		SSLRootCert: configFields["SSLROOTCERT"],
		SSL:         configFields["SSL"],
		SSLCA:       configFields["SSLCA"],
		Options:     configFields["OPTIONS"],
	}

	logger.Infof("Found external data source: %s (config name: %s) with database: %s (type: %s, sslmode: %s, ssl: %s, sslca: %s)",
		name, dataSource.ConfigName, dataSource.Database, dataSource.Type, dataSource.SSLMode, dataSource.SSL, dataSource.SSLCA)

	return dataSource, true
}
