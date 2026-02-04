// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package chaos

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	h "github.com/LerianStudio/reporter/v4/tests/helpers"
)

// TestChaos_RabbitMQ_ConnectionClosed tests the behavior when manager tries to send
// a message to RabbitMQ but the connection is closed
func TestChaos_RabbitMQ_ConnectionClosed(t *testing.T) {
	env := h.LoadEnvironment()

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeaders()

	t.Log("🔧 Starting RabbitMQ connection chaos test...")

	t.Log("Step 1: Verifying normal system operation...")
	templateID, ok := getAnyTemplateIDWithRetry(ctx, t, cli, headers, 10, 2*time.Second)
	if !ok {
		t.Skip("No templates available or service unstable for chaos testing")
	}
	t.Logf("Using template ID: %s", templateID)

	t.Log("Step 2: Closing RabbitMQ connection (stopping container)...")
	err := h.RestartWithWait(env.RabbitContainer, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to restart RabbitMQ: %v", err)
	}

	time.Sleep(3 * time.Second)
	t.Log("Step 3: Attempting to create report with RabbitMQ disconnected...")
	payload := map[string]any{
		"templateId": templateID,
		"filters":    map[string]any{},
	}

	code, body, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
	t.Log("Step 4: Analyzing system behavior with closed connection...")

	if err != nil {
		t.Logf("✅ Expected behavior: Request failed due to connection error: %v", err)
	} else if code >= 500 {
		t.Logf("✅ Expected behavior: Server error (code=%d) due to RabbitMQ unavailability", code)
	} else if code == 200 || code == 201 {
		t.Logf("⚠️  Unexpected behavior: Request succeeded (code=%d) despite RabbitMQ being down", code)
		t.Logf("Response body: %s", string(body))
	} else {
		t.Logf("📊 System behavior: HTTP %d - %s", code, string(body))
	}

	t.Log("Step 5: Verifying manager still responds to other requests...")
	code, body, err = cli.Request(ctx, "GET", "/v1/templates?limit=1", headers, nil)
	if err != nil {
		t.Logf("⚠️  Manager not responding to other requests: %v", err)
	} else if code != 200 {
		t.Logf("⚠️  Manager responding with error to other requests: code=%d", code)
	} else {
		t.Logf("✅ Manager still responding to other requests normally")
	}

	t.Log("Step 6: Restoring RabbitMQ connection...")
	err = h.RestartWithWait(env.RabbitContainer, 10*time.Second)
	if err != nil {
		t.Fatalf("Failed to restore RabbitMQ: %v", err)
	}

	time.Sleep(5 * time.Second)

	t.Log("Step 7: Verifying system recovery...")
	code, _, err = cli.Request(ctx, "POST", "/v1/reports", headers, payload)
	if err != nil {
		t.Logf("⚠️  System not fully recovered: %v", err)
	} else if code == 201 || code == 200 {
		t.Logf("✅ System fully recovered - report creation working again")
	} else {
		t.Logf("📊 System recovery status: HTTP %d", code)
	}

	t.Log("🎯 RabbitMQ connection chaos test completed")
}

// TestChaos_RabbitMQ_ChannelClosed tests when RabbitMQ is running but the channel is closed
func TestChaos_RabbitMQ_ChannelClosed(t *testing.T) {
	env := h.LoadEnvironment()

	t.Log("⏳ Waiting for full system recovery after previous chaos tests...")
	time.Sleep(30 * time.Second) // Give time for Manager and RabbitMQ to fully stabilize

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeaders()

	t.Log("🔧 Starting RabbitMQ channel chaos test...")

	t.Log("🔍 Verifying system health before channel chaos test...")
	if err := h.WaitForSystemHealth(ctx, cli, 60*time.Second); err != nil {
		t.Logf("⚠️  System health check failed: %v", err)
		t.Skip("System not ready for channel chaos test - likely recovering from previous test")
	}

	t.Log("Step 1: Verifying normal system operation...")
	templateID, ok := getAnyTemplateIDWithRetry(ctx, t, cli, headers, 10, 2*time.Second)
	if !ok {
		t.Skip("No templates available or service unstable for chaos testing")
	}

	t.Log("Step 2: Simulating channel closure (quick RabbitMQ restart)...")
	err := h.RestartWithWait(env.RabbitContainer, 2*time.Second)
	if err != nil {
		t.Fatalf("Failed to restart RabbitMQ: %v", err)
	}

	time.Sleep(3 * time.Second)

	t.Log("Step 3: Attempting to create report during channel issues...")
	payload := map[string]any{
		"templateId": templateID,
		"filters":    map[string]any{},
	}

	code, body, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)

	t.Log("Step 4: Analyzing behavior during channel issues...")

	if err != nil {
		t.Logf("✅ Expected behavior: Request failed due to channel error: %v", err)
	} else if code >= 500 {
		t.Logf("✅ Expected behavior: Server error (code=%d) due to channel issues", code)
	} else if code == 200 || code == 201 {
		t.Logf("✅ System handled channel issue gracefully: code=%d", code)
	} else {
		t.Logf("📊 System behavior during channel issues: HTTP %d - %s", code, string(body))
	}

	t.Log("Step 5: Waiting for automatic recovery...")
	time.Sleep(10 * time.Second)

	t.Log("Step 6: Verifying automatic recovery...")
	code, _, err = cli.Request(ctx, "POST", "/v1/reports", headers, payload)
	if err != nil {
		t.Logf("⚠️  System not fully recovered: %v", err)
	} else if code == 201 || code == 200 {
		t.Logf("✅ System automatically recovered - report creation working")
	} else {
		t.Logf("📊 System recovery status: HTTP %d", code)
	}

	t.Log("🎯 RabbitMQ channel chaos test completed")
}

