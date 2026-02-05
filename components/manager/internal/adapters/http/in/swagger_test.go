// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/LerianStudio/reporter/components/manager/api"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithSwaggerEnvConfig(t *testing.T) {
	originalTitle := api.SwaggerInfo.Title
	originalDescription := api.SwaggerInfo.Description
	originalVersion := api.SwaggerInfo.Version
	originalHost := api.SwaggerInfo.Host
	originalBasePath := api.SwaggerInfo.BasePath
	originalSchemes := api.SwaggerInfo.Schemes

	defer func() {
		api.SwaggerInfo.Title = originalTitle
		api.SwaggerInfo.Description = originalDescription
		api.SwaggerInfo.Version = originalVersion
		api.SwaggerInfo.Host = originalHost
		api.SwaggerInfo.BasePath = originalBasePath
		api.SwaggerInfo.Schemes = originalSchemes
	}()

	tests := []struct {
		name           string
		envVars        map[string]string
		expectedTitle  string
		expectedDesc   string
		expectedVer    string
		expectedHost   string
		expectedBase   string
		expectedScheme []string
	}{
		{
			name:           "No environment variables set",
			envVars:        map[string]string{},
			expectedTitle:  originalTitle,
			expectedDesc:   originalDescription,
			expectedVer:    originalVersion,
			expectedHost:   originalHost,
			expectedBase:   originalBasePath,
			expectedScheme: originalSchemes,
		},
		{
			name: "All environment variables set",
			envVars: map[string]string{
				"SWAGGER_TITLE":       "Test API",
				"SWAGGER_DESCRIPTION": "Test Description",
				"SWAGGER_VERSION":     "2.0.0",
				"SWAGGER_HOST":        "localhost:8080",
				"SWAGGER_BASE_PATH":   "/api/v2",
				"SWAGGER_SCHEMES":     "https",
			},
			expectedTitle:  "Test API",
			expectedDesc:   "Test Description",
			expectedVer:    "2.0.0",
			expectedHost:   "localhost:8080",
			expectedBase:   "/api/v2",
			expectedScheme: []string{"https"},
		},
		{
			name: "Only title set",
			envVars: map[string]string{
				"SWAGGER_TITLE": "Custom Title",
			},
			expectedTitle:  "Custom Title",
			expectedDesc:   originalDescription,
			expectedVer:    originalVersion,
			expectedHost:   originalHost,
			expectedBase:   originalBasePath,
			expectedScheme: originalSchemes,
		},
		{
			name: "Invalid host format is ignored",
			envVars: map[string]string{
				"SWAGGER_HOST": "not a valid host format!!!",
			},
			expectedTitle:  originalTitle,
			expectedDesc:   originalDescription,
			expectedVer:    originalVersion,
			expectedHost:   originalHost,
			expectedBase:   originalBasePath,
			expectedScheme: originalSchemes,
		},
		{
			name: "Valid host with port",
			envVars: map[string]string{
				"SWAGGER_HOST": "api.example.com:443",
			},
			expectedTitle:  originalTitle,
			expectedDesc:   originalDescription,
			expectedVer:    originalVersion,
			expectedHost:   "api.example.com:443",
			expectedBase:   originalBasePath,
			expectedScheme: originalSchemes,
		},
		{
			name: "Multiple schemes",
			envVars: map[string]string{
				"SWAGGER_SCHEMES": "http",
			},
			expectedTitle:  originalTitle,
			expectedDesc:   originalDescription,
			expectedVer:    originalVersion,
			expectedHost:   originalHost,
			expectedBase:   originalBasePath,
			expectedScheme: []string{"http"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api.SwaggerInfo.Title = originalTitle
			api.SwaggerInfo.Description = originalDescription
			api.SwaggerInfo.Version = originalVersion
			api.SwaggerInfo.Host = originalHost
			api.SwaggerInfo.BasePath = originalBasePath
			api.SwaggerInfo.Schemes = originalSchemes

			for key := range tt.envVars {
				os.Unsetenv(key)
			}

			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Get("/swagger/*", WithSwaggerEnvConfig(), func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
			resp, err := app.Test(req)

			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.expectedTitle, api.SwaggerInfo.Title)
			assert.Equal(t, tt.expectedDesc, api.SwaggerInfo.Description)
			assert.Equal(t, tt.expectedVer, api.SwaggerInfo.Version)
			assert.Equal(t, tt.expectedHost, api.SwaggerInfo.Host)
			assert.Equal(t, tt.expectedBase, api.SwaggerInfo.BasePath)

			if len(tt.envVars["SWAGGER_SCHEMES"]) > 0 {
				assert.Equal(t, tt.expectedScheme, api.SwaggerInfo.Schemes)
			}
		})
	}
}

func TestWithSwaggerEnvConfig_EmptyValues(t *testing.T) {
	originalTitle := api.SwaggerInfo.Title

	defer func() {
		api.SwaggerInfo.Title = originalTitle
		os.Unsetenv("SWAGGER_TITLE")
	}()

	os.Setenv("SWAGGER_TITLE", "")

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Get("/swagger/*", WithSwaggerEnvConfig(), func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, originalTitle, api.SwaggerInfo.Title)
}

func TestWithSwaggerEnvConfig_DelimiterSettings(t *testing.T) {
	originalLeftDelim := api.SwaggerInfo.LeftDelim
	originalRightDelim := api.SwaggerInfo.RightDelim

	defer func() {
		api.SwaggerInfo.LeftDelim = originalLeftDelim
		api.SwaggerInfo.RightDelim = originalRightDelim
		os.Unsetenv("SWAGGER_LEFT_DELIM")
		os.Unsetenv("SWAGGER_RIGHT_DELIM")
	}()

	os.Setenv("SWAGGER_LEFT_DELIM", "[[")
	os.Setenv("SWAGGER_RIGHT_DELIM", "]]")

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Get("/swagger/*", WithSwaggerEnvConfig(), func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "[[", api.SwaggerInfo.LeftDelim)
	assert.Equal(t, "]]", api.SwaggerInfo.RightDelim)
}
