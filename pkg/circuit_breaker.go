package pkg

import (
	"fmt"
	"sync"

	"github.com/LerianStudio/reporter/v4/pkg/constant"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/sony/gobreaker"
)

// CircuitBreakerManager manages circuit breakers for datasources
type CircuitBreakerManager struct {
	breakers map[string]*gobreaker.CircuitBreaker
	mu       sync.RWMutex
	logger   log.Logger
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(logger log.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		logger:   logger,
	}
}

// GetOrCreate returns existing circuit breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreate(datasourceName string) *gobreaker.CircuitBreaker {
	cbm.mu.RLock()
	breaker, exists := cbm.breakers[datasourceName]
	cbm.mu.RUnlock()

	if exists {
		return breaker
	}

	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	// Double-check after acquiring write lock
	if breaker, exists = cbm.breakers[datasourceName]; exists {
		return breaker
	}

	// Create new circuit breaker with configuration
	settings := gobreaker.Settings{
		Name:        fmt.Sprintf("datasource-%s", datasourceName),
		MaxRequests: constant.CircuitBreakerMaxRequests,
		Interval:    constant.CircuitBreakerInterval,
		Timeout:     constant.CircuitBreakerTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.ConsecutiveFailures >= constant.CircuitBreakerThreshold ||
				(counts.Requests >= 10 && failureRatio >= 0.5)
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			cbm.logger.Warnf("Circuit Breaker [%s] state changed: %s -> %s", name, from.String(), to.String())

			switch to {
			case gobreaker.StateOpen:
				cbm.logger.Errorf("Circuit Breaker [%s] OPENED - datasource is unhealthy, requests will fast-fail", name)
			case gobreaker.StateHalfOpen:
				cbm.logger.Infof("Circuit Breaker [%s] HALF-OPEN - testing datasource recovery", name)
			case gobreaker.StateClosed:
				cbm.logger.Infof("Circuit Breaker [%s] CLOSED - datasource is healthy", name)
			}
		},
	}

	breaker = gobreaker.NewCircuitBreaker(settings)
	cbm.breakers[datasourceName] = breaker

	cbm.logger.Infof("Created circuit breaker for datasource: %s", datasourceName)

	return breaker
}

// Execute runs a function through the circuit breaker
func (cbm *CircuitBreakerManager) Execute(datasourceName string, fn func() (any, error)) (any, error) {
	breaker := cbm.GetOrCreate(datasourceName)

	result, err := breaker.Execute(fn)
	if err != nil {
		if err == gobreaker.ErrOpenState {
			cbm.logger.Warnf("Circuit breaker [%s] is OPEN - request rejected immediately", datasourceName)
			return nil, fmt.Errorf("datasource %s is currently unavailable (circuit breaker open): %w", datasourceName, err)
		}

		if err == gobreaker.ErrTooManyRequests {
			cbm.logger.Warnf("Circuit breaker [%s] is HALF-OPEN - too many test requests", datasourceName)
			return nil, fmt.Errorf("datasource %s is recovering (too many requests): %w", datasourceName, err)
		}
	}

	return result, err
}

// GetState returns the current state of a circuit breaker
func (cbm *CircuitBreakerManager) GetState(datasourceName string) string {
	cbm.mu.RLock()
	breaker, exists := cbm.breakers[datasourceName]
	cbm.mu.RUnlock()

	if !exists {
		return "not_initialized"
	}

	state := breaker.State()
	switch state {
	case gobreaker.StateClosed:
		return constant.CircuitBreakerStateClosed
	case gobreaker.StateOpen:
		return constant.CircuitBreakerStateOpen
	case gobreaker.StateHalfOpen:
		return constant.CircuitBreakerStateHalfOpen
	default:
		return "unknown"
	}
}

// GetCounts returns the current counts for a circuit breaker
func (cbm *CircuitBreakerManager) GetCounts(datasourceName string) gobreaker.Counts {
	cbm.mu.RLock()
	breaker, exists := cbm.breakers[datasourceName]
	cbm.mu.RUnlock()

	if !exists {
		return gobreaker.Counts{}
	}

	return breaker.Counts()
}

// Reset resets a circuit breaker to closed state by creating a new instance
func (cbm *CircuitBreakerManager) Reset(datasourceName string) {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	if _, exists := cbm.breakers[datasourceName]; exists {
		cbm.logger.Infof("Manually resetting circuit breaker for datasource: %s", datasourceName)

		settings := gobreaker.Settings{
			Name:        fmt.Sprintf("datasource-%s", datasourceName),
			MaxRequests: constant.CircuitBreakerMaxRequests,
			Interval:    constant.CircuitBreakerInterval,
			Timeout:     constant.CircuitBreakerTimeout,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.ConsecutiveFailures >= constant.CircuitBreakerThreshold ||
					(counts.Requests >= 10 && failureRatio >= 0.5)
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				cbm.logger.Warnf("Circuit Breaker [%s] state changed: %s -> %s", name, from.String(), to.String())

				switch to {
				case gobreaker.StateOpen:
					cbm.logger.Errorf("Circuit Breaker [%s] OPENED - datasource is unhealthy, requests will fast-fail", name)
				case gobreaker.StateHalfOpen:
					cbm.logger.Infof("Circuit Breaker [%s] HALF-OPEN - testing datasource recovery", name)
				case gobreaker.StateClosed:
					cbm.logger.Infof("Circuit Breaker [%s] CLOSED - datasource is healthy", name)
				}
			},
		}

		breaker := gobreaker.NewCircuitBreaker(settings)
		cbm.breakers[datasourceName] = breaker
		cbm.logger.Infof("Circuit breaker reset completed for datasource: %s", datasourceName)
	}
}

// IsHealthy returns true if the circuit breaker is in a healthy state (closed or half-open)
// Returns false if the circuit breaker is open (datasource is unhealthy)
func (cbm *CircuitBreakerManager) IsHealthy(datasourceName string) bool {
	state := cbm.GetState(datasourceName)
	return state != constant.CircuitBreakerStateOpen
}

// ShouldAllowRetry determines if a retry should be attempted based on circuit breaker state
func (cbm *CircuitBreakerManager) ShouldAllowRetry(datasourceName string) bool {
	cbm.mu.RLock()
	breaker, exists := cbm.breakers[datasourceName]
	cbm.mu.RUnlock()

	if !exists {
		return true
	}

	state := breaker.State()
	counts := breaker.Counts()

	if state == gobreaker.StateOpen {
		cbm.logger.Warnf("Circuit breaker for '%s' is OPEN - blocking retry attempt", datasourceName)
		return false
	}

	if state == gobreaker.StateHalfOpen && counts.Requests >= constant.CircuitBreakerMaxRequests {
		cbm.logger.Warnf("Circuit breaker for '%s' is HALF-OPEN and at max capacity - blocking retry attempt", datasourceName)
		return false
	}

	return true
}
