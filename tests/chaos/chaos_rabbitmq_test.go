//go:build chaos

// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package chaos

import (
	"context"
	"os"
	"testing"
	"time"

	h "github.com/LerianStudio/reporter/tests/utils"
)

// Restarts RabbitMQ container and validates recovery of the system
func TestIntegration_Chaos_RabbitMQ_RestartAndRecover(t *testing.T) {
	if os.Getenv("CHAOS") != "1" {
		t.Skip("Set CHAOS=1 to run chaos tests")
	}
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	// Phase 1: Restart RabbitMQ container
	t.Log("Phase 1: Restarting RabbitMQ container...")
	if err := RestartRabbitMQ(5 * time.Second); err != nil {
		t.Fatalf("failed to restart rabbitmq: %v", err)
	}

	// Phase 2: Wait for RabbitMQ to fully stabilize after restart
	// RabbitMQ needs time to re-establish topology (exchanges, queues, bindings)
	t.Log("Phase 2: Waiting for RabbitMQ to stabilize after restart...")
	time.Sleep(15 * time.Second)

	// Phase 3: Verify system recovery by checking Manager health
	t.Log("Phase 3: Verifying system recovery...")
	cli := h.NewHTTPClient(GetManagerAddress(), 30*time.Second)
	headers := h.AuthHeaders()

	var lastErr error
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		code, _, _, err := cli.RequestFull(ctx, "GET", "/health", headers, nil)
		cancel()

		if err == nil && code == 200 {
			t.Log("Phase 3: Manager is healthy after RabbitMQ restart")
			return
		}

		lastErr = err
		t.Logf("Phase 3: Health check attempt %d/10 failed (code=%d, err=%v), retrying...", i+1, code, err)
		time.Sleep(3 * time.Second)
	}

	t.Fatalf("Manager did not recover after RabbitMQ restart: %v", lastErr)
}
