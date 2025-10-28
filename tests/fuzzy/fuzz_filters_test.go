package fuzzy

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	h "github.com/LerianStudio/reporter/v4/tests/helpers"
)

// FuzzReportFilters tests report generation with various malformed filter inputs
// Expected: Should handle gracefully without server crashes
func FuzzReportFilters(f *testing.F) {
	// Seed corpus with various filter patterns
	f.Add(`{"eq": ["value"]}`)
	f.Add(`{"gt": [100]}`)
	f.Add(`{"invalid_operator": ["test"]}`)
	f.Add(`{"eq": []}`)
	f.Add(`{"eq": null}`)
	f.Add(`{"nested": {"deep": {"very": {"deep": "value"}}}}`)
	f.Add(`{"eq": ["', DROP TABLE users; --"]}`)
	f.Add(`{"eq": ["\u0000\u0001\u0002"]}`)

	env := h.LoadEnvironment()

	if env.DefaultOrgID == "" {
		f.Skip("X-Organization-Id not configured; set ORG_ID or X_ORGANIZATION_ID")
	}

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg(env.DefaultOrgID)

	// Use a fixed template ID for testing (will return 404 but that's ok, we're testing filter validation)
	templateID := "00000000-0000-0000-0000-000000000000"

	f.Fuzz(func(t *testing.T, filterJSON string) {
		// Limit size
		if len(filterJSON) > 5000 {
			filterJSON = filterJSON[:5000]
		}

		// Try to parse as JSON
		var filterData any
		if err := json.Unmarshal([]byte(filterJSON), &filterData); err != nil {
			// Invalid JSON, skip
			t.Skip("invalid JSON")
			return
		}

		// Construct report payload with fuzzed filters
		payload := map[string]any{
			"templateId": templateID,
			"filters": map[string]any{
				"midaz_onboarding": map[string]any{
					"organization": map[string]any{
						"id": filterData, // Fuzzed filter
					},
				},
			},
		}

		reportCode, reportBody, reportErr := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
		if reportErr != nil {
			t.Logf("Request error (acceptable): %v", reportErr)
			return
		}

		// Server should NEVER crash (5xx)
		if reportCode >= 500 {
			t.Fatalf("SERVER ERROR on fuzzed filter: code=%d body=%s filter=%s", reportCode, string(reportBody), filterJSON)
		}

		// If report was created, check that it doesn't crash the worker
		if reportCode == 200 || reportCode == 201 {
			var reportResp struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(reportBody, &reportResp); err == nil && reportResp.ID != "" {
				// Wait for processing
				time.Sleep(3 * time.Second)

				// Check status (should be either Finished or Error, never cause server crash)
				statusCode, statusBody, _ := cli.Request(ctx, "GET", "/v1/reports/"+reportResp.ID, headers, nil)
				if statusCode >= 500 {
					t.Fatalf("SERVER ERROR on status check: code=%d body=%s filter=%s", statusCode, string(statusBody), filterJSON)
				}

				t.Logf("Report processed: %s (filter: %s)", reportResp.ID, filterJSON)
			}
		}
	})
}
