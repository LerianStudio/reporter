package services

import (
	"context"
	"plugin-smart-templates/v2/components/manager/internal/adapters/rabbitmq"
	"plugin-smart-templates/v2/pkg/constant"
	"plugin-smart-templates/v2/pkg/model"
	"plugin-smart-templates/v2/pkg/mongodb/report"
	"plugin-smart-templates/v2/pkg/mongodb/template"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_createReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockReportRepo := report.NewMockRepository(ctrl)
	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)
	reportId := uuid.New()
	tempId := uuid.New()
	orgId := uuid.New()

	reportSvc := &UseCase{
		TemplateRepo: mockTempRepo,
		ReportRepo:   mockReportRepo,
		RabbitMQRepo: mockRabbitMQ,
	}

	mappedFields := map[string]map[string][]string{
		"midaz_transaction_metadata": {
			"transaction": {"metadata"},
		},
		"midaz_onboarding": {
			"asset":        {"name", "type", "code"},
			"organization": {"legal_document", "legal_name", "doing_business_as", "address"},
			"ledger":       {"name", "status"},
		},
	}

	outputFormat := "xml"

	reportInput := &model.CreateReportInput{
		TemplateID: tempId.String(),
		Filters:    nil,
	}

	reportEntity := &report.Report{
		ID:         reportId,
		TemplateID: tempId,
		Filters:    nil,
		Status:     "processing",
	}

	tests := []struct {
		name           string
		reportInput    *model.CreateReportInput
		orgId          uuid.UUID
		mockSetup      func()
		expectErr      bool
		expectedResult *report.Report
	}{
		{
			name:        "Success - Create a report",
			reportInput: reportInput,
			orgId:       orgId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(reportEntity, nil)

				mockRabbitMQ.EXPECT().
					ProducerDefault(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectErr: false,
			expectedResult: &report.Report{
				ID:         reportId,
				TemplateID: tempId,
				Filters:    nil,
				Status:     "processing",
			},
		},
		{
			name:        "Error - Find mapped fields and output format",
			reportInput: reportInput,
			orgId:       orgId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			expectedResult: nil,
		},
		{
			name:        "Error - Create report",
			reportInput: reportInput,
			orgId:       orgId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			expectedResult: nil,
		},

		{
			name:        "Error - Send message on RabbitMQ",
			reportInput: reportInput,
			orgId:       orgId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(reportEntity, nil)

				mockRabbitMQ.EXPECT().
					ProducerDefault(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr: false,
			expectedResult: &report.Report{
				ID:         reportId,
				TemplateID: tempId,
				Filters:    nil,
				Status:     "processing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := reportSvc.CreateReport(ctx, tt.reportInput, tt.orgId)

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
