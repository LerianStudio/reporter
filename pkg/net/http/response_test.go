package http

import (
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestUnauthorized(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return Unauthorized(c, "AUTH-001", "Unauthorized", "Invalid credentials")
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusUnauthorized, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "AUTH-001", body["code"])
	assert.Equal(t, "Unauthorized", body["title"])
	assert.Equal(t, "Invalid credentials", body["message"])
}

func TestForbidden(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return Forbidden(c, "AUTH-002", "Forbidden", "Access denied")
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusForbidden, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "AUTH-002", body["code"])
	assert.Equal(t, "Forbidden", body["title"])
	assert.Equal(t, "Access denied", body["message"])
}

func TestBadRequest(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return BadRequest(c, map[string]string{"error": "invalid input"})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusBadRequest, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "invalid input", body["error"])
}

func TestCreated(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return Created(c, map[string]string{"id": "123"})
	})

	req := httptest.NewRequest(stdhttp.MethodPost, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusCreated, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "123", body["id"])
}

func TestOK(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return OK(c, map[string]string{"status": "success"})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusOK, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "success", body["status"])
}

func TestNoContent(t *testing.T) {
	app := fiber.New()
	app.Delete("/test", func(c *fiber.Ctx) error {
		return NoContent(c)
	})

	req := httptest.NewRequest(stdhttp.MethodDelete, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusNoContent, resp.StatusCode)
}

func TestAccepted(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return Accepted(c, map[string]string{"status": "processing"})
	})

	req := httptest.NewRequest(stdhttp.MethodPost, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusAccepted, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "processing", body["status"])
}

func TestPartialContent(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return PartialContent(c, map[string]string{"chunk": "1"})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusPartialContent, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "1", body["chunk"])
}

func TestRangeNotSatisfiable(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return RangeNotSatisfiable(c)
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusRequestedRangeNotSatisfiable, resp.StatusCode)
}

func TestNotFound(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return NotFound(c, "NOT-001", "Not Found", "Resource not found")
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusNotFound, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "NOT-001", body["code"])
	assert.Equal(t, "Not Found", body["title"])
	assert.Equal(t, "Resource not found", body["message"])
}

func TestConflict(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return Conflict(c, "CONF-001", "Conflict", "Resource already exists")
	})

	req := httptest.NewRequest(stdhttp.MethodPost, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusConflict, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "CONF-001", body["code"])
	assert.Equal(t, "Conflict", body["title"])
	assert.Equal(t, "Resource already exists", body["message"])
}

func TestNotImplemented(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return NotImplemented(c, "Feature not available")
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusNotImplemented, resp.StatusCode)

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, float64(stdhttp.StatusNotImplemented), body["code"])
	assert.Equal(t, "Not Implemented", body["title"])
	assert.Equal(t, "Feature not available", body["message"])
}

func TestUnprocessableEntity(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return UnprocessableEntity(c, "UNP-001", "Unprocessable", "Invalid entity")
	})

	req := httptest.NewRequest(stdhttp.MethodPost, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusUnprocessableEntity, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "UNP-001", body["code"])
	assert.Equal(t, "Unprocessable", body["title"])
	assert.Equal(t, "Invalid entity", body["message"])
}

func TestInternalServerError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return InternalServerError(c, "INT-001", "Internal Error", "Something went wrong")
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusInternalServerError, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "INT-001", body["code"])
	assert.Equal(t, "Internal Error", body["title"])
	assert.Equal(t, "Something went wrong", body["message"])
}

func TestJSONResponseError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return JSONResponseError(c, pkg.ResponseError{
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

func TestJSONResponse(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return JSONResponse(c, stdhttp.StatusTeapot, map[string]string{"tea": "earl grey"})
	})

	req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, stdhttp.StatusTeapot, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "earl grey", body["tea"])
}
