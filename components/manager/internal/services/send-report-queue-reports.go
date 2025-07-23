package services

import (
	"context"
	"os"
	"plugin-smart-templates/pkg/model"

	libCommons "github.com/LerianStudio/lib-commons/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/commons/opentelemetry"
)

// SendReportQueueReports sends a report to the queue of a generation reports message to a RabbitMQ queue for further processing.
// It uses context for logger and tracer management and handles data serialization and queue message construction.
func (uc *UseCase) SendReportQueueReports(ctx context.Context, reportMessage model.ReportMessage) {
	logger := libCommons.NewLoggerFromContext(ctx)
	tracer := libCommons.NewTracerFromContext(ctx)

	ctxLogTransaction, spanLogTransaction := tracer.Start(ctx, "command.send_report_queue_reports")
	defer spanLogTransaction.End()

	err := libOpentelemetry.SetSpanAttributesFromStruct(&spanLogTransaction, "report_message", reportMessage)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanLogTransaction, "Failed to convert report message to JSON string", err)
	}

	if _, err := uc.RabbitMQRepo.ProducerDefault(
		ctxLogTransaction,
		os.Getenv("RABBITMQ_EXCHANGE"),
		os.Getenv("RABBITMQ_GENERATE_REPORT_KEY"),
		reportMessage,
	); err != nil {
		logger.Fatalf("Failed to send message: %s", err.Error())
	}

	logger.Infof("Report sent to genrate report queue successfully")
}
