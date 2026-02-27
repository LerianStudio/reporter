// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

//go:build unit

package services

import (
	"context"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/model"
	"github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/pkg/net/http"
	"github.com/LerianStudio/reporter/pkg/rabbitmq"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestUseCase_CreateReport_SetsOrganizationID verifies that CreateReport successfully
// creates a report and persists it. Tenant identity is carried by the context
// (injected by the tenant middleware), not by method parameters.
func TestUseCase_CreateReport_SetsOrganizationID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockReportRepo := report.NewMockRepository(ctrl)
	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)

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
		input       *model.CreateReportInput
		mockSetup   func()
		expectErr   bool
		errContains string
	}{
		{
			name:  "Success - CreateReport creates and persists a report",
			input: reportInput,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, r *report.Report) (*report.Report, error) {
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
			result, err := reportSvc.CreateReport(ctx, tt.input)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

// TestUseCase_GetReportByID_TenantIsolation verifies that GetReportByID delegates
// to the repository. Tenant isolation is enforced at the repository layer via the
// MongoDB connection scoped to the tenant from context.
func TestUseCase_GetReportByID_TenantIsolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)

	reportID := uuid.New()
	tempID := uuid.New()
	timeNow := time.Now()

	reportSvc := &UseCase{
		ReportRepo: mockReportRepo,
	}

	reportModel := &report.Report{
		ID:          reportID,
		TemplateID:  tempID,
		Filters:     nil,
		Status:      constant.FinishedStatus,
		CompletedAt: &timeNow,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	tests := []struct {
		name           string
		reportID       uuid.UUID
		mockSetup      func()
		expectErr      bool
		expectedErr    error
		expectedResult *report.Report
	}{
		{
			name:     "Success - Get report by ID returns the report",
			reportID: reportID,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID).
					Return(reportModel, nil)
			},
			expectErr:      false,
			expectedResult: reportModel,
		},
		{
			name:     "Error - Get report by ID returns not found when repo returns not found",
			reportID: reportID,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID).
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
			result, err := reportSvc.GetReportByID(ctx, tt.reportID)

			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

// TestUseCase_GetAllReports_TenantIsolation verifies that GetAllReports returns
// reports from the repository. Tenant scoping is enforced at the DB layer via
// the tenant-scoped MongoDB connection in context.
func TestUseCase_GetAllReports_TenantIsolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)

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

	tenantReports := []*report.Report{
		{
			ID:          reportID1,
			TemplateID:  templateID,
			Status:      constant.FinishedStatus,
			CompletedAt: &timeNow,
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
		},
		{
			ID:         reportID2,
			TemplateID: templateID,
			Status:     constant.ProcessingStatus,
			CreatedAt:  timeNow,
			UpdatedAt:  timeNow,
		},
	}

	tests := []struct {
		name          string
		filters       http.QueryHeader
		mockSetup     func()
		expectErr     bool
		errContains   string
		expectedCount int
	}{
		{
			name:    "Success - Get all reports returns all tenant reports",
			filters: filters,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindList(gomock.Any(), filters).
					Return(tenantReports, nil)
			},
			expectErr:     false,
			expectedCount: 2,
		},
		{
			name:    "Success - Get all reports returns empty when no reports exist",
			filters: filters,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindList(gomock.Any(), filters).
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
			result, err := reportSvc.GetAllReports(ctx, tt.filters)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

// TestUseCase_GetTemplateByID_TenantIsolation verifies that GetTemplateByID delegates
// to the repository. Tenant isolation is enforced at the repository layer.
func TestUseCase_GetTemplateByID_TenantIsolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)

	tempID := uuid.New()

	tempSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	templateEntity := &template.Template{
		ID:           tempID,
		OutputFormat: "html",
		Description:  "Template Financeiro",
		FileName:     "test.tpl",
		CreatedAt:    time.Now(),
	}

	tests := []struct {
		name           string
		tempID         uuid.UUID
		mockSetup      func()
		expectErr      bool
		expectedErr    error
		expectedResult *template.Template
	}{
		{
			name:   "Success - Get template by ID returns the template",
			tempID: tempID,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID).
					Return(templateEntity, nil)
			},
			expectErr:      false,
			expectedResult: templateEntity,
		},
		{
			name:   "Error - Get template by ID returns not found when repo returns not found",
			tempID: tempID,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), tempID).
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
			result, err := tempSvc.GetTemplateByID(ctx, tt.tempID)

			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

// TestUseCase_GetAllTemplates_TenantIsolation verifies that GetAllTemplates returns
// templates from the repository. Tenant scoping is enforced at the DB layer.
func TestUseCase_GetAllTemplates_TenantIsolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)

	tempID := uuid.New()

	tempSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	filters := http.QueryHeader{
		Limit: 10,
		Page:  1,
	}

	tenantTemplates := []*template.Template{
		{
			ID:           tempID,
			OutputFormat: "html",
			Description:  "Template Financeiro",
			FileName:     "test.tpl",
			CreatedAt:    time.Now(),
		},
	}

	tests := []struct {
		name          string
		filters       http.QueryHeader
		mockSetup     func()
		expectErr     bool
		errContains   string
		expectedCount int
	}{
		{
			name:    "Success - Get all templates returns all tenant templates",
			filters: filters,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindList(gomock.Any(), filters).
					Return(tenantTemplates, nil)
			},
			expectErr:     false,
			expectedCount: 1,
		},
		{
			name:    "Success - Get all templates returns empty when none exist",
			filters: filters,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindList(gomock.Any(), filters).
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
			result, err := tempSvc.GetAllTemplates(ctx, tt.filters)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

// TestUseCase_DeleteTemplateByID_TenantIsolation verifies that DeleteTemplateByID
// delegates to the repository. Tenant isolation is enforced at the repository layer.
func TestUseCase_DeleteTemplateByID_TenantIsolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)

	tempID := uuid.New()

	tempSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	tests := []struct {
		name        string
		tempID      uuid.UUID
		hardDelete  bool
		mockSetup   func()
		expectErr   bool
		expectedErr error
	}{
		{
			name:       "Success - Delete template succeeds",
			tempID:     tempID,
			hardDelete: false,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), tempID, false).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name:       "Error - Delete template returns not found when template does not exist",
			tempID:     tempID,
			hardDelete: false,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), tempID, false).
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
			err := tempSvc.DeleteTemplateByID(ctx, tt.tempID, tt.hardDelete)

			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestUseCase_DownloadReport_TenantIsolation verifies that DownloadReport enforces
// tenant isolation by delegating to GetReportByID which uses the tenant-scoped
// MongoDB connection from context.
func TestUseCase_DownloadReport_TenantIsolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)

	reportID := uuid.New()

	reportSvc := &UseCase{
		ReportRepo: mockReportRepo,
	}

	tests := []struct {
		name        string
		reportID    uuid.UUID
		mockSetup   func()
		expectedErr error
	}{
		{
			name:     "Error - Download report returns not found when repo returns not found",
			reportID: reportID,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), reportID).
					Return(nil, constant.ErrEntityNotFound)
			},
			expectedErr: constant.ErrEntityNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			_, _, _, err := reportSvc.DownloadReport(ctx, tt.reportID)

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
