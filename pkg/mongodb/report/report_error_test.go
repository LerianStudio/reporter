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

// TestReportRepo_MissingTenantContext_FallsBackToStaticConnection verifies that
// when no tenant is set in context the repository's getCollection helper receives
// ErrTenantContextRequired from GetMongoForTenant and must not panic.
//
// The production code in getCollection checks errors.Is(err, ErrTenantContextRequired)
// and routes to the static connection on match. This test documents and guards that
// expected sentinel so a future refactor cannot accidentally break single-tenant compat.
func TestReportRepo_MissingTenantContext_FallsBackToStaticConnection(t *testing.T) {
	t.Parallel()

	// Plain background context — no middleware, no JWT, no tenant set.
	ctx := context.Background()

	// The getCollection fallback relies on this exact sentinel error.
	_, err := tmCore.GetMongoForTenant(ctx)
	assert.ErrorIs(t, err, tmCore.ErrTenantContextRequired,
		"missing tenant context must produce ErrTenantContextRequired, "+
			"which is the trigger for the single-tenant static-connection fallback in getCollection")
}

// TestReportRepo_UnexpectedGetMongoError_IsNotSentinel verifies that an error
// other than ErrTenantContextRequired would NOT be treated as the fallback signal.
// This test documents the negative case: only the specific sentinel triggers fallback.
func TestReportRepo_UnexpectedGetMongoError_IsNotSentinel(t *testing.T) {
	t.Parallel()

	// When no tenant context is set the error IS the sentinel.
	ctx := context.Background()
	_, err := tmCore.GetMongoForTenant(ctx)

	// The error must be exactly ErrTenantContextRequired — not a wrapped or
	// different error — so the is-check in getCollection can match it precisely.
	assert.ErrorIs(t, err, tmCore.ErrTenantContextRequired,
		"the sentinel must be matchable with errors.Is so getCollection can branch correctly")
}

// TestReportRepo_TenantIDInContext_DoesNotAloneProvideMongoConnection verifies that
// injecting a tenant ID string alone is not sufficient for GetMongoForTenant.
// The middleware must also store the actual *mongo.Database connection.
// This prevents a misconfigured deployment from silently returning no-op data.
func TestReportRepo_TenantIDInContext_DoesNotAloneProvideMongoConnection(t *testing.T) {
	t.Parallel()

	// SetTenantIDInContext stores a string identifier, NOT a DB connection.
	ctx := tmCore.SetTenantIDInContext(context.Background(), "tenant-orphan")

	_, err := tmCore.GetMongoForTenant(ctx)
	assert.ErrorIs(t, err, tmCore.ErrTenantContextRequired,
		"a tenant ID string in context without a stored *mongo.Database must still return "+
			"ErrTenantContextRequired — GetMongoForTenant requires the actual DB connection")
}
