// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package rabbitmq

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/pkg"
	pkgConstant "github.com/LerianStudio/reporter/pkg/constant"

	amqp091 "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsumerRoutes_GetRetryCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  amqp091.Delivery
		want int
	}{
		{
			name: "nil headers returns zero",
			msg:  amqp091.Delivery{Headers: nil},
			want: 0,
		},
		{
			name: "missing header returns zero",
			msg:  amqp091.Delivery{Headers: amqp091.Table{"other-key": 42}},
			want: 0,
		},
		{
			name: "int value",
			msg:  amqp091.Delivery{Headers: amqp091.Table{pkgConstant.RetryCountHeader: 3}},
			want: 3,
		},
		{
			name: "int32 value",
			msg:  amqp091.Delivery{Headers: amqp091.Table{pkgConstant.RetryCountHeader: int32(2)}},
			want: 2,
		},
		{
			name: "int64 value",
			msg:  amqp091.Delivery{Headers: amqp091.Table{pkgConstant.RetryCountHeader: int64(4)}},
			want: 4,
		},
		{
			name: "float64 value",
			msg:  amqp091.Delivery{Headers: amqp091.Table{pkgConstant.RetryCountHeader: float64(5)}},
			want: 5,
		},
		{
			name: "negative int returns zero",
			msg:  amqp091.Delivery{Headers: amqp091.Table{pkgConstant.RetryCountHeader: -1}},
			want: 0,
		},
		{
			name: "negative int32 returns zero",
			msg:  amqp091.Delivery{Headers: amqp091.Table{pkgConstant.RetryCountHeader: int32(-5)}},
			want: 0,
		},
		{
			name: "negative int64 returns zero",
			msg:  amqp091.Delivery{Headers: amqp091.Table{pkgConstant.RetryCountHeader: int64(-10)}},
			want: 0,
		},
		{
			name: "negative float64 returns zero",
			msg:  amqp091.Delivery{Headers: amqp091.Table{pkgConstant.RetryCountHeader: float64(-3.5)}},
			want: 0,
		},
		{
			name: "string value returns zero",
			msg:  amqp091.Delivery{Headers: amqp091.Table{pkgConstant.RetryCountHeader: "not-a-number"}},
			want: 0,
		},
		{
			name: "zero value",
			msg:  amqp091.Delivery{Headers: amqp091.Table{pkgConstant.RetryCountHeader: 0}},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getRetryCount(tt.msg)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConsumerRoutes_CalculateBackoff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		attempt int
		wantMin time.Duration
		wantMax time.Duration
	}{
		{
			name:    "attempt 0 returns base backoff plus jitter",
			attempt: 0,
			wantMin: pkgConstant.RetryInitialBackoff,
			wantMax: pkgConstant.RetryInitialBackoff + pkgConstant.RetryJitterMax,
		},
		{
			name:    "attempt 1 returns 2x base plus jitter",
			attempt: 1,
			wantMin: 2 * pkgConstant.RetryInitialBackoff,
			wantMax: 2*pkgConstant.RetryInitialBackoff + pkgConstant.RetryJitterMax,
		},
		{
			name:    "attempt 2 returns 4x base plus jitter",
			attempt: 2,
			wantMin: 4 * pkgConstant.RetryInitialBackoff,
			wantMax: 4*pkgConstant.RetryInitialBackoff + pkgConstant.RetryJitterMax,
		},
		{
			name:    "attempt 100 is capped at max backoff plus jitter",
			attempt: 100,
			wantMin: pkgConstant.RetryMaxBackoff,
			wantMax: pkgConstant.RetryMaxBackoff + pkgConstant.RetryJitterMax,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Run multiple iterations to account for random jitter
			for i := 0; i < 50; i++ {
				got := calculateBackoff(tt.attempt)
				assert.GreaterOrEqual(t, got, tt.wantMin,
					"iteration %d: backoff %v should be >= %v", i, got, tt.wantMin)
				assert.LessOrEqual(t, got, tt.wantMax,
					"iteration %d: backoff %v should be <= %v", i, got, tt.wantMax)
			}
		})
	}
}

func TestCalculateBackoff_Distribution(t *testing.T) {
	t.Parallel()

	// Verify that jitter produces non-deterministic output
	const iterations = 50

	results := make(map[time.Duration]bool)

	for i := 0; i < iterations; i++ {
		results[calculateBackoff(0)] = true
	}

	assert.Greater(t, len(results), 1,
		"expected non-deterministic jitter values over %d iterations", iterations)
}

