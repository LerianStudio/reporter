// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validWorkerConfig returns a Config with all required fields populated.
func validWorkerConfig() *Config {
	return &Config{
		RabbitMQHost:                "localhost",
		RabbitMQPortAMQP:            "5672",
		RabbitMQUser:                "guest",
		RabbitMQPass:                "guest",
		RabbitMQGenerateReportQueue: "reporter.generate-report.queue",
		MongoDBHost:                 "localhost",
		MongoDBName:                 "reporter",
		ObjectStorageEndpoint:       "http://localhost:8333",
	}
}

func TestConfig_Validate_ValidConfig(t *testing.T) {
	t.Parallel()

	cfg := validWorkerConfig()
	err := cfg.Validate()
	require.NoError(t, err)
}

func TestConfig_Validate_AllFieldsMissing(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	err := cfg.Validate()
	require.Error(t, err)

	errMsg := err.Error()
	assert.Contains(t, errMsg, "config validation failed:")
	assert.Contains(t, errMsg, "RABBITMQ_HOST is required")
	assert.Contains(t, errMsg, "RABBITMQ_PORT_AMQP is required")
	assert.Contains(t, errMsg, "RABBITMQ_DEFAULT_USER is required")
	assert.Contains(t, errMsg, "RABBITMQ_DEFAULT_PASS is required")
	assert.Contains(t, errMsg, "RABBITMQ_GENERATE_REPORT_QUEUE is required")
	assert.Contains(t, errMsg, "MONGO_HOST is required")
	assert.Contains(t, errMsg, "MONGO_NAME is required")
	assert.Contains(t, errMsg, "OBJECT_STORAGE_ENDPOINT is required")
}

