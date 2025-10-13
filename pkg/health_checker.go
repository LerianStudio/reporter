package pkg

import (
	"context"
	"plugin-smart-templates/v3/pkg/constant"
	"sync"
	"time"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
)

// HealthChecker performs periodic health checks on datasources and attempts reconnection
type HealthChecker struct {
	dataSources           *map[string]DataSource
	circuitBreakerManager *CircuitBreakerManager
	logger                log.Logger
	stopChan              chan struct{}
	wg                    sync.WaitGroup
	mu                    sync.RWMutex
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(
	dataSources *map[string]DataSource,
	circuitBreakerManager *CircuitBreakerManager,
	logger log.Logger,
) *HealthChecker {
	return &HealthChecker{
		dataSources:           dataSources,
		circuitBreakerManager: circuitBreakerManager,
		logger:                logger,
		stopChan:              make(chan struct{}),
	}
}

// Start begins the health check loop in a separate goroutine
func (hc *HealthChecker) Start() {
	hc.wg.Add(1)
	go hc.healthCheckLoop()
	hc.logger.Info("🏥 Health checker started - checking datasources every 30s")
}

// Stop gracefully stops the health checker
func (hc *HealthChecker) Stop() {
	close(hc.stopChan)
	hc.wg.Wait()
	hc.logger.Info("🏥 Health checker stopped")
}

// healthCheckLoop runs the periodic health checks
func (hc *HealthChecker) healthCheckLoop() {
	defer hc.wg.Done()

	ticker := time.NewTicker(constant.HealthCheckInterval)
	defer ticker.Stop()

	// Run initial check after a short delay
	time.Sleep(5 * time.Second)
	hc.performHealthChecks()

	for {
		select {
		case <-ticker.C:
			hc.performHealthChecks()
		case <-hc.stopChan:
			return
		}
	}
}

// performHealthChecks checks all datasources and attempts reconnection if needed
func (hc *HealthChecker) performHealthChecks() {
	hc.mu.RLock()
	dataSources := *hc.dataSources
	hc.mu.RUnlock()

	hc.logger.Info("🔍 Performing health checks on all datasources...")

	unavailableCount := 0
	reconnectedCount := 0

	for name, ds := range dataSources {
		// Check if datasource needs healing
		if hc.needsHealing(name, ds) {
			unavailableCount++
			hc.logger.Infof("🔧 Attempting to heal datasource '%s' (status: %s)", name, ds.Status)

			if hc.attemptReconnection(name, &ds) {
				reconnectedCount++

				// Update datasource in map
				hc.mu.Lock()
				(*hc.dataSources)[name] = ds
				hc.mu.Unlock()

				// Reset circuit breaker
				hc.circuitBreakerManager.Reset(name)
				hc.logger.Infof("✅ Datasource '%s' reconnected successfully - circuit breaker reset", name)
			} else {
				hc.logger.Warnf("⚠️  Failed to reconnect datasource '%s' - will retry in %v", name, constant.HealthCheckInterval)
			}
		}
	}

	if unavailableCount > 0 {
		hc.logger.Infof("🏥 Health check complete: %d datasources needed healing, %d reconnected", unavailableCount, reconnectedCount)
	} else {
		hc.logger.Debug("✅ All datasources healthy")
	}
}

// needsHealing determines if a datasource needs reconnection attempt
func (hc *HealthChecker) needsHealing(name string, ds DataSource) bool {
	// Datasource is unavailable
	if ds.Status == constant.DataSourceStatusUnavailable {
		return true
	}

	// Datasource is not initialized
	if !ds.Initialized {
		return true
	}

	// Circuit breaker is open (datasource is unhealthy)
	if !hc.circuitBreakerManager.IsHealthy(name) {
		cbState := hc.circuitBreakerManager.GetState(name)
		if cbState == constant.CircuitBreakerStateOpen {
			return true
		}
	}

	return false
}

// attemptReconnection tries to reconnect to a datasource
func (hc *HealthChecker) attemptReconnection(name string, ds *DataSource) bool {
	ctx, cancel := context.WithTimeout(context.Background(), constant.HealthCheckTimeout)
	defer cancel()

	hc.logger.Infof("🔌 Attempting reconnection to datasource '%s'...", name)

	// Create a temporary map for ConnectToDataSource
	tempMap := make(map[string]DataSource)
	tempMap[name] = *ds

	// Reset retry count before attempting reconnection
	ds.RetryCount = 0
	ds.LastAttempt = time.Now()

	// Attempt connection (single attempt, no retry loop)
	err := ConnectToDataSource(name, ds, hc.logger, tempMap)
	if err != nil {
		hc.logger.Errorf("Failed to reconnect datasource '%s': %v", name, err)
		ds.Status = constant.DataSourceStatusUnavailable
		ds.LastError = err
		return false
	}

	// Check if connection is actually working with a ping
	if !hc.pingDataSource(ctx, name, ds) {
		hc.logger.Errorf("Reconnection to '%s' succeeded but ping failed", name)
		ds.Status = constant.DataSourceStatusDegraded
		return false
	}

	// Update datasource status
	ds.Status = constant.DataSourceStatusAvailable
	ds.Initialized = true
	ds.LastError = nil

	return true
}

// pingDataSource performs a simple ping to verify datasource connectivity
func (hc *HealthChecker) pingDataSource(ctx context.Context, name string, ds *DataSource) bool {
	switch ds.DatabaseType {
	case PostgreSQLType:
		if ds.PostgresRepository == nil {
			return false
		}
		// Try to get schema as a ping (lightweight operation)
		_, err := ds.PostgresRepository.GetDatabaseSchema(ctx)
		return err == nil

	case MongoDBType:
		if ds.MongoDBRepository == nil {
			return false
		}
		// Try to get schema as a ping (lightweight operation)
		_, err := ds.MongoDBRepository.GetDatabaseSchema(ctx)
		return err == nil

	default:
		hc.logger.Warnf("Unknown database type for datasource '%s': %s", name, ds.DatabaseType)
		return false
	}
}

// GetHealthStatus returns the current health status of all datasources
func (hc *HealthChecker) GetHealthStatus() map[string]string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	status := make(map[string]string)
	for name, ds := range *hc.dataSources {
		cbState := hc.circuitBreakerManager.GetState(name)
		status[name] = ds.Status + " (CB: " + cbState + ")"
	}

	return status
}
