// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimitConfig holds the configuration for the three-tier rate limiter.
// Each tier applies independent limits to different route categories:
//   - Enabled: feature toggle; when false, the middleware is a passthrough
//   - GlobalMax: catch-all limit for general API requests
//   - ExportMax: limit for resource-intensive export/download operations
//   - DispatchMax: limit for write operations (POST, PUT, PATCH, DELETE)
//   - Storage: optional fiber.Storage backend (e.g. Redis) for distributed counting
type RateLimitConfig struct {
	Enabled     bool
	GlobalMax   int
	ExportMax   int
	DispatchMax int
	Window      time.Duration
	Storage     RateLimitStorage
}

// healthPaths lists endpoints excluded from rate limiting.
// These paths must always remain accessible for orchestration and monitoring.
var healthPaths = map[string]bool{
	"/health":  true,
	"/ready":   true,
	"/version": true,
}

// isHealthPath returns true if the given path is a health/readiness/version endpoint.
func isHealthPath(path string) bool {
	return healthPaths[path]
}

// isExportPath returns true if the request path ends with "/download",
// matching known export/download route patterns. Uses HasSuffix instead
// of Contains to avoid false positives on paths that merely contain
// "/download" as a substring (e.g., "/v1/download-configs/abc").
func isExportPath(path string) bool {
	return strings.HasSuffix(path, "/download")
}

// isDispatchMethod returns true if the HTTP method is a write operation.
func isDispatchMethod(method string) bool {
	return method == fiber.MethodPost ||
		method == fiber.MethodPut ||
		method == fiber.MethodPatch ||
		method == fiber.MethodDelete
}

// RateLimiterMiddleware returns a Fiber handler that enforces three independent
// rate limit tiers based on request path and HTTP method. Each tier maintains
// its own counter, so exhausting one tier does not affect the others.
//
// When cfg.Enabled is false, returns a passthrough handler that calls c.Next().
//
// Tier selection logic:
//  1. Health/ready/version endpoints: bypassed entirely (no rate limiting)
//  2. Paths ending with "/download": export tier (ExportMax)
//  3. POST/PUT/PATCH/DELETE methods: dispatch tier (DispatchMax)
//  4. All other requests: global tier (GlobalMax)
//
// Rate-limited responses return HTTP 429 with a structured JSON body and
// a Retry-After header indicating when the client may retry.
func RateLimiterMiddleware(cfg RateLimitConfig) fiber.Handler {
	if !cfg.Enabled {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	limitReached := newRateLimitReachedHandler(cfg.Window)

	globalLimiter := limiter.New(limiter.Config{
		Max:        cfg.GlobalMax,
		Expiration: cfg.Window,
		Storage:    cfg.Storage,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "global:" + c.IP()
		},
		LimitReached: limitReached,
	})

	exportLimiter := limiter.New(limiter.Config{
		Max:        cfg.ExportMax,
		Expiration: cfg.Window,
		Storage:    cfg.Storage,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "export:" + c.IP()
		},
		LimitReached: limitReached,
	})

	dispatchLimiter := limiter.New(limiter.Config{
		Max:        cfg.DispatchMax,
		Expiration: cfg.Window,
		Storage:    cfg.Storage,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "dispatch:" + c.IP()
		},
		LimitReached: limitReached,
	})

	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Health endpoints are never rate-limited
		if isHealthPath(path) {
			return c.Next()
		}

		// Route to the appropriate tier
		if isExportPath(path) {
			return exportLimiter(c)
		}

		if isDispatchMethod(c.Method()) {
			return dispatchLimiter(c)
		}

		return globalLimiter(c)
	}
}

// rateLimitErrorResponse is the structured JSON body returned when a rate
// limit tier is exhausted. It follows the project's standard error envelope.
type rateLimitErrorResponse struct {
	Code    string `json:"code"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// newRateLimitReachedHandler returns a handler that is called when a rate limit
// tier is exhausted. It sets the Retry-After header using the configured window
// duration (in seconds) and returns HTTP 429 with a structured JSON body.
func newRateLimitReachedHandler(window time.Duration) fiber.Handler {
	retryAfterSeconds := fmt.Sprintf("%d", int(window.Seconds()))

	return func(c *fiber.Ctx) error {
		retryAfter := c.GetRespHeader("Retry-After")
		if retryAfter == "" {
			c.Set("Retry-After", retryAfterSeconds)
		}

		return c.Status(fiber.StatusTooManyRequests).JSON(rateLimitErrorResponse{
			Code:    "TPL-0429",
			Title:   "Too Many Requests",
			Message: "Rate limit exceeded. Please retry after " + retryAfterSeconds + " seconds.",
		})
	}
}
