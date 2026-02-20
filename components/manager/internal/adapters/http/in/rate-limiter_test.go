// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRateLimiterMiddleware_ReturnsHandler verifies that RateLimiterMiddleware
// returns a valid Fiber handler that can be mounted on an app.
func TestRateLimiterMiddleware_ReturnsHandler(t *testing.T) {
	t.Parallel()

	cfg := RateLimitConfig{
		Enabled:     true,
		GlobalMax:   100,
		ExportMax:   10,
		DispatchMax: 50,
		Window:      60 * time.Second,
	}

	handler := RateLimiterMiddleware(cfg)
	require.NotNil(t, handler, "RateLimiterMiddleware must return a non-nil handler")
}

// TestRateLimiterMiddleware_DisabledPassthrough verifies that when Enabled is
// false, the middleware passes all requests through without rate limiting.
func TestRateLimiterMiddleware_DisabledPassthrough(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	cfg := RateLimitConfig{
		Enabled:     false,
		GlobalMax:   1,
		ExportMax:   1,
		DispatchMax: 1,
		Window:      60 * time.Second,
	}

	app.Use(RateLimiterMiddleware(cfg))
	app.Get("/v1/templates", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	// Send many requests -- all should succeed even with limit=1
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/templates", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Request %d should pass through when rate limiting is disabled", i+1)
	}
}

// TestRateLimiterMiddleware_AllowsRequestsWithinLimit verifies that requests
// within the configured limit receive a 200 OK response, not 429.
func TestRateLimiterMiddleware_AllowsRequestsWithinLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		path   string
		method string
		limit  int
	}{
		{
			name:   "Global limit - GET templates within limit",
			path:   "/v1/templates",
			method: http.MethodGet,
			limit:  100,
		},
		{
			name:   "Global limit - GET reports within limit",
			path:   "/v1/reports",
			method: http.MethodGet,
			limit:  100,
		},
		{
			name:   "Export limit - GET download within limit",
			path:   "/v1/reports/123/download",
			method: http.MethodGet,
			limit:  10,
		},
		{
			name:   "Dispatch limit - POST reports within limit",
			path:   "/v1/reports",
			method: http.MethodPost,
			limit:  50,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			cfg := RateLimitConfig{
				Enabled:     true,
				GlobalMax:   100,
				ExportMax:   10,
				DispatchMax: 50,
				Window:      60 * time.Second,
			}

			app.Use(RateLimiterMiddleware(cfg))
			app.Add(tt.method, tt.path, func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp, err := app.Test(req)

			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode,
				"Request within limit should receive 200, not rate-limited")
		})
	}
}

// TestRateLimiterMiddleware_Returns429WhenLimitExceeded verifies that
// the rate limiter returns HTTP 429 Too Many Requests when the configured
// limit is exceeded.
func TestRateLimiterMiddleware_Returns429WhenLimitExceeded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		method      string
		limit       int
		requestsFn  func(cfg RateLimitConfig) RateLimitConfig
		description string
	}{
		{
			name:   "Global tier - exceeds 3 req limit",
			path:   "/v1/templates",
			method: http.MethodGet,
			limit:  3,
			requestsFn: func(cfg RateLimitConfig) RateLimitConfig {
				cfg.GlobalMax = 3
				return cfg
			},
			description: "Global rate limit should reject after 3 requests",
		},
		{
			name:   "Export tier - exceeds 2 req limit on download",
			path:   "/v1/reports/test-id/download",
			method: http.MethodGet,
			limit:  2,
			requestsFn: func(cfg RateLimitConfig) RateLimitConfig {
				cfg.ExportMax = 2
				return cfg
			},
			description: "Export rate limit should reject download after 2 requests",
		},
		{
			name:   "Dispatch tier - exceeds 2 req limit on create",
			path:   "/v1/reports",
			method: http.MethodPost,
			limit:  2,
			requestsFn: func(cfg RateLimitConfig) RateLimitConfig {
				cfg.DispatchMax = 2
				return cfg
			},
			description: "Dispatch rate limit should reject create after 2 requests",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			baseCfg := RateLimitConfig{
				Enabled:     true,
				GlobalMax:   100,
				ExportMax:   10,
				DispatchMax: 50,
				Window:      60 * time.Second,
			}
			cfg := tt.requestsFn(baseCfg)

			app.Use(RateLimiterMiddleware(cfg))
			app.Add(tt.method, tt.path, func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			// Send requests up to the limit -- all should succeed
			for i := 0; i < tt.limit; i++ {
				req := httptest.NewRequest(tt.method, tt.path, nil)
				resp, err := app.Test(req)

				require.NoError(t, err)
				resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode,
					"Request %d of %d should succeed", i+1, tt.limit)
			}

			// The next request should be rate-limited (429)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp, err := app.Test(req)

			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, tt.description)
		})
	}
}

