// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

//go:build unit

package rabbitmq

import (
	"context"
	"testing"

	tmcore "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/core"
	amqp091 "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

// TestConsumer_Tenant_ExtractedFromXTenantIDHeader verifies that extractTenantIDFromHeaders
// reads the X-Tenant-ID header and stores the tenant ID in the returned context so that
// downstream repository calls can route to the correct tenant database.
func TestConsumer_Tenant_ExtractedFromXTenantIDHeader(t *testing.T) {
	t.Parallel()

	headers := amqp091.Table{
		"X-Tenant-ID": "tenant-from-producer",
	}

	ctx := extractTenantIDFromHeaders(context.Background(), headers)

	tenantID := tmcore.GetTenantIDFromContext(ctx)
	assert.Equal(t, "tenant-from-producer", tenantID,
		"tenant ID from X-Tenant-ID header must be stored in context for downstream repository use")
}

// TestConsumer_Tenant_NotInjectedWhenHeaderAbsent verifies that when the AMQP message
// has no X-Tenant-ID header (legacy single-tenant message), the context is returned
// unchanged and no tenant ID is injected. This preserves backward compatibility.
func TestConsumer_Tenant_NotInjectedWhenHeaderAbsent(t *testing.T) {
	t.Parallel()

	headers := amqp091.Table{
		"x-request-id": "some-request-id",
	}

	ctx := extractTenantIDFromHeaders(context.Background(), headers)

	tenantID := tmcore.GetTenantIDFromContext(ctx)
	assert.Empty(t, tenantID,
		"tenant ID must be empty when X-Tenant-ID header is absent (single-tenant backward compat)")
}

// TestConsumer_Tenant_NotInjectedWhenHeaderIsNilTable verifies that a nil headers
// table does not panic and leaves context without a tenant ID.
func TestConsumer_Tenant_NotInjectedWhenHeaderIsNilTable(t *testing.T) {
	t.Parallel()

	ctx := extractTenantIDFromHeaders(context.Background(), nil)

	tenantID := tmcore.GetTenantIDFromContext(ctx)
	assert.Empty(t, tenantID,
		"nil headers table must not panic and must leave tenant ID empty")
}

// TestConsumer_Tenant_MultipleMessages_IndependentContexts verifies that processing
// two messages in sequence results in independent contexts — tenant A's context must
// not bleed into tenant B's processing.
func TestConsumer_Tenant_MultipleMessages_IndependentContexts(t *testing.T) {
	t.Parallel()

	headersA := amqp091.Table{"X-Tenant-ID": "tenant-A"}
	headersB := amqp091.Table{"X-Tenant-ID": "tenant-B"}

	ctxA := extractTenantIDFromHeaders(context.Background(), headersA)
	ctxB := extractTenantIDFromHeaders(context.Background(), headersB)

	tenantA := tmcore.GetTenantIDFromContext(ctxA)
	tenantB := tmcore.GetTenantIDFromContext(ctxB)

	assert.Equal(t, "tenant-A", tenantA, "context A must contain tenant-A")
	assert.Equal(t, "tenant-B", tenantB, "context B must contain tenant-B")
	assert.NotEqual(t, tenantA, tenantB,
		"tenant contexts from independent messages must not leak into each other")
}