func TestConfig_Validate_SingleFieldMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		modify      func(cfg *Config)
		expectedErr string
	}{
		{
			name:        "missing RabbitMQHost",
			modify:      func(cfg *Config) { cfg.RabbitMQHost = "" },
			expectedErr: "RABBITMQ_HOST is required",
		},
		{
			name:        "missing RabbitMQPortAMQP",
			modify:      func(cfg *Config) { cfg.RabbitMQPortAMQP = "" },
			expectedErr: "RABBITMQ_PORT_AMQP is required",
		},
		{
			name:        "missing RabbitMQUser",
			modify:      func(cfg *Config) { cfg.RabbitMQUser = "" },
			expectedErr: "RABBITMQ_DEFAULT_USER is required",
		},
		{
			name:        "missing RabbitMQPass",
			modify:      func(cfg *Config) { cfg.RabbitMQPass = "" },
			expectedErr: "RABBITMQ_DEFAULT_PASS is required",
		},
		{
			name:        "missing RabbitMQGenerateReportQueue",
			modify:      func(cfg *Config) { cfg.RabbitMQGenerateReportQueue = "" },
			expectedErr: "RABBITMQ_GENERATE_REPORT_QUEUE is required",
		},
		{
			name:        "missing MongoDBHost",
			modify:      func(cfg *Config) { cfg.MongoDBHost = "" },
			expectedErr: "MONGO_HOST is required",
		},
		{
			name:        "missing MongoDBName",
			modify:      func(cfg *Config) { cfg.MongoDBName = "" },
			expectedErr: "MONGO_NAME is required",
		},
		{
			name:        "missing ObjectStorageEndpoint",
			modify:      func(cfg *Config) { cfg.ObjectStorageEndpoint = "" },
			expectedErr: "OBJECT_STORAGE_ENDPOINT is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := validWorkerConfig()
			tt.modify(cfg)

			err := cfg.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestConfig_Validate_MultipleFieldsMissing(t *testing.T) {
	t.Parallel()

	cfg := validWorkerConfig()
	cfg.RabbitMQHost = ""
	cfg.MongoDBHost = ""
	cfg.ObjectStorageEndpoint = ""

	err := cfg.Validate()
	require.Error(t, err)

	errMsg := err.Error()
	assert.Contains(t, errMsg, "RABBITMQ_HOST is required")
	assert.Contains(t, errMsg, "MONGO_HOST is required")
	assert.Contains(t, errMsg, "OBJECT_STORAGE_ENDPOINT is required")

	// Verify multi-line format
	lines := strings.Split(errMsg, "\n")
	assert.GreaterOrEqual(t, len(lines), 4) // header + 3 errors
}

func TestConfig_ValidateProductionConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		modify      func(cfg *Config)
		expectErr   bool
		errContains []string
	}{
		{
			name: "valid production config",
			modify: func(cfg *Config) {
				cfg.EnvName = "production"
				cfg.EnableTelemetry = true
				cfg.MongoDBPassword = "real-secret"
				cfg.RabbitMQPass = "real-secret"
				cfg.ObjectStorageSecretKey = "real-secret"
				cfg.CryptoHashSecretKeyPluginCRM = "real-secret"
				cfg.CryptoEncryptSecretKeyPluginCRM = "real-secret"
			},
			expectErr: false,
		},
		{
			name: "production requires telemetry enabled",
			modify: func(cfg *Config) {
				cfg.EnvName = "production"
				cfg.EnableTelemetry = false
				cfg.MongoDBPassword = "real-secret"
				cfg.RabbitMQPass = "real-secret"
				cfg.ObjectStorageSecretKey = "real-secret"
				cfg.CryptoHashSecretKeyPluginCRM = "real-secret"
				cfg.CryptoEncryptSecretKeyPluginCRM = "real-secret"
			},
			expectErr:   true,
			errContains: []string{"ENABLE_TELEMETRY must be true in production"},
		},
		{
			name: "production rejects default placeholder password",
			modify: func(cfg *Config) {
				cfg.EnvName = "production"
				cfg.EnableTelemetry = true
				cfg.MongoDBPassword = "CHANGE_ME"
				cfg.RabbitMQPass = "real-secret"
				cfg.ObjectStorageSecretKey = "real-secret"
				cfg.CryptoHashSecretKeyPluginCRM = "real-secret"
				cfg.CryptoEncryptSecretKeyPluginCRM = "real-secret"
			},
			expectErr:   true,
			errContains: []string{"MONGO_PASSWORD must not use the default placeholder in production"},
		},
		{
			name: "production rejects empty secrets",
			modify: func(cfg *Config) {
				cfg.EnvName = "production"
				cfg.EnableTelemetry = true
				cfg.MongoDBPassword = ""
				cfg.RabbitMQPass = ""
				cfg.ObjectStorageSecretKey = "real-secret"
				cfg.CryptoHashSecretKeyPluginCRM = "real-secret"
				cfg.CryptoEncryptSecretKeyPluginCRM = "real-secret"
			},
			expectErr:   true,
			errContains: []string{"MONGO_PASSWORD must not be empty in production", "RABBITMQ_DEFAULT_PASS must not be empty in production"},
		},
		{
			name: "production rejects empty crypto keys",
			modify: func(cfg *Config) {
				cfg.EnvName = "production"
				cfg.EnableTelemetry = true
				cfg.MongoDBPassword = "real-secret"
				cfg.RabbitMQPass = "real-secret"
				cfg.ObjectStorageSecretKey = "real-secret"
				cfg.CryptoHashSecretKeyPluginCRM = ""
				cfg.CryptoEncryptSecretKeyPluginCRM = ""
			},
			expectErr: true,
			errContains: []string{
				"CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM must not be empty in production",
				"CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM must not be empty in production",
			},
		},
		{
			name: "non-production skips production validation",
			modify: func(cfg *Config) {
				cfg.EnvName = "staging"
				cfg.EnableTelemetry = false
				cfg.MongoDBPassword = "CHANGE_ME"
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := validWorkerConfig()
			tt.modify(cfg)

			err := cfg.Validate()

			if tt.expectErr {
				require.Error(t, err)
				for _, expected := range tt.errContains {
					assert.Contains(t, err.Error(), expected)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_Validate_OptionalFieldsCanBeEmpty(t *testing.T) {
	t.Parallel()

	cfg := validWorkerConfig()
	// These fields are optional and should not cause validation errors
	cfg.EnvName = ""
	cfg.LogLevel = ""
	cfg.MongoDBPassword = ""
	cfg.MongoDBParameters = ""
	cfg.RabbitMQHealthCheckURL = ""

	err := cfg.Validate()
	require.NoError(t, err)
}
