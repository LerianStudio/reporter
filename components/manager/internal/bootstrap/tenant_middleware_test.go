// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

//go:build unit

package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitTenantMiddleware_ReturnsNil_WhenDisabled(t *testing.T) {
	t.Parallel()
	// When MultiTenantEnabled=false, initTenantMiddleware must return nil
	// (no middleware registered, single-tenant passthrough)
	result := initTenantMiddlewareForTest(false, "")
	assert.Nil(t, result, "must return nil when multi-tenant is disabled")
}

func TestInitTenantMiddleware_ReturnsNil_WhenNoAddress(t *testing.T) {
	t.Parallel()
	result := initTenantMiddlewareForTest(true, "")
	assert.Nil(t, result, "must return nil when MultiTenantURL is empty even if enabled")
}

func TestInitTenantMiddleware_ReturnsNonNil_WhenEnabledWithAddress(t *testing.T) {
	t.Parallel()
	result := initTenantMiddlewareForTest(true, "http://tenant-manager:8080")
	assert.NotNil(t, result, "must return a handler when multi-tenant is enabled with an address")
}
