// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pkg

import (
	"context"
	"errors"
	"testing"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/stretchr/testify/assert"
)

func TestNewCircuitBreakerManager(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	cbm := NewCircuitBreakerManager(logger)

	assert.NotNil(t, cbm)
	assert.NotNil(t, cbm.breakers)
	assert.Equal(t, 0, len(cbm.breakers))
}

func TestCircuitBreakerManager_GetOrCreate(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	tests := []struct {
		name           string
		datasourceName string
	}{
		{
			name:           "Create new circuit breaker",
			datasourceName: "test_db_1",
		},
		{
			name:           "Get existing circuit breaker",
			datasourceName: "test_db_1",
		},
		{
			name:           "Create another circuit breaker",
			datasourceName: "test_db_2",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot use t.Parallel() because subtests share cbm state
			// and the assertion after the loop depends on subtests completing

			breaker := cbm.GetOrCreate(tt.datasourceName)
			assert.NotNil(t, breaker)
		})
	}

	// Verify both breakers exist
	assert.Equal(t, 2, len(cbm.breakers))
}

func TestCircuitBreakerManager_Execute_Success(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	result, err := cbm.Execute("test_db", func() (any, error) {
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestCircuitBreakerManager_Execute_Error(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	result, err := cbm.Execute("test_db", func() (any, error) {
		return nil, errors.New("test error")
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "test error")
}

func TestCircuitBreakerManager_GetState(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	tests := []struct {
		name           string
		datasourceName string
		setup          func()
		expectedState  string
	}{
		{
			name:           "Not initialized state",
			datasourceName: "unknown_db",
			setup:          func() {},
			expectedState:  "not_initialized",
		},
		{
			name:           "Closed state (healthy)",
			datasourceName: "healthy_db",
			setup: func() {
				cbm.GetOrCreate("healthy_db")
			},
			expectedState: constant.CircuitBreakerStateClosed,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.setup()
			state := cbm.GetState(tt.datasourceName)
			assert.Equal(t, tt.expectedState, state)
		})
	}
}

func TestCircuitBreakerManager_GetCounts(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	// Test non-existent breaker
	counts := cbm.GetCounts("non_existent")
	assert.Equal(t, uint32(0), counts.Requests)

	// Create breaker and execute some requests
	cbm.GetOrCreate("test_db")
	_, _ = cbm.Execute("test_db", func() (any, error) {
		return "ok", nil
	})

	counts = cbm.GetCounts("test_db")
	assert.Equal(t, uint32(1), counts.Requests)
	assert.Equal(t, uint32(1), counts.TotalSuccesses)
}

func TestCircuitBreakerManager_Reset(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	// Create and use breaker
	cbm.GetOrCreate("test_db")
	_, _ = cbm.Execute("test_db", func() (any, error) {
		return nil, errors.New("error")
	})

	counts := cbm.GetCounts("test_db")
	assert.Equal(t, uint32(1), counts.TotalFailures)

	// Reset breaker
	cbm.Reset("test_db")

	// After reset, counts should be zero
	counts = cbm.GetCounts("test_db")
	assert.Equal(t, uint32(0), counts.Requests)
}

func TestCircuitBreakerManager_Reset_NonExistent(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	// Should not panic when resetting non-existent breaker
	assert.NotPanics(t, func() {
		cbm.Reset("non_existent")
	})
}

func TestCircuitBreakerManager_IsHealthy(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	tests := []struct {
		name           string
		datasourceName string
		setup          func()
		expected       bool
	}{
		{
			name:           "Not initialized - returns true",
			datasourceName: "unknown_db",
			setup:          func() {},
			expected:       true,
		},
		{
			name:           "Closed state - returns true",
			datasourceName: "healthy_db",
			setup: func() {
				cbm.GetOrCreate("healthy_db")
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.setup()
			result := cbm.IsHealthy(tt.datasourceName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCircuitBreakerManager_ShouldAllowRetry(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	tests := []struct {
		name           string
		datasourceName string
		setup          func()
		expected       bool
	}{
		{
			name:           "Not initialized - allows retry",
			datasourceName: "unknown_db",
			setup:          func() {},
			expected:       true,
		},
		{
			name:           "Closed state - allows retry",
			datasourceName: "healthy_db",
			setup: func() {
				cbm.GetOrCreate("healthy_db")
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.setup()
			result := cbm.ShouldAllowRetry(tt.datasourceName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCircuitBreakerManager_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	// Test concurrent access doesn't cause race conditions
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			cbm.GetOrCreate("concurrent_db")
			_, _ = cbm.Execute("concurrent_db", func() (any, error) {
				return id, nil
			})
			cbm.GetState("concurrent_db")
			cbm.GetCounts("concurrent_db")
			cbm.IsHealthy("concurrent_db")
			cbm.ShouldAllowRetry("concurrent_db")
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify breaker was created
	assert.Equal(t, 1, len(cbm.breakers))
}
