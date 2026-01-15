package http

import (
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestWithError_EntityNotFoundError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return WithError(c, pkg.EntityNotFoundError{
			Code:    "NOT-001",
			Title:   "Not Found",
			Message: "Entity not found",
		})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusNotFound, resp.StatusCode)
}

func TestWithError_EntityConflictError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return WithError(c, pkg.EntityConflictError{
			Code:    "CONF-001",
			Title:   "Conflict",
			Message: "Entity already exists",
		})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusConflict, resp.StatusCode)
}

func TestWithError_ValidationError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return WithError(c, pkg.ValidationError{
			Code:    "VAL-001",
			Title:   "Validation Error",
			Message: "Invalid input",
		})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusBadRequest, resp.StatusCode)
}

func TestWithError_UnprocessableOperationError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return WithError(c, pkg.UnprocessableOperationError{
			Code:    "UNP-001",
			Title:   "Unprocessable",
			Message: "Cannot process",
		})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusUnprocessableEntity, resp.StatusCode)
}

func TestWithError_UnauthorizedError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return WithError(c, pkg.UnauthorizedError{
			Code:    "AUTH-001",
			Title:   "Unauthorized",
			Message: "Not authenticated",
		})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusUnauthorized, resp.StatusCode)
}

func TestWithError_ForbiddenError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return WithError(c, pkg.ForbiddenError{
			Code:    "FORB-001",
			Title:   "Forbidden",
			Message: "Access denied",
		})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusForbidden, resp.StatusCode)
}

func TestWithError_ValidationKnownFieldsError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return WithError(c, pkg.ValidationKnownFieldsError{
			Code:    "VAL-002",
			Title:   "Validation Error",
			Message: "Field validation failed",
		})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusBadRequest, resp.StatusCode)
}

func TestWithError_ValidationUnknownFieldsError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return WithError(c, pkg.ValidationUnknownFieldsError{
			Code:    "VAL-003",
			Title:   "Unknown Fields",
			Message: "Unknown fields provided",
		})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusBadRequest, resp.StatusCode)
}

func TestWithError_ResponseError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return WithError(c, pkg.ResponseError{
			Code:    stdhttp.StatusBadGateway,
			Title:   "Bad Gateway",
			Message: "Upstream error",
		})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusBadGateway, resp.StatusCode)
}

func TestWithError_UnknownError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return WithError(c, errors.New("unknown error"))
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusInternalServerError, resp.StatusCode)
}
