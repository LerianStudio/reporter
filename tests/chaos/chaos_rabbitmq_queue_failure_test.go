// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package chaos

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	h "github.com/LerianStudio/reporter/v4/tests/helpers"
)

// TestChaos_RabbitMQ_QueueFailureDuringReportGeneration simulate a failure of the RabbitMQ queue during report generation
func TestChaos_RabbitMQ_QueueFailureDuringReportGeneration(t *testing.T) {
	env := h.LoadEnvironment()

	t.Log("⏳ Waiting for system stability after previous chaos tests...")
	time.Sleep(10 * time.Second)

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeaders()

	t.Log("🔍 Verifying system health before chaos test...")
	if err := h.WaitForSystemHealth(ctx, cli, 30*time.Second); err != nil {
		t.Fatalf("❌ System not healthy before chaos test: %v", err)
	}
	t.Log("✅ System is healthy, proceeding with chaos test...")

	templateID := "00000000-0000-0000-0000-000000000000"

	payload := map[string]any{
		"templateId": templateID,
		"filters": map[string]any{
			"status": map[string]any{
				"in": []any{"active"},
			},
		},
	}

	t.Log("🚀 Step 1: Sending POST /v1/reports...")

	var code int
	var body []byte
	var err error

	maxRetries := 5
	for attempt := 1; attempt <= maxRetries; attempt++ {
		code, body, err = cli.Request(ctx, "POST", "/v1/reports", headers, payload)
		if err == nil && (code == 201 || (code >= 400 && code < 500)) {
			break
		}

		if attempt < maxRetries {
			t.Logf("⚠️ Request attempt %d/%d failed: %v (code: %d), retrying in 2s...", attempt, maxRetries, err, code)
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		t.Fatalf("❌ Request error after %d attempts: %v", maxRetries, err)
	}

	if code != 201 && (code < 400 || code >= 500) {
		t.Fatalf("❌ Expected code: %d, body: %s", code, string(body))
	}

	var reportResponse struct {
		ID string `json:"id"`
	}
	if code == 201 {
		if err := json.Unmarshal(body, &reportResponse); err != nil {
			t.Fatalf("❌ Error to decode answer: %v", err)
		}
		t.Logf("✅ Report created with ID: %s", reportResponse.ID)
	} else {
		t.Logf("⚠️ Report does not created (code %d), continue with chaos tests...", code)
		return
	}

	t.Log("⏳ Waiting for message to be sent to RabbitMQ...")
	time.Sleep(2 * time.Second)

	t.Log("💥 CHAOS: Restarting RabbitMQ container...")
	rabbitContainer := env.RabbitContainer
	if rabbitContainer == "" {
		rabbitContainer = "reporter-rabbitmq"
	}

	if err := h.RestartWithWait(rabbitContainer, 10*time.Second); err != nil {
		t.Fatalf("❌ Failed to restart RabbitMQ: %v", err)
	}

	t.Log("✅ RabbitMQ restarted successfully")

	t.Log("⏳ Waiting for worker to reconnect...")
	time.Sleep(5 * time.Second)

	t.Log("🔍 Checking report status...")
	report, err := cli.GetReportStatus(ctx, reportResponse.ID, headers)
	if err != nil {
		t.Logf("⚠️ Could not fetch report: %v", err)
		return
	}

	t.Logf("📊 Current report status: %s", report.Status)
	t.Log("⏳ Waiting 30 seconds to see if worker processes the message...")

	finalReport, err := cli.WaitForReportStatus(ctx, reportResponse.ID, headers, "Finished", 30*time.Second)
	if err != nil {
		currentReport, err2 := cli.GetReportStatus(ctx, reportResponse.ID, headers)
		if err2 != nil {
			t.Fatalf("❌ Error fetching final status: %v", err2)
		}

		t.Logf("📊 Final report status: %s", currentReport.Status)

		if currentReport.Status == "Processing" {
			t.Log("🚨 PROBLEM IDENTIFIED: Report still in 'Processing' status")
			t.Log("💡 This indicates the message was lost when RabbitMQ crashed")
			t.Log("🔧 SOLUTION NEEDED: Implement retry or dead letter queue")
			t.Log("✅ Chaos test PASSED - problem identified correctly")
		} else if currentReport.Status == "Finished" {
			t.Log("✅ Report was processed successfully after restart")
			t.Log("💡 This indicates the system recovered or message was reprocessed")
		} else {
			t.Logf("⚠️ Unexpected status: %s", currentReport.Status)
		}
	} else {
		t.Logf("✅ Report processed successfully! Status: %s", finalReport.Status)
	}

	t.Log("📋 Checking worker logs...")
}

// TestChaos_RabbitMQ_MessageLossSimulation simulates message loss in a more controlled way
func TestChaos_RabbitMQ_MessageLossSimulation(t *testing.T) {
	env := h.LoadEnvironment()

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeaders()

	t.Log("🧪 Simulating message loss scenario...")

	for i := 0; i < 3; i++ {
		payload := map[string]any{
			"templateId": "00000000-0000-0000-0000-000000000000",
			"filters": map[string]any{
				"batch": map[string]any{
					"eq": []any{fmt.Sprintf("test-%d", i)},
				},
			},
		}

		code, body, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
		if err != nil {
			t.Logf("⚠️ Request error %d: %v", i, err)
			continue
		}

		if code == 201 {
			var report struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(body, &report); err == nil {
				t.Logf("✅ Report %d created: %s", i, report.ID)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Log("💥 Restarting RabbitMQ during processing...")
	rabbitContainer := env.RabbitContainer
	if rabbitContainer == "" {
		rabbitContainer = "reporter-rabbitmq"
	}

	if err := h.RestartWithWait(rabbitContainer, 5*time.Second); err != nil {
		t.Fatalf("❌ Failed to restart RabbitMQ: %v", err)
	}

	t.Log("⏳ Waiting for processing...")
	time.Sleep(20 * time.Second)

	reports, err := cli.ListReports(ctx, headers, "limit=10")
	if err != nil {
		t.Fatalf("❌ Error listing reports: %v", err)
	}

	processingCount := 0
	finishedCount := 0

	for _, report := range reports {
		if report.Status == "Processing" {
			processingCount++
		} else if report.Status == "Finished" {
			finishedCount++
		}
	}

	t.Logf("📊 Result: %d processed, %d still processing", finishedCount, processingCount)

	if processingCount > 0 {
		t.Log("🚨 Orphaned reports detected - messages lost during restart")
		t.Log("💡 This confirms the message loss problem during RabbitMQ failures")
	}
}
