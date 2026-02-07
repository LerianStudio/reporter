// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"

	"github.com/LerianStudio/reporter/pkg/constant"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/google/uuid"
)

// shouldSkipProcessing checks if report should be skipped due to idempotency.
func (uc *UseCase) shouldSkipProcessing(ctx context.Context, reportID uuid.UUID, logger log.Logger) bool {
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
	report, err := uc.ReportDataRepo.FindByID(ctx, reportID)
	if err != nil {
		logger.Debugf("Could not check report status for %s (may be first attempt): %v", reportID, err)
		return "", err
	}

	logger.Debugf("Report %s current status: %s", reportID, report.Status)

	return report.Status, nil
}
