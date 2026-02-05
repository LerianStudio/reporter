// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pkg

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb"
	pg "github.com/LerianStudio/reporter/pkg/postgres"

	libConstant "github.com/LerianStudio/lib-commons/v2/commons/constants"
	"github.com/LerianStudio/lib-commons/v2/commons/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// registeredDataSourceIDs holds the immutable set of valid datasource IDs.
// This is populated once at startup and never modified, providing a source of truth
// for validating datasource names and preventing map corruption from invalid IDs.
var (
	registeredDataSourceIDs     = make(map[string]struct{})
	registeredDataSourceIDsOnce sync.Once
	registeredDataSourceIDsLock sync.RWMutex
)

// initRegisteredDataSourceIDs initializes the immutable set of valid datasource IDs.
// This should be called once at startup before any datasource operations.
func initRegisteredDataSourceIDs(ids []string) {
	registeredDataSourceIDsOnce.Do(func() {
		registeredDataSourceIDsLock.Lock()
		defer registeredDataSourceIDsLock.Unlock()

		for _, id := range ids {
			registeredDataSourceIDs[id] = struct{}{}
		}
	})
}

// RegisterDataSourceIDsForTesting allows tests to register datasource IDs.
// This should ONLY be used in tests. In production, IDs are registered at startup.
func RegisterDataSourceIDsForTesting(ids []string) {
	registeredDataSourceIDsLock.Lock()
	defer registeredDataSourceIDsLock.Unlock()

	for _, id := range ids {
		registeredDataSourceIDs[id] = struct{}{}
	}
}

// ResetRegisteredDataSourceIDsForTesting clears all registered IDs and resets the sync.Once.
// This should ONLY be used in tests to ensure test isolation.
func ResetRegisteredDataSourceIDsForTesting() {
	registeredDataSourceIDsLock.Lock()
	defer registeredDataSourceIDsLock.Unlock()

	registeredDataSourceIDs = make(map[string]struct{})
	registeredDataSourceIDsOnce = sync.Once{}
}

// IsValidDataSourceID checks if a datasource ID was registered at startup.
// This is the authoritative check for valid datasource names.
func IsValidDataSourceID(id string) bool {
	registeredDataSourceIDsLock.RLock()
	defer registeredDataSourceIDsLock.RUnlock()

	_, exists := registeredDataSourceIDs[id]

	return exists
}

// DataSourceConfig represents the configuration required to establish a connection to a data source.
// Fields include name, connection details, authentication, database, type, and SSL mode.
type DataSourceConfig struct {
	ConfigName          string
	Name                string
	Host                string
	Port                string
	User                string
	Password            string
	Database            string
	Type                string
	SSLMode             string
	SSLCert             string
	SSLRootCert         string
	SSL                 string
	SSLCA               string
	Options             string
	MidazOrganizationID string // Used for CRM datasources to construct collection names
}

// GetSchemas returns the configured schemas for this datasource.
// It reads from the environment variable DATASOURCE_{NAME}_SCHEMAS.
// If not configured, it defaults to ["public"].
func (c *DataSourceConfig) GetSchemas() []string {
	envKey := "DATASOURCE_" + strings.ToUpper(strings.ReplaceAll(c.ConfigName, "-", "_")) + "_SCHEMAS"
	schemasStr := os.Getenv(envKey)

	if schemasStr == "" {
		return []string{"public"}
	}

	rawSchemas := strings.Split(schemasStr, ",")
	schemas := make([]string, 0, len(rawSchemas))

	for _, s := range rawSchemas {
		s = strings.TrimSpace(s)
		if s != "" {
			schemas = append(schemas, s)
		}
	}

	if len(schemas) == 0 {
		return []string{"public"}
	}

	return schemas
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

	// Status indicates the current health status of the datasource
	Status string

	// LastError stores the most recent error encountered
	LastError error

	// LastAttempt stores the timestamp of the last connection attempt
	LastAttempt time.Time

	// RetryCount tracks how many times we've attempted to connect
	RetryCount int

	// Schemas holds the list of database schemas to query (PostgreSQL only)
	// Defaults to ["public"] if not configured
	Schemas []string

	// MidazOrganizationID holds the Midaz organization ID for CRM datasources
	// Used to construct collection names like "holder_{org_id}"
	MidazOrganizationID string
}

