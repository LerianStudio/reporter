// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

//go:build unit

package redis_test

import (
	"context"
	"testing"

	tmCore "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/core"
	tmValkey "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/valkey"
	"github.com/stretchr/testify/assert"
)

// TestRedisRepo_TenantContext_KeyIsPrefixed verifies that when a tenant ID is present
// in context, GetKeyFromContext returns a key prefixed with the tenant namespace.
// The redis consumer uses this function on every Set/Get/Del call to enforce tenant
// isolation at the key level without requiring separate Redis instances per tenant.
func TestRedisRepo_TenantContext_KeyIsPrefixed(t *testing.T) {
	t.Parallel()

	tenantID := "tenant123"
	baseKey := "report:status"

	ctx := tmCore.SetTenantIDInContext(context.Background(), tenantID)
	result := tmValkey.GetKeyFromContext(ctx, baseKey)

	// The prefix format enforced by lib-commons is "tenant:{tenantID}:{key}".
	expected := tmValkey.GetKey(tenantID, baseKey)
	assert.Equal(t, expected, result,
		"GetKeyFromContext must prefix the key with the tenant namespace when tenant is in context")
	assert.Contains(t, result, tenantID,
		"the resulting key must contain the tenant ID")
	assert.Contains(t, result, baseKey,
		"the resulting key must still contain the original key")
}

// TestRedisRepo_NoTenantContext_KeyIsUnchanged verifies that when no tenant ID is set
// in context, GetKeyFromContext returns the key exactly as provided (backward compat).
// This is the single-tenant path: no prefix is added, no panic occurs.
func TestRedisRepo_NoTenantContext_KeyIsUnchanged(t *testing.T) {
	t.Parallel()

	baseKey := "report:status"
	ctx := context.Background() // no tenant set

	result := tmValkey.GetKeyFromContext(ctx, baseKey)

	assert.Equal(t, baseKey, result,
		"GetKeyFromContext must return the key unchanged when no tenant is in context (single-tenant mode)")
}

// TestRedisRepo_DifferentTenants_ProduceDifferentKeys verifies that two tenants with
// the same logical key produce different Redis keys, ensuring tenant isolation.
func TestRedisRepo_DifferentTenants_ProduceDifferentKeys(t *testing.T) {
	t.Parallel()

	baseKey := "idempotency:abc123"

	ctxA := tmCore.SetTenantIDInContext(context.Background(), "tenant-alpha")
	ctxB := tmCore.SetTenantIDInContext(context.Background(), "tenant-beta")

	keyA := tmValkey.GetKeyFromContext(ctxA, baseKey)
	keyB := tmValkey.GetKeyFromContext(ctxB, baseKey)

	assert.NotEqual(t, keyA, keyB,
		"the same logical key for different tenants must produce different Redis keys; "+
			"tenant-alpha and tenant-beta must not share cache entries")
}

// TestRedisRepo_SameTenant_ProducesSameKey verifies that the same tenant ID always
// produces the same key prefix, making Redis lookups deterministic across requests.
func TestRedisRepo_SameTenant_ProducesSameKey(t *testing.T) {
	t.Parallel()

	tenantID := "deterministic-tenant"
	baseKey := "session:token"

	ctx1 := tmCore.SetTenantIDInContext(context.Background(), tenantID)
	ctx2 := tmCore.SetTenantIDInContext(context.Background(), tenantID)

	key1 := tmValkey.GetKeyFromContext(ctx1, baseKey)
	key2 := tmValkey.GetKeyFromContext(ctx2, baseKey)

	assert.Equal(t, key1, key2,
		"the same tenant ID must always produce the same prefixed key for deterministic cache lookups")
}
