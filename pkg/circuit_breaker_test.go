package pkg

import (
	"errors"
	"testing"

	"github.com/LerianStudio/reporter/v4/pkg/constant"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewCircuitBreakerManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	cbm := NewCircuitBreakerManager(mockLogger)

	assert.NotNil(t, cbm)
	assert.NotNil(t, cbm.breakers)
	assert.Equal(t, mockLogger, cbm.logger)
}

func TestCircuitBreakerManager_GetOrCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	cbm := NewCircuitBreakerManager(mockLogger)

	// First call should create a new circuit breaker
	breaker1 := cbm.GetOrCreate("test-datasource")
	assert.NotNil(t, breaker1)

	// Second call should return the same circuit breaker
	breaker2 := cbm.GetOrCreate("test-datasource")
	assert.Equal(t, breaker1, breaker2)

	// Different datasource should create a new circuit breaker
	breaker3 := cbm.GetOrCreate("another-datasource")
	assert.NotNil(t, breaker3)
	assert.NotEqual(t, breaker1, breaker3)
}

func TestCircuitBreakerManager_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	cbm := NewCircuitBreakerManager(mockLogger)

	result, err := cbm.Execute("test-datasource", func() (any, error) {
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestCircuitBreakerManager_Execute_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	cbm := NewCircuitBreakerManager(mockLogger)

	expectedErr := errors.New("test error")
	result, err := cbm.Execute("test-datasource", func() (any, error) {
		return nil, expectedErr
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
}

func TestCircuitBreakerManager_GetState_NotInitialized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	cbm := NewCircuitBreakerManager(mockLogger)

	state := cbm.GetState("non-existent")

	assert.Equal(t, "not_initialized", state)
}

func TestCircuitBreakerManager_GetState_Closed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	cbm := NewCircuitBreakerManager(mockLogger)

	// Create circuit breaker
	cbm.GetOrCreate("test-datasource")

	state := cbm.GetState("test-datasource")

	assert.Equal(t, constant.CircuitBreakerStateClosed, state)
}

func TestCircuitBreakerManager_GetCounts_NotInitialized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	cbm := NewCircuitBreakerManager(mockLogger)

	counts := cbm.GetCounts("non-existent")

	assert.Equal(t, uint32(0), counts.Requests)
	assert.Equal(t, uint32(0), counts.TotalSuccesses)
	assert.Equal(t, uint32(0), counts.TotalFailures)
}

func TestCircuitBreakerManager_GetCounts_WithRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	cbm := NewCircuitBreakerManager(mockLogger)

	// Execute some successful requests
	_, _ = cbm.Execute("test-datasource", func() (any, error) {
		return "success", nil
	})

	counts := cbm.GetCounts("test-datasource")

	assert.Equal(t, uint32(1), counts.Requests)
	assert.Equal(t, uint32(1), counts.TotalSuccesses)
}

func TestCircuitBreakerManager_Reset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	cbm := NewCircuitBreakerManager(mockLogger)

	// Create and use circuit breaker
	cbm.GetOrCreate("test-datasource")
	_, _ = cbm.Execute("test-datasource", func() (any, error) {
		return "success", nil
	})

	// Reset
	cbm.Reset("test-datasource")

	// Counts should be reset
	counts := cbm.GetCounts("test-datasource")
	assert.Equal(t, uint32(0), counts.Requests)
}

func TestCircuitBreakerManager_Reset_NonExistent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	cbm := NewCircuitBreakerManager(mockLogger)

	// Should not panic
	cbm.Reset("non-existent")
}

func TestCircuitBreakerManager_IsHealthy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	cbm := NewCircuitBreakerManager(mockLogger)

	// Not initialized should be considered healthy
	assert.True(t, cbm.IsHealthy("non-existent"))

	// Create circuit breaker
	cbm.GetOrCreate("test-datasource")

	// Closed state should be healthy
	assert.True(t, cbm.IsHealthy("test-datasource"))
}

func TestCircuitBreakerManager_ShouldAllowRetry_NotInitialized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	cbm := NewCircuitBreakerManager(mockLogger)

	// Not initialized should allow retry
	assert.True(t, cbm.ShouldAllowRetry("non-existent"))
}

func TestCircuitBreakerManager_ShouldAllowRetry_Closed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	cbm := NewCircuitBreakerManager(mockLogger)

	cbm.GetOrCreate("test-datasource")

	// Closed state should allow retry
	assert.True(t, cbm.ShouldAllowRetry("test-datasource"))
}
