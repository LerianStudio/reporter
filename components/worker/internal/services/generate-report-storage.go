// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"strings"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libOtel "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	// otel/trace is used for trace.Tracer parameter type in saveReport
	"go.opentelemetry.io/otel/trace"
)

// mimeTypes maps file extensions to their corresponding MIME content types
var mimeTypes = map[string]string{
	"txt":  "text/plain",
	"html": "text/html",
	"json": "application/json",
	"csv":  "text/csv",
}

// saveReport handles saving the generated report file to the report repository and logs any encountered errors.
// It determines the object name, content type, and stores the file using the ReportSeaweedFS interface.
// If ReportTTL is configured, the file will be saved with TTL (Time To Live).
// Returns an error if the file storage operation fails.
func (uc *UseCase) saveReport(ctx context.Context, tracer trace.Tracer, message GenerateReportMessage, out string, logger log.Logger) error {
	ctx, spanSaveReport := tracer.Start(ctx, "service.report.save_report")
	defer spanSaveReport.End()

	outputFormat := strings.ToLower(message.OutputFormat)
	contentType := getContentType(outputFormat)
	objectName := message.TemplateID.String() + "/" + message.ReportID.String() + "." + outputFormat

	err := uc.ReportSeaweedFS.Put(ctx, objectName, contentType, []byte(out), uc.ReportTTL)
	if err != nil {
		libOtel.HandleSpanError(&spanSaveReport, "Error putting report file.", err)

		logger.Errorf("Error putting report file: %s", err.Error())

		return err
	}

	if uc.ReportTTL != "" {
		logger.Infof("Saving report with TTL: %s", uc.ReportTTL)
	}

	return nil
}

// getContentType returns the MIME type for a given file extension.
// If the extension is not recognized, it returns "text/plain".
func getContentType(ext string) string {
	if contentType, ok := mimeTypes[ext]; ok {
		return contentType
	}

	return "text/plain"
}
