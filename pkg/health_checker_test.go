// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pkg

import (
	"context"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/pkg/constant"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libConstants "github.com/LerianStudio/lib-commons/v2/commons/constants"
	"github.com/stretchr/testify/assert"
)

func TestNewHealthChecker(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	assert.NotNil(t, hc)
	assert.NotNil(t, hc.dataSources)
	assert.NotNil(t, hc.circuitBreakerManager)
	assert.NotNil(t, hc.logger)
	assert.NotNil(t, hc.stopChan)
}

func TestHealthChecker_StartAndStop(t *testing.T) {
	t.Parallel()

	// Skip this test in short mode since it requires waiting for the health checker's
	// initial 5-second delay to complete before Stop() can take effect
	if testing.Short() {
		t.Skip("Skipping test in short mode - requires waiting for health check delay")
	}

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	// Start should not block
	hc.Start()

	// Wait for the initial delay (5 seconds) plus buffer time
	// so that the health check loop enters its select statement
	time.Sleep(6 * time.Second)

	// Stop should gracefully terminate once in the select loop
	done := make(chan struct{})
	go func() {
		hc.Stop()
		close(done)
	}()

	// Should stop quickly once stopChan is closed
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Health checker did not stop within expected time")
	}
}

func TestHealthChecker_StartNotBlocking(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	// Start should not block - verify it returns immediately
	started := make(chan struct{})
	go func() {
		hc.Start()
		close(started)
	}()

	select {
	case <-started:
		// Success - Start() returned immediately
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Start() should not block")
	}

	// Note: Not calling Stop() here to avoid the 5s+ delay in tests.
	// The goroutine will be cleaned up when the test exits.
}

