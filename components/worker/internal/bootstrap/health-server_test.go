// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libRabbitMQ "github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthServer_HandleHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "returns 200 alive",
			expectedStatus: http.StatusOK,
			expectedBody:   "alive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hs := NewHealthServer("0", nil, &log.NoneLogger{})

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()

			hs.handleHealth(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			var body map[string]string
			err := json.Unmarshal(rec.Body.Bytes(), &body)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, body["status"])
		})
	}
}

func TestHealthServer_HandleReady_NilConnection(t *testing.T) {
	t.Parallel()

	hs := NewHealthServer("0", nil, &log.NoneLogger{})

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	hs.handleReady(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var body map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "not_ready", body["status"])

	deps, ok := body["dependencies"].(map[string]any)
	require.True(t, ok)

	rabbit, ok := deps["rabbitmq"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "not_ready", rabbit["status"])
	assert.Equal(t, "connection not configured", rabbit["message"])
}

func TestHealthServer_HandleReady_DisconnectedRabbitMQ(t *testing.T) {
	t.Parallel()

	conn := &libRabbitMQ.RabbitMQConnection{
		Connected:  false,
		Connection: nil,
	}

	hs := NewHealthServer("0", conn, &log.NoneLogger{})

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	hs.handleReady(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var body map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "not_ready", body["status"])

	deps, ok := body["dependencies"].(map[string]any)
	require.True(t, ok)

	rabbit, ok := deps["rabbitmq"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "not_ready", rabbit["status"])
	assert.Equal(t, "connection is closed", rabbit["message"])
}

func TestHealthServer_HandleReady_ConnectedButNilAMQP(t *testing.T) {
	t.Parallel()

	// Connection marked as connected but with nil AMQP connection object
	// This simulates a state where the connection was lost after initial connect
	conn := &libRabbitMQ.RabbitMQConnection{
		Connected:  true,
		Connection: nil,
	}

	hs := NewHealthServer("0", conn, &log.NoneLogger{})

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	hs.handleReady(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var body map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "not_ready", body["status"])
}

func TestHealthServer_CheckRabbitMQ_NilConnection(t *testing.T) {
	t.Parallel()

	hs := &HealthServer{
		rabbitMQConnection: nil,
		logger:             &log.NoneLogger{},
	}

	status := hs.checkRabbitMQ()
	assert.Equal(t, "not_ready", status.Status)
	assert.Equal(t, "connection not configured", status.Message)
}

func TestHealthServer_CheckRabbitMQ_NotConnected(t *testing.T) {
	t.Parallel()

	hs := &HealthServer{
		rabbitMQConnection: &libRabbitMQ.RabbitMQConnection{
			Connected: false,
		},
		logger: &log.NoneLogger{},
	}

	status := hs.checkRabbitMQ()
	assert.Equal(t, "not_ready", status.Status)
	assert.Equal(t, "connection is closed", status.Message)
}

func TestHealthServer_NewHealthServer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		port   string
		conn   *libRabbitMQ.RabbitMQConnection
		logger log.Logger
	}{
		{
			name:   "creates server with default port",
			port:   "4006",
			conn:   nil,
			logger: &log.NoneLogger{},
		},
		{
			name:   "creates server with custom port",
			port:   "9090",
			conn:   &libRabbitMQ.RabbitMQConnection{},
			logger: &log.NoneLogger{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hs := NewHealthServer(tt.port, tt.conn, tt.logger)
			require.NotNil(t, hs)
			require.NotNil(t, hs.server)
			assert.Equal(t, ":"+tt.port, hs.server.Addr)
			assert.Equal(t, tt.conn, hs.rabbitMQConnection)
		})
	}
}

func TestHealthServer_HandleHealth_ResponseFormat(t *testing.T) {
	t.Parallel()

	hs := NewHealthServer("0", nil, &log.NoneLogger{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	hs.handleHealth(rec, req)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestHealthServer_HandleReady_ResponseFormat(t *testing.T) {
	t.Parallel()

	hs := NewHealthServer("0", nil, &log.NoneLogger{})

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	hs.handleReady(rec, req)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestHealthServer_StartAndShutdown(t *testing.T) {
	t.Parallel()

	// Use port 0 so the OS assigns a random available port
	hs := NewHealthServer("0", nil, &log.NoneLogger{})
	require.NotNil(t, hs)

	// Shutdown should not panic even if Start was not called
	hs.Shutdown()
}

func TestHealthServer_Start_UsesPanicRecovery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		expectedName string
	}{
		{
			name:         "Success - Start uses GoNamed with health-server label",
			expectedName: "health-server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var capturedName string

			hs := NewHealthServer("0", nil, &log.NoneLogger{})

			// Override the goroutine launcher to capture the name argument
			hs.goNamedFn = func(_ log.Logger, name string, fn func()) {
				capturedName = name
				// Execute fn in the current goroutine to avoid port binding in tests
				go fn()
			}

			hs.Start()
			hs.Shutdown()

			assert.Equal(t, tt.expectedName, capturedName,
				"Start() must use goNamedFn with the name %q for panic recovery", tt.expectedName)
		})
	}
}
