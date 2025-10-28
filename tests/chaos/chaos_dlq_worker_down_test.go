package chaos

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	h "github.com/LerianStudio/reporter/v4/tests/helpers"
)

// TestChaos_DLQ_WorkerDownThenRabbitMQCrash tests message persistence when both Worker and RabbitMQ fail
func TestChaos_DLQ_WorkerDownThenRabbitMQCrash(t *testing.T) {
	env := h.LoadEnvironment()
	if env.DefaultOrgID == "" {
		t.Skip("X-Organization-Id not configured; set ORG_ID or X_ORGANIZATION_ID")
	}

	t.Log("⏳ Waiting for system stability...")
	time.Sleep(5 * time.Second)

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg(env.DefaultOrgID)

	t.Log("🔍 Step 1: Verifying system health...")
	if err := h.WaitForSystemHealth(ctx, cli, 70*time.Second); err != nil {
		t.Fatalf("❌ System not healthy: %v", err)
	}
	t.Log("✅ System is healthy")

	// Get Worker container name
	workerContainer := env.WorkerContainer
	if workerContainer == "" {
		workerContainer = "reporter-worker"
	}

	rabbitContainer := env.RabbitContainer
	if rabbitContainer == "" {
		rabbitContainer = "reporter-rabbitmq"
	}
	t.Log("📄 Step 3: Fetching existing template...")

	listCode, listBody, err := cli.Request(ctx, "GET", "/v1/templates?limit=1", headers, nil)
	if err != nil {
		t.Fatalf("❌ Failed to list templates: %v", err)
	}

	if listCode != 200 {
		t.Fatalf("❌ Failed to list templates (HTTP %d): %s", listCode, string(listBody))
	}

	var templateList struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}

	if err := json.Unmarshal(listBody, &templateList); err != nil {
		t.Fatalf("❌ Error decoding template list: %v", err)
	}

	if len(templateList.Items) == 0 {
		t.Skip("⚠️ No templates found in system - skipping test")
	}

	templateID := templateList.Items[0].ID
	t.Logf("✅ Using template: %s", templateID)

	// Step 3: Stop Worker NOW (after getting template)
	t.Log("💥 Step 3: Stopping Worker (messages will accumulate)...")
	if err := h.StopContainer(workerContainer); err != nil {
		t.Fatalf("❌ Failed to stop Worker: %v", err)
	}
	t.Log("✅ Worker stopped")

	// Wait a bit for worker to fully stop
	time.Sleep(3 * time.Second)

	// Step 4: Create report (Worker is down, so message will sit in queue)
	payload := map[string]any{
		"templateId": templateID,
		"filters": map[string]any{
			"midaz_onboarding": map[string]any{
				"organization": map[string]any{
					"id": map[string]any{
						"eq": []any{env.DefaultOrgID},
					},
				},
			},
		},
	}

	t.Log("🚀 Step 4: Creating report (Worker is DOWN - message will queue)...")

	// Retry request in case of temporary connection issues when Worker was stopped
	var code int
	var body []byte
	var reqErr error

	maxRetries := 5
	for attempt := 1; attempt <= maxRetries; attempt++ {
		code, body, reqErr = cli.Request(ctx, "POST", "/v1/reports", headers, payload)
		if reqErr == nil && (code == 200 || code == 201) {
			break
		}

		if attempt < maxRetries {
			t.Logf("⚠️ Request attempt %d/%d failed: %v (code: %d), retrying in 2s...", attempt, maxRetries, reqErr, code)
			time.Sleep(2 * time.Second)
		}
	}

	if reqErr != nil {
		t.Fatalf("❌ Request error after %d attempts: %v", maxRetries, reqErr)
	}

	if code != 200 && code != 201 {
		t.Fatalf("❌ Expected 200/201, got %d: %s", code, string(body))
	}

	var reportResponse struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &reportResponse); err != nil {
		t.Fatalf("❌ Error decoding response: %v", err)
	}

	reportID := reportResponse.ID
	t.Logf("✅ Report created: %s (message in queue, Worker DOWN)", reportID)

	// Verify report is in "Processing" state
	initialReport, err := cli.GetReportStatus(ctx, reportID, headers)
	if err != nil {
		t.Fatalf("❌ Failed to get report status: %v", err)
	}
	t.Logf("📊 Report status (Worker DOWN): %s", initialReport.Status)

	if initialReport.Status != "Processing" {
		t.Logf("⚠️ Expected 'Processing', got '%s'", initialReport.Status)
	}

	// Step 5: Wait for message to be in queue
	t.Log("⏳ Step 5: Waiting for message to be published to RabbitMQ queue (5s)...")
	time.Sleep(5 * time.Second)

	// Step 6: Crash RabbitMQ (message should be in queue)
	t.Log("💥 Step 6: CHAOS - Stopping RabbitMQ (message is in queue)...")
	if err := h.StopContainer(rabbitContainer); err != nil {
		t.Fatalf("❌ Failed to stop RabbitMQ: %v", err)
	}
	t.Log("✅ RabbitMQ stopped - message should be persisted on disk (durable)")

	// Simulate downtime
	t.Log("⏳ Simulating downtime (10 seconds)...")
	time.Sleep(10 * time.Second)

	// Check report status - should still be "Processing"
	t.Log("🔍 Step 7: Checking report status during TOTAL DOWNTIME (Worker + RabbitMQ)...")
	downtimeReport, err := cli.GetReportStatus(ctx, reportID, headers)
	if err != nil {
		t.Logf("⚠️ Could not fetch report: %v", err)
	} else {
		t.Logf("📊 Status during downtime: %s", downtimeReport.Status)
		if downtimeReport.Status != "Processing" {
			t.Logf("⚠️ Unexpected: status is '%s' (expected 'Processing')", downtimeReport.Status)
		} else {
			t.Log("✅ Status is 'Processing' as expected (no Worker to process)")
		}
	}

	// Step 8: Restart RabbitMQ FIRST
	t.Log("🔄 Step 8: RECOVERY - Restarting RabbitMQ (Worker still DOWN)...")
	if err := h.StartWithWait(rabbitContainer, 15*time.Second); err != nil {
		t.Fatalf("❌ Failed to start RabbitMQ: %v", err)
	}
	t.Log("✅ RabbitMQ restarted")

	// Wait for RabbitMQ to load definitions and restore queues
	t.Log("⏳ Waiting for RabbitMQ to load definitions (10s)...")
	time.Sleep(10 * time.Second)

	// Step 9: Restart Worker (should consume message from queue)
	t.Log("🔄 Step 9: RECOVERY - Restarting Worker...")
	if err := h.StartWithWait(workerContainer, 10*time.Second); err != nil {
		t.Fatalf("❌ Failed to start Worker: %v", err)
	}
	t.Log("✅ Worker restarted - should now consume message from queue")

	// Step 10: Wait for processing
	t.Log("⏳ Step 10: Waiting for Worker to reconnect and process message...")
	time.Sleep(10 * time.Second)

	// Step 11: Check final status
	t.Log("🔍 Step 11: Checking if message was reprocessed...")
	t.Log("⏳ Waiting up to 60 seconds for reprocessing...")

	finalReport, err := cli.WaitForReportStatus(ctx, reportID, headers, "Finished", 60*time.Second)
	if err != nil {
		// If not finished, check what status it is
		currentReport, err2 := cli.GetReportStatus(ctx, reportID, headers)
		if err2 != nil {
			t.Fatalf("❌ Error fetching final status: %v", err2)
		}

		t.Logf("📊 Final report status: %s", currentReport.Status)

		if currentReport.Status == "Processing" {
			t.Error("❌ TEST FAILED: Message was LOST!")
			t.Error("💡 Report stuck in 'Processing' - message did NOT persist through RabbitMQ crash")
			t.Error("🔧 DLQ/Durable queue implementation is NOT working")
			t.FailNow()
		} else if currentReport.Status == "Error" {
			t.Log("✅ Report processed with Error (message was reprocessed)")
			t.Log("💡 Message persisted through crash but processing failed")
		} else {
			t.Logf("⚠️ Unexpected final status: %s", currentReport.Status)
		}
	} else {
		t.Log("✅✅✅ TEST PASSED SUCCESSFULLY! ✅✅✅")
		t.Logf("📊 Final status: %s", finalReport.Status)
		t.Log("💡 Message persisted through RabbitMQ crash AND was reprocessed successfully!")
		t.Log("🎉 DLQ/Durable queue implementation is WORKING PERFECTLY!")
	}

	// Final verification
	finalReport, _ = cli.GetReportStatus(ctx, reportID, headers)
	if finalReport.Status == "Processing" {
		t.Fatalf("❌ CRITICAL FAILURE: Message was LOST - report stuck in Processing")
	}

	t.Log("✅ TEST PASSED: Message was NOT lost during Worker + RabbitMQ failure")
	t.Logf("📊 Final status: %s", finalReport.Status)
}
