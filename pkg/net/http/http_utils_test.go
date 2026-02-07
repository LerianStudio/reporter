// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package http

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateParameters_Defaults(t *testing.T) {
	params := map[string]string{}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)

	// Check default values
	assert.Equal(t, 10, result.Limit)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, "desc", result.SortOrder)
	assert.Equal(t, "", result.OutputFormat)
	assert.Equal(t, "", result.Description)
	assert.Equal(t, "", result.Status)
	assert.Equal(t, "", result.Cursor)
	assert.False(t, result.UseMetadata)
	assert.Nil(t, result.Metadata)
}

func TestValidateParameters_AllParameters(t *testing.T) {
	templateID := uuid.New()

	params := map[string]string{
		"outputFormat": "PDF",
		"description":  "Test description",
		"status":       "Finished",
		"templateId":   templateID.String(),
		"limit":        "20",
		"page":         "2",
		"sortOrder":    "asc",
		"createdAt":    "2024-01-15",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)

	assert.Equal(t, "PDF", result.OutputFormat)
	assert.Equal(t, "Test description", result.Description)
	assert.Equal(t, "Finished", result.Status)
	assert.Equal(t, templateID, result.TemplateID)
	assert.Equal(t, 20, result.Limit)
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, "asc", result.SortOrder)
	assert.Equal(t, 2024, result.CreatedAt.Year())
	assert.Equal(t, 1, int(result.CreatedAt.Month()))
	assert.Equal(t, 15, result.CreatedAt.Day())
}

func TestValidateParameters_Metadata(t *testing.T) {
	params := map[string]string{
		"metadata.customField": "customValue",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)

	assert.True(t, result.UseMetadata)
	assert.NotNil(t, result.Metadata)
}

func TestValidateParameters_InvalidOutputFormat(t *testing.T) {
	params := map[string]string{
		"outputFormat": "INVALID_FORMAT",
	}

	_, err := ValidateParameters(params)
	assert.Error(t, err)
}

func TestValidateParameters_ValidOutputFormats(t *testing.T) {
	formats := []string{"PDF", "pdf", "HTML", "html", "CSV", "csv", "XML", "xml", "TXT", "txt"}

	for _, format := range formats {
		t.Run("Format_"+format, func(t *testing.T) {
			params := map[string]string{
				"outputFormat": format,
			}

			result, err := ValidateParameters(params)
			assert.NoError(t, err)
			assert.Equal(t, format, result.OutputFormat)
		})
	}
}

func TestValidateParameters_InvalidSortOrder(t *testing.T) {
	params := map[string]string{
		"sortOrder": "invalid",
	}

	_, err := ValidateParameters(params)
	assert.Error(t, err)
}

func TestValidateParameters_ValidSortOrders(t *testing.T) {
	sortOrders := []string{"asc", "ASC", "desc", "DESC", "Asc", "Desc"}

	for _, order := range sortOrders {
		t.Run("SortOrder_"+order, func(t *testing.T) {
			params := map[string]string{
				"sortOrder": order,
			}

			result, err := ValidateParameters(params)
			assert.NoError(t, err)
			// Result is lowercased
			assert.Contains(t, []string{"asc", "desc"}, result.SortOrder)
		})
	}
}

func TestValidateParameters_PaginationLimitExceeded(t *testing.T) {
	// Set max pagination limit for test
	t.Setenv("MAX_PAGINATION_LIMIT", "100")

	params := map[string]string{
		"limit": "150",
	}

	_, err := ValidateParameters(params)
	assert.Error(t, err)
}

func TestValidateParameters_ValidCursor(t *testing.T) {
	cursor := Cursor{
		ID:         "123",
		PointsNext: true,
	}
	cursorJSON, _ := json.Marshal(cursor)
	encodedCursor := base64.StdEncoding.EncodeToString(cursorJSON)

	params := map[string]string{
		"cursor": encodedCursor,
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.Equal(t, encodedCursor, result.Cursor)
}

func TestValidateParameters_InvalidCursor(t *testing.T) {
	params := map[string]string{
		"cursor": "invalid-cursor-not-base64",
	}

	_, err := ValidateParameters(params)
	assert.Error(t, err)
}

func TestValidateParameters_InvalidTemplateID(t *testing.T) {
	params := map[string]string{
		"templateId": "not-a-uuid",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	// Invalid UUID results in zero UUID
	assert.Equal(t, uuid.UUID{}, result.TemplateID)
}

func TestValidateParameters_InvalidLimit(t *testing.T) {
	params := map[string]string{
		"limit": "not-a-number",
	}

	result, err := ValidateParameters(params)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "TPL-0019")
}

func TestValidateParameters_InvalidPage(t *testing.T) {
	params := map[string]string{
		"page": "not-a-number",
	}

	result, err := ValidateParameters(params)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "TPL-0019")
}

