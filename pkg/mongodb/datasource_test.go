package mongodb

import (
	"testing"

	"plugin-smart-templates/v2/pkg/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDiscoverFieldsRecursively(t *testing.T) {
	ds := &ExternalDataSource{}

	testDoc := bson.M{
		"_id":  primitive.ObjectID{},
		"name": "Test Organization",
		"legal_person": bson.M{
			"trade_name": "Legal Name",
			"document":   "12345678901",
			"representative": bson.M{
				"name":  "Representative Name",
				"email": "rep@example.com",
				"role":  "CEO",
			},
			"address": bson.M{
				"street":  "Main Street",
				"city":    "São Paulo",
				"country": "Brazil",
			},
		},
		"addresses": bson.A{
			bson.M{
				"type":   "primary",
				"street": "Address 1",
				"city":   "City 1",
				"details": bson.M{
					"zip_code": "12345",
					"state":    "SP",
				},
			},
			bson.M{
				"type":   "secondary",
				"street": "Address 2",
				"city":   "City 2",
			},
		},
		"natural_person": bson.M{
			"first_name": "John",
			"last_name":  "Doe",
			"contact": bson.M{
				"email": "john@example.com",
				"phone": "123456789",
				"addresses": bson.A{
					bson.M{
						"type":   "home",
						"street": "Home Street",
					},
				},
			},
		},
		"metadata": bson.M{
			"source":  "test",
			"version": "1.0",
		},
		"tags":       bson.A{"tag1", "tag2", "tag3"},
		"active":     true,
		"count":      42,
		"created_at": primitive.DateTime(1640995200000), // 2022-01-01
	}

	allFields := make(map[string]bool)
	fieldTypes := make(map[string]string)

	ds.discoverFieldsRecursively(testDoc, "", allFields, fieldTypes)
	expectedFields := []string{
		"_id",
		"name",
		"legal_person",
		"legal_person.trade_name",
		"legal_person.document",
		"legal_person.representative",
		"legal_person.representative.name",
		"legal_person.representative.email",
		"legal_person.representative.role",
		"legal_person.address",
		"legal_person.address.street",
		"legal_person.address.city",
		"legal_person.address.country",
		"addresses",
		"addresses.0.type",
		"addresses.0.street",
		"addresses.0.city",
		"addresses.0.details",
		"addresses.0.details.zip_code",
		"addresses.0.details.state",
		"natural_person",
		"natural_person.first_name",
		"natural_person.last_name",
		"natural_person.contact",
		"natural_person.contact.email",
		"natural_person.contact.phone",
		"natural_person.contact.addresses",
		"natural_person.contact.addresses.0.type",
		"natural_person.contact.addresses.0.street",
		"metadata",
		"metadata.source",
		"metadata.version",
		"tags",
		"active",
		"count",
		"created_at",
	}

	for _, expectedField := range expectedFields {
		if !allFields[expectedField] {
			t.Errorf("Expected field '%s' not found in discovered fields", expectedField)
		}
	}

	expectedTypes := map[string]string{
		"_id":          "objectId",
		"name":         "string",
		"active":       "boolean",
		"count":        "number",
		"created_at":   "date",
		"tags":         "array",
		"legal_person": "object",
		"addresses":    "array",
	}

	for field, expectedType := range expectedTypes {
		if actualType, exists := fieldTypes[field]; !exists {
			t.Errorf("Field type not found for '%s'", field)
		} else if actualType != expectedType {
			t.Errorf("Expected type '%s' for field '%s', got '%s'", expectedType, field, actualType)
		}
	}

	t.Logf("Discovered %d fields total", len(allFields))
}

func TestInferDataType(t *testing.T) {
	ds := &ExternalDataSource{}

	testCases := []struct {
		value    any
		expected string
	}{
		{"hello", "string"},
		{42, "number"},
		{3.14, "number"},
		{true, "boolean"},
		{false, "boolean"},
		{bson.A{"a", "b"}, "array"},
		{bson.M{"key": "value"}, "object"},
		{bson.D{primitive.E{Key: "key", Value: "value"}}, "object"},
		{primitive.DateTime(1640995200000), "date"},
		{primitive.ObjectID{}, "objectId"},
		{primitive.Binary{Data: []byte("test")}, "binData"},
		{primitive.Regex{Pattern: "test"}, "regex"},
		{primitive.Timestamp{T: 1640995200}, "timestamp"},
		{primitive.Decimal128{}, "decimal"},
		{primitive.MinKey{}, "minKey/maxKey"},
		{primitive.MaxKey{}, "minKey/maxKey"},
		{nil, "unknown"},
		{[]byte("test"), "unknown"},
	}

	for _, tc := range testCases {
		result := ds.inferDataType(tc.value)
		if result != tc.expected {
			t.Errorf("Expected type '%s' for value %v, got '%s'", tc.expected, tc.value, result)
		}
	}
}