// ConnectToDataSource establishes a connection to a data source if not already initialized.
func ConnectToDataSource(databaseName string, dataSource *DataSource, logger log.Logger, externalDataSources map[string]DataSource) error {
	// Primary validation: check against immutable set of registered IDs (source of truth)
	if !IsValidDataSourceID(databaseName) {
		logger.Errorf("Attempted to connect to unregistered datasource: %s - not in immutable registry, operation rejected", databaseName)
		return fmt.Errorf("cannot connect to unregistered datasource: %s", databaseName)
	}

	// Secondary validation: ensure datasource exists in the runtime map
	if _, exists := externalDataSources[databaseName]; !exists {
		logger.Errorf("Datasource %s is registered but not in runtime map - possible corruption", databaseName)
		return fmt.Errorf("datasource %s not found in runtime map", databaseName)
	}

	dataSource.LastAttempt = time.Now()
	dataSource.RetryCount++

	var err error

	switch dataSource.DatabaseType {
	case PostgreSQLType:
		dataSource.PostgresRepository, err = pg.NewDataSourceRepository(dataSource.DatabaseConfig)
		if err != nil {
			dataSource.Status = libConstant.DataSourceStatusUnavailable
			dataSource.LastError = err
			logger.Errorf("Failed to establish PostgreSQL connection to %s: %v", databaseName, err)

			return fmt.Errorf("failed to establish PostgreSQL connection to %s: %w", databaseName, err)
		}

		logger.Infof("Established PostgreSQL connection to %s database", databaseName)

		dataSource.Status = libConstant.DataSourceStatusAvailable

	case MongoDBType:
		dataSource.MongoDBRepository, err = mongodb.NewDataSourceRepository(dataSource.MongoURI, dataSource.MongoDBName, logger)
		if err != nil {
			dataSource.Status = libConstant.DataSourceStatusUnavailable
			dataSource.LastError = err
			logger.Errorf("Failed to establish MongoDB connection to %s: %v", databaseName, err)

			return fmt.Errorf("failed to establish MongoDB connection to %s: %w", databaseName, err)
		}

		logger.Infof("Established MongoDB connection to %s database", databaseName)

		dataSource.Status = libConstant.DataSourceStatusAvailable

	default:
		dataSource.Status = libConstant.DataSourceStatusUnavailable
		dataSource.LastError = fmt.Errorf("unsupported database type: %s", dataSource.DatabaseType)

		return fmt.Errorf("unsupported database type: %s for database: %s", dataSource.DatabaseType, databaseName)
	}

	dataSource.Initialized = true
	dataSource.LastError = nil
	externalDataSources[databaseName] = *dataSource

	return nil
}

// isFatalError checks if an error is fatal (no point in retrying)
func isFatalError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// DNS/network errors that won't be fixed by retrying
	fatalPatterns := []string{
		"no such host",
		"lookup",
		"server misbehaving",
		"connection refused",
		"unsupported database type",
		"invalid connection string",
		"authentication failed",
		"authorization failed",
		"access denied",
	}

	for _, pattern := range fatalPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// ConnectToDataSourceWithRetry attempts to connect to a datasource with exponential backoff retry logic.
func ConnectToDataSourceWithRetry(databaseName string, dataSource *DataSource, logger log.Logger, externalDataSources map[string]DataSource) error {
	backoff := constant.DataSourceInitialBackoff

	for attempt := 0; attempt <= constant.DataSourceMaxRetries; attempt++ {
		if attempt > 0 {
			logger.Warnf("Retry attempt %d/%d for datasource %s after %v", attempt, constant.DataSourceMaxRetries, databaseName, backoff)
			time.Sleep(backoff)

			// Calculate next backoff (exponential with max cap)
			backoff = time.Duration(float64(backoff) * constant.DataSourceBackoffMultiplier)
			if backoff > constant.DataSourceMaxBackoff {
				backoff = constant.DataSourceMaxBackoff
			}
		}

		err := ConnectToDataSource(databaseName, dataSource, logger, externalDataSources)
		if err == nil {
			logger.Infof("Successfully connected to datasource %s on attempt %d", databaseName, attempt+1)
			return nil
		}

		logger.Errorf("Failed to connect to datasource %s (attempt %d/%d): %v", databaseName, attempt+1, constant.DataSourceMaxRetries+1, err)

		// Check if error is fatal (no point in retrying)
		if isFatalError(err) {
			logger.Warnf("⚠️  Fatal error detected for datasource %s - skipping remaining retries", databaseName)
			break
		}

		// Don't retry on last attempt
		if attempt == constant.DataSourceMaxRetries {
			break
		}
	}

	logger.Errorf("Exhausted all retry attempts for datasource %s - marking as unavailable", databaseName)

	dataSource.Status = libConstant.DataSourceStatusUnavailable
	externalDataSources[databaseName] = *dataSource

	return fmt.Errorf("failed to connect to datasource %s after %d attempts", databaseName, constant.DataSourceMaxRetries+1)
}

