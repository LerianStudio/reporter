package postgres

import (
	"encoding/json"
	"testing"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/LerianStudio/reporter/v4/pkg/model"
	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
)

// mockLogger implements log.Logger for testing
type mockLogger struct{}

func (m *mockLogger) Info(args ...any)                               {}
func (m *mockLogger) Infof(format string, args ...any)               {}
func (m *mockLogger) Infoln(args ...any)                             {}
func (m *mockLogger) Error(args ...any)                              {}
func (m *mockLogger) Errorf(format string, args ...any)              {}
func (m *mockLogger) Errorln(args ...any)                            {}
func (m *mockLogger) Warn(args ...any)                               {}
func (m *mockLogger) Warnf(format string, args ...any)               {}
func (m *mockLogger) Warnln(args ...any)                             {}
func (m *mockLogger) Debug(args ...any)                              {}
func (m *mockLogger) Debugf(format string, args ...any)              {}
func (m *mockLogger) Debugln(args ...any)                            {}
func (m *mockLogger) Fatal(args ...any)                              {}
func (m *mockLogger) Fatalf(format string, args ...any)              {}
func (m *mockLogger) Fatalln(args ...any)                            {}
func (m *mockLogger) WithFields(fields ...any) log.Logger            { return m }
func (m *mockLogger) WithDefaultMessageTemplate(s string) log.Logger { return m }
func (m *mockLogger) Sync() error                                    { return nil }

func newMockLogger() log.Logger {
	return &mockLogger{}
}

// Tests for ValidateFieldsInSchemaPostgres
func TestValidateFieldsInSchemaPostgres_AllFieldsExist(t *testing.T) {
	schema := TableSchema{
		TableName: "users",
		Columns: []ColumnInformation{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "varchar"},
			{Name: "email", DataType: "varchar"},
		},
	}
	expectedFields := []string{"id", "name", "email"}
	var count int32

	missing := ValidateFieldsInSchemaPostgres(expectedFields, schema, &count)

	assert.Empty(t, missing)
	assert.Equal(t, int32(3), count)
}

func TestValidateFieldsInSchemaPostgres_SomeFieldsMissing(t *testing.T) {
	schema := TableSchema{
		TableName: "users",
		Columns: []ColumnInformation{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "varchar"},
		},
	}
	expectedFields := []string{"id", "name", "email", "age"}
	var count int32

	missing := ValidateFieldsInSchemaPostgres(expectedFields, schema, &count)

	assert.Len(t, missing, 2)
	assert.Contains(t, missing, "email")
	assert.Contains(t, missing, "age")
	assert.Equal(t, int32(4), count)
}

func TestValidateFieldsInSchemaPostgres_CaseInsensitive(t *testing.T) {
	schema := TableSchema{
		TableName: "users",
		Columns: []ColumnInformation{
			{Name: "ID", DataType: "uuid"},
			{Name: "Name", DataType: "varchar"},
		},
	}
	expectedFields := []string{"id", "name"}
	var count int32

	missing := ValidateFieldsInSchemaPostgres(expectedFields, schema, &count)

	assert.Empty(t, missing)
}

func TestValidateFieldsInSchemaPostgres_EmptySchema(t *testing.T) {
	schema := TableSchema{
		TableName: "empty_table",
		Columns:   []ColumnInformation{},
	}
	expectedFields := []string{"id", "name"}
	var count int32

	missing := ValidateFieldsInSchemaPostgres(expectedFields, schema, &count)

	assert.Len(t, missing, 2)
}

// Tests for TableSchema JSON serialization
func TestTableSchema_JSONSerialization(t *testing.T) {
	schema := TableSchema{
		TableName: "users",
		Columns: []ColumnInformation{
			{Name: "id", DataType: "uuid", IsNullable: false, IsPrimaryKey: true},
			{Name: "email", DataType: "varchar", IsNullable: true, IsPrimaryKey: false},
		},
	}

	data, err := json.Marshal(schema)
	assert.NoError(t, err)

	var result TableSchema
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, "users", result.TableName)
	assert.Len(t, result.Columns, 2)
	assert.Equal(t, "id", result.Columns[0].Name)
	assert.True(t, result.Columns[0].IsPrimaryKey)
}

