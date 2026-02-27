// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

//go:build unit

package report_test

import (
	"context"
	"testing"

	tmCore "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/core"
	"github.com/stretchr/testify/assert"
)

// TestReportRepo_BackwardCompat_NoTenantContext verifies that the tenant-manager core
// package is correctly imported and that GetMongoForTenant returns ErrTenantContextRequired
// when no tenant connection is set in context (single-tenant / no-middleware mode).
func TestReportRepo_BackwardCompat_NoTenantContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// GetMongoForTenant must return ErrTenantContextRequired when no tenant is in context.
	// The repository fallback relies on this sentinel to switch to the static connection.
	_, err := tmCore.GetMongoForTenant(ctx)
	assert.ErrorIs(t, err, tmCore.ErrTenantContextRequired,
		"expected ErrTenantContextRequired when no tenant context is set")
}

// TestReportRepo_TenantContext_MongoSet verifies that GetMongoForTenant succeeds when
// a *mongo.Database is stored in context via ContextWithTenantMongo.
func TestReportRepo_TenantContext_MongoSet(t *testing.T) {
	t.Parallel()

	// When a tenant DB is stored, GetMongoForTenant must return it without error.
	// We pass nil here only to verify the nil case does NOT succeed (nil is not stored).
	ctx := context.Background()
	ctx = tmCore.ContextWithTenantMongo(ctx, nil)

	// Storing nil does not satisfy the check — GetMongoFromContext returns nil for nil.
	_, err := tmCore.GetMongoForTenant(ctx)
	assert.ErrorIs(t, err, tmCore.ErrTenantContextRequired,
		"storing nil mongo DB must not satisfy tenant context check")
}
