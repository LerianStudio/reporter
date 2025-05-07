package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/mongodb/report"
	"testing"
	"time"
)

func Test_getReportById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)
	reportId := uuid.New()
	tempId := uuid.New()
	orgId := uuid.New()
	timeNow := time.Now()

	reportSvc := &UseCase{
		ReportRepo: mockReportRepo,
	}

	reportModel := &report.Report{
		ID:          reportId,
		TemplateID:  tempId,
		LedgerID:    nil,
		Filters:     nil,
		Status:      "Finished",
		CompletedAt: &timeNow,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
		DeletedAt:   nil,
	}

	tests := []struct {
		name           string
		orgId          uuid.UUID
		tempId         uuid.UUID
		reportId       uuid.UUID
		mockSetup      func()
		expectErr      bool
		expectedResult *report.Report
	}{
		{
			name:   "Success - Get a report by id",
			orgId:  orgId,
			tempId: tempId,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(reportModel, nil)
			},
			expectErr: false,
			expectedResult: &report.Report{
				ID:          reportId,
				TemplateID:  tempId,
				LedgerID:    nil,
				Filters:     nil,
				Status:      "Finished",
				CompletedAt: &timeNow,
				CreatedAt:   timeNow,
				UpdatedAt:   timeNow,
				DeletedAt:   nil,
			},
		},
		{
			name:   "Error - Get a report by id",
			orgId:  orgId,
			tempId: tempId,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			expectedResult: nil,
		},
		{
			name:   "Error - Get a report by id not found",
			orgId:  orgId,
			tempId: tempId,
			mockSetup: func() {
				mockReportRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, mongo.ErrNoDocuments)
			},
			expectErr:      true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := reportSvc.GetReportByID(ctx, tt.reportId, tt.orgId)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
