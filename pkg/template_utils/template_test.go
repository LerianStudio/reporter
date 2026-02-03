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

func TestMappedFieldsOfTemplate_DIMPFilters(t *testing.T) {
	template := `|0000|{{ midaz_onboarding.onboarding.0.legal_document|replace:".:"|replace:"/:"|replace:"-:" }}|{{ midaz_onboarding.onboarding.0.legal_name }}|
{% for acc in midaz_onboarding.account|where:"type:cacc" %}|1100|SP|{{ acc.id }}|{{ acc.alias|replace:"@:"|replace:"_:/" }}|
{% endfor %}|TOTAL_SP|{{ midaz_transaction.transaction|where:"status:APPROVED"|sum:"amount" }}|
|9900|1100|{{ midaz_transaction.transaction|count:"tipo:1100" }}|`

	result := MappedFieldsOfTemplate(template)

	t.Logf("Result: %+v", result)

	assert.NotNil(t, result)

	// Check midaz_onboarding
	assert.Contains(t, result, "midaz_onboarding")
	midazOnboarding := result["midaz_onboarding"]
	t.Logf("midaz_onboarding: %+v", midazOnboarding)

	// Check onboarding fields
	assert.Contains(t, midazOnboarding, "onboarding")
	onboardingFields := midazOnboarding["onboarding"]
	assert.Contains(t, onboardingFields, "legal_document", "legal_document should be mapped")
	assert.Contains(t, onboardingFields, "legal_name", "legal_name should be mapped")

	// Check account fields (from for loop with where filter)
	assert.Contains(t, midazOnboarding, "account")
	accountFields := midazOnboarding["account"]
	t.Logf("account fields: %+v", accountFields)
	assert.Contains(t, accountFields, "type", "type should be mapped from where filter")
	assert.Contains(t, accountFields, "id", "id should be mapped")
	assert.Contains(t, accountFields, "alias", "alias should be mapped")

	// Check midaz_transaction
	assert.Contains(t, result, "midaz_transaction")
	midazTransaction := result["midaz_transaction"]
	t.Logf("midaz_transaction: %+v", midazTransaction)

	assert.Contains(t, midazTransaction, "transaction")
	transactionFields := midazTransaction["transaction"]
	t.Logf("transaction fields: %+v", transactionFields)
	assert.Contains(t, transactionFields, "status", "status should be mapped from where filter")
	assert.Contains(t, transactionFields, "amount", "amount should be mapped from sum filter")
	assert.Contains(t, transactionFields, "tipo", "tipo should be mapped from count filter")
}

func TestMappedFieldsOfTemplate_WhereFilter(t *testing.T) {
	template := `{{ operations|where:"uf:SP"|sum:"value" }}`

	result := MappedFieldsOfTemplate(template)

	assert.NotNil(t, result)
	// Note: "operations" alone isn't a valid datasource.collection path,
	// but the filter fields should still be captured if we had a proper path
}

func TestMappedFieldsOfTemplate_ChainedFilters(t *testing.T) {
	template := `{{ midaz_data.records|where:"active:true"|where:"type:A"|sum:"amount"|count:"status:done" }}`

	result := MappedFieldsOfTemplate(template)

	assert.NotNil(t, result)
	assert.Contains(t, result, "midaz_data")

	midazData := result["midaz_data"]
	assert.Contains(t, midazData, "records")

	recordsFields := midazData["records"]
	assert.Contains(t, recordsFields, "active", "active should be mapped from first where filter")
	assert.Contains(t, recordsFields, "type", "type should be mapped from second where filter")
	assert.Contains(t, recordsFields, "amount", "amount should be mapped from sum filter")
	assert.Contains(t, recordsFields, "status", "status should be mapped from count filter")
}

func TestMappedFieldsOfTemplate_ForLoopWithWhereFilter(t *testing.T) {
	template := `{% for item in midaz_source.items|where:"category:electronics" %}
{{ item.name }} - {{ item.price }}
{% endfor %}`

	result := MappedFieldsOfTemplate(template)

	assert.NotNil(t, result)
	assert.Contains(t, result, "midaz_source")

	midazSource := result["midaz_source"]
	assert.Contains(t, midazSource, "items")

	itemsFields := midazSource["items"]
	assert.Contains(t, itemsFields, "category", "category should be mapped from where filter in for loop")
	assert.Contains(t, itemsFields, "name", "name should be mapped from loop variable usage")
	assert.Contains(t, itemsFields, "price", "price should be mapped from loop variable usage")
}

func TestMappedFieldsOfTemplate_CountByTagQuotedValues(t *testing.T) {
	template := `{% count_by midaz_onboarding.account if type == "cacc" %}`

	result := MappedFieldsOfTemplate(template)

	t.Logf("Result: %+v", result)

	assert.NotNil(t, result)
	assert.Contains(t, result, "midaz_onboarding")

	midazOnboarding := result["midaz_onboarding"]
	assert.Contains(t, midazOnboarding, "account")

	accountFields := midazOnboarding["account"]
	t.Logf("account fields: %+v", accountFields)

	// Should contain "type" (the field being compared)
	assert.Contains(t, accountFields, "type", "type should be mapped from count_by condition")

	// Should NOT contain "cacc" (the quoted string value)
	assert.NotContains(t, accountFields, "cacc", "cacc should NOT be mapped - it's a value, not a field")
	assert.NotContains(t, accountFields, `"cacc"`, `"cacc" should NOT be mapped - it's a quoted value`)
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
