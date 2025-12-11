package services

import (
	"context"
	"os"
	"testing"

	"github.com/LerianStudio/reporter/v4/components/manager/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/v4/pkg/model"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestSendReportQueueReports(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)

	// Set environment variables for the test
	os.Setenv("RABBITMQ_EXCHANGE", "test-exchange")
	os.Setenv("RABBITMQ_GENERATE_REPORT_KEY", "test-routing-key")
	defer func() {
		os.Unsetenv("RABBITMQ_EXCHANGE")
		os.Unsetenv("RABBITMQ_GENERATE_REPORT_KEY")
	}()

	reportSvc := &UseCase{
		RabbitMQRepo: mockRabbitMQ,
	}

	reportID := uuid.New()
	templateID := uuid.New()

	reportMessage := model.ReportMessage{
		ReportID:     reportID,
		TemplateID:   templateID,
		OutputFormat: "pdf",
		MappedFields: map[string]map[string][]string{
			"datasource1": {
				"table1": {"field1", "field2"},
			},
		},
		Filters: nil,
	}

	tests := []struct {
		name          string
		reportMessage model.ReportMessage
		mockSetup     func()
	}{
		{
			name:          "Success - Send report to queue",
			reportMessage: reportMessage,
			mockSetup: func() {
				mockRabbitMQ.EXPECT().
					ProducerDefault(gomock.Any(), "test-exchange", "test-routing-key", reportMessage).
					Return(nil, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()

			// This function doesn't return anything, just verify it doesn't panic
			reportSvc.SendReportQueueReports(ctx, tt.reportMessage)
		})
	}
}

func TestSendReportQueueReports_WithFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)

	os.Setenv("RABBITMQ_EXCHANGE", "test-exchange")
	os.Setenv("RABBITMQ_GENERATE_REPORT_KEY", "test-routing-key")
	defer func() {
		os.Unsetenv("RABBITMQ_EXCHANGE")
		os.Unsetenv("RABBITMQ_GENERATE_REPORT_KEY")
	}()

	reportSvc := &UseCase{
		RabbitMQRepo: mockRabbitMQ,
	}

	reportMessage := model.ReportMessage{
		ReportID:     uuid.New(),
		TemplateID:   uuid.New(),
		OutputFormat: "csv",
		MappedFields: map[string]map[string][]string{
			"datasource": {
				"users": {"id", "name", "email"},
			},
		},
		Filters: map[string]map[string]map[string]model.FilterCondition{
			"datasource": {
				"users": {
					"status": {
						Equals: []any{"active"},
					},
				},
			},
		},
	}

	mockRabbitMQ.EXPECT().
		ProducerDefault(gomock.Any(), "test-exchange", "test-routing-key", reportMessage).
		Return(nil, nil)

	ctx := context.Background()
	reportSvc.SendReportQueueReports(ctx, reportMessage)
}
