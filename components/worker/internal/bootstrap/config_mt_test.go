// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

//go:build unit

package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerConfig_MultiTenantEnabled_DefaultFalse(t *testing.T) {
	t.Parallel()
	cfg := Config{}
	assert.False(t, cfg.MultiTenantEnabled)
}

func TestWorkerConfig_MultiTenant_ValidWithoutURLWhenDisabled(t *testing.T) {
	t.Parallel()
	// Documents that a fully-configured worker validates cleanly with
	// MultiTenantEnabled=false and MultiTenantURL empty — the backward
	// compatibility contract: no tenant vars required in single-tenant mode.
	cfg := validWorkerConfig()
	cfg.MultiTenantEnabled = false
	cfg.MultiTenantURL = ""

	assert.NoError(t, cfg.Validate(),
		"worker config must validate with MultiTenantEnabled=false (no tenant vars required)")
}

func TestWorkerConfig_MultiTenant_ErrorWhenEnabledWithoutURL(t *testing.T) {
	t.Parallel()
	cfg := validWorkerConfig()
	cfg.MultiTenantEnabled = true
	cfg.MultiTenantURL = ""
	cfg.MultiTenantCircuitBreakerThreshold = 5
	cfg.MultiTenantCircuitBreakerTimeoutSec = 30

	err := cfg.Validate()
	require.Error(t, err, "Validate() must return error when MultiTenantEnabled=true and MultiTenantURL is empty")
	assert.Contains(t, err.Error(), "MULTI_TENANT_URL is required when MULTI_TENANT_ENABLED=true")
}

func TestWorkerConfig_MultiTenant_ValidWhenEnabledWithURL(t *testing.T) {
	t.Parallel()
	cfg := validWorkerConfig()
	cfg.MultiTenantEnabled = true
	cfg.MultiTenantURL = "http://tenant-manager:8080"
	cfg.MultiTenantCircuitBreakerThreshold = 5
	cfg.MultiTenantCircuitBreakerTimeoutSec = 30

	assert.NoError(t, cfg.Validate(),
		"Validate() must pass when MultiTenantEnabled=true and MultiTenantURL is set")
}

func TestWorkerConfig_MultiTenant_ErrorWhenCircuitBreakerThresholdZero(t *testing.T) {
	t.Parallel()
	cfg := validWorkerConfig()
	cfg.MultiTenantEnabled = true
	cfg.MultiTenantURL = "http://tenant-manager:8080"
	cfg.MultiTenantCircuitBreakerThreshold = 0

	err := cfg.Validate()
	require.Error(t, err,
		"Validate() must return error when MultiTenantEnabled=true and CircuitBreakerThreshold=0")
	assert.Contains(t, err.Error(),
		"MULTI_TENANT_CIRCUIT_BREAKER_THRESHOLD must be > 0 when MULTI_TENANT_ENABLED=true")
}

func TestWorkerConfig_MultiTenant_ErrorWhenCircuitBreakerTimeoutZero(t *testing.T) {
	t.Parallel()
	cfg := validWorkerConfig()
	cfg.MultiTenantEnabled = true
	cfg.MultiTenantURL = "http://tenant-manager:8080"
	cfg.MultiTenantCircuitBreakerThreshold = 5
	cfg.MultiTenantCircuitBreakerTimeoutSec = 0

	err := cfg.Validate()
	require.Error(t, err,
		"Validate() must return error when CircuitBreakerThreshold>0 and CircuitBreakerTimeoutSec=0")
	assert.Contains(t, err.Error(),
		"MULTI_TENANT_CIRCUIT_BREAKER_TIMEOUT_SEC must be > 0 when MULTI_TENANT_CIRCUIT_BREAKER_THRESHOLD > 0")
}

func TestWorkerConfig_MultiTenant_CanonicalFieldsExist(t *testing.T) {
	t.Parallel()
	// Verify all 7 canonical multi-tenant fields exist and have correct zero/default values.
	cfg := Config{}
	assert.False(t, cfg.MultiTenantEnabled)
	assert.Empty(t, cfg.MultiTenantURL)
	assert.Empty(t, cfg.MultiTenantEnvironment)
	assert.Zero(t, cfg.MultiTenantMaxTenantPools)
	assert.Zero(t, cfg.MultiTenantIdleTimeoutSec)
	assert.Zero(t, cfg.MultiTenantCircuitBreakerThreshold)
	assert.Zero(t, cfg.MultiTenantCircuitBreakerTimeoutSec)
}
