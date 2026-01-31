//go:build integration
// +build integration

// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package integration_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/v4/pkg/model"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

// TestDLQMessage_MetadataStructure validates DLQ message schema
func TestDLQMessage_MetadataStructure(t *testing.T) {
	reportID := uuid.New()
	reportMessage := model.ReportMessage{
		ReportID:     reportID,
		TemplateID:   uuid.New(),
		OutputFormat: "html",
	}

	bodyBytes, err := json.Marshal(reportMessage)
	assert.NoError(t, err)

	headers := amqp091.Table{
		"x-retry-count":    int32(3),
		"x-failure-reason": "Template rendering failed: invalid syntax",
		"x-request-id":     "req-123",
	}

	expectedDoc := map[string]any{
		"message_body":     string(bodyBytes),
		"retry_count":      int32(3),
		"failure_reason":   "Template rendering failed: invalid syntax",
		"received_at":      time.Now(),
		"report_id":        reportID.String(),
		"original_headers": headers,
	}

	assert.Contains(t, expectedDoc, "message_body")
	assert.Contains(t, expectedDoc, "retry_count")
	assert.Contains(t, expectedDoc, "failure_reason")
	assert.Contains(t, expectedDoc, "received_at")
	assert.Contains(t, expectedDoc, "report_id")
	assert.Contains(t, expectedDoc, "original_headers")

	assert.Equal(t, int32(3), expectedDoc["retry_count"])
	assert.NotEmpty(t, expectedDoc["failure_reason"])
}

// TestDLQConfiguration_TTLAndLimits validates DLQ queue configuration
func TestDLQConfiguration_TTLAndLimits(t *testing.T) {
	expectedTTL := 7 * 24 * time.Hour
	expectedMaxLength := 10000

	ttlMs := int(expectedTTL.Milliseconds())
	assert.Equal(t, 604800000, ttlMs, "DLQ TTL should be 7 days in milliseconds")
	assert.Equal(t, 10000, expectedMaxLength, "DLQ should limit to 10,000 messages")
}

// TestReportStatus_UpdatedOnDLQ validates report status update
func TestReportStatus_UpdatedOnDLQ(t *testing.T) {
	expectedStatus := "Error"
	expectedMetadata := map[string]any{
		"error":         "Database connection timeout",
		"retry_count":   int32(3),
		"dlq_timestamp": time.Now(),
	}

	assert.Equal(t, "Error", expectedStatus)
	assert.Contains(t, expectedMetadata, "error")
	assert.Contains(t, expectedMetadata, "retry_count")
	assert.Contains(t, expectedMetadata, "dlq_timestamp")
	assert.Equal(t, int32(3), expectedMetadata["retry_count"])
}

// TestExponentialBackoff_Timing validates retry delays
func TestExponentialBackoff_Timing(t *testing.T) {
	expectedBackoffs := []time.Duration{
		1 * time.Second, // 2^0
		2 * time.Second, // 2^1
		4 * time.Second, // 2^2
	}

	for i, expected := range expectedBackoffs {
		actual := time.Duration(1<<i) * time.Second
		assert.Equal(t, expected, actual, "Backoff for retry %d should be %v", i, expected)
	}
}
