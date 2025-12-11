package in

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/LerianStudio/reporter/v4/components/manager/api"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithSwaggerEnvConfig(t *testing.T) {
	// Save original values
	originalTitle := api.SwaggerInfo.Title
	originalDescription := api.SwaggerInfo.Description
	originalVersion := api.SwaggerInfo.Version
	originalHost := api.SwaggerInfo.Host
	originalBasePath := api.SwaggerInfo.BasePath
	originalSchemes := api.SwaggerInfo.Schemes

	// Restore original values after test
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
		expectedHost   string
		expectedScheme string
		setupEnv       func()
		cleanupEnv     func()
	}{
		{
			name:    "Success - Sets all environment variables",
			envVars: map[string]string{},
			setupEnv: func() {
				os.Setenv("SWAGGER_TITLE", "Test API")
				os.Setenv("SWAGGER_DESCRIPTION", "Test Description")
				os.Setenv("SWAGGER_VERSION", "1.0.0")
				os.Setenv("SWAGGER_HOST", "localhost:8080")
				os.Setenv("SWAGGER_BASE_PATH", "/api")
				os.Setenv("SWAGGER_SCHEMES", "https")
			},
			cleanupEnv: func() {
				os.Unsetenv("SWAGGER_TITLE")
				os.Unsetenv("SWAGGER_DESCRIPTION")
				os.Unsetenv("SWAGGER_VERSION")
				os.Unsetenv("SWAGGER_HOST")
				os.Unsetenv("SWAGGER_BASE_PATH")
				os.Unsetenv("SWAGGER_SCHEMES")
			},
			expectedTitle:  "Test API",
			expectedHost:   "localhost:8080",
			expectedScheme: "https",
		},
		{
			name:    "Success - No environment variables set",
			envVars: map[string]string{},
			setupEnv: func() {
				// Ensure env vars are not set
				os.Unsetenv("SWAGGER_TITLE")
				os.Unsetenv("SWAGGER_DESCRIPTION")
				os.Unsetenv("SWAGGER_VERSION")
				os.Unsetenv("SWAGGER_HOST")
				os.Unsetenv("SWAGGER_BASE_PATH")
				os.Unsetenv("SWAGGER_SCHEMES")
			},
			cleanupEnv:     func() {},
			expectedTitle:  originalTitle,
			expectedHost:   originalHost,
			expectedScheme: "",
		},
		{
			name:    "Success - Invalid host is skipped",
			envVars: map[string]string{},
			setupEnv: func() {
				os.Setenv("SWAGGER_TITLE", "Valid Title")
				os.Setenv("SWAGGER_HOST", "") // Empty host should be skipped
			},
			cleanupEnv: func() {
				os.Unsetenv("SWAGGER_TITLE")
				os.Unsetenv("SWAGGER_HOST")
			},
			expectedTitle: "Valid Title",
			expectedHost:  originalHost,
		},
		{
			name:    "Success - Only title set",
			envVars: map[string]string{},
			setupEnv: func() {
				os.Setenv("SWAGGER_TITLE", "Only Title")
			},
			cleanupEnv: func() {
				os.Unsetenv("SWAGGER_TITLE")
			},
			expectedTitle: "Only Title",
			expectedHost:  originalHost,
		},
		{
			name:    "Success - Left and right delimiters",
			envVars: map[string]string{},
			setupEnv: func() {
				os.Setenv("SWAGGER_LEFT_DELIM", "<<")
				os.Setenv("SWAGGER_RIGHT_DELIM", ">>")
			},
			cleanupEnv: func() {
				os.Unsetenv("SWAGGER_LEFT_DELIM")
				os.Unsetenv("SWAGGER_RIGHT_DELIM")
			},
			expectedTitle: originalTitle,
			expectedHost:  originalHost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset swagger info before each test
			api.SwaggerInfo.Title = originalTitle
			api.SwaggerInfo.Description = originalDescription
			api.SwaggerInfo.Version = originalVersion
			api.SwaggerInfo.Host = originalHost
			api.SwaggerInfo.BasePath = originalBasePath
			api.SwaggerInfo.Schemes = originalSchemes

			tt.setupEnv()
			defer tt.cleanupEnv()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Get("/swagger/*", WithSwaggerEnvConfig(), func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/swagger/index.html", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			if tt.expectedTitle != "" {
				assert.Equal(t, tt.expectedTitle, api.SwaggerInfo.Title)
			}

			if tt.expectedHost != "" {
				assert.Equal(t, tt.expectedHost, api.SwaggerInfo.Host)
			}

			if tt.expectedScheme != "" {
				assert.Contains(t, api.SwaggerInfo.Schemes, tt.expectedScheme)
			}
		})
	}
}

func TestWithSwaggerEnvConfig_HostValidation(t *testing.T) {
	originalHost := api.SwaggerInfo.Host
	defer func() {
		api.SwaggerInfo.Host = originalHost
	}()

	tests := []struct {
		name         string
		hostValue    string
		shouldUpdate bool
	}{
		{
			name:         "Valid host with port",
			hostValue:    "api.example.com:8080",
			shouldUpdate: true,
		},
		{
			name:         "Valid localhost",
			hostValue:    "localhost:3000",
			shouldUpdate: true,
		},
		{
			name:         "Empty host - skipped",
			hostValue:    "",
			shouldUpdate: false,
		},
		{
			name:         "Invalid host format - skipped",
			hostValue:    "://invalid-host",
			shouldUpdate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api.SwaggerInfo.Host = originalHost

			os.Setenv("SWAGGER_HOST", tt.hostValue)
			defer os.Unsetenv("SWAGGER_HOST")

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Get("/swagger/*", WithSwaggerEnvConfig(), func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/swagger/index.html", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			if tt.shouldUpdate && tt.hostValue != "" {
				assert.Equal(t, tt.hostValue, api.SwaggerInfo.Host)
			}
		})
	}
}
