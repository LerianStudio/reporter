package pkg

import (
	"os"
	"testing"
	"time"

	libConstant "github.com/LerianStudio/lib-commons/v2/commons/constants"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/LerianStudio/lib-commons/v2/commons/zap"
)

// mockLogger creates a simple logger for testing
func mockLogger() log.Logger {
	return zap.InitializeLogger()
}

// =============================================================================
// ConnectToDataSource Tests
// =============================================================================

func TestConnectToDataSource_RejectsUnregisteredDatasource(t *testing.T) {
	t.Parallel()

	logger := mockLogger()
	externalDataSources := make(map[string]DataSource)

	// Register only "valid_datasource"
	externalDataSources["valid_datasource"] = DataSource{
		DatabaseType: PostgreSQLType,
		Initialized:  false,
	}

	// Try to connect to an unregistered datasource
	unregisteredDS := &DataSource{
		DatabaseType: PostgreSQLType,
	}

	err := ConnectToDataSource("unregistered_datasource", unregisteredDS, logger, externalDataSources)

	if err == nil {
		t.Error("Expected error when connecting to unregistered datasource, got nil")
	}

	expectedErrMsg := "cannot connect to unregistered datasource: unregistered_datasource"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message %q, got %q", expectedErrMsg, err.Error())
	}

	// Verify the map was NOT corrupted
	if _, exists := externalDataSources["unregistered_datasource"]; exists {
		t.Error("Unregistered datasource should NOT be added to the map")
	}

	// Verify original datasource is still there
	if _, exists := externalDataSources["valid_datasource"]; !exists {
		t.Error("Valid datasource should still exist in the map")
	}
}

func TestConnectToDataSource_RejectsTemplateIdAsDatasourceName(t *testing.T) {
	t.Parallel()

	logger := mockLogger()
	externalDataSources := make(map[string]DataSource)

	// Register valid datasources
	externalDataSources["midaz_onboarding"] = DataSource{
		DatabaseType: PostgreSQLType,
		Initialized:  false,
	}

	// Simulate frontend bug: sending templateId as datasource name
	templateIdAsDatasource := "019abd3d-13c8-7692-8067-a9a9d42d9b41"
	invalidDS := &DataSource{
		DatabaseType: PostgreSQLType,
	}

	err := ConnectToDataSource(templateIdAsDatasource, invalidDS, logger, externalDataSources)

	if err == nil {
		t.Error("Expected error when using templateId as datasource name, got nil")
	}

	// Verify the map was NOT corrupted with the templateId
	if _, exists := externalDataSources[templateIdAsDatasource]; exists {
		t.Errorf("TemplateId %q should NOT be added to the map", templateIdAsDatasource)
	}

	// Verify map size hasn't changed
	if len(externalDataSources) != 1 {
		t.Errorf("Expected map to have 1 entry, got %d", len(externalDataSources))
	}
}

func TestConnectToDataSource_UnsupportedDatabaseType(t *testing.T) {
	t.Parallel()

	logger := mockLogger()
	externalDataSources := make(map[string]DataSource)

	// Register datasource with unsupported type
	externalDataSources["test_datasource"] = DataSource{
		DatabaseType: "unsupported_db_type",
		Initialized:  false,
	}

	ds := &DataSource{
		DatabaseType: "unsupported_db_type",
	}

	err := ConnectToDataSource("test_datasource", ds, logger, externalDataSources)

	if err == nil {
		t.Error("Expected error for unsupported database type, got nil")
	}

	expectedContains := "unsupported database type"
	if err.Error() == "" || !contains(err.Error(), expectedContains) {
		t.Errorf("Expected error to contain %q, got %q", expectedContains, err.Error())
	}

	// Verify status was set to unavailable
	if ds.Status != libConstant.DataSourceStatusUnavailable {
		t.Errorf("Expected status to be %q, got %q", libConstant.DataSourceStatusUnavailable, ds.Status)
	}

	// Verify LastError was set
	if ds.LastError == nil {
		t.Error("Expected LastError to be set")
	}
}

func TestConnectToDataSource_IncrementsRetryCount(t *testing.T) {
	t.Parallel()

	logger := mockLogger()
	externalDataSources := make(map[string]DataSource)

	externalDataSources["test_datasource"] = DataSource{
		DatabaseType: "unsupported_type",
		Initialized:  false,
		RetryCount:   0,
	}

	ds := &DataSource{
		DatabaseType: "unsupported_type",
		RetryCount:   0,
	}

	_ = ConnectToDataSource("test_datasource", ds, logger, externalDataSources)

	if ds.RetryCount != 1 {
		t.Errorf("Expected RetryCount to be 1, got %d", ds.RetryCount)
	}

	if ds.LastAttempt.IsZero() {
		t.Error("Expected LastAttempt to be set")
	}
}

