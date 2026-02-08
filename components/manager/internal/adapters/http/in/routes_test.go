// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedHeader string
		expectedValue  string
	}{
		{
			name:           "X-Content-Type-Options header is set to nosniff",
			method:         http.MethodGet,
			path:           "/health",
			expectedHeader: "X-Content-Type-Options",
			expectedValue:  "nosniff",
		},
		{
			name:           "X-Frame-Options header is set to DENY",
			method:         http.MethodGet,
			path:           "/health",
			expectedHeader: "X-Frame-Options",
			expectedValue:  "DENY",
		},
		{
			name:           "X-XSS-Protection header is set to 0",
			method:         http.MethodGet,
			path:           "/health",
			expectedHeader: "X-XSS-Protection",
			expectedValue:  "0",
		},
		{
			name:           "Strict-Transport-Security header is set with max-age and includeSubDomains",
			method:         http.MethodGet,
			path:           "/health",
			expectedHeader: "Strict-Transport-Security",
			expectedValue:  "max-age=31536000; includeSubDomains",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			// Apply SecurityHeaders middleware (does not exist yet - test must FAIL)
			app.Use(SecurityHeaders())

			app.Get("/health", func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp, err := app.Test(req)

			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			headerValue := resp.Header.Get(tt.expectedHeader)
			assert.Equal(t, tt.expectedValue, headerValue,
				"Expected header %s to be %q, got %q",
				tt.expectedHeader, tt.expectedValue, headerValue,
			)
		})
	}
}

func TestSecurityHeaders_AllHeadersOnSingleResponse(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Apply SecurityHeaders middleware (does not exist yet - test must FAIL)
	app.Use(SecurityHeaders())

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// All four security headers must be present on a single response
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"),
		"X-Content-Type-Options must be nosniff")
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"),
		"X-Frame-Options must be DENY")
	assert.Equal(t, "0", resp.Header.Get("X-XSS-Protection"),
		"X-XSS-Protection must be 0")
	assert.Equal(t, "max-age=31536000; includeSubDomains", resp.Header.Get("Strict-Transport-Security"),
		"Strict-Transport-Security must enforce HSTS with includeSubDomains")
}

func TestRecoverMiddleware_PanicDoesNotCrashServer(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Apply RecoverMiddleware (does not exist yet - test must FAIL).
	// This function should wrap Fiber's recover.New() so we can test
	// that it is wired into the middleware stack.
	app.Use(RecoverMiddleware())

	app.Get("/panic", func(_ *fiber.Ctx) error {
		panic("intentional test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	resp, err := app.Test(req)

	require.NoError(t, err, "Server must not crash on panic when recover middleware is present")
	defer resp.Body.Close()

	// Fiber's recover middleware returns 500 Internal Server Error by default
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode,
		"Recovered panic should return 500, not crash the server")
}

func TestRecoverMiddleware_NonPanicRouteUnaffected(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Apply RecoverMiddleware (does not exist yet - test must FAIL)
	app.Use(RecoverMiddleware())

	app.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode,
		"Non-panicking routes must work normally with recover middleware")
}
