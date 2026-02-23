//go:build fuzz

// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// safeRateLimiterTest creates a Fiber app with the given rate limit config,
// builds a request from method+path, and returns any panic message. If no
// panic occurs, returns "". This helper isolates the middleware's panic
// behavior so fuzz tests can report panics instead of crashing.
func safeRateLimiterTest(cfg RateLimitConfig, method, path string) (panicMsg string, statusCode int, testErr error) {
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

	// Use http.NewRequest (returns error) instead of httptest.NewRequest (panics)
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		return "", 0, err
	}

	defer resp.Body.Close()

	_, _ = io.ReadAll(resp.Body)

	return "", resp.StatusCode, nil
}

// FuzzRateLimiterMiddleware_Config tests that the RateLimiterMiddleware never
// panics regardless of the configuration values provided. The fuzzer generates
// random integers for GlobalMax, ExportMax, and DispatchMax. The middleware
// must initialise and handle a request without crashing.
func FuzzRateLimiterMiddleware_Config(f *testing.F) {
	// Seed corpus: 7 entries across all required categories
	// Category 1 (Valid): default production values
	f.Add(100, 10, 50, "/v1/templates", "GET")
	// Category 2 (Boundary): minimum positive values
	f.Add(1, 1, 1, "/v1/reports", "POST")
	// Category 3 (Boundary): large values
	f.Add(999999, 999999, 999999, "/v1/reports/abc/download", "GET")
	// Category 4 (Empty/zero): all zeros
	f.Add(0, 0, 0, "/health", "GET")
	// Category 5 (Negative): negative values
	f.Add(-1, -10, -100, "/ready", "GET")
	// Category 6 (Mixed): one zero, others positive
	f.Add(50, 0, 25, "/v1/templates", "DELETE")
	// Category 7 (Unicode path): international characters
	f.Add(10, 5, 20, "/v1/\u30c6\u30b9\u30c8", "PATCH")

	f.Fuzz(func(t *testing.T, globalMax, exportMax, dispatchMax int, path, method string) {
		// Bound inputs to prevent resource exhaustion
		if len(path) > 256 {
			path = path[:256]
		}

		if len(method) > 16 {
			method = method[:16]
		}

		// Ensure path is a valid URI target (must start with "/" for HTTP requests)
		if path == "" || path[0] != '/' {
			path = "/" + path
		}

		// Skip empty methods (http.NewRequest rejects them)
		if method == "" {
			method = "GET"
		}

		cfg := RateLimitConfig{
			GlobalMax:   globalMax,
			ExportMax:   exportMax,
			DispatchMax: dispatchMax,
			Window:      1 * time.Second,
		}

		panicMsg, statusCode, err := safeRateLimiterTest(cfg, method, path)

		if panicMsg != "" {
			t.Fatalf("RateLimiterMiddleware panicked for config{global=%d, export=%d, dispatch=%d} path=%q method=%q: %s",
				globalMax, exportMax, dispatchMax, path, method, panicMsg)
		}

		if err != nil {
			// Request construction or connection errors are acceptable for fuzzed inputs
			return
		}

		// The middleware must never produce a 5xx server error from configuration issues
		if statusCode >= 500 {
			t.Fatalf("server error %d for config{global=%d, export=%d, dispatch=%d} path=%q method=%q",
				statusCode, globalMax, exportMax, dispatchMax, path, method)
		}
	})
}

// FuzzRateLimiterMiddleware_Path tests that the middleware correctly handles
// any path string without panicking. The fuzzer generates random paths to
// exercise the health-check, export, and dispatch routing logic.
func FuzzRateLimiterMiddleware_Path(f *testing.F) {
	// Seed corpus: 7 entries across all required categories
	// Category 1 (Valid): health path
	f.Add("/health")
	// Category 2 (Valid): ready path
	f.Add("/ready")
	// Category 3 (Valid): export/download path
	f.Add("/v1/reports/abc/download")
	// Category 4 (Valid): normal API path
	f.Add("/v1/templates")
	// Category 5 (Empty/nil): root path
	f.Add("/")
	// Category 6 (Security): SQL injection in path
	f.Add("/v1/reports/' OR 1=1 --/download")
	// Category 7 (Security): path traversal
	f.Add("/../../../etc/passwd")

	f.Fuzz(func(t *testing.T, path string) {
		if len(path) > 1024 {
			path = path[:1024]
		}

		// Ensure path is a valid URI target
		if path == "" || path[0] != '/' {
			path = "/" + path
		}

		cfg := RateLimitConfig{
			GlobalMax:   100,
			ExportMax:   10,
			DispatchMax: 50,
			Window:      1 * time.Second,
		}

		panicMsg, statusCode, err := safeRateLimiterTest(cfg, http.MethodGet, path)

		if panicMsg != "" {
			t.Fatalf("RateLimiterMiddleware panicked for path=%q: %s", path, panicMsg)
		}

		if err != nil {
			return
		}

		if statusCode >= 500 {
			t.Fatalf("server error %d for path=%q", statusCode, path)
		}
	})
}

// FuzzRateLimiterMiddleware_Method tests that the middleware correctly
// classifies HTTP methods for dispatch tier routing. The fuzzer generates
// random method strings to exercise isDispatchMethod logic.
func FuzzRateLimiterMiddleware_Method(f *testing.F) {
	// Seed corpus: 7 entries across all required categories
	// Category 1 (Valid): standard read method
	f.Add("GET")
	// Category 2 (Valid): dispatch methods
	f.Add("POST")
	f.Add("PUT")
	// Category 3 (Valid): more dispatch methods
	f.Add("PATCH")
	f.Add("DELETE")
	// Category 4 (Invalid): unknown HTTP method
	f.Add("INVALID")
	// Category 5 (Boundary): OPTIONS method
	f.Add("OPTIONS")

	f.Fuzz(func(t *testing.T, method string) {
		if len(method) > 64 {
			method = method[:64]
		}

		// Skip empty methods (not valid HTTP, http.NewRequest rejects them)
		if method == "" {
			return
		}

		cfg := RateLimitConfig{
			GlobalMax:   100,
			ExportMax:   10,
			DispatchMax: 50,
			Window:      1 * time.Second,
		}

		panicMsg, statusCode, err := safeRateLimiterTest(cfg, method, "/v1/templates")

		if panicMsg != "" {
			t.Fatalf("RateLimiterMiddleware panicked for method=%q: %s", method, panicMsg)
		}

		if err != nil {
			return
		}

		if statusCode >= 500 {
			t.Fatalf("server error %d for method=%q", statusCode, method)
		}
	})
}
