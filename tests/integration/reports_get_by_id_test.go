package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	h "plugin-smart-templates/v2/tests/helpers"
)

// TestIntegration_Reports_GetByID_ValidID tests GET /v1/reports/{id} with a valid report ID
func TestIntegration_Reports_GetByID_ValidID(t *testing.T) {
	env := h.LoadEnvironment()
	if env.DefaultOrgID == "" {
		t.Skip("X-Organization-Id not configured; set ORG_ID or X_ORGANIZATION_ID")
	}
	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg(env.DefaultOrgID)

	code, body, err := cli.Request(ctx, "GET", "/v1/reports?limit=1", headers, nil)
	if err != nil {
		t.Fatalf("list reports error: %v", err)
	}
	if code != 200 {
		t.Fatalf("list reports failed: code=%d body=%s", code, string(body))
	}

	var reports struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	_ = json.Unmarshal(body, &reports)

	if len(reports.Items) == 0 {
		t.Skip("No reports found to test GET by ID")
	}

	reportID := reports.Items[0].ID
	t.Logf("Testing GET /v1/reports/%s", reportID)

	code, body, err = cli.Request(ctx, "GET", fmt.Sprintf("/v1/reports/%s", reportID), headers, nil)
	if err != nil {
		t.Fatalf("get report by ID error: %v", err)
	}
	if code != 200 {
		t.Fatalf("get report by ID failed: code=%d body=%s", code, string(body))
	}

	var report struct {
		ID          string         `json:"id"`
		Status      string         `json:"status"`
		TemplateID  string         `json:"templateId"`
		CreatedAt   string         `json:"createdAt"`
		UpdatedAt   string         `json:"updatedAt"`
		CompletedAt string         `json:"completedAt"`
		DeletedAt   string         `json:"deletedAt"`
		Filters     map[string]any `json:"filters"`
		Metadata    map[string]any `json:"metadata"`
	}
	_ = json.Unmarshal(body, &report)

	if report.ID == "" {
		t.Fatalf("Report ID is empty")
	}
	if report.Status == "" {
		t.Fatalf("Report status is empty")
	}
	if report.TemplateID == "" {
		t.Fatalf("Report templateId is empty")
	}
	if report.CreatedAt == "" {
		t.Fatalf("Report createdAt is empty")
	}
	if report.UpdatedAt == "" {
		t.Fatalf("Report updatedAt is empty")
	}

	if report.ID != reportID {
		t.Fatalf("Report ID mismatch: expected %s, got %s", reportID, report.ID)
	}

	validStatuses := []string{"Processing", "Finished", "Error"}
	statusValid := false
	for _, validStatus := range validStatuses {
		if report.Status == validStatus {
			statusValid = true
			break
		}
	}
	if !statusValid {
		t.Fatalf("Invalid report status: %s", report.Status)
	}

	t.Logf("✅ Report retrieved successfully:")
	t.Logf("   - ID: %s", report.ID)
	t.Logf("   - Status: %s", report.Status)
	t.Logf("   - TemplateID: %s", report.TemplateID)
	t.Logf("   - CreatedAt: %s", report.CreatedAt)
	t.Logf("   - UpdatedAt: %s", report.UpdatedAt)
	if report.CompletedAt != "" {
		t.Logf("   - CompletedAt: %s", report.CompletedAt)
	}
}

// TestIntegration_Reports_GetByID_InvalidID tests GET /v1/reports/{id} with an invalid report ID
func TestIntegration_Reports_GetByID_InvalidID(t *testing.T) {
	env := h.LoadEnvironment()
	if env.DefaultOrgID == "" {
		t.Skip("X-Organization-Id not configured; set ORG_ID or X_ORGANIZATION_ID")
	}
	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg(env.DefaultOrgID)

	invalidID := "00000000-0000-0000-0000-000000000000"
	code, body, err := cli.Request(ctx, "GET", fmt.Sprintf("/v1/reports/%s", invalidID), headers, nil)
	if err != nil {
		t.Fatalf("get report by invalid ID error: %v", err)
	}

	if code != 404 {
		t.Fatalf("Expected 404 for invalid report ID, got: code=%d body=%s", code, string(body))
	}

	var errorResp struct {
		Title   string `json:"title"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &errorResp)

	if errorResp.Title == "" {
		t.Fatalf("Error response missing title")
	}
	if errorResp.Code == "" {
		t.Fatalf("Error response missing code")
	}
	if errorResp.Message == "" {
		t.Fatalf("Error response missing message")
	}

	t.Logf("✅ Invalid report ID handled correctly:")
	t.Logf("   - Status: 404")
	t.Logf("   - Title: %s", errorResp.Title)
	t.Logf("   - Code: %s", errorResp.Code)
	t.Logf("   - Message: %s", errorResp.Message)
}

// TestIntegration_Reports_GetByID_StatusFinished tests GET /v1/reports/{id} for a finished report
func TestIntegration_Reports_GetByID_StatusFinished(t *testing.T) {
	env := h.LoadEnvironment()
	if env.DefaultOrgID == "" {
		t.Skip("X-Organization-Id not configured; set ORG_ID or X_ORGANIZATION_ID")
	}
	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg(env.DefaultOrgID)

	code, body, err := cli.Request(ctx, "GET", "/v1/reports?status=Finished&limit=1", headers, nil)
	if err != nil {
		t.Fatalf("list finished reports error: %v", err)
	}
	if code != 200 {
		t.Fatalf("list finished reports failed: code=%d body=%s", code, string(body))
	}

	var reports struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	_ = json.Unmarshal(body, &reports)

	if len(reports.Items) == 0 {
		t.Skip("No finished reports found to test")
	}

	reportID := reports.Items[0].ID
	t.Logf("Testing GET /v1/reports/%s (finished report)", reportID)

	code, body, err = cli.Request(ctx, "GET", fmt.Sprintf("/v1/reports/%s", reportID), headers, nil)
	if err != nil {
		t.Fatalf("get finished report error: %v", err)
	}
	if code != 200 {
		t.Fatalf("get finished report failed: code=%d body=%s", code, string(body))
	}

	var report struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		CompletedAt string `json:"completedAt"`
		CreatedAt   string `json:"createdAt"`
		UpdatedAt   string `json:"updatedAt"`
	}
	_ = json.Unmarshal(body, &report)

	if report.Status != "Finished" {
		t.Fatalf("Expected report status 'finished', got '%s'", report.Status)
	}

	if report.CompletedAt == "" {
		t.Fatalf("Finished report should have completedAt field filled")
	}

	if report.CreatedAt != "" && report.CompletedAt != "" {
		t.Logf("✅ Report completion timeline:")
		t.Logf("   - CreatedAt: %s", report.CreatedAt)
		t.Logf("   - CompletedAt: %s", report.CompletedAt)
	}

	t.Logf("✅ Finished report retrieved successfully:")
	t.Logf("   - ID: %s", report.ID)
	t.Logf("   - Status: %s", report.Status)
	t.Logf("   - CompletedAt: %s", report.CompletedAt)
}
