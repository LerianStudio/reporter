// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"strings"
	"time"

	"github.com/LerianStudio/reporter/pkg/constant"

	"github.com/LerianStudio/lib-commons/v3/commons/log"
	tmclient "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/client"
	tmmiddleware "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/middleware"
	tmmongo "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/mongo"
	"github.com/gofiber/fiber/v2"
)

// tenantBypassPaths lists HTTP path prefixes that must bypass tenant resolution.
// Health, readiness, swagger, and version endpoints are infrastructure endpoints
// that do not carry a tenant JWT and must never be blocked by the tenant middleware.
var tenantBypassPaths = []string{
	"/health",
	"/ready",
	"/swagger",
	"/version",
}

// initTenantMiddleware constructs the TenantMiddleware for the manager component.
//
// Returns nil (no-op) in two cases:
//   - MultiTenantEnabled is false: single-tenant passthrough, zero performance impact.
//   - MultiTenantURL is empty: URL required to reach the Tenant Manager API.
//
// When enabled, it creates:
//  1. A Tenant Manager HTTP client pointing at cfg.MultiTenantURL with circuit breaker.
//  2. A MongoDB connection manager scoped to ApplicationName + ModuleManager.
//  3. A TenantMiddleware that resolves per-tenant MongoDB connections from JWT.
//
// The returned fiber.Handler wraps WithTenantDB with a bypass check so that
// /health, /ready, /swagger, and /version skip tenant resolution entirely.
func initTenantMiddleware(cfg *Config, logger log.Logger) fiber.Handler {
	if !cfg.MultiTenantEnabled || cfg.MultiTenantURL == "" {
		return nil
	}

	var clientOpts []tmclient.ClientOption

	if cfg.MultiTenantCircuitBreakerThreshold > 0 {
		cbTimeout := time.Duration(cfg.MultiTenantCircuitBreakerTimeoutSec) * time.Second
		clientOpts = append(clientOpts,
			tmclient.WithCircuitBreaker(
				cfg.MultiTenantCircuitBreakerThreshold,
				cbTimeout,
			),
		)
	}

	tmClient := tmclient.NewClient(cfg.MultiTenantURL, logger, clientOpts...)

	mongoManager := tmmongo.NewManager(
		tmClient,
		constant.ApplicationName,
		tmmongo.WithModule(constant.ModuleManager),
		tmmongo.WithLogger(logger),
		tmmongo.WithMaxTenantPools(cfg.MultiTenantMaxTenantPools),
		tmmongo.WithIdleTimeout(time.Duration(cfg.MultiTenantIdleTimeoutSec)*time.Second),
	)

	tenantMid := tmmiddleware.NewTenantMiddleware(
		tmmiddleware.WithMongoManager(mongoManager),
	)

	return func(c *fiber.Ctx) error {
		// Bypass tenant resolution for infrastructure endpoints.
		// /swagger uses prefix match because it serves multiple sub-paths (e.g. /swagger/index.html).
		// All other bypass paths use exact match to prevent path-prefix bypass attacks.
		path := c.Path()

		for _, prefix := range tenantBypassPaths {
			if prefix == "/swagger" {
				if strings.HasPrefix(path, prefix) {
					return c.Next()
				}
			} else {
				if path == prefix {
					return c.Next()
				}
			}
		}

		return tenantMid.WithTenantDB(c)
	}
}
