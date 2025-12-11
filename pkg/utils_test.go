package pkg

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMapNumKinds(t *testing.T) {
	numKinds := GetMapNumKinds()

	assert.True(t, numKinds[reflect.Int])
	assert.True(t, numKinds[reflect.Int8])
	assert.True(t, numKinds[reflect.Int16])
	assert.True(t, numKinds[reflect.Int32])
	assert.True(t, numKinds[reflect.Int64])
	assert.True(t, numKinds[reflect.Float32])
	assert.True(t, numKinds[reflect.Float64])
	assert.False(t, numKinds[reflect.String])
}

func TestIsNilOrEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected bool
	}{
		{
			name:     "nil string",
			input:    nil,
			expected: true,
		},
		{
			name:     "empty string",
			input:    strPtr(""),
			expected: true,
		},
		{
			name:     "whitespace only",
			input:    strPtr("   "),
			expected: true,
		},
		{
			name:     "null string",
			input:    strPtr("null"),
			expected: true,
		},
		{
			name:     "nil string literal",
			input:    strPtr("nil"),
			expected: true,
		},
		{
			name:     "valid string",
			input:    strPtr("value"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNilOrEmpty(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateFormDataFields(t *testing.T) {
	tests := []struct {
		name        string
		outFormat   *string
		description *string
		expectError bool
	}{
		{
			name:        "valid fields",
			outFormat:   strPtr("html"),
			description: strPtr("Test description"),
			expectError: false,
		},
		{
			name:        "nil outFormat",
			outFormat:   nil,
			description: strPtr("Test"),
			expectError: true,
		},
		{
			name:        "nil description",
			outFormat:   strPtr("html"),
			description: nil,
			expectError: true,
		},
		{
			name:        "invalid outFormat",
			outFormat:   strPtr("invalid"),
			description: strPtr("Test"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormDataFields(tt.outFormat, tt.description)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsOutputFormatValuesValid(t *testing.T) {
	tests := []struct {
		format   string
		expected bool
	}{
		{"html", true},
		{"HTML", true},
		{"pdf", true},
		{"PDF", true},
		{"csv", true},
		{"CSV", true},
		{"xml", true},
		{"XML", true},
		{"txt", true},
		{"TXT", true},
		{"invalid", false},
		{"doc", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			format := tt.format
			result := IsOutputFormatValuesValid(&format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateFileFormat(t *testing.T) {
	tests := []struct {
		name         string
		outFormat    string
		templateFile string
		expectError  bool
	}{
		{
			name:         "valid HTML",
			outFormat:    "html",
			templateFile: "<html><body>Test</body></html>",
			expectError:  false,
		},
		{
			name:         "valid HTML with DOCTYPE",
			outFormat:    "HTML",
			templateFile: "<!DOCTYPE html><html>Test</html>",
			expectError:  false,
		},
		{
			name:         "invalid HTML",
			outFormat:    "html",
			templateFile: "plain text",
			expectError:  true,
		},
		{
			name:         "valid PDF (HTML content)",
			outFormat:    "pdf",
			templateFile: "<html><body>Test</body></html>",
			expectError:  false,
		},
		{
			name:         "valid XML",
			outFormat:    "xml",
			templateFile: "<?xml version=\"1.0\"?><root></root>",
			expectError:  false,
		},
		{
			name:         "valid XML with just tag",
			outFormat:    "xml",
			templateFile: "<root><item>Test</item></root>",
			expectError:  false,
		},
		{
			name:         "invalid XML",
			outFormat:    "xml",
			templateFile: "plain text without tags",
			expectError:  true,
		},
		{
			name:         "valid CSV",
			outFormat:    "csv",
			templateFile: "col1,col2\nval1,val2",
			expectError:  false,
		},
		{
			name:         "valid CSV with semicolon",
			outFormat:    "CSV",
			templateFile: "col1;col2\nval1;val2",
			expectError:  false,
		},
		{
			name:         "invalid CSV - single line",
			outFormat:    "csv",
			templateFile: "single line without separator",
			expectError:  true,
		},
		{
			name:         "valid TXT",
			outFormat:    "txt",
			templateFile: "Some text content",
			expectError:  false,
		},
		{
			name:         "invalid TXT - empty",
			outFormat:    "txt",
			templateFile: "   ",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileFormat(tt.outFormat, tt.templateFile)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateServerAddress(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"localhost:8080", "localhost:8080"},
		{"192.168.1.1:3000", "192.168.1.1:3000"},
		{"example.com:443", "example.com:443"},
		{"invalid", ""},
		{"no-port", ""},
		{":8080", ""},
		{"host:", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ValidateServerAddress(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeInt64ToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int
	}{
		{
			name:     "normal value",
			input:    100,
			expected: 100,
		},
		{
			name:     "zero",
			input:    0,
			expected: 0,
		},
		{
			name:     "negative",
			input:    -100,
			expected: -100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeInt64ToInt(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}
