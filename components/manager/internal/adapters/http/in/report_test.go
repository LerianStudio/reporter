package in

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/v4/components/manager/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/model"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb/template"
	reportSeaweed "github.com/LerianStudio/reporter/v4/pkg/seaweedfs/report"

	"github.com/LerianStudio/reporter/v4/components/manager/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestReportHandler_CreateReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockReportRepo := report.NewMockRepository(ctrl)
	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)

	orgID := uuid.New()
	tempID := uuid.New()
	reportID := uuid.New()

	tests := []struct {
		name           string
		payload        model.CreateReportInput
		mockSetup      func()
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Success - Create report",
			payload: model.CreateReportInput{
				TemplateID: tempID.String(),
				Filters:    nil,
			},
			mockSetup: func() {
				outputFormat := "pdf"
				mappedFields := map[string]map[string][]string{
					"midaz_onboarding": {
						"account": {"id", "name"},
					},
				}

				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&report.Report{
						ID:         reportID,
						TemplateID: tempID,
						Status:     constant.ProcessingStatus,
					}, nil)

				mockRabbitMQ.EXPECT().
					ProducerDefault(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectError:    false,
		},
		{
			name: "Error - Template not found",
			payload: model.CreateReportInput{
				TemplateID: tempID.String(),
				Filters:    nil,
			},
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "template"))
			},
			expectedStatus: fiber.StatusNotFound,
			expectError:    true,
		},
		{
			name: "Error - Report creation fails",
			payload: model.CreateReportInput{
				TemplateID: tempID.String(),
				Filters:    nil,
			},
			mockSetup: func() {
				outputFormat := "pdf"
				mappedFields := map[string]map[string][]string{
					"midaz_onboarding": {
						"account": {"id", "name"},
					},
				}

				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &services.UseCase{
				TemplateRepo: mockTempRepo,
				ReportRepo:   mockReportRepo,
				RabbitMQRepo: mockRabbitMQ,
			}

			handler := &ReportHandler{
				Service: svc,
			}

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Post("/v1/reports", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.SetUserContext(context.Background())
				return handler.CreateReport(&tt.payload, c)
			})

			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/v1/reports", bytes.NewReader(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestReportHandler_GetReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)

	orgID := uuid.New()
	reportID := uuid.New()
	tempID := uuid.New()

	now := time.Now()

	tests := []struct {
		name           string
		reportID       uuid.UUID
		mockSetup      func()
		expectedStatus int
		expectError    bool
	}{
		{
			name:     "Success - Get report by ID",
			reportID: reportID,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID, orgID).
					Return(&report.Report{
						ID:          reportID,
						TemplateID:  tempID,
						Status:      constant.FinishedStatus,
						CreatedAt:   now,
						CompletedAt: &now,
					}, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectError:    false,
		},
		{
			name:     "Error - Report not found",
			reportID: reportID,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID, orgID).
					Return(nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "report"))
			},
			expectedStatus: fiber.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &services.UseCase{
				ReportRepo: mockReportRepo,
			}

			handler := &ReportHandler{
				Service: svc,
			}

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Get("/v1/reports/:id", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.Locals("id", tt.reportID)
				c.SetUserContext(context.Background())
				return handler.GetReport(c)
			})

			req := httptest.NewRequest("GET", "/v1/reports/"+tt.reportID.String(), nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if !tt.expectError {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var result report.Report
				err = json.Unmarshal(body, &result)
				require.NoError(t, err)

				assert.Equal(t, tt.reportID, result.ID)
			}
		})
	}
}

