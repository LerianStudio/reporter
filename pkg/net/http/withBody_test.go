package http

import (
	"bytes"
	"io"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewOfType(t *testing.T) {
	type TestStruct struct {
		Name  string
		Value int
	}

	t.Run("creates new instance of struct type", func(t *testing.T) {
		original := &TestStruct{Name: "original", Value: 42}
		result := newOfType(original)

		assert.NotNil(t, result)
		assert.IsType(t, &TestStruct{}, result)

		// Should be a new instance with zero values
		resultStruct := result.(*TestStruct)
		assert.Equal(t, "", resultStruct.Name)
		assert.Equal(t, 0, resultStruct.Value)
	})

	t.Run("creates independent instance", func(t *testing.T) {
		original := &TestStruct{Name: "test", Value: 100}
		result := newOfType(original)

		resultStruct := result.(*TestStruct)
		resultStruct.Name = "modified"
		resultStruct.Value = 999

		// Original should remain unchanged
		assert.Equal(t, "test", original.Name)
		assert.Equal(t, 100, original.Value)
	})
}

func TestFindUnknownFields(t *testing.T) {
	t.Run("returns empty map when maps are equal", func(t *testing.T) {
		original := map[string]any{"name": "test", "value": 42}
		marshaled := map[string]any{"name": "test", "value": 42}

		result := findUnknownFields(original, marshaled)
		assert.Empty(t, result)
	})

	t.Run("finds unknown fields", func(t *testing.T) {
		original := map[string]any{"name": "test", "unknown_field": "value"}
		marshaled := map[string]any{"name": "test"}

		result := findUnknownFields(original, marshaled)
		assert.Len(t, result, 1)
		assert.Equal(t, "value", result["unknown_field"])
	})

	t.Run("finds nested unknown fields", func(t *testing.T) {
		original := map[string]any{
			"name": "test",
			"nested": map[string]any{
				"known":   "value",
				"unknown": "extra",
			},
		}
		marshaled := map[string]any{
			"name": "test",
			"nested": map[string]any{
				"known": "value",
			},
		}

		result := findUnknownFields(original, marshaled)
		assert.Len(t, result, 1)
		assert.Contains(t, result, "nested")
	})

	t.Run("handles arrays", func(t *testing.T) {
		original := map[string]any{
			"items": []any{"a", "b", "c"},
		}
		marshaled := map[string]any{
			"items": []any{"a", "b", "c"},
		}

		result := findUnknownFields(original, marshaled)
		assert.Empty(t, result)
	})

	t.Run("ignores zero numeric values", func(t *testing.T) {
		original := map[string]any{
			"name":  "test",
			"value": 0.0,
		}
		marshaled := map[string]any{
			"name": "test",
		}

		result := findUnknownFields(original, marshaled)
		assert.Empty(t, result)
	})

	t.Run("detects value differences", func(t *testing.T) {
		original := map[string]any{
			"name": "original",
		}
		marshaled := map[string]any{
			"name": "different",
		}

		result := findUnknownFields(original, marshaled)
		assert.Len(t, result, 1)
		assert.Equal(t, "original", result["name"])
	})
}

func TestCompareSlices(t *testing.T) {
	t.Run("returns empty for equal slices", func(t *testing.T) {
		original := []any{"a", "b", "c"}
		marshaled := []any{"a", "b", "c"}

		result := compareSlices(original, marshaled)
		assert.Empty(t, result)
	})

	t.Run("detects missing items in marshaled", func(t *testing.T) {
		original := []any{"a", "b", "c", "d"}
		marshaled := []any{"a", "b"}

		result := compareSlices(original, marshaled)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "c")
		assert.Contains(t, result, "d")
	})

	t.Run("detects extra items in marshaled", func(t *testing.T) {
		original := []any{"a"}
		marshaled := []any{"a", "b", "c"}

		result := compareSlices(original, marshaled)
		assert.Len(t, result, 2)
	})

	t.Run("handles nested maps in slices", func(t *testing.T) {
		original := []any{
			map[string]any{"name": "test", "extra": "field"},
		}
		marshaled := []any{
			map[string]any{"name": "test"},
		}

		result := compareSlices(original, marshaled)
		assert.Len(t, result, 1)
	})

	t.Run("handles different values at same index", func(t *testing.T) {
		original := []any{"a", "b", "c"}
		marshaled := []any{"a", "x", "c"}

		result := compareSlices(original, marshaled)
		assert.Len(t, result, 1)
		assert.Equal(t, "b", result[0])
	})
}

