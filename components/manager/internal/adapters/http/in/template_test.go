// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/components/manager/internal/services"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	templateSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/template"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func setupTemplateTestApp(handler *TemplateHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	return app
}

func setupTemplateContextMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := zap.NewNop().Sugar()
		tracer := noop.NewTracerProvider().Tracer("test")

		ctx := context.WithValue(c.UserContext(), "logger", logger)
		ctx = context.WithValue(ctx, "tracer", tracer)
		ctx = context.WithValue(ctx, "requestId", "test-request-id")

		c.SetUserContext(ctx)

		return c.Next()
	}
}

func createMultipartForm(t *testing.T, filename, content, outputFormat, description string) (*bytes.Buffer, string) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add template file
	part, err := writer.CreateFormFile("template", filename)
	assert.NoError(t, err)
	_, err = part.Write([]byte(content))
	assert.NoError(t, err)

	// Add outputFormat field
	err = writer.WriteField("outputFormat", outputFormat)
	assert.NoError(t, err)

	// Add description field
	err = writer.WriteField("description", description)
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	return body, writer.FormDataContentType()
}

func TestTemplateHandler_GetTemplateByID(t *testing.T) {
	t.Parallel()

	templateID := uuid.New()
	templateEntity := &template.Template{
		ID:           templateID,
		OutputFormat: "HTML",
		Description:  "Test Template",
		FileName:     templateID.String() + ".tpl",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tests := []struct {
		name           string
		templateID     string
		mockSetup      func(mockTemplateRepo *template.MockRepository)
		expectedStatus int
		expectError    bool
	}{
		{
			name:       "Success - Get template by ID",
			templateID: templateID.String(),
			mockSetup: func(mockTemplateRepo *template.MockRepository) {
				mockTemplateRepo.EXPECT().
					FindByID(gomock.Any(), templateID).
					Return(templateEntity, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:       "Error - Template not found",
			templateID: templateID.String(),
			mockSetup: func(mockTemplateRepo *template.MockRepository) {
				mockTemplateRepo.EXPECT().
					FindByID(gomock.Any(), templateID).
					Return(nil, errors.New("template not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "Error - Invalid UUID",
			templateID:     "invalid-uuid",
			mockSetup:      func(mockTemplateRepo *template.MockRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTemplateRepo := template.NewMockRepository(ctrl)
			mockSeaweedFS := templateSeaweedFS.NewMockRepository(ctrl)

			tt.mockSetup(mockTemplateRepo)

			useCase := &services.UseCase{
				TemplateRepo:      mockTemplateRepo,
				TemplateSeaweedFS: mockSeaweedFS,
			}
			handler := &TemplateHandler{service: useCase}

			app := setupTemplateTestApp(handler)
			app.Get("/templates/:id", setupTemplateContextMiddleware(), ParsePathParametersUUID, handler.GetTemplateByID)

			req := httptest.NewRequest(http.MethodGet, "/templates/"+tt.templateID, nil)
			resp, err := app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestTemplateHandler_GetAllTemplates(t *testing.T) {
	t.Parallel()

	templates := []*template.Template{
		{
			ID:           uuid.New(),
			OutputFormat: "HTML",
			Description:  "Test Template 1",
			FileName:     "template1.tpl",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			OutputFormat: "XML",
			Description:  "Test Template 2",
			FileName:     "template2.tpl",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(mockTemplateRepo *template.MockRepository)
		expectedStatus int
		expectError    bool
	}{
		{
			name:        "Success - Get all templates with default pagination",
			queryParams: "",
			mockSetup: func(mockTemplateRepo *template.MockRepository) {
				mockTemplateRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return(templates, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "Success - Get all templates with custom pagination",
			queryParams: "?limit=5&page=2",
			mockSetup: func(mockTemplateRepo *template.MockRepository) {
				mockTemplateRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return(templates, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "Success - Get all templates with filter",
			queryParams: "?outputFormat=HTML",
			mockSetup: func(mockTemplateRepo *template.MockRepository) {
				mockTemplateRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return([]*template.Template{templates[0]}, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "Error - Repository error",
			queryParams: "",
			mockSetup: func(mockTemplateRepo *template.MockRepository) {
				mockTemplateRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "Error - Invalid output format",
			queryParams:    "?outputFormat=INVALID",
			mockSetup:      func(mockTemplateRepo *template.MockRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTemplateRepo := template.NewMockRepository(ctrl)
			mockSeaweedFS := templateSeaweedFS.NewMockRepository(ctrl)

			tt.mockSetup(mockTemplateRepo)

			useCase := &services.UseCase{
				TemplateRepo:      mockTemplateRepo,
				TemplateSeaweedFS: mockSeaweedFS,
			}
			handler := &TemplateHandler{service: useCase}

			app := setupTemplateTestApp(handler)
			app.Get("/templates", setupTemplateContextMiddleware(), handler.GetAllTemplates)

			req := httptest.NewRequest(http.MethodGet, "/templates"+tt.queryParams, nil)
			resp, err := app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestTemplateHandler_DeleteTemplateByID(t *testing.T) {
	t.Parallel()

	templateID := uuid.New()

	tests := []struct {
		name           string
		templateID     string
		mockSetup      func(mockTemplateRepo *template.MockRepository)
		expectedStatus int
		expectError    bool
	}{
		{
			name:       "Success - Delete template",
			templateID: templateID.String(),
			mockSetup: func(mockTemplateRepo *template.MockRepository) {
				mockTemplateRepo.EXPECT().
					Delete(gomock.Any(), templateID, false).
					Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		},
		{
			name:       "Error - Template not found",
			templateID: templateID.String(),
			mockSetup: func(mockTemplateRepo *template.MockRepository) {
				mockTemplateRepo.EXPECT().
					Delete(gomock.Any(), templateID, false).
					Return(errors.New("template not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "Error - Invalid UUID",
			templateID:     "invalid-uuid",
			mockSetup:      func(mockTemplateRepo *template.MockRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTemplateRepo := template.NewMockRepository(ctrl)
			mockSeaweedFS := templateSeaweedFS.NewMockRepository(ctrl)

			tt.mockSetup(mockTemplateRepo)

			useCase := &services.UseCase{
				TemplateRepo:      mockTemplateRepo,
				TemplateSeaweedFS: mockSeaweedFS,
			}
			handler := &TemplateHandler{service: useCase}

			app := setupTemplateTestApp(handler)
			app.Delete("/templates/:id", setupTemplateContextMiddleware(), ParsePathParametersUUID, handler.DeleteTemplateByID)

			req := httptest.NewRequest(http.MethodDelete, "/templates/"+tt.templateID, nil)
			resp, err := app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestTemplateHandler_GetAllTemplates_EmptyResult(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTemplateRepo := template.NewMockRepository(ctrl)
	mockSeaweedFS := templateSeaweedFS.NewMockRepository(ctrl)

	useCase := &services.UseCase{
		TemplateRepo:      mockTemplateRepo,
		TemplateSeaweedFS: mockSeaweedFS,
	}

	handler := &TemplateHandler{service: useCase}

	mockTemplateRepo.EXPECT().
		FindList(gomock.Any(), gomock.Any()).
		Return([]*template.Template{}, nil)

	app := setupTemplateTestApp(handler)
	app.Get("/templates", setupTemplateContextMiddleware(), handler.GetAllTemplates)

	req := httptest.NewRequest(http.MethodGet, "/templates", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "items")
}

func TestTemplateHandler_CreateTemplate_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		filename       string
		content        string
		outputFormat   string
		description    string
		expectedStatus int
	}{
		{
			name:           "Error - Invalid file format (not .tpl)",
			filename:       "template.txt",
			content:        "some content",
			outputFormat:   "html",
			description:    "Test template",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - Invalid output format",
			filename:       "template.tpl",
			content:        "<html>content</html>",
			outputFormat:   "invalid",
			description:    "Test template",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - Empty description",
			filename:       "template.tpl",
			content:        "<html>content</html>",
			outputFormat:   "html",
			description:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - Empty output format",
			filename:       "template.tpl",
			content:        "<html>content</html>",
			outputFormat:   "",
			description:    "Test template",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTemplateRepo := template.NewMockRepository(ctrl)
			mockSeaweedFS := templateSeaweedFS.NewMockRepository(ctrl)

			useCase := &services.UseCase{
				TemplateRepo:      mockTemplateRepo,
				TemplateSeaweedFS: mockSeaweedFS,
			}
			handler := &TemplateHandler{service: useCase}

			app := setupTemplateTestApp(handler)
			app.Post("/templates", setupTemplateContextMiddleware(), handler.CreateTemplate)

			body, contentType := createMultipartForm(t, tt.filename, tt.content, tt.outputFormat, tt.description)

			req := httptest.NewRequest(http.MethodPost, "/templates", body)
			req.Header.Set("Content-Type", contentType)

			resp, err := app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestTemplateHandler_CreateTemplate_EmptyFile(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTemplateRepo := template.NewMockRepository(ctrl)
	mockSeaweedFS := templateSeaweedFS.NewMockRepository(ctrl)

	useCase := &services.UseCase{
		TemplateRepo:      mockTemplateRepo,
		TemplateSeaweedFS: mockSeaweedFS,
	}

	handler := &TemplateHandler{service: useCase}

	app := setupTemplateTestApp(handler)
	app.Post("/templates", setupTemplateContextMiddleware(), handler.CreateTemplate)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("template", "template.tpl")
	assert.NoError(t, err)
	_, err = part.Write([]byte(""))
	assert.NoError(t, err)

	err = writer.WriteField("outputFormat", "html")
	assert.NoError(t, err)
	err = writer.WriteField("description", "Test description")
	assert.NoError(t, err)
	err = writer.Close()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/templates", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestTemplateHandler_CreateTemplate_NoFile(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTemplateRepo := template.NewMockRepository(ctrl)
	mockSeaweedFS := templateSeaweedFS.NewMockRepository(ctrl)

	useCase := &services.UseCase{
		TemplateRepo:      mockTemplateRepo,
		TemplateSeaweedFS: mockSeaweedFS,
	}

	handler := &TemplateHandler{service: useCase}

	app := setupTemplateTestApp(handler)
	app.Post("/templates", setupTemplateContextMiddleware(), handler.CreateTemplate)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	err := writer.WriteField("outputFormat", "html")
	assert.NoError(t, err)
	err = writer.WriteField("description", "Test description")
	assert.NoError(t, err)
	err = writer.Close()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/templates", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestTemplateHandler_UpdateTemplateByID_ValidationErrors(t *testing.T) {
	t.Parallel()

	templateID := uuid.New()

	tests := []struct {
		name           string
		templateID     string
		filename       string
		content        string
		outputFormat   string
		description    string
		expectedStatus int
	}{
		{
			name:           "Error - Invalid UUID",
			templateID:     "invalid-uuid",
			filename:       "template.tpl",
			content:        "content",
			outputFormat:   "html",
			description:    "Test",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - Invalid file format (not .tpl)",
			templateID:     templateID.String(),
			filename:       "template.txt",
			content:        "content",
			outputFormat:   "html",
			description:    "description",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTemplateRepo := template.NewMockRepository(ctrl)
			mockSeaweedFS := templateSeaweedFS.NewMockRepository(ctrl)

			useCase := &services.UseCase{
				TemplateRepo:      mockTemplateRepo,
				TemplateSeaweedFS: mockSeaweedFS,
			}
			handler := &TemplateHandler{service: useCase}

			app := setupTemplateTestApp(handler)
			app.Patch("/templates/:id", setupTemplateContextMiddleware(), ParsePathParametersUUID, handler.UpdateTemplateByID)

			body, contentType := createMultipartForm(t, tt.filename, tt.content, tt.outputFormat, tt.description)

			req := httptest.NewRequest(http.MethodPatch, "/templates/"+tt.templateID, body)
			req.Header.Set("Content-Type", contentType)

			resp, err := app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestNewTemplateHandler_NilService(t *testing.T) {
	t.Parallel()

	handler, err := NewTemplateHandler(nil)

	assert.Nil(t, handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service must not be nil")
}

func TestNewTemplateHandler_ValidService(t *testing.T) {
	t.Parallel()

	svc := &services.UseCase{}

	handler, err := NewTemplateHandler(svc)

	assert.NotNil(t, handler)
	assert.NoError(t, err)
}
