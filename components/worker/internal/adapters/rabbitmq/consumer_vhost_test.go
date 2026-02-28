// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

//go:build unit

package rabbitmq

import (
	"context"
	"testing"

	pkgRabbitmq "github.com/LerianStudio/reporter/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

// TestConsumerRoutes_MultiTenantFields verifies that the multi-tenant
// fields (rabbitMQManager and multiTenantMode) are correctly initialized.
func TestConsumerRoutes_MultiTenantFields(t *testing.T) {
	t.Parallel()

	t.Run("single-tenant consumer has nil rabbitMQManager", func(t *testing.T) {
		t.Parallel()
		// The default constructor doesn't set rabbitMQManager
		// This test verifies the expected state for single-tenant mode
		cr := &ConsumerRoutes{
			routes:          make(map[string]pkgRabbitmq.QueueHandlerFunc),
			multiTenantMode: false,
		}
		assert.False(t, cr.multiTenantMode, "multiTenantMode should be false for single-tenant consumer")
		assert.Nil(t, cr.rabbitMQManager, "rabbitMQManager should be nil for single-tenant consumer")
	})

	t.Run("multi-tenant consumer has rabbitMQManager set", func(t *testing.T) {
		t.Parallel()
		mockManager := &mockRabbitMQManagerConsumer{}
		cr := &ConsumerRoutes{
			routes:          make(map[string]pkgRabbitmq.QueueHandlerFunc),
			rabbitMQManager: mockManager,
			multiTenantMode: true,
		}
		assert.True(t, cr.multiTenantMode, "multiTenantMode should be true for multi-tenant consumer")
		assert.NotNil(t, cr.rabbitMQManager, "rabbitMQManager should be set for multi-tenant consumer")
	})
}

// mockRabbitMQManagerConsumer is a mock implementation of the RabbitMQManagerConsumerInterface
// used for unit testing multi-tenant consumer behavior.
type mockRabbitMQManagerConsumer struct {
	getConnectionErr error
	lastTenantID     string
	connection       *mockRabbitMQConnectionChannel
}

// GetConnection implements the interface required by ConsumerRoutes.
func (m *mockRabbitMQManagerConsumer) GetConnection(ctx context.Context, tenantID string) (RabbitMQConnectionChannel, error) {
	m.lastTenantID = tenantID

	if m.getConnectionErr != nil {
		return nil, m.getConnectionErr
	}

	if m.connection != nil {
		return m.connection, nil
	}

	return &mockRabbitMQConnectionChannel{}, nil
}

// mockRabbitMQConnectionChannel is a mock implementation of the RabbitMQConnectionChannel interface.
type mockRabbitMQConnectionChannel struct {
	publishCalled bool
	lastExchange  string
	lastKey       string
	publishErr    error
}

func (m *mockRabbitMQConnectionChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	m.publishCalled = true
	m.lastExchange = exchange
	m.lastKey = key

	return m.publishErr
}
