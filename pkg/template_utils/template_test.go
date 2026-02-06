// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package template_utils

import (
	"testing"

	"github.com/LerianStudio/reporter/pkg/constant"
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

func TestParseDatabaseReference(t *testing.T) {
	tests := []struct {
		name           string
		ref            string
		wantDatabase   string
		wantSchema     string
		wantTable      string
		wantErr        bool
		wantErrContain string
	}{
		{
			name:         "legacy format - database.table",
			ref:          "midaz_onboarding.account",
			wantDatabase: "midaz_onboarding",
			wantSchema:   "",
			wantTable:    "account",
			wantErr:      false,
		},
		{
			name:         "new format - database:schema.table",
			ref:          "external_db:sales.orders",
			wantDatabase: "external_db",
			wantSchema:   "sales",
			wantTable:    "orders",
			wantErr:      false,
		},
		{
			name:         "new format with public schema",
			ref:          "midaz:public.users",
			wantDatabase: "midaz",
			wantSchema:   "public",
			wantTable:    "users",
			wantErr:      false,
		},
		{
			name:           "invalid format - no separator",
			ref:            "invalidformat",
			wantErr:        true,
			wantErrContain: "invalid format",
		},
		{
			name:           "invalid format - colon but no dot after",
			ref:            "database:schematable",
			wantErr:        true,
			wantErrContain: "expected schema.table",
		},
		{
			name:         "legacy format with underscore in names",
			ref:          "midaz_transaction.balance_entry",
			wantDatabase: "midaz_transaction",
			wantSchema:   "",
			wantTable:    "balance_entry",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, schema, table, err := ParseDatabaseReference(tt.ref)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrContain != "" {
					assert.Contains(t, err.Error(), tt.wantErrContain)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantDatabase, database, "database mismatch")
			assert.Equal(t, tt.wantSchema, schema, "schema mismatch")
			assert.Equal(t, tt.wantTable, table, "table mismatch")
		})
	}
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

func TestCleanPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "legacy format - database.table.field",
			path:     "midaz_transaction.balance.amount",
			expected: []string{"midaz_transaction", "balance", "amount"},
		},
		{
			name:     "legacy format with array index",
			path:     "midaz_transaction.balance.0.amount",
			expected: []string{"midaz_transaction", "balance", "amount"},
		},
		{
			name:     "schema format - database:schema.table.field",
			path:     "external_db:sales.orders.amount",
			expected: []string{"external_db", "sales__orders", "amount"},
		},
		{
			name:     "schema format with array index",
			path:     "external_db:sales.orders.0.amount",
			expected: []string{"external_db", "sales__orders", "amount"},
		},
		{
			name:     "schema format - public schema",
			path:     "midaz:public.users.name",
			expected: []string{"midaz", "public__users", "name"},
		},
		{
			name:     "simple two parts",
			path:     "database.table",
			expected: []string{"database", "table"},
		},
		{
			name:     "schema format - two parts (database:schema.table)",
			path:     "database:schema.table",
			expected: []string{"database", "schema__table"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMappedFieldsOfTemplate_SchemaFormat(t *testing.T) {
	template := `{% for tx in external_db:sales.orders %}
Transaction ID: {{ tx.id }}
Amount: {{ tx.amount }}
Status: {{ tx.status }}
{% endfor %}`

	result := MappedFieldsOfTemplate(template)

	assert.NotNil(t, result)
	assert.Contains(t, result, "external_db")

	externalDB := result["external_db"]
	assert.Contains(t, externalDB, "sales__orders")

	txFields := externalDB["sales__orders"]
	assert.Contains(t, txFields, "id")
	assert.Contains(t, txFields, "amount")
	assert.Contains(t, txFields, "status")
}

func TestMappedFieldsOfTemplate_MixedFormats(t *testing.T) {
	// Template uses original syntax with dot (database:schema.table)
	template := `{% for acc in midaz_onboarding.account %}
Account: {{ acc.alias }}
{% endfor %}

{% for tx in external_db:sales.orders %}
Amount: {{ tx.amount }}
{% endfor %}`

	result := MappedFieldsOfTemplate(template)

	// Check legacy format
	assert.Contains(t, result, "midaz_onboarding")
	assert.Contains(t, result["midaz_onboarding"], "account")
	assert.Contains(t, result["midaz_onboarding"]["account"], "alias")

	// Check schema format (schema.table becomes schema__table internally)
	assert.Contains(t, result, "external_db")
	assert.Contains(t, result["external_db"], "sales__orders")
	assert.Contains(t, result["external_db"]["sales__orders"], "amount")
}

func TestMappedFieldsOfTemplate_ExplicitSchemaCalcTag(t *testing.T) {
	template := `{% for tx in external_db:sales.orders %}
Amount: {{ tx.amount }}
{% endfor %}
Total: {% calc external_db:sales.orders.0.amount + external_db:sales.orders.1.amount %}`

	result := MappedFieldsOfTemplate(template)

	assert.Contains(t, result, "external_db")
	assert.Contains(t, result["external_db"], "sales__orders")
	assert.Contains(t, result["external_db"]["sales__orders"], "amount")
}

func TestMappedFieldsOfTemplate_ExplicitSchemaIfTag(t *testing.T) {
	template := `{% if external_db:sales.orders.0.status == "completed" %}
Completed: {{ external_db:sales.orders.0.amount }}
{% endif %}`

	result := MappedFieldsOfTemplate(template)

	assert.Contains(t, result, "external_db")
	assert.Contains(t, result["external_db"], "sales__orders")
	assert.Contains(t, result["external_db"]["sales__orders"], "status")
	assert.Contains(t, result["external_db"]["sales__orders"], "amount")
}

func TestExtractIfFromExpression_ExplicitSchemaSyntax(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected []string
	}{
		{
			name:     "explicit schema path",
			expr:     "external_db:sales.orders.0.amount",
			expected: []string{"external_db:sales.orders.0.amount"},
		},
		{
			name:     "explicit schema in comparison",
			expr:     "external_db:sales.orders.0.status == 'completed'",
			expected: []string{"external_db:sales.orders.0.status"},
		},
		{
			name:     "mixed formats",
			expr:     "external_db:sales.orders.0.amount + midaz.balance.0.value",
			expected: []string{"external_db:sales.orders.0.amount", "midaz.balance.value"},
		},
		{
			name:     "legacy format preserved",
			expr:     "midaz_transaction.balance.0.amount",
			expected: []string{"midaz_transaction.balance.amount"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractIfFromExpression(tt.expr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFieldsFromExpression_ExplicitSchemaSyntax(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected []string
	}{
		{
			name:     "explicit schema path",
			expr:     "external_db:sales.orders.0.amount",
			expected: []string{"external_db:sales.orders.0.amount"},
		},
		{
			name:     "explicit schema with filter",
			expr:     `external_db:sales.orders|where:"status:completed"`,
			expected: []string{"external_db:sales.orders"},
		},
		{
			name:     "legacy format preserved",
			expr:     "midaz_transaction.balance.0.amount",
			expected: []string{"midaz_transaction.balance.0.amount"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFieldsFromExpression(tt.expr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMappedFieldsOfTemplate_AllTagsWithExplicitSchema(t *testing.T) {
	// Test all template tags with explicit schema syntax
	template := `
{# For loop with schema #}
{% for tx in external_db:sales.orders %}
	ID: {{ tx.id }}
	Amount: {{ tx.amount }}
{% endfor %}

{# Variable expression with schema #}
First Order: {{ external_db:sales.orders.0.reference_id }}

{# Calc tag with schema #}
Total: {% calc external_db:sales.orders.0.amount + external_db:sales.orders.1.amount %}

{# If tag with schema #}
{% if external_db:sales.orders.0.status == "completed" %}
Completed!
{% endif %}

{# Set tag with schema #}
{% set total = external_db:sales.orders.0.fee %}

{# For loop with DIMP filter and schema #}
{% for tx in external_db:sales.orders|where:"status:completed" %}
	{{ tx.description }}
{% endfor %}

{# Sum_by aggregation with schema #}
{% sum_by external_db:sales.orders.amount if tx.status == "completed" %}

{# With tag and filter function #}
{% with filtered = filter(external_db:sales.orders, "status", "completed") %}
	{{ filtered }}
{% endwith %}
`

	result := MappedFieldsOfTemplate(template)

	// Verify all fields are extracted for the explicit schema datasource
	// Note: schema.table becomes schema__table for Pongo2 compatibility
	assert.Contains(t, result, "external_db", "Should contain external_db datasource")
	assert.Contains(t, result["external_db"], "sales__orders", "Should contain sales__orders table")

	orderFields := result["external_db"]["sales__orders"]

	expectedFields := []string{"id", "amount", "reference_id", "status", "fee", "description"}
	for _, field := range expectedFields {
		assert.Contains(t, orderFields, field, "Field %s should be mapped", field)
	}
}

func TestMappedFieldsOfTemplate_MixedLegacyAndSchemaFormats(t *testing.T) {
	// Template uses original syntax with dot (database:schema.table)
	template := `
{# Legacy format #}
{% for acc in midaz_onboarding.account %}
	{{ acc.alias }}
{% endfor %}

{# Explicit schema format - user writes schema.table with dot #}
{% for tx in external_db:sales.orders %}
	{{ tx.amount }}
{% endfor %}

{# Both in same calc expression #}
{% calc midaz_onboarding.account.0.balance + external_db:sales.orders.0.amount %}
`

	result := MappedFieldsOfTemplate(template)

	// Check legacy format
	assert.Contains(t, result, "midaz_onboarding")
	assert.Contains(t, result["midaz_onboarding"], "account")
	assert.Contains(t, result["midaz_onboarding"]["account"], "alias")
	assert.Contains(t, result["midaz_onboarding"]["account"], "balance")

	// Check schema format (schema.table becomes schema__table internally)
	assert.Contains(t, result, "external_db")
	assert.Contains(t, result["external_db"], "sales__orders")
	assert.Contains(t, result["external_db"]["sales__orders"], "amount")
}

func TestMappedFieldsOfTemplate_AggregationNestedFields(t *testing.T) {
	// Test that nested JSONB field paths in aggregation "by" clauses
	// are correctly treated as field paths, not datasource references
	template := `
{# Aggregation with nested JSONB field in "by" clause #}
<ValorReceita>{% sum_by external_db:analytics.transfers by "fee_charge.totalAmount" if status == "COMPLETED" %}</ValorReceita>

{# Another aggregation with deeply nested field #}
<TotalFees>{% sum_by external_db:analytics.transactions by "metadata.fees.amount" if type == "PAYMENT" %}</TotalFees>

{# Count with nested field #}
<Count>{% count_by datasource:public.orders by "customer.address.city" if active == "true" %}</Count>

{# Avg with nested field #}
<Average>{% avg_by datasource:public.products by "pricing.discount.percentage" if available == "true" %}</Average>

{# Simple aggregation without nested field (should still work) #}
<SimpleSum>{% sum_by external_db:analytics.sales by "amount" if region == "SOUTH" %}</SimpleSum>
`

	result := MappedFieldsOfTemplate(template)

	t.Logf("Result: %+v", result)

	// Check external_db datasource
	assert.Contains(t, result, "external_db")

	// analytics__transfers should contain the nested field path as-is
	assert.Contains(t, result["external_db"], "analytics__transfers")
	transfersFields := result["external_db"]["analytics__transfers"]
	assert.Contains(t, transfersFields, "fee_charge.totalAmount", "Nested field 'fee_charge.totalAmount' should be preserved as-is")
	assert.Contains(t, transfersFields, "status", "Condition field 'status' should be mapped")

	// analytics__transactions should contain deeply nested field path
	assert.Contains(t, result["external_db"], "analytics__transactions")
	transactionsFields := result["external_db"]["analytics__transactions"]
	assert.Contains(t, transactionsFields, "metadata.fees.amount", "Deeply nested field 'metadata.fees.amount' should be preserved as-is")
	assert.Contains(t, transactionsFields, "type", "Condition field 'type' should be mapped")

	// analytics__sales should contain simple field
	assert.Contains(t, result["external_db"], "analytics__sales")
	salesFields := result["external_db"]["analytics__sales"]
	assert.Contains(t, salesFields, "amount", "Simple field 'amount' should be mapped")
	assert.Contains(t, salesFields, "region", "Condition field 'region' should be mapped")

	// Check datasource
	assert.Contains(t, result, "datasource")

	// public__orders should contain nested field
	assert.Contains(t, result["datasource"], "public__orders")
	ordersFields := result["datasource"]["public__orders"]
	assert.Contains(t, ordersFields, "customer.address.city", "Nested field 'customer.address.city' should be preserved as-is")
	assert.Contains(t, ordersFields, "active", "Condition field 'active' should be mapped")

	// public__products should contain nested field
	assert.Contains(t, result["datasource"], "public__products")
	productsFields := result["datasource"]["public__products"]
	assert.Contains(t, productsFields, "pricing.discount.percentage", "Nested field 'pricing.discount.percentage' should be preserved as-is")
	assert.Contains(t, productsFields, "available", "Condition field 'available' should be mapped")
}

func TestMappedFieldsOfTemplate_AggregationNestedFieldsNotTreatedAsDatasource(t *testing.T) {
	// Regression test: ensure nested field paths like "fee_charge.totalAmount"
	// are NOT incorrectly split and treated as datasource references
	template := `{% sum_by db:schema.table by "fee_charge.totalAmount" if status == "COMPLETED" %}`

	result := MappedFieldsOfTemplate(template)

	t.Logf("Result: %+v", result)

	// Should have "db" as datasource
	assert.Contains(t, result, "db")

	// Should have "schema__table" as collection
	assert.Contains(t, result["db"], "schema__table")

	// Should NOT have "fee_charge" as a separate datasource or collection
	// This was the bug: fee_charge.totalAmount was being split and fee_charge
	// was treated as a datasource
	_, hasFeeCharge := result["fee_charge"]
	assert.False(t, hasFeeCharge, "fee_charge should NOT be treated as a datasource")

	_, hasFeeChargeInDb := result["db"]["fee_charge"]
	assert.False(t, hasFeeChargeInDb, "fee_charge should NOT be treated as a collection under db")

	// The field should be preserved as-is in the correct collection
	tableFields := result["db"]["schema__table"]
	assert.Contains(t, tableFields, "fee_charge.totalAmount", "Nested field should be preserved as complete path")
}

func TestMappedFieldsOfTemplate_AggregationCompoundConditions(t *testing.T) {
	// Test that compound conditions with "and" extract all field names
	template := `{% sum_by db:schema.transfers by "amount" if transfer_type == "CASHIN" and destination_person_type == "NATURAL_PERSON" and status == "COMPLETED" %}`

	result := MappedFieldsOfTemplate(template)

	t.Logf("Result: %+v", result)

	// Should have "db" as datasource
	assert.Contains(t, result, "db")

	// Should have "schema__transfers" as collection
	assert.Contains(t, result["db"], "schema__transfers")

	// All fields from the compound condition should be extracted
	tableFields := result["db"]["schema__transfers"]
	assert.Contains(t, tableFields, "amount", "Field 'amount' from 'by' clause should be mapped")
	assert.Contains(t, tableFields, "transfer_type", "Field 'transfer_type' from first condition should be mapped")
	assert.Contains(t, tableFields, "destination_person_type", "Field 'destination_person_type' from second condition should be mapped")
	assert.Contains(t, tableFields, "status", "Field 'status' from third condition should be mapped")
}

func TestMappedFieldsOfTemplate_AggregationNestedJSONBWithCompoundConditions(t *testing.T) {
	// Regression test: nested JSONB field path in "by" clause combined with compound conditions
	// This was causing TPL-0031 error because fee_charge.totalAmount was incorrectly treated as a datasource
	template := `{% sum_by sales_db:payment.transfers by "fee_charge.totalAmount" if transfer_type == "CASHIN" and destination_person_type == "NATURAL_PERSON" and status == "COMPLETED" %}`

	result := MappedFieldsOfTemplate(template)

	t.Logf("Result: %+v", result)

	// Should have "sales_db" as datasource
	assert.Contains(t, result, "sales_db")

	// Should have "payment__transfers" as collection
	assert.Contains(t, result["sales_db"], "payment__transfers")

	// Should NOT have "fee_charge" as a separate datasource
	_, hasFeeCharge := result["fee_charge"]
	assert.False(t, hasFeeCharge, "fee_charge should NOT be treated as a datasource")

	// The nested field should be preserved as-is
	tableFields := result["sales_db"]["payment__transfers"]
	assert.Contains(t, tableFields, "fee_charge.totalAmount", "Nested JSONB field should be preserved as complete path")
	assert.Contains(t, tableFields, "transfer_type", "Condition field should be mapped")
	assert.Contains(t, tableFields, "destination_person_type", "Condition field should be mapped")
	assert.Contains(t, tableFields, "status", "Condition field should be mapped")
}

func TestExtractFieldsFromConditions(t *testing.T) {
	tests := []struct {
		name       string
		conditions string
		expected   []string
	}{
		{
			name:       "single_condition",
			conditions: `transfer_type == "CASHIN"`,
			expected:   []string{"transfer_type", `"CASHIN"`},
		},
		{
			name:       "two_conditions",
			conditions: `transfer_type == "CASHIN" and status == "COMPLETED"`,
			expected:   []string{"transfer_type", `"CASHIN"`, "status", `"COMPLETED"`},
		},
		{
			name:       "three_conditions",
			conditions: `transfer_type == "CASHIN" and destination_person_type == "NATURAL_PERSON" and status == "COMPLETED"`,
			expected:   []string{"transfer_type", `"CASHIN"`, "destination_person_type", `"NATURAL_PERSON"`, "status", `"COMPLETED"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFieldsFromConditions(tt.conditions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateNoScriptTag(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		expectError bool
	}{
		// Valid templates (should pass)
		{
			name:        "Valid - Simple Pongo2 template",
			template:    `{% for item in db.table %}{{ item.name }}{% endfor %}`,
			expectError: false,
		},
		{
			name:        "Valid - Template with 'on' in text content",
			template:    `<div>organization</div><p>connection</p>`,
			expectError: false,
		},
		{
			name:        "Valid - Template with 'on' in variable names",
			template:    `{{ person.name }} {{ common_field }}`,
			expectError: false,
		},
		{
			name:        "Valid - HTML without XSS vectors",
			template:    `<html><body><h1>Report</h1><p>{{ data.content }}</p></body></html>`,
			expectError: false,
		},

		// Invalid templates - Script tags (original vulnerability)
		{
			name:        "Invalid - Exact <script> tag",
			template:    `<html><script>alert('x')</script></html>`,
			expectError: true,
		},
		{
			name:        "Invalid - Script tag with src attribute",
			template:    `<script src="evil.js"></script>`,
			expectError: true,
		},
		{
			name:        "Invalid - Script tag with type attribute",
			template:    `<script type="text/javascript">alert(1)</script>`,
			expectError: true,
		},
		{
			name:        "Invalid - Script tag with multiple attributes",
			template:    `<script type="text/javascript" src="evil.js" defer></script>`,
			expectError: true,
		},
		{
			name:        "Invalid - Script tag case insensitive",
			template:    `<SCRIPT>alert(1)</SCRIPT>`,
			expectError: true,
		},
		{
			name:        "Invalid - Script tag mixed case",
			template:    `<ScRiPt>alert(1)</sCrIpT>`,
			expectError: true,
		},

		// Invalid templates - Inline event handlers (XSS bypasses)
		{
			name:        "Invalid - img onerror XSS",
			template:    `<img src="x" onerror="alert(1)">`,
			expectError: true,
		},
		{
			name:        "Invalid - svg onload XSS",
			template:    `<svg onload="alert(1)">`,
			expectError: true,
		},
		{
			name:        "Invalid - div onclick XSS",
			template:    `<div onclick="malicious()">Click me</div>`,
			expectError: true,
		},
		{
			name:        "Invalid - body onload XSS",
			template:    `<body onload="stealData()">`,
			expectError: true,
		},
		{
			name:        "Invalid - input onfocus XSS",
			template:    `<input type="text" onfocus="alert(1)">`,
			expectError: true,
		},
		{
			name:        "Invalid - a onmouseover XSS",
			template:    `<a href="#" onmouseover="alert(1)">Hover me</a>`,
			expectError: true,
		},
		{
			name:        "Invalid - iframe onload XSS",
			template:    `<iframe src="page.html" onload="stealCookies()"></iframe>`,
			expectError: true,
		},
		{
			name:        "Invalid - Event handler case insensitive",
			template:    `<img OnError="alert(1)">`,
			expectError: true,
		},
		{
			name:        "Invalid - Event handler with spaces",
			template:    `<img onerror = "alert(1)">`,
			expectError: true,
		},
		{
			name:        "Invalid - Multiple event handlers",
			template:    `<div onclick="x()" onmouseover="y()">Content</div>`,
			expectError: true,
		},

		// Edge cases
		{
			name:        "Invalid - Script tag at end of file",
			template:    `<html><body>content</body><script>alert(1)</script>`,
			expectError: true,
		},
		{
			name:        "Invalid - Multiple script tags",
			template:    `<script>x()</script><div>content</div><script>y()</script>`,
			expectError: true,
		},
		{
			name:        "Invalid - Script tag with newlines",
			template:    "<script>\nalert(1)\n</script>",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNoScriptTag(tt.template)

			if tt.expectError {
				assert.Error(t, err, "Expected error for template: %s", tt.template)
				assert.Equal(t, constant.ErrScriptTagDetected, err, "Should return ErrScriptTagDetected")
			} else {
				assert.NoError(t, err, "Expected no error for template: %s", tt.template)
			}
		})
	}
}
