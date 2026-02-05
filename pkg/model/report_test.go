// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package model

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFilterCondition_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		filter   FilterCondition
		expected string
	}{
		{
			name: "Equals filter",
			filter: FilterCondition{
				Equals: []any{"active", "pending"},
			},
			expected: `{"eq":["active","pending"]}`,
		},
		{
			name: "GreaterThan filter",
			filter: FilterCondition{
				GreaterThan: []any{100},
			},
			expected: `{"gt":[100]}`,
		},
		{
			name: "Between filter",
			filter: FilterCondition{
				Between: []any{100, 1000},
			},
			expected: `{"between":[100,1000]}`,
		},
		{
			name: "In filter",
			filter: FilterCondition{
				In: []any{"active", "pending", "suspended"},
			},
			expected: `{"in":["active","pending","suspended"]}`,
		},
		{
			name: "NotIn filter",
			filter: FilterCondition{
				NotIn: []any{"deleted", "archived"},
			},
			expected: `{"nin":["deleted","archived"]}`,
		},
		{
			name:     "Empty filter",
			filter:   FilterCondition{},
			expected: `{}`,
		},
		{
			name: "Combined filters",
			filter: FilterCondition{
				GreaterOrEqual: []any{"2025-06-01"},
				LessOrEqual:    []any{"2025-06-30"},
			},
			expected: `{"gte":["2025-06-01"],"lte":["2025-06-30"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.filter)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestFilterCondition_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected FilterCondition
	}{
		{
			name: "Equals filter",
			json: `{"eq":["active"]}`,
			expected: FilterCondition{
				Equals: []any{"active"},
			},
		},
		{
			name: "GreaterThan with number",
			json: `{"gt":[100]}`,
			expected: FilterCondition{
				GreaterThan: []any{float64(100)}, // JSON numbers unmarshal as float64
			},
		},
		{
			name: "Between filter",
			json: `{"between":[100,500]}`,
			expected: FilterCondition{
				Between: []any{float64(100), float64(500)},
			},
		},
		{
			name: "Complex filter",
			json: `{"gte":["2025-01-01"],"lte":["2025-12-31"],"in":["active","pending"]}`,
			expected: FilterCondition{
				GreaterOrEqual: []any{"2025-01-01"},
				LessOrEqual:    []any{"2025-12-31"},
				In:             []any{"active", "pending"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result FilterCondition
			err := json.Unmarshal([]byte(tt.json), &result)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateReportInput_JSONMarshal(t *testing.T) {
	input := CreateReportInput{
		TemplateID: "00000000-0000-0000-0000-000000000001",
		Filters: map[string]map[string]map[string]FilterCondition{
			"database": {
				"table": {
					"status": {
						Equals: []any{"active"},
					},
				},
			},
		},
	}

	data, err := json.Marshal(input)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, "00000000-0000-0000-0000-000000000001", result["templateId"])
	assert.NotNil(t, result["filters"])
}

func TestCreateReportInput_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"templateId": "00000000-0000-0000-0000-000000000001",
		"filters": {
			"midaz_onboarding": {
				"account": {
					"status": {"eq": ["active"]}
				}
			}
		}
	}`

	var input CreateReportInput
	err := json.Unmarshal([]byte(jsonData), &input)
	assert.NoError(t, err)

	assert.Equal(t, "00000000-0000-0000-0000-000000000001", input.TemplateID)
	assert.NotNil(t, input.Filters)
	assert.NotNil(t, input.Filters["midaz_onboarding"])
	assert.NotNil(t, input.Filters["midaz_onboarding"]["account"])
	assert.NotNil(t, input.Filters["midaz_onboarding"]["account"]["status"])
}

func TestReportMessage_JSONMarshal(t *testing.T) {
	templateID := uuid.New()
	reportID := uuid.New()

	msg := ReportMessage{
		TemplateID:   templateID,
		ReportID:     reportID,
		OutputFormat: "pdf",
		Filters: map[string]map[string]map[string]FilterCondition{
			"database": {
				"table": {
					"id": {
						In: []any{"1", "2", "3"},
					},
				},
			},
		},
		MappedFields: map[string]map[string][]string{
			"database": {
				"table": {"id", "name", "status"},
			},
		},
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, templateID.String(), result["templateId"])
	assert.Equal(t, reportID.String(), result["reportId"])
	assert.Equal(t, "pdf", result["outputFormat"])
}

func TestReportMessage_JSONUnmarshal(t *testing.T) {
	templateID := uuid.New()
	reportID := uuid.New()

	jsonData := `{
		"templateId": "` + templateID.String() + `",
		"reportId": "` + reportID.String() + `",
		"outputFormat": "html",
		"filters": {
			"db": {
				"users": {
					"age": {"gte": [18]}
				}
			}
		},
		"mappedFields": {
			"db": {
				"users": ["id", "name", "email"]
			}
		}
	}`

	var msg ReportMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	assert.NoError(t, err)

	assert.Equal(t, templateID, msg.TemplateID)
	assert.Equal(t, reportID, msg.ReportID)
	assert.Equal(t, "html", msg.OutputFormat)
	assert.NotNil(t, msg.Filters)
	assert.NotNil(t, msg.MappedFields)
}

func TestFilterCondition_AllOperators(t *testing.T) {
	filter := FilterCondition{
		Equals:         []any{"value1"},
		GreaterThan:    []any{10},
		GreaterOrEqual: []any{5},
		LessThan:       []any{100},
		LessOrEqual:    []any{50},
		Between:        []any{1, 10},
		In:             []any{"a", "b"},
		NotIn:          []any{"x", "y"},
	}

	data, err := json.Marshal(filter)
	assert.NoError(t, err)

	var result FilterCondition
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.NotNil(t, result.Equals)
	assert.NotNil(t, result.GreaterThan)
	assert.NotNil(t, result.GreaterOrEqual)
	assert.NotNil(t, result.LessThan)
	assert.NotNil(t, result.LessOrEqual)
	assert.NotNil(t, result.Between)
	assert.NotNil(t, result.In)
	assert.NotNil(t, result.NotIn)
}

func TestReportMessage_EmptyFilters(t *testing.T) {
	msg := ReportMessage{
		TemplateID:   uuid.New(),
		ReportID:     uuid.New(),
		OutputFormat: "xml",
		Filters:      nil,
		MappedFields: nil,
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	var result ReportMessage
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Nil(t, result.Filters)
	assert.Nil(t, result.MappedFields)
}
