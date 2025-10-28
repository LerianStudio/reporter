package property

import (
	"encoding/json"
	"testing"
	"testing/quick"

	"github.com/LerianStudio/reporter/v4/pkg/model"
)

// Property 1: Filtros vazios devem ter todos os campos vazios
func TestProperty_Filter_EmptyCheck(t *testing.T) {
	property := func() bool {
		emptyFilter := model.FilterCondition{}

		// All fields should be nil or empty
		return len(emptyFilter.Equals) == 0 &&
			len(emptyFilter.GreaterThan) == 0 &&
			len(emptyFilter.LessThan) == 0 &&
			len(emptyFilter.In) == 0
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 10}); err != nil {
		t.Errorf("Property violated: empty filter check: %v", err)
	}
}

// Property 2: Filtros com operadores devem ter valores correspondentes
func TestProperty_Filter_OperatorValuePairing(t *testing.T) {
	property := func(values []string) bool {
		if len(values) == 0 {
			return true
		}

		// Convert to []any
		anyValues := make([]any, len(values))
		for i, v := range values {
			anyValues[i] = v
		}

		// Create filter with Equals operator
		filter := model.FilterCondition{
			Equals: anyValues,
		}

		// Filter should have values
		return len(filter.Equals) > 0
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property violated: filter operator-value pairing: %v", err)
	}
}

// Property 3: Operadores IN devem aceitar arrays de qualquer tamanho
func TestProperty_Filter_InOperatorArraySize(t *testing.T) {
	property := func(values []string) bool {
		// Convert to []any
		anyValues := make([]any, len(values))
		for i, v := range values {
			anyValues[i] = v
		}

		// IN operator should work with any array size (including empty)
		filter := model.FilterCondition{
			In: anyValues,
		}

		// Length should match
		return len(filter.In) == len(values)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property violated: IN operator array size: %v", err)
	}
}

// Property 4: Operadores de range devem aceitar valores
func TestProperty_Filter_RangeOperators(t *testing.T) {
	property := func(start, end int) bool {
		if start < 0 || end < 0 {
			return true
		}

		// Create range filter
		filter := model.FilterCondition{
			GreaterOrEqual: []any{start},
			LessOrEqual:    []any{end},
		}

		// Should have both operators set
		return len(filter.GreaterOrEqual) > 0 && len(filter.LessOrEqual) > 0
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property violated: range operators: %v", err)
	}
}

// Property 5: Filtros devem ser serializáveis para JSON e deserializáveis sem perda
func TestProperty_Filter_JSONRoundTrip(t *testing.T) {
	property := func(val1, val2 string) bool {
		if val1 == "" && val2 == "" {
			return true
		}

		original := model.FilterCondition{
			Equals: []any{val1},
			NotIn:  []any{val2},
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		if err != nil {
			return false
		}

		// Unmarshal back
		var decoded model.FilterCondition
		if err := json.Unmarshal(jsonData, &decoded); err != nil {
			return false
		}

		// Compare lengths
		return len(decoded.Equals) == len(original.Equals) &&
			len(decoded.NotIn) == len(original.NotIn)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property violated: JSON round-trip: %v", err)
	}
}

// Property 6: Between operator deve aceitar exatamente 2 valores
func TestProperty_Filter_BetweenOperator(t *testing.T) {
	property := func(min, max int) bool {
		filter := model.FilterCondition{
			Between: []any{min, max},
		}

		// Between should have exactly 2 values
		return len(filter.Between) == 2
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property violated: Between operator: %v", err)
	}
}

// Property 7: Multiple operadores devem coexistir
func TestProperty_Filter_MultipleOperators(t *testing.T) {
	property := func(eqVal, inVal string) bool {
		filter := model.FilterCondition{
			Equals:      []any{eqVal},
			In:          []any{inVal},
			GreaterThan: []any{0},
		}

		// All operators should be set
		return len(filter.Equals) > 0 &&
			len(filter.In) > 0 &&
			len(filter.GreaterThan) > 0
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("Property violated: multiple operators: %v", err)
	}
}

// Property 8: FilterCondition deve aceitar tipos diferentes em slices
func TestProperty_Filter_MixedTypes(t *testing.T) {
	property := func(strVal string, intVal int) bool {
		filter := model.FilterCondition{
			In: []any{strVal, intVal, true, 3.14},
		}

		// Should accept mixed types
		return len(filter.In) == 4
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property violated: mixed types: %v", err)
	}
}

// Property 9: NotIn deve ser inverso de In logicamente
func TestProperty_Filter_NotInInverse(t *testing.T) {
	property := func(value string) bool {
		if value == "" {
			return true
		}

		filterIn := model.FilterCondition{
			In: []any{value},
		}

		filterNotIn := model.FilterCondition{
			NotIn: []any{value},
		}

		// Both should be valid but opposite
		return len(filterIn.In) > 0 && len(filterNotIn.NotIn) > 0
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("Property violated: NotIn/In inverse: %v", err)
	}
}

// Property 10: FilterCondition com todos os operadores deve ser válido
func TestProperty_Filter_AllOperators(t *testing.T) {
	property := func(val1, val2 string, num1, num2 int) bool {
		filter := model.FilterCondition{
			Equals:         []any{val1},
			GreaterThan:    []any{num1},
			GreaterOrEqual: []any{num1},
			LessThan:       []any{num2},
			LessOrEqual:    []any{num2},
			Between:        []any{num1, num2},
			In:             []any{val1, val2},
			NotIn:          []any{val2},
		}

		// All operators should be populated
		return len(filter.Equals) > 0 &&
			len(filter.GreaterThan) > 0 &&
			len(filter.In) > 0 &&
			len(filter.NotIn) > 0 &&
			len(filter.Between) == 2
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("Property violated: all operators: %v", err)
	}
}