// Tests for ColumnInformation JSON serialization
func TestColumnInformation_JSONSerialization(t *testing.T) {
	col := ColumnInformation{
		Name:         "user_id",
		DataType:     "uuid",
		IsNullable:   false,
		IsPrimaryKey: true,
	}

	data, err := json.Marshal(col)
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, "user_id", result["name"])
	assert.Equal(t, "uuid", result["data_type"])
	assert.Equal(t, false, result["is_nullable"])
	assert.Equal(t, true, result["is_primary_key"])
}

// Tests for buildDynamicFilters
func TestBuildDynamicFilters_WithValidFilters(t *testing.T) {
	schema := []TableSchema{
		{
			TableName: "users",
			Columns: []ColumnInformation{
				{Name: "id", DataType: "uuid"},
				{Name: "status", DataType: "varchar"},
			},
		},
	}
	filter := map[string][]any{
		"status": {"active", "pending"},
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select("*").From("users")

	result := buildDynamicFilters(queryBuilder, schema, "users", filter)

	sql, args, err := result.ToSql()
	assert.NoError(t, err)
	assert.Contains(t, sql, "WHERE")
	assert.Contains(t, sql, "status IN")
	assert.Len(t, args, 2)
}

func TestBuildDynamicFilters_WithInvalidColumn(t *testing.T) {
	schema := []TableSchema{
		{
			TableName: "users",
			Columns: []ColumnInformation{
				{Name: "id", DataType: "uuid"},
			},
		},
	}
	filter := map[string][]any{
		"nonexistent": {"value"},
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select("*").From("users")

	result := buildDynamicFilters(queryBuilder, schema, "users", filter)

	sql, _, err := result.ToSql()
	assert.NoError(t, err)
	assert.NotContains(t, sql, "WHERE") // No filter applied
}

func TestBuildDynamicFilters_EmptyFilter(t *testing.T) {
	schema := []TableSchema{
		{
			TableName: "users",
			Columns: []ColumnInformation{
				{Name: "id", DataType: "uuid"},
			},
		},
	}
	filter := map[string][]any{}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select("*").From("users")

	result := buildDynamicFilters(queryBuilder, schema, "users", filter)

	sql, _, err := result.ToSql()
	assert.NoError(t, err)
	assert.NotContains(t, sql, "WHERE")
}

// Tests for applyFilter
func TestApplyFilter_SingleValue(t *testing.T) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select("*").From("users")

	result := applyFilter(queryBuilder, "status", []any{"active"})

	sql, args, err := result.ToSql()
	assert.NoError(t, err)
	assert.Contains(t, sql, "status IN")
	assert.Len(t, args, 1)
	assert.Equal(t, "active", args[0])
}

func TestApplyFilter_MultipleValues(t *testing.T) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select("*").From("users")

	result := applyFilter(queryBuilder, "status", []any{"active", "pending", "completed"})

	sql, args, err := result.ToSql()
	assert.NoError(t, err)
	assert.Contains(t, sql, "status IN")
	assert.Len(t, args, 3)
}

func TestApplyFilter_EmptyValues(t *testing.T) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select("*").From("users")

	result := applyFilter(queryBuilder, "status", []any{})

	sql, _, err := result.ToSql()
	assert.NoError(t, err)
	assert.NotContains(t, sql, "WHERE")
}

// Tests for isFilterConditionEmpty
func TestIsFilterConditionEmpty_Empty(t *testing.T) {
	condition := model.FilterCondition{}
	assert.True(t, isFilterConditionEmpty(condition))
}

func TestIsFilterConditionEmpty_WithEquals(t *testing.T) {
	condition := model.FilterCondition{
		Equals: []any{"value"},
	}
	assert.False(t, isFilterConditionEmpty(condition))
}

func TestIsFilterConditionEmpty_WithGreaterThan(t *testing.T) {
	condition := model.FilterCondition{
		GreaterThan: []any{100},
	}
	assert.False(t, isFilterConditionEmpty(condition))
}

