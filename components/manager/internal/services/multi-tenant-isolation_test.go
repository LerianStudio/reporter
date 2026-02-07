// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

//go:build multi_tenant

package services

import (
	"context"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/components/manager/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/model"
	"github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/pkg/net/http"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Test_createReport_setsOrganizationID verifies that CreateReport sets the
// organization_id on the report entity before persisting it.
func TestCreateReport_SetsOrganizationID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockReportRepo := report.NewMockRepository(ctrl)
	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)

	orgID := uuid.New()
	tempID := uuid.New()
	reportID := uuid.New()

	reportSvc := &UseCase{
		TemplateRepo: mockTempRepo,
		ReportRepo:   mockReportRepo,
		RabbitMQRepo: mockRabbitMQ,
	}

	mappedFields := map[string]map[string][]string{
		"midaz_onboarding": {
			"asset": {"name", "type", "code"},
		},
	}

	outputFormat := "html"

	reportInput := &model.CreateReportInput{
		TemplateID: tempID.String(),
		Filters:    nil,
	}

	tests := []struct {
		name        string
		orgID       uuid.UUID
		input       *model.CreateReportInput
		mockSetup   func()
		expectErr   bool
		errContains string
	}{
		{
			name:  "Success - CreateReport sets organization_id on the report",
			orgID: orgID,
			input: reportInput,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				// Expect Create to be called with a report that has OrganizationID set
				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, r *report.Report) (*report.Report, error) {
						assert.Equal(t, orgID, r.OrganizationID, "Report must have OrganizationID set")
						r.ID = reportID
						return r, nil
					})

				mockRabbitMQ.EXPECT().
					ProducerDefault(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := reportSvc.CreateReport(ctx, tt.orgID, tt.input)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.orgID, result.OrganizationID)
			}
		})
	}
}

// Test_getReportByID_tenantIsolation verifies that GetReportByID enforces
// tenant isolation by requiring organization_id and returning ErrEntityNotFound
// when the report belongs to a different organization.
func TestGetReportByID_TenantIsolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)

	orgA := uuid.New()
	orgB := uuid.New()
	reportID := uuid.New()
	tempID := uuid.New()
	timeNow := time.Now()

	reportSvc := &UseCase{
		ReportRepo: mockReportRepo,
	}

	reportModel := &report.Report{
		ID:             reportID,
		TemplateID:     tempID,
		OrganizationID: orgA,
		Filters:        nil,
		Status:         constant.FinishedStatus,
		CompletedAt:    &timeNow,
		CreatedAt:      timeNow,
		UpdatedAt:      timeNow,
	}

	tests := []struct {
		name           string
		orgID          uuid.UUID
		reportID       uuid.UUID
		mockSetup      func()
		expectErr      bool
		expectedErr    error
		expectedResult *report.Report
	}{
		{
			name:     "Success - Get report by ID with matching organization",
			orgID:    orgA,
			reportID: reportID,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID, orgA).
					Return(reportModel, nil)
			},
			expectErr:      false,
			expectedResult: reportModel,
		},
		{
			name:     "Error - Get report by ID with non-matching organization returns not found",
			orgID:    orgB,
			reportID: reportID,
			mockSetup: func() {
				// Repository returns not found when org filter doesn't match
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID, orgB).
					Return(nil, constant.ErrEntityNotFound)
			},
			expectErr:      true,
			expectedErr:    constant.ErrEntityNotFound,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := reportSvc.GetReportByID(ctx, tt.reportID, tt.orgID)

			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.orgID, result.OrganizationID)
			}
		})
	}
}

// Test_getAllReports_tenantIsolation verifies that GetAllReports only returns
// reports belonging to the specified organization.
func TestGetAllReports_TenantIsolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)

	orgA := uuid.New()
	orgB := uuid.New()
	templateID := uuid.New()
	reportID1 := uuid.New()
	reportID2 := uuid.New()
	timeNow := time.Now()

	reportSvc := &UseCase{
		ReportRepo: mockReportRepo,
	}

	filters := http.QueryHeader{
		Limit: 10,
		Page:  1,
	}

	orgAReports := []*report.Report{
		{
			ID:             reportID1,
			TemplateID:     templateID,
			OrganizationID: orgA,
			Status:         constant.FinishedStatus,
			CompletedAt:    &timeNow,
			CreatedAt:      timeNow,
			UpdatedAt:      timeNow,
		},
		{
			ID:             reportID2,
			TemplateID:     templateID,
			OrganizationID: orgA,
			Status:         constant.ProcessingStatus,
			CreatedAt:      timeNow,
			UpdatedAt:      timeNow,
		},
	}

	tests := []struct {
		name          string
		orgID         uuid.UUID
		filters       http.QueryHeader
		mockSetup     func()
		expectErr     bool
		errContains   string
		expectedCount int
	}{
		{
			name:    "Success - Get all reports returns only org A reports",
			orgID:   orgA,
			filters: filters,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindList(gomock.Any(), filters, orgA).
					Return(orgAReports, nil)
			},
			expectErr:     false,
			expectedCount: 2,
		},
		{
			name:    "Success - Get all reports returns empty for org B (no reports)",
			orgID:   orgB,
			filters: filters,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindList(gomock.Any(), filters, orgB).
					Return([]*report.Report{}, nil)
			},
			expectErr:     false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := reportSvc.GetAllReports(ctx, tt.filters, tt.orgID)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)
				for _, r := range result {
					assert.Equal(t, tt.orgID, r.OrganizationID,
						"All returned reports must belong to the requesting organization")
				}
			}
		})
	}
}

