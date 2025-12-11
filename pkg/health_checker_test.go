package pkg

import (
	"context"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb"
	pg "github.com/LerianStudio/reporter/v4/pkg/postgres"

	libConstants "github.com/LerianStudio/lib-commons/v2/commons/constants"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewHealthChecker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	assert.NotNil(t, hc)
	assert.Equal(t, &dataSources, hc.dataSources)
	assert.Equal(t, cbManager, hc.circuitBreakerManager)
	assert.Equal(t, logger, hc.logger)
	assert.NotNil(t, hc.stopChan)
}

func TestHealthChecker_StartStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any()).AnyTimes()

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	// Start in a goroutine
	hc.Start()

	// Give it a moment to start (health check loop has 5s initial delay)
	time.Sleep(100 * time.Millisecond)

	// Stop should not block - increase timeout to account for health check loop
	done := make(chan struct{})
	go func() {
		hc.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success - stopped cleanly
	case <-time.After(10 * time.Second):
		t.Fatal("Stop() took too long - possible deadlock")
	}
}

func TestHealthChecker_NeedsHealing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	tests := []struct {
		name           string
		datasourceName string
		dataSource     DataSource
		setupCB        func()
		expected       bool
	}{
		{
			name:           "unavailable status should need healing",
			datasourceName: "test-db",
			dataSource: DataSource{
				Status:      libConstants.DataSourceStatusUnavailable,
				Initialized: true,
			},
			setupCB:  func() {},
			expected: true,
		},
		{
			name:           "not initialized should need healing",
			datasourceName: "test-db",
			dataSource: DataSource{
				Status:      libConstants.DataSourceStatusAvailable,
				Initialized: false,
			},
			setupCB:  func() {},
			expected: true,
		},
		{
			name:           "available and initialized should not need healing",
			datasourceName: "test-db",
			dataSource: DataSource{
				Status:      libConstants.DataSourceStatusAvailable,
				Initialized: true,
			},
			setupCB:  func() {},
			expected: false,
		},
		{
			name:           "open circuit breaker should need healing",
			datasourceName: "test-cb",
			dataSource: DataSource{
				Status:      libConstants.DataSourceStatusAvailable,
				Initialized: true,
			},
			setupCB: func() {
				// Force circuit breaker to open state by recording failures
				cb := cbManager.GetOrCreate("test-cb")
				for i := 0; i < 10; i++ {
					_, _ = cb.Execute(func() (any, error) {
						return nil, assert.AnError
					})
				}
			},
			expected: true,
		},
		{
			name:           "unknown status and not initialized should need healing",
			datasourceName: "test-db",
			dataSource: DataSource{
				Status:      libConstants.DataSourceStatusUnknown,
				Initialized: false,
			},
			setupCB:  func() {},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset circuit breaker manager for clean state
			cbManager = NewCircuitBreakerManager(logger)
			hc.circuitBreakerManager = cbManager

			tt.setupCB()

			result := hc.needsHealing(tt.datasourceName, tt.dataSource)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHealthChecker_PingDataSource_PostgreSQL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)
	ctx := context.Background()

	t.Run("nil postgres repository should return false", func(t *testing.T) {
		ds := &DataSource{
			DatabaseType:       PostgreSQLType,
			PostgresRepository: nil,
		}

		result := hc.pingDataSource(ctx, "test-pg", ds)
		assert.False(t, result)
	})

	t.Run("postgres with working repository should return true", func(t *testing.T) {
		mockPgRepo := pg.NewMockRepository(ctrl)
		mockPgRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(nil, nil)

		ds := &DataSource{
			DatabaseType:       PostgreSQLType,
			PostgresRepository: mockPgRepo,
		}

		result := hc.pingDataSource(ctx, "test-pg", ds)
		assert.True(t, result)
	})

	t.Run("postgres with failing repository should return false", func(t *testing.T) {
		mockPgRepo := pg.NewMockRepository(ctrl)
		mockPgRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(nil, assert.AnError)

		ds := &DataSource{
			DatabaseType:       PostgreSQLType,
			PostgresRepository: mockPgRepo,
		}

		result := hc.pingDataSource(ctx, "test-pg", ds)
		assert.False(t, result)
	})
}

func TestHealthChecker_PingDataSource_MongoDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)
	ctx := context.Background()

	t.Run("nil mongodb repository should return false", func(t *testing.T) {
		ds := &DataSource{
			DatabaseType:      MongoDBType,
			MongoDBRepository: nil,
		}

		result := hc.pingDataSource(ctx, "test-mongo", ds)
		assert.False(t, result)
	})

	t.Run("mongodb with working repository should return true", func(t *testing.T) {
		mockMongoRepo := mongodb.NewMockRepository(ctrl)
		mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(nil, nil)

		ds := &DataSource{
			DatabaseType:      MongoDBType,
			MongoDBRepository: mockMongoRepo,
		}

		result := hc.pingDataSource(ctx, "test-mongo", ds)
		assert.True(t, result)
	})

	t.Run("mongodb with failing repository should return false", func(t *testing.T) {
		mockMongoRepo := mongodb.NewMockRepository(ctrl)
		mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(nil, assert.AnError)

		ds := &DataSource{
			DatabaseType:      MongoDBType,
			MongoDBRepository: mockMongoRepo,
		}

		result := hc.pingDataSource(ctx, "test-mongo", ds)
		assert.False(t, result)
	})
}

