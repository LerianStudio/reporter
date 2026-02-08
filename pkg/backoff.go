// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pkg

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/LerianStudio/reporter/pkg/constant"
)

// FullJitter returns a random duration in [0, baseDelay], capped at ProducerMaxBackoff.
// Full jitter prevents thundering herd when multiple producers reconnect simultaneously
// after a RabbitMQ restart. Uses crypto/rand for unbiased distribution.
func FullJitter(baseDelay time.Duration) time.Duration {
	if baseDelay <= 0 {
		return 0
	}

	cap := baseDelay
	if cap > constant.ProducerMaxBackoff {
		cap = constant.ProducerMaxBackoff
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(cap)))
	if err != nil {
		// Fallback to half the base delay if crypto/rand fails (extremely unlikely).
		return cap / 2
	}

	return time.Duration(n.Int64())
}

// NextBackoff doubles the current delay, capped at ProducerMaxBackoff.
func NextBackoff(current time.Duration) time.Duration {
	next := time.Duration(float64(current) * constant.ProducerBackoffFactor)
	if next > constant.ProducerMaxBackoff {
		return constant.ProducerMaxBackoff
	}

	return next
}