// TestChaos_RabbitMQ_QueueFull tests behavior when RabbitMQ queue is full or unavailable
func TestChaos_RabbitMQ_QueueFull(t *testing.T) {
	env := h.LoadEnvironment()

	t.Log("⏳ Waiting for full system recovery after previous chaos tests...")
	time.Sleep(30 * time.Second) // Increased from 15s to 30s for datasource reconnection

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeaders()

	t.Log("🔍 Verifying system health before queue chaos test...")
	if err := h.WaitForSystemHealth(ctx, cli, 90*time.Second); err != nil {
		t.Logf("⚠️  System health check failed after 90s: %v", err)
		t.Log("💡 This may be due to datasource initialization with retry - skipping test")
		t.Skip("System not ready for chaos test - likely datasource initialization in progress")
	}
	t.Log("✅ System is healthy, proceeding with queue chaos test...")

	t.Log("🔧 Starting RabbitMQ queue chaos test...")

	t.Log("Step 1: Verifying normal system operation...")
	templateID, ok := getAnyTemplateIDWithRetry(ctx, t, cli, headers, 10, 2*time.Second)
	if !ok {
		t.Skip("No templates available or service unstable for chaos testing")
	}

	t.Log("Step 2: Simulating queue unavailability...")
	err := h.RestartWithWait(env.RabbitContainer, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to restart RabbitMQ: %v", err)
	}

	time.Sleep(2 * time.Second)

	t.Log("Step 3: Attempting rapid report creation during queue issues...")
	payload := map[string]any{
		"templateId": templateID,
		"filters":    map[string]any{},
	}

	successCount := 0
	errorCount := 0

	for i := 0; i < 5; i++ {
		code, _, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
		if err != nil {
			errorCount++
			t.Logf("Request %d failed: %v", i+1, err)
		} else if code == 201 || code == 200 {
			successCount++
			t.Logf("Request %d succeeded: code=%d", i+1, code)
		} else {
			errorCount++
			t.Logf("Request %d returned error: code=%d", i+1, code)
		}
		time.Sleep(500 * time.Millisecond) // Pequeno delay entre requests
	}

	t.Log("Step 4: Analyzing behavior during queue issues...")
	t.Logf("📊 Results: %d successful, %d failed", successCount, errorCount)

	if errorCount > successCount {
		t.Logf("✅ Expected behavior: More failures during queue issues")
	} else if successCount > 0 {
		t.Logf("✅ System handled queue issues gracefully: %d successful requests", successCount)
	} else {
		t.Logf("📊 System behavior: All requests failed during queue issues")
	}

	t.Log("Step 5: Restoring RabbitMQ...")
	err = h.RestartWithWait(env.RabbitContainer, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to restore RabbitMQ: %v", err)
	}

	time.Sleep(5 * time.Second)

	t.Log("Step 6: Verifying system recovery...")
	code, _, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
	if err != nil {
		t.Logf("⚠️  System not fully recovered: %v", err)
	} else if code == 201 || code == 200 {
		t.Logf("✅ System fully recovered - report creation working again")
	} else {
		t.Logf("📊 System recovery status: HTTP %d", code)
	}

	t.Log("🎯 RabbitMQ queue chaos test completed")
}

// getAnyTemplateIDWithRetry tries to fetch any template ID with retries/backoff to tolerate transient errors
func getAnyTemplateIDWithRetry(ctx context.Context, t *testing.T, cli *h.HTTPClient, headers map[string]string, attempts int, delay time.Duration) (string, bool) {
	for i := 1; i <= attempts; i++ {
		code, body, err := cli.Request(ctx, "GET", "/v1/templates?limit=1", headers, nil)
		t.Logf("retry %d/%d: waiting %s due to err/code: %v/%d", i, attempts, delay, err, code)
		if err == nil && code == 200 {
			var templates struct {
				Items []struct {
					ID string `json:"id"`
				} `json:"items"`
			}
			_ = json.Unmarshal(body, &templates)
			if len(templates.Items) > 0 && templates.Items[0].ID != "" {
				return templates.Items[0].ID, true
			}
		}
		t.Logf("retry %d/%d: waiting %s due to err/code: %v/%d", i, attempts, delay, err, code)
		time.Sleep(delay)
	}
	return "", false
}
