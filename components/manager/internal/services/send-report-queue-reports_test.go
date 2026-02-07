// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"testing"

	"github.com/LerianStudio/reporter/components/manager/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/pkg/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_SendReportQueueReports(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)

	reportID := uuid.New()
	templateID := uuid.New()

	// Set environment variables for the test
	t.Setenv("RABBITMQ_EXCHANGE", "test-exchange")
	t.Setenv("RABBITMQ_GENERATE_REPORT_KEY", "test-key")

	tests := []struct {
		name          string
		reportMessage model.ReportMessage
		mockSetup     func()
		expectErr     bool
	}{
		{
			name: "Success - Send report to queue",
			reportMessage: model.ReportMessage{
				ReportID:     reportID,
				TemplateID:   templateID,
				OutputFormat: "pdf",
				MappedFields: map[string]map[string][]string{
					"midaz_onboarding": {
						"asset": {"name", "type", "code"},
					},
				},
				Filters: nil,
			},
			mockSetup: func() {
				mockRabbitMQ.EXPECT().
					ProducerDefault(gomock.Any(), "test-exchange", "test-key", gomock.Any()).
					Return(nil, nil)
			},
		},
		{
			name: "Success - Send report with filters",
			reportMessage: model.ReportMessage{
				ReportID:     reportID,
				TemplateID:   templateID,
				OutputFormat: "xml",
				MappedFields: map[string]map[string][]string{
					"midaz_transaction_metadata": {
						"transaction": {"metadata"},
					},
				},
				Filters: map[string]map[string]map[string]model.FilterCondition{
					"midaz_transaction_metadata": {
						"transaction": {
							"id": {
								Equals: []any{"123"},
							},
						},
					},
				},
			},
			mockSetup: func() {
				mockRabbitMQ.EXPECT().
					ProducerDefault(gomock.Any(), "test-exchange", "test-key", gomock.Any()).
					Return(nil, nil)
			},
		},
		{
			name: "Success - Send report with empty mapped fields",
			reportMessage: model.ReportMessage{
				ReportID:     reportID,
				TemplateID:   templateID,
				OutputFormat: "html",
				MappedFields: map[string]map[string][]string{},
				Filters:      nil,
			},
			mockSetup: func() {
				mockRabbitMQ.EXPECT().
					ProducerDefault(gomock.Any(), "test-exchange", "test-key", gomock.Any()).
					Return(nil, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &UseCase{
				RabbitMQRepo: mockRabbitMQ,
			}

			ctx := context.Background()

			err := svc.SendReportQueueReports(ctx, tt.reportMessage)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_SendReportQueueReports_WithDifferentOutputFormats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)

	reportID := uuid.New()
	templateID := uuid.New()

	t.Setenv("RABBITMQ_EXCHANGE", "test-exchange")
	t.Setenv("RABBITMQ_GENERATE_REPORT_KEY", "test-key")

	outputFormats := []string{"pdf", "xml", "html", "txt", "csv"}

	for _, format := range outputFormats {
		t.Run("Success - OutputFormat "+format, func(t *testing.T) {
			mockRabbitMQ.EXPECT().
				ProducerDefault(gomock.Any(), "test-exchange", "test-key", gomock.Any()).
				Return(nil, nil)

			svc := &UseCase{
				RabbitMQRepo: mockRabbitMQ,
			}

			reportMessage := model.ReportMessage{
				ReportID:     reportID,
				TemplateID:   templateID,
				OutputFormat: format,
				MappedFields: map[string]map[string][]string{
					"test_db": {
						"test_table": {"field1", "field2"},
					},
				},
			}

			ctx := context.Background()

			err := svc.SendReportQueueReports(ctx, reportMessage)
			assert.NoError(t, err)
		})
	}
}
