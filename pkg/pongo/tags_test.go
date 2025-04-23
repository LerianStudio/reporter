package pongo

import (
	"testing"

	"github.com/flosch/pongo2/v6"
	"github.com/stretchr/testify/assert"
)

func TestSumByTag(t *testing.T) {
	tplStr := `{% sum_by data by "amount" scale 2 %}`
	tpl, err := pongo2.FromString(tplStr)
	assert.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"amount": 1000},
			{"amount": 2500},
			{"amount": 1500},
		},
	}

	out, err := tpl.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "50.00", out)
}

func TestCountByTagWithFilter(t *testing.T) {
	tplStr := `{% count_by data if amount > 1000 %}`
	tpl, err := pongo2.FromString(tplStr)
	assert.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"amount": 1000},
			{"amount": 2500},
			{"amount": 1500},
		},
	}

	out, err := tpl.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "2", out)
}

func TestAvgByTag(t *testing.T) {
	tplStr := `{% avg_by data by "amount" scale 1 %}`
	tpl, err := pongo2.FromString(tplStr)
	assert.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"amount": 1000},
			{"amount": 2000},
			{"amount": 3000},
		},
	}

	out, err := tpl.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "200.0", out)
}

func TestMinByTag(t *testing.T) {
	tplStr := `{% min_by data by "amount" scale 2 %}`
	tpl, err := pongo2.FromString(tplStr)
	assert.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"amount": 4000},
			{"amount": 1500},
			{"amount": 5000},
		},
	}

	out, err := tpl.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "15.00", out)
}

func TestMaxByTag(t *testing.T) {
	tplStr := `{% max_by data by "amount" scale 2 %}`
	tpl, err := pongo2.FromString(tplStr)
	assert.NoError(t, err)

	ctx := pongo2.Context{
		"data": []map[string]any{
			{"amount": 1000},
			{"amount": 8000},
			{"amount": 3200},
		},
	}

	out, err := tpl.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "80.00", out)
}
