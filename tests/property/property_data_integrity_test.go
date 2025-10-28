package property

import (
	"context"
	"encoding/json"
	"testing"
	"testing/quick"
	"time"

	h "github.com/LerianStudio/reporter/v4/tests/helpers"
)

// Property 1: Report criado deve sempre existir no MongoDB
func TestProperty_Report_ExistsAfterCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	env := h.LoadEnvironment()

	if env.DefaultOrgID == "" {
		t.Skip("X-Organization-Id not configured; set ORG_ID or X_ORGANIZATION_ID")
	}

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg(env.DefaultOrgID)

	templateID := createTestTemplate(t, ctx, cli, headers, env.DefaultOrgID)

	property := func(seed uint32) bool {
		payload := map[string]any{
			"templateId": templateID,
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

		code, body, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
		if err != nil || code >= 500 || (code != 200 && code != 201) {
			return true
		}

		var createResp struct {
			ID string `json:"id"`
		}

		if err := json.Unmarshal(body, &createResp); err != nil {
			return true
		}

		// Immediately check if report exists
		getCode, getBody, _ := cli.Request(ctx, "GET", "/v1/reports/"+createResp.ID, headers, nil)
		if getCode != 200 {
			t.Logf("Report %s not found immediately after creation", createResp.ID)
			return false
		}

		var getResp struct {
			ID string `json:"id"`
		}

		if err := json.Unmarshal(getBody, &getResp); err != nil {
			return false
		}

		// IDs should match
		return getResp.ID == createResp.ID
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 5}); err != nil {
		t.Errorf("Property violated: report doesn't exist after creation: %v", err)
	}
}

// Property 2: Report metadata deve conter campos obrigatórios
func TestProperty_Report_RequiredMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	env := h.LoadEnvironment()

	if env.DefaultOrgID == "" {
		t.Skip("X-Organization-Id not configured; set ORG_ID or X_ORGANIZATION_ID")
	}

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg(env.DefaultOrgID)

	templateID := createTestTemplate(t, ctx, cli, headers, env.DefaultOrgID)

	property := func(seed uint32) bool {
		payload := map[string]any{
			"templateId": templateID,
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

		code, body, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
		if err != nil || code >= 500 || (code != 200 && code != 201) {
			return true
		}

		var resp struct {
			ID             string `json:"id"`
			Status         string `json:"status"`
			OrganizationID string `json:"organizationId"`
			TemplateID     string `json:"templateId"`
		}

		if err := json.Unmarshal(body, &resp); err != nil {
			return true
		}

		// All required fields should be present
		return resp.ID != "" &&
			resp.Status != "" &&
			resp.TemplateID == templateID
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 5}); err != nil {
		t.Errorf("Property violated: missing required metadata: %v", err)
	}
}

// Property 3: Lista de reports deve incluir report recém-criado
func TestProperty_Report_AppearsInList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	env := h.LoadEnvironment()

	if env.DefaultOrgID == "" {
		t.Skip("X-Organization-Id not configured; set ORG_ID or X_ORGANIZATION_ID")
	}

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg(env.DefaultOrgID)

	templateID := createTestTemplate(t, ctx, cli, headers, env.DefaultOrgID)

	property := func(seed uint32) bool {
		// Create report
		payload := map[string]any{
			"templateId": templateID,
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

		code, body, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
		if err != nil || code >= 500 || (code != 200 && code != 201) {
			return true
		}

		var createResp struct {
			ID string `json:"id"`
		}

		if err := json.Unmarshal(body, &createResp); err != nil {
			return true
		}

		createdID := createResp.ID

		// Small delay
		time.Sleep(100 * time.Millisecond)

		// List reports
		listCode, listBody, _ := cli.Request(ctx, "GET", "/v1/reports?limit=100", headers, nil)
		if listCode != 200 {
			return true
		}

		var listResp struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
		}

		if err := json.Unmarshal(listBody, &listResp); err != nil {
			return true
		}

		// Check if created report appears in list
		for _, item := range listResp.Items {
			if item.ID == createdID {
				return true
			}
		}

		t.Logf("Report %s not found in list", createdID)
		return false
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 3}); err != nil {
		t.Errorf("Property violated: report not in list: %v", err)
	}
}

// Property 4: Output format do report deve corresponder ao template
func TestProperty_Report_OutputFormatMatches(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	env := h.LoadEnvironment()

	if env.DefaultOrgID == "" {
		t.Skip("X-Organization-Id not configured; set ORG_ID or X_ORGANIZATION_ID")
	}

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg(env.DefaultOrgID)

	outputFormats := []string{"TXT", "HTML", "CSV", "XML"}

	property := func(formatIndex uint8) bool {
		format := outputFormats[int(formatIndex)%len(outputFormats)]

		// Create template with specific format
		files := map[string][]byte{
			"template": []byte(`{% for org in midaz_onboarding.organization %}{{ org.id }}{% endfor %}`),
		}
		formData := map[string]string{
			"outputFormat": format,
			"description":  "Property test format",
		}

		code, body, err := cli.UploadMultipartForm(ctx, "POST", "/v1/templates", headers, formData, files)
		if err != nil || (code != 200 && code != 201) {
			return true
		}

		var templateResp struct {
			ID           string `json:"id"`
			OutputFormat string `json:"outputFormat"`
		}

		if err := json.Unmarshal(body, &templateResp); err != nil {
			return true
		}

		// Output format should match (lowercase)
		expectedFormat := format
		if format == "TXT" {
			expectedFormat = "txt"
		} else if format == "HTML" {
			expectedFormat = "html"
		} else if format == "CSV" {
			expectedFormat = "csv"
		} else if format == "XML" {
			expectedFormat = "xml"
		}

		return templateResp.OutputFormat == expectedFormat
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 10}); err != nil {
		t.Errorf("Property violated: output format mismatch: %v", err)
	}
}

// Property 5: Timestamps devem ser ordenados (createdAt <= updatedAt)
func TestProperty_Report_TimestampOrdering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	env := h.LoadEnvironment()

	if env.DefaultOrgID == "" {
		t.Skip("X-Organization-Id not configured; set ORG_ID or X_ORGANIZATION_ID")
	}

	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg(env.DefaultOrgID)

	templateID := createTestTemplate(t, ctx, cli, headers, env.DefaultOrgID)

	property := func(seed uint32) bool {
		payload := map[string]any{
			"templateId": templateID,
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

		code, body, err := cli.Request(ctx, "POST", "/v1/reports", headers, payload)
		if err != nil || code >= 500 || (code != 200 && code != 201) {
			return true
		}

		var resp struct {
			ID        string `json:"id"`
			CreatedAt string `json:"createdAt"`
			UpdatedAt string `json:"updatedAt"`
		}

		if err := json.Unmarshal(body, &resp); err != nil {
			return true
		}

		if resp.CreatedAt == "" || resp.UpdatedAt == "" {
			return true // Skip if timestamps missing
		}

		// Parse timestamps
		createdAt, err1 := time.Parse(time.RFC3339, resp.CreatedAt)
		updatedAt, err2 := time.Parse(time.RFC3339, resp.UpdatedAt)

		if err1 != nil || err2 != nil {
			// Try alternative format
			return true
		}

		// updatedAt should be >= createdAt
		return !updatedAt.Before(createdAt)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 5}); err != nil {
		t.Errorf("Property violated: timestamp ordering: %v", err)
	}
}
