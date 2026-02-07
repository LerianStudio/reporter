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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSendReportQueueReports(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)

	reportID := uuid.New()
	templateID := uuid.New()

	tests := []struct {
		name          string
		reportMessage model.ReportMessage
		mockSetup     func()
		expectErr     bool
		errContains   string
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &UseCase{
				RabbitMQRepo:              mockRabbitMQ,
				RabbitMQExchange:          "test-exchange",
				RabbitMQGenerateReportKey: "test-key",
			}

			ctx := context.Background()

			err := svc.SendReportQueueReports(ctx, tt.reportMessage)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSendReportQueueReports_WithDifferentOutputFormats(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)

	reportID := uuid.New()
	templateID := uuid.New()

	outputFormats := []string{"pdf", "xml", "html", "txt", "csv"}

	for _, format := range outputFormats {
		format := format
		t.Run("Success - OutputFormat "+format, func(t *testing.T) {
			mockRabbitMQ.EXPECT().
				ProducerDefault(gomock.Any(), "test-exchange", "test-key", gomock.Any()).
				Return(nil, nil)

			svc := &UseCase{
				RabbitMQRepo:              mockRabbitMQ,
				RabbitMQExchange:          "test-exchange",
				RabbitMQGenerateReportKey: "test-key",
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
