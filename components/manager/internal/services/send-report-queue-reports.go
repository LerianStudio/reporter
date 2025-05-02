package services

import (
	"context"
	libCommons "github.com/LerianStudio/lib-commons/commons"
	"os"
	"plugin-template-engine/pkg/model"
)

// SendReportQueueReports sends report to queue of generation reports message to a RabbitMQ queue for further processing.
// It utilizes context for logger and tracer management and handles data serialization and queue message construction.
func (uc *UseCase) SendReportQueueReports(ctx context.Context, reportMessage model.ReportMessage) {
	logger := libCommons.NewLoggerFromContext(ctx)
	tracer := libCommons.NewTracerFromContext(ctx)

	ctxLogTransaction, spanLogTransaction := tracer.Start(ctx, "command.send_report_queue_reports")
	defer spanLogTransaction.End()

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