func TestIsMoreSpecificType(t *testing.T) {
	ds := &ExternalDataSource{}

	testCases := []struct {
		newType     string
		currentType string
		expected    bool
	}{
		{"objectId", "string", true},
		{"date", "string", true},
		{"number", "string", true},
		{"string", "objectId", false},
		{"unknown", "string", false},
		{"objectId", "objectId", false},
		{"date", "timestamp", true},
		{"decimal", "number", true},
	}

	for _, tc := range testCases {
		result := ds.isMoreSpecificType(tc.newType, tc.currentType)
		if result != tc.expected {
			t.Errorf("Expected %v for isMoreSpecificType('%s', '%s'), got %v",
				tc.expected, tc.newType, tc.currentType, result)
		}
	}
}

func TestCalculateOptimalSampleSize(t *testing.T) {
	ds := &ExternalDataSource{}

	testCases := []struct {
		totalDocs int64
		expected  int
	}{
		{100, 100},
		{1000, 1000},
		{5000, 1000},
		{10000, 1000},
		{50000, 2000},
		{100000, 2000},
		{500000, 5000},
		{1000000, 5000},
		{5000000, 10000},
	}

	for _, tc := range testCases {
		result := ds.calculateOptimalSampleSize(tc.totalDocs)
		if result != tc.expected {
			t.Errorf("Expected sample size %d for %d docs, got %d",
				tc.expected, tc.totalDocs, result)
		}
	}
}

func TestConvertBsonToMap(t *testing.T) {
	testDoc := bson.M{
		"_id":  primitive.ObjectID{},
		"name": "Test",
		"nested": bson.M{
			"value": 42,
			"array": bson.A{"a", "b", "c"},
			"deep": bson.M{
				"level": "deep",
			},
		},
		"array": bson.A{
			bson.M{"item": "first"},
			bson.M{"item": "second"},
		},
		"date":   primitive.DateTime(1640995200000),
		"binary": primitive.Binary{Data: []byte("test")},
	}

	result := convertBsonToMap(testDoc)

	// Check root level fields
	if result["name"] != "Test" {
		t.Errorf("Expected 'Test' for name, got %v", result["name"])
	}

	if nested, ok := result["nested"].(map[string]any); !ok {
		t.Error("Expected nested to be map[string]any")
	} else {
		if nested["value"] != 42 {
			t.Errorf("Expected 42 for nested.value, got %v", nested["value"])
		}
	}

	if array, ok := result["array"].([]any); !ok {
		t.Error("Expected array to be []any")
	} else if len(array) != 2 {
		t.Errorf("Expected array length 2, got %d", len(array))
	}

	if date, ok := result["date"]; !ok {
		t.Error("Expected date field to exist")
	} else {
		if _, ok := date.(primitive.DateTime); ok {
			t.Error("Expected date to be converted from primitive.DateTime")
		}
	}

	if binary, ok := result["binary"]; !ok {
		t.Error("Expected binary field to exist")
	} else {
		if _, ok := binary.(string); !ok {
			t.Errorf("Expected binary to be converted to string, got %T", binary)
		}
	}
}

func TestConvertBsonValue(t *testing.T) {
	testCases := []struct {
		name     string
		value    any
		expected any
	}{
		{
			name:     "string",
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "number",
			value:    42,
			expected: 42,
		},
		{
			name:     "bson.M",
			value:    bson.M{"key": "value"},
			expected: map[string]any{"key": "value"},
		},
		{
			name:     "bson.A",
			value:    bson.A{"a", "b"},
			expected: []any{"a", "b"},
		},
		{
			name:     "primitive.DateTime",
			value:    primitive.DateTime(1640995200000),
			expected: primitive.DateTime(1640995200000).Time(),
		},
		{
			name:     "primitive.ObjectID",
			value:    primitive.ObjectID{},
			expected: primitive.ObjectID{}.Hex(),
		},
		{
			name:     "primitive.Binary",
			value:    primitive.Binary{Data: []byte("test")},
			expected: "74657374", // hex of "test"
		},
		{
			name:     "nil",
			value:    nil,
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := convertBsonValue(tc.value)

			switch tc.name {
			case "bson.M":
				if _, ok := result.(map[string]any); !ok {
					t.Errorf("Expected map[string]any, got %T", result)
				}
			case "bson.A":
				if _, ok := result.([]any); !ok {
					t.Errorf("Expected []any, got %T", result)
				}
			case "primitive.DateTime":
				if _, ok := result.(primitive.DateTime); ok {
					t.Error("Expected time.Time, got primitive.DateTime")
				}
			case "primitive.ObjectID":
				if _, ok := result.(string); !ok {
					t.Errorf("Expected string, got %T", result)
				}
			case "primitive.Binary":
				if _, ok := result.(string); !ok {
					t.Errorf("Expected string, got %T", result)
				}
			default:
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			}
		})
	}
}

