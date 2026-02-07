// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
)

var UUIDPathParameter = "id"

// SecurityHeaders returns a Fiber middleware that sets standard HTTP security
// headers on every response. The headers mitigate common browser-side attacks
// such as MIME-type sniffing, clickjacking, and reflected XSS.
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "0")

		return c.Next()
	}
}

// RecoverMiddleware returns a Fiber middleware that recovers from panics
// inside handlers, preventing a single panicking request from crashing the
// entire server process. It delegates to Fiber's built-in recover middleware.
func RecoverMiddleware() fiber.Handler {
	return recover.New()
}

// ParsePathParametersUUID convert and validate if the path parameter is UUID
func ParsePathParametersUUID(c *fiber.Ctx) error {
	pathParam := c.Params(UUIDPathParameter)

	if commons.IsNilOrEmpty(&pathParam) {
		err := pkg.ValidateBusinessError(constant.ErrInvalidPathParameter, "", UUIDPathParameter)
		return http.WithError(c, err)
	}

	parsedPathUUID, errPath := uuid.Parse(pathParam)
	if errPath != nil {
		err := pkg.ValidateBusinessError(constant.ErrInvalidPathParameter, "", UUIDPathParameter)
		return http.WithError(c, err)
	}

	c.Locals(UUIDPathParameter, parsedPathUUID)

	return c.Next()
}
