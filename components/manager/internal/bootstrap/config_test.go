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

// validManagerConfig returns a Config with all required fields populated.
func validManagerConfig() *Config {
	return &Config{
		ServerAddress:             "localhost:4005",
		MongoDBHost:               "localhost",
		MongoDBName:               "reporter",
		RabbitMQHost:              "localhost",
		RabbitMQPortAMQP:          "5672",
		RabbitMQUser:              "guest",
		RabbitMQPass:              "guest",
		RabbitMQGenerateReportQueue: "reporter.generate-report.queue",
		RedisHost:                 "localhost:6379",
		ObjectStorageEndpoint:     "http://localhost:8333",
	}
}

func TestConfig_Validate_ValidConfig(t *testing.T) {
	t.Parallel()

	cfg := validManagerConfig()
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_AllFieldsMissing(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	err := cfg.Validate()
	require.Error(t, err)

	errMsg := err.Error()
	assert.Contains(t, errMsg, "config validation failed:")
	assert.Contains(t, errMsg, "SERVER_ADDRESS is required")
	assert.Contains(t, errMsg, "MONGO_HOST is required")
	assert.Contains(t, errMsg, "MONGO_NAME is required")
	assert.Contains(t, errMsg, "RABBITMQ_HOST is required")
	assert.Contains(t, errMsg, "RABBITMQ_PORT_AMQP is required")
	assert.Contains(t, errMsg, "RABBITMQ_DEFAULT_USER is required")
	assert.Contains(t, errMsg, "RABBITMQ_DEFAULT_PASS is required")
	assert.Contains(t, errMsg, "RABBITMQ_GENERATE_REPORT_QUEUE is required")
	assert.Contains(t, errMsg, "REDIS_HOST is required")
	assert.Contains(t, errMsg, "OBJECT_STORAGE_ENDPOINT is required")
}

func TestConfig_Validate_SingleFieldMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		modify       func(cfg *Config)
		expectedErr  string
	}{
		{
			name:        "missing ServerAddress",
			modify:      func(cfg *Config) { cfg.ServerAddress = "" },
			expectedErr: "SERVER_ADDRESS is required",
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
			name:        "missing RedisHost",
			modify:      func(cfg *Config) { cfg.RedisHost = "" },
			expectedErr: "REDIS_HOST is required",
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

			cfg := validManagerConfig()
			tt.modify(cfg)

			err := cfg.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestConfig_Validate_MultipleFieldsMissing(t *testing.T) {
	t.Parallel()

	cfg := validManagerConfig()
	cfg.ServerAddress = ""
	cfg.RedisHost = ""
	cfg.MongoDBHost = ""

	err := cfg.Validate()
	require.Error(t, err)

	errMsg := err.Error()
	assert.Contains(t, errMsg, "SERVER_ADDRESS is required")
	assert.Contains(t, errMsg, "REDIS_HOST is required")
	assert.Contains(t, errMsg, "MONGO_HOST is required")

	// Verify multi-line format
	lines := strings.Split(errMsg, "\n")
	assert.GreaterOrEqual(t, len(lines), 4) // header + 3 errors
}

func TestConfig_Validate_OptionalFieldsCanBeEmpty(t *testing.T) {
	t.Parallel()

	cfg := validManagerConfig()
	// These fields are optional and should not cause validation errors
	cfg.EnvName = ""
	cfg.LogLevel = ""
	cfg.MongoDBPassword = ""
	cfg.MongoDBParameters = ""
	cfg.RedisPassword = ""
	cfg.AuthAddress = ""

	err := cfg.Validate()
	assert.NoError(t, err)
}
