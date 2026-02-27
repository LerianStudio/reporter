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
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

// TestReportRepo_TenantContext_FlowsToMongo verifies the branching logic documented
// in getCollection: when a real *mongo.Database is stored via ContextWithTenantMongo,
// GetMongoForTenant must succeed so the repository can use the tenant-scoped connection
// instead of the static one.
func TestReportRepo_TenantContext_FlowsToMongo(t *testing.T) {
	t.Parallel()

	// Construct a minimal *mongo.Database client just to satisfy the context storage.
	// We use mongo.NewClient only to obtain a typed value; no connection is made.
	client, err := mongo.NewClient()
	require.NoError(t, err, "mongo.NewClient must not error in unit test setup")

	db := client.Database("tenant_test_db")

	ctx := tmCore.ContextWithTenantMongo(context.Background(), db)

	// GetMongoForTenant must succeed when a non-nil *mongo.Database is in context.
	got, err := tmCore.GetMongoForTenant(ctx)
	require.NoError(t, err,
		"GetMongoForTenant must succeed when a real *mongo.Database is in context (multi-tenant path)")
	assert.NotNil(t, got, "returned *mongo.Database must not be nil")
	assert.Equal(t, db, got, "returned DB must be the same instance stored in context")
}

// TestReportRepo_NoTenantContext_FallsBackToStaticConnection verifies that when no
// tenant context is set, GetMongoForTenant returns ErrTenantContextRequired.
// The repository getCollection function catches exactly this sentinel to fall back
// to the static MongoConnection — the backward-compatible single-tenant path.
func TestReportRepo_NoTenantContext_FallsBackToStaticConnection(t *testing.T) {
	t.Parallel()

	ctx := context.Background() // no tenant set

	_, err := tmCore.GetMongoForTenant(ctx)
	assert.ErrorIs(t, err, tmCore.ErrTenantContextRequired,
		"GetMongoForTenant must return ErrTenantContextRequired when no tenant context is set; "+
			"this sentinel triggers the static-connection fallback in getCollection")
}

// TestReportRepo_TenantContextIsolation verifies that two independent contexts with
// different tenant databases do not leak into one another.
func TestReportRepo_TenantContextIsolation(t *testing.T) {
	t.Parallel()

	client, err := mongo.NewClient()
	require.NoError(t, err)

	dbA := client.Database("tenant_a_db")
	dbB := client.Database("tenant_b_db")

	ctxA := tmCore.ContextWithTenantMongo(context.Background(), dbA)
	ctxB := tmCore.ContextWithTenantMongo(context.Background(), dbB)

	gotA, errA := tmCore.GetMongoForTenant(ctxA)
	gotB, errB := tmCore.GetMongoForTenant(ctxB)

	require.NoError(t, errA)
	require.NoError(t, errB)

	assert.Equal(t, dbA, gotA, "context A must return tenant A database")
	assert.Equal(t, dbB, gotB, "context B must return tenant B database")
	assert.NotEqual(t, gotA, gotB,
		"tenant A and tenant B must resolve to distinct database instances")
}

// TestReportRepo_TenantContext_NilDB_FallsBack verifies that storing a nil
// *mongo.Database in context does not satisfy GetMongoForTenant — the repository
// must still fall back to the static connection rather than panicking downstream.
func TestReportRepo_TenantContext_NilDB_FallsBack(t *testing.T) {
	t.Parallel()

	ctx := tmCore.ContextWithTenantMongo(context.Background(), nil)

	_, err := tmCore.GetMongoForTenant(ctx)
	assert.ErrorIs(t, err, tmCore.ErrTenantContextRequired,
		"a nil *mongo.Database stored in context must not satisfy GetMongoForTenant; "+
			"fallback to static connection must be preserved")
}
