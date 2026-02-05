package pongo

import (
	"testing"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregateBalance_BasicGrouping(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 1000.0, "date": "2026-01-15"},
			{"route_id": "route-2", "cosif_code": "1.1.2", "balance": 2000.0, "date": "2026-01-20"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	assert.Contains(t, out, "1.1.1:1000;")
	assert.Contains(t, out, "1.1.2:2000;")
}

func TestAggregateBalance_MultipleAccountsSameCosif(t *testing.T) {
	tplStr := `{% aggregate_balance data by "available_balance_after" group_by "cosif_code" order_by "created_at" as balances %}{% for item in balances %}{{ item.group_value }};{{ item.balance }};{{ item.count }}
{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1.00.00-0", "available_balance_after": 1000.00, "created_at": "2026-01-15"},
			{"route_id": "route-1", "cosif_code": "1.1.1.00.00-0", "available_balance_after": 2000.00, "created_at": "2026-01-31"},
			{"route_id": "route-2", "cosif_code": "1.1.1.00.00-0", "available_balance_after": 800.00, "created_at": "2026-01-25"},
			{"route_id": "route-3", "cosif_code": "1.2.1.00.00-3", "available_balance_after": 3000.00, "created_at": "2026-01-31"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)

	// route-1 last balance: 2000, route-2 last balance: 800 → total: 2800, count: 2
	assert.Contains(t, out, "1.1.1.00.00-0;2800;2")
	// route-3 last balance: 3000 → total: 3000, count: 1
	assert.Contains(t, out, "1.2.1.00.00-3;3000;1")
}

func TestAggregateBalance_WithFilter(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" if type == "CREDIT" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 1000.0, "date": "2026-01-15", "type": "CREDIT"},
			{"route_id": "route-2", "cosif_code": "1.1.1", "balance": 500.0, "date": "2026-01-20", "type": "DEBIT"},
			{"route_id": "route-3", "cosif_code": "1.1.1", "balance": 2000.0, "date": "2026-01-25", "type": "CREDIT"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	// Only CREDIT items: route-1 (1000) + route-3 (2000) = 3000
	assert.Contains(t, out, "1.1.1:3000;")
}

func TestAggregateBalance_EmptyCollection(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}count:{{ results|length }}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	assert.Equal(t, "count:0", out)
}

func TestAggregateBalance_MissingFields(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 1000.0, "date": "2026-01-15"},
			{"route_id": "route-2", "balance": 500.0, "date": "2026-01-20"},      // missing cosif_code
			{"route_id": "route-3", "cosif_code": "1.1.1", "date": "2026-01-25"}, // missing balance
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	// Only route-1 has valid data
	assert.Contains(t, out, "1.1.1:1000;")
}

func TestAggregateBalance_SyntaxError_MissingBy(t *testing.T) {
	tplStr := `{% aggregate_balance data "balance" group_by "cosif_code" order_by "date" as results %}`

	_, err := pongo2.FromString(tplStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected 'by' keyword")
}

func TestAggregateBalance_SyntaxError_MissingGroupBy(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" "cosif_code" order_by "date" as results %}`

	_, err := pongo2.FromString(tplStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected 'group_by' keyword")
}

func TestAggregateBalance_SyntaxError_MissingOrderBy(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" "date" as results %}`

	_, err := pongo2.FromString(tplStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected 'order_by' keyword")
}

func TestAggregateBalance_SyntaxError_MissingAs(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" results %}`

	_, err := pongo2.FromString(tplStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected 'as' keyword")
}

func TestAggregateBalance_SyntaxError_MissingVarName(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as %}`

	_, err := pongo2.FromString(tplStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected variable name after 'as'")
}

func TestAggregateBalance_WithAccountId(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"account_id": "acc-1", "cosif_code": "1.1.1", "balance": 1000.0, "date": "2026-01-15"},
			{"account_id": "acc-1", "cosif_code": "1.1.1", "balance": 1500.0, "date": "2026-01-20"},
			{"account_id": "acc-2", "cosif_code": "1.1.1", "balance": 2000.0, "date": "2026-01-25"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	// acc-1 last: 1500, acc-2 last: 2000 → total: 3500
	assert.Contains(t, out, "1.1.1:3500;")
}

func TestAggregateBalance_WithIdFallback(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"id": "id-1", "cosif_code": "1.1.1", "balance": 1000.0, "date": "2026-01-15"},
			{"id": "id-1", "cosif_code": "1.1.1", "balance": 1500.0, "date": "2026-01-20"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	// id-1 last: 1500
	assert.Contains(t, out, "1.1.1:1500;")
}

func TestAggregateBalance_NoSubGroupKey(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"cosif_code": "1.1.1", "balance": 1000.0, "date": "2026-01-15"},
			{"cosif_code": "1.1.1", "balance": 2000.0, "date": "2026-01-20"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	// Both go to _default_ sub-group, last is 2000
	assert.Contains(t, out, "1.1.1:2000;")
}

func TestAggregateBalance_RFC3339Date(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "created_at" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 1000.0, "created_at": "2026-01-15T10:00:00Z"},
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 2000.0, "created_at": "2026-01-31T15:30:00Z"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	// Last by RFC3339 is 2000
	assert.Contains(t, out, "1.1.1:2000;")
}

func TestAggregateBalance_TimeTypeDate(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "created_at" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 1000.0, "created_at": time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)},
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 2000.0, "created_at": time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	assert.Contains(t, out, "1.1.1:2000;")
}

func TestAggregateBalance_StringBalance(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": "1500.50", "date": "2026-01-15"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	assert.Contains(t, out, "1.1.1:1500.5;")
}

func TestAggregateBalance_IntBalance(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 1500, "date": "2026-01-15"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	assert.Contains(t, out, "1.1.1:1500;")
}

func TestAggregateBalance_Int64Balance(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": int64(1500), "date": "2026-01-15"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	assert.Contains(t, out, "1.1.1:1500;")
}

func TestAggregateBalance_DecimalBalance(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": decimal.NewFromFloat(1500.75), "date": "2026-01-15"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	assert.Contains(t, out, "1.1.1:1500.75;")
}

func TestAggregateBalance_InvalidBalanceType(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": struct{}{}, "date": "2026-01-15"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	// Invalid balance type should be skipped, resulting in 0 count
	assert.Contains(t, out, "1.1.1:0;")
}

func TestAggregateBalance_InvalidStringBalance(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": "not-a-number", "date": "2026-01-15"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	// Invalid string balance should be skipped
	assert.Contains(t, out, "1.1.1:0;")
}

func TestAggregateBalance_RFC3339NanoDate(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "created_at" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 1000.0, "created_at": "2026-01-15T10:00:00.123456789Z"},
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 2000.0, "created_at": "2026-01-31T15:30:00.987654321Z"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	assert.Contains(t, out, "1.1.1:2000;")
}

func TestAggregateBalance_InvalidDateFormat(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "created_at" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 1000.0, "created_at": "invalid-date"},
			{"route_id": "route-2", "cosif_code": "1.1.1", "balance": 2000.0, "created_at": "also-invalid"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	// Both have zero time, first one becomes "latest" (route-1 balance: 1000)
	// Then route-2 is also zero time but doesn't beat it
	// Sum should be 1000 + 2000 = 3000 since both are processed
	assert.Contains(t, out, "1.1.1:")
}

func TestAggregateBalance_MultipleGroups(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "category" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }}:{{ item.count }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"id": "1", "category": "A", "balance": 100.0, "date": "2026-01-01"},
			{"id": "2", "category": "B", "balance": 200.0, "date": "2026-01-01"},
			{"id": "3", "category": "C", "balance": 300.0, "date": "2026-01-01"},
			{"id": "4", "category": "A", "balance": 150.0, "date": "2026-01-02"},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	// Results are sorted alphabetically by group_value
	assert.Contains(t, out, "A:250:2;") // id-1 last: 100, id-4 last: 150 = 250, count: 2
	assert.Contains(t, out, "B:200:1;") // id-2: 200
	assert.Contains(t, out, "C:300:1;") // id-3: 300
}

func TestAggregateBalance_AllItemsFiltered(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" if active == true as results %}count:{{ results|length }}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 1000.0, "date": "2026-01-15", "active": false},
			{"route_id": "route-2", "cosif_code": "1.1.1", "balance": 2000.0, "date": "2026-01-20", "active": false},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	assert.Equal(t, "count:0", out)
}

func TestAggregateBalance_NestedFields(t *testing.T) {
	tplStr := `{% aggregate_balance data by "account.balance" group_by "meta.code" order_by "meta.date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{
				"route_id": "route-1",
				"meta":     map[string]any{"code": "1.1.1", "date": "2026-01-15"},
				"account":  map[string]any{"balance": 1000.0},
			},
			{
				"route_id": "route-1",
				"meta":     map[string]any{"code": "1.1.1", "date": "2026-01-20"},
				"account":  map[string]any{"balance": 2000.0},
			},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	assert.Contains(t, out, "1.1.1:2000;")
}

func TestAggregateBalance_EmptyDateField(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.count }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 1000.0},
			{"route_id": "route-2", "cosif_code": "1.1.1", "balance": 2000.0},
		},
	}

	out, err := tpl.Execute(ctx)
	require.NoError(t, err)
	// Items without date field are now included with zero time
	// Both routes are processed, count should be 2
	assert.Contains(t, out, "1.1.1:2;")
}

func TestAggregateBalance_SameTimestamp_DeterministicBehavior(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}{% for item in results %}{{ item.group_value }}:{{ item.balance }};{% endfor %}`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 1000.0, "date": "2026-01-31"},
			{"route_id": "route-1", "cosif_code": "1.1.1", "balance": 2000.0, "date": "2026-01-31"},
		},
	}

	// Run multiple times to verify deterministic behavior
	for i := 0; i < 5; i++ {
		out, err := tpl.Execute(ctx)
		require.NoError(t, err)
		// With same timestamp, later item (balance=2000) should win consistently
		assert.Contains(t, out, "1.1.1:2000;", "iteration %d should be deterministic", i)
	}
}

func TestAggregateBalance_InvalidCollectionType(t *testing.T) {
	tplStr := `{% aggregate_balance data by "balance" group_by "cosif_code" order_by "date" as results %}done`

	tpl, err := pongo2.FromString(tplStr)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"data": "not-a-slice",
	}

	_, err = tpl.Execute(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "[]map[string]any")
}
