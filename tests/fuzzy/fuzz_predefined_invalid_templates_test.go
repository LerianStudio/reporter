// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package fuzzy

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	h "github.com/LerianStudio/reporter/v4/tests/helpers"
)

// TestPredefinedInvalidTemplates tests pre-defined templates that should fail gracefully
func TestPredefinedInvalidTemplates(t *testing.T) {
	env := h.LoadEnvironment()

	if env.DefaultOrgID == "" {
		t.Skip("X-Organization-Id not configured; set ORG_ID or X_ORGANIZATION_ID")
	}

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg(env.DefaultOrgID)

	// Get all template files from templates directory
	templatesDir := "./templates"
	templateFiles, err := filepath.Glob(filepath.Join(templatesDir, "*.tpl"))
	if err != nil {
		t.Fatalf("Failed to find template files: %v", err)
	}

	if len(templateFiles) == 0 {
		t.Skip("No template files found in ./templates directory")
	}

	for _, templateFile := range templateFiles {
		templateName := filepath.Base(templateFile)

		t.Run(templateName, func(t *testing.T) {
			// Read template content
			content, err := os.ReadFile(templateFile)
			if err != nil {
				t.Fatalf("Failed to read template %s: %v", templateName, err)
			}

			t.Logf("Testing invalid template: %s", templateName)

			// Upload template
			files := map[string][]byte{
				"template": content,
			}

			formData := map[string]string{
				"outputFormat": "TXT",
				"description":  "Invalid template test: " + templateName,
			}

			code, body, err := cli.UploadMultipartForm(ctx, "POST", "/v1/templates", headers, formData, files)
			if err != nil {
				t.Logf("Request error (may be expected): %v", err)
				return
			}

			// Server should NEVER crash (5xx)
			if code >= 500 {
				t.Fatalf("SERVER ERROR on template %s: code=%d body=%s", templateName, code, string(body))
			}

			// Template creation might succeed or fail with 4xx
			if code == 200 || code == 201 {
				var resp struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(body, &resp); err == nil && resp.ID != "" {
					t.Logf("Template %s accepted with ID: %s", templateName, resp.ID)

					// Try to generate report (expected to fail at render time)
					payload := map[string]any{
						"templateId": resp.ID,
						"filters": map[string]any{
							"midaz_onboarding": map[string]any{
								"organization": map[string]any{
									"id": map[string]any{
										"eq": []string{env.DefaultOrgID},
									},
								},
							},
						},
					}

					reportCode, reportBody, reportErr := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
					if reportErr != nil {
						t.Logf("Report generation request failed: %v", reportErr)
						return
					}

					// Server should NEVER crash (5xx)
					if reportCode >= 500 {
						t.Fatalf("SERVER ERROR on report generation for %s: code=%d body=%s", templateName, reportCode, string(reportBody))
					}

					// If report was created, check status
					if reportCode == 200 || reportCode == 201 {
						var reportResp struct {
							ID string `json:"id"`
						}
						if err := json.Unmarshal(reportBody, &reportResp); err == nil && reportResp.ID != "" {
							t.Logf("Report created: %s, waiting for processing...", reportResp.ID)

							// Wait for processing
							time.Sleep(5 * time.Second)

							// Check status
							statusCode, statusBody, _ := cli.Request(ctx, "GET", "/v1/reports/"+reportResp.ID, headers, nil)

							// Server should NEVER crash (5xx)
							if statusCode >= 500 {
								t.Fatalf("SERVER ERROR on status check for %s: code=%d body=%s", templateName, statusCode, string(statusBody))
							}

							if statusCode == 200 {
								var statusResp struct {
									Status string `json:"status"`
								}
								if err := json.Unmarshal(statusBody, &statusResp); err == nil {
									t.Logf("Report status for %s: %s", templateName, statusResp.Status)

									// Expected to be "Error" for invalid templates
									if statusResp.Status == "Error" {
										t.Logf("✅ Template %s correctly failed with Error status", templateName)
									} else if statusResp.Status == "Finished" {
										t.Errorf("❌ Template %s unexpectedly succeeded but was expected to fail", templateName)
									} else {
										t.Logf("Report status: %s", statusResp.Status)
									}
								}
							}
						}
					} else {
						t.Logf("Report creation rejected for %s: code=%d", templateName, reportCode)
					}
				}
			} else {
				t.Logf("Template %s rejected at creation: code=%d", templateName, code)
			}
		})
	}
}
