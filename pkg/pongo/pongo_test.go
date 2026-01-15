package pongo

import (
	"strings"
	"testing"

	"github.com/flosch/pongo2/v6"
	"github.com/stretchr/testify/assert"
)

// TestInit_FiltersRegistered verifies that all custom filters are registered during init
func TestInit_FiltersRegistered(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		contains string
	}{
		{
			name:     "percent_of filter registered",
			template: "{{ 25|percent_of:100 }}",
			context:  pongo2.Context{},
			contains: "25.00%",
		},
		{
			name:     "slice_str filter registered",
			template: `{{ value|slice_str:"0:5" }}`,
			context:  pongo2.Context{"value": "HelloWorld"},
			contains: "Hello",
		},
		{
			name:     "strip_zeros filter registered",
			template: "{{ value|strip_zeros }}",
			context:  pongo2.Context{"value": "100.50000"},
			contains: "100.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			assert.NoError(t, err, "Filter should be registered")

			result, err := tpl.Execute(tt.context)
			assert.NoError(t, err)
			assert.Contains(t, result, tt.contains)
		})
	}
}

// TestInit_TagsRegistered verifies that all custom tags are registered during init
func TestInit_TagsRegistered(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		validate func(t *testing.T, result string, err error)
	}{
		{
			name:     "sum_by tag registered",
			template: `{% sum_by items by "value" %}`,
			context: pongo2.Context{
				"items": []map[string]any{
					{"value": 10},
					{"value": 20},
					{"value": 30},
				},
			},
			validate: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, result, "60")
			},
		},
		{
			name:     "count_by tag registered",
			template: `{% count_by items %}`,
			context: pongo2.Context{
				"items": []map[string]any{
					{"name": "a"},
					{"name": "b"},
					{"name": "c"},
				},
			},
			validate: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, result, "3")
			},
		},
		{
			name:     "avg_by tag registered",
			template: `{% avg_by items by "value" %}`,
			context: pongo2.Context{
				"items": []map[string]any{
					{"value": 10},
					{"value": 20},
					{"value": 30},
				},
			},
			validate: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, result, "20")
			},
		},
		{
			name:     "min_by tag registered",
			template: `{% min_by items by "value" %}`,
			context: pongo2.Context{
				"items": []map[string]any{
					{"value": 30},
					{"value": 10},
					{"value": 20},
				},
			},
			validate: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, result, "10")
			},
		},
		{
			name:     "max_by tag registered",
			template: `{% max_by items by "value" %}`,
			context: pongo2.Context{
				"items": []map[string]any{
					{"value": 30},
					{"value": 10},
					{"value": 20},
				},
			},
			validate: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, result, "30")
			},
		},
		{
			name:     "date_time tag registered",
			template: `{% date_time "2006-01-02" %}`,
			context:  pongo2.Context{},
			validate: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				// Should contain a date in YYYY-MM-DD format
				assert.Regexp(t, `\d{4}-\d{2}-\d{2}`, result)
			},
		},
		{
			name:     "calc tag registered",
			template: "{% calc (10 + 5) %}",
			context:  pongo2.Context{},
			validate: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, result, "15")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			if err != nil {
				tt.validate(t, "", err)
				return
			}

			result, err := tpl.Execute(tt.context)
			tt.validate(t, result, err)
		})
	}
}

// TestInit_PercentOfFilter_EdgeCases tests edge cases for percent_of filter
func TestInit_PercentOfFilter_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		context     pongo2.Context
		expectError bool
		contains    string
	}{
		{
			name:        "zero numerator",
			template:    "{{ 0|percent_of:100 }}",
			context:     pongo2.Context{},
			expectError: false,
			contains:    "0.00%",
		},
		{
			name:        "large numbers",
			template:    "{{ 1000000|percent_of:2000000 }}",
			context:     pongo2.Context{},
			expectError: false,
			contains:    "50.00%",
		},
		{
			name:        "decimal result",
			template:    "{{ 1|percent_of:3 }}",
			context:     pongo2.Context{},
			expectError: false,
			contains:    "33.33%",
		},
		{
			name:        "zero denominator",
			template:    "{{ 10|percent_of:0 }}",
			context:     pongo2.Context{},
			expectError: true,
			contains:    "NaN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			assert.NoError(t, err)

			result, err := tpl.Execute(tt.context)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, result, tt.contains)
			}
		})
	}
}

// TestInit_SliceFilter_EdgeCases tests edge cases for slice_str filter
func TestInit_SliceFilter_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		expected string
	}{
		{
			name:     "slice from start",
			template: `{{ value|slice_str:"0:3" }}`,
			context:  pongo2.Context{"value": "Hello"},
			expected: "Hel",
		},
		{
			name:     "slice middle",
			template: `{{ value|slice_str:"2:5" }}`,
			context:  pongo2.Context{"value": "Hello World"},
			expected: "llo",
		},
		{
			name:     "empty string",
			template: `{{ value|slice_str:"0:5" }}`,
			context:  pongo2.Context{"value": ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			assert.NoError(t, err)

			result, err := tpl.Execute(tt.context)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, strings.TrimSpace(result))
		})
	}
}

