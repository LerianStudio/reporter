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

// TestChaos_DLQ_RecoveryAfterRabbitMQFailure tests that messages are not lost when RabbitMQ crashes
func TestChaos_DLQ_RecoveryAfterRabbitMQFailure(t *testing.T) {
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
	t.Log("📄 Step 2: Fetching existing template...")

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
		t.Skip("⚠️ No templates found in system - skipping test. Create a template first.")
	}

	templateID := templateList.Items[0].ID
	t.Logf("✅ Using existing template: %s", templateID)

	// Step 3: Create report with proper filters
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

	t.Log("🚀 Step 3: Creating report via Manager...")
	code, body, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
	if err != nil {
		t.Fatalf("❌ Request error: %v", err)
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
	t.Logf("✅ Report created successfully! ID: %s", reportID)

	// Verify report is in "Processing" state
	t.Log("🔍 Step 3: Verifying initial report status...")
	initialReport, err := cli.GetReportStatus(ctx, reportID, headers)
	if err != nil {
		t.Fatalf("❌ Failed to get report status: %v", err)
	}
	t.Logf("📊 Initial status: %s", initialReport.Status)

	// Wait a bit for message to be published to RabbitMQ
	t.Log("⏳ Waiting for message to be published to RabbitMQ (2s)...")
	time.Sleep(2 * time.Second)

	// CHAOS: Crash RabbitMQ
	t.Log("💥 Step 4: CHAOS - Stopping RabbitMQ (simulating crash)...")
	rabbitContainer := env.RabbitContainer
	if rabbitContainer == "" {
		rabbitContainer = "reporter-rabbitmq"
	}

	if err := h.StopContainer(rabbitContainer); err != nil {
		t.Fatalf("❌ Failed to stop RabbitMQ: %v", err)
	}
	t.Log("✅ RabbitMQ stopped (simulating crash)")

	// Wait during "downtime"
	t.Log("⏳ Simulating downtime (5 seconds)...")
	time.Sleep(5 * time.Second)

	// Check report status during downtime - should still be "Processing"
	t.Log("🔍 Step 5: Checking report status during RabbitMQ downtime...")
	downtimeReport, err := cli.GetReportStatus(ctx, reportID, headers)
	if err != nil {
		t.Logf("⚠️ Could not fetch report during downtime: %v", err)
	} else {
		t.Logf("📊 Status during downtime: %s", downtimeReport.Status)
		if downtimeReport.Status != "Processing" {
			t.Logf("⚠️ Unexpected: report status changed during RabbitMQ downtime!")
		}
	}

	// RECOVERY: Restart RabbitMQ
	t.Log("🔄 Step 6: RECOVERY - Starting RabbitMQ...")
	if err := h.StartWithWait(rabbitContainer, 15*time.Second); err != nil {
		t.Fatalf("❌ Failed to start RabbitMQ: %v", err)
	}
	t.Log("✅ RabbitMQ started successfully")

	// Wait for RabbitMQ to fully initialize and load definitions
	t.Log("⏳ Waiting for RabbitMQ to fully initialize (10s)...")
	time.Sleep(10 * time.Second)

	// Wait for worker to reconnect and process the message
	t.Log("⏳ Step 7: Waiting for worker to reconnect and process message...")
	time.Sleep(5 * time.Second)

	// Check report status - should eventually be processed
	t.Log("🔍 Step 8: Checking final report status...")
	t.Log("⏳ Waiting up to 60 seconds for message to be reprocessed...")

	finalReport, err := cli.WaitForReportStatus(ctx, reportID, headers, "Finished", 60*time.Second)
	if err != nil {
		// If not finished, check what status it ended up in
		currentReport, err2 := cli.GetReportStatus(ctx, reportID, headers)
		if err2 != nil {
			t.Fatalf("❌ Error fetching final status: %v", err2)
		}

		t.Logf("📊 Final report status: %s", currentReport.Status)

		if currentReport.Status == "Processing" {
			t.Error("❌ FAILURE: Report stuck in 'Processing' status")
			t.Error("💡 This indicates the message was lost or not reprocessed")
			t.Error("🔧 DLQ/DLX implementation may not be working correctly")
		} else if currentReport.Status == "Error" {
			t.Log("✅ Report status updated to 'Error' (message was reprocessed)")
		} else {
			t.Logf("⚠️ Unexpected final status: %s", currentReport.Status)
		}
	} else {
		t.Log("✅ SUCCESS: Report processed successfully after RabbitMQ recovery!")
		t.Logf("📊 Final status: %s", finalReport.Status)
		t.Log("💡 Message persisted through RabbitMQ crash and was reprocessed")
	}

	// Final verification
	finalReport, _ = cli.GetReportStatus(ctx, reportID, headers)
	if finalReport.Status == "Processing" {
		t.Fatalf("❌ TEST FAILED: Message was lost - report stuck in Processing")
	}

	t.Log("✅ TEST PASSED: Message was not lost during RabbitMQ failure")
}
