// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package constant

import "time"

// RabbitMQ Retry Configuration
const (
	// MaxMessageRetries is the maximum number of retry attempts before sending to DLQ.
	MaxMessageRetries = 5

	// RetryInitialBackoff is the base delay for exponential backoff calculation.
	RetryInitialBackoff = 1 * time.Second

	// RetryMaxBackoff is the upper bound for the backoff delay.
	RetryMaxBackoff = 30 * time.Second

	// RetryJitterMax is the maximum random jitter added to backoff to prevent thundering herd.
	RetryJitterMax = 500 * time.Millisecond

	// RetryCountHeader is the RabbitMQ message header key for tracking retry attempts.
	RetryCountHeader = "x-retry-count"

	// RetryFailureReasonHeader is the RabbitMQ message header key for tracking the last failure reason.
	RetryFailureReasonHeader = "x-failure-reason"
)
