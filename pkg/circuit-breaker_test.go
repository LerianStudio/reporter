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

func TestCircuitBreakerManager_New(t *testing.T) {
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

func TestCircuitBreakerManager_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		fn             func() (any, error)
		expectedResult any
		expectError    bool
		errContains    string
	}{
		{
			name: "Success - returns result without error",
			fn: func() (any, error) {
				return "success", nil
			},
			expectedResult: "success",
			expectError:    false,
		},
		{
			name: "Error - returns nil result with error",
			fn: func() (any, error) {
				return nil, errors.New("test error")
			},
			expectedResult: nil,
			expectError:    true,
			errContains:    "test error",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
			cbm := NewCircuitBreakerManager(logger)

			result, err := cbm.Execute("test_db", tt.fn)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
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

// tripCircuitBreaker sends enough consecutive failures to open the circuit breaker.
func tripCircuitBreaker(cbm *CircuitBreakerManager, datasourceName string) {
	cb := cbm.GetOrCreate(datasourceName)
	for i := 0; i < int(constant.CircuitBreakerThreshold)+5; i++ {
		_, _ = cb.Execute(func() (any, error) {
			return nil, errors.New("deliberate failure")
		})
	}
}

func TestCircuitBreakerManager_Execute_OpenState(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	// Trip the circuit breaker to open state
	tripCircuitBreaker(cbm, "open_db")

	// Verify state is open
	state := cbm.GetState("open_db")
	assert.Equal(t, constant.CircuitBreakerStateOpen, state)

	// Execute through the manager should wrap the error
	result, err := cbm.Execute("open_db", func() (any, error) {
		return "should not run", nil
	})

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "currently unavailable")
	assert.Contains(t, err.Error(), "circuit breaker open")
}

func TestCircuitBreakerManager_GetState_OpenState(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	tripCircuitBreaker(cbm, "state_open_db")

	state := cbm.GetState("state_open_db")
	assert.Equal(t, constant.CircuitBreakerStateOpen, state)
}

func TestCircuitBreakerManager_IsHealthy_OpenState(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	tripCircuitBreaker(cbm, "unhealthy_db")

	// Open state should not be healthy
	assert.False(t, cbm.IsHealthy("unhealthy_db"))
}

func TestCircuitBreakerManager_ShouldAllowRetry_OpenState(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	tripCircuitBreaker(cbm, "retry_open_db")

	// Open state should block retries
	assert.False(t, cbm.ShouldAllowRetry("retry_open_db"))
}

func TestCircuitBreakerManager_Reset_AfterOpen(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	// Trip the circuit breaker
	tripCircuitBreaker(cbm, "reset_open_db")
	assert.Equal(t, constant.CircuitBreakerStateOpen, cbm.GetState("reset_open_db"))

	// Reset it
	cbm.Reset("reset_open_db")

	// After reset, state should be closed and counts should be zero
	assert.Equal(t, constant.CircuitBreakerStateClosed, cbm.GetState("reset_open_db"))

	counts := cbm.GetCounts("reset_open_db")
	assert.Equal(t, uint32(0), counts.Requests)
	assert.Equal(t, uint32(0), counts.TotalFailures)

	// Should be able to execute again
	result, err := cbm.Execute("reset_open_db", func() (any, error) {
		return "recovered", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "recovered", result)
}

func TestCircuitBreakerManager_Execute_TooManyRequests(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	// Trip the circuit breaker first
	tripCircuitBreaker(cbm, "toomany_db")
	assert.Equal(t, constant.CircuitBreakerStateOpen, cbm.GetState("toomany_db"))

	// Wait for circuit breaker timeout to transition to half-open
	// CircuitBreakerTimeout is typically short in test scenarios
	// Note: gobreaker transitions to half-open when requests come in after timeout
	// The Execute wraps ErrTooManyRequests but this requires half-open state
	// which needs waiting for the timeout. For unit test purposes, we verify the open state path.

	// Verify the open state error wrapping
	_, err := cbm.Execute("toomany_db", func() (any, error) {
		return nil, nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "currently unavailable")
}

func TestCircuitBreakerManager_Reset_PreservesOtherBreakers(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	// Create two breakers
	cbm.GetOrCreate("db_a")
	cbm.GetOrCreate("db_b")

	// Trip one
	tripCircuitBreaker(cbm, "db_a")
	assert.Equal(t, constant.CircuitBreakerStateOpen, cbm.GetState("db_a"))

	// Execute on the other should still work
	result, err := cbm.Execute("db_b", func() (any, error) {
		return "ok", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "ok", result)

	// Reset only db_a
	cbm.Reset("db_a")
	assert.Equal(t, constant.CircuitBreakerStateClosed, cbm.GetState("db_a"))

	// db_b should still be closed (it was never tripped)
	assert.Equal(t, constant.CircuitBreakerStateClosed, cbm.GetState("db_b"))
}

func TestCircuitBreakerManager_Reset_MultipleResets(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	cbm.GetOrCreate("multi_reset_db")

	// Trip and reset multiple times
	for i := 0; i < 3; i++ {
		tripCircuitBreaker(cbm, "multi_reset_db")
		assert.Equal(t, constant.CircuitBreakerStateOpen, cbm.GetState("multi_reset_db"))

		cbm.Reset("multi_reset_db")
		assert.Equal(t, constant.CircuitBreakerStateClosed, cbm.GetState("multi_reset_db"))

		// Verify it works after reset
		result, err := cbm.Execute("multi_reset_db", func() (any, error) {
			return "attempt", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "attempt", result)
	}
}

func TestCircuitBreakerManager_ReadyToTrip_FailureRatio(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	// Intersperse successes and failures to avoid hitting the consecutive failure
	// threshold (15), while reaching MinRequests (10) with >= 50% failure ratio.
	// Pattern: fail, success, fail, success, fail, success, fail, success, fail, fail
	// = 10 requests, 6 failures, 4 successes → ratio 0.6 >= 0.5

	cb := cbm.GetOrCreate("ratio_db")
	pattern := []bool{false, true, false, true, false, true, false, true, false, false}

	for _, success := range pattern {
		if success {
			_, _ = cb.Execute(func() (any, error) {
				return "ok", nil
			})
		} else {
			_, _ = cb.Execute(func() (any, error) {
				return nil, errors.New("failure")
			})
		}
	}

	// Circuit breaker should have tripped due to failure ratio
	state := cbm.GetState("ratio_db")
	assert.Equal(t, constant.CircuitBreakerStateOpen, state,
		"circuit breaker should open via failure ratio path (6/10 = 0.6 >= 0.5)")
}

func TestCircuitBreakerManager_ReadyToTrip_BelowMinRequests(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	// Send fewer than MinRequests (10) with high failure ratio
	// but not enough consecutive failures to hit threshold (15)
	cb := cbm.GetOrCreate("below_min_db")
	pattern := []bool{false, true, false, true, false} // 5 requests, 3 failures

	for _, success := range pattern {
		if success {
			_, _ = cb.Execute(func() (any, error) {
				return "ok", nil
			})
		} else {
			_, _ = cb.Execute(func() (any, error) {
				return nil, errors.New("failure")
			})
		}
	}

	// Should stay closed: below MinRequests and below consecutive threshold
	state := cbm.GetState("below_min_db")
	assert.Equal(t, constant.CircuitBreakerStateClosed, state,
		"circuit breaker should stay closed when below MinRequests")
}

func TestCircuitBreakerManager_ReadyToTrip_ZeroRequests(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	// Create breaker but don't send any requests - should not panic
	cbm.GetOrCreate("zero_db")

	state := cbm.GetState("zero_db")
	assert.Equal(t, constant.CircuitBreakerStateClosed, state,
		"circuit breaker should stay closed with zero requests")
}

func TestCircuitBreakerManager_GetOrCreate_Idempotent(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	cbm := NewCircuitBreakerManager(logger)

	breaker1 := cbm.GetOrCreate("idempotent_db")
	breaker2 := cbm.GetOrCreate("idempotent_db")

	// Should return the same instance
	assert.Equal(t, breaker1, breaker2)
	assert.Equal(t, 1, len(cbm.breakers))
}
