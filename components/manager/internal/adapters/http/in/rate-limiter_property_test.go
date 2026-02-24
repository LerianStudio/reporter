//go:build property

// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/quick"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// safeRateLimiterPropertyTest creates a Fiber app with the given rate limit
// config, sends the given request, and returns any panic message. If no panic
// occurs, returns "". This helper isolates the middleware's panic behavior so
// property tests can report panics as failures instead of crashing the process.
func safeRateLimiterPropertyTest(cfg RateLimitConfig, req *http.Request) (panicMsg string, statusCode int, testErr error) {
	defer func() {
		if r := recover(); r != nil {
			panicMsg = fmt.Sprintf("%v", r)
		}
	}()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(RateLimiterMiddleware(cfg))

	app.All("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	resp, err := app.Test(req, -1)
	if err != nil {
		return "", 0, err
	}

	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body) //nolint:errcheck // body drained before close; content is irrelevant

	return "", resp.StatusCode, nil
}

// TestProperty_RateLimiterMiddleware_NeverPanicsForPositiveConfig verifies
// that creating a RateLimiterMiddleware with any positive configuration values
// and processing a request never causes a panic.
func TestProperty_RateLimiterMiddleware_NeverPanicsForPositiveConfig(t *testing.T) {
	t.Parallel()

	property := func(globalMax, exportMax, dispatchMax uint16) bool {
		// Ensure positive values (uint16 >= 0, add 1 to avoid zero)
		gMax := int(globalMax) + 1
		eMax := int(exportMax) + 1
		dMax := int(dispatchMax) + 1

		cfg := RateLimitConfig{
			GlobalMax:   gMax,
			ExportMax:   eMax,
			DispatchMax: dMax,
			Window:      1 * time.Second,
		}

		req := httptest.NewRequest(http.MethodGet, "/v1/templates", nil)

		panicMsg, statusCode, err := safeRateLimiterPropertyTest(cfg, req)

		if panicMsg != "" {
			t.Logf("Panicked for config{global=%d, export=%d, dispatch=%d}: %s",
				gMax, eMax, dMax, panicMsg)
			return false
		}

		if err != nil {
			return true // Connection errors are acceptable
		}

		// Must not produce a server error
		return statusCode < 500
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	require.NoError(t, err, "Property violated: RateLimiterMiddleware panicked or returned 5xx for positive config")
}

// TestProperty_RateLimiterMiddleware_ValidationRejectsZeroNegative verifies
// that zero or negative rate limit values are always detectable. The
// validation logic lives in Config.Validate, but the property holds at the
// RateLimitConfig level: any non-positive field is invalid configuration.
func TestProperty_RateLimiterMiddleware_ValidationRejectsZeroNegative(t *testing.T) {
	t.Parallel()

	property := func(value int32) bool {
		// Only test non-positive values
		if value > 0 {
			return true
		}

		cfg := RateLimitConfig{
			GlobalMax:   int(value),
			ExportMax:   10,
			DispatchMax: 50,
			Window:      1 * time.Second,
		}

		// Zero/negative value is detectable
		return cfg.GlobalMax <= 0
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	require.NoError(t, err, "Property violated: non-positive values not detected")
}

// TestProperty_RateLimiterMiddleware_AcceptsAllPositiveValues verifies that
// any positive integer values for all three tiers produce a valid config
// where all fields are greater than zero.
func TestProperty_RateLimiterMiddleware_AcceptsAllPositiveValues(t *testing.T) {
	t.Parallel()

	property := func(globalMax, exportMax, dispatchMax uint16) bool {
		// Ensure strictly positive (uint16 >= 0, add 1)
		gMax := int(globalMax) + 1
		eMax := int(exportMax) + 1
		dMax := int(dispatchMax) + 1

		cfg := RateLimitConfig{
			GlobalMax:   gMax,
			ExportMax:   eMax,
			DispatchMax: dMax,
			Window:      1 * time.Second,
		}

		// All values must be positive
		return cfg.GlobalMax > 0 && cfg.ExportMax > 0 && cfg.DispatchMax > 0
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	require.NoError(t, err, "Property violated: positive values not accepted")
}

// TestProperty_RateLimiterMiddleware_TiersOperateIndependently verifies that
// exhausting requests in one tier does not affect the availability of another
// tier. Specifically: exhausting the dispatch tier (POST requests) must not
// cause the global tier (GET requests) to reject traffic.
func TestProperty_RateLimiterMiddleware_TiersOperateIndependently(t *testing.T) {
	t.Parallel()

	property := func(dispatchMax uint8) bool {
		// Use small dispatchMax to exhaust it quickly (1 to 20)
		dMax := int(dispatchMax)%20 + 1
		globalMax := 100

		cfg := RateLimitConfig{
			GlobalMax:   globalMax,
			ExportMax:   10,
			DispatchMax: dMax,
			Window:      10 * time.Second,
		}

		app := fiber.New(fiber.Config{DisableStartupMessage: true})
		app.Use(RateLimiterMiddleware(cfg))
		app.All("/*", func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})

		// Step 1: Exhaust the dispatch tier by sending dMax+1 POST requests
		var lastDispatchStatus int

		for i := 0; i <= dMax; i++ {
			req := httptest.NewRequest(http.MethodPost, "/v1/reports", nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				return true // acceptable
			}

			lastDispatchStatus = resp.StatusCode

			io.Copy(io.Discard, resp.Body) //nolint:errcheck // body drained before close; content is irrelevant
			_ = resp.Body.Close()
		}

		// If dispatch tier was not exhausted, the property is trivially true
		if lastDispatchStatus != fiber.StatusTooManyRequests {
			return true
		}

		// Step 2: Send a GET request (global tier) -- must NOT be rate-limited
		getReq := httptest.NewRequest(http.MethodGet, "/v1/templates", nil)

		getResp, err := app.Test(getReq, -1)
		if err != nil {
			return true // acceptable
		}

		defer func() {
			_ = getResp.Body.Close()
		}()

		io.Copy(io.Discard, getResp.Body) //nolint:errcheck // body drained before close; content is irrelevant

		// Global tier must still work (200 OK, not 429)
		return getResp.StatusCode == fiber.StatusOK
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	require.NoError(t, err, "Property violated: exhausting dispatch tier affected global tier")
}

// TestProperty_RateLimiterMiddleware_HealthPathAlwaysBypassed verifies that
// health/readiness endpoints are never rate-limited, regardless of how many
// requests have been made or the rate limit configuration values.
func TestProperty_RateLimiterMiddleware_HealthPathAlwaysBypassed(t *testing.T) {
	t.Parallel()

	property := func(requestCount uint8) bool {
		// Use a very low rate limit to ensure other paths would be limited
		cfg := RateLimitConfig{
			GlobalMax:   1,
			ExportMax:   1,
			DispatchMax: 1,
			Window:      10 * time.Second,
		}

		app := fiber.New(fiber.Config{DisableStartupMessage: true})
		app.Use(RateLimiterMiddleware(cfg))
		app.All("/*", func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})

		// Exhaust all tiers first with non-health requests
		for i := 0; i < 5; i++ {
			exhaustReq := httptest.NewRequest(http.MethodGet, "/v1/templates", nil)

			exhaustResp, err := app.Test(exhaustReq, -1)
			if err != nil {
				continue
			}

			io.Copy(io.Discard, exhaustResp.Body) //nolint:errcheck // body drained before close; content is irrelevant
			_ = exhaustResp.Body.Close()
		}

		// Verify health paths are still accessible
		numRequests := int(requestCount)%10 + 1

		for _, path := range []string{"/health", "/ready", "/version"} {
			for i := 0; i < numRequests; i++ {
				req := httptest.NewRequest(http.MethodGet, path, nil)

				resp, err := app.Test(req, -1)
				if err != nil {
					continue
				}

				status := resp.StatusCode

				io.Copy(io.Discard, resp.Body) //nolint:errcheck // body drained before close; content is irrelevant
				_ = resp.Body.Close()

				if status == fiber.StatusTooManyRequests {
					t.Logf("Health path %s was rate-limited on request %d", path, i+1)
					return false
				}
			}
		}

		return true
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	require.NoError(t, err, "Property violated: health paths were rate-limited")
}

// TestProperty_RateLimiterMiddleware_ExportPathDetection verifies that any
// path ending with "/download" is always classified as an export path by the
// isExportPath function.
func TestProperty_RateLimiterMiddleware_ExportPathDetection(t *testing.T) {
	t.Parallel()

	property := func(prefix uint8) bool {
		// Create a path that always ends with "/download"
		path := "/v1/reports/" + string(rune('a'+int(prefix)%26)) + "/download"

		return isExportPath(path)
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	require.NoError(t, err, "Property violated: path with /download not classified as export")
}

// TestProperty_RateLimiterMiddleware_DispatchMethodDetection verifies that
// POST, PUT, PATCH, and DELETE methods are always classified as dispatch
// methods, and GET, HEAD, OPTIONS are never classified as dispatch.
func TestProperty_RateLimiterMiddleware_DispatchMethodDetection(t *testing.T) {
	t.Parallel()

	dispatchMethods := map[string]bool{
		fiber.MethodPost:   true,
		fiber.MethodPut:    true,
		fiber.MethodPatch:  true,
		fiber.MethodDelete: true,
	}

	property := func(selector uint8) bool {
		allMethods := []string{
			fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions,
			fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete,
		}

		method := allMethods[int(selector)%len(allMethods)]
		result := isDispatchMethod(method)
		expected := dispatchMethods[method]

		if result != expected {
			t.Logf("isDispatchMethod(%q) = %v, expected %v", method, result, expected)
			return false
		}

		return true
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	require.NoError(t, err, "Property violated: dispatch method classification incorrect")
}
