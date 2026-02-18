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
	"github.com/stretchr/testify/require"
)

// Restarts RabbitMQ container and validates recovery of the system
func TestIntegration_Chaos_RabbitMQ_RestartAndRecover(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because this test manipulates shared infrastructure (restarts RabbitMQ).
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

	// Phase 2: Wait for RabbitMQ to fully stabilize and Manager to recover
	// RabbitMQ needs time to re-establish topology (exchanges, queues, bindings)
	t.Log("Phase 2: Waiting for RabbitMQ to stabilize and Manager to recover...")
	cli := h.NewHTTPClient(GetManagerAddress(), 30*time.Second)

	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		code, _, err := cli.Request(ctx, "GET", "/ready", nil, nil)

		return err == nil && code == 200
	}, 90*time.Second, 2*time.Second, "Manager did not recover after RabbitMQ restart")

	t.Log("Phase 3: Manager is healthy after RabbitMQ restart")
}
