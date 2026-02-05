// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"os"

	"github.com/LerianStudio/reporter/pkg/model"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
)

// SendReportQueueReports sends a report to the queue of a generation reports message to a RabbitMQ queue for further processing.
// It uses context for logger and tracer management and handles data serialization and queue message construction.
func (uc *UseCase) SendReportQueueReports(ctx context.Context, reportMessage model.ReportMessage) {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.send_report_queue")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", reportMessage)
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
