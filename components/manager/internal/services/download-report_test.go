// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	reportSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/report"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDownloadReport(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)
	mockTempRepo := template.NewMockRepository(ctrl)
	mockReportStorage := reportSeaweedFS.NewMockRepository(ctrl)

	reportId := uuid.New()
	tempId := uuid.New()
	timeNow := time.Now()

	reportSvc := &UseCase{
		ReportRepo:      mockReportRepo,
		TemplateRepo:    mockTempRepo,
		ReportSeaweedFS: mockReportStorage,
	}

	finishedReport := &report.Report{
		ID:          reportId,
		TemplateID:  tempId,
		Filters:     nil,
		Status:      constant.FinishedStatus,
		CompletedAt: &timeNow,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
		DeletedAt:   nil,
	}

	processingReport := &report.Report{
		ID:         reportId,
		TemplateID: tempId,
		Filters:    nil,
		Status:     constant.ProcessingStatus,
		CreatedAt:  timeNow,
		UpdatedAt:  timeNow,
		DeletedAt:  nil,
	}

	templateEntity := &template.Template{
		ID:           tempId,
		OutputFormat: "pdf",
		Description:  "Template Financeiro",
		FileName:     tempId.String() + "_1744119295.tpl",
		CreatedAt:    timeNow,
		UpdatedAt:    timeNow,
	}

	expectedFileBytes := []byte("report-file-content")

	tests := []struct {
		name          string
		reportId      uuid.UUID
		mockSetup     func()
		expectErr     bool
		expectedBytes []byte
	}{
		{
			name:     "Success - Download finished report",
			reportId: reportId,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(finishedReport, nil)

				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(templateEntity, nil)

				mockReportStorage.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(expectedFileBytes, nil)
			},
			expectErr:     false,
			expectedBytes: expectedFileBytes,
		},
		{
			name:     "Error - GetReportByID fails",
			reportId: reportId,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr:     true,
			expectedBytes: nil,
		},
		{
			name:     "Error - Report status not finished",
			reportId: reportId,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(processingReport, nil)
			},
			expectErr:     true,
			expectedBytes: nil,
		},
		{
			name:     "Error - GetTemplateByID fails",
			reportId: reportId,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(finishedReport, nil)

				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("template not found"))
			},
			expectErr:     true,
			expectedBytes: nil,
		},
		{
			name:     "Error - Storage Get fails",
			reportId: reportId,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(finishedReport, nil)

				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(templateEntity, nil)

				mockReportStorage.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("storage unavailable"))
			},
			expectErr:     true,
			expectedBytes: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			fileBytes, objectName, contentType, err := reportSvc.DownloadReport(ctx, tt.reportId)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, fileBytes)
				assert.Empty(t, objectName)
				assert.Empty(t, contentType)
			} else {
				require.NoError(t, err)
				require.NotNil(t, fileBytes)
				assert.Equal(t, tt.expectedBytes, fileBytes)
				assert.NotEmpty(t, objectName)
				assert.NotEmpty(t, contentType)
			}
		})
	}
}
