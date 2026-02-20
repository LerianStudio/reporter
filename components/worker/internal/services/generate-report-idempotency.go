// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"

	"github.com/LerianStudio/reporter/pkg/constant"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libOtel "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// shouldSkipProcessing checks if report should be skipped due to idempotency.
func (uc *UseCase) shouldSkipProcessing(ctx context.Context, reportID uuid.UUID, logger log.Logger) bool {
	_, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.report.should_skip_processing")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.report_id", reportID.String()),
	)

	reportStatus, err := uc.checkReportStatus(ctx, reportID, logger)
	if err == nil {
		if reportStatus == constant.FinishedStatus {
			logger.Infof("Report %s is already finished, skipping reprocessing", reportID)
			return true
		}

		if reportStatus == constant.ErrorStatus {
			logger.Warnf("Report %s is in error state, skipping reprocessing", reportID)
			return true
		}
	}

	return false
}

// checkReportStatus checks the current status of a report to implement idempotency.
func (uc *UseCase) checkReportStatus(ctx context.Context, reportID uuid.UUID, logger log.Logger) (string, error) {
	_, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.report.check_report_status")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.report_id", reportID.String()),
	)

	report, err := uc.ReportDataRepo.FindByID(ctx, reportID)
	if err != nil {
		libOtel.HandleSpanError(&span, "Failed to check report status", err)

		logger.Debugf("Could not check report status for %s (may be first attempt): %v", reportID, err)

		return "", err
	}

	logger.Debugf("Report %s current status: %s", reportID, report.Status)

	return report.Status, nil
}
