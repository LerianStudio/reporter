// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePathParametersUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		pathParam      string
		expectedStatus int
		expectError    bool
		expectLocals   bool
	}{
		{
			name:           "Success - Valid UUID",
			pathParam:      uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
			expectLocals:   true,
		},
		{
			name:           "Error - Invalid UUID format",
			pathParam:      "invalid-uuid-format",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectLocals:   false,
		},
		{
			name:           "Error - Partial UUID",
			pathParam:      "550e8400-e29b-41d4",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectLocals:   false,
		},
		{
			name:           "Error - UUID with invalid characters",
			pathParam:      "550e8400-e29b-41d4-a716-44665544000g",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectLocals:   false,
		},
		{
			name:           "Success - UUID with uppercase letters",
			pathParam:      "550E8400-E29B-41D4-A716-446655440000",
			expectedStatus: http.StatusOK,
			expectError:    false,
			expectLocals:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			var capturedID uuid.UUID
			var localsSet bool

			app.Get("/test/:id", ParsePathParametersUUID, func(c *fiber.Ctx) error {
				if id, ok := c.Locals(UUIDPathParameter).(uuid.UUID); ok {
					capturedID = id
					localsSet = true
				}
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test/"+tt.pathParam, nil)
			resp, err := app.Test(req)

			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectLocals {
				assert.True(t, localsSet, "Expected locals to be set")
				assert.NotEqual(t, uuid.Nil, capturedID, "Expected valid UUID in locals")
			}

			if tt.expectError {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var errorResponse map[string]interface{}
				err = json.Unmarshal(body, &errorResponse)
				require.NoError(t, err)

				assert.Contains(t, errorResponse, "code")
			}
		})
	}
}

func TestParsePathParametersUUID_SpecificUUID(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	expectedUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	var capturedID uuid.UUID

	app.Get("/test/:id", ParsePathParametersUUID, func(c *fiber.Ctx) error {
		capturedID = c.Locals(UUIDPathParameter).(uuid.UUID)
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test/"+expectedUUID.String(), nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expectedUUID, capturedID)
}

func TestParsePathParametersUUID_WithDifferentRoutes(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	validUUID := uuid.New()

	app.Get("/templates/:id", ParsePathParametersUUID, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"type": "template", "id": c.Locals(UUIDPathParameter)})
	})

	app.Get("/reports/:id", ParsePathParametersUUID, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"type": "report", "id": c.Locals(UUIDPathParameter)})
	})

	tests := []struct {
		name         string
		route        string
		expectedType string
	}{
		{
			name:         "Template route",
			route:        "/templates/" + validUUID.String(),
			expectedType: "template",
		},
		{
			name:         "Report route",
			route:        "/reports/" + validUUID.String(),
			expectedType: "report",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.route, nil)
			resp, err := app.Test(req)

			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var response map[string]interface{}
			err = json.Unmarshal(body, &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedType, response["type"])
		})
	}
}

func TestUUIDPathParameter_Constant(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "id", UUIDPathParameter)
}

func TestParseUUIDPathParam_CustomParamName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		pathParam      string
		expectedStatus int
		expectError    bool
		expectLocals   bool
	}{
		{
			name:           "Success - Valid UUID for dataSourceId",
			pathParam:      uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
			expectLocals:   true,
		},
		{
			name:           "Error - Non-UUID string for dataSourceId",
			pathParam:      "mongo_ds",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectLocals:   false,
		},
		{
			name:           "Error - Empty-like string for dataSourceId",
			pathParam:      "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectLocals:   false,
		},
		{
			name:           "Success - UUID with uppercase letters for dataSourceId",
			pathParam:      "550E8400-E29B-41D4-A716-446655440000",
			expectedStatus: http.StatusOK,
			expectError:    false,
			expectLocals:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			const paramName = "dataSourceId"

			var capturedID uuid.UUID
			var localsSet bool

			app.Get("/data-sources/:dataSourceId", ParseUUIDPathParam(paramName), func(c *fiber.Ctx) error {
				if id, ok := c.Locals(paramName).(uuid.UUID); ok {
					capturedID = id
					localsSet = true
				}
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/data-sources/"+tt.pathParam, nil)
			resp, err := app.Test(req)

			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectLocals {
				assert.True(t, localsSet, "Expected locals to be set")
				assert.NotEqual(t, uuid.Nil, capturedID, "Expected valid UUID in locals")
			}

			if tt.expectError {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var errorResponse map[string]interface{}
				err = json.Unmarshal(body, &errorResponse)
				require.NoError(t, err)

				assert.Contains(t, errorResponse, "code")
			}
		})
	}
}

func TestParseUUIDPathParam_SpecificUUID(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	const paramName = "dataSourceId"
	expectedUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	var capturedID uuid.UUID

	app.Get("/data-sources/:dataSourceId", ParseUUIDPathParam(paramName), func(c *fiber.Ctx) error {
		capturedID = c.Locals(paramName).(uuid.UUID)
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/data-sources/"+expectedUUID.String(), nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expectedUUID, capturedID)
}
