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

// TestRedisKey_WithTenant_HasPrefix verifies that GetKeyFromContext prefixes a key
// with the tenant namespace when a tenant ID is present in context.
func TestRedisKey_WithTenant_HasPrefix(t *testing.T) {
	t.Parallel()

	ctx := tmCore.SetTenantIDInContext(context.Background(), "acme-corp")
	key := tmValkey.GetKeyFromContext(ctx, "report:abc123")

	// lib-commons GetKey format: "tenant:{tenantID}:{key}"
	expected := tmValkey.GetKey("acme-corp", "report:abc123")
	assert.Equal(t, expected, key,
		"GetKeyFromContext must prefix the key with the tenant namespace")
}

// TestRedisKey_WithoutTenant_NoPrefix verifies that GetKeyFromContext returns the key
// unchanged when no tenant ID is in context, preserving backward compatibility.
func TestRedisKey_WithoutTenant_NoPrefix(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	key := tmValkey.GetKeyFromContext(ctx, "report:abc123")

	assert.Equal(t, "report:abc123", key,
		"GetKeyFromContext must return the key unchanged when no tenant is in context")
}
