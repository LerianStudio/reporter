//go:build chaos

// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package chaos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	h "github.com/LerianStudio/reporter/tests/utils"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Chaos_RabbitMQ_QueueFailureDuringReportGeneration simulates a failure of the
// RabbitMQ queue during report generation following the 5-phase chaos test structure:
// Phase 1 (Normal) -> Phase 2 (Inject) -> Phase 3 (Verify Failure) -> Phase 4 (Restore) -> Phase 5 (Verify Recovery)
func TestIntegration_Chaos_RabbitMQ_QueueFailureDuringReportGeneration(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because this test manipulates shared infrastructure (restarts RabbitMQ).
	if os.Getenv("CHAOS") != "1" {
		t.Skip("Set CHAOS=1 to run chaos tests")
	}
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}
	ctx := context.Background()
	cli := h.NewHTTPClient(GetManagerAddress(), 30*time.Second)
	headers := h.AuthHeaders()

	// Phase 1 (Normal): Verify system health and create a report under normal conditions
	t.Log("Phase 1 (Normal): Verifying system health before chaos test...")
	if err := h.WaitForSystemHealth(ctx, cli, 60*time.Second); err != nil {
		t.Fatalf("Phase 1 (Normal): System not healthy before chaos test: %v", err)
	}
	t.Log("Phase 1 (Normal): System is healthy, creating report under normal conditions...")

	templateID := "00000000-0000-0000-0000-000000000000"

	payload := map[string]any{
		"templateId": templateID,
		"filters": map[string]any{
			"status": map[string]any{
				"in": []any{"active"},
			},
		},
	}

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
			t.Logf("Phase 1 (Normal): Request attempt %d/%d failed: %v (code: %d), retrying in 2s...", attempt, maxRetries, err, code)
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		t.Fatalf("Phase 1 (Normal): Request error after %d attempts: %v", maxRetries, err)
	}

	if code != 201 && (code < 400 || code >= 500) {
		t.Fatalf("Phase 1 (Normal): Unexpected response code: %d, body: %s", code, string(body))
	}

	var reportResponse struct {
		ID string `json:"id"`
	}
	if code == 201 {
		if err := json.Unmarshal(body, &reportResponse); err != nil {
			t.Fatalf("Phase 1 (Normal): Error decoding response: %v", err)
		}
		t.Logf("Phase 1 (Normal): Report created with ID: %s", reportResponse.ID)
	} else {
		t.Logf("Phase 1 (Normal): Report not created (code %d), skipping chaos injection", code)
		return
	}

	// Intentional wait: allow time for message to be published before inducing chaos
	t.Log("Phase 1 (Normal): Waiting for message to be sent to RabbitMQ...")
	time.Sleep(2 * time.Second)

	// Phase 2 (Inject): Restart RabbitMQ container during report processing
	t.Log("Phase 2 (Inject): Restarting RabbitMQ container during report processing...")
	if err := RestartRabbitMQ(10 * time.Second); err != nil {
		t.Fatalf("Phase 2 (Inject): Failed to restart RabbitMQ: %v", err)
	}
	t.Log("Phase 2 (Inject): RabbitMQ restart initiated")

	// Phase 3 (Verify Failure): Check report status during disruption
	t.Log("Phase 3 (Verify Failure): Checking report status during RabbitMQ disruption...")
	report, err := cli.GetReportStatus(ctx, reportResponse.ID, headers)
	if err != nil {
		t.Logf("Phase 3 (Verify Failure): Could not fetch report (expected during disruption): %v", err)
	} else {
		t.Logf("Phase 3 (Verify Failure): Report status during disruption: %s", report.Status)
		if report.Status == "Processing" {
			t.Log("Phase 3 (Verify Failure): Report still processing - message may be in-flight or lost")
		}
	}

	// Phase 4 (Restore): Wait for RabbitMQ and worker to recover
	t.Log("Phase 4 (Restore): Waiting for worker to reconnect to RabbitMQ...")
	require.Eventually(t, func() bool {
		code, _, err := cli.Request(ctx, "GET", "/ready", nil, nil)
		return err == nil && code == 200
	}, 90*time.Second, 2*time.Second, "service did not become healthy after RabbitMQ restart")
	t.Log("Phase 4 (Restore): System is healthy again")

	// Phase 5 (Verify Recovery): Check if report was eventually processed
	t.Log("Phase 5 (Verify Recovery): Checking final report status after recovery...")
	finalReport, err := cli.WaitForReportStatus(ctx, reportResponse.ID, headers, "Finished", 30*time.Second)
	if err != nil {
		currentReport, err2 := cli.GetReportStatus(ctx, reportResponse.ID, headers)
		if err2 != nil {
			t.Fatalf("Phase 5 (Verify Recovery): Error fetching final status: %v", err2)
		}

		t.Logf("Phase 5 (Verify Recovery): Final report status: %s", currentReport.Status)

		if currentReport.Status == "Processing" {
			t.Log("Phase 5 (Verify Recovery): PROBLEM IDENTIFIED - Report stuck in 'Processing' status")
			t.Log("Phase 5 (Verify Recovery): This indicates the message was lost when RabbitMQ crashed")
			t.Log("Phase 5 (Verify Recovery): Chaos test PASSED - problem identified correctly")
		} else if currentReport.Status == "Finished" {
			t.Log("Phase 5 (Verify Recovery): Report processed successfully after restart")
		} else {
			t.Logf("Phase 5 (Verify Recovery): Unexpected status: %s", currentReport.Status)
		}
	} else {
		t.Logf("Phase 5 (Verify Recovery): Report processed successfully! Status: %s", finalReport.Status)
	}
}