func TestIsFilterConditionEmpty_WithBetween(t *testing.T) {
	condition := model.FilterCondition{
		Between: []any{100, 200},
	}
	assert.False(t, isFilterConditionEmpty(condition))
}

func TestIsFilterConditionEmpty_WithIn(t *testing.T) {
	condition := model.FilterCondition{
		In: []any{"a", "b", "c"},
	}
	assert.False(t, isFilterConditionEmpty(condition))
}

func TestIsFilterConditionEmpty_WithNotIn(t *testing.T) {
	condition := model.FilterCondition{
		NotIn: []any{"deleted"},
	}
	assert.False(t, isFilterConditionEmpty(condition))
}

// Tests for validateFilterCondition
func TestValidateFilterCondition_ValidBetween(t *testing.T) {
	condition := model.FilterCondition{
		Between: []any{100, 200},
	}
	err := validateFilterCondition("amount", condition)
	assert.NoError(t, err)
}

func TestValidateFilterCondition_InvalidBetweenSingleValue(t *testing.T) {
	condition := model.FilterCondition{
		Between: []any{100},
	}
	err := validateFilterCondition("amount", condition)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have exactly 2 values")
}

func TestValidateFilterCondition_InvalidBetweenThreeValues(t *testing.T) {
	condition := model.FilterCondition{
		Between: []any{100, 200, 300},
	}
	err := validateFilterCondition("amount", condition)
	assert.Error(t, err)
}

func TestValidateFilterCondition_ValidSingleValueOperators(t *testing.T) {
	tests := []struct {
		name      string
		condition model.FilterCondition
	}{
		{
			name: "valid gt",
			condition: model.FilterCondition{
				GreaterThan: []any{100},
			},
		},
		{
			name: "valid gte",
			condition: model.FilterCondition{
				GreaterOrEqual: []any{100},
			},
		},
		{
			name: "valid lt",
			condition: model.FilterCondition{
				LessThan: []any{100},
			},
		},
		{
			name: "valid lte",
			condition: model.FilterCondition{
				LessOrEqual: []any{100},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFilterCondition("amount", tt.condition)
			assert.NoError(t, err)
		})
	}
}

func TestValidateFilterCondition_InvalidGtMultipleValues(t *testing.T) {
	condition := model.FilterCondition{
		GreaterThan: []any{100, 200},
	}
	err := validateFilterCondition("amount", condition)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have exactly 1 value")
}

func TestValidateFilterCondition_UUIDFieldValidation(t *testing.T) {
	condition := model.FilterCondition{
		Equals: []any{"not-a-uuid"},
	}
	err := validateFilterCondition("user_id", condition)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UUID")
}

func TestValidateFilterCondition_ValidUUIDField(t *testing.T) {
	condition := model.FilterCondition{
		Equals: []any{"550e8400-e29b-41d4-a716-446655440000"},
	}
	err := validateFilterCondition("user_id", condition)
	assert.NoError(t, err)
}

