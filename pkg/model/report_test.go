package model

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFilterCondition_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		filter   FilterCondition
		expected map[string]any
	}{
		{
			name: "equals filter",
			filter: FilterCondition{
				Equals: []any{"active", "pending"},
			},
			expected: map[string]any{
				"eq": []any{"active", "pending"},
			},
		},
		{
			name: "greater than filter",
			filter: FilterCondition{
				GreaterThan: []any{100},
			},
			expected: map[string]any{
				"gt": []any{float64(100)}, // JSON numbers are float64
			},
		},
		{
			name: "between filter",
			filter: FilterCondition{
				Between: []any{100, 1000},
			},
			expected: map[string]any{
				"between": []any{float64(100), float64(1000)},
			},
		},
		{
			name: "in filter",
			filter: FilterCondition{
				In: []any{"active", "pending", "suspended"},
			},
			expected: map[string]any{
				"in": []any{"active", "pending", "suspended"},
			},
		},
		{
			name: "not in filter",
			filter: FilterCondition{
				NotIn: []any{"deleted", "archived"},
			},
			expected: map[string]any{
				"nin": []any{"deleted", "archived"},
			},
		},
		{
			name: "combined filters",
			filter: FilterCondition{
				GreaterOrEqual: []any{100},
				LessOrEqual:    []any{1000},
			},
			expected: map[string]any{
				"gte": []any{float64(100)},
				"lte": []any{float64(1000)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.filter)
			assert.NoError(t, err)

			var result map[string]any
			err = json.Unmarshal(data, &result)
			assert.NoError(t, err)

			for key := range tt.expected {
				assert.Contains(t, result, key)
			}
		})
	}
}

func TestFilterCondition_OmitEmpty(t *testing.T) {
	// Empty filter should produce empty JSON object
	filter := FilterCondition{}
	data, err := json.Marshal(filter)
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(data))
}

func TestCreateReportInput_JSONSerialization(t *testing.T) {
	input := CreateReportInput{
		TemplateID: "00000000-0000-0000-0000-000000000001",
		Filters: map[string]map[string]map[string]FilterCondition{
			"datasource1": {
				"entity1": {
					"field1": {
						Equals: []any{"value1"},
					},
				},
			},
		},
	}

	data, err := json.Marshal(input)
	assert.NoError(t, err)

	var result CreateReportInput
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, input.TemplateID, result.TemplateID)
	assert.NotNil(t, result.Filters["datasource1"])
}

func TestReportMessage_JSONSerialization(t *testing.T) {
	templateID := uuid.New()
	reportID := uuid.New()

	msg := ReportMessage{
		TemplateID:   templateID,
		ReportID:     reportID,
		OutputFormat: "html",
		Filters: map[string]map[string]map[string]FilterCondition{
			"datasource": {
				"entity": {
					"field": {
						Equals: []any{"value"},
					},
				},
			},
		},
		MappedFields: map[string]map[string][]string{
			"datasource": {
				"entity": {"field1", "field2"},
			},
		},
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	var result ReportMessage
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, templateID, result.TemplateID)
	assert.Equal(t, reportID, result.ReportID)
	assert.Equal(t, "html", result.OutputFormat)
	assert.NotNil(t, result.Filters)
	assert.NotNil(t, result.MappedFields)
}

func TestReportMessage_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"templateId": "00000000-0000-0000-0000-000000000001",
		"reportId": "00000000-0000-0000-0000-000000000002",
		"outputFormat": "pdf",
		"filters": {
			"ds": {
				"ent": {
					"fld": {
						"eq": ["val1", "val2"]
					}
				}
			}
		},
		"mappedFields": {
			"ds": {
				"ent": ["field1"]
			}
		}
	}`

	var msg ReportMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	assert.NoError(t, err)

	assert.Equal(t, "00000000-0000-0000-0000-000000000001", msg.TemplateID.String())
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", msg.ReportID.String())
	assert.Equal(t, "pdf", msg.OutputFormat)
	assert.Equal(t, []any{"val1", "val2"}, msg.Filters["ds"]["ent"]["fld"].Equals)
	assert.Equal(t, []string{"field1"}, msg.MappedFields["ds"]["ent"])
}