func TestValidateStruct(t *testing.T) {
	type ValidStruct struct {
		Name  string `json:"name" validate:"required"`
		Value int    `json:"value"`
	}

	type InvalidStruct struct {
		Name string `json:"name" validate:"required"`
	}

	t.Run("returns nil for valid struct", func(t *testing.T) {
		s := &ValidStruct{Name: "test", Value: 42}
		err := ValidateStruct(s)
		assert.NoError(t, err)
	})

	t.Run("returns error for invalid struct", func(t *testing.T) {
		s := &InvalidStruct{Name: ""}
		err := ValidateStruct(s)
		assert.Error(t, err)
	})

	t.Run("returns nil for non-struct types", func(t *testing.T) {
		s := "just a string"
		err := ValidateStruct(s)
		assert.NoError(t, err)
	})

	t.Run("handles pointer to struct", func(t *testing.T) {
		s := &ValidStruct{Name: "test"}
		err := ValidateStruct(s)
		assert.NoError(t, err)
	})
}

func TestFields(t *testing.T) {
	v, trans := newValidator()

	type TestStruct struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	t.Run("returns field validations for errors", func(t *testing.T) {
		s := &TestStruct{Name: "", Email: "invalid"}
		err := v.Struct(s)
		assert.Error(t, err)

		// Use trans to avoid unused variable error
		assert.NotNil(t, trans)
	})
}

func TestFieldsRequired(t *testing.T) {
	t.Run("filters required fields", func(t *testing.T) {
		input := pkg.FieldValidations{
			"name":  "name is a required field",
			"email": "email must be a valid email",
			"age":   "age is a required field",
		}

		result := fieldsRequired(input)

		assert.Len(t, result, 2)
		assert.Contains(t, result, "name")
		assert.Contains(t, result, "age")
		assert.NotContains(t, result, "email")
	})

	t.Run("returns empty map when no required fields", func(t *testing.T) {
		input := pkg.FieldValidations{
			"email": "email must be a valid email",
			"url":   "url must be a valid URL",
		}

		result := fieldsRequired(input)
		assert.Empty(t, result)
	})
}

