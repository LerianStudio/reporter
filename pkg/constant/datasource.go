package constant

import "time"

// Database Query Timeouts
const (
	QueryTimeoutFast       = 5 * time.Second
	QueryTimeoutMedium     = 10 * time.Second
	QueryTimeoutSlow       = 15 * time.Second
	SchemaDiscoveryTimeout = 30 * time.Second
	ConnectionTimeout      = 5 * time.Second
)

// Circuit Breaker Configuration
const (
	CircuitBreakerMaxRequests uint32 = 3
	CircuitBreakerInterval           = 2 * time.Minute
	CircuitBreakerTimeout            = 30 * time.Second
	CircuitBreakerThreshold   uint32 = 15
)

// PostgreSQL Pool Configuration
const (
	PostgresMaxOpenConns    = 25
	PostgresMaxIdleConns    = 10
	PostgresConnMaxLifetime = 5 * time.Minute
	PostgresConnMaxIdleTime = 1 * time.Minute
)

// MongoDB Pool Configuration
const (
	MongoDBMaxPoolSize     uint64 = 100
	MongoDBMinPoolSize     uint64 = 10
	MongoDBMaxConnIdleTime        = 1 * time.Minute
)

// Circuit Breaker State Names
const (
	CircuitBreakerStateClosed   = "closed"
	CircuitBreakerStateOpen     = "open"
	CircuitBreakerStateHalfOpen = "half-open"
)

// DataSource Initialization Retry Configuration
const (
	DataSourceMaxRetries        = 3
	DataSourceInitialBackoff    = 1 * time.Second
	DataSourceMaxBackoff        = 10 * time.Second
	DataSourceBackoffMultiplier = 2.0
)

// DataSource Status
const (
	DataSourceStatusUnknown     = "unknown"
	DataSourceStatusAvailable   = "available"
	DataSourceStatusUnavailable = "unavailable"
	DataSourceStatusDegraded    = "degraded"
)

// Health Check Configuration
const (
	HealthCheckInterval = 30 * time.Second
	HealthCheckTimeout  = 5 * time.Second
)