// ExternalDatasourceConnectionsLazy initializes datasource configurations WITHOUT attempting connections.
// Useful for components that connect on-demand (like Manager).
func ExternalDatasourceConnectionsLazy(logger log.Logger) map[string]DataSource {
	externalDataSources := make(map[string]DataSource)

	dataSourceConfigs := getDataSourceConfigs(logger)

	// Collect valid IDs and register them in the immutable set
	validIDs := make([]string, 0, len(dataSourceConfigs))
	for _, dataSource := range dataSourceConfigs {
		validIDs = append(validIDs, dataSource.ConfigName)
	}

	initRegisteredDataSourceIDs(validIDs)
	logger.Infof("Registered %d immutable datasource IDs: %v", len(validIDs), validIDs)

	for _, dataSource := range dataSourceConfigs {
		var ds DataSource

		switch strings.ToLower(dataSource.Type) {
		case MongoDBType:
			ds = initMongoDataSource(dataSource, logger)
		case PostgreSQLType:
			ds = initPostgresDataSource(dataSource, logger)
		default:
			logger.Errorf("Unsupported database type '%s' for data source '%s'.", dataSource.Type, dataSource.Name)
			continue
		}

		// Add datasource WITHOUT attempting connection
		externalDataSources[dataSource.ConfigName] = ds
		logger.Infof("Datasource '%s' configured (lazy mode - will connect on first use)", dataSource.ConfigName)
	}

	logger.Infof("Datasource lazy initialization complete: %d configured", len(externalDataSources))

	return externalDataSources
}

