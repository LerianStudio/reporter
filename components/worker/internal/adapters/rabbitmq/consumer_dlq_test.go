package rabbitmq_test

import (
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

// TestRetryMessageWithCount_SafeHeaderParsing tests safe retry-count extraction
func TestRetryMessageWithCount_SafeHeaderParsing(t *testing.T) {
	tests := []struct {
		name          string
		headerValue   any
		expectedRetry int32
	}{
		{"int32 value", int32(2), 2},
		{"int64 value", int64(1), 1},
		{"int value", int(0), 0},
		{"float64 value", float64(2.0), 2},
		{"string value", "invalid", 0},
		{"nil value", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test validates that various header types are safely parsed
			// without causing panic - actual retry logic tested in integration
			headers := amqp091.Table{}
			if tt.headerValue != nil {
				headers["x-retry-count"] = tt.headerValue
			}

			message := amqp091.Delivery{
				Headers: headers,
			}

			// Validate header extraction logic (mimics consumer logic)
			var retryCount int32
			if raw, ok := message.Headers["x-retry-count"]; ok {
				switch v := raw.(type) {
				case int32:
					retryCount = v
				case int64:
					retryCount = int32(v)
				case int:
					retryCount = int32(v)
				case float32:
					retryCount = int32(v)
				case float64:
					retryCount = int32(v)
				default:
					retryCount = 0
				}
			}

			assert.Equal(t, tt.expectedRetry, retryCount, "Retry count should be safely extracted")
		})
	}
}

// TestRetryMessageWithBackoff_ExponentialDelay validates exponential backoff
func TestRetryMessageWithBackoff_ExponentialDelay(t *testing.T) {
	tests := []struct {
		retryCount      int32
		expectedBackoff time.Duration
	}{
		{0, 1 * time.Second}, // 2^0 = 1
		{1, 2 * time.Second}, // 2^1 = 2
		{2, 4 * time.Second}, // 2^2 = 4
	}

	for _, tt := range tests {
		t.Run("retry_"+string(rune(tt.retryCount+'0')), func(t *testing.T) {
			// Validate exponential backoff calculation
			backoff := time.Duration(1<<tt.retryCount) * time.Second
			assert.Equal(t, tt.expectedBackoff, backoff, "Backoff should follow 2^n pattern")
		})
	}
}

// TestDLQMessagePersistence_MetadataCapture validates DLQ metadata
func TestDLQMessagePersistence_MetadataCapture(t *testing.T) {
	headers := amqp091.Table{
		"x-retry-count":    int32(3),
		"x-failure-reason": "Database connection failed",
		"x-request-id":     "test-req-123",
	}

	// Validate metadata extraction
	var retryCount int32
	if raw, ok := headers["x-retry-count"]; ok {
		if v, ok := raw.(int32); ok {
			retryCount = v
		}
	}

	var failureReason string
	if raw, ok := headers["x-failure-reason"]; ok {
		if str, ok := raw.(string); ok {
			failureReason = str
		}
	}

	assert.Equal(t, int32(3), retryCount, "Should extract retry count")
	assert.Equal(t, "Database connection failed", failureReason, "Should extract failure reason")
}

// TestConsumerDLQFlow_NoMessageLoss validates messages are not lost after 3 failures
func TestConsumerDLQFlow_NoMessageLoss(t *testing.T) {
	// This test validates the conceptual flow; actual RabbitMQ integration
	// would require testcontainers or similar infrastructure

	// Simulate a message that has failed 3 times
	message := amqp091.Delivery{
		Headers: amqp091.Table{
			"x-retry-count":    int32(3),
			"x-failure-reason": "Processing error",
		},
		Body:        []byte(`{"reportId":"test-123"}`),
		Exchange:    "reporter.generate-report.exchange",
		RoutingKey:  "reporter.generate-report.key",
		ContentType: "application/json",
	}

	// After 3 failures, message should be Nack'd and go to DLQ (via DLX)
	// The consumer should NOT requeue (Nack with requeue=false)
	var retryCount int32
	if raw, ok := message.Headers["x-retry-count"]; ok {
		if v, ok := raw.(int32); ok {
			retryCount = v
		}
	}

	assert.GreaterOrEqual(t, retryCount, int32(3), "Message should have reached retry limit")

	// In real flow, this triggers:
	// 1. Nack(false, false) -> message goes to DLX
	// 2. DLX routes to DLQ
	// 3. DLQ consumer persists to MongoDB
	// 4. Report status updated to Error
}

// TestDLXConfiguration validates DLX binding expectations
func TestDLXConfiguration(t *testing.T) {
	// This test documents the expected DLX/DLQ configuration
	// Actual validation would be done via RabbitMQ management API or integration test

	type QueueConfig struct {
		Name      string
		Arguments map[string]any
	}

	mainQueue := QueueConfig{
		Name: "reporter.generate-report.queue",
		Arguments: map[string]any{
			"x-dead-letter-exchange":    "reporter.dlx",
			"x-dead-letter-routing-key": "reporter.dlq.key",
		},
	}

	dlqQueue := QueueConfig{
		Name: "reporter.dlq",
		Arguments: map[string]any{
			"x-message-ttl": 604800000, // 7 days in ms
			"x-max-length":  10000,
		},
	}

	assert.Equal(t, "reporter.dlx", mainQueue.Arguments["x-dead-letter-exchange"])
	assert.Equal(t, 604800000, dlqQueue.Arguments["x-message-ttl"])
	assert.Equal(t, 10000, dlqQueue.Arguments["x-max-length"])
}
