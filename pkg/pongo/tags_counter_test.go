package pongo

import (
	"testing"

	"github.com/flosch/pongo2/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCounterTag_BasicIncrement(t *testing.T) {
	ResetCounters()

	tpl, err := pongo2.FromString(`{% counter "1100" %}{% counter "1100" %}{% counter "1100" %}{% counter_show "1100" %}`)
	require.NoError(t, err)

	result, err := tpl.Execute(nil)
	require.NoError(t, err)

	assert.Equal(t, "3", result)
}

func TestCounterTag_MultipleCounters(t *testing.T) {
	ResetCounters()

	tpl, err := pongo2.FromString(`{% counter "1100" %}{% counter "1100" %}{% counter "1101" %}{% counter "1101" %}{% counter "1101" %}{% counter_show "1100" %}-{% counter_show "1101" %}`)
	require.NoError(t, err)

	result, err := tpl.Execute(nil)
	require.NoError(t, err)

	assert.Equal(t, "2-3", result)
}

func TestCounterTag_SumMultipleCounters(t *testing.T) {
	ResetCounters()

	tpl, err := pongo2.FromString(`{% counter "1100" %}{% counter "1100" %}{% counter "1101" %}{% counter "1101" %}{% counter "1101" %}{% counter_show "1100" "1101" %}`)
	require.NoError(t, err)

	result, err := tpl.Execute(nil)
	require.NoError(t, err)

	assert.Equal(t, "5", result) // 2 + 3 = 5
}

func TestCounterTag_InLoop(t *testing.T) {
	ResetCounters()

	tpl, err := pongo2.FromString(`{% for i in items %}|1100|{% counter "1100" %}{{ i.name }}|
{% endfor %}Total: {% counter_show "1100" %}`)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"items": []map[string]any{
			{"name": "Item1"},
			{"name": "Item2"},
			{"name": "Item3"},
		},
	}

	result, err := tpl.Execute(ctx)
	require.NoError(t, err)

	expected := `|1100|Item1|
|1100|Item2|
|1100|Item3|
Total: 3`
	assert.Equal(t, expected, result)
}

func TestCounterTag_NestedLoops(t *testing.T) {
	ResetCounters()

	tpl, err := pongo2.FromString(`{% for acc in accounts %}|1100|{% counter "1100" %}{{ acc.id }}|
{% for det in acc.details %}|1101|{% counter "1101" %}{{ det.value }}|
{% endfor %}{% endfor %}|9900|1100|{% counter_show "1100" %}|
|9900|1101|{% counter_show "1101" %}|
|9900|TOTAL|{% counter_show "1100" "1101" %}|`)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"accounts": []map[string]any{
			{
				"id": "ACC1",
				"details": []map[string]any{
					{"value": "D1"},
					{"value": "D2"},
				},
			},
			{
				"id": "ACC2",
				"details": []map[string]any{
					{"value": "D3"},
				},
			},
		},
	}

	result, err := tpl.Execute(ctx)
	require.NoError(t, err)

	assert.Contains(t, result, "|9900|1100|2|")
	assert.Contains(t, result, "|9900|1101|3|")
	assert.Contains(t, result, "|9900|TOTAL|5|")
}

func TestCounterTag_ZeroCounter(t *testing.T) {
	ResetCounters()

	tpl, err := pongo2.FromString(`{% counter_show "nonexistent" %}`)
	require.NoError(t, err)

	result, err := tpl.Execute(nil)
	require.NoError(t, err)

	assert.Equal(t, "0", result)
}

func TestCounterTag_ResetBetweenRenders(t *testing.T) {
	ResetCounters()

	// First render
	tpl, err := pongo2.FromString(`{% counter "test" %}{% counter "test" %}{% counter_show "test" %}`)
	require.NoError(t, err)

	result1, err := tpl.Execute(nil)
	require.NoError(t, err)
	assert.Equal(t, "2", result1)

	// Reset and second render
	ResetCounters()

	result2, err := tpl.Execute(nil)
	require.NoError(t, err)
	assert.Equal(t, "2", result2)
}

func TestCounterTag_DIMPExample(t *testing.T) {
	ResetCounters()

	// Simulates a DIMP report structure
	tpl, err := pongo2.FromString(`|0000|12345678901234|EMPRESA|
{% for acc in accounts %}|1100|{% counter "1100" %}SP|{{ acc.id }}|{{ acc.alias }}|
{% endfor %}{% for tx in transactions %}|1102|{% counter "1102" %}{{ tx.id }}|{{ tx.amount }}|
{% endfor %}|TOTAL_SP|1500.00|
|9900|1100|{% counter_show "1100" %}|
|9900|1102|{% counter_show "1102" %}|`)
	require.NoError(t, err)

	ctx := pongo2.Context{
		"accounts": []map[string]any{
			{"id": "ACC001", "alias": "account/123"},
			{"id": "ACC002", "alias": "account/456"},
			{"id": "ACC003", "alias": "account/789"},
		},
		"transactions": []map[string]any{
			{"id": "TX001", "amount": "500.00"},
			{"id": "TX002", "amount": "1000.00"},
		},
	}

	result, err := tpl.Execute(ctx)
	require.NoError(t, err)

	assert.Contains(t, result, "|9900|1100|3|")
	assert.Contains(t, result, "|9900|1102|2|")
}

func TestCounterTag_SumThreeCounters(t *testing.T) {
	ResetCounters()

	tpl, err := pongo2.FromString(`{% counter "A" %}{% counter "A" %}{% counter "B" %}{% counter "B" %}{% counter "B" %}{% counter "C" %}{% counter_show "A" "B" "C" %}`)
	require.NoError(t, err)

	result, err := tpl.Execute(nil)
	require.NoError(t, err)

	assert.Equal(t, "6", result) // 2 + 3 + 1 = 6
}

func TestGetCounter(t *testing.T) {
	ResetCounters()

	tpl, err := pongo2.FromString(`{% counter "test" %}{% counter "test" %}{% counter "test" %}`)
	require.NoError(t, err)

	_, err = tpl.Execute(nil)
	require.NoError(t, err)

	assert.Equal(t, 3, GetCounter("test"))
	assert.Equal(t, 0, GetCounter("nonexistent"))
}