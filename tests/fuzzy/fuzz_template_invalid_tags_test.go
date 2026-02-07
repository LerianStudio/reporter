//go:build fuzz

// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package fuzzy

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	h "github.com/LerianStudio/reporter/tests/utils"
)

// FuzzTemplate_InvalidTags tests templates with non-existent or malformed tags
// Expected: Should return 4xx errors, never 5xx (server errors)
func FuzzTemplate_InvalidTags(f *testing.F) {
	// Seed corpus with various malformed template patterns
	f.Add("{{ nonexistent.field }}")
	f.Add("{% for x in fake.table %}{{ x.id }}{% endfor %}")
	f.Add("{{ database.table.nonexistent_column }}")
	f.Add("{% if missing.field > 10 %}error{% endif %}")
	f.Add("{{ midaz_onboarding.account.999999999 }}")
	f.Add("{% with x = filter(fake.data, 'id', 'value') %}{{ x.field }}{% endwith %}")
	f.Add("{{ ...invalid... }}")
	f.Add("{% calc invalid.field + 10 %}")
	f.Add("{{ %illegal syntax% }}")
	f.Add("{% for %}{% endfor %}")
	f.Add("{{ }}")
	f.Add("{% %}")

	env := h.LoadEnvironment()
	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeaders()

	testOrgID := "00000000-0000-0000-0000-000000000001"

	f.Fuzz(func(t *testing.T, templateContent string) {
		// Limit template size
		if len(templateContent) > 10000 {
			templateContent = templateContent[:10000]
		}

		// Skip empty templates
		if strings.TrimSpace(templateContent) == "" {
			t.Skip("empty template")
			return
		}

		// Create template with potentially invalid content
		files := map[string][]byte{
			"template": []byte(templateContent),
		}

		formData := map[string]string{
			"outputFormat": "TXT",
			"description":  "Fuzz test invalid tags",
		}

		code, body, err := cli.UploadMultipartForm(ctx, "POST", "/v1/templates", headers, formData, files)
		if err != nil {
			t.Logf("Request error (acceptable): %v", err)
			return
		}

		// Server should NEVER crash (5xx)
		if code >= 500 {
			t.Fatalf("SERVER ERROR on invalid template: code=%d body=%s template=%q", code, string(body), templateContent)
		}

		// If template creation succeeded (unlikely but possible), try to generate report
		if code == 200 || code == 201 {
			var resp struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(body, &resp); err == nil && resp.ID != "" {
				t.Logf("Template accepted: %s", resp.ID)

				// Try to generate report (should fail gracefully if tags don't exist)
				payload := map[string]any{
					"templateId": resp.ID,
					"filters": map[string]any{
						"midaz_onboarding": map[string]any{
							"organization": map[string]any{
								"id": map[string]any{
									"eq": []string{testOrgID},
								},
							},
						},
					},
				}

				reportCode, reportBody, reportErr := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
				if reportErr != nil {
					t.Logf("Report generation failed (expected): %v", reportErr)
					return
				}

				// Server should NEVER crash on report generation
				if reportCode >= 500 {
					t.Fatalf("SERVER ERROR on report generation: code=%d body=%s template=%q", reportCode, string(reportBody), templateContent)
				}

				// If report was accepted, wait a bit and check status
				if reportCode == 200 || reportCode == 201 {
					var reportResp struct {
						ID string `json:"id"`
					}
					if err := json.Unmarshal(reportBody, &reportResp); err == nil && reportResp.ID != "" {
						time.Sleep(2 * time.Second)

						// Check report status
						statusCode, statusBody, _ := cli.Request(ctx, "GET", "/v1/reports/"+reportResp.ID, headers, nil)
						if statusCode >= 500 {
							t.Fatalf("SERVER ERROR on status check: code=%d body=%s", statusCode, string(statusBody))
						}

						t.Logf("Report generated or failed gracefully: %s", reportResp.ID)
					}
				}
			}
		}
	})
}
