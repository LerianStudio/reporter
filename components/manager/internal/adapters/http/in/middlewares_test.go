package in

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePathParametersUUID(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name           string
		pathParam      string
		expectedStatus int
		expectNext     bool
		expectedUUID   uuid.UUID
	}{
		{
			name:           "Success - Valid UUID path parameter",
			pathParam:      validUUID.String(),
			expectedStatus: fiber.StatusOK,
			expectNext:     true,
			expectedUUID:   validUUID,
		},
		{
			name:           "Error - Invalid UUID format",
			pathParam:      "invalid-uuid",
			expectedStatus: fiber.StatusBadRequest,
			expectNext:     false,
		},
		{
			name:           "Error - Partial UUID",
			pathParam:      "123e4567-e89b",
			expectedStatus: fiber.StatusBadRequest,
			expectNext:     false,
		},
		{
			name:           "Error - UUID with extra characters",
			pathParam:      validUUID.String() + "-extra",
			expectedStatus: fiber.StatusBadRequest,
			expectNext:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			var capturedUUID uuid.UUID
			nextCalled := false

			app.Get("/test/:id", ParsePathParametersUUID, func(c *fiber.Ctx) error {
				nextCalled = true
				if val := c.Locals(UUIDPathParameter); val != nil {
					capturedUUID = val.(uuid.UUID)
				}
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test/"+tt.pathParam, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectNext, nextCalled)

			if tt.expectNext {
				assert.Equal(t, tt.expectedUUID, capturedUUID)
			}
		})
	}
}

func TestParsePathParametersUUID_NoPathParam(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Route without :id parameter - simulates missing path param
	app.Get("/test", ParsePathParametersUUID, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestParseHeaderParameters(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name           string
		headerValue    string
		includeHeader  bool
		expectedStatus int
		expectNext     bool
		expectedUUID   uuid.UUID
	}{
		{
			name:           "Success - Valid UUID header",
			headerValue:    validUUID.String(),
			includeHeader:  true,
			expectedStatus: fiber.StatusOK,
			expectNext:     true,
			expectedUUID:   validUUID,
		},
		{
			name:           "Error - Missing header",
			headerValue:    "",
			includeHeader:  false,
			expectedStatus: fiber.StatusBadRequest,
			expectNext:     false,
		},
		{
			name:           "Error - Empty header value",
			headerValue:    "",
			includeHeader:  true,
			expectedStatus: fiber.StatusBadRequest,
			expectNext:     false,
		},
		{
			name:           "Error - Invalid UUID format",
			headerValue:    "not-a-uuid",
			includeHeader:  true,
			expectedStatus: fiber.StatusBadRequest,
			expectNext:     false,
		},
		{
			name:           "Error - Partial UUID",
			headerValue:    "123e4567-e89b-12d3",
			includeHeader:  true,
			expectedStatus: fiber.StatusBadRequest,
			expectNext:     false,
		},
		{
			name:           "Error - UUID with invalid characters",
			headerValue:    validUUID.String() + "!@#",
			includeHeader:  true,
			expectedStatus: fiber.StatusBadRequest,
			expectNext:     false,
		},
		{
			name:           "Error - Numeric string",
			headerValue:    "12345678901234567890",
			includeHeader:  true,
			expectedStatus: fiber.StatusBadRequest,
			expectNext:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			var capturedUUID uuid.UUID
			nextCalled := false

			app.Get("/test", ParseHeaderParameters, func(c *fiber.Ctx) error {
				nextCalled = true
				if val := c.Locals(OrgIDHeaderParameter); val != nil {
					capturedUUID = val.(uuid.UUID)
				}
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.includeHeader {
				req.Header.Set(OrgIDHeaderParameter, tt.headerValue)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectNext, nextCalled)

			if tt.expectNext {
				assert.Equal(t, tt.expectedUUID, capturedUUID)
			}
		})
	}
}

func TestParseHeaderParameters_ErrorResponse(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Get("/test", ParseHeaderParameters, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(OrgIDHeaderParameter, "invalid")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Verify error response contains expected error structure
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "code")
}

func TestMiddlewareChain(t *testing.T) {
	validOrgID := uuid.New()
	validID := uuid.New()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	var capturedOrgID, capturedID uuid.UUID

	app.Get("/test/:id",
		ParseHeaderParameters,
		ParsePathParametersUUID,
		func(c *fiber.Ctx) error {
			if val := c.Locals(OrgIDHeaderParameter); val != nil {
				capturedOrgID = val.(uuid.UUID)
			}
			if val := c.Locals(UUIDPathParameter); val != nil {
				capturedID = val.(uuid.UUID)
			}
			return c.SendStatus(fiber.StatusOK)
		},
	)

	req := httptest.NewRequest("GET", "/test/"+validID.String(), nil)
	req.Header.Set(OrgIDHeaderParameter, validOrgID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, validOrgID, capturedOrgID)
	assert.Equal(t, validID, capturedID)
}

func TestMiddlewareChain_FailsOnHeader(t *testing.T) {
	validID := uuid.New()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	pathMiddlewareCalled := false

	app.Get("/test/:id",
		ParseHeaderParameters,
		func(c *fiber.Ctx) error {
			pathMiddlewareCalled = true
			return ParsePathParametersUUID(c)
		},
		func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		},
	)

	req := httptest.NewRequest("GET", "/test/"+validID.String(), nil)
	// Missing X-Organization-Id header

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.False(t, pathMiddlewareCalled, "Path middleware should not be called when header validation fails")
}

func TestMiddlewareChain_FailsOnPath(t *testing.T) {
	validOrgID := uuid.New()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	handlerCalled := false

	app.Get("/test/:id",
		ParseHeaderParameters,
		ParsePathParametersUUID,
		func(c *fiber.Ctx) error {
			handlerCalled = true
			return c.SendStatus(fiber.StatusOK)
		},
	)

	req := httptest.NewRequest("GET", "/test/invalid-uuid", nil)
	req.Header.Set(OrgIDHeaderParameter, validOrgID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.False(t, handlerCalled, "Handler should not be called when path validation fails")
}

func TestUUIDPathParameter_Constant(t *testing.T) {
	assert.Equal(t, "id", UUIDPathParameter)
}

func TestOrgIDHeaderParameter_Constant(t *testing.T) {
	assert.Equal(t, "X-Organization-Id", OrgIDHeaderParameter)
}
