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
	assert.NoError(t, err)
	// Invalid number defaults to 0
	assert.Equal(t, 0, result.Limit)
}

func TestValidateParameters_InvalidPage(t *testing.T) {
	params := map[string]string{
		"page": "not-a-number",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	// Invalid number defaults to 0
	assert.Equal(t, 0, result.Page)
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
