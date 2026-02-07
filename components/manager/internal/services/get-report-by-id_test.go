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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"
)

func TestGetReportByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)
	reportId := uuid.New()
	tempId := uuid.New()
	timeNow := time.Now()

	reportSvc := &UseCase{
		ReportRepo: mockReportRepo,
	}

	reportModel := &report.Report{
		ID:          reportId,
		TemplateID:  tempId,
		Filters:     nil,
		Status:      constant.FinishedStatus,
		CompletedAt: &timeNow,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
		DeletedAt:   nil,
	}

	tests := []struct {
		name           string
		tempId         uuid.UUID
		reportId       uuid.UUID
		mockSetup      func()
		expectErr      bool
		errContains    string
		expectedResult *report.Report
	}{
		{
			name:   "Success - Get a report by id",
			tempId: tempId,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(reportModel, nil)
			},
			expectErr: false,
			expectedResult: &report.Report{
				ID:          reportId,
				TemplateID:  tempId,
				Filters:     nil,
				Status:      constant.FinishedStatus,
				CompletedAt: &timeNow,
				CreatedAt:   timeNow,
				UpdatedAt:   timeNow,
				DeletedAt:   nil,
			},
		},
		{
			name:   "Error - Get a report by id",
			tempId: tempId,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			errContains:    constant.ErrInternalServer.Error(),
			expectedResult: nil,
		},
		{
			name:   "Error - Get a report by id not found",
			tempId: tempId,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(nil, mongo.ErrNoDocuments)
			},
			expectErr:      true,
			errContains:    "No report entity was found",
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := reportSvc.GetReportByID(ctx, tt.reportId)

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