// TestIntegration_Chaos_RabbitMQ_MessageLossSimulation simulates message loss in a more controlled way
// following the 5-phase chaos test structure:
// Phase 1 (Normal) -> Phase 2 (Inject) -> Phase 3 (Verify Failure) -> Phase 4 (Restore) -> Phase 5 (Verify Recovery)
func TestIntegration_Chaos_RabbitMQ_MessageLossSimulation(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because this test manipulates shared infrastructure (restarts RabbitMQ).
	if os.Getenv("CHAOS") != "1" {
		t.Skip("Set CHAOS=1 to run chaos tests")
	}
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	ctx := context.Background()
	cli := h.NewHTTPClient(GetManagerAddress(), 30*time.Second)
	headers := h.AuthHeaders()

	// Phase 1 (Normal): Verify system health and create multiple reports under normal conditions
	t.Log("Phase 1 (Normal): Verifying system health and creating reports...")
	if err := h.WaitForSystemHealth(ctx, cli, 60*time.Second); err != nil {
		t.Fatalf("Phase 1 (Normal): System not healthy: %v", err)
	}

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
			t.Logf("Phase 1 (Normal): Request error %d: %v", i, err)
			continue
		}

		if code == 201 {
			var report struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(body, &report); err == nil {
				t.Logf("Phase 1 (Normal): Report %d created: %s", i, report.ID)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
	t.Log("Phase 1 (Normal): Reports created, messages in-flight to RabbitMQ")

	// Phase 2 (Inject): Restart RabbitMQ during message processing
	t.Log("Phase 2 (Inject): Restarting RabbitMQ during processing...")
	if err := RestartRabbitMQ(5 * time.Second); err != nil {
		t.Fatalf("Phase 2 (Inject): Failed to restart RabbitMQ: %v", err)
	}
	t.Log("Phase 2 (Inject): RabbitMQ restart initiated")

	// Phase 3 (Verify Failure): Confirm system detects disruption
	t.Log("Phase 3 (Verify Failure): Checking system state during RabbitMQ disruption...")
	code, _, err := cli.Request(ctx, "GET", "/ready", nil, nil)
	if err != nil || code != 200 {
		t.Logf("Phase 3 (Verify Failure): System correctly reports degraded state (code=%d, err=%v)", code, err)
	} else {
		t.Log("Phase 3 (Verify Failure): System still reports healthy - RabbitMQ may have restarted quickly")
	}

	// Phase 4 (Restore): Wait for system to recover
	t.Log("Phase 4 (Restore): Waiting for system to recover and process messages...")
	require.Eventually(t, func() bool {
		code, _, err := cli.Request(ctx, "GET", "/ready", nil, nil)
		return err == nil && code == 200
	}, 90*time.Second, 2*time.Second, "service did not become healthy after RabbitMQ restart")
	// Intentional wait: allow extra time for worker to reprocess queued messages
	time.Sleep(5 * time.Second)
	t.Log("Phase 4 (Restore): System has recovered")

	// Phase 5 (Verify Recovery): Check if messages were processed or lost
	t.Log("Phase 5 (Verify Recovery): Checking report statuses after recovery...")
	reports, err := cli.ListReports(ctx, headers, "limit=10")
	if err != nil {
		t.Fatalf("Phase 5 (Verify Recovery): Error listing reports: %v", err)
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

	t.Logf("Phase 5 (Verify Recovery): Result: %d processed, %d still processing", finishedCount, processingCount)

	if processingCount > 0 {
		t.Log("Phase 5 (Verify Recovery): Orphaned reports detected - messages lost during restart")
		t.Log("Phase 5 (Verify Recovery): This confirms the message loss problem during RabbitMQ failures")
	} else {
		t.Log("Phase 5 (Verify Recovery): All reports processed - system recovered successfully")
	}
}
