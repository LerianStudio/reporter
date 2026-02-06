// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	templateUtils "github.com/LerianStudio/reporter/pkg/template_utils"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// DownloadReport retrieves the report file bytes, file name, and content type for a given report ID.
// It validates the report status, fetches the associated template for output format, constructs the
// storage object name, and downloads the file from object storage.
func (uc *UseCase) DownloadReport(ctx context.Context, id uuid.UUID) ([]byte, string, string, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.download_report")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.report_id", id.String()),
	)

	logger.Infof("Downloading report for id %v", id)

	// Fetch the report
	reportModel, err := uc.GetReportByID(ctx, id)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to retrieve report on query", err)

		logger.Errorf("Failed to retrieve Report with ID: %s, Error: %s", id, err.Error())

		return nil, "", "", err
	}

	// Validate report status
	if reportModel.Status != constant.FinishedStatus {
		errStatus := pkg.ValidateBusinessError(constant.ErrReportStatusNotFinished, "")

		logger.Errorf("Report with ID %s is not Finished", id)

		return nil, "", "", errStatus
	}

	// Fetch the associated template for output format
	templateModel, err := uc.GetTemplateByID(ctx, reportModel.TemplateID)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to retrieve template on query", err)

		logger.Errorf("Failed to retrieve Template with ID: %s, Error: %s", reportModel.TemplateID, err.Error())

		return nil, "", "", err
	}

	// Construct the storage object name
	objectName := templateModel.ID.String() + "/" + reportModel.ID.String() + "." + templateModel.OutputFormat

	// Download the file from storage
	fileBytes, errFile := uc.ReportSeaweedFS.Get(ctx, objectName)
	if errFile != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to download file from storage", errFile)

		logger.Errorf("Failed to download file from storage: %s", errFile.Error())

		return nil, "", "", errFile
	}

	logger.Infof("Downloaded report file from storage: %s (size: %d bytes)", objectName, len(fileBytes))

	// Determine content type from the template output format
	contentType := templateUtils.GetMimeType(templateModel.OutputFormat)

	// Construct proper filename for download (reportID.extension, not templateID/reportID.extension)
	fileName := reportModel.ID.String() + "." + templateModel.OutputFormat

	return fileBytes, fileName, contentType, nil
}
