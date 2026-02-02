package template_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMappedFieldsOfTemplate_CalcTag(t *testing.T) {
	template := `{% for balance in midaz_transaction.balance %}
Alias: {{ balance.alias }}
Balance: {{ balance.available }}
Sum: {% calc balance.available + 1.2 %}
Sum Complex: {% calc (balance.available + 1.2) * balance.on_hold - balance.available / 2 %}
{% endfor %}

Sum: {% calc midaz_transaction.balance.3.available + 1.2 %}`

	result := MappedFieldsOfTemplate(template)

	// Verify that all required fields are mapped
	assert.NotNil(t, result)
	assert.Contains(t, result, "midaz_transaction")

	midazTransaction := result["midaz_transaction"]
	assert.Contains(t, midazTransaction, "balance")

	balanceFields := midazTransaction["balance"]

	// Check that all fields used in the template are mapped
	expectedFields := []string{"alias", "available", "on_hold"}
	for _, field := range expectedFields {
		assert.Contains(t, balanceFields, field, "Field %s should be mapped", field)
	}
}

func TestMappedFieldsOfTemplate_CalcTagComplex(t *testing.T) {
	template := `{% calc (midaz_transaction.balance.0.initial_balance + midaz_transaction.balance.0.final_balance) * 1.2 %}`

	result := MappedFieldsOfTemplate(template)

	assert.NotNil(t, result)
	assert.Contains(t, result, "midaz_transaction")

	midazTransaction := result["midaz_transaction"]
	assert.Contains(t, midazTransaction, "balance")

	balanceFields := midazTransaction["balance"]
	expectedFields := []string{"initial_balance", "final_balance"}

	for _, field := range expectedFields {
		assert.Contains(t, balanceFields, field, "Field %s should be mapped", field)
	}
}

func TestMappedFieldsOfTemplate_NestedLoopWithParentVariable(t *testing.T) {
	template := `
{% for alias in plugin_crm.aliases %}
  Alias ID: {{ alias.account_id }}
  {% for related_party in alias.related_parties %}
    Role: {{ related_party.role }}
    Start: {{ related_party.start_date }}
  {% endfor %}
{% endfor %}`

	result := MappedFieldsOfTemplate(template)

	// Should map to plugin_crm
	assert.NotNil(t, result)
	assert.Contains(t, result, "plugin_crm", "plugin_crm should exist in result")

	pluginCRM := result["plugin_crm"]
	assert.Contains(t, pluginCRM, "aliases", "aliases should exist under plugin_crm")

	// Get aliases fields
	aliasesFields := pluginCRM["aliases"]

	// Should contain account_id (direct field from alias)
	assert.Contains(t, aliasesFields, "account_id", "account_id should be in aliases fields")

	// Should contain related_parties (nested loop resolved to same table)
	assert.Contains(t, aliasesFields, "related_parties", "related_parties should be captured under aliases")

	// Should contain nested fields with prefix (related_parties.field)
	assert.Contains(t, aliasesFields, "related_parties.role", "related_parties.role should be in aliases fields (from nested loop)")
	assert.Contains(t, aliasesFields, "related_parties.start_date", "related_parties.start_date should be in aliases fields (from nested loop)")
}

func TestResolveNestedVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string][]string
		expected map[string][]string
	}{
		{
			name: "single level - no resolution needed",
			input: map[string][]string{
				"account": {"midaz_onboarding", "account"},
			},
			expected: map[string][]string{
				"account": {"midaz_onboarding", "account"},
			},
		},
		{
			name: "nested loop - resolve parent variable",
			input: map[string][]string{
				"alias":         {"plugin_crm", "aliases"},
				"related_party": {"alias", "related_parties"},
			},
			expected: map[string][]string{
				"alias":         {"plugin_crm", "aliases"},
				"related_party": {"plugin_crm", "aliases", "related_parties"},
			},
		},
		{
			name: "triple nested - resolve chain",
			input: map[string][]string{
				"account":     {"midaz", "accounts"},
				"transaction": {"account", "transactions"},
				"entry":       {"transaction", "entries"},
			},
			expected: map[string][]string{
				"account":     {"midaz", "accounts"},
				"transaction": {"midaz", "accounts", "transactions"},
				"entry":       {"midaz", "accounts", "transactions", "entries"},
			},
		},
		{
			name: "independent loops - no resolution",
			input: map[string][]string{
				"account": {"midaz_onboarding", "account"},
				"balance": {"midaz_transaction", "balance"},
			},
			expected: map[string][]string{
				"account": {"midaz_onboarding", "account"},
				"balance": {"midaz_transaction", "balance"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			input := make(map[string][]string)
			for k, v := range tt.input {
				input[k] = append([]string{}, v...)
			}

			resolveNestedVariables(input)

			for varName, expectedPath := range tt.expected {
				assert.Equal(t, expectedPath, input[varName], "Variable %s should have correct resolved path", varName)
			}
		})
	}
}