func TestReportHandler_GetAllReports(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)

	orgID := uuid.New()
	reportID1 := uuid.New()
	reportID2 := uuid.New()
	tempID := uuid.New()

	now := time.Now()

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func()
		expectedStatus int
		expectedLen    int
	}{
		{
			name:        "Success - Get all reports",
			queryParams: "?limit=10&page=1",
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return([]*report.Report{
						{
							ID:          reportID1,
							TemplateID:  tempID,
							Status:      constant.FinishedStatus,
							CreatedAt:   now,
							CompletedAt: &now,
						},
						{
							ID:          reportID2,
							TemplateID:  tempID,
							Status:      constant.ProcessingStatus,
							CreatedAt:   now,
							CompletedAt: nil,
						},
					}, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedLen:    2,
		},
		{
			name:        "Success - Get all reports with empty result",
			queryParams: "?limit=10&page=1",
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return([]*report.Report{}, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedLen:    0,
		},
		{
			name:        "Error - Database error",
			queryParams: "?limit=10&page=1",
			mockSetup: func() {
				mockReportRepo.EXPECT().
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
				ReportRepo: mockReportRepo,
			}

			handler := &ReportHandler{
				Service: svc,
			}

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Get("/v1/reports", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.SetUserContext(context.Background())
				return handler.GetAllReports(c)
			})

			req := httptest.NewRequest("GET", "/v1/reports"+tt.queryParams, nil)
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

func TestReportHandler_GetAllReports_ValidationErrors(t *testing.T) {
	orgID := uuid.New()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Error - Limit exceeds max pagination limit",
			queryParams:    "?limit=1000&page=1",
			expectedStatus: fiber.StatusBadRequest, // ErrPaginationLimitExceeded maps to ValidationError
		},
		{
			name:           "Error - Invalid sort order",
			queryParams:    "?limit=10&page=1&sortOrder=invalid",
			expectedStatus: fiber.StatusBadRequest, // ErrInvalidSortOrder maps to ValidationError
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockReportRepo := report.NewMockRepository(ctrl)

			svc := &services.UseCase{
				ReportRepo: mockReportRepo,
			}

			handler := &ReportHandler{
				Service: svc,
			}

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Get("/v1/reports", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.SetUserContext(context.Background())
				return handler.GetAllReports(c)
			})

			req := httptest.NewRequest("GET", "/v1/reports"+tt.queryParams, nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestReportHandler_GetDownloadReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)
	mockTempRepo := template.NewMockRepository(ctrl)
	mockSeaweedFS := reportSeaweed.NewMockRepository(ctrl)

	orgID := uuid.New()
	reportID := uuid.New()
	tempID := uuid.New()

	now := time.Now()

	tests := []struct {
		name           string
		reportID       uuid.UUID
		mockSetup      func()
		expectedStatus int
		expectError    bool
	}{
		{
			name:     "Success - Download report",
			reportID: reportID,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID, orgID).
					Return(&report.Report{
						ID:          reportID,
						TemplateID:  tempID,
						Status:      constant.FinishedStatus,
						CreatedAt:   now,
						CompletedAt: &now,
					}, nil)

				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID, orgID).
					Return(&template.Template{
						ID:           tempID,
						OutputFormat: "pdf",
						FileName:     tempID.String() + ".tpl",
					}, nil)

				mockSeaweedFS.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return([]byte("PDF content here"), nil)
			},
			expectedStatus: fiber.StatusOK,
			expectError:    false,
		},
		{
			name:     "Error - Report not found",
			reportID: reportID,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID, orgID).
					Return(nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "report"))
			},
			expectedStatus: fiber.StatusNotFound,
			expectError:    true,
		},
		{
			name:     "Error - Report not finished",
			reportID: reportID,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID, orgID).
					Return(&report.Report{
						ID:         reportID,
						TemplateID: tempID,
						Status:     constant.ProcessingStatus,
						CreatedAt:  now,
					}, nil)
			},
			expectedStatus: fiber.StatusBadRequest, // ErrReportStatusNotFinished returns ValidationError (400)
			expectError:    true,
		},
		{
			name:     "Error - Template not found",
			reportID: reportID,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID, orgID).
					Return(&report.Report{
						ID:          reportID,
						TemplateID:  tempID,
						Status:      constant.FinishedStatus,
						CreatedAt:   now,
						CompletedAt: &now,
					}, nil)

				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID, orgID).
					Return(nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "template"))
			},
			expectedStatus: fiber.StatusNotFound,
			expectError:    true,
		},
		{
			name:     "Error - File not found in SeaweedFS",
			reportID: reportID,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID, orgID).
					Return(&report.Report{
						ID:          reportID,
						TemplateID:  tempID,
						Status:      constant.FinishedStatus,
						CreatedAt:   now,
						CompletedAt: &now,
					}, nil)

				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID, orgID).
					Return(&template.Template{
						ID:           tempID,
						OutputFormat: "pdf",
						FileName:     tempID.String() + ".tpl",
					}, nil)

				mockSeaweedFS.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &services.UseCase{
				ReportRepo:      mockReportRepo,
				TemplateRepo:    mockTempRepo,
				ReportSeaweedFS: mockSeaweedFS,
			}

			handler := &ReportHandler{
				Service: svc,
			}

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			app.Get("/v1/reports/:id/download", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.Locals("id", tt.reportID)
				c.SetUserContext(context.Background())
				return handler.GetDownloadReport(c)
			})

			req := httptest.NewRequest("GET", "/v1/reports/"+tt.reportID.String()+"/download", nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if !tt.expectError {
				assert.Contains(t, resp.Header.Get("Content-Type"), "application/pdf")
				assert.Contains(t, resp.Header.Get("Content-Disposition"), "attachment")
			}
		})
	}
}
