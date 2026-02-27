// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

//go:build unit

package bootstrap

import (
	libLog "github.com/LerianStudio/lib-commons/v3/commons/log"
	"github.com/gofiber/fiber/v2"
)

// initTenantMiddlewareForTest exercises initTenantMiddleware with the given
// enabled flag and URL, returning the fiber.Handler or nil.
// Uses log.NoneLogger to avoid requiring a real logging backend in unit tests.
func initTenantMiddlewareForTest(enabled bool, url string) fiber.Handler {
	cfg := &Config{
		MultiTenantEnabled: enabled,
		MultiTenantURL:     url,
	}

	return initTenantMiddleware(cfg, &libLog.NoneLogger{})
}
