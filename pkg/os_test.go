package pkg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "returns env value when set",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			envValue:     "actual_value",
			setEnv:       true,
			expected:     "actual_value",
		},
		{
			name:         "returns default when env not set",
			key:          "TEST_ENV_VAR_NOT_SET",
			defaultValue: "default_value",
			envValue:     "",
			setEnv:       false,
			expected:     "default_value",
		},
		{
			name:         "returns default when env is empty string",
			key:          "TEST_ENV_VAR_EMPTY",
			defaultValue: "fallback",
			envValue:     "",
			setEnv:       true,
			expected:     "fallback",
		},
		{
			name:         "returns default when env is whitespace only",
			key:          "TEST_ENV_VAR_WHITESPACE",
			defaultValue: "fallback",
			envValue:     "   ",
			setEnv:       true,
			expected:     "fallback",
		},
		{
			name:         "returns env value with spaces trimmed check",
			key:          "TEST_ENV_VAR_SPACES",
			defaultValue: "default",
			envValue:     "  value_with_spaces  ",
			setEnv:       true,
			expected:     "  value_with_spaces  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before and after test
			os.Unsetenv(tt.key)
			defer os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			}

			result := GetEnvOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetenvBoolOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		setEnv       bool
		expected     bool
	}{
		{
			name:         "returns true when env is 'true'",
			key:          "TEST_BOOL_TRUE",
			defaultValue: false,
			envValue:     "true",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "returns false when env is 'false'",
			key:          "TEST_BOOL_FALSE",
			defaultValue: true,
			envValue:     "false",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "returns true when env is '1'",
			key:          "TEST_BOOL_ONE",
			defaultValue: false,
			envValue:     "1",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "returns false when env is '0'",
			key:          "TEST_BOOL_ZERO",
			defaultValue: true,
			envValue:     "0",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "returns default when env not set",
			key:          "TEST_BOOL_NOT_SET",
			defaultValue: true,
			envValue:     "",
			setEnv:       false,
			expected:     true,
		},
		{
			name:         "returns default when env is invalid",
			key:          "TEST_BOOL_INVALID",
			defaultValue: true,
			envValue:     "invalid",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "returns default false when env is invalid",
			key:          "TEST_BOOL_INVALID_FALSE",
			defaultValue: false,
			envValue:     "not_a_bool",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "returns true when env is 'TRUE' uppercase",
			key:          "TEST_BOOL_UPPERCASE",
			defaultValue: false,
			envValue:     "TRUE",
			setEnv:       true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)
			defer os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			}

			result := GetenvBoolOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetenvIntOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int64
		envValue     string
		setEnv       bool
		expected     int64
	}{
		{
			name:         "returns int value when valid",
			key:          "TEST_INT_VALID",
			defaultValue: 0,
			envValue:     "42",
			setEnv:       true,
			expected:     42,
		},
		{
			name:         "returns negative int value",
			key:          "TEST_INT_NEGATIVE",
			defaultValue: 0,
			envValue:     "-100",
			setEnv:       true,
			expected:     -100,
		},
		{
			name:         "returns default when env not set",
			key:          "TEST_INT_NOT_SET",
			defaultValue: 999,
			envValue:     "",
			setEnv:       false,
			expected:     999,
		},
		{
			name:         "returns default when env is invalid",
			key:          "TEST_INT_INVALID",
			defaultValue: 50,
			envValue:     "not_a_number",
			setEnv:       true,
			expected:     50,
		},
		{
			name:         "returns default when env is float",
			key:          "TEST_INT_FLOAT",
			defaultValue: 10,
			envValue:     "3.14",
			setEnv:       true,
			expected:     10,
		},
		{
			name:         "returns zero when valid zero",
			key:          "TEST_INT_ZERO",
			defaultValue: 100,
			envValue:     "0",
			setEnv:       true,
			expected:     0,
		},
		{
			name:         "returns large int64 value",
			key:          "TEST_INT_LARGE",
			defaultValue: 0,
			envValue:     "9223372036854775807",
			setEnv:       true,
			expected:     9223372036854775807,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)
			defer os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			}

			result := GetenvIntOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetConfigFromEnvVars(t *testing.T) {
	type TestConfig struct {
		StringField string `env:"TEST_STRING_FIELD"`
		IntField    int64  `env:"TEST_INT_FIELD"`
		BoolField   bool   `env:"TEST_BOOL_FIELD"`
		NoTagField  string
	}

	t.Run("sets string field from env", func(t *testing.T) {
		os.Setenv("TEST_STRING_FIELD", "test_value")
		defer os.Unsetenv("TEST_STRING_FIELD")

		config := &TestConfig{}
		err := SetConfigFromEnvVars(config)

		assert.NoError(t, err)
		assert.Equal(t, "test_value", config.StringField)
	})

	t.Run("sets int field from env", func(t *testing.T) {
		os.Setenv("TEST_INT_FIELD", "123")
		defer os.Unsetenv("TEST_INT_FIELD")

		config := &TestConfig{}
		err := SetConfigFromEnvVars(config)

		assert.NoError(t, err)
		assert.Equal(t, int64(123), config.IntField)
	})

	t.Run("sets bool field from env", func(t *testing.T) {
		os.Setenv("TEST_BOOL_FIELD", "true")
		defer os.Unsetenv("TEST_BOOL_FIELD")

		config := &TestConfig{}
		err := SetConfigFromEnvVars(config)

		assert.NoError(t, err)
		assert.True(t, config.BoolField)
	})

	t.Run("sets multiple fields from env", func(t *testing.T) {
		os.Setenv("TEST_STRING_FIELD", "multi_test")
		os.Setenv("TEST_INT_FIELD", "456")
		os.Setenv("TEST_BOOL_FIELD", "true")
		defer func() {
			os.Unsetenv("TEST_STRING_FIELD")
			os.Unsetenv("TEST_INT_FIELD")
			os.Unsetenv("TEST_BOOL_FIELD")
		}()

		config := &TestConfig{}
		err := SetConfigFromEnvVars(config)

		assert.NoError(t, err)
		assert.Equal(t, "multi_test", config.StringField)
		assert.Equal(t, int64(456), config.IntField)
		assert.True(t, config.BoolField)
	})

	t.Run("returns error when not a pointer", func(t *testing.T) {
		config := TestConfig{}
		err := SetConfigFromEnvVars(config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be an pointer")
	})

	t.Run("ignores fields without env tag", func(t *testing.T) {
		config := &TestConfig{NoTagField: "initial"}
		err := SetConfigFromEnvVars(config)

		assert.NoError(t, err)
		assert.Equal(t, "initial", config.NoTagField)
	})
}

func TestSetConfigFromEnvVars_IntTypes(t *testing.T) {
	type IntTypesConfig struct {
		Int8Field  int8  `env:"TEST_INT8"`
		Int16Field int16 `env:"TEST_INT16"`
		Int32Field int32 `env:"TEST_INT32"`
		IntField   int   `env:"TEST_INT"`
	}

	t.Run("sets various int types", func(t *testing.T) {
		os.Setenv("TEST_INT8", "127")
		os.Setenv("TEST_INT16", "32767")
		os.Setenv("TEST_INT32", "2147483647")
		os.Setenv("TEST_INT", "999")
		defer func() {
			os.Unsetenv("TEST_INT8")
			os.Unsetenv("TEST_INT16")
			os.Unsetenv("TEST_INT32")
			os.Unsetenv("TEST_INT")
		}()

		config := &IntTypesConfig{}
		err := SetConfigFromEnvVars(config)

		assert.NoError(t, err)
		assert.Equal(t, int8(127), config.Int8Field)
		assert.Equal(t, int16(32767), config.Int16Field)
		assert.Equal(t, int32(2147483647), config.Int32Field)
		assert.Equal(t, int(999), config.IntField)
	})
}

func TestEnsureConfigFromEnvVars(t *testing.T) {
	type TestConfig struct {
		Field string `env:"TEST_ENSURE_FIELD"`
	}

	t.Run("returns config when successful", func(t *testing.T) {
		os.Setenv("TEST_ENSURE_FIELD", "value")
		defer os.Unsetenv("TEST_ENSURE_FIELD")

		config := &TestConfig{}
		result := EnsureConfigFromEnvVars(config)

		assert.NotNil(t, result)
		assert.Equal(t, "value", config.Field)
	})

	t.Run("panics when not a pointer", func(t *testing.T) {
		config := TestConfig{}

		assert.Panics(t, func() {
			EnsureConfigFromEnvVars(config)
		})
	})
}

func TestLocalEnvConfig_Struct(t *testing.T) {
	config := LocalEnvConfig{
		Initialized: true,
	}

	assert.True(t, config.Initialized)

	config.Initialized = false
	assert.False(t, config.Initialized)
}
