package in

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/v4/components/manager/internal/services"
	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/model"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb/template"
	templateSeaweed "github.com/LerianStudio/reporter/v4/pkg/seaweedfs/template"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestTemplateHandler_GetTemplateByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)

	orgID := uuid.New()
	tempID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		templateID     uuid.UUID
		mockSetup      func()
		expectedStatus int
		expectError    bool
	}{
		{
			name:       "Success - Get template by ID",
			templateID: tempID,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID, orgID).
					Return(&template.Template{
						ID:           tempID,
						OutputFormat: "pdf",
						Description:  "Test template",
						FileName:     tempID.String() + ".tpl",
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectError:    false,
		},
		{
			name:       "Error - Template not found",
			templateID: tempID,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID, orgID).
					Return(nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "template"))
			},
			expectedStatus: fiber.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &services.UseCase{
				TemplateRepo: mockTempRepo,
			}

			handler := &TemplateHandler{
				Service: svc,
			}

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Get("/v1/templates/:id", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.Locals("id", tt.templateID)
				c.SetUserContext(context.Background())
				return handler.GetTemplateByID(c)
			})

			req := httptest.NewRequest("GET", "/v1/templates/"+tt.templateID.String(), nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if !tt.expectError {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var result template.Template
				err = json.Unmarshal(body, &result)
				require.NoError(t, err)

				assert.Equal(t, tt.templateID, result.ID)
			}
		})
	}
}

func TestTemplateHandler_GetAllTemplates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)

	orgID := uuid.New()
	tempID1 := uuid.New()
	tempID2 := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func()
		expectedStatus int
		expectedLen    int
	}{
		{
			name:        "Success - Get all templates",
			queryParams: "?limit=10&page=1",
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return([]*template.Template{
						{
							ID:           tempID1,
							OutputFormat: "pdf",
							Description:  "Test template 1",
							FileName:     tempID1.String() + ".tpl",
							CreatedAt:    now,
							UpdatedAt:    now,
						},
						{
							ID:           tempID2,
							OutputFormat: "html",
							Description:  "Test template 2",
							FileName:     tempID2.String() + ".tpl",
							CreatedAt:    now,
							UpdatedAt:    now,
						},
					}, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedLen:    2,
		},
		{
			name:        "Success - Get all templates with empty result",
			queryParams: "?limit=10&page=1",
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return([]*template.Template{}, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedLen:    0,
		},
		{
			name:        "Error - Database error",
			queryParams: "?limit=10&page=1",
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &services.UseCase{
				TemplateRepo: mockTempRepo,
			}

			handler := &TemplateHandler{
				Service: svc,
			}

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Get("/v1/templates", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.SetUserContext(context.Background())
				return handler.GetAllTemplates(c)
			})

			req := httptest.NewRequest("GET", "/v1/templates"+tt.queryParams, nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == fiber.StatusOK {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var result model.Pagination
				err = json.Unmarshal(body, &result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedLen, result.Total)
			}
		})
	}
}

