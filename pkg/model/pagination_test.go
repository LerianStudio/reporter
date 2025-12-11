package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagination_SetItems(t *testing.T) {
	tests := []struct {
		name  string
		items any
	}{
		{
			name:  "set slice of strings",
			items: []string{"item1", "item2", "item3"},
		},
		{
			name:  "set slice of integers",
			items: []int{1, 2, 3},
		},
		{
			name:  "set slice of structs",
			items: []struct{ Name string }{{Name: "test1"}, {Name: "test2"}},
		},
		{
			name:  "set empty slice",
			items: []string{},
		},
		{
			name:  "set nil",
			items: nil,
		},
		{
			name:  "set map",
			items: map[string]int{"a": 1, "b": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pagination{}
			p.SetItems(tt.items)
			assert.Equal(t, tt.items, p.Items)
		})
	}
}

func TestPagination_SetTotal(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		expected int
	}{
		{
			name:     "set positive total",
			total:    100,
			expected: 100,
		},
		{
			name:     "set zero total",
			total:    0,
			expected: 0,
		},
		{
			name:     "set large total",
			total:    1000000,
			expected: 1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pagination{}
			p.SetTotal(tt.total)
			assert.Equal(t, tt.expected, p.Total)
		})
	}
}

func TestPagination_JSONSerialization(t *testing.T) {
	items := []string{"item1", "item2"}
	p := &Pagination{
		Items: items,
		Page:  1,
		Limit: 10,
		Total: 100,
	}

	data, err := json.Marshal(p)
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Contains(t, result, "items")
	assert.Contains(t, result, "page")
	assert.Contains(t, result, "limit")
	assert.Contains(t, result, "total")
}

func TestPagination_JSONOmitEmpty(t *testing.T) {
	// Page should be omitted when zero due to omitempty tag
	p := &Pagination{
		Items: []string{"item"},
		Page:  0,
		Limit: 10,
		Total: 1,
	}

	data, err := json.Marshal(p)
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	// Page should be omitted (not present in JSON) when zero
	_, pageExists := result["page"]
	assert.False(t, pageExists, "page should be omitted when zero")
}
