package pkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsValidDate(t *testing.T) {
	tests := []struct {
		date     string
		expected bool
	}{
		{"2024-01-15", true},
		{"2024-12-31", true},
		{"2024-02-29", true},  // Leap year
		{"2023-02-29", false}, // Not leap year
		{"2024-13-01", false}, // Invalid month
		{"2024-01-32", false}, // Invalid day
		{"24-01-15", false},   // Wrong format
		{"2024/01/15", false}, // Wrong separator
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.date, func(t *testing.T) {
			result := IsValidDate(tt.date)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsInitialDateBeforeFinalDate(t *testing.T) {
	tests := []struct {
		name     string
		initial  time.Time
		final    time.Time
		expected bool
	}{
		{
			name:     "initial before final",
			initial:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			final:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "initial equal to final",
			initial:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			final:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "initial after final",
			initial:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			final:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInitialDateBeforeFinalDate(tt.initial, tt.final)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsDateRangeWithinMonthLimit(t *testing.T) {
	tests := []struct {
		name     string
		initial  time.Time
		final    time.Time
		limit    int
		expected bool
	}{
		{
			name:     "within limit",
			initial:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			final:    time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			limit:    3,
			expected: true,
		},
		{
			name:     "exactly at limit",
			initial:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			final:    time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			limit:    3,
			expected: true,
		},
		{
			name:     "exceeds limit",
			initial:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			final:    time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
			limit:    3,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDateRangeWithinMonthLimit(tt.initial, tt.final, tt.limit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeDate(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Without days adjustment
	result := NormalizeDate(date, nil)
	assert.Equal(t, "2024-01-15", result)

	// With positive days
	days := 5
	result = NormalizeDate(date, &days)
	assert.Equal(t, "2024-01-20", result)

	// With negative days
	days = -5
	result = NormalizeDate(date, &days)
	assert.Equal(t, "2024-01-10", result)
}
