package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	h "plugin-smart-templates/v2/tests/helpers"
)

// GET /v1/templates — filters and pagination
func TestIntegration_Templates_ListWithFiltersAndPagination(t *testing.T) {
	t.Parallel()
	env := h.LoadEnvironment()
	if env.DefaultOrgID == "" {
		t.Skip("X-Organization-Id not configured; set ORG_ID or X_ORGANIZATION_ID")
	}
	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg(env.DefaultOrgID)

	path1 := "/v1/templates?limit=1&page=1&outputFormat=HTML"
	code, body, err := cli.Request(ctx, "GET", path1, headers, nil)
	if err != nil || code != 200 {
		t.Fatalf("list page1 code=%d err=%v body=%s", code, err, string(body))
	}
	var page1 struct {
		Items []map[string]any `json:"items"`
	}
	_ = json.Unmarshal(body, &page1)

	path2 := "/v1/templates?limit=1&page=2&outputFormat=HTML"
	code, body, err = cli.Request(ctx, "GET", path2, headers, nil)
	if err != nil || code != 200 {
		t.Fatalf("list page2 code=%d err=%v body=%s", code, err, string(body))
	}
	var page2 struct {
		Items []map[string]any `json:"items"`
	}
	_ = json.Unmarshal(body, &page2)

	seen := map[string]bool{}
	for _, it := range page1.Items {
		if id, ok := it["id"].(string); ok {
			seen[id] = true
		}
	}
	for _, it := range page2.Items {
		if id, ok := it["id"].(string); ok {
			if seen[id] {
				t.Fatalf("duplicate across pages: %s", id)
			}
		}
	}
}

// POST /v1/templates — create template with invalid payload
func TestIntegration_Templates_Create_BadRequest(t *testing.T) {
	t.Parallel()
	env := h.LoadEnvironment()
	ctx := context.Background()
	cli := h.NewHTTPClient(env.ManagerURL, env.HTTPTimeout)
	headers := h.AuthHeadersWithOrg("00000000-0000-0000-0000-000000000000")

	payload := map[string]any{"description": "x", "outputFormat": "HTML", "templateFile": fmt.Sprintf("%s", "not-binary")}
	code, body, err := cli.Request(ctx, "POST", "/v1/templates", headers, payload)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	if code < 400 || code >= 500 {
		t.Fatalf("expected 4xx, got %d body=%s", code, string(body))
	}
}
