// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/net/http"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_getAllReports(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)
	templateId := uuid.New()
	reportId1 := uuid.New()
	reportId2 := uuid.New()
	timeNow := time.Now()

	reportSvc := &UseCase{
		ReportRepo: mockReportRepo,
	}

	filters := http.QueryHeader{
		Limit:  10,
		Page:   1,
		Status: constant.FinishedStatus,
	}

	mockReports := []*report.Report{
		{
			ID:          reportId1,
			TemplateID:  templateId,
			Filters:     nil,
			Status:      constant.FinishedStatus,
			CompletedAt: &timeNow,
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
			DeletedAt:   nil,
		},
		{
			ID:          reportId2,
			TemplateID:  templateId,
			Filters:     nil,
			Status:      constant.ProcessingStatus,
			CompletedAt: nil,
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
			DeletedAt:   nil,
		},
	}

	tests := []struct {
		name           string
		filters        http.QueryHeader
		mockSetup      func()
		expectErr      bool
		expectedResult []*report.Report
		expectedCount  int
	}{
		{
			name:    "Success - Get all reports",
			filters: filters,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return(mockReports, nil)
			},
			expectErr:      false,
			expectedResult: mockReports,
			expectedCount:  2,
		},
		{
			name:    "Success - Get all reports with status filter",
			filters: filters,
			mockSetup: func() {
				filteredReports := []*report.Report{mockReports[0]} // Only finished reports
				mockReportRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return(filteredReports, nil)
			},
			expectErr:      false,
			expectedResult: []*report.Report{mockReports[0]},
			expectedCount:  1,
		},
		{
			name:    "Error - Failed to retrieve reports",
			filters: filters,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			expectedResult: nil,
			expectedCount:  0,
		},
		{
			name:    "Success - Empty result set",
			filters: filters,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindList(gomock.Any(), gomock.Any()).
					Return([]*report.Report{}, nil)
			},
			expectErr:      false, // Empty result set is valid, returns empty slice
			expectedResult: []*report.Report{},
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := reportSvc.GetAllReports(ctx, tt.filters)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)
				if tt.expectedCount > 0 {
					assert.Equal(t, tt.expectedResult[0].ID, result[0].ID)
					assert.Equal(t, tt.expectedResult[0].Status, result[0].Status)
				}
			}
		})
	}
}