// TestRateLimiterMiddleware_DifferentTiersHaveDifferentLimits verifies that
// different route groups (global, export, dispatch) apply independent limits,
// so exhausting one tier does not affect another.
func TestRateLimiterMiddleware_DifferentTiersHaveDifferentLimits(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	cfg := RateLimitConfig{
		Enabled:     true,
		GlobalMax:   100,
		ExportMax:   2,
		DispatchMax: 50,
		Window:      60 * time.Second,
	}

	app.Use(RateLimiterMiddleware(cfg))

	app.Get("/v1/reports/:id/download", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})
	app.Get("/v1/templates", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	// Exhaust the export tier (2 requests)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/reports/abc/download", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		resp.Body.Close()
	}

	// Export tier should be exhausted -- 429
	req := httptest.NewRequest(http.MethodGet, "/v1/reports/abc/download", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode,
		"Export tier should be rate-limited after 2 requests")

	// Global tier should still allow requests to other routes
	req = httptest.NewRequest(http.MethodGet, "/v1/templates", nil)
	resp, err = app.Test(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode,
		"Global tier should still work even when export tier is exhausted")
}

// TestRateLimiterMiddleware_HealthEndpointExcluded verifies that health
// check endpoints are not rate-limited.
func TestRateLimiterMiddleware_HealthEndpointExcluded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{
			name: "Health endpoint excluded from rate limiting",
			path: "/health",
		},
		{
			name: "Ready endpoint excluded from rate limiting",
			path: "/ready",
		},
		{
			name: "Version endpoint excluded from rate limiting",
			path: "/version",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			// Use extremely low limit to prove exclusion
			cfg := RateLimitConfig{
				Enabled:     true,
				GlobalMax:   1,
				ExportMax:   1,
				DispatchMax: 1,
				Window:      60 * time.Second,
			}

			app.Use(RateLimiterMiddleware(cfg))
			app.Get(tt.path, func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			// Send multiple requests -- all should succeed (not rate-limited)
			for i := 0; i < 5; i++ {
				req := httptest.NewRequest(http.MethodGet, tt.path, nil)
				resp, err := app.Test(req)

				require.NoError(t, err)
				resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode,
					"Request %d to %s should not be rate-limited", i+1, tt.path)
			}
		})
	}
}

// TestRateLimiterMiddleware_429ResponseContainsRetryAfterHeader verifies
// that rate-limited responses include the Retry-After header.
func TestRateLimiterMiddleware_429ResponseContainsRetryAfterHeader(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	cfg := RateLimitConfig{
		Enabled:     true,
		GlobalMax:   1,
		ExportMax:   1,
		DispatchMax: 1,
		Window:      60 * time.Second,
	}

	app.Use(RateLimiterMiddleware(cfg))
	app.Get("/v1/templates", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	// First request succeeds
	req := httptest.NewRequest(http.MethodGet, "/v1/templates", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	resp.Body.Close()

	// Second request should be rate-limited with Retry-After header
	req = httptest.NewRequest(http.MethodGet, "/v1/templates", nil)
	resp, err = app.Test(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	retryAfter := resp.Header.Get("Retry-After")
	assert.NotEmpty(t, retryAfter,
		"429 response must include Retry-After header")
	assert.Regexp(t, `^[0-9]+$`, retryAfter,
		"Retry-After must be numeric seconds")
}

// TestRateLimiterMiddleware_429ResponseContainsStructuredJSON verifies that
// rate-limited responses return a structured JSON body with code, title, and message.
func TestRateLimiterMiddleware_429ResponseContainsStructuredJSON(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	cfg := RateLimitConfig{
		Enabled:     true,
		GlobalMax:   1,
		ExportMax:   1,
		DispatchMax: 1,
		Window:      60 * time.Second,
	}

	app.Use(RateLimiterMiddleware(cfg))
	app.Get("/v1/templates", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	// First request succeeds
	req := httptest.NewRequest(http.MethodGet, "/v1/templates", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	resp.Body.Close()

	// Second request should be rate-limited with JSON body
	req = httptest.NewRequest(http.MethodGet, "/v1/templates", nil)
	resp, err = app.Test(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var errResp map[string]string
	err = json.Unmarshal(body, &errResp)
	require.NoError(t, err, "429 response body must be valid JSON")

	assert.Equal(t, "TPL-0429", errResp["code"],
		"429 response must have code TPL-0429")
	assert.Equal(t, "Too Many Requests", errResp["title"],
		"429 response must have title 'Too Many Requests'")
	assert.NotEmpty(t, errResp["message"],
		"429 response must have a non-empty message")
}

// NOTE: Zero-value RateLimitConfig is rejected at the application layer by
// Config.validateRateLimitBounds() in the bootstrap package. That validation
// ensures zero/negative values never reach the middleware. Tests for this
// boundary check live in bootstrap/config_test.go (TestConfig_Validate_RateLimitBounds).
