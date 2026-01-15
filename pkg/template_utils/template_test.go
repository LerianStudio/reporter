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

func TestRegexBlockIfOnPlaceholder(t *testing.T) {
	t.Run("extracts fields from simple if condition", func(t *testing.T) {
		template := `{% if midaz_transaction.status == "active" %}Active{% endif %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockIfOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "midaz_transaction")
	})

	t.Run("extracts fields from if with loop variable", func(t *testing.T) {
		template := `{% for item in midaz_transaction.items %}{% if item.active == true %}{{ item.name }}{% endif %}{% endfor %}`
		resultRegex := make(map[string]any)
		variableMap := map[string][]string{
			"item": {"midaz_transaction", "items"},
		}

		regexBlockIfOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "midaz_transaction")
	})

	t.Run("extracts fields from complex if expression", func(t *testing.T) {
		template := `{% if data.user.name and data.user.email %}Show user{% endif %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockIfOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "data")
	})

	t.Run("handles if with dash syntax", func(t *testing.T) {
		template := `{%- if report.status == "done" -%}Done{%- endif -%}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockIfOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "report")
	})

	t.Run("ignores single part paths", func(t *testing.T) {
		template := `{% if active %}Show{% endif %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockIfOnPlaceholder(template, resultRegex, variableMap)

		assert.Empty(t, resultRegex)
	})
}

func TestRegexBlockSetOnPlaceholder(t *testing.T) {
	t.Run("extracts fields from set statement", func(t *testing.T) {
		template := `{% set total = midaz_transaction.amount %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockSetOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "midaz_transaction")
	})

	t.Run("extracts fields from set with loop variable", func(t *testing.T) {
		template := `{% for item in orders.items %}{% set price = item.price %}{% endfor %}`
		resultRegex := make(map[string]any)
		variableMap := map[string][]string{
			"item": {"orders", "items"},
		}

		regexBlockSetOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "orders")
	})

	t.Run("extracts fields from set with complex expression", func(t *testing.T) {
		template := `{% set result = data.config.value + data.config.offset %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockSetOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "data")
	})

	t.Run("handles set with dash syntax", func(t *testing.T) {
		template := `{%- set name = user.profile.name -%}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockSetOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "user")
	})
}

func TestRegexBlockAggregationBlocksOnPlaceholder(t *testing.T) {
	t.Run("extracts fields from count_by aggregation", func(t *testing.T) {
		template := `{% count_by midaz_transaction.items if item.status == "active" %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockAggregationBlocksOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "midaz_transaction")
	})

	t.Run("extracts fields from sum_by aggregation", func(t *testing.T) {
		template := `{% sum_by midaz_transaction.balance if balance.type == "credit" %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockAggregationBlocksOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "midaz_transaction")
	})

	t.Run("extracts fields from avg_by aggregation", func(t *testing.T) {
		template := `{% avg_by orders.items if item.active == true %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockAggregationBlocksOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "orders")
	})

	t.Run("extracts fields from min_by aggregation", func(t *testing.T) {
		template := `{% min_by data.prices if price.valid == true %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockAggregationBlocksOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "data")
	})

	t.Run("extracts fields from max_by aggregation", func(t *testing.T) {
		template := `{% max_by report.values if value.active == true %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockAggregationBlocksOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "report")
	})

	t.Run("extracts fields from aggregation with by clause", func(t *testing.T) {
		template := `{% sum_by midaz_transaction.items by "category" if item.status == "active" %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockAggregationBlocksOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "midaz_transaction")
	})

	t.Run("handles aggregation with existing variable map", func(t *testing.T) {
		template := `{% count_by orders.items if item.type == "shipped" %}`
		resultRegex := make(map[string]any)
		variableMap := map[string][]string{
			"item": {"orders", "items"},
		}

		regexBlockAggregationBlocksOnPlaceholder(template, resultRegex, variableMap)

		assert.Contains(t, resultRegex, "orders")
	})

	t.Run("ignores invalid aggregation expressions", func(t *testing.T) {
		template := `{% count_by invalid %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockAggregationBlocksOnPlaceholder(template, resultRegex, variableMap)

		// Should not panic
		assert.NotNil(t, resultRegex)
	})
}

func TestRegexBlockForWithFilterOnPlaceholder(t *testing.T) {
	t.Run("extracts fields from for with filter", func(t *testing.T) {
		template := `{% for item in filter(midaz_transaction.items, "active", true) %}{{ item.name }}{% endfor %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockForWithFilterOnPlaceholder(resultRegex, variableMap, template)

		assert.Contains(t, variableMap, "item")
	})

	t.Run("extracts filter parameters", func(t *testing.T) {
		template := `{% for order in filter(data.orders, "status", "pending") %}{{ order.id }}{% endfor %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockForWithFilterOnPlaceholder(resultRegex, variableMap, template)

		assert.Contains(t, variableMap, "order")
		assert.Equal(t, []string{"data", "orders"}, variableMap["order"])
	})

	t.Run("handles filter with multiple parameters", func(t *testing.T) {
		template := `{% for user in filter(system.users, "role", "admin", user.department.name) %}{{ user.email }}{% endfor %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockForWithFilterOnPlaceholder(resultRegex, variableMap, template)

		assert.Contains(t, variableMap, "user")
	})

	t.Run("handles filter with existing variable map", func(t *testing.T) {
		template := `{% for item in filter(orders.items, "type", order.type) %}{{ item.price }}{% endfor %}`
		resultRegex := make(map[string]any)
		variableMap := map[string][]string{
			"order": {"data", "orders"},
		}

		regexBlockForWithFilterOnPlaceholder(resultRegex, variableMap, template)

		assert.Contains(t, variableMap, "item")
	})

	t.Run("handles filter with empty parameter", func(t *testing.T) {
		template := `{% for item in filter(data.items, , "value") %}{{ item.name }}{% endfor %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockForWithFilterOnPlaceholder(resultRegex, variableMap, template)

		// Should not panic
		assert.NotNil(t, resultRegex)
	})

	t.Run("handles filter with single part path parameter", func(t *testing.T) {
		template := `{% for item in filter(data.items, "field", simple_value) %}{{ item.name }}{% endfor %}`
		resultRegex := make(map[string]any)
		variableMap := make(map[string][]string)

		regexBlockForWithFilterOnPlaceholder(resultRegex, variableMap, template)

		assert.Contains(t, variableMap, "item")
	})
}

func TestRegexBlockWithOnPlaceholder(t *testing.T) {
	t.Run("extracts fields from with filter statement", func(t *testing.T) {
		template := `{% with filtered = filter(midaz_transaction.items, "active", true) %}{{ filtered }}{% endwith %}`
		variableMap := make(map[string][]string)

		result := regexBlockWithOnPlaceholder(variableMap, template)

		assert.Contains(t, variableMap, "filtered")
		assert.NotNil(t, result)
	})

	t.Run("extracts fields from simple with statement", func(t *testing.T) {
		template := `{% with data = report.summary %}{{ data.total }}{% endwith %}`
		variableMap := make(map[string][]string)

		result := regexBlockWithOnPlaceholder(variableMap, template)

		assert.Contains(t, variableMap, "data")
		assert.NotNil(t, result)
	})

	t.Run("handles with filter multiple parameters", func(t *testing.T) {
		template := `{% with items = filter(data.orders, "status", "pending", order.category) %}{{ items }}{% endwith %}`
		variableMap := map[string][]string{
			"order": {"system", "orders"},
		}

		result := regexBlockWithOnPlaceholder(variableMap, template)

		assert.Contains(t, variableMap, "items")
		assert.NotNil(t, result)
	})

	t.Run("handles with statement with dash syntax", func(t *testing.T) {
		template := `{%- with config = settings.database -%}{{ config.host }}{%- endwith -%}`
		variableMap := make(map[string][]string)

		result := regexBlockWithOnPlaceholder(variableMap, template)

		assert.Contains(t, variableMap, "config")
		assert.NotNil(t, result)
	})

	t.Run("handles with filter with empty parameters", func(t *testing.T) {
		template := `{% with items = filter(data.items, , "test") %}{{ items }}{% endwith %}`
		variableMap := make(map[string][]string)

		result := regexBlockWithOnPlaceholder(variableMap, template)

		// Should not panic
		assert.NotNil(t, result)
	})

	t.Run("handles with filter with single part path", func(t *testing.T) {
		template := `{% with items = filter(data.items, "field", simple) %}{{ items }}{% endwith %}`
		variableMap := make(map[string][]string)

		result := regexBlockWithOnPlaceholder(variableMap, template)

		assert.NotNil(t, result)
	})

	t.Run("handles multiple with statements", func(t *testing.T) {
		template := `{% with a = data.first %}{% with b = data.second %}{{ a }}{{ b }}{% endwith %}{% endwith %}`
		variableMap := make(map[string][]string)

		result := regexBlockWithOnPlaceholder(variableMap, template)

		assert.Contains(t, variableMap, "a")
		assert.Contains(t, variableMap, "b")
		assert.NotNil(t, result)
	})

	t.Run("handles with filter using loop variable", func(t *testing.T) {
		template := `{% with filtered = filter(data.items, "type", item.category) %}{{ filtered }}{% endwith %}`
		variableMap := map[string][]string{
			"item": {"orders", "items"},
		}

		result := regexBlockWithOnPlaceholder(variableMap, template)

		assert.Contains(t, variableMap, "filtered")
		assert.NotNil(t, result)
	})
}

func TestMappedFieldsOfTemplate_IfBlock(t *testing.T) {
	template := `{% if midaz_transaction.status == "active" %}
Status: {{ midaz_transaction.status }}
{% endif %}`

	result := MappedFieldsOfTemplate(template)

	assert.NotNil(t, result)
	assert.Contains(t, result, "midaz_transaction")
}

func TestMappedFieldsOfTemplate_SetBlock(t *testing.T) {
	template := `{% set total = midaz_transaction.amount %}
Total: {{ total }}`

	result := MappedFieldsOfTemplate(template)

	assert.NotNil(t, result)
	assert.Contains(t, result, "midaz_transaction")
}

func TestMappedFieldsOfTemplate_WithBlock(t *testing.T) {
	template := `{% with summary = report.data %}
{{ summary.total }}
{% endwith %}`

	result := MappedFieldsOfTemplate(template)

	assert.NotNil(t, result)
	assert.Contains(t, result, "report")
}

func TestMappedFieldsOfTemplate_AggregationBlock(t *testing.T) {
	template := `Total: {% count_by midaz_transaction.items if item.active == true %}`

	result := MappedFieldsOfTemplate(template)

	assert.NotNil(t, result)
	assert.Contains(t, result, "midaz_transaction")
}

func TestMappedFieldsOfTemplate_ComplexTemplate(t *testing.T) {
	template := `
{% for item in midaz_transaction.items %}
{% if item.status == "active" %}
{% set price = item.price %}
Name: {{ item.name }}
Price: {% calc item.price * 1.1 %}
{% endif %}
{% endfor %}
Count: {% count_by midaz_transaction.items if item.type == "credit" %}
`

	result := MappedFieldsOfTemplate(template)

	assert.NotNil(t, result)
	assert.Contains(t, result, "midaz_transaction")
}

func TestRegexBlockForOnPlaceholder(t *testing.T) {
	t.Run("extracts simple for loop variable", func(t *testing.T) {
		template := `{% for item in data.items %}{{ item.name }}{% endfor %}`

		result := regexBlockForOnPlaceholder(template)

		assert.Contains(t, result, "item")
		assert.Equal(t, []string{"data", "items"}, result["item"])
	})

	t.Run("extracts for loop with 3 level path", func(t *testing.T) {
		template := `{% for order in system.data.orders %}{{ order.id }}{% endfor %}`

		result := regexBlockForOnPlaceholder(template)

		assert.Contains(t, result, "order")
		assert.Equal(t, []string{"system", "data", "orders"}, result["order"])
	})

	t.Run("extracts for loop with single level path", func(t *testing.T) {
		template := `{% for item in items %}{{ item.name }}{% endfor %}`

		result := regexBlockForOnPlaceholder(template)

		assert.Contains(t, result, "item")
	})

	t.Run("handles multiple for loops", func(t *testing.T) {
		template := `{% for a in data.a %}{% for b in data.b %}{{ a }}{{ b }}{% endfor %}{% endfor %}`

		result := regexBlockForOnPlaceholder(template)

		assert.Contains(t, result, "a")
		assert.Contains(t, result, "b")
	})

	t.Run("handles for loop with dash syntax", func(t *testing.T) {
		template := `{%- for item in data.items -%}{{ item }}{%- endfor -%}`

		result := regexBlockForOnPlaceholder(template)

		assert.Contains(t, result, "item")
	})
}

func TestGetMimeType(t *testing.T) {
	tests := []struct {
		name         string
		outputFormat string
		expected     string
	}{
		{
			name:         "xml format",
			outputFormat: "xml",
			expected:     "application/xml",
		},
		{
			name:         "XML uppercase",
			outputFormat: "XML",
			expected:     "application/xml",
		},
		{
			name:         "html format",
			outputFormat: "html",
			expected:     "text/html",
		},
		{
			name:         "HTML uppercase",
			outputFormat: "HTML",
			expected:     "text/html",
		},
		{
			name:         "csv format",
			outputFormat: "csv",
			expected:     "text/csv",
		},
		{
			name:         "CSV uppercase",
			outputFormat: "CSV",
			expected:     "text/csv",
		},
		{
			name:         "txt format",
			outputFormat: "txt",
			expected:     "text/plain",
		},
		{
			name:         "TXT uppercase",
			outputFormat: "TXT",
			expected:     "text/plain",
		},
		{
			name:         "pdf format",
			outputFormat: "pdf",
			expected:     "application/pdf",
		},
		{
			name:         "PDF uppercase",
			outputFormat: "PDF",
			expected:     "application/pdf",
		},
		{
			name:         "unknown format",
			outputFormat: "unknown",
			expected:     "application/octet-stream",
		},
		{
			name:         "empty format",
			outputFormat: "",
			expected:     "application/octet-stream",
		},
		{
			name:         "json format (not defined)",
			outputFormat: "json",
			expected:     "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMimeType(tt.outputFormat)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateNoScriptTag(t *testing.T) {
	t.Run("returns nil for template without script tags", func(t *testing.T) {
		template := `<html><body><h1>Hello</h1></body></html>`
		err := ValidateNoScriptTag(template)
		assert.NoError(t, err)
	})

	t.Run("returns error for template with script opening tag", func(t *testing.T) {
		template := `<html><body><script>alert('xss')</script></body></html>`
		err := ValidateNoScriptTag(template)
		assert.Error(t, err)
	})

	t.Run("returns error for template with uppercase SCRIPT tag", func(t *testing.T) {
		template := `<html><body><SCRIPT>alert('xss')</SCRIPT></body></html>`
		err := ValidateNoScriptTag(template)
		assert.Error(t, err)
	})

	t.Run("returns error for template with mixed case Script tag", func(t *testing.T) {
		template := `<html><body><Script>alert('xss')</Script></body></html>`
		err := ValidateNoScriptTag(template)
		assert.Error(t, err)
	})

	t.Run("returns error for template with only opening script tag", func(t *testing.T) {
		template := `<html><body><script>some code</body></html>`
		err := ValidateNoScriptTag(template)
		assert.Error(t, err)
	})

	t.Run("returns error for template with only closing script tag", func(t *testing.T) {
		template := `<html><body>some code</script></body></html>`
		err := ValidateNoScriptTag(template)
		assert.Error(t, err)
	})

	t.Run("returns nil for template with script in text content", func(t *testing.T) {
		template := `<html><body><p>The word script is allowed</p></body></html>`
		err := ValidateNoScriptTag(template)
		assert.NoError(t, err)
	})

	t.Run("returns nil for empty template", func(t *testing.T) {
		template := ``
		err := ValidateNoScriptTag(template)
		assert.NoError(t, err)
	})

	t.Run("returns nil for plain text template", func(t *testing.T) {
		template := `This is just plain text without any HTML tags`
		err := ValidateNoScriptTag(template)
		assert.NoError(t, err)
	})
}