func TestConnectToDataSource_MultipleInvalidRequestsDoNotCorruptMap(t *testing.T) {
	t.Parallel()

	logger := mockLogger()
	externalDataSources := make(map[string]DataSource)

	// Register valid datasources
	validDatasources := []string{"midaz_onboarding", "midaz_transaction", "plugin_crm"}
	for _, name := range validDatasources {
		externalDataSources[name] = DataSource{
			DatabaseType: PostgreSQLType,
			Initialized:  false,
		}
	}

	initialCount := len(externalDataSources)

	// Try multiple invalid datasource names (simulating frontend bugs)
	invalidNames := []string{
		"019abd3d-13c8-7692-8067-a9a9d42d9b41",
		"fake-uuid-1234-5678",
		"another-invalid-datasource",
		"not-a-real-database",
		"abd3d-13c8-7692-",
	}

	for _, invalidName := range invalidNames {
		ds := &DataSource{DatabaseType: PostgreSQLType}
		_ = ConnectToDataSource(invalidName, ds, logger, externalDataSources)
	}

	// Verify map was NOT corrupted
	if len(externalDataSources) != initialCount {
		t.Errorf("Map size changed from %d to %d after invalid requests", initialCount, len(externalDataSources))
	}

	// Verify no invalid names were added
	for _, invalidName := range invalidNames {
		if _, exists := externalDataSources[invalidName]; exists {
			t.Errorf("Invalid datasource %q was added to the map", invalidName)
		}
	}

	// Verify valid datasources still exist
	for _, validName := range validDatasources {
		if _, exists := externalDataSources[validName]; !exists {
			t.Errorf("Valid datasource %q was removed from the map", validName)
		}
	}
}

// =============================================================================
// isFatalError Tests
// =============================================================================

func TestIsFatalError_NilError(t *testing.T) {
	t.Parallel()

	if isFatalError(nil) {
		t.Error("Expected nil error to not be fatal")
	}
}

func TestIsFatalError_DNSErrors(t *testing.T) {
	t.Parallel()

	fatalErrors := []string{
		"lookup hostname: no such host",
		"dial tcp: lookup database.example.com: no such host",
		"server misbehaving",
		"connection refused",
	}

	for _, errMsg := range fatalErrors {
		err := &testError{msg: errMsg}
		if !isFatalError(err) {
			t.Errorf("Expected %q to be fatal error", errMsg)
		}
	}
}

func TestIsFatalError_AuthErrors(t *testing.T) {
	t.Parallel()

	fatalErrors := []string{
		"authentication failed for user 'admin'",
		"authorization failed: invalid credentials",
		"access denied for user 'root'",
	}

	for _, errMsg := range fatalErrors {
		err := &testError{msg: errMsg}
		if !isFatalError(err) {
			t.Errorf("Expected %q to be fatal error", errMsg)
		}
	}
}

func TestIsFatalError_ConfigErrors(t *testing.T) {
	t.Parallel()

	fatalErrors := []string{
		"unsupported database type: oracle",
		"invalid connection string: missing host",
	}

	for _, errMsg := range fatalErrors {
		err := &testError{msg: errMsg}
		if !isFatalError(err) {
			t.Errorf("Expected %q to be fatal error", errMsg)
		}
	}
}

func TestIsFatalError_NonFatalErrors(t *testing.T) {
	t.Parallel()

	nonFatalErrors := []string{
		"connection timeout",
		"network unreachable",
		"i/o timeout",
		"context deadline exceeded",
		"temporary failure",
	}

	for _, errMsg := range nonFatalErrors {
		err := &testError{msg: errMsg}
		if isFatalError(err) {
			t.Errorf("Expected %q to NOT be fatal error", errMsg)
		}
	}
}

func TestIsFatalError_CaseInsensitive(t *testing.T) {
	t.Parallel()

	// Should match regardless of case
	testCases := []string{
		"NO SUCH HOST",
		"No Such Host",
		"CONNECTION REFUSED",
		"Connection Refused",
		"AUTHENTICATION FAILED",
		"Authentication Failed",
	}

	for _, errMsg := range testCases {
		err := &testError{msg: errMsg}
		if !isFatalError(err) {
			t.Errorf("Expected %q to be fatal error (case insensitive)", errMsg)
		}
	}
}

