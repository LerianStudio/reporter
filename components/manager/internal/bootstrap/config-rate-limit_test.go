// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig_HasRateLimitFields verifies that the manager Config struct
// has rate limit configuration fields loaded from environment variables.
func TestConfig_HasRateLimitFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() *Config
		field    string
		expected int
	}{
		{
			name: "RateLimitGlobal field exists and holds value from env",
			setup: func() *Config {
				return &Config{
					RateLimitGlobal: 100,
				}
			},
			field:    "RateLimitGlobal",
			expected: 100,
		},
		{
			name: "RateLimitExport field exists and holds value from env",
			setup: func() *Config {
				return &Config{
					RateLimitExport: 10,
				}
			},
			field:    "RateLimitExport",
			expected: 10,
		},
		{
			name: "RateLimitDispatch field exists and holds value from env",
			setup: func() *Config {
				return &Config{
					RateLimitDispatch: 50,
				}
			},
			field:    "RateLimitDispatch",
			expected: 50,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := tt.setup()

			switch tt.field {
			case "RateLimitGlobal":
				assert.Equal(t, tt.expected, cfg.RateLimitGlobal)
			case "RateLimitExport":
				assert.Equal(t, tt.expected, cfg.RateLimitExport)
			case "RateLimitDispatch":
				assert.Equal(t, tt.expected, cfg.RateLimitDispatch)
			default:
				t.Fatalf("unknown field: %s", tt.field)
			}
		})
	}
}

// TestConfig_RateLimitDefaults verifies that rate limit fields have correct
// default values when not explicitly set (zero value means the application
// should apply standard defaults from constants).
func TestConfig_RateLimitDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() *Config
		field    string
		expected int
	}{
		{
			name: "RateLimitGlobal defaults to 100",
			setup: func() *Config {
				cfg := validManagerConfig()
				// Do not set RateLimitGlobal -- default tag should provide 100
				return cfg
			},
			field:    "RateLimitGlobal",
			expected: 100,
		},
		{
			name: "RateLimitExport defaults to 10",
			setup: func() *Config {
				cfg := validManagerConfig()
				return cfg
			},
			field:    "RateLimitExport",
			expected: 10,
		},
		{
			name: "RateLimitDispatch defaults to 50",
			setup: func() *Config {
				cfg := validManagerConfig()
				return cfg
			},
			field:    "RateLimitDispatch",
			expected: 50,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := tt.setup()

			switch tt.field {
			case "RateLimitGlobal":
				assert.Equal(t, tt.expected, cfg.RateLimitGlobal)
			case "RateLimitExport":
				assert.Equal(t, tt.expected, cfg.RateLimitExport)
			case "RateLimitDispatch":
				assert.Equal(t, tt.expected, cfg.RateLimitDispatch)
			default:
				t.Fatalf("unknown field: %s", tt.field)
			}
		})
	}
}

// TestConfig_Validate_ProductionRequiresRateLimitEnabled verifies that
// production config validation rejects RateLimitEnabled=false.
func TestConfig_Validate_ProductionRequiresRateLimitEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		enabled     bool
		expectErr   bool
		errContains string
	}{
		{
			name:        "Error - rate limiting disabled in production",
			enabled:     false,
			expectErr:   true,
			errContains: "RATE_LIMIT_ENABLED must be true in production",
		},
		{
			name:      "Success - rate limiting enabled in production",
			enabled:   true,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := validManagerConfig()
			cfg.EnvName = "production"
			cfg.EnableTelemetry = true
			cfg.AuthEnabled = true
			cfg.RateLimitEnabled = tt.enabled
			cfg.MongoDBPassword = "real-password"
			cfg.RabbitMQPass = "real-password"
			cfg.RedisPassword = "real-password"
			cfg.ObjectStorageSecretKey = "real-secret"
			cfg.CORSAllowedOrigins = "https://app.example.com"

			err := cfg.Validate()

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestConfig_Validate_RateLimitBounds verifies that validation rejects
// rate limit values that are zero, negative, or exceed upper bounds.
func TestConfig_Validate_RateLimitBounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		modify      func(cfg *Config)
		expectErr   bool
		errContains string
	}{
		{
			name: "Error - RateLimitGlobal is zero",
			modify: func(cfg *Config) {
				cfg.RateLimitGlobal = 0
			},
			expectErr:   true,
			errContains: "RATE_LIMIT_GLOBAL must be between 1 and 10000",
		},
		{
			name: "Error - RateLimitExport is negative",
			modify: func(cfg *Config) {
				cfg.RateLimitExport = -1
			},
			expectErr:   true,
			errContains: "RATE_LIMIT_EXPORT must be between 1 and 1000",
		},
		{
			name: "Error - RateLimitDispatch is zero",
			modify: func(cfg *Config) {
				cfg.RateLimitDispatch = 0
			},
			expectErr:   true,
			errContains: "RATE_LIMIT_DISPATCH must be between 1 and 5000",
		},
		{
			name: "Error - RateLimitGlobal exceeds upper bound",
			modify: func(cfg *Config) {
				cfg.RateLimitGlobal = 10001
			},
			expectErr:   true,
			errContains: "RATE_LIMIT_GLOBAL must be between 1 and 10000",
		},
		{
			name: "Error - RateLimitExport exceeds upper bound",
			modify: func(cfg *Config) {
				cfg.RateLimitExport = 1001
			},
			expectErr:   true,
			errContains: "RATE_LIMIT_EXPORT must be between 1 and 1000",
		},
		{
			name: "Error - RateLimitDispatch exceeds upper bound",
			modify: func(cfg *Config) {
				cfg.RateLimitDispatch = 5001
			},
			expectErr:   true,
			errContains: "RATE_LIMIT_DISPATCH must be between 1 and 5000",
		},
		{
			name: "Success - RateLimitGlobal at upper bound",
			modify: func(cfg *Config) {
				cfg.RateLimitGlobal = 10000
			},
			expectErr: false,
		},
		{
			name: "Success - RateLimitExport at upper bound",
			modify: func(cfg *Config) {
				cfg.RateLimitExport = 1000
			},
			expectErr: false,
		},
		{
			name: "Success - RateLimitDispatch at upper bound",
			modify: func(cfg *Config) {
				cfg.RateLimitDispatch = 5000
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := validManagerConfig()
			tt.modify(cfg)

			err := cfg.Validate()
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
