// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pongo

import (
	"context"
	"strings"
	"testing"

	"github.com/LerianStudio/lib-commons/v2/commons/zap"
	"github.com/stretchr/testify/assert"
)

func TestRenderFromBytes_Success(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()
	tpl := []byte("Hello, {{ person._.0.name }}!")

	data := map[string]map[string][]map[string]any{
		"person": {
			"_": {
				{"name": "World"},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, World!", out)
}

func TestRenderFromBytes_SyntaxError(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()
	tpl := []byte("Hello, {{ name !")
	data := map[string]map[string][]map[string]any{
		"name": {
			"_": {
				{"name": "World"},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.Error(t, err)
	assert.Empty(t, out)
}

func TestRender_ArithmeticExpression(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()
	tpl := []byte(`Initial Balance: {{ midaz_transaction.balance.0.initial_balance }}
Final Balance: {{ midaz_transaction.balance.0.final_balance }}
Calculation: {% calc (100 + 200) * 1.2 %}`)

	data := map[string]map[string][]map[string]any{
		"midaz_transaction": {
			"balance": {
				{"initial_balance": 100.0, "final_balance": 200.0},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "Initial Balance: 100")
	assert.Contains(t, out, "Final Balance: 200")
	assert.Contains(t, out, "Calculation: 360")
}

func TestRender_ArithmeticExpressionWithVariables(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()
	tpl := []byte(`Initial Balance: {{ midaz_transaction.balance.0.initial_balance }}
Final Balance: {{ midaz_transaction.balance.0.final_balance }}
Calculation: {% calc (midaz_transaction.balance.0.initial_balance + midaz_transaction.balance.0.final_balance) * 1.2 %}`)

	data := map[string]map[string][]map[string]any{
		"midaz_transaction": {
			"balance": {
				{"initial_balance": 100.0, "final_balance": 200.0},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "Initial Balance: 100")
	assert.Contains(t, out, "Final Balance: 200")
	assert.Contains(t, out, "Calculation: 360")
}

func TestRender_CalcTag(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()
	tpl := []byte(`{% calc midaz_transaction.balance.0.initial_balance + midaz_transaction.balance.0.final_balance %}`)

	data := map[string]map[string][]map[string]any{
		"midaz_transaction": {
			"balance": {
				{"initial_balance": 100.0, "final_balance": 200.0},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "300")
}

func TestRender_CalcTagWithEmptyValue(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()
	tpl := []byte(`{% for balance in midaz_transaction.balance %}
Alias: {{ balance.alias }}
Balance: {{ balance.available }}
Sum: {% calc balance.available + 1.2 %}
{% endfor %}`)

	data := map[string]map[string][]map[string]any{
		"midaz_transaction": {
			"balance": {
				{"alias": "Account1", "available": ""},
				{"alias": "Account2", "available": "100.5"},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "Alias: Account1")
	assert.Contains(t, out, "Balance: ")
	assert.Contains(t, out, "Sum: 1.2")
	assert.Contains(t, out, "Alias: Account2")
	assert.Contains(t, out, "Balance: 100.5")
	assert.Contains(t, out, "Sum: 101.7")
}

func TestRender_CalcTagComplexExpression(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()
	tpl := []byte(`{% for balance in midaz_transaction.balance %}
Alias: {{ balance.alias }}
Balance: {{ balance.available }}
Sum: {% calc balance.available + 1.2 %}
Sum Complex: {% calc (balance.available + 1.2) * balance.on_hold - balance.available / 2 %}
{% endfor %}`)

	data := map[string]map[string][]map[string]any{
		"midaz_transaction": {
			"balance": {
				{"alias": "Account1", "available": "0", "on_hold": "3.50"},
				{"alias": "Account2", "available": "100.5", "on_hold": "2.0"},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "Alias: Account1")
	assert.Contains(t, out, "Balance: 0")
	assert.Contains(t, out, "Sum: 1.2")
	assert.Contains(t, out, "Sum Complex: 4.2")
	assert.Contains(t, out, "Alias: Account2")
	assert.Contains(t, out, "Balance: 100.5")
	assert.Contains(t, out, "Sum: 101.7")
	assert.Contains(t, out, "Sum Complex: 153.15")
}

func TestRender_CalcTagPowerOperation(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()
	tpl := []byte(`{% for balance in midaz_transaction.balance %}
Alias: {{ balance.alias }}
Balance: {{ balance.available }}
Power: {% calc balance.available ** 2 %}
{% endfor %}

Sum: {% calc (midaz_transaction.balance.3.available + 1.2) ** 2 %}`)

	data := map[string]map[string][]map[string]any{
		"midaz_transaction": {
			"balance": {
				{"alias": "Account1", "available": "3"},
				{"alias": "Account2", "available": "4"},
				{"alias": "Account3", "available": "5"},
				{"alias": "Account4", "available": "10"},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "Alias: Account1")
	assert.Contains(t, out, "Balance: 3")
	assert.Contains(t, out, "Power: 9")
	assert.Contains(t, out, "Alias: Account2")
	assert.Contains(t, out, "Balance: 4")
	assert.Contains(t, out, "Power: 16")
	assert.Contains(t, out, "Alias: Account3")
	assert.Contains(t, out, "Balance: 5")
	assert.Contains(t, out, "Power: 25")
	assert.True(t, strings.Contains(out, "Sum: 125.4") || strings.Contains(out, "Sum: 125.44") || strings.Contains(out, "Sum: 125.43999999999998"))
}

func TestRender_CalcTagIndexAccess(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()
	tpl := []byte(`{% for balance in midaz_transaction.balance %}
Alias: {{ balance.alias }}
Balance: {{ balance.available }}
Sum: {% calc balance.available + 1.2 %}
Sum Complex: {% calc (balance.available + 1.2) * balance.on_hold - balance.available / 2 %}
{% endfor %}

Sum: {% calc (midaz_transaction.balance.3.available + 1.2) ** 2 %}`)

	data := map[string]map[string][]map[string]any{
		"midaz_transaction": {
			"balance": {
				{"alias": "Account1", "available": "0", "on_hold": "3.50"}, // Only 3 elements, index 3 doesn't exist
				{"alias": "Account2", "available": "100.5", "on_hold": "2.0"},
				{"alias": "Account3", "available": "5", "on_hold": "1.0"},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "Alias: Account1")
	assert.Contains(t, out, "Alias: Account2")
	assert.Contains(t, out, "Alias: Account3")
	assert.Contains(t, out, "Sum: 1.44")
}

func TestRender_CalcTagScientificNotation(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()
	tpl := []byte(`Power Small: {% calc 0.1 ** 3 %}
Power Large: {% calc 1000 ** 2 %}
Power Fractional: {% calc 2.5 ** 0.5 %}`)

	data := map[string]map[string][]map[string]any{
		"midaz_transaction": {
			"balance": {
				{"alias": "Account1", "available": "0.1"},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "Power Small: 0.001")
	assert.Contains(t, out, "Power Large: 1000000")
	assert.Contains(t, out, "Power Fractional: 1.5811388301")
}

func TestPreprocessSchemaReferences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "converts schema syntax in for loop",
			input:    `{% for tx in external_db:sales.orders %}{{ tx.id }}{% endfor %}`,
			expected: `{% for tx in external_db.sales__orders %}{{ tx.id }}{% endfor %}`,
		},
		{
			name:     "converts multiple schema references",
			input:    `{% for tx in db1:schema1.table1 %}{% endfor %}{% for acc in db2:schema2.table2 %}{% endfor %}`,
			expected: `{% for tx in db1.schema1__table1 %}{% endfor %}{% for acc in db2.schema2__table2 %}{% endfor %}`,
		},
		{
			name:     "preserves legacy format",
			input:    `{% for tx in midaz_transaction.balance %}{{ tx.amount }}{% endfor %}`,
			expected: `{% for tx in midaz_transaction.balance %}{{ tx.amount }}{% endfor %}`,
		},
		{
			name:     "handles mixed formats",
			input:    `{% for tx in external_db:sales.orders %}{% endfor %}{% for acc in midaz.account %}{% endfor %}`,
			expected: `{% for tx in external_db.sales__orders %}{% endfor %}{% for acc in midaz.account %}{% endfor %}`,
		},
		{
			name:     "converts direct access with index",
			input:    `{{ external_db:sales.orders.0.id }}`,
			expected: `{{ external_db.sales__orders.0.id }}`,
		},
		{
			name:     "handles schema in calc expression",
			input:    `{% calc external_db:sales.orders.0.amount + external_db:sales.orders.1.amount %}`,
			expected: `{% calc external_db.sales__orders.0.amount + external_db.sales__orders.1.amount %}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := preprocessSchemaReferences(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRender_ExplicitSchemaFormat(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()

	// Template uses explicit schema syntax that will be preprocessed to sales__orders
	tpl := []byte(`{% for order in external_db:sales.orders %}
ID: {{ order.id }}, Amount: {{ order.amount }}
{% endfor %}`)

	// Data is stored using double underscore key (schema__table) for Pongo2 compatibility
	data := map[string]map[string][]map[string]any{
		"external_db": {
			"sales__orders": {
				{"id": "TX001", "amount": 100.50},
				{"id": "TX002", "amount": 200.00},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "ID: TX001, Amount: 100.5")
	assert.Contains(t, out, "ID: TX002, Amount: 200")
}

func TestRender_ExplicitSchemaDirectAccess(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()

	// Direct access to schema-qualified data with index
	tpl := []byte(`First Order ID: {{ external_db:sales.orders.0.id }}
First Amount: {{ external_db:sales.orders.0.amount }}`)

	data := map[string]map[string][]map[string]any{
		"external_db": {
			"sales__orders": {
				{"id": "TX001", "amount": 100.50},
				{"id": "TX002", "amount": 200.00},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "First Order ID: TX001")
	assert.Contains(t, out, "First Amount: 100.5")
}

func TestRender_ExplicitSchemaIfTag(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()

	tpl := []byte(`{% if external_db:sales.orders %}Has orders{% endif %}`)

	data := map[string]map[string][]map[string]any{
		"external_db": {
			"sales__orders": {
				{"id": "TX001", "amount": 100.50},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "Has orders")
}

func TestRender_ExplicitSchemaCalcTag(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()

	tpl := []byte(`Total: {% calc external_db:sales.orders.0.amount + external_db:sales.orders.1.amount %}`)

	data := map[string]map[string][]map[string]any{
		"external_db": {
			"sales__orders": {
				{"id": "TX001", "amount": 100.50},
				{"id": "TX002", "amount": 200.00},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "Total: 300.5")
}

func TestRender_MixedLegacyAndSchemaFormats(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()

	tpl := []byte(`Legacy: {{ midaz.account.0.alias }}
Schema: {{ external_db:sales.orders.0.id }}`)

	data := map[string]map[string][]map[string]any{
		"midaz": {
			"account": {
				{"alias": "ACCT001"},
			},
		},
		"external_db": {
			"sales__orders": {
				{"id": "TX001"},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Contains(t, out, "Legacy: ACCT001")
	assert.Contains(t, out, "Schema: TX001")
}
