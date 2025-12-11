package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryHeader_ToOffsetPagination(t *testing.T) {
	qh := &QueryHeader{
		Limit:     25,
		Page:      3,
		SortOrder: "asc",
		Alias:     "test-alias",
		Cursor:    "some-cursor", // Should not be included in result
	}

	result := qh.ToOffsetPagination()

	assert.Equal(t, 25, result.Limit)
	assert.Equal(t, 3, result.Page)
	assert.Equal(t, "asc", result.SortOrder)
	assert.Equal(t, "test-alias", result.Alias)
	assert.Empty(t, result.Cursor) // Cursor should be empty
}

func TestQueryHeader_ToOffsetPagination_DefaultValues(t *testing.T) {
	qh := &QueryHeader{}

	result := qh.ToOffsetPagination()

	assert.Equal(t, 0, result.Limit)
	assert.Equal(t, 0, result.Page)
	assert.Empty(t, result.SortOrder)
	assert.Empty(t, result.Alias)
}
