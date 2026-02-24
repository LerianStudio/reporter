// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/model"
	pkgHTTP "github.com/LerianStudio/reporter/pkg/net/http"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libOtel "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"

	// otel/attribute is used for span attribute types (no lib-commons wrapper available)
	"go.opentelemetry.io/otel/attribute"
	// otel/trace is used for trace.Span parameter types in internal helpers
	"go.opentelemetry.io/otel/trace"
)

// GenerateReportMessage contains the information needed to generate a report.
type GenerateReportMessage struct {
	// TemplateID is the unique identifier of the template to be used for report generation.
	TemplateID uuid.UUID `json:"templateId"`

	// ReportID uniquely identifies this report generation request
	ReportID uuid.UUID `json:"reportId"`

	// OutputFormat specifies the format of the generated report (e.g., html, csv, json).
	OutputFormat string `json:"outputFormat"`

	// DataQueries maps database names to tables and their fields.
	// Format: map[databaseName]map[tableName][]fieldName.
	// Example: {"onboarding": {"organization": ["name"], "ledger": ["id"]}}.
	DataQueries map[string]map[string][]string `json:"mappedFields"`

	// Filters specify advanced filtering criteria using FilterCondition for complex queries.
	// Format: map[databaseName]map[tableName]map[fieldName]model.FilterCondition
	// Example: {"db": {"table": {"created_at": {"gte": ["2025-06-01"], "lte": ["2025-06-30"]}}}}
	Filters map[string]map[string]map[string]model.FilterCondition `json:"filters"`
}

// GenerateReport handles a report generation request by loading a template file,
// processing it, and storing the final report in the report repository.
func (uc *UseCase) GenerateReport(ctx context.Context, body []byte) error {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.report.generate")
	defer span.End()

	span.SetAttributes(attribute.String("app.request.request_id", reqId))

	message, err := uc.parseMessage(ctx, body, &span, logger)
	if err != nil {
		return err
	}

	span.SetAttributes(
		attribute.String("app.request.report_id", message.ReportID.String()),
		attribute.String("app.request.template_id", message.TemplateID.String()),
	)

	if skip := uc.shouldSkipProcessing(ctx, message.ReportID, logger); skip {
		return nil
	}

	templateBytes, err := uc.loadTemplate(ctx, message, &span)
	if err != nil {
		return err
	}

	result := make(map[string]map[string][]map[string]any)

	if err := uc.queryExternalData(ctx, message, result); err != nil {
		return uc.handleErrorWithUpdate(ctx, message.ReportID, &span, "Error querying external data", err, logger)
	}

	renderedOutput, err := uc.renderTemplate(ctx, templateBytes, result, message, &span)
	if err != nil {
		return err
	}

	finalOutput, err := uc.convertToPDFIfNeeded(ctx, message, renderedOutput, &span)
	if err != nil {
		return err
	}

	if err := uc.saveReport(ctx, message, finalOutput); err != nil {
		return uc.handleErrorWithUpdate(ctx, message.ReportID, &span, "Error saving report", err, logger)
	}

	if err := uc.markReportAsFinished(ctx, message.ReportID, &span, logger); err != nil {
		return err
	}

	return nil
}

// parseMessage parses the RabbitMQ message body into GenerateReportMessage struct.
func (uc *UseCase) parseMessage(ctx context.Context, body []byte, span *trace.Span, logger log.Logger) (GenerateReportMessage, error) {
	var message GenerateReportMessage

	err := json.Unmarshal(body, &message)
	if err != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, message.ReportID, err.Error()); errUpdate != nil {
			libOtel.HandleSpanError(span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return message, errUpdate
		}

		libOtel.HandleSpanError(span, "Error unmarshalling message.", err)
		logger.Errorf("Error unmarshalling message: %s", err.Error())

		return message, err
	}

	return message, nil
}

// markReportAsFinished updates report status to finished.
func (uc *UseCase) markReportAsFinished(ctx context.Context, reportID uuid.UUID, span *trace.Span, logger log.Logger) error {
	err := uc.ReportDataRepo.UpdateReportStatusById(ctx, constant.FinishedStatus, reportID, time.Now(), nil)
	if err != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, reportID, err.Error()); errUpdate != nil {
			libOtel.HandleSpanError(span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return errUpdate
		}

		libOtel.HandleSpanError(span, "Error to update report status.", err)
		logger.Errorf("Error saving report: %s", err.Error())

		return err
	}

	return nil
}

// handleErrorWithUpdate logs error and updates report status to error.
func (uc *UseCase) handleErrorWithUpdate(ctx context.Context, reportID uuid.UUID, span *trace.Span, errorMsg string, err error, logger log.Logger) error {
	if errUpdate := uc.updateReportWithErrors(ctx, reportID, err.Error()); errUpdate != nil {
		libOtel.HandleSpanError(span, "Error to update report status with error.", errUpdate)
		logger.Errorf("Error update report status with error: %s", errUpdate.Error())

		return errUpdate
	}

	if pkgHTTP.IsBusinessError(err) {
		libOtel.HandleSpanBusinessErrorEvent(span, errorMsg, err)
	} else {
		libOtel.HandleSpanError(span, errorMsg, err)
	}

	logger.Errorf("%s: %s", errorMsg, err.Error())

	return err
}

// updateReportWithErrors updates the status of a report to "Error" with metadata containing the provided error message.
func (uc *UseCase) updateReportWithErrors(ctx context.Context, reportId uuid.UUID, errorMessage string) error {
	_, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.report.update_report_with_errors")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.report_id", reportId.String()),
	)

	metadata := make(map[string]any)
	metadata["error"] = errorMessage

	errUpdate := uc.ReportDataRepo.UpdateReportStatusById(ctx, constant.ErrorStatus,
		reportId, time.Now(), metadata)
	if errUpdate != nil {
		libOtel.HandleSpanError(&span, "Failed to update report with error status", errUpdate)

		return errUpdate
	}

	return nil
}