// =============================================================================
// collectDataSourceNames Tests
// =============================================================================

func TestCollectDataSourceNames_FindsConfiguredDatasources(t *testing.T) {
	// Clear and set test environment variables
	clearDatasourceEnvVars()
	defer clearDatasourceEnvVars()

	os.Setenv("DATASOURCE_TEST_DB_CONFIG_NAME", "test_db")
	os.Setenv("DATASOURCE_ANOTHER_DB_CONFIG_NAME", "another_db")

	names := collectDataSourceNames()

	if len(names) != 2 {
		t.Errorf("Expected 2 datasource names, got %d", len(names))
	}

	if !names["test_db"] {
		t.Error("Expected 'test_db' to be in the names map")
	}

	if !names["another_db"] {
		t.Error("Expected 'another_db' to be in the names map")
	}
}

func TestCollectDataSourceNames_IgnoresNonDatasourceVars(t *testing.T) {
	clearDatasourceEnvVars()
	defer clearDatasourceEnvVars()

	// Set non-datasource env vars
	os.Setenv("SOME_OTHER_VAR", "value")
	os.Setenv("DATABASE_HOST", "localhost")
	os.Setenv("DATASOURCE_TEST_HOST", "localhost") // Missing CONFIG_NAME suffix

	// Set one valid datasource
	os.Setenv("DATASOURCE_VALID_CONFIG_NAME", "valid")

	names := collectDataSourceNames()

	if len(names) != 1 {
		t.Errorf("Expected 1 datasource name, got %d", len(names))
	}

	if !names["valid"] {
		t.Error("Expected 'valid' to be in the names map")
	}
}

func TestCollectDataSourceNames_EmptyWhenNoDatasources(t *testing.T) {
	clearDatasourceEnvVars()
	defer clearDatasourceEnvVars()

	names := collectDataSourceNames()

	if len(names) != 0 {
		t.Errorf("Expected 0 datasource names, got %d", len(names))
	}
}

func TestCollectDataSourceNames_ConvertsToLowercase(t *testing.T) {
	clearDatasourceEnvVars()
	defer clearDatasourceEnvVars()

	os.Setenv("DATASOURCE_UPPERCASE_DB_CONFIG_NAME", "uppercase_db")

	names := collectDataSourceNames()

	// The key should be lowercase
	if !names["uppercase_db"] {
		t.Error("Expected 'uppercase_db' (lowercase) to be in the names map")
	}
}

// =============================================================================
// buildDataSourceConfig Tests
// =============================================================================

func TestBuildDataSourceConfig_ReadsAllFields(t *testing.T) {
	clearDatasourceEnvVars()
	defer clearDatasourceEnvVars()

	os.Setenv("DATASOURCE_TEST_CONFIG_NAME", "test_datasource")
	os.Setenv("DATASOURCE_TEST_HOST", "localhost")
	os.Setenv("DATASOURCE_TEST_PORT", "5432")
	os.Setenv("DATASOURCE_TEST_USER", "postgres")
	os.Setenv("DATASOURCE_TEST_PASSWORD", "secret")
	os.Setenv("DATASOURCE_TEST_DATABASE", "testdb")
	os.Setenv("DATASOURCE_TEST_TYPE", "postgresql")
	os.Setenv("DATASOURCE_TEST_SSLMODE", "require")
	os.Setenv("DATASOURCE_TEST_SSLROOTCERT", "/path/to/cert")

	logger := mockLogger()
	config, isComplete := buildDataSourceConfig("test", logger)

	if !isComplete {
		t.Error("Expected config to be complete")
	}

	if config.ConfigName != "test_datasource" {
		t.Errorf("Expected ConfigName 'test_datasource', got %q", config.ConfigName)
	}

	if config.Host != "localhost" {
		t.Errorf("Expected Host 'localhost', got %q", config.Host)
	}

	if config.Port != "5432" {
		t.Errorf("Expected Port '5432', got %q", config.Port)
	}

	if config.User != "postgres" {
		t.Errorf("Expected User 'postgres', got %q", config.User)
	}

	if config.Password != "secret" {
		t.Errorf("Expected Password 'secret', got %q", config.Password)
	}

	if config.Database != "testdb" {
		t.Errorf("Expected Database 'testdb', got %q", config.Database)
	}

	if config.Type != "postgresql" {
		t.Errorf("Expected Type 'postgresql', got %q", config.Type)
	}

	if config.SSLMode != "require" {
		t.Errorf("Expected SSLMode 'require', got %q", config.SSLMode)
	}
}