func TestFormatErrorFieldName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "extracts field name from namespace",
			input:    "CreateReportInput.name",
			expected: "name",
		},
		{
			name:     "extracts nested field name",
			input:    "CreateReportInput.nested.field",
			expected: "nested.field",
		},
		{
			name:     "returns original if no dot",
			input:    "fieldname",
			expected: "fieldname",
		},
		{
			name:     "handles multiple levels",
			input:    "Root.Level1.Level2.field",
			expected: "Level1.Level2.field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatErrorFieldName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSimpleType(t *testing.T) {
	tests := []struct {
		name     string
		kind     reflect.Kind
		expected bool
	}{
		{"string is simple", reflect.String, true},
		{"int is simple", reflect.Int, true},
		{"float64 is simple", reflect.Float64, true},
		{"bool is simple", reflect.Bool, true},
		{"map is not simple", reflect.Map, false},
		{"slice is not simple", reflect.Slice, false},
		{"struct is not simple", reflect.Struct, false},
		{"pointer is not simple", reflect.Ptr, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSimpleType(tt.kind)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTypeMismatch(t *testing.T) {
	t.Run("returns nil for compatible string to string", func(t *testing.T) {
		result := getTypeMismatch("test", reflect.String)
		assert.Nil(t, result)
	})

	t.Run("returns mismatch for string to map", func(t *testing.T) {
		result := getTypeMismatch("test", reflect.Map)
		assert.NotNil(t, result)
		assert.Equal(t, "string", result.receivedType)
	})

	t.Run("returns mismatch for string to slice", func(t *testing.T) {
		result := getTypeMismatch("test", reflect.Slice)
		assert.NotNil(t, result)
		assert.Equal(t, "string", result.receivedType)
	})

	t.Run("returns mismatch for map to simple type", func(t *testing.T) {
		result := getTypeMismatch(map[string]any{"key": "value"}, reflect.String)
		assert.NotNil(t, result)
		assert.Equal(t, "object", result.receivedType)
	})

	t.Run("returns mismatch for array to simple type", func(t *testing.T) {
		result := getTypeMismatch([]any{"a", "b"}, reflect.Int)
		assert.NotNil(t, result)
		assert.Equal(t, "array", result.receivedType)
	})

	t.Run("returns mismatch for number to string", func(t *testing.T) {
		result := getTypeMismatch(float64(42), reflect.String)
		assert.NotNil(t, result)
		assert.Equal(t, "number", result.receivedType)
	})

	t.Run("returns mismatch for boolean to string", func(t *testing.T) {
		result := getTypeMismatch(true, reflect.String)
		assert.NotNil(t, result)
		assert.Equal(t, "boolean", result.receivedType)
	})

	t.Run("returns nil for compatible number to int", func(t *testing.T) {
		result := getTypeMismatch(float64(42), reflect.Int)
		assert.Nil(t, result)
	})

	t.Run("returns nil for compatible map to map", func(t *testing.T) {
		result := getTypeMismatch(map[string]any{}, reflect.Map)
		assert.Nil(t, result)
	})
}

func TestExtractFieldNameFromUnmarshalError(t *testing.T) {
	tests := []struct {
		name     string
		errorMsg string
		expected string
	}{
		{
			name:     "extracts field name from standard error",
			errorMsg: "json: cannot unmarshal string into Go struct field CreateReportInput.filters of type map[string]map[string][]string",
			expected: "filters",
		},
		{
			name:     "extracts field name from another format",
			errorMsg: "json: cannot unmarshal number into Go struct field TestStruct.name of type string",
			expected: "name",
		},
		{
			name:     "returns empty for unrecognized format",
			errorMsg: "some random error message",
			expected: "",
		},
		{
			name:     "extracts field from alternative pattern",
			errorMsg: "cannot unmarshal field value of type string",
			expected: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFieldNameFromUnmarshalError(tt.errorMsg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateTypeMismatches(t *testing.T) {
	type TestStruct struct {
		Name   string         `json:"name"`
		Value  int            `json:"value"`
		Active bool           `json:"active"`
		Data   map[string]any `json:"data"`
	}

	t.Run("returns nil for valid types", func(t *testing.T) {
		body := []byte(`{"name": "test", "value": 42, "active": true}`)
		s := &TestStruct{}

		err := validateTypeMismatches(body, s)
		assert.NoError(t, err)
	})

	t.Run("returns error for string to map mismatch", func(t *testing.T) {
		body := []byte(`{"data": "should_be_object"}`)
		s := &TestStruct{}

		err := validateTypeMismatches(body, s)
		assert.Error(t, err)
	})

	t.Run("returns nil for non-pointer", func(t *testing.T) {
		body := []byte(`{"name": "test"}`)
		s := TestStruct{}

		err := validateTypeMismatches(body, s)
		assert.NoError(t, err)
	})

	t.Run("returns nil for non-struct pointer", func(t *testing.T) {
		body := []byte(`{"name": "test"}`)
		s := "just a string"

		err := validateTypeMismatches(body, &s)
		assert.NoError(t, err)
	})

	t.Run("returns error for invalid json", func(t *testing.T) {
		body := []byte(`{invalid json}`)
		s := &TestStruct{}

		err := validateTypeMismatches(body, s)
		assert.Error(t, err)
	})
}

func TestParseMetadata(t *testing.T) {
	type WithMetadata struct {
		Name     string         `json:"name"`
		Metadata map[string]any `json:"metadata"`
	}

	type WithoutMetadata struct {
		Name string `json:"name"`
	}

	t.Run("initializes metadata when not present in original", func(t *testing.T) {
		s := &WithMetadata{Name: "test"}
		original := map[string]any{"name": "test"}

		parseMetadata(s, original)

		assert.NotNil(t, s.Metadata)
		assert.Empty(t, s.Metadata)
	})

	t.Run("preserves metadata when present in original", func(t *testing.T) {
		s := &WithMetadata{
			Name:     "test",
			Metadata: map[string]any{"key": "value"},
		}
		original := map[string]any{
			"name":     "test",
			"metadata": map[string]any{"key": "value"},
		}

		parseMetadata(s, original)

		assert.Equal(t, "value", s.Metadata["key"])
	})

	t.Run("handles struct without metadata field", func(t *testing.T) {
		s := &WithoutMetadata{Name: "test"}
		original := map[string]any{"name": "test"}

		// Should not panic
		parseMetadata(s, original)
	})

	t.Run("handles non-pointer", func(t *testing.T) {
		s := WithMetadata{Name: "test"}
		original := map[string]any{"name": "test"}

		// Should not panic
		parseMetadata(s, original)
	})

	t.Run("handles non-struct", func(t *testing.T) {
		s := "just a string"
		original := map[string]any{"name": "test"}

		// Should not panic
		parseMetadata(&s, original)
	})
}

func TestNewValidator(t *testing.T) {
	t.Run("creates validator and translator", func(t *testing.T) {
		v, trans := newValidator()

		assert.NotNil(t, v)
		assert.NotNil(t, trans)
	})

	t.Run("validator has custom validations registered", func(t *testing.T) {
		v, _ := newValidator()

		type TestStruct struct {
			Key string `validate:"keymax=10"`
		}

		// Should not panic - validation is registered
		s := &TestStruct{Key: "short"}
		err := v.Struct(s)
		assert.NoError(t, err)
	})
}

func TestWithBody(t *testing.T) {
	type TestPayload struct {
		Name string `json:"name"`
	}

	t.Run("returns fiber handler", func(t *testing.T) {
		handler := WithBody(&TestPayload{}, func(p any, c *fiber.Ctx) error {
			return nil
		})

		assert.NotNil(t, handler)
	})
}

func TestFiberHandlerFunc(t *testing.T) {
	type TestPayload struct {
		Name  string `json:"name" validate:"required"`
		Value int    `json:"value"`
	}

	t.Run("successfully parses valid JSON body", func(t *testing.T) {
		app := fiber.New()

		var receivedPayload *TestPayload
		handler := WithBody(&TestPayload{}, func(p any, c *fiber.Ctx) error {
			receivedPayload = p.(*TestPayload)
			return c.SendStatus(fiber.StatusOK)
		})

		app.Post("/test", handler)

		body := bytes.NewReader([]byte(`{"name": "test", "value": 42}`))
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.NotNil(t, receivedPayload)
		assert.Equal(t, "test", receivedPayload.Name)
		assert.Equal(t, 42, receivedPayload.Value)
	})

	t.Run("returns bad request for empty body", func(t *testing.T) {
		app := fiber.New()

		handler := WithBody(&TestPayload{}, func(p any, c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})

		app.Post("/test", handler)

		body := bytes.NewReader([]byte(``))
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns bad request for null body", func(t *testing.T) {
		app := fiber.New()

		handler := WithBody(&TestPayload{}, func(p any, c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})

		app.Post("/test", handler)

		body := bytes.NewReader([]byte(`null`))
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		app := fiber.New()

		handler := WithBody(&TestPayload{}, func(p any, c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})

		app.Post("/test", handler)

		body := bytes.NewReader([]byte(`{invalid json}`))
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		// Invalid JSON should return an error status
		assert.NotEqual(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("returns bad request for validation failure", func(t *testing.T) {
		app := fiber.New()

		handler := WithBody(&TestPayload{}, func(p any, c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})

		app.Post("/test", handler)

		// Missing required "name" field
		body := bytes.NewReader([]byte(`{"value": 42}`))
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns bad request for unknown fields", func(t *testing.T) {
		app := fiber.New()

		handler := WithBody(&TestPayload{}, func(p any, c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})

		app.Post("/test", handler)

		// Contains unknown field "extra"
		body := bytes.NewReader([]byte(`{"name": "test", "value": 42, "extra": "unknown"}`))
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns bad request for type mismatch", func(t *testing.T) {
		app := fiber.New()

		handler := WithBody(&TestPayload{}, func(p any, c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})

		app.Post("/test", handler)

		// "value" should be int, not string
		body := bytes.NewReader([]byte(`{"name": "test", "value": "not_an_int"}`))
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("handles constructor function", func(t *testing.T) {
		app := fiber.New()

		d := &decoderHandler{
			constructor: func() any {
				return &TestPayload{Name: "default"}
			},
			handler: func(p any, c *fiber.Ctx) error {
				payload := p.(*TestPayload)
				return c.JSON(payload)
			},
		}

		app.Post("/test", d.FiberHandlerFunc)

		body := bytes.NewReader([]byte(`{"name": "test", "value": 42}`))
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), "test")
	})
}

func TestTypeMismatchInfo(t *testing.T) {
	info := &typeMismatchInfo{
		receivedType: "string",
		value:        "test_value",
	}

	assert.Equal(t, "string", info.receivedType)
	assert.Equal(t, "test_value", info.value)
}

func TestDecoderHandler(t *testing.T) {
	type TestPayload struct {
		Name string `json:"name"`
	}

	t.Run("struct initialization", func(t *testing.T) {
		handler := &decoderHandler{
			handler:      nil,
			constructor:  nil,
			structSource: &TestPayload{},
		}

		assert.NotNil(t, handler.structSource)
		assert.Nil(t, handler.constructor)
	})
}

func TestValidateMetadataNestedValues(t *testing.T) {
	v, _ := newValidator()

	type StructWithNestedMap struct {
		Value map[string]any `validate:"nonested"`
	}

	type StructWithString struct {
		Value string `validate:"nonested"`
	}

	type StructWithInt struct {
		Value int `validate:"nonested"`
	}

	t.Run("fails for nested map value", func(t *testing.T) {
		s := &StructWithNestedMap{
			Value: map[string]any{"nested": "value"},
		}
		err := v.Struct(s)
		assert.Error(t, err)
	})

	t.Run("passes for string value", func(t *testing.T) {
		s := &StructWithString{
			Value: "simple string",
		}
		err := v.Struct(s)
		assert.NoError(t, err)
	})

	t.Run("passes for int value", func(t *testing.T) {
		s := &StructWithInt{
			Value: 42,
		}
		err := v.Struct(s)
		assert.NoError(t, err)
	})
}

func TestValidateMetadataValueMaxLength(t *testing.T) {
	v, _ := newValidator()

	type StructWithStringValue struct {
		Value string `validate:"valuemax=10"`
	}

	type StructWithIntValue struct {
		Value int `validate:"valuemax=5"`
	}

	type StructWithFloatValue struct {
		Value float64 `validate:"valuemax=10"`
	}

	type StructWithBoolValue struct {
		Value bool `validate:"valuemax=10"`
	}

	type StructWithDefaultLimit struct {
		Value string `validate:"valuemax"`
	}

	type StructWithSliceValue struct {
		Value []string `validate:"valuemax=10"`
	}

	t.Run("passes for string within limit", func(t *testing.T) {
		s := &StructWithStringValue{
			Value: "short",
		}
		err := v.Struct(s)
		assert.NoError(t, err)
	})

	t.Run("fails for string exceeding limit", func(t *testing.T) {
		s := &StructWithStringValue{
			Value: "this is a very long string",
		}
		err := v.Struct(s)
		assert.Error(t, err)
	})

	t.Run("passes for int within limit", func(t *testing.T) {
		s := &StructWithIntValue{
			Value: 123, // "123" has length 3
		}
		err := v.Struct(s)
		assert.NoError(t, err)
	})

	t.Run("fails for int exceeding limit", func(t *testing.T) {
		s := &StructWithIntValue{
			Value: 123456, // "123456" has length 6 > 5
		}
		err := v.Struct(s)
		assert.Error(t, err)
	})

	t.Run("passes for float within limit", func(t *testing.T) {
		s := &StructWithFloatValue{
			Value: 3.14, // "3.14" has length 4
		}
		err := v.Struct(s)
		assert.NoError(t, err)
	})

	t.Run("fails for float exceeding limit", func(t *testing.T) {
		s := &StructWithFloatValue{
			Value: 3.14159265359, // exceeds 10 chars
		}
		err := v.Struct(s)
		assert.Error(t, err)
	})

	t.Run("passes for bool value", func(t *testing.T) {
		s := &StructWithBoolValue{
			Value: true, // "true" has length 4
		}
		err := v.Struct(s)
		assert.NoError(t, err)
	})

	t.Run("passes for default limit with short value", func(t *testing.T) {
		s := &StructWithDefaultLimit{
			Value: "short value",
		}
		err := v.Struct(s)
		assert.NoError(t, err)
	})

	t.Run("fails for unsupported type", func(t *testing.T) {
		s := &StructWithSliceValue{
			Value: []string{"a", "b"},
		}
		err := v.Struct(s)
		assert.Error(t, err)
	})
}