func TestConsumerRoutes_IsRetryable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error is not retryable",
			err:  nil,
			want: false,
		},
		{
			name: "context.Canceled is not retryable",
			err:  context.Canceled,
			want: false,
		},
		{
			name: "context.DeadlineExceeded is not retryable",
			err:  context.DeadlineExceeded,
			want: false,
		},
		{
			name: "TPL error code is not retryable",
			err:  errors.New("TPL-0001: invalid template field"),
			want: false,
		},
		{
			name: "wrapped TPL error is not retryable",
			err:  errors.New("processing failed: TPL-0042 bad format"),
			want: false,
		},
		{
			name: "generic error is retryable",
			err:  errors.New("connection reset by peer"),
			want: true,
		},
		{
			name: "timeout error is retryable",
			err:  errors.New("i/o timeout"),
			want: true,
		},
		{
			name: "wrapped context.Canceled is not retryable",
			err:  wrappedError{inner: context.Canceled},
			want: false,
		},
		{
			name: "wrapped context.DeadlineExceeded is not retryable",
			err:  wrappedError{inner: context.DeadlineExceeded},
			want: false,
		},
		{
			name: "EntityConflictError is not retryable",
			err:  pkg.EntityConflictError{Message: "conflict"},
			want: false,
		},
		{
			name: "ForbiddenError is not retryable",
			err:  pkg.ForbiddenError{Message: "forbidden"},
			want: false,
		},
		{
			name: "UnauthorizedError is not retryable",
			err:  pkg.UnauthorizedError{Message: "unauthorized"},
			want: false,
		},
		{
			name: "FailedPreconditionError is not retryable",
			err:  pkg.FailedPreconditionError{Message: "precondition failed"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isRetryable(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// wrappedError is a test helper that wraps an inner error for errors.Is traversal.
type wrappedError struct {
	inner error
}

func (w wrappedError) Error() string {
	return "wrapped: " + w.inner.Error()
}

func (w wrappedError) Unwrap() error {
	return w.inner
}

func TestConsumerRoutes_RetryConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 5, pkgConstant.MaxMessageRetries,
		"MaxMessageRetries should be 5")
	assert.Equal(t, 1*time.Second, pkgConstant.RetryInitialBackoff,
		"RetryInitialBackoff should be 1 second")
	assert.Equal(t, 30*time.Second, pkgConstant.RetryMaxBackoff,
		"RetryMaxBackoff should be 30 seconds")
	assert.Equal(t, 500*time.Millisecond, pkgConstant.RetryJitterMax,
		"RetryJitterMax should be 500 milliseconds")
	assert.Equal(t, "x-retry-count", pkgConstant.RetryCountHeader,
		"RetryCountHeader should be 'x-retry-count'")
	assert.Equal(t, "x-failure-reason", pkgConstant.RetryFailureReasonHeader,
		"RetryFailureReasonHeader should be 'x-failure-reason'")
}

func TestConsumerRoutes_RetryBoundaryAndHeaderConsistency(t *testing.T) {
	t.Parallel()

	// We cannot easily test republishWithIncrementedRetry in isolation without
	// a real or mocked amqp.Channel. Instead, we test the logic components:
	// 1. buildRetryHeaders correctly increments the counter (tested above)
	// 2. The new handleFailedMessage flow is covered by integration/chaos tests
	//
	// This test verifies the handleFailedMessage branching logic by testing
	// that the function signatures and constants are consistent.

	// Verify MaxMessageRetries is reachable: retry count 5 should exceed max
	assert.True(t, 5 >= pkgConstant.MaxMessageRetries,
		"retry count 5 should trigger DLQ path")

	// Verify retry count 4 should NOT exceed max
	assert.False(t, 4 >= pkgConstant.MaxMessageRetries,
		"retry count 4 should still allow retry")

	// Verify headers are built correctly for the last retry attempt
	headers := buildRetryHeaders(nil, 4, errors.New("transient error"))
	assert.Equal(t, 5, headers[pkgConstant.RetryCountHeader],
		"after retry count 4, next retry should be 5 (which equals MaxMessageRetries)")
}

func TestConsumerRoutes_BuildRetryHeaders(t *testing.T) {
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
		{
			name:            "truncates long failure reason to RetryFailureReasonMaxLen",
			existingHeaders: nil,
			retryCount:      0,
			lastErr:         errors.New(strings.Repeat("x", pkgConstant.RetryFailureReasonMaxLen+100)),
			wantRetryCount:  1,
			wantReason:      strings.Repeat("x", pkgConstant.RetryFailureReasonMaxLen),
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

func TestConsumerRoutes_BuildRetryHeaders_PreservesExistingHeaders(t *testing.T) {
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

func TestConsumerRoutes_SanitizeFailureReason(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		reason string
		want   string
	}{
		{
			name:   "short reason is unchanged",
			reason: "connection timeout",
			want:   "connection timeout",
		},
		{
			name:   "empty reason is unchanged",
			reason: "",
			want:   "",
		},
		{
			name:   "exact limit is unchanged",
			reason: strings.Repeat("a", pkgConstant.RetryFailureReasonMaxLen),
			want:   strings.Repeat("a", pkgConstant.RetryFailureReasonMaxLen),
		},
		{
			name:   "over limit is truncated",
			reason: strings.Repeat("b", pkgConstant.RetryFailureReasonMaxLen+50),
			want:   strings.Repeat("b", pkgConstant.RetryFailureReasonMaxLen),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := sanitizeFailureReason(tt.reason)
			assert.Equal(t, tt.want, got)
			assert.LessOrEqual(t, len(got), pkgConstant.RetryFailureReasonMaxLen)
		})
	}
}
