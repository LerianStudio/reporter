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
