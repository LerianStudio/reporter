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

// TestProducer_HeadersContainTenantID_WhenTenantInContext verifies that when a tenant
// ID is present in the context, the X-Tenant-ID header is injected into the AMQP
// headers table before publishing. This is the core multi-tenant producer behaviour.
func TestProducer_HeadersContainTenantID_WhenTenantInContext(t *testing.T) {
	t.Parallel()

	ctx := tmcore.SetTenantIDInContext(context.Background(), "tenant-abc")
	tenantID := tmcore.GetTenantIDFromContext(ctx)

	// buildHeaders is the extraction-and-injection logic we will add to
	// ProducerDefault. We test the invariant here so the test will fail
	// RED until the production code sets headers["X-Tenant-ID"].
	headers := buildProducerHeaders(ctx, tenantID)

	val, ok := headers["X-Tenant-ID"]
	assert.True(t, ok, "X-Tenant-ID header must be present when tenant is in context")
	assert.Equal(t, "tenant-abc", val, "X-Tenant-ID header must equal the context tenant ID")
}

// TestProducer_NoTenantHeader_WhenNoTenantContext verifies that when no tenant ID is
// set in the context (single-tenant / legacy deployment), the X-Tenant-ID header is
// NOT added to the AMQP headers table. This preserves backward compatibility.
func TestProducer_NoTenantHeader_WhenNoTenantContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background() // no tenant set
	tenantID := tmcore.GetTenantIDFromContext(ctx)

	headers := buildProducerHeaders(ctx, tenantID)

	_, ok := headers["X-Tenant-ID"]
	assert.False(t, ok, "X-Tenant-ID header must NOT be present when no tenant is in context")
}

// TestProducer_EmptyTenantIDNotInjected verifies that an explicitly empty tenant ID
// (e.g. from a bug in upstream middleware) does not result in a blank X-Tenant-ID
// header being published. Only non-empty tenant IDs are injected.
func TestProducer_EmptyTenantIDNotInjected(t *testing.T) {
	t.Parallel()

	ctx := tmcore.SetTenantIDInContext(context.Background(), "")
	tenantID := tmcore.GetTenantIDFromContext(ctx)

	headers := buildProducerHeaders(ctx, tenantID)

	_, ok := headers["X-Tenant-ID"]
	assert.False(t, ok, "X-Tenant-ID header must NOT be present when tenant ID is empty string")
}
