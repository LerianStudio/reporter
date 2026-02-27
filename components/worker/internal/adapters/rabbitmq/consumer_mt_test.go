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

// TestConsumer_ExtractsTenantID_FromHeaders verifies that when an AMQP message
// carries an X-Tenant-ID header, the tenant ID is correctly extracted and stored
// in the context so downstream repository calls can use tenant-scoped connections.
func TestConsumer_ExtractsTenantID_FromHeaders(t *testing.T) {
	t.Parallel()

	headers := amqp091.Table{
		"X-Tenant-ID": "tenant-xyz",
	}

	ctx := extractTenantIDFromHeaders(context.Background(), headers)

	tenantID := tmcore.GetTenantIDFromContext(ctx)
	assert.Equal(t, "tenant-xyz", tenantID, "tenant ID must be propagated from AMQP header into context")
}

// TestConsumer_NoTenantInContext_WhenHeaderAbsent verifies that when an AMQP message
// does not carry an X-Tenant-ID header (single-tenant / legacy mode), the context
// remains unchanged and no tenant ID is injected. This preserves backward compat.
func TestConsumer_NoTenantInContext_WhenHeaderAbsent(t *testing.T) {
	t.Parallel()

	headers := amqp091.Table{
		"x-retry-count": 0,
	}

	ctx := extractTenantIDFromHeaders(context.Background(), headers)

	tenantID := tmcore.GetTenantIDFromContext(ctx)
	assert.Equal(t, "", tenantID, "tenant ID must be empty when X-Tenant-ID header is absent")
}

// TestConsumer_NoTenantInContext_WhenHeaderEmpty verifies that a blank X-Tenant-ID
// header value does not inject an empty string into the context.
func TestConsumer_NoTenantInContext_WhenHeaderEmpty(t *testing.T) {
	t.Parallel()

	headers := amqp091.Table{
		"X-Tenant-ID": "",
	}

	ctx := extractTenantIDFromHeaders(context.Background(), headers)

	tenantID := tmcore.GetTenantIDFromContext(ctx)
	assert.Equal(t, "", tenantID, "an empty X-Tenant-ID header must not inject a blank tenant ID")
}

// TestConsumer_NoTenantInContext_WhenHeaderWrongType verifies that a non-string
// X-Tenant-ID header value (malformed) does not cause a panic and leaves the
// context without a tenant ID.
func TestConsumer_NoTenantInContext_WhenHeaderWrongType(t *testing.T) {
	t.Parallel()

	headers := amqp091.Table{
		"X-Tenant-ID": 12345, // wrong type: int instead of string
	}

	ctx := extractTenantIDFromHeaders(context.Background(), headers)

	tenantID := tmcore.GetTenantIDFromContext(ctx)
	assert.Equal(t, "", tenantID, "a non-string X-Tenant-ID header must not crash or inject garbage")
}
