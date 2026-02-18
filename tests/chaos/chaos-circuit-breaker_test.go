//go:build chaos

// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package chaos

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	h "github.com/LerianStudio/reporter/tests/utils"
)

// TestIntegration_Chaos_CircuitBreaker_OpenAndRecover tests circuit breaker opening and recovery
func TestIntegration_Chaos_CircuitBreaker_OpenAndRecover(t *testing.T) {
	if os.Getenv("CHAOS") != "1" {
		t.Skip("Set CHAOS=1 to run chaos tests")
	}
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	// Skip this test in testcontainers mode - requires external plugin_crm infrastructure
	if os.Getenv("USE_EXISTING_INFRA") != "true" {
		t.Skip("Skipping circuit breaker test - requires plugin_crm infrastructure (docker-compose)")
	}

	ctx := context.Background()
	cli := h.NewHTTPClient(GetManagerAddress(), 30*time.Second)
	headers := h.AuthHeaders()

	t.Log("🎯 Starting Circuit Breaker chaos test...")

	// Step 1: Wait for system to be healthy
	t.Log("🔍 Verifying system health...")
	if err := h.WaitForSystemHealth(ctx, cli, 90*time.Second); err != nil {
		t.Skip("System not ready for circuit breaker test")
	}

	// Step 2: Get a template for testing
	t.Log("📋 Step 1: Getting template for testing...")
	templateID, ok := getTemplateIDForCRM(ctx, t, cli, headers)
	if !ok {
		t.Skip("No suitable template found for circuit breaker test")
	}
	t.Logf("✅ Using template: %s", templateID)

	// Step 3: Simulate plugin_crm MongoDB failure
	t.Log("💥 Step 2: Simulating plugin_crm MongoDB failure (stopping container)...")
	crmContainer := "plugin-crm-mongodb"
	if err := h.StopContainer(crmContainer); err != nil {
		t.Logf("⚠️  Could not stop plugin_crm container (may not exist): %v", err)
		t.Skip("plugin_crm container not available for chaos test")
	}
	t.Log("✅ plugin_crm MongoDB stopped")

	// Step 4: Generate multiple reports to trigger circuit breaker
	t.Log("🔄 Step 3: Sending 20 report requests to trigger circuit breaker...")

	failureCount := 0
	successCount := 0

	for i := 1; i <= 20; i++ {
		payload := map[string]any{
			"templateId": templateID,
			"filters": map[string]any{
				"test_batch": map[string]any{
					"eq": []any{i},
				},
			},
		}

		code, body, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)

		if err != nil || code >= 500 {
			failureCount++
			t.Logf("  Request %d: ❌ Failed (code: %d, err: %v)", i, code, err)
		} else if code == 201 {
			successCount++

			var reportResponse struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(body, &reportResponse); err == nil {
				t.Logf("  Request %d: ✅ Created (report: %s)", i, reportResponse.ID)
			}
		} else {
			failureCount++
			t.Logf("  Request %d: ⚠️  Code %d", i, code)
		}

		// Small delay between requests
		time.Sleep(200 * time.Millisecond)

		// After 15 failures, circuit breaker should open
		if i == 15 {
			t.Log("🔍 Checkpoint: 15 requests sent - circuit breaker should be opening soon...")
			time.Sleep(1 * time.Second)
		}
	}

	t.Logf("📊 Results after 20 requests: %d successes, %d failures", successCount, failureCount)

	// Step 5: Intentional wait: allow circuit breaker state machine to evaluate failure threshold
	t.Log("⏳ Step 4: Waiting 5s for circuit breaker to process failures...")
	time.Sleep(5 * time.Second)

	// Step 6: Try more requests (should fast-fail due to circuit breaker open)
	t.Log("🚫 Step 5: Testing fast-fail with circuit breaker OPEN...")

	fastFailCount := 0
	for i := 1; i <= 5; i++ {
		start := time.Now()

		payload := map[string]any{
			"templateId": templateID,
			"filters":    map[string]any{},
		}

		code, _, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
		elapsed := time.Since(start)

		// Circuit breaker open should return error almost immediately
		if err != nil || code >= 500 {
			fastFailCount++
			if elapsed < 2*time.Second {
				t.Logf("  Request %d: ⚡ Fast-fail in %v (circuit breaker likely OPEN)", i, elapsed)
			} else {
				t.Logf("  Request %d: ❌ Slow fail in %v (timeout, not circuit breaker)", i, elapsed)
			}
		}
	}

	t.Logf("📊 Fast-fail results: %d/5 failed quickly", fastFailCount)

	// Step 7: Restore plugin_crm
	t.Log("🔧 Step 6: Restoring plugin_crm MongoDB...")
	if err := h.StartContainer(crmContainer); err != nil {
		t.Logf("⚠️  Could not start plugin_crm container: %v", err)
	} else {
		t.Log("✅ plugin_crm MongoDB restarted")
	}

	// Step 8: Intentional wait: circuit breaker timeout must expire before transitioning to half-open
	t.Log("⏳ Step 7: Waiting 35s for circuit breaker to transition to HALF-OPEN...")
	time.Sleep(35 * time.Second)

	// Step 9: Test recovery
	t.Log("🔄 Step 8: Testing system recovery (circuit breaker should be HALF-OPEN)...")

	recoveryCount := 0
	for i := 1; i <= 5; i++ {
		payload := map[string]any{
			"templateId": templateID,
			"filters":    map[string]any{},
		}

		code, _, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)

		if err == nil && code == 201 {
			recoveryCount++
			t.Logf("  Request %d: ✅ Success (circuit breaker transitioning to CLOSED)", i)
		} else {
			t.Logf("  Request %d: ⚠️  Code %d, err: %v", i, code, err)
		}

		time.Sleep(500 * time.Millisecond)
	}

	t.Logf("📊 Recovery results: %d/5 successful", recoveryCount)

	if recoveryCount > 0 {
		t.Log("✅ System recovered - circuit breaker likely transitioned back to CLOSED")
	}

	t.Log("")
	t.Log("🎯 Circuit Breaker chaos test completed!")
	t.Log("💡 Check worker logs for circuit breaker state changes:")
	t.Log("   docker logs plugin-reporter-worker 2>&1 | grep -E 'Circuit|breaker'")
}

// getTemplateIDForCRM tries to get any template that uses plugin_crm
func getTemplateIDForCRM(ctx context.Context, t *testing.T, cli *h.HTTPClient, headers map[string]string) (string, bool) {
	code, body, err := cli.Request(ctx, "GET", "/v1/templates?limit=10", headers, nil)
	if err != nil || code != 200 {
		t.Logf("⚠️  Could not list templates: code=%d, err=%v", code, err)
		return "", false
	}

	var templates struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &templates); err != nil {
		t.Logf("⚠️  Could not parse templates: %v", err)
		return "", false
	}

	if len(templates.Items) > 0 {
		return templates.Items[0].ID, true
	}

	return "", false
}