// TestInit_StripZerosFilter_EdgeCases tests edge cases for strip_zeros filter
func TestInit_StripZerosFilter_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		expected string
	}{
		{
			name:     "trailing zeros",
			template: "{{ value|strip_zeros }}",
			context:  pongo2.Context{"value": "123.4500"},
			expected: "123.45",
		},
		{
			name:     "all decimal zeros",
			template: "{{ value|strip_zeros }}",
			context:  pongo2.Context{"value": "100.0000"},
			expected: "100",
		},
		{
			name:     "no trailing zeros",
			template: "{{ value|strip_zeros }}",
			context:  pongo2.Context{"value": "123.45"},
			expected: "123.45",
		},
		{
			name:     "integer value",
			template: "{{ value|strip_zeros }}",
			context:  pongo2.Context{"value": "100"},
			expected: "100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			assert.NoError(t, err)

			result, err := tpl.Execute(tt.context)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, strings.TrimSpace(result))
		})
	}
}

// TestInit_CalcTag_Operations tests various calc tag operations
func TestInit_CalcTag_Operations(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		contains string
	}{
		{
			name:     "addition",
			template: "{% calc 10 + 5 %}",
			context:  pongo2.Context{},
			contains: "15",
		},
		{
			name:     "subtraction",
			template: "{% calc 10 - 3 %}",
			context:  pongo2.Context{},
			contains: "7",
		},
		{
			name:     "multiplication",
			template: "{% calc 10 * 5 %}",
			context:  pongo2.Context{},
			contains: "50",
		},
		{
			name:     "division",
			template: "{% calc 20 / 4 %}",
			context:  pongo2.Context{},
			contains: "5",
		},
		{
			name:     "with variables",
			template: "{% calc a + b %}",
			context:  pongo2.Context{"a": 15, "b": 25},
			contains: "40",
		},
		{
			name:     "complex expression",
			template: "{% calc (10 + 5) * 2 %}",
			context:  pongo2.Context{},
			contains: "30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			assert.NoError(t, err)

			result, err := tpl.Execute(tt.context)
			assert.NoError(t, err)
			assert.Contains(t, result, tt.contains)
		})
	}
}

// TestInit_AggregationTags_EmptyArray tests aggregation tags with empty arrays
func TestInit_AggregationTags_EmptyArray(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  pongo2.Context
	}{
		{
			name:     "sum_by empty array",
			template: `{% sum_by items by "value" %}`,
			context:  pongo2.Context{"items": []map[string]any{}},
		},
		{
			name:     "count_by empty array",
			template: `{% count_by items %}`,
			context:  pongo2.Context{"items": []map[string]any{}},
		},
		{
			name:     "avg_by empty array",
			template: `{% avg_by items by "value" %}`,
			context:  pongo2.Context{"items": []map[string]any{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			assert.NoError(t, err)

			// Should not panic with empty arrays
			_, err = tpl.Execute(tt.context)
			// Error or no error is acceptable, just shouldn't panic
			_ = err
		})
	}
}

// TestInit_AllFiltersAndTagsAccessible verifies all filters and tags can be parsed
func TestInit_AllFiltersAndTagsAccessible(t *testing.T) {
	// Test all filters are parseable
	filters := []string{
		"{{ value|percent_of:100 }}",
		`{{ value|slice_str:"0:5" }}`,
		"{{ value|strip_zeros }}",
	}

	for _, filter := range filters {
		t.Run("filter_"+filter, func(t *testing.T) {
			_, err := pongo2.FromString(filter)
			assert.NoError(t, err, "Filter should be parseable: %s", filter)
		})
	}

	// Test all tags are parseable
	tags := []string{
		`{% sum_by items by "value" %}`,
		`{% count_by items %}`,
		`{% avg_by items by "value" %}`,
		`{% min_by items by "value" %}`,
		`{% max_by items by "value" %}`,
		`{% date_time "2006-01-02" %}`,
		`{% calc 1 + 1 %}`,
	}

	for _, tag := range tags {
		t.Run("tag_"+tag, func(t *testing.T) {
			_, err := pongo2.FromString(tag)
			assert.NoError(t, err, "Tag should be parseable: %s", tag)
		})
	}
}

// TestInit_DateTimeTag_BasicUsage tests date_time tag basic usage
func TestInit_DateTimeTag_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		template string
		pattern  string
	}{
		{
			name:     "date format YYYY-MM-DD",
			template: `{% date_time "2006-01-02" %}`,
			pattern:  `\d{4}-\d{2}-\d{2}`,
		},
		{
			name:     "date format DD/MM/YYYY",
			template: `{% date_time "02/01/2006" %}`,
			pattern:  `\d{2}/\d{2}/\d{4}`,
		},
		{
			name:     "date format MM-DD-YYYY",
			template: `{% date_time "01-02-2006" %}`,
			pattern:  `\d{2}-\d{2}-\d{4}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			assert.NoError(t, err)

			result, err := tpl.Execute(pongo2.Context{})
			assert.NoError(t, err)
			// Should contain a valid date
			assert.Regexp(t, tt.pattern, result)
		})
	}
}

// TestInit_DateTimeTag_Formats tests date_time tag with different formats
func TestInit_DateTimeTag_Formats(t *testing.T) {
	tests := []struct {
		name     string
		template string
		pattern  string
	}{
		{
			name:     "date only format",
			template: `{% date_time "2006-01-02" %}`,
			pattern:  `\d{4}-\d{2}-\d{2}`,
		},
		{
			name:     "date and time format",
			template: `{% date_time "2006-01-02 15:04:05" %}`,
			pattern:  `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`,
		},
		{
			name:     "custom format",
			template: `{% date_time "02/01/2006" %}`,
			pattern:  `\d{2}/\d{2}/\d{4}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			assert.NoError(t, err)

			result, err := tpl.Execute(pongo2.Context{})
			assert.NoError(t, err)
			assert.Regexp(t, tt.pattern, result)
		})
	}
}
