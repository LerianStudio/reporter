// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pkg

import (
	"context"
	"errors"
	"fmt"
	"testing"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsFatalError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error returns false",
			err:      nil,
			expected: false,
		},
		{
			name:     "connection refused is fatal",
			err:      errors.New("dial tcp 127.0.0.1:5432: connection refused"),
			expected: true,
		},
		{
			name:     "no such host is fatal",
			err:      errors.New("dial tcp: lookup db.example.com: no such host"),
			expected: true,
		},
		{
			name:     "DNS lookup failure is fatal",
			err:      errors.New("lookup db.example.com on 8.8.8.8:53: no such host"),
			expected: true,
		},
		{
			name:     "server misbehaving is fatal",
			err:      errors.New("lookup db.example.com: server misbehaving"),
			expected: true,
		},
		{
			name:     "unsupported database type is fatal",
			err:      errors.New("unsupported database type: oracle"),
			expected: true,
		},
		{
			name:     "invalid connection string is fatal",
			err:      errors.New("invalid connection string: missing host"),
			expected: true,
		},
		{
			name:     "authentication failed is fatal",
			err:      errors.New("authentication failed for user 'admin'"),
			expected: true,
		},
		{
			name:     "authorization failed is fatal",
			err:      errors.New("authorization failed: insufficient privileges"),
			expected: true,
		},
		{
			name:     "access denied is fatal",
			err:      errors.New("access denied for user 'readonly'@'10.0.0.1'"),
			expected: true,
		},
		{
			name:     "timeout error is retryable",
			err:      errors.New("dial tcp 127.0.0.1:5432: i/o timeout"),
			expected: false,
		},
		{
			name:     "connection reset is retryable",
			err:      errors.New("read: connection reset by peer"),
			expected: false,
		},
		{
			name:     "generic error is retryable",
			err:      errors.New("something went wrong"),
			expected: false,
		},
		{
			name:     "EOF error is retryable",
			err:      errors.New("unexpected EOF"),
			expected: false,
		},
		{
			name:     "case insensitive matching - CONNECTION REFUSED",
			err:      errors.New("CONNECTION REFUSED by server"),
			expected: true,
		},
		{
			name:     "case insensitive matching - Authentication Failed uppercase",
			err:      errors.New("Authentication Failed for user postgres"),
			expected: true,
		},
		{
			name:     "partial match - contains connection refused in longer message",
			err:      errors.New("failed to connect to postgres: connection refused at host:5432"),
			expected: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isFatalError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegisterDataSourceIDsForTesting(t *testing.T) {
	// Note: Cannot use t.Parallel() because it modifies package-level state

	// Reset to ensure clean state
	ResetRegisteredDataSourceIDsForTesting()

	t.Cleanup(func() {
		ResetRegisteredDataSourceIDsForTesting()
	})

	RegisterDataSourceIDsForTesting([]string{"test_db_1", "test_db_2"})

	assert.True(t, IsValidDataSourceID("test_db_1"))
	assert.True(t, IsValidDataSourceID("test_db_2"))
	assert.False(t, IsValidDataSourceID("test_db_3"))
}

func TestResetRegisteredDataSourceIDsForTesting(t *testing.T) {
	// Note: Cannot use t.Parallel() because it modifies package-level state

	t.Cleanup(func() {
		ResetRegisteredDataSourceIDsForTesting()
	})

	// Register some IDs
	RegisterDataSourceIDsForTesting([]string{"reset_db_1", "reset_db_2"})
	assert.True(t, IsValidDataSourceID("reset_db_1"))

	// Reset should clear all IDs
	ResetRegisteredDataSourceIDsForTesting()

	assert.False(t, IsValidDataSourceID("reset_db_1"))
	assert.False(t, IsValidDataSourceID("reset_db_2"))
}

func TestIsValidDataSourceID(t *testing.T) {
	// Note: Cannot use t.Parallel() because it modifies package-level state

	ResetRegisteredDataSourceIDsForTesting()

	t.Cleanup(func() {
		ResetRegisteredDataSourceIDsForTesting()
	})

	tests := []struct {
		name     string
		id       string
		register []string
		expected bool
	}{
		{
			name:     "valid registered ID",
			id:       "valid_db",
			register: []string{"valid_db"},
			expected: true,
		},
		{
			name:     "unregistered ID returns false",
			id:       "unknown_db",
			register: []string{"other_db"},
			expected: false,
		},
		{
			name:     "empty string ID returns false",
			id:       "",
			register: []string{"some_db"},
			expected: false,
		},
	}

	for _, tt := range tests {
		ResetRegisteredDataSourceIDsForTesting()
		RegisterDataSourceIDsForTesting(tt.register)

		t.Run(tt.name, func(t *testing.T) {
			result := IsValidDataSourceID(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInitRegisteredDataSourceIDs(t *testing.T) {
	// Note: Cannot use t.Parallel() because it modifies package-level state

	ResetRegisteredDataSourceIDsForTesting()

	t.Cleanup(func() {
		ResetRegisteredDataSourceIDsForTesting()
	})

	initRegisteredDataSourceIDs([]string{"init_db_1", "init_db_2"})

	assert.True(t, IsValidDataSourceID("init_db_1"))
	assert.True(t, IsValidDataSourceID("init_db_2"))
	assert.False(t, IsValidDataSourceID("init_db_3"))

	// Second call should be a no-op (sync.Once)
	initRegisteredDataSourceIDs([]string{"init_db_3"})
	assert.False(t, IsValidDataSourceID("init_db_3"), "sync.Once should prevent second registration")
}

func TestCollectDataSourceNames(t *testing.T) {
	// Note: Cannot use t.Parallel() because t.Setenv is used

	t.Run("collects datasource names from environment", func(t *testing.T) {
		t.Setenv("DATASOURCE_MIDAZ_LEDGER_CONFIG_NAME", "midaz-ledger")
		t.Setenv("DATASOURCE_MIDAZ_LEDGER_HOST", "localhost")
		t.Setenv("DATASOURCE_CRM_CONFIG_NAME", "crm")

		names := collectDataSourceNames()

		assert.True(t, names["midaz_ledger"], "should find midaz_ledger datasource")
		assert.True(t, names["crm"], "should find crm datasource")
	})

	t.Run("returns empty map when no datasource envs exist", func(t *testing.T) {
		// Clear relevant env vars by setting to empty
		// collectDataSourceNames looks for DATASOURCE_*_CONFIG_NAME pattern
		names := collectDataSourceNames()

		// We cannot guarantee empty since other tests or system may have set env vars,
		// but at minimum the function should not panic
		assert.NotNil(t, names)
	})
}

func TestBuildDataSourceConfig(t *testing.T) {
	// Note: Cannot use t.Parallel() because t.Setenv is used

	t.Run("builds complete config from environment", func(t *testing.T) {
		t.Setenv("DATASOURCE_TEST_DB_CONFIG_NAME", "test-db")
		t.Setenv("DATASOURCE_TEST_DB_HOST", "localhost")
		t.Setenv("DATASOURCE_TEST_DB_PORT", "5432")
		t.Setenv("DATASOURCE_TEST_DB_USER", "admin")
		t.Setenv("DATASOURCE_TEST_DB_PASSWORD", "secret")
		t.Setenv("DATASOURCE_TEST_DB_DATABASE", "testdb")
		t.Setenv("DATASOURCE_TEST_DB_TYPE", "postgresql")
		t.Setenv("DATASOURCE_TEST_DB_SSLMODE", "disable")

		logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

		config, isComplete := buildDataSourceConfig("test_db", logger)

		assert.True(t, isComplete)
		assert.Equal(t, "test-db", config.ConfigName)
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, "5432", config.Port)
		assert.Equal(t, "admin", config.User)
		assert.Equal(t, "secret", config.Password)
		assert.Equal(t, "testdb", config.Database)
		assert.Equal(t, "postgresql", config.Type)
		assert.Equal(t, "disable", config.SSLMode)
	})

	t.Run("builds config with MongoDB fields", func(t *testing.T) {
		t.Setenv("DATASOURCE_MONGO_DB_CONFIG_NAME", "mongo-db")
		t.Setenv("DATASOURCE_MONGO_DB_HOST", "mongo-host")
		t.Setenv("DATASOURCE_MONGO_DB_PORT", "27017")
		t.Setenv("DATASOURCE_MONGO_DB_USER", "mongouser")
		t.Setenv("DATASOURCE_MONGO_DB_PASSWORD", "mongosecret")
		t.Setenv("DATASOURCE_MONGO_DB_DATABASE", "reporterdb")
		t.Setenv("DATASOURCE_MONGO_DB_TYPE", "mongodb")
		t.Setenv("DATASOURCE_MONGO_DB_SSL", "true")
		t.Setenv("DATASOURCE_MONGO_DB_SSLCA", "/path/to/ca.pem")
		t.Setenv("DATASOURCE_MONGO_DB_OPTIONS", "authSource=admin")

		logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

		config, isComplete := buildDataSourceConfig("mongo_db", logger)

		assert.True(t, isComplete)
		assert.Equal(t, "mongo-db", config.ConfigName)
		assert.Equal(t, "mongodb", config.Type)
		assert.Equal(t, "true", config.SSL)
		assert.Equal(t, "/path/to/ca.pem", config.SSLCA)
		assert.Equal(t, "authSource=admin", config.Options)
	})

	t.Run("builds config with midaz organization ID", func(t *testing.T) {
		t.Setenv("DATASOURCE_CRM_DS_CONFIG_NAME", "crm-ds")
		t.Setenv("DATASOURCE_CRM_DS_HOST", "localhost")
		t.Setenv("DATASOURCE_CRM_DS_PORT", "27017")
		t.Setenv("DATASOURCE_CRM_DS_USER", "user")
		t.Setenv("DATASOURCE_CRM_DS_PASSWORD", "pass")
		t.Setenv("DATASOURCE_CRM_DS_DATABASE", "crm")
		t.Setenv("DATASOURCE_CRM_DS_TYPE", "mongodb")
		t.Setenv("DATASOURCE_CRM_DS_MIDAZ_ORGANIZATION_ID", "org-123-456")

		logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

		config, isComplete := buildDataSourceConfig("crm_ds", logger)

		assert.True(t, isComplete)
		assert.Equal(t, "org-123-456", config.MidazOrganizationID)
	})
}

func TestGetDataSourceConfigs(t *testing.T) {
	// Note: Cannot use t.Parallel() because t.Setenv is used

	t.Run("returns configs from environment", func(t *testing.T) {
		t.Setenv("DATASOURCE_INTEG_DB_CONFIG_NAME", "integ-db")
		t.Setenv("DATASOURCE_INTEG_DB_HOST", "localhost")
		t.Setenv("DATASOURCE_INTEG_DB_PORT", "5432")
		t.Setenv("DATASOURCE_INTEG_DB_USER", "user")
		t.Setenv("DATASOURCE_INTEG_DB_PASSWORD", "pass")
		t.Setenv("DATASOURCE_INTEG_DB_DATABASE", "integdb")
		t.Setenv("DATASOURCE_INTEG_DB_TYPE", "postgresql")
		t.Setenv("DATASOURCE_INTEG_DB_SSLMODE", "disable")

		logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

		configs := getDataSourceConfigs(logger)

		// Should find at least one datasource (the one we set up)
		found := false
		for _, c := range configs {
			if c.ConfigName == "integ-db" {
				found = true
				assert.Equal(t, "localhost", c.Host)
				assert.Equal(t, "5432", c.Port)

				break
			}
		}

		assert.True(t, found, "should find the integ-db datasource in configs")
	})
}

func TestGetDataSourceConfigs_ReturnsSlice(t *testing.T) {
	// Note: Cannot use t.Parallel() because t.Setenv is used
	// This test verifies the function executes without error.
	// The result may or may not be empty depending on environment state.

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	// Should not panic regardless of environment state
	assert.NotPanics(t, func() {
		_ = getDataSourceConfigs(logger)
	})
}

func TestGetDataSourceEnv(t *testing.T) {
	// Note: Cannot use t.Parallel() because t.Setenv is used

	tests := []struct {
		name     string
		dsName   string
		field    string
		envKey   string
		envValue string
		expected string
	}{
		{
			name:     "simple name and field",
			dsName:   "mydb",
			field:    "HOST",
			envKey:   "DATASOURCE_MYDB_HOST",
			envValue: "localhost",
			expected: "localhost",
		},
		{
			name:     "name with hyphen converted to underscore",
			dsName:   "my-db",
			field:    "PORT",
			envKey:   "DATASOURCE_MY_DB_PORT",
			envValue: "5432",
			expected: "5432",
		},
		{
			name:     "unset env returns empty string",
			dsName:   "nonexistent",
			field:    "HOST",
			envKey:   "",
			envValue: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}

			result := getDataSourceEnv(tt.dsName, tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConnectToDataSource_UnregisteredDatasource(t *testing.T) {
	// Note: Cannot use t.Parallel() because it modifies package-level state

	ResetRegisteredDataSourceIDsForTesting()

	t.Cleanup(func() {
		ResetRegisteredDataSourceIDsForTesting()
	})

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	ds := &DataSource{DatabaseType: PostgreSQLType}
	externalDS := map[string]DataSource{"unregistered_db": *ds}

	err := ConnectToDataSource("unregistered_db", ds, logger, externalDS)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unregistered datasource")
}

func TestConnectToDataSource_NotInRuntimeMap(t *testing.T) {
	// Note: Cannot use t.Parallel() because it modifies package-level state

	ResetRegisteredDataSourceIDsForTesting()
	RegisterDataSourceIDsForTesting([]string{"orphan_db"})

	t.Cleanup(func() {
		ResetRegisteredDataSourceIDsForTesting()
	})

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	ds := &DataSource{DatabaseType: PostgreSQLType}
	externalDS := map[string]DataSource{} // empty map - not in runtime

	err := ConnectToDataSource("orphan_db", ds, logger, externalDS)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in runtime map")
}

func TestConnectToDataSource_UnsupportedDatabaseType(t *testing.T) {
	// Note: Cannot use t.Parallel() because it modifies package-level state

	ResetRegisteredDataSourceIDsForTesting()
	RegisterDataSourceIDsForTesting([]string{"unsupported_db"})

	t.Cleanup(func() {
		ResetRegisteredDataSourceIDsForTesting()
	})

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	ds := &DataSource{DatabaseType: "oracle"}
	externalDS := map[string]DataSource{"unsupported_db": *ds}

	err := ConnectToDataSource("unsupported_db", ds, logger, externalDS)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported database type")
	assert.Contains(t, err.Error(), "oracle")
	assert.NotNil(t, ds.LastError)
}

func TestConnectToDataSource_MongoDBInvalidURI(t *testing.T) {
	// Note: Cannot use t.Parallel() because it modifies package-level state

	ResetRegisteredDataSourceIDsForTesting()
	RegisterDataSourceIDsForTesting([]string{"mongo_bad_db"})

	t.Cleanup(func() {
		ResetRegisteredDataSourceIDsForTesting()
	})

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	ds := &DataSource{
		DatabaseType: MongoDBType,
		MongoURI:     "invalid://not-a-valid-uri",
		MongoDBName:  "testdb",
	}
	externalDS := map[string]DataSource{"mongo_bad_db": *ds}

	err := ConnectToDataSource("mongo_bad_db", ds, logger, externalDS)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MongoDB")
}

func TestConnectToDataSource_IncreasesRetryCount(t *testing.T) {
	// Note: Cannot use t.Parallel() because it modifies package-level state

	ResetRegisteredDataSourceIDsForTesting()
	RegisterDataSourceIDsForTesting([]string{"retry_db"})

	t.Cleanup(func() {
		ResetRegisteredDataSourceIDsForTesting()
	})

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	ds := &DataSource{
		DatabaseType: "oracle", // unsupported, so it will fail predictably
		RetryCount:   0,
	}
	externalDS := map[string]DataSource{"retry_db": *ds}

	_ = ConnectToDataSource("retry_db", ds, logger, externalDS)

	assert.Equal(t, 1, ds.RetryCount, "RetryCount should be incremented")
	assert.False(t, ds.LastAttempt.IsZero(), "LastAttempt should be set")
}

func TestConnectToDataSourceWithRetry_UnsupportedType(t *testing.T) {
	// Note: Cannot use t.Parallel() because it modifies package-level state

	ResetRegisteredDataSourceIDsForTesting()
	RegisterDataSourceIDsForTesting([]string{"retry_unsupported_db"})

	t.Cleanup(func() {
		ResetRegisteredDataSourceIDsForTesting()
	})

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	ds := &DataSource{
		DatabaseType: "oracle", // unsupported - will be fatal, skipping retries
	}
	externalDS := map[string]DataSource{"retry_unsupported_db": *ds}

	err := ConnectToDataSourceWithRetry("retry_unsupported_db", ds, logger, externalDS)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "retry_unsupported_db")
}

func TestConnectToDataSourceWithRetry_FatalErrorSkipsRetries(t *testing.T) {
	// Note: Cannot use t.Parallel() because it modifies package-level state

	ResetRegisteredDataSourceIDsForTesting()
	RegisterDataSourceIDsForTesting([]string{"fatal_err_db"})

	t.Cleanup(func() {
		ResetRegisteredDataSourceIDsForTesting()
	})

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	// "unsupported database type" is in the fatalPatterns list, so it should skip retries
	ds := &DataSource{
		DatabaseType: "cassandra", // unsupported = fatal
	}
	externalDS := map[string]DataSource{"fatal_err_db": *ds}

	err := ConnectToDataSourceWithRetry("fatal_err_db", ds, logger, externalDS)
	require.Error(t, err)

	// RetryCount should be low since fatal errors skip retries
	// First attempt increments to 1, fatal detected -> break
	assert.LessOrEqual(t, ds.RetryCount, 2, "fatal errors should skip remaining retries")
}

func TestDataSourceConfig_GetSchemas_EmptyAfterTrim(t *testing.T) {
	// Note: Cannot use t.Parallel() because t.Setenv is used

	t.Setenv("DATASOURCE_TRIM_TEST_SCHEMAS", ", , ,")

	config := DataSourceConfig{
		ConfigName: "trim-test",
	}

	schemas := config.GetSchemas()

	// All entries are empty after trim, so should return default
	assert.Equal(t, []string{"public"}, schemas)
}

func TestDataSourceConfig_GetSchemas_MixedSpacing(t *testing.T) {
	// Note: Cannot use t.Parallel() because t.Setenv is used

	envKey := fmt.Sprintf("DATASOURCE_%s_SCHEMAS", toEnvFormat("mixed-spacing"))
	t.Setenv(envKey, "  schema_a , schema_b ,schema_c  ")

	config := DataSourceConfig{
		ConfigName: "mixed-spacing",
	}

	schemas := config.GetSchemas()
	assert.Equal(t, []string{"schema_a", "schema_b", "schema_c"}, schemas)
}
