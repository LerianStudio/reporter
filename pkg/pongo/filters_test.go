// Copyright (c) 2025 Lerian Studio. All rights reserved.
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
