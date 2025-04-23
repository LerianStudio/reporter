package pongo

import (
	"testing"

	"github.com/flosch/pongo2/v6"
	"github.com/stretchr/testify/assert"
)

func TestScaleFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		scale    int
		expect   string
		hasError bool
	}{
		{"int", 10000, 2, "100.00", false},
		{"int_small", 135, 1, "13.5", false},
		{"string_number", "20000", 3, "20.000", false},
		{"int64", int64(123456), 4, "12.3456", false},
		{"invalid_string", "abc", 2, "NaN", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := scaleFilter(pongo2.AsValue(test.input), pongo2.AsValue(test.scale))
			t.Logf("input=%v (%T), scale=%d → output=%s, err=%v", test.input, test.input, test.scale, val.String(), err)

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