func TestBuildDataSourceConfig_ReadsMongoDBFields(t *testing.T) {
	clearDatasourceEnvVars()
	defer clearDatasourceEnvVars()

	os.Setenv("DATASOURCE_MONGO_CONFIG_NAME", "mongo_datasource")
	os.Setenv("DATASOURCE_MONGO_HOST", "mongodb.example.com")
	os.Setenv("DATASOURCE_MONGO_PORT", "27017")
	os.Setenv("DATASOURCE_MONGO_USER", "mongo_user")
	os.Setenv("DATASOURCE_MONGO_PASSWORD", "mongo_secret")
	os.Setenv("DATASOURCE_MONGO_DATABASE", "mongodb")
	os.Setenv("DATASOURCE_MONGO_TYPE", "mongodb")
	os.Setenv("DATASOURCE_MONGO_SSL", "true")
	os.Setenv("DATASOURCE_MONGO_SSLCA", "/path/to/ca.pem")
	os.Setenv("DATASOURCE_MONGO_OPTIONS", "replicaSet=rs0&authSource=admin")

	logger := mockLogger()
	config, isComplete := buildDataSourceConfig("mongo", logger)

	if !isComplete {
		t.Error("Expected config to be complete")
	}

	if config.Type != "mongodb" {
		t.Errorf("Expected Type 'mongodb', got %q", config.Type)
	}

	if config.SSL != "true" {
		t.Errorf("Expected SSL 'true', got %q", config.SSL)
	}

	if config.SSLCA != "/path/to/ca.pem" {
		t.Errorf("Expected SSLCA '/path/to/ca.pem', got %q", config.SSLCA)
	}

	if config.Options != "replicaSet=rs0&authSource=admin" {
		t.Errorf("Expected Options 'replicaSet=rs0&authSource=admin', got %q", config.Options)
	}
}

func TestBuildDataSourceConfig_CaseInsensitiveName(t *testing.T) {
	clearDatasourceEnvVars()
	defer clearDatasourceEnvVars()

	os.Setenv("DATASOURCE_MYDB_CONFIG_NAME", "mydb_config")
	os.Setenv("DATASOURCE_MYDB_HOST", "localhost")
	os.Setenv("DATASOURCE_MYDB_TYPE", "postgresql")

	logger := mockLogger()

	// Should work with lowercase name
	config, _ := buildDataSourceConfig("mydb", logger)
	if config.ConfigName != "mydb_config" {
		t.Errorf("Expected ConfigName 'mydb_config', got %q", config.ConfigName)
	}
}

// =============================================================================
// ExternalDatasourceConnectionsLazy Tests
// =============================================================================

func TestExternalDatasourceConnectionsLazy_ReturnsEmptyMapWhenNoConfig(t *testing.T) {
	clearDatasourceEnvVars()
	defer clearDatasourceEnvVars()

	logger := mockLogger()
	datasources := ExternalDatasourceConnectionsLazy(logger)

	if len(datasources) != 0 {
		t.Errorf("Expected empty map, got %d entries", len(datasources))
	}
}

func TestExternalDatasourceConnectionsLazy_SkipsUnsupportedTypes(t *testing.T) {
	clearDatasourceEnvVars()
	defer clearDatasourceEnvVars()

	// Configure a datasource with unsupported type
	os.Setenv("DATASOURCE_ORACLE_CONFIG_NAME", "oracle_db")
	os.Setenv("DATASOURCE_ORACLE_HOST", "oracle.example.com")
	os.Setenv("DATASOURCE_ORACLE_PORT", "1521")
	os.Setenv("DATASOURCE_ORACLE_USER", "oracle_user")
	os.Setenv("DATASOURCE_ORACLE_PASSWORD", "oracle_pass")
	os.Setenv("DATASOURCE_ORACLE_DATABASE", "oracledb")
	os.Setenv("DATASOURCE_ORACLE_TYPE", "oracle") // Unsupported

	logger := mockLogger()
	datasources := ExternalDatasourceConnectionsLazy(logger)

	if len(datasources) != 0 {
		t.Errorf("Expected empty map for unsupported type, got %d entries", len(datasources))
	}
}

// =============================================================================
// DataSource struct Tests
// =============================================================================

func TestDataSource_DefaultValues(t *testing.T) {
	t.Parallel()

	ds := DataSource{}

	if ds.Initialized {
		t.Error("Expected Initialized to be false by default")
	}

	if ds.RetryCount != 0 {
		t.Error("Expected RetryCount to be 0 by default")
	}

	if !ds.LastAttempt.IsZero() {
		t.Error("Expected LastAttempt to be zero time by default")
	}

	if ds.LastError != nil {
		t.Error("Expected LastError to be nil by default")
	}
}