func TestTemplateHandler_GetAllTemplates_ValidationErrors(t *testing.T) {
	orgID := uuid.New()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Error - Limit exceeds max pagination limit",
			queryParams:    "?limit=1000&page=1",
			expectedStatus: fiber.StatusNotFound, // ErrPaginationLimitExceeded maps to EntityNotFoundError
		},
		{
			name:           "Error - Invalid sort order",
			queryParams:    "?limit=10&page=1&sortOrder=invalid",
			expectedStatus: fiber.StatusNotFound, // ErrInvalidSortOrder maps to EntityNotFoundError
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTempRepo := template.NewMockRepository(ctrl)

			svc := &services.UseCase{
				TemplateRepo: mockTempRepo,
			}

			handler := &TemplateHandler{
				Service: svc,
			}

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Get("/v1/templates", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.SetUserContext(context.Background())
				return handler.GetAllTemplates(c)
			})

			req := httptest.NewRequest("GET", "/v1/templates"+tt.queryParams, nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestTemplateHandler_DeleteTemplateByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)

	orgID := uuid.New()
	tempID := uuid.New()

	tests := []struct {
		name           string
		templateID     uuid.UUID
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:       "Success - Delete template by ID",
			templateID: tempID,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), tempID, orgID, false).
					Return(nil)
			},
			expectedStatus: fiber.StatusNoContent,
		},
		{
			name:       "Error - Template not found",
			templateID: tempID,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), tempID, orgID, false).
					Return(pkg.ValidateBusinessError(constant.ErrEntityNotFound, "template"))
			},
			expectedStatus: fiber.StatusNotFound,
		},
		{
			name:       "Error - Internal server error",
			templateID: tempID,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), tempID, orgID, false).
					Return(constant.ErrInternalServer)
			},
			expectedStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &services.UseCase{
				TemplateRepo: mockTempRepo,
			}

			handler := &TemplateHandler{
				Service: svc,
			}

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Delete("/v1/templates/:id", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.Locals("id", tt.templateID)
				c.SetUserContext(context.Background())
				return handler.DeleteTemplateByID(c)
			})

			req := httptest.NewRequest("DELETE", "/v1/templates/"+tt.templateID.String(), nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestTemplateHandler_CreateTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockSeaweedFS := templateSeaweed.NewMockRepository(ctrl)

	orgID := uuid.New()
	tempID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		outputFormat   string
		description    string
		fileContent    string
		mockSetup      func()
		expectedStatus int
		expectError    bool
	}{
		{
			name:         "Success - Create HTML template",
			outputFormat: "html",
			description:  "Test HTML template",
			fileContent:  "<!DOCTYPE html><html><body>{{ data }}</body></html>",
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&template.Template{
						ID:           tempID,
						OutputFormat: "html",
						Description:  "Test HTML template",
						FileName:     tempID.String() + ".tpl",
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)

				mockSeaweedFS.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: fiber.StatusCreated,
			expectError:    false,
		},
		{
			name:         "Success - Create PDF template",
			outputFormat: "pdf",
			description:  "Test PDF template",
			fileContent:  "<!DOCTYPE html><html><body>PDF Content</body></html>",
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&template.Template{
						ID:           tempID,
						OutputFormat: "pdf",
						Description:  "Test PDF template",
						FileName:     tempID.String() + ".tpl",
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)

				mockSeaweedFS.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: fiber.StatusCreated,
			expectError:    false,
		},
		{
			name:         "Error - Invalid output format",
			outputFormat: "invalid",
			description:  "Test invalid template",
			fileContent:  "some content",
			mockSetup:    func() {},
			expectedStatus: fiber.StatusBadRequest,
			expectError:    true,
		},
		{
			name:         "Error - Empty description",
			outputFormat: "html",
			description:  "",
			fileContent:  "<!DOCTYPE html><html><body>Content</body></html>",
			mockSetup:    func() {},
			expectedStatus: fiber.StatusBadRequest,
			expectError:    true,
		},
		{
			name:         "Error - Database error on create",
			outputFormat: "html",
			description:  "Test template",
			fileContent:  "<!DOCTYPE html><html><body>Content</body></html>",
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:         "Error - SeaweedFS error on put",
			outputFormat: "html",
			description:  "Test template",
			fileContent:  "<!DOCTYPE html><html><body>Content</body></html>",
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&template.Template{
						ID:           tempID,
						OutputFormat: "html",
						Description:  "Test template",
						FileName:     tempID.String() + ".tpl",
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)

				mockSeaweedFS.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(constant.ErrInternalServer)

				// Compensating transaction: rollback
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), tempID, orgID, true).
					Return(nil)
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:         "Error - SeaweedFS error on put with rollback failure",
			outputFormat: "html",
			description:  "Test template rollback",
			fileContent:  "<!DOCTYPE html><html><body>Content</body></html>",
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&template.Template{
						ID:           tempID,
						OutputFormat: "html",
						Description:  "Test template rollback",
						FileName:     tempID.String() + ".tpl",
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)

				mockSeaweedFS.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(constant.ErrInternalServer)

				// Compensating transaction: rollback fails
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), tempID, orgID, true).
					Return(constant.ErrInternalServer)
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:         "Error - File format validation error (PDF with invalid content)",
			outputFormat: "pdf",
			description:  "Test PDF template with invalid content",
			fileContent:  "This is not valid HTML content for PDF",
			mockSetup:    func() {},
			expectedStatus: fiber.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &services.UseCase{
				TemplateRepo:      mockTempRepo,
				TemplateSeaweedFS: mockSeaweedFS,
			}

			handler := &TemplateHandler{
				Service: svc,
			}

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Post("/v1/templates", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.SetUserContext(context.Background())
				return handler.CreateTemplate(c)
			})

			// Create multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("outputFormat", tt.outputFormat)
			_ = writer.WriteField("description", tt.description)

			part, _ := writer.CreateFormFile("template", "test.tpl")
			_, _ = part.Write([]byte(tt.fileContent))

			writer.Close()

			req := httptest.NewRequest("POST", "/v1/templates", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestTemplateHandler_CreateTemplate_InvalidFileExtension(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockSeaweedFS := templateSeaweed.NewMockRepository(ctrl)

	orgID := uuid.New()

	svc := &services.UseCase{
		TemplateRepo:      mockTempRepo,
		TemplateSeaweedFS: mockSeaweedFS,
	}

	handler := &TemplateHandler{
		Service: svc,
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Post("/v1/templates", func(c *fiber.Ctx) error {
		c.Locals("X-Organization-Id", orgID)
		c.SetUserContext(context.Background())
		return handler.CreateTemplate(c)
	})

	// Create multipart form with invalid file extension
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("outputFormat", "html")
	_ = writer.WriteField("description", "Test template")

	// Create file with invalid extension (.txt instead of .tpl)
	part, _ := writer.CreateFormFile("template", "test.txt")
	_, _ = part.Write([]byte("<!DOCTYPE html><html><body>Content</body></html>"))

	writer.Close()

	req := httptest.NewRequest("POST", "/v1/templates", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestTemplateHandler_CreateTemplate_EmptyFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockSeaweedFS := templateSeaweed.NewMockRepository(ctrl)

	orgID := uuid.New()

	svc := &services.UseCase{
		TemplateRepo:      mockTempRepo,
		TemplateSeaweedFS: mockSeaweedFS,
	}

	handler := &TemplateHandler{
		Service: svc,
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Post("/v1/templates", func(c *fiber.Ctx) error {
		c.Locals("X-Organization-Id", orgID)
		c.SetUserContext(context.Background())
		return handler.CreateTemplate(c)
	})

	// Create multipart form with empty file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("outputFormat", "html")
	_ = writer.WriteField("description", "Test template")

	// Create empty file
	part, _ := writer.CreateFormFile("template", "test.tpl")
	_, _ = part.Write([]byte("")) // Empty content

	writer.Close()

	req := httptest.NewRequest("POST", "/v1/templates", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestTemplateHandler_CreateTemplate_NoFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockSeaweedFS := templateSeaweed.NewMockRepository(ctrl)

	orgID := uuid.New()

	svc := &services.UseCase{
		TemplateRepo:      mockTempRepo,
		TemplateSeaweedFS: mockSeaweedFS,
	}

	handler := &TemplateHandler{
		Service: svc,
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Post("/v1/templates", func(c *fiber.Ctx) error {
		c.Locals("X-Organization-Id", orgID)
		c.SetUserContext(context.Background())
		return handler.CreateTemplate(c)
	})

	// Create multipart form without file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("outputFormat", "html")
	_ = writer.WriteField("description", "Test template")

	writer.Close()

	req := httptest.NewRequest("POST", "/v1/templates", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestTemplateHandler_UpdateTemplateByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockSeaweedFS := templateSeaweed.NewMockRepository(ctrl)

	orgID := uuid.New()
	tempID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		templateID     uuid.UUID
		outputFormat   string
		description    string
		fileContent    string
		includeFile    bool
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:         "Success - Update template with new file",
			templateID:   tempID,
			outputFormat: "html",
			description:  "Updated template",
			fileContent:  "<!DOCTYPE html><html><body>Updated</body></html>",
			includeFile:  true,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Update(gomock.Any(), tempID, orgID, gomock.Any()).
					Return(nil)

				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID, orgID).
					Return(&template.Template{
						ID:           tempID,
						OutputFormat: "html",
						Description:  "Updated template",
						FileName:     tempID.String() + ".tpl",
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)

				mockSeaweedFS.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: fiber.StatusOK,
		},
		{
			name:         "Success - Update template without file (description only)",
			templateID:   tempID,
			outputFormat: "", // Empty outputFormat when no file is included
			description:  "Updated description only",
			fileContent:  "",
			includeFile:  false,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Update(gomock.Any(), tempID, orgID, gomock.Any()).
					Return(nil)

				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID, orgID).
					Return(&template.Template{
						ID:           tempID,
						OutputFormat: "html",
						Description:  "Updated description only",
						FileName:     tempID.String() + ".tpl",
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)
			},
			expectedStatus: fiber.StatusOK,
		},
		{
			name:         "Error - Template not found on update",
			templateID:   tempID,
			outputFormat: "", // Empty outputFormat to avoid ErrOutputFormatWithoutTemplateFile
			description:  "Updated template",
			fileContent:  "",
			includeFile:  false,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Update(gomock.Any(), tempID, orgID, gomock.Any()).
					Return(pkg.ValidateBusinessError(constant.ErrEntityNotFound, "template"))
			},
			expectedStatus: fiber.StatusNotFound,
		},
		{
			name:         "Error - SeaweedFS error on update",
			templateID:   tempID,
			outputFormat: "html",
			description:  "Updated template",
			fileContent:  "<!DOCTYPE html><html><body>Updated</body></html>",
			includeFile:  true,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Update(gomock.Any(), tempID, orgID, gomock.Any()).
					Return(nil)

				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID, orgID).
					Return(&template.Template{
						ID:           tempID,
						OutputFormat: "html",
						Description:  "Updated template",
						FileName:     tempID.String() + ".tpl",
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)

				mockSeaweedFS.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(constant.ErrInternalServer)
			},
			expectedStatus: fiber.StatusInternalServerError,
		},
		{
			name:         "Error - GetTemplateByID fails after update",
			templateID:   tempID,
			outputFormat: "",
			description:  "Updated template",
			fileContent:  "",
			includeFile:  false,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Update(gomock.Any(), tempID, orgID, gomock.Any()).
					Return(nil)

				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID, orgID).
					Return(nil, constant.ErrInternalServer)
			},
			expectedStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &services.UseCase{
				TemplateRepo:      mockTempRepo,
				TemplateSeaweedFS: mockSeaweedFS,
			}

			handler := &TemplateHandler{
				Service: svc,
			}

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Patch("/v1/templates/:id", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.Locals("id", tt.templateID)
				c.SetUserContext(context.Background())
				return handler.UpdateTemplateByID(c)
			})

			// Create multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("outputFormat", tt.outputFormat)
			_ = writer.WriteField("description", tt.description)

			if tt.includeFile {
				part, _ := writer.CreateFormFile("template", "test.tpl")
				_, _ = part.Write([]byte(tt.fileContent))
			}

			writer.Close()

			req := httptest.NewRequest("PATCH", "/v1/templates/"+tt.templateID.String(), body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
