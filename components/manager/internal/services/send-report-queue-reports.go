package services

import (
	"context"
	"os"
	"plugin-smart-templates/pkg/model"

	libCommons "github.com/LerianStudio/lib-commons/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
)

// SendReportQueueReports sends a report to the queue of a generation reports message to a RabbitMQ queue for further processing.
// It uses context for logger and tracer management and handles data serialization and queue message construction.
func (uc *UseCase) SendReportQueueReports(ctx context.Context, reportMessage model.ReportMessage) {
	logger := libCommons.NewLoggerFromContext(ctx)
	tracer := libCommons.NewTracerFromContext(ctx)
	reqId := libCommons.NewHeaderIDFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.send_report_queue")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err := libOpentelemetry.SetSpanAttributesFromStructWithObfuscation(&span, "app.request.payload", reportMessage)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert report message to JSON string", err)
	}

	if _, err := uc.RabbitMQRepo.ProducerDefault(
		ctx,
		os.Getenv("RABBITMQ_EXCHANGE"),
		os.Getenv("RABBITMQ_GENERATE_REPORT_KEY"),
		reportMessage,
	); err != nil {
		logger.Fatalf("Failed to send message: %s", err.Error())
	}

	logger.Infof("Report sent to genrate report queue successfully")
}