func TestDataSourceConfig_AllFieldsAccessible(t *testing.T) {
	t.Parallel()

	config := DataSourceConfig{
		ConfigName:  "test_config",
		Name:        "test",
		Host:        "localhost",
		Port:        "5432",
		User:        "user",
		Password:    "pass",
		Database:    "testdb",
		Type:        "postgresql",
		SSLMode:     "require",
		SSLCert:     "/path/cert",
		SSLRootCert: "/path/root",
		SSL:         "true",
		SSLCA:       "/path/ca",
		Options:     "option=value",
	}

	if config.ConfigName != "test_config" {
		t.Error("ConfigName field not accessible")
	}

	if config.Options != "option=value" {
		t.Error("Options field not accessible")
	}
}

// =============================================================================
// Integration-style Tests
// =============================================================================

func TestMapCorruptionPrevention_FullScenario(t *testing.T) {
	t.Parallel()

	logger := mockLogger()
	externalDataSources := make(map[string]DataSource)

	// Setup: Register legitimate datasources
	legitimateDatasources := map[string]DataSource{
		"midaz_onboarding": {
			DatabaseType: PostgreSQLType,
			Initialized:  false,
			Status:       libConstant.DataSourceStatusUnknown,
		},
		"midaz_transaction": {
			DatabaseType: PostgreSQLType,
			Initialized:  false,
			Status:       libConstant.DataSourceStatusUnknown,
		},
		"plugin_crm": {
			DatabaseType: MongoDBType,
			Initialized:  false,
			Status:       libConstant.DataSourceStatusUnknown,
		},
	}

	for name, ds := range legitimateDatasources {
		externalDataSources[name] = ds
	}

	// Record initial state
	initialIDs := make(map[string]bool)
	for id := range externalDataSources {
		initialIDs[id] = true
	}

	// Attack: Try to corrupt with various invalid datasource names
	attackPayloads := []string{
		// UUID fragments (common frontend bug)
		"019abd3d-13c8-7692-8067-a9a9d42d9b41",
		"abd3d-13c8-7692-8067-a9a9",
		"a7d2d-994a-7",
		"abd40-d07d",
		// Random strings
		"fake-datasource",
		"../../etc/passwd",
		"'; DROP TABLE users; --",
		"<script>alert('xss')</script>",
		// Empty and whitespace
		"",
		"   ",
	}

	for _, payload := range attackPayloads {
		ds := &DataSource{DatabaseType: PostgreSQLType}
		_ = ConnectToDataSource(payload, ds, logger, externalDataSources)
	}

	// Verify: Map should be unchanged
	if len(externalDataSources) != len(legitimateDatasources) {
		t.Errorf("Map size changed: expected %d, got %d",
			len(legitimateDatasources), len(externalDataSources))
	}

	// Verify: No new entries
	for id := range externalDataSources {
		if !initialIDs[id] {
			t.Errorf("Unexpected datasource appeared: %q", id)
		}
	}

	// Verify: All original entries still exist
	for id := range initialIDs {
		if _, exists := externalDataSources[id]; !exists {
			t.Errorf("Original datasource %q was removed", id)
		}
	}
}

func TestConnectToDataSource_RetryCountPersists(t *testing.T) {
	t.Parallel()

	logger := mockLogger()
	externalDataSources := make(map[string]DataSource)

	externalDataSources["test_ds"] = DataSource{
		DatabaseType: "unsupported",
		Initialized:  false,
		RetryCount:   5, // Already has 5 retries
	}

	ds := externalDataSources["test_ds"]

	_ = ConnectToDataSource("test_ds", &ds, logger, externalDataSources)

	// RetryCount should have incremented
	if ds.RetryCount != 6 {
		t.Errorf("Expected RetryCount to be 6, got %d", ds.RetryCount)
	}

	// LastAttempt should be recent
	if time.Since(ds.LastAttempt) > time.Second {
		t.Error("LastAttempt should be recent")
	}
}

// =============================================================================
// Helper functions and types
// =============================================================================

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func clearDatasourceEnvVars() {
	// Get all env vars and remove DATASOURCE_* ones
	for _, env := range os.Environ() {
		if len(env) > 11 && env[:11] == "DATASOURCE_" {
			parts := splitOnce(env, '=')
			if len(parts) > 0 {
				os.Unsetenv(parts[0])
			}
		}
	}
}

func splitOnce(s string, sep byte) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}
