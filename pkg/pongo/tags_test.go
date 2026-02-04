// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pongo

import (
	"testing"

	"github.com/flosch/pongo2/v6"
	"github.com/stretchr/testify/assert"
)

func TestSumByTag(t *testing.T) {
	tplStr := `{% sum_by data by "amount" %}`
	tpl, err := pongo2.FromString(tplStr)
	assert.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"amount": 1000},
			{"amount": 2500},
			{"amount": 1500},
		},
	}

	out, err := tpl.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "5000", out)
}

func TestCountByTagWithFilter(t *testing.T) {
	tplStr := `{% count_by data if amount > 1000 %}`
	tpl, err := pongo2.FromString(tplStr)
	assert.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"amount": 1000},
			{"amount": 2500},
			{"amount": 1500},
		},
	}

	out, err := tpl.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "2", out)
}

func TestAvgByTag(t *testing.T) {
	tplStr := `{% avg_by data by "amount" %}`
	tpl, err := pongo2.FromString(tplStr)
	assert.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"amount": 1000},
			{"amount": 2000},
			{"amount": 3000},
		},
	}

	out, err := tpl.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "2000", out)
}

func TestMinByTag(t *testing.T) {
	tplStr := `{% min_by data by "amount" %}`
	tpl, err := pongo2.FromString(tplStr)
	assert.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"amount": 4000},
			{"amount": 1500},
			{"amount": 5000},
		},
	}

	out, err := tpl.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "1500", out)
}

func TestMaxByTag(t *testing.T) {
	tplStr := `{% max_by data by "amount" %}`
	tpl, err := pongo2.FromString(tplStr)
	assert.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"amount": 1000},
			{"amount": 8000},
			{"amount": 3200},
		},
	}

	out, err := tpl.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "8000", out)
}

func TestCalcTag_BasicOperations(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"addition", `{% calc 10 + 5 %}`, "15"},
		{"subtraction", `{% calc 10 - 5 %}`, "5"},
		{"multiplication", `{% calc 10 * 5 %}`, "50"},
		{"division", `{% calc 10 / 5 %}`, "2"},
		{"power", `{% calc 2 ** 3 %}`, "8"},
		{"parentheses", `{% calc (10 + 5) * 2 %}`, "30"},
		{"decimal", `{% calc 10.5 + 5.5 %}`, "16"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			assert.NoError(t, err)

			out, err := tpl.Execute(pongo2.Context{})
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestCalcTag_NegativeNumbers(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		// Simple negative number
		{"single_negative", `{% calc -5 %}`, "-5"},

		// Subtraction resulting in negative
		{"subtraction_negative_result", `{% calc 3 - 10 %}`, "-7"},

		// Multiplication with negative
		{"multiply_negative", `{% calc -5 * 2 %}`, "-10"},
		{"multiply_two_negatives", `{% calc -5 * -2 %}`, "10"},

		// Division with negative
		{"divide_negative", `{% calc -10 / 2 %}`, "-5"},
		{"divide_by_negative", `{% calc 10 / -2 %}`, "-5"},

		// Parentheses with negative
		{"parentheses_negative", `{% calc (-5) + 10 %}`, "5"},
		{"parentheses_double_negative", `{% calc (-5) * (-2) %}`, "10"},

		// Addition with negative operand (5 + -3 = 2)
		{"add_negative_operand", `{% calc 5 + -3 %}`, "2"},

		// Subtraction of negative (5 - -3 = 8)
		{"subtract_negative_operand", `{% calc 5 - -3 %}`, "8"},

		// Complex with negatives
		{"complex_negative", `{% calc (-10 + 5) * 2 %}`, "-10"},
		{"complex_negative_result", `{% calc (5 - 15) / 2 %}`, "-5"},

		// Power with negative base
		{"power_negative_base", `{% calc -2 ** 2 %}`, "4"},
		{"power_negative_base_odd", `{% calc -2 ** 3 %}`, "-8"},

		// Decimal negatives
		{"decimal_negative", `{% calc -5.5 + 2.5 %}`, "-3"},
		{"decimal_negative_result", `{% calc 2.5 - 5.5 %}`, "-3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			assert.NoError(t, err, "template parsing should not fail")

			out, err := tpl.Execute(pongo2.Context{})
			assert.NoError(t, err, "template execution should not fail")
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestCalcTag_DivisionByZero(t *testing.T) {
	tpl, err := pongo2.FromString(`{% calc 10 / 0 %}`)
	assert.NoError(t, err, "template parsing should not fail")

	_, err = tpl.Execute(pongo2.Context{})
	assert.Error(t, err, "division by zero should return an error")
}

func TestCalcTag_NegativeWithVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		expected string
	}{
		{
			name:     "variable_with_negative_value",
			template: `{% calc value + 10 %}`,
			context:  pongo2.Context{"value": -5},
			expected: "5",
		},
		{
			name:     "subtract_from_negative_variable",
			template: `{% calc value - 3 %}`,
			context:  pongo2.Context{"value": -5},
			expected: "-8",
		},
		{
			name:     "multiply_negative_variables",
			template: `{% calc a * b %}`,
			context:  pongo2.Context{"a": -5, "b": -2},
			expected: "10",
		},
		{
			name:     "nested_negative_values",
			template: `{% calc data.amount * 2 %}`,
			context: pongo2.Context{
				"data": map[string]any{"amount": -100},
			},
			expected: "-200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			assert.NoError(t, err, "template parsing should not fail")

			out, err := tpl.Execute(tt.context)
			assert.NoError(t, err, "template execution should not fail")
			assert.Equal(t, tt.expected, out)
		})
	}
}
