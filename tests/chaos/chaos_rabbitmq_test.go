// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package chaos

import (
	"testing"
	"time"
)

// Restarts RabbitMQ container and validates recovery of the system
func TestChaos_RabbitMQ_RestartAndRecover(t *testing.T) {
	if err := RestartRabbitMQ(5 * time.Second); err != nil {
		t.Fatalf("failed to restart rabbitmq: %v", err)
	}
}