func TestHealthChecker_NeedsHealing(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	tests := []struct {
		name           string
		datasourceName string
		ds             DataSource
		setup          func()
		expected       bool
	}{
		{
			name:           "Unavailable status - needs healing",
			datasourceName: "test_db_1",
			ds: DataSource{
				Status:      libConstants.DataSourceStatusUnavailable,
				Initialized: true,
			},
			setup:    func() {},
			expected: true,
		},
		{
			name:           "Not initialized - needs healing",
			datasourceName: "test_db_2",
			ds: DataSource{
				Status:      libConstants.DataSourceStatusAvailable,
				Initialized: false,
			},
			setup:    func() {},
			expected: true,
		},
		{
			name:           "Available and initialized - no healing needed",
			datasourceName: "test_db_3",
			ds: DataSource{
				Status:      libConstants.DataSourceStatusAvailable,
				Initialized: true,
			},
			setup: func() {
				// Ensure circuit breaker is healthy
				cbManager.GetOrCreate("test_db_3")
			},
			expected: false,
		},
		{
			name:           "Circuit breaker open - needs healing",
			datasourceName: "test_db_4",
			ds: DataSource{
				Status:      libConstants.DataSourceStatusAvailable,
				Initialized: true,
			},
			setup: func() {
				// Force circuit breaker to open state
				cb := cbManager.GetOrCreate("test_db_4")
				// Execute enough failures to trip the circuit breaker
				for i := 0; i < 10; i++ {
					_, _ = cb.Execute(func() (any, error) {
						return nil, assert.AnError
					})
				}
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.setup()
			result := hc.needsHealing(tt.datasourceName, tt.ds)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHealthChecker_GetHealthStatus(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	dataSources["db1"] = DataSource{
		Status:       libConstants.DataSourceStatusAvailable,
		DatabaseType: PostgreSQLType,
		Initialized:  true,
	}
	dataSources["db2"] = DataSource{
		Status:       libConstants.DataSourceStatusUnavailable,
		DatabaseType: MongoDBType,
		Initialized:  false,
	}

	cbManager := NewCircuitBreakerManager(logger)
	cbManager.GetOrCreate("db1")

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	status := hc.GetHealthStatus()

	assert.Len(t, status, 2)
	assert.Contains(t, status["db1"], libConstants.DataSourceStatusAvailable)
	assert.Contains(t, status["db2"], libConstants.DataSourceStatusUnavailable)
}

func TestHealthChecker_GetHealthStatus_Empty(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	status := hc.GetHealthStatus()

	assert.Empty(t, status)
}

func TestHealthChecker_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	for i := 0; i < 5; i++ {
		name := "db" + string(rune('0'+i))
		dataSources[name] = DataSource{
			Status:       libConstants.DataSourceStatusAvailable,
			DatabaseType: PostgreSQLType,
			Initialized:  true,
		}
	}

	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	// Test concurrent read access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_ = hc.GetHealthStatus()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}
}

func TestHealthChecker_PingDataSource_UnknownType(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	ds := &DataSource{
		DatabaseType: "unknown_type",
		Initialized:  true,
	}

	// Should return false for unknown database type
	result := hc.pingDataSource(nil, "test_db", ds)
	assert.False(t, result)
}

func TestHealthChecker_PingDataSource_NilPostgresRepository(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	ds := &DataSource{
		DatabaseType:       PostgreSQLType,
		PostgresRepository: nil,
		Initialized:        true,
	}

	result := hc.pingDataSource(nil, "test_db", ds)
	assert.False(t, result)
}

func TestHealthChecker_PingDataSource_NilMongoDBRepository(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	ds := &DataSource{
		DatabaseType:      MongoDBType,
		MongoDBRepository: nil,
		Initialized:       true,
	}

	result := hc.pingDataSource(nil, "test_db", ds)
	assert.False(t, result)
}

func TestHealthChecker_PerformHealthChecks_AllHealthy(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	dataSources["healthy_db"] = DataSource{
		Status:       libConstants.DataSourceStatusAvailable,
		DatabaseType: PostgreSQLType,
		Initialized:  true,
	}

	cbManager := NewCircuitBreakerManager(logger)
	cbManager.GetOrCreate("healthy_db")

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	// Should not panic when all datasources are healthy
	assert.NotPanics(t, func() {
		hc.performHealthChecks()
	})
}

func TestHealthChecker_PerformHealthChecks_WithUnavailable(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	dataSources["unavailable_db"] = DataSource{
		Status:       libConstants.DataSourceStatusUnavailable,
		DatabaseType: PostgreSQLType,
		Initialized:  false,
	}

	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	// Should not panic when handling unavailable datasources
	assert.NotPanics(t, func() {
		hc.performHealthChecks()
	})
}

func TestHealthCheckConstants(t *testing.T) {
	t.Parallel()

	// Verify health check constants are defined properly
	assert.Equal(t, 30*time.Second, constant.HealthCheckInterval)
	assert.Equal(t, 5*time.Second, constant.HealthCheckTimeout)
}

func TestHealthChecker_AttemptReconnection_NilConnection(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	ds := &DataSource{
		DatabaseType:   PostgreSQLType,
		Initialized:    false,
		Status:         libConstants.DataSourceStatusUnavailable,
		DatabaseConfig: nil, // No connection config
	}

	// Should return false when connection cannot be established
	result := hc.attemptReconnection("test_db", ds)
	assert.False(t, result)
	assert.Equal(t, libConstants.DataSourceStatusUnavailable, ds.Status)
}

func TestHealthChecker_MultipleDataSources(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	dataSources := make(map[string]DataSource)
	dataSources["db1"] = DataSource{
		Status:       libConstants.DataSourceStatusAvailable,
		DatabaseType: PostgreSQLType,
		Initialized:  true,
	}
	dataSources["db2"] = DataSource{
		Status:       libConstants.DataSourceStatusAvailable,
		DatabaseType: MongoDBType,
		Initialized:  true,
	}
	dataSources["db3"] = DataSource{
		Status:       libConstants.DataSourceStatusUnavailable,
		DatabaseType: PostgreSQLType,
		Initialized:  false,
	}

	cbManager := NewCircuitBreakerManager(logger)
	cbManager.GetOrCreate("db1")
	cbManager.GetOrCreate("db2")

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	status := hc.GetHealthStatus()
	assert.Len(t, status, 3)
}
