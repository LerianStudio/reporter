// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libRabbitMQ "github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
)

const (
	// healthServerReadTimeout is the maximum duration for reading the entire request.
	healthServerReadTimeout = 5 * time.Second

	// healthServerWriteTimeout is the maximum duration before timing out writes of the response.
	healthServerWriteTimeout = 5 * time.Second

	// healthServerIdleTimeout is the maximum duration an idle connection will remain open.
	healthServerIdleTimeout = 30 * time.Second

	// healthServerShutdownTimeout is the maximum duration to wait for the server to shutdown gracefully.
	healthServerShutdownTimeout = 5 * time.Second
)

// HealthServer provides HTTP liveness and readiness endpoints for the worker.
// It runs as a lightweight goroutine alongside the RabbitMQ consumer.
type HealthServer struct {
	server             *http.Server
	rabbitMQConnection *libRabbitMQ.RabbitMQConnection
	logger             log.Logger
}

// NewHealthServer creates a new HealthServer bound to the given port.
// The rabbitMQConnection is used by the /ready endpoint to verify RabbitMQ connectivity.
func NewHealthServer(port string, rabbitMQConnection *libRabbitMQ.RabbitMQConnection, logger log.Logger) *HealthServer {
	hs := &HealthServer{
		rabbitMQConnection: rabbitMQConnection,
		logger:             logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", hs.handleHealth)
	mux.HandleFunc("/ready", hs.handleReady)

	hs.server = &http.Server{
		Addr:         net.JoinHostPort("", port),
		Handler:      mux,
		ReadTimeout:  healthServerReadTimeout,
		WriteTimeout: healthServerWriteTimeout,
		IdleTimeout:  healthServerIdleTimeout,
	}

	return hs
}

// Start begins listening for health check requests in a background goroutine.
func (hs *HealthServer) Start() {
	go func() {
		hs.logger.Infof("Health server listening on %s", hs.server.Addr)

		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			hs.logger.Errorf("Health server error: %v", err)
		}
	}()
}

// Shutdown gracefully stops the health server.
func (hs *HealthServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), healthServerShutdownTimeout)
	defer cancel()

	if err := hs.server.Shutdown(ctx); err != nil {
		hs.logger.Errorf("Health server shutdown error: %v", err)
	}
}

// handleHealth is the liveness probe handler.
// Returns 200 OK if the process is alive. No dependency checks.
func (hs *HealthServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := map[string]string{"status": "alive"}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		hs.logger.Errorf("Failed to encode health response: %v", err)
	}
}

// handleReady is the readiness probe handler.
// Returns 200 OK only when RabbitMQ is connected and healthy. Returns 503 otherwise.
func (hs *HealthServer) handleReady(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rabbitStatus := hs.checkRabbitMQ()

	if rabbitStatus.Status != "ready" {
		w.WriteHeader(http.StatusServiceUnavailable)

		resp := map[string]any{
			"status": "not_ready",
			"dependencies": map[string]any{
				"rabbitmq": rabbitStatus,
			},
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			hs.logger.Errorf("Failed to encode readiness response: %v", err)
		}

		return
	}

	w.WriteHeader(http.StatusOK)

	resp := map[string]any{
		"status": "ready",
		"dependencies": map[string]any{
			"rabbitmq": rabbitStatus,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		hs.logger.Errorf("Failed to encode readiness response: %v", err)
	}
}

// dependencyStatus represents the health state of a single dependency.
type dependencyStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// checkRabbitMQ verifies the RabbitMQ connection is alive and healthy.
// Mirrors the same check pattern used by the manager's readiness endpoint.
func (hs *HealthServer) checkRabbitMQ() *dependencyStatus {
	if hs.rabbitMQConnection == nil {
		return &dependencyStatus{Status: "not_ready", Message: "connection not configured"}
	}

	if !hs.rabbitMQConnection.Connected || hs.rabbitMQConnection.Connection == nil || hs.rabbitMQConnection.Connection.IsClosed() {
		return &dependencyStatus{Status: "not_ready", Message: "connection is closed"}
	}

	if !hs.rabbitMQConnection.HealthCheck() {
		return &dependencyStatus{Status: "not_ready", Message: "health check failed"}
	}

	return &dependencyStatus{Status: "ready"}
}