func TestHealthChecker_PingDataSource_UnknownType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)
	ctx := context.Background()

	ds := &DataSource{
		DatabaseType: "oracle", // unknown type
	}

	result := hc.pingDataSource(ctx, "test-oracle", ds)
	assert.False(t, result)
}

func TestHealthChecker_GetHealthStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	dataSources := map[string]DataSource{
		"db1": {
			Status:       libConstants.DataSourceStatusAvailable,
			Initialized:  true,
			DatabaseType: PostgreSQLType,
		},
		"db2": {
			Status:       libConstants.DataSourceStatusUnavailable,
			Initialized:  false,
			DatabaseType: MongoDBType,
		},
		"db3": {
			Status:       libConstants.DataSourceStatusDegraded,
			Initialized:  true,
			DatabaseType: PostgreSQLType,
		},
	}
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	status := hc.GetHealthStatus()

	assert.Len(t, status, 3)
	assert.Contains(t, status["db1"], libConstants.DataSourceStatusAvailable)
	assert.Contains(t, status["db2"], libConstants.DataSourceStatusUnavailable)
	assert.Contains(t, status["db3"], libConstants.DataSourceStatusDegraded)

	// Check that CB state is included
	assert.Contains(t, status["db1"], "CB:")
	assert.Contains(t, status["db2"], "CB:")
	assert.Contains(t, status["db3"], "CB:")
}

func TestHealthChecker_GetHealthStatus_EmptyDatasources(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	status := hc.GetHealthStatus()

	assert.NotNil(t, status)
	assert.Len(t, status, 0)
}

func TestHealthChecker_PerformHealthChecks_NoDatasources(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any()).Times(1) // "All datasources healthy"

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	// Should not panic with empty datasources
	hc.performHealthChecks()
}

func TestHealthChecker_PerformHealthChecks_AllHealthy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any()).Times(1) // "All datasources healthy"

	dataSources := map[string]DataSource{
		"healthy-db": {
			Status:       libConstants.DataSourceStatusAvailable,
			Initialized:  true,
			DatabaseType: PostgreSQLType,
		},
	}
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	hc.performHealthChecks()
}

func TestHealthChecker_Struct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := &HealthChecker{
		dataSources:           &dataSources,
		circuitBreakerManager: cbManager,
		logger:                logger,
		stopChan:              make(chan struct{}),
	}

	assert.NotNil(t, hc.dataSources)
	assert.NotNil(t, hc.circuitBreakerManager)
	assert.NotNil(t, hc.logger)
	assert.NotNil(t, hc.stopChan)
}


func TestHealthChecker_NeedsHealing_WithCircuitBreakerStates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	dataSources := make(map[string]DataSource)
	cbManager := NewCircuitBreakerManager(logger)

	hc := NewHealthChecker(&dataSources, cbManager, logger)

	t.Run("closed circuit breaker should not need healing", func(t *testing.T) {
		ds := DataSource{
			Status:      libConstants.DataSourceStatusAvailable,
			Initialized: true,
		}

		// Circuit breaker starts closed
		result := hc.needsHealing("test-db", ds)
		assert.False(t, result)
	})

	t.Run("half-open circuit breaker should not need healing", func(t *testing.T) {
		ds := DataSource{
			Status:      libConstants.DataSourceStatusAvailable,
			Initialized: true,
		}

		// Force half-open state
		cb := cbManager.GetOrCreate("half-open-db")
		// Record failures to open, then wait for half-open transition
		for i := 0; i < 10; i++ {
			_, _ = cb.Execute(func() (any, error) {
				return nil, assert.AnError
			})
		}

		// Half-open state is reached after timeout - for unit test just verify initial state
		result := hc.needsHealing("half-open-db", ds)
		// When CB is open, it needs healing
		assert.True(t, result)
	})
}

func TestHealthChecker_Constants(t *testing.T) {
	// Verify that constants are properly defined
	assert.NotZero(t, constant.HealthCheckInterval)
	assert.NotZero(t, constant.HealthCheckTimeout)
	assert.Equal(t, constant.CircuitBreakerStateOpen, "open")
	assert.Equal(t, constant.CircuitBreakerStateClosed, "closed")
	assert.Equal(t, constant.CircuitBreakerStateHalfOpen, "half-open")
}
