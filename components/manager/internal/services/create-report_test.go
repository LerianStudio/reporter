// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"testing"

	"github.com/LerianStudio/reporter/components/manager/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/model"
	"github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"

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
		mockSetup      func()
		expectErr      bool
		expectedResult *report.Report
	}{
		{
			name:        "Success - Create a report",
			reportInput: reportInput,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
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
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any()).
					Return(nil, nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			expectedResult: nil,
		},
		{
			name:        "Error - Create report",
			reportInput: reportInput,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			expectedResult: nil,
		},

		{
			name:        "Error - Send message on RabbitMQ",
			reportInput: reportInput,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
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
			result, err := reportSvc.CreateReport(ctx, tt.reportInput)

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

func TestConvertFiltersToMappedFieldsType(t *testing.T) {
	uc := &UseCase{}

	tests := []struct {
		name     string
		input    map[string]map[string]map[string]model.FilterCondition
		expected map[string]map[string][]string
	}{
		{
			name: "Single datasource, single table, single field",
			input: map[string]map[string]map[string]model.FilterCondition{
				"midaz_onboarding": {
					"organization": {
						"id": {Equals: []any{"123"}},
					},
				},
			},
			expected: map[string]map[string][]string{
				"midaz_onboarding": {
					"organization": {"id"},
				},
			},
		},
		{
			name: "Single datasource, single table, multiple fields (max 3)",
			input: map[string]map[string]map[string]model.FilterCondition{
				"midaz_onboarding": {
					"organization": {
						"id":     {Equals: []any{"123"}},
						"name":   {In: []any{"Test"}},
						"status": {Equals: []any{"active"}},
						"extra":  {Equals: []any{"ignored"}},
					},
				},
			},
			expected: map[string]map[string][]string{
				"midaz_onboarding": {
					"organization": {"id", "name", "status"},
				},
			},
		},
		{
			name: "Multiple datasources and tables",
			input: map[string]map[string]map[string]model.FilterCondition{
				"datasource_one": {
					"organization": {
						"id": {Equals: []any{"123"}},
					},
					"ledger": {
						"status": {Equals: []any{"active"}},
					},
				},
				"datasource_two": {
					"analytics.transfers": {
						"amount": {GreaterThan: []any{100}},
					},
				},
			},
			expected: map[string]map[string][]string{
				"datasource_one": {
					"organization": {"id"},
					"ledger":       {"status"},
				},
				"datasource_two": {
					"analytics.transfers": {"amount"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uc.convertFiltersToMappedFieldsType(tt.input)

			// Verify structure matches (can't guarantee order of keys in maps)
			assert.Equal(t, len(tt.expected), len(result))
			for datasource, tables := range tt.expected {
				assert.Contains(t, result, datasource)
				assert.Equal(t, len(tables), len(result[datasource]))
				for table, fields := range tables {
					assert.Contains(t, result[datasource], table)
					assert.Equal(t, len(fields), len(result[datasource][table]))
				}
			}
		})
	}
}