// ExternalDatasourceConnections initializes and returns a map of external data source connections.
// Uses graceful degradation - continues initialization even if some datasources fail.
// Attempts connection with retry for each datasource (use for Worker).
func ExternalDatasourceConnections(logger log.Logger) map[string]DataSource {
	externalDataSources := make(map[string]DataSource)

	dataSourceConfigs := getDataSourceConfigs(logger)

	// Collect valid IDs and register them in the immutable set
	validIDs := make([]string, 0, len(dataSourceConfigs))
	for _, dataSource := range dataSourceConfigs {
		validIDs = append(validIDs, dataSource.ConfigName)
	}

	initRegisteredDataSourceIDs(validIDs)
	logger.Infof("Registered %d immutable datasource IDs: %v", len(validIDs), validIDs)

	for _, dataSource := range dataSourceConfigs {
		var ds DataSource

		switch strings.ToLower(dataSource.Type) {
		case MongoDBType:
			ds = initMongoDataSource(dataSource, logger)
		case PostgreSQLType:
			ds = initPostgresDataSource(dataSource, logger)
		default:
			logger.Errorf("Unsupported database type '%s' for data source '%s'.", dataSource.Type, dataSource.Name)
			continue
		}

		externalDataSources[dataSource.ConfigName] = ds

		// Attempt connection with retry
		err := ConnectToDataSourceWithRetry(dataSource.ConfigName, &ds, logger, externalDataSources)
		if err != nil {
			logger.Errorf("Datasource '%s' is UNAVAILABLE - system will continue without it: %v", dataSource.ConfigName, err)
			externalDataSources[dataSource.ConfigName] = ds
		} else {
			logger.Infof("Datasource '%s' initialized successfully", dataSource.ConfigName)
			externalDataSources[dataSource.ConfigName] = ds
		}
	}

	available := 0
	unavailable := 0

	for name, ds := range externalDataSources {
		if ds.Status == libConstant.DataSourceStatusAvailable {
			available++
		} else {
			unavailable++

			logger.Warnf("Datasource '%s' status: %s", name, ds.Status)
		}
	}

	logger.Infof("Datasource initialization complete: %d available, %d unavailable", available, unavailable)

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

	ctx, cancel := context.WithTimeout(context.Background(), constant.ConnectionTimeout)
	defer cancel()

	// Configure MongoDB client with pool settings and shorter timeouts
	clientOpts := options.Client().
		ApplyURI(mongoURI).
		SetMaxPoolSize(constant.MongoDBMaxPoolSize).
		SetMinPoolSize(constant.MongoDBMinPoolSize).
		SetMaxConnIdleTime(constant.MongoDBMaxConnIdleTime).
		SetConnectTimeout(constant.ConnectionTimeout).
		SetServerSelectionTimeout(constant.ConnectionTimeout)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		logger.Errorf("Failed to connect to MongoDB [%s]: %v", dataSource.ConfigName, err)
	} else if err := client.Ping(ctx, nil); err != nil {
		logger.Errorf("Failed to ping MongoDB [%s]: %v", dataSource.ConfigName, err)
	} else {
		logger.Infof("Successfully connected to MongoDB [%s] with pool config (max: %d, min: %d)",
			dataSource.ConfigName, constant.MongoDBMaxPoolSize, constant.MongoDBMinPoolSize)
	}

	// Only disconnect if client was successfully created
	if client != nil {
		_ = client.Disconnect(ctx)
	}

	return DataSource{
		DatabaseType:        MongoDBType,
		MongoURI:            mongoURI,
		MongoDBName:         dataSource.Database,
		Initialized:         false,
		Status:              libConstant.DataSourceStatusUnknown,
		LastAttempt:         time.Time{},
		RetryCount:          0,
		MidazOrganizationID: dataSource.MidazOrganizationID,
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
		MaxOpenConnections: constant.PostgresMaxOpenConns,
		MaxIdleConnections: constant.PostgresMaxIdleConns,
	}
	if err := connection.Connect(); err != nil {
		logger.Errorf("Failed to connect to Postgres [%s]: %v", dataSource.ConfigName, err)
	} else {
		logger.Infof("Successfully connected to Postgres [%s] with pool config (max: %d, idle: %d)",
			dataSource.ConfigName, constant.PostgresMaxOpenConns, constant.PostgresMaxIdleConns)
	}

	return DataSource{
		DatabaseType:   dataSource.Type,
		DatabaseConfig: connection,
		Initialized:    false,
		Status:         libConstant.DataSourceStatusUnknown,
		LastAttempt:    time.Time{},
		RetryCount:     0,
		Schemas:        dataSource.GetSchemas(),
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
		"CONFIG_NAME":           os.Getenv(fmt.Sprintf("%s%s_CONFIG_NAME", prefixPattern, upperName)),
		"HOST":                  os.Getenv(fmt.Sprintf("%s%s_HOST", prefixPattern, upperName)),
		"PORT":                  os.Getenv(fmt.Sprintf("%s%s_PORT", prefixPattern, upperName)),
		"USER":                  os.Getenv(fmt.Sprintf("%s%s_USER", prefixPattern, upperName)),
		"PASSWORD":              os.Getenv(fmt.Sprintf("%s%s_PASSWORD", prefixPattern, upperName)),
		"DATABASE":              os.Getenv(fmt.Sprintf("%s%s_DATABASE", prefixPattern, upperName)),
		"TYPE":                  os.Getenv(fmt.Sprintf("%s%s_TYPE", prefixPattern, upperName)),
		"SSLMODE":               os.Getenv(fmt.Sprintf("%s%s_SSLMODE", prefixPattern, upperName)),
		"SSLROOTCERT":           os.Getenv(fmt.Sprintf("%s%s_SSLROOTCERT", prefixPattern, upperName)),
		"SSL":                   os.Getenv(fmt.Sprintf("%s%s_SSL", prefixPattern, upperName)),                   // For MongoDB SSL
		"SSLCA":                 os.Getenv(fmt.Sprintf("%s%s_SSLCA", prefixPattern, upperName)),                 // For MongoDB CA file
		"OPTIONS":               os.Getenv(fmt.Sprintf("%s%s_OPTIONS", prefixPattern, upperName)),               // For MongoDB URI options
		"MIDAZ_ORGANIZATION_ID": os.Getenv(fmt.Sprintf("%s%s_MIDAZ_ORGANIZATION_ID", prefixPattern, upperName)), // For CRM collection names
	}

	dataSource := DataSourceConfig{
		Name:                name,
		ConfigName:          configFields["CONFIG_NAME"],
		Host:                configFields["HOST"],
		Port:                configFields["PORT"],
		User:                configFields["USER"],
		Password:            configFields["PASSWORD"],
		Database:            configFields["DATABASE"],
		Type:                configFields["TYPE"],
		SSLMode:             configFields["SSLMODE"],
		SSLRootCert:         configFields["SSLROOTCERT"],
		SSL:                 configFields["SSL"],
		SSLCA:               configFields["SSLCA"],
		Options:             configFields["OPTIONS"],
		MidazOrganizationID: configFields["MIDAZ_ORGANIZATION_ID"],
	}

	logger.Infof("Found external data source: %s (config name: %s) with database: %s (type: %s, sslmode: %s, ssl: %s, sslca: %s)",
		name, dataSource.ConfigName, dataSource.Database, dataSource.Type, dataSource.SSLMode, dataSource.SSL, dataSource.SSLCA)

	return dataSource, true
}