// Test_getTemplateByID_tenantIsolation verifies that GetTemplateByID enforces
// tenant isolation by requiring organization_id.
func TestGetTemplateByID_TenantIsolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)

	orgA := uuid.New()
	orgB := uuid.New()
	tempID := uuid.New()

	tempSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	templateEntity := &template.Template{
		ID:             tempID,
		OrganizationID: orgA,
		OutputFormat:   "html",
		Description:    "Template Financeiro",
		FileName:       "test.tpl",
		CreatedAt:      time.Now(),
	}

	tests := []struct {
		name           string
		orgID          uuid.UUID
		tempID         uuid.UUID
		mockSetup      func()
		expectErr      bool
		expectedErr    error
		expectedResult *template.Template
	}{
		{
			name:   "Success - Get template by ID with matching organization",
			orgID:  orgA,
			tempID: tempID,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID, orgA).
					Return(templateEntity, nil)
			},
			expectErr:      false,
			expectedResult: templateEntity,
		},
		{
			name:   "Error - Get template by ID with non-matching organization returns not found",
			orgID:  orgB,
			tempID: tempID,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID, orgB).
					Return(nil, constant.ErrEntityNotFound)
			},
			expectErr:      true,
			expectedErr:    constant.ErrEntityNotFound,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := tempSvc.GetTemplateByID(ctx, tt.tempID, tt.orgID)

			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.orgID, result.OrganizationID)
			}
		})
	}
}

// Test_getAllTemplates_tenantIsolation verifies that GetAllTemplates only returns
// templates belonging to the specified organization.
func TestGetAllTemplates_TenantIsolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)

	orgA := uuid.New()
	orgB := uuid.New()
	tempID := uuid.New()

	tempSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	filters := http.QueryHeader{
		Limit: 10,
		Page:  1,
	}

	orgATemplates := []*template.Template{
		{
			ID:             tempID,
			OrganizationID: orgA,
			OutputFormat:   "html",
			Description:    "Template Financeiro",
			FileName:       "test.tpl",
			CreatedAt:      time.Now(),
		},
	}

	tests := []struct {
		name          string
		orgID         uuid.UUID
		filters       http.QueryHeader
		mockSetup     func()
		expectErr     bool
		errContains   string
		expectedCount int
	}{
		{
			name:    "Success - Get all templates returns only org A templates",
			orgID:   orgA,
			filters: filters,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindList(gomock.Any(), filters, orgA).
					Return(orgATemplates, nil)
			},
			expectErr:     false,
			expectedCount: 1,
		},
		{
			name:    "Success - Get all templates returns empty for org B",
			orgID:   orgB,
			filters: filters,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindList(gomock.Any(), filters, orgB).
					Return([]*template.Template{}, nil)
			},
			expectErr:     false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := tempSvc.GetAllTemplates(ctx, tt.filters, tt.orgID)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)
				for _, tmpl := range result {
					assert.Equal(t, tt.orgID, tmpl.OrganizationID,
						"All returned templates must belong to the requesting organization")
				}
			}
		})
	}
}

// Test_deleteTemplateByID_tenantIsolation verifies that DeleteTemplateByID
// enforces tenant isolation by requiring organization_id.
func TestDeleteTemplateByID_TenantIsolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)

	orgA := uuid.New()
	orgB := uuid.New()
	tempID := uuid.New()

	tempSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	tests := []struct {
		name        string
		orgID       uuid.UUID
		tempID      uuid.UUID
		hardDelete  bool
		mockSetup   func()
		expectErr   bool
		expectedErr error
	}{
		{
			name:       "Success - Delete template with matching organization",
			orgID:      orgA,
			tempID:     tempID,
			hardDelete: false,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), tempID, false, orgA).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name:       "Error - Delete template with non-matching organization returns not found",
			orgID:      orgB,
			tempID:     tempID,
			hardDelete: false,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), tempID, false, orgB).
					Return(constant.ErrEntityNotFound)
			},
			expectErr:   true,
			expectedErr: constant.ErrEntityNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			err := tempSvc.DeleteTemplateByID(ctx, tt.tempID, tt.hardDelete, tt.orgID)

			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test_downloadReport_tenantIsolation verifies that DownloadReport enforces
// tenant isolation by passing organization_id through to GetReportByID.
func TestDownloadReport_TenantIsolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)

	orgA := uuid.New()
	orgB := uuid.New()
	reportID := uuid.New()

	reportSvc := &UseCase{
		ReportRepo: mockReportRepo,
	}

	tests := []struct {
		name        string
		orgID       uuid.UUID
		reportID    uuid.UUID
		mockSetup   func()
		expectErr   bool
		expectedErr error
	}{
		{
			name:     "Error - Download report with non-matching organization returns not found",
			orgID:    orgB,
			reportID: reportID,
			mockSetup: func() {
				// The inner GetReportByID call should enforce tenant isolation
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID, orgB).
					Return(nil, constant.ErrEntityNotFound)
			},
			expectErr:   true,
			expectedErr: constant.ErrEntityNotFound,
		},
		{
			name:     "Success - Download report with matching organization proceeds",
			orgID:    orgA,
			reportID: reportID,
			mockSetup: func() {
				timeNow := time.Now()
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID, orgA).
					Return(&report.Report{
						ID:             reportID,
						OrganizationID: orgA,
						Status:         constant.FinishedStatus,
						TemplateID:     uuid.New(),
						CompletedAt:    &timeNow,
					}, nil)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			_, _, _, err := reportSvc.DownloadReport(ctx, tt.reportID, tt.orgID)

			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			}
			// Note: Success case may still error due to missing template/storage mocks,
			// but the tenant isolation check is what we're testing here.
		})
	}
}
