// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package http

import (
	"net/http"

	"github.com/LerianStudio/reporter/pkg"

	"github.com/gofiber/fiber/v2"
)

// Unauthorized sends an HTTP 401 Unauthorized response with a custom code, title and message.
func Unauthorized(c *fiber.Ctx, code, title, message string) error {
	return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
		"code":    code,
		"title":   title,
		"message": message,
	})
}

// Forbidden sends an HTTP 403 Forbidden response with a custom code, title and message.
func Forbidden(c *fiber.Ctx, code, title, message string) error {
	return c.Status(http.StatusForbidden).JSON(fiber.Map{
		"code":    code,
		"title":   title,
		"message": message,
	})
}

// BadRequest sends an HTTP 400 Bad Request response with a custom body.
func BadRequest(c *fiber.Ctx, s any) error {
	return c.Status(http.StatusBadRequest).JSON(s)
}

// NotFound sends an HTTP 404 Not Found response with a custom code, title and message.
func NotFound(c *fiber.Ctx, code, title, message string) error {
	return c.Status(http.StatusNotFound).JSON(fiber.Map{
		"code":    code,
		"title":   title,
		"message": message,
	})
}

// Conflict sends an HTTP 409 Conflict response with a custom code, title and message.
func Conflict(c *fiber.Ctx, code, title, message string) error {
	return c.Status(http.StatusConflict).JSON(fiber.Map{
		"code":    code,
		"title":   title,
		"message": message,
	})
}

// UnprocessableEntity sends an HTTP 422 Unprocessable Entity response with a custom code, title and message.
func UnprocessableEntity(c *fiber.Ctx, code, title, message string) error {
	return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
		"code":    code,
		"title":   title,
		"message": message,
	})
}

// InternalServerError sends an HTTP 500 Internal Server Error response
func InternalServerError(c *fiber.Ctx, code, title, message string) error {
	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
		"code":    code,
		"title":   title,
		"message": message,
	})
}

// JSONResponseError sends a JSON formatted error response with a custom error struct.
func JSONResponseError(c *fiber.Ctx, err pkg.ResponseError) error {
	return c.Status(err.Code).JSON(err)
}
