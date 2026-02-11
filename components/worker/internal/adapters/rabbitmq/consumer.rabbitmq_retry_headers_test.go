// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package rabbitmq

import (
	"errors"
	"testing"

	pkgConstant "github.com/LerianStudio/reporter/pkg/constant"

	amqp091 "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRetryHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		existingHeaders amqp091.Table
		retryCount      int
		lastErr         error
		wantRetryCount  int
		wantReason      string
	}{
		{
			name:            "first retry with nil headers",
			existingHeaders: nil,
			retryCount:      0,
			lastErr:         errors.New("connection timeout"),
			wantRetryCount:  1,
			wantReason:      "connection timeout",
		},
		{
			name: "subsequent retry preserves and increments",
			existingHeaders: amqp091.Table{
				pkgConstant.RetryCountHeader:         3,
				pkgConstant.RetryFailureReasonHeader: "previous error",
			},
			retryCount:     3,
			lastErr:        errors.New("network unreachable"),
			wantRetryCount: 4,
			wantReason:     "network unreachable",
		},
		{
			name: "preserves existing custom headers",
			existingHeaders: amqp091.Table{
				"x-custom-header": "custom-value",
				"x-trace-id":      "abc-123",
			},
			retryCount:     1,
			lastErr:        errors.New("temporary failure"),
			wantRetryCount: 2,
			wantReason:     "temporary failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headers := buildRetryHeaders(tt.existingHeaders, tt.retryCount, tt.lastErr)

			require.NotNil(t, headers)
			assert.Equal(t, tt.wantRetryCount, headers[pkgConstant.RetryCountHeader])
			assert.Equal(t, tt.wantReason, headers[pkgConstant.RetryFailureReasonHeader])
		})
	}
}

func TestBuildRetryHeaders_PreservesExistingHeaders(t *testing.T) {
	t.Parallel()

	existingHeaders := amqp091.Table{
		"x-custom-header": "custom-value",
		"x-trace-id":      "trace-abc-123",
		"x-org-id":        "org-456",
	}

	headers := buildRetryHeaders(existingHeaders, 2, errors.New("some error"))

	require.NotNil(t, headers)

	// Verify retry headers are set correctly
	assert.Equal(t, 3, headers[pkgConstant.RetryCountHeader])
	assert.Equal(t, "some error", headers[pkgConstant.RetryFailureReasonHeader])

	// Verify existing custom headers are preserved
	assert.Equal(t, "custom-value", headers["x-custom-header"])
	assert.Equal(t, "trace-abc-123", headers["x-trace-id"])
	assert.Equal(t, "org-456", headers["x-org-id"])
}
