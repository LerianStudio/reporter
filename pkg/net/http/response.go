// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package http

import (
	"github.com/LerianStudio/reporter/pkg"

	commonsHTTP "github.com/LerianStudio/lib-commons/v2/commons/net/http"
	"github.com/gofiber/fiber/v2"
)

// Unauthorized sends an HTTP 401 Unauthorized response with a custom code, title and message.
// Delegates to lib-commons commonsHTTP.Unauthorized for consistency.
func Unauthorized(c *fiber.Ctx, code, title, message string) error {
	return commonsHTTP.Unauthorized(c, code, title, message)
}

// Forbidden sends an HTTP 403 Forbidden response with a custom code, title and message.
// Delegates to lib-commons commonsHTTP.Forbidden for consistency.
func Forbidden(c *fiber.Ctx, code, title, message string) error {
	return commonsHTTP.Forbidden(c, code, title, message)
}

// BadRequest sends an HTTP 400 Bad Request response with a custom body.
// Delegates to lib-commons commonsHTTP.BadRequest for consistency.
func BadRequest(c *fiber.Ctx, s any) error {
	return commonsHTTP.BadRequest(c, s)
}

// NotFound sends an HTTP 404 Not Found response with a custom code, title and message.
// Delegates to lib-commons commonsHTTP.NotFound for consistency.
func NotFound(c *fiber.Ctx, code, title, message string) error {
	return commonsHTTP.NotFound(c, code, title, message)
}

// Conflict sends an HTTP 409 Conflict response with a custom code, title and message.
// Delegates to lib-commons commonsHTTP.Conflict for consistency.
func Conflict(c *fiber.Ctx, code, title, message string) error {
	return commonsHTTP.Conflict(c, code, title, message)
}

// UnprocessableEntity sends an HTTP 422 Unprocessable Entity response with a custom code, title and message.
// Delegates to lib-commons commonsHTTP.UnprocessableEntity for consistency.
func UnprocessableEntity(c *fiber.Ctx, code, title, message string) error {
	return commonsHTTP.UnprocessableEntity(c, code, title, message)
}

// InternalServerError sends an HTTP 500 Internal Server Error response.
// Delegates to lib-commons commonsHTTP.InternalServerError for consistency.
func InternalServerError(c *fiber.Ctx, code, title, message string) error {
	return commonsHTTP.InternalServerError(c, code, title, message)
}

// JSONResponseError sends a JSON formatted error response with a custom error struct.
// Note: This uses project-level pkg.ResponseError (not commons.Response) because the
// type includes a Code int field for HTTP status, which differs from lib-commons' Response type.
// This is an accepted deviation documented for future migration.
func JSONResponseError(c *fiber.Ctx, err pkg.ResponseError) error {
	return c.Status(err.Code).JSON(err)
}
