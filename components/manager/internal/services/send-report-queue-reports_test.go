// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"os"
	"testing"

	"github.com/LerianStudio/reporter/components/manager/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/pkg/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSendReportQueueReports(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)

	reportID := uuid.New()
	templateID := uuid.New()

	// Set environment variables for the test
	os.Setenv("RABBITMQ_EXCHANGE", "test-exchange")
	os.Setenv("RABBITMQ_GENERATE_REPORT_KEY", "test-key")
	defer func() {
		os.Unsetenv("RABBITMQ_EXCHANGE")
		os.Unsetenv("RABBITMQ_GENERATE_REPORT_KEY")
	}()

	tests := []struct {
		name          string
		reportMessage model.ReportMessage
		mockSetup     func()
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

			// This function doesn't return anything, so we just verify it doesn't panic
			// and the mock expectations are met
			assert.NotPanics(t, func() {
				svc.SendReportQueueReports(ctx, tt.reportMessage)
			})
		})
	}
}

func TestSendReportQueueReports_WithDifferentOutputFormats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)

	reportID := uuid.New()
	templateID := uuid.New()

	os.Setenv("RABBITMQ_EXCHANGE", "test-exchange")
	os.Setenv("RABBITMQ_GENERATE_REPORT_KEY", "test-key")
	defer func() {
		os.Unsetenv("RABBITMQ_EXCHANGE")
		os.Unsetenv("RABBITMQ_GENERATE_REPORT_KEY")
	}()

	outputFormats := []string{"pdf", "xml", "html", "txt", "csv"}

	for _, format := range outputFormats {
		t.Run("OutputFormat_"+format, func(t *testing.T) {
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
			assert.NotPanics(t, func() {
				svc.SendReportQueueReports(ctx, reportMessage)
			})
		})
	}
}