func TestIsFilterConditionEmpty(t *testing.T) {
	testCases := []struct {
		name      string
		condition map[string]any
		expected  bool
	}{
		{
			name:      "empty condition",
			condition: map[string]any{},
			expected:  true,
		},
		{
			name: "condition with equals",
			condition: map[string]any{
				"Equals": []any{"value"},
			},
			expected: false,
		},
		{
			name: "condition with greater than",
			condition: map[string]any{
				"GreaterThan": []any{42},
			},
			expected: false,
		},
		{
			name: "condition with between",
			condition: map[string]any{
				"Between": []any{1, 10},
			},
			expected: false,
		},
		{
			name: "condition with in",
			condition: map[string]any{
				"In": []any{"a", "b", "c"},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			condition := model.FilterCondition{}

			if equals, ok := tc.condition["Equals"].([]any); ok {
				condition.Equals = equals
			}
			if gt, ok := tc.condition["GreaterThan"].([]any); ok {
				condition.GreaterThan = gt
			}
			if gte, ok := tc.condition["GreaterOrEqual"].([]any); ok {
				condition.GreaterOrEqual = gte
			}
			if lt, ok := tc.condition["LessThan"].([]any); ok {
				condition.LessThan = lt
			}
			if lte, ok := tc.condition["LessOrEqual"].([]any); ok {
				condition.LessOrEqual = lte
			}
			if between, ok := tc.condition["Between"].([]any); ok {
				condition.Between = between
			}
			if in, ok := tc.condition["In"].([]any); ok {
				condition.In = in
			}
			if notIn, ok := tc.condition["NotIn"].([]any); ok {
				condition.NotIn = notIn
			}

			result := isFilterConditionEmpty(condition)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestValidateFieldsInSchemaMongo(t *testing.T) {
	schema := CollectionSchema{
		CollectionName: "test_collection",
		Fields: []FieldInformation{
			{Name: "name", DataType: "string"},
			{Name: "age", DataType: "number"},
			{Name: "active", DataType: "boolean"},
			{Name: "legal_person", DataType: "object"},
			{Name: "legal_person.name", DataType: "string"},
			{Name: "addresses", DataType: "array"},
			{Name: "addresses.0.type", DataType: "string"},
		},
	}

	testCases := []struct {
		name            string
		expectedFields  []string
		expectedCount   int32
		expectedMissing []string
	}{
		{
			name:            "all fields exist",
			expectedFields:  []string{"name", "age", "active"},
			expectedCount:   3,
			expectedMissing: []string{},
		},
		{
			name:            "some fields missing",
			expectedFields:  []string{"name", "nonexistent", "age", "missing"},
			expectedCount:   4,
			expectedMissing: []string{"nonexistent", "missing"},
		},
		{
			name:            "nested fields exist",
			expectedFields:  []string{"legal_person", "legal_person.name", "addresses.0.type"},
			expectedCount:   3,
			expectedMissing: []string{},
		},
		{
			name:            "case insensitive",
			expectedFields:  []string{"NAME", "Age", "ACTIVE"},
			expectedCount:   3,
			expectedMissing: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count := int32(0)
			missing := ValidateFieldsInSchemaMongo(tc.expectedFields, schema, &count)

			if count != tc.expectedCount {
				t.Errorf("Expected count %d, got %d", tc.expectedCount, count)
			}

			if len(missing) != len(tc.expectedMissing) {
				t.Errorf("Expected %d missing fields, got %d", len(tc.expectedMissing), len(missing))
			}

			for _, expectedMissing := range tc.expectedMissing {
				found := false
				for _, actualMissing := range missing {
					if actualMissing == expectedMissing {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected missing field '%s' not found in result", expectedMissing)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkDiscoverFieldsRecursively(b *testing.B) {
	ds := &ExternalDataSource{}
	testDoc := bson.M{
		"_id":  "test-id",
		"name": "Test Organization",
		"legal_person": bson.M{
			"trade_name": "Legal Name",
			"representative": bson.M{
				"name": "Representative Name",
				"contact": bson.M{
					"email": "rep@example.com",
					"phone": "123456789",
				},
			},
		},
		"addresses": bson.A{
			bson.M{
				"type": "primary",
				"details": bson.M{
					"street": "Main Street",
					"city":   "São Paulo",
				},
			},
		},
		"metadata": bson.M{
			"source":  "test",
			"version": "1.0",
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		allFields := make(map[string]bool)
		fieldTypes := make(map[string]string)
		ds.discoverFieldsRecursively(testDoc, "", allFields, fieldTypes)
	}
}

func BenchmarkInferDataType(b *testing.B) {
	ds := &ExternalDataSource{}
	testValues := []any{
		"string",
		42,
		true,
		bson.A{"a", "b"},
		bson.M{"key": "value"},
		primitive.DateTime(1640995200000),
		primitive.ObjectID{},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, value := range testValues {
			ds.inferDataType(value)
		}
	}
}

func BenchmarkConvertBsonToMap(b *testing.B) {
	testDoc := bson.M{
		"_id":  primitive.ObjectID{},
		"name": "Test",
		"nested": bson.M{
			"value": 42,
			"array": bson.A{"a", "b", "c"},
		},
		"array": bson.A{
			bson.M{"item": "first"},
			bson.M{"item": "second"},
		},
		"date": primitive.DateTime(1640995200000),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		convertBsonToMap(testDoc)
	}
}