// Tests for isLikelyUUIDField
func TestIsLikelyUUIDField(t *testing.T) {
	tests := []struct {
		fieldName string
		expected  bool
	}{
		{"id", true},
		{"user_id", true},
		{"account_id", true},
		{"template_id", true},
		{"organization_id", true},
		{"uuid", true},
		{"name", false},
		{"email", false},
		{"status", false},
		{"created_at", false},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := isLikelyUUIDField(tt.fieldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for isValidUUIDFormat
func TestIsValidUUIDFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"00000000-0000-0000-0000-000000000000", true},
		{"not-a-uuid", false},
		{"550e8400e29b41d4a716446655440000", true}, // Without dashes is also valid for uuid.Parse
		{"", false},
		{"550e8400-e29b-41d4-a716", false}, // Incomplete
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isValidUUIDFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for isDateField
func TestIsDateField(t *testing.T) {
	tests := []struct {
		fieldName string
		expected  bool
	}{
		{"created_at", true},
		{"updated_at", true},
		{"deleted_at", true},
		{"completed_at", true},
		{"birth_date", true},
		{"start_date", true},
		{"end_time", true},
		{"name", false},
		{"email", false},
		{"status", false},
		{"user_id", false},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := isDateField(tt.fieldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for isDateString
func TestIsDateString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"date only", "2025-01-15", true},
		{"datetime with T", "2025-01-15T10:30:00", true},
		{"datetime with timezone", "2025-01-15T10:30:00Z", true},
		{"too short", "2025-01", false},
		{"no dash", "20250115", false},
		{"number", 123, false},
		{"nil", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDateString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for parseJSONBField
func TestParseJSONBField_Nil(t *testing.T) {
	result := parseJSONBField(nil, newMockLogger())
	assert.Nil(t, result)
}

func TestParseJSONBField_Map(t *testing.T) {
	jsonData := []uint8(`{"key": "value", "num": 123}`)
	result := parseJSONBField(jsonData, newMockLogger())

	mapResult, ok := result.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "value", mapResult["key"])
	assert.Equal(t, float64(123), mapResult["num"])
}

func TestParseJSONBField_Array(t *testing.T) {
	jsonData := []uint8(`["a", "b", "c"]`)
	result := parseJSONBField(jsonData, newMockLogger())

	arrayResult, ok := result.([]any)
	assert.True(t, ok)
	assert.Len(t, arrayResult, 3)
	assert.Equal(t, "a", arrayResult[0])
}

func TestParseJSONBField_String(t *testing.T) {
	jsonData := []uint8(`"hello world"`)
	result := parseJSONBField(jsonData, newMockLogger())

	stringResult, ok := result.(string)
	assert.True(t, ok)
	assert.Equal(t, "hello world", stringResult)
}

func TestParseJSONBField_InvalidJSON(t *testing.T) {
	// Invalid JSON should return original value
	jsonData := []uint8(`not valid json`)
	result := parseJSONBField(jsonData, newMockLogger())

	// Should return original []uint8 when parsing fails
	_, ok := result.([]uint8)
	assert.True(t, ok)
}

func TestParseJSONBField_NonByteValue(t *testing.T) {
	// Non-byte values should be returned as-is
	result := parseJSONBField(123, newMockLogger())
	assert.Equal(t, 123, result)

	result = parseJSONBField("string", newMockLogger())
	assert.Equal(t, "string", result)
}

// Tests for createRowMap
func TestCreateRowMap(t *testing.T) {
	columns := []string{"id", "name", "data"}
	values := []any{
		"123",
		"John",
		[]uint8(`{"key": "value"}`),
	}

	result := createRowMap(columns, values, newMockLogger())

	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "John", result["name"])
	// JSONB field should be parsed
	dataMap, ok := result["data"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "value", dataMap["key"])
}

// Tests for validateUUIDFieldValues
func TestValidateUUIDFieldValues_ValidUUIDs(t *testing.T) {
	condition := model.FilterCondition{
		Equals: []any{"550e8400-e29b-41d4-a716-446655440000"},
		In:     []any{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"},
	}
	err := validateUUIDFieldValues("user_id", condition)
	assert.NoError(t, err)
}

func TestValidateUUIDFieldValues_InvalidUUID(t *testing.T) {
	condition := model.FilterCondition{
		Equals: []any{"not-a-valid-uuid"},
	}
	err := validateUUIDFieldValues("user_id", condition)
	assert.Error(t, err)
}

func TestValidateUUIDFieldValues_MixedTypes(t *testing.T) {
	// Non-string values should be skipped
	condition := model.FilterCondition{
		Equals: []any{123, 456},
	}
	err := validateUUIDFieldValues("user_id", condition)
	assert.NoError(t, err) // Non-strings are not validated as UUIDs
}

// Tests for Connection struct
func TestConnection_GetDB_NotConnected(t *testing.T) {
	conn := &Connection{
		Logger: newMockLogger(),
	}

	// Without a valid connection string, this should fail
	_, err := conn.GetDB()
	assert.Error(t, err)
}
