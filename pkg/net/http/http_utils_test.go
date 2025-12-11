package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime/multipart"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidateParameters_EmptyParams(t *testing.T) {
	params := map[string]string{}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Check defaults
	assert.Equal(t, 10, result.Limit)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, "desc", result.SortOrder)
}

func TestValidateParameters_WithLimit(t *testing.T) {
	params := map[string]string{
		"limit": "50",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.Equal(t, 50, result.Limit)
}

func TestValidateParameters_WithPage(t *testing.T) {
	params := map[string]string{
		"page": "5",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.Equal(t, 5, result.Page)
}

func TestValidateParameters_WithDescription(t *testing.T) {
	params := map[string]string{
		"description": "test description",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.Equal(t, "test description", result.Description)
}

func TestValidateParameters_WithStatus(t *testing.T) {
	params := map[string]string{
		"status": "active",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.Equal(t, "active", result.Status)
}

func TestValidateParameters_WithValidTemplateID(t *testing.T) {
	validUUID := uuid.New().String()
	params := map[string]string{
		"templateId": validUUID,
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.Equal(t, validUUID, result.TemplateID.String())
}

func TestValidateParameters_WithInvalidTemplateID(t *testing.T) {
	params := map[string]string{
		"templateId": "not-a-valid-uuid",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	// Invalid UUID should result in zero UUID
	assert.Equal(t, uuid.Nil, result.TemplateID)
}

func TestValidateParameters_WithSortOrderAsc(t *testing.T) {
	params := map[string]string{
		"sortOrder": "ASC",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.Equal(t, "asc", result.SortOrder) // Should be lowercased
}

func TestValidateParameters_WithSortOrderDesc(t *testing.T) {
	params := map[string]string{
		"sortOrder": "DESC",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.Equal(t, "desc", result.SortOrder)
}

func TestValidateParameters_WithInvalidSortOrder(t *testing.T) {
	params := map[string]string{
		"sortOrder": "invalid",
	}

	_, err := ValidateParameters(params)
	assert.Error(t, err)
}

func TestValidateParameters_WithCreatedAt(t *testing.T) {
	params := map[string]string{
		"createdAt": "2025-06-15",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.Equal(t, 2025, result.CreatedAt.Year())
	assert.Equal(t, 6, int(result.CreatedAt.Month()))
	assert.Equal(t, 15, result.CreatedAt.Day())
}

func TestValidateParameters_WithMetadata(t *testing.T) {
	params := map[string]string{
		"metadata.key": "value",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.NotNil(t, result.Metadata)
	assert.True(t, result.UseMetadata)
}

func TestValidateParameters_WithValidOutputFormat(t *testing.T) {
	// Valid formats: html, pdf, csv, xml, txt
	tests := []string{"html", "pdf", "csv", "xml", "txt"}

	for _, format := range tests {
		t.Run(format, func(t *testing.T) {
			params := map[string]string{
				"outputFormat": format,
			}

			result, err := ValidateParameters(params)
			assert.NoError(t, err)
			assert.Equal(t, format, result.OutputFormat)
		})
	}
}

func TestValidateParameters_WithInvalidOutputFormat(t *testing.T) {
	params := map[string]string{
		"outputFormat": "invalid-format",
	}

	_, err := ValidateParameters(params)
	assert.Error(t, err)
}

func TestValidateParameters_WithValidCursor(t *testing.T) {
	cursor := Cursor{
		ID:         "test-id",
		PointsNext: true,
	}
	jsonData, _ := json.Marshal(cursor)
	encoded := base64.StdEncoding.EncodeToString(jsonData)

	params := map[string]string{
		"cursor": encoded,
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.Equal(t, encoded, result.Cursor)
}

func TestValidateParameters_WithInvalidCursor(t *testing.T) {
	params := map[string]string{
		"cursor": "invalid-cursor-not-base64",
	}

	_, err := ValidateParameters(params)
	assert.Error(t, err)
}

func TestValidateParameters_LimitExceedsMax(t *testing.T) {
	// Default max is 100
	params := map[string]string{
		"limit": "150",
	}

	_, err := ValidateParameters(params)
	assert.Error(t, err)
}

func TestValidateParameters_AllParams(t *testing.T) {
	validUUID := uuid.New().String()
	cursor := Cursor{ID: "id", PointsNext: true}
	jsonData, _ := json.Marshal(cursor)
	encodedCursor := base64.StdEncoding.EncodeToString(jsonData)

	params := map[string]string{
		"limit":        "20",
		"page":         "2",
		"sortOrder":    "asc",
		"templateId":   validUUID,
		"description":  "test desc",
		"status":       "completed",
		"createdAt":    "2025-01-15",
		"cursor":       encodedCursor,
		"outputFormat": "pdf",
	}

	result, err := ValidateParameters(params)
	assert.NoError(t, err)
	assert.Equal(t, 20, result.Limit)
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, "asc", result.SortOrder)
	assert.Equal(t, validUUID, result.TemplateID.String())
	assert.Equal(t, "test desc", result.Description)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, encodedCursor, result.Cursor)
	assert.Equal(t, "pdf", result.OutputFormat)
}

// Helper to create multipart file header for testing
func createTestFileHeader(filename string, content string) (*multipart.FileHeader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	_, err = io.WriteString(part, content)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(1024 * 1024)
	if err != nil {
		return nil, err
	}

	files := form.File["file"]
	if len(files) == 0 {
		return nil, nil
	}

	return files[0], nil
}

func TestGetFileFromHeader_ValidTPLFile(t *testing.T) {
	content := "template content here"
	fileHeader, err := createTestFileHeader("test.tpl", content)
	assert.NoError(t, err)

	result, err := GetFileFromHeader(fileHeader)
	assert.NoError(t, err)
	assert.Equal(t, content, result)
}

func TestGetFileFromHeader_InvalidExtension(t *testing.T) {
	fileHeader, err := createTestFileHeader("test.txt", "content")
	assert.NoError(t, err)

	_, err = GetFileFromHeader(fileHeader)
	assert.Error(t, err)
}

func TestGetFileFromHeader_EmptyFile(t *testing.T) {
	// Create file header with empty content
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create an empty file
	part, err := writer.CreateFormFile("file", "empty.tpl")
	assert.NoError(t, err)
	// Write nothing to part

	_ = part
	err = writer.Close()
	assert.NoError(t, err)

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(1024 * 1024)
	assert.NoError(t, err)

	files := form.File["file"]
	assert.Len(t, files, 1)

	_, err = GetFileFromHeader(files[0])
	assert.Error(t, err)
}

func TestReadMultipartFile_ValidFile(t *testing.T) {
	content := "file content for reading"
	fileHeader, err := createTestFileHeader("test.txt", content)
	assert.NoError(t, err)

	result, err := ReadMultipartFile(fileHeader)
	assert.NoError(t, err)
	assert.Equal(t, content, string(result))
}

func TestReadMultipartFile_BinaryContent(t *testing.T) {
	content := "\x00\x01\x02\x03binary\xff\xfe"
	fileHeader, err := createTestFileHeader("binary.bin", content)
	assert.NoError(t, err)

	result, err := ReadMultipartFile(fileHeader)
	assert.NoError(t, err)
	assert.Equal(t, []byte(content), result)
}

func TestValidatePagination_ThroughValidateParameters(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]string
		expectError bool
	}{
		{
			name:        "valid pagination",
			params:      map[string]string{"limit": "50", "sortOrder": "asc"},
			expectError: false,
		},
		{
			name:        "limit at max",
			params:      map[string]string{"limit": "100"},
			expectError: false,
		},
		{
			name:        "limit over max",
			params:      map[string]string{"limit": "101"},
			expectError: true,
		},
		{
			name:        "invalid sort order",
			params:      map[string]string{"sortOrder": "random"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateParameters(tt.params)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateParameters_CustomMaxPaginationLimit(t *testing.T) {
	// Save original value
	original := os.Getenv("MAX_PAGINATION_LIMIT")
	defer os.Setenv("MAX_PAGINATION_LIMIT", original)

	// Set custom max limit
	os.Setenv("MAX_PAGINATION_LIMIT", "50")

	params := map[string]string{
		"limit": "51",
	}

	_, err := ValidateParameters(params)
	assert.Error(t, err)

	// Should work with limit at 50
	params["limit"] = "50"
	_, err = ValidateParameters(params)
	assert.NoError(t, err)
}
