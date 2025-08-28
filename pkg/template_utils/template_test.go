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
