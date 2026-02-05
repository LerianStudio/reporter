// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pongo

import (
	"testing"

	"github.com/flosch/pongo2/v6"
	"github.com/stretchr/testify/assert"
)

func TestPercentOfFilter(t *testing.T) {
	tests := []struct {
		name     string
		num      any
		total    any
		expect   string
		hasError bool
	}{
		{"basic", 25, 100, "25.00%", false},
		{"fraction", 1, 4, "25.00%", false},
		{"string_inputs", "500", "1000", "50.00%", false},
		{"zero_denominator", 10, 0, "NaN", true},
		{"invalid_input", "abc", 100, "NaN", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := percentOfFilter(pongo2.AsValue(test.num), pongo2.AsValue(test.total))
			t.Logf("num=%v, total=%v → output=%s, err=%v", test.num, test.total, val.String(), err)

			if test.hasError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			assert.Equal(t, test.expect, val.String())
		})
	}
}

func TestStripZerosFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"integer", 100, "100"},
		{"int64", int64(200), "200"},
		{"float_with_zeros", 3.14000, "3.14"},
		{"float_whole_number", 5.0, "5"},
		{"string_decimal", "123.45000", "123.45"},
		{"string_integer", "100", "100"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := stripZerosFilter(pongo2.AsValue(test.input), pongo2.AsValue(""))
			assert.Nil(t, err)
			assert.Equal(t, test.expected, val.String())
		})
	}
}

func TestStripZerosFilter_InvalidString(t *testing.T) {
	val, err := stripZerosFilter(pongo2.AsValue("not_a_number"), pongo2.AsValue(""))
	assert.NotNil(t, err)
	assert.Equal(t, "NaN", val.String())
}

func TestSliceFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		param    string
		expected string
		hasError bool
	}{
		{"basic_slice", "Hello World", "0:5", "Hello", false},
		{"middle_slice", "Hello World", "6:11", "World", false},
		{"full_string", "Test", "0:4", "Test", false},
		{"empty_result", "Test", "0:0", "", false},
		{"out_of_bounds_end", "Test", "0:100", "Test", false},
		{"invalid_format", "Test", "0-5", "", true},
		{"invalid_start", "Test", "abc:5", "", true},
		{"invalid_end", "Test", "0:xyz", "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := sliceFilter(pongo2.AsValue(test.input), pongo2.AsValue(test.param))

			if test.hasError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, val.String())
			}
		})
	}
}

func TestSliceFilter_NegativeStart(t *testing.T) {
	val, err := sliceFilter(pongo2.AsValue("Hello"), pongo2.AsValue("-5:5"))
	assert.Nil(t, err)
	assert.Equal(t, "Hello", val.String())
}

func TestSliceFilter_StartGreaterThanEnd(t *testing.T) {
	val, err := sliceFilter(pongo2.AsValue("Hello"), pongo2.AsValue("5:2"))
	assert.Nil(t, err)
	assert.Equal(t, "", val.String())
}

func TestEvaluateArithmeticExpression_BasicOperations(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		expected   float64
		wantErr    bool
	}{
		{"addition", "5+3", 8, false},
		{"subtraction", "10-4", 6, false},
		{"multiplication", "6*7", 42, false},
		{"division", "20/4", 5, false},
		{"power", "2**3", 8, false},
		{"complex_expression", "2+3*4", 14, false},
		{"parentheses", "(2+3)*4", 20, false},
		{"negative_number", "-5+10", 5, false},
		{"decimal_numbers", "3.5*2", 7, false},
		{"division_decimal_result", "7/2", 3.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateArithmeticExpression(tt.expression)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expected, result, 0.0001)
			}
		})
	}
}

func TestEvaluateArithmeticExpression_Errors(t *testing.T) {
	tests := []struct {
		name       string
		expression string
	}{
		{"empty_expression", ""},
		{"division_by_zero", "10/0"},
		{"invalid_character", "5+abc"},
		{"unmatched_parentheses", "(5+3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluateArithmeticExpression(tt.expression)
			assert.Error(t, err)
		})
	}
}

func TestEvaluateArithmeticExpression_Precedence(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		expected   float64
	}{
		{"mult_before_add", "2+3*4", 14},
		{"div_before_sub", "10-6/2", 7},
		{"power_before_mult", "2*3**2", 18},
		{"nested_parentheses", "((2+3)*4)+1", 21},
		{"multiple_operations", "1+2*3-4/2+5", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateArithmeticExpression(tt.expression)
			assert.NoError(t, err)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestParseNumber(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedStr    string
		expectedLength int
	}{
		{"simple_integer", "123", "123", 3},
		{"decimal_number", "3.14", "3.14", 4},
		{"negative_number", "-42", "-42", 3},
		{"number_followed_by_operator", "123+", "123", 3},
		{"empty_string", "", "", 0},
		{"starts_with_non_digit", "abc", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, length := parseNumber(tt.input)
			assert.Equal(t, tt.expectedStr, str)
			assert.Equal(t, tt.expectedLength, length)
		})
	}
}

func TestIsDigit(t *testing.T) {
	tests := []struct {
		char     byte
		expected bool
	}{
		{'0', true},
		{'5', true},
		{'9', true},
		{'a', false},
		{'z', false},
		{'.', false},
		{'-', false},
		{'+', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			result := isDigit(tt.char)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"integer_as_float", 42.0, "42.0000000000"},
		{"decimal", 3.14, "3.1400000000"},
		{"negative", -5.5, "-5.5000000000"},
		{"zero", 0.0, "0.0000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenizeExpression(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		wantErr    bool
		numTokens  int
	}{
		{"simple_addition", "2+3", false, 3},
		{"multiple_operations", "1+2*3", false, 5},
		{"negative_number_at_start", "-5+3", false, 3},
		{"power_operator", "2**3", false, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := tokenizeExpression(tt.expression)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, tokens, tt.numTokens)
			}
		})
	}
}

func TestEvaluateTokens_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		want    float64
		wantErr bool
	}{
		{"empty_tokens", []token{}, 0, true},
		{"single_number", []token{{isOperator: false, value: 42}}, 42, false},
		{"single_operator", []token{{isOperator: true, operator: "+"}}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateTokens(tt.tokens)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.want, result, 0.0001)
			}
		})
	}
}

func TestEvaluateWithParentheses(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		expected   float64
		wantErr    bool
	}{
		{"simple_parentheses", "(5+3)", 8, false},
		{"nested_parentheses", "((2+3)*4)", 20, false},
		{"parentheses_change_precedence", "(2+3)*4", 20, false},
		{"multiple_parentheses_groups", "(2+3)*(4+1)", 25, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateWithParentheses(tt.expression)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expected, result, 0.0001)
			}
		})
	}
}

func TestToken_Struct(t *testing.T) {
	numToken := token{isOperator: false, value: 42.5}
	assert.False(t, numToken.isOperator)
	assert.Equal(t, 42.5, numToken.value)

	opToken := token{isOperator: true, operator: "+"}
	assert.True(t, opToken.isOperator)
	assert.Equal(t, "+", opToken.operator)
}
