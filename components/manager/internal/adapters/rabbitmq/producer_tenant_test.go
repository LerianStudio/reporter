// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

//go:build unit

package rabbitmq

import (
	"context"
	"testing"

	tmcore "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/core"
	"github.com/stretchr/testify/assert"
)

// TestProducer_Tenant_HeaderInjectedIntoBuildProducerHeaders verifies that
// SetTenantIDInContext stores a tenant ID retrievable by GetTenantIDFromContext,
// and that buildProducerHeaders injects it as X-Tenant-ID when non-empty.
func TestProducer_Tenant_HeaderInjectedIntoBuildProducerHeaders(t *testing.T) {
	t.Parallel()

	ctx := tmcore.SetTenantIDInContext(context.Background(), "tenant123")

	// Verify round-trip: what we set is what GetTenantIDFromContext returns.
	tenantID := tmcore.GetTenantIDFromContext(ctx)
	assert.Equal(t, "tenant123", tenantID,
		"GetTenantIDFromContext must return the value stored by SetTenantIDInContext")

	headers := buildProducerHeaders(ctx, tenantID)

	val, ok := headers["X-Tenant-ID"]
	assert.True(t, ok, "X-Tenant-ID must be present in headers when tenant is in context")
	assert.Equal(t, "tenant123", val, "X-Tenant-ID must equal the tenant ID from context")
}

// TestProducer_Tenant_HeaderAbsentWithoutContext verifies that when no tenant ID is
// stored in context, buildProducerHeaders does NOT add X-Tenant-ID to headers.
// This preserves backward compatibility with single-tenant deployments.
func TestProducer_Tenant_HeaderAbsentWithoutContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := tmcore.GetTenantIDFromContext(ctx)

	assert.Empty(t, tenantID,
		"GetTenantIDFromContext must return empty string when no tenant is in context")

	headers := buildProducerHeaders(ctx, tenantID)

	_, ok := headers["X-Tenant-ID"]
	assert.False(t, ok,
		"X-Tenant-ID must NOT be present in headers when no tenant is in context (backward compat)")
}

// TestProducer_Tenant_ContextRoundTrip verifies the full set/get round-trip for the
// tenant ID via the lib-commons API. This documents the contract that the producer
// depends on: whatever the HTTP middleware stores, the producer can read.
func TestProducer_Tenant_ContextRoundTrip(t *testing.T) {
	t.Parallel()

	tenants := []string{"org-abc", "org-xyz", "acme-corp", "tenant-001"}

	for _, tenant := range tenants {
		tenant := tenant
		t.Run(tenant, func(t *testing.T) {
			t.Parallel()

			ctx := tmcore.SetTenantIDInContext(context.Background(), tenant)
			got := tmcore.GetTenantIDFromContext(ctx)

			assert.Equal(t, tenant, got,
				"tenant ID round-trip must be exact for tenant %s", tenant)
		})
	}
}