func TestQueryHeader_ToOffsetPagination(t *testing.T) {
	qh := &QueryHeader{
		Limit:     20,
		Page:      3,
		SortOrder: "asc",
		Alias:     "test_alias",
		Cursor:    "some_cursor",
	}

	pagination := qh.ToOffsetPagination()

	assert.Equal(t, 20, pagination.Limit)
	assert.Equal(t, 3, pagination.Page)
	assert.Equal(t, "asc", pagination.SortOrder)
	assert.Equal(t, "test_alias", pagination.Alias)
	// Cursor is not included in ToOffsetPagination
	assert.Empty(t, pagination.Cursor)
}

func TestPagination_Struct(t *testing.T) {
	pagination := Pagination{
		Limit:     10,
		Page:      1,
		Cursor:    "cursor",
		SortOrder: "desc",
		Alias:     "alias",
	}

	assert.Equal(t, 10, pagination.Limit)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, "cursor", pagination.Cursor)
	assert.Equal(t, "desc", pagination.SortOrder)
	assert.Equal(t, "alias", pagination.Alias)
}

func TestHeaderConstants(t *testing.T) {
	// Test that constants are defined
	assert.Equal(t, "User-Agent", headerUserAgent)
	assert.Equal(t, ".tpl", fileExtension)
	assert.Equal(t, "X-TTL", idempotencyTTL)
}

// TestValidateParameters_QueryParamParseErrors verifies that ValidateParameters
// returns a validation error for non-numeric limit/page values and out-of-bounds
// values instead of silently defaulting.
// REFACTOR-005A: These tests must FAIL against the current code.
func TestValidateParameters_QueryParamParseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		params      map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name:        "non-numeric limit returns error",
			params:      map[string]string{"limit": "abc"},
			wantErr:     true,
			errContains: "TPL-0019",
		},
		{
			name:        "non-numeric page returns error",
			params:      map[string]string{"page": "xyz"},
			wantErr:     true,
			errContains: "TPL-0019",
		},
		{
			name:        "negative limit returns error",
			params:      map[string]string{"limit": "-1"},
			wantErr:     true,
			errContains: "TPL-0019",
		},
		{
			name:        "zero page returns error",
			params:      map[string]string{"page": "0"},
			wantErr:     true,
			errContains: "TPL-0019",
		},
		{
			name:        "zero limit returns error",
			params:      map[string]string{"limit": "0"},
			wantErr:     true,
			errContains: "TPL-0019",
		},
		{
			name:        "negative page returns error",
			params:      map[string]string{"page": "-5"},
			wantErr:     true,
			errContains: "TPL-0019",
		},
		{
			name:        "float limit returns error",
			params:      map[string]string{"limit": "10.5"},
			wantErr:     true,
			errContains: "TPL-0019",
		},
		{
			name:        "float page returns error",
			params:      map[string]string{"page": "1.5"},
			wantErr:     true,
			errContains: "TPL-0019",
		},
		{
			name:    "valid limit and page succeeds",
			params:  map[string]string{"limit": "10", "page": "1"},
			wantErr: false,
		},
		{
			name:    "valid large page succeeds",
			params:  map[string]string{"limit": "25", "page": "100"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ValidateParameters(tt.params)

			if tt.wantErr {
				require.Error(t, err, "expected error for params %v but got nil", tt.params)
				assert.Contains(t, err.Error(), tt.errContains,
					"error should contain code %q, got: %s", tt.errContains, err.Error())
				assert.Nil(t, result, "result should be nil when error is returned")
			} else {
				require.NoError(t, err, "unexpected error for params %v: %v", tt.params, err)
				assert.NotNil(t, result, "result should not be nil for valid params")
				assert.Greater(t, result.Limit, 0, "limit must be positive")
				assert.GreaterOrEqual(t, result.Page, 1, "page must be >= 1")
			}
		})
	}
}
