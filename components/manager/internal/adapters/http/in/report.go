// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"bytes"

	"github.com/LerianStudio/reporter/components/manager/internal/services"
	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/model"
	_ "github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/net/http"
	templateUtils "github.com/LerianStudio/reporter/pkg/template_utils"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

type ReportHandler struct {
	Service *services.UseCase
}

// CreateReport is a method that creates a report.
//
//	@Summary		Create a Report
//	@Description	Create a Report of existent template with the input payload
//	@Tags			Reports
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string					false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Param			reports			body		model.CreateReportInput	true	"Report Input"
//	@Success		201				{object}	report.Report
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		404				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/reports [post]
func (rh *ReportHandler) CreateReport(p any, c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.create_report")
	defer span.End()

	c.SetUserContext(ctx)

	payload := p.(*model.CreateReportInput)
	logger.Infof("Request to create a report with details: %#v", payload)

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", payload)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert payload to JSON string", err)
	}

	reportOut, err := rh.Service.CreateReport(ctx, payload)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to create report", err)

		return http.WithError(c, err)
	}

	logger.Infof("Successfully created a report %v", reportOut)

	return http.OK(c, reportOut)
}

// GetDownloadReport is a method to make download of a report.
//
//	@Summary		Download a Report
//	@Description	Make a download of a Report passing the ID
//	@Tags			Reports
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Param			id				path		string	true	"Report ID"
//	@Success		200				{file}		any
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		404				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/reports/{id}/download [get]
func (rh *ReportHandler) GetDownloadReport(c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_report_download")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating get a Report with ID: %s", id)

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.report_id", id.String()),
	)

	reportModel, err := rh.Service.GetReportByID(ctx, id)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to retrieve report on query", err)

		logger.Errorf("Failed to retrieve Report with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	if reportModel.Status != constant.FinishedStatus {
		errStatus := pkg.ValidateBusinessError(constant.ErrReportStatusNotFinished, "")
		libOpentelemetry.HandleSpanError(&span, "Report is not Finished", errStatus)

		logger.Errorf("Report with ID %s is not Finished", id)

		return http.WithError(c, errStatus)
	}

	templateModel, err := rh.Service.GetTemplateByID(ctx, reportModel.TemplateID)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to retrieve template on query", err)

		logger.Errorf("Failed to retrieve Template with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	objectName := templateModel.ID.String() + "/" + reportModel.ID.String() + "." + templateModel.OutputFormat

	fileBytes, errFile := rh.Service.ReportSeaweedFS.Get(ctx, objectName)
	if errFile != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to download file from SeaweedFS", errFile)

		logger.Errorf("Failed to download file from SeaweedFS: %s", errFile.Error())

		return http.WithError(c, errFile)
	}

	logger.Infof("Downloaded report file from SeaweedFS: %s (size: %d bytes)", objectName, len(fileBytes))

	contentType := templateUtils.GetMimeType(templateModel.OutputFormat)
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename=\""+objectName+"\"")

	return c.SendStream(bytes.NewReader(fileBytes))
}

// GetReport is a method to get a report information.
//
//	@Summary		Get a Report
//	@Description	Get information of a Report passing the ID
//	@Tags			Reports
//	@Accept			json
//	@Produce		json
//	@Param			Authorization		header		string	false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Param			id					path		string	true	"Report ID"
//	@Success		200					{object}	report.Report
//	@Failure		400					{object}	pkg.HTTPError
//	@Failure		404					{object}	pkg.HTTPError
//	@Failure		500					{object}	pkg.HTTPError
//	@Router			/v1/reports/{id}																					 [get]
func (rh *ReportHandler) GetReport(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_report")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating get a Report with ID: %s", id)

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.report_id", id.String()),
	)

	reportModel, err := rh.Service.GetReportByID(ctx, id)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to retrieve report on query", err)

		logger.Errorf("Failed to retrieve Report with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	return http.OK(c, reportModel)
}

// GetAllReports is a method that recovery all Reports information.
//
//	@Summary		Get all reports
//	@Description	List all the reports
//	@Tags			Reports
//	@Produce		json
//	@Param			Authorization	header		string	false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Param			status			query		string	false	"Report status (processing, finished, error)"
//	@Param			templateId		query		string	false	"Template ID"
//	@Param			createdAt		query		string	false	"Created at date (YYYY-MM-DD)"
//	@Param			limit			query		int		false	"Limit"	default(10)
//	@Param			page			query		int		false	"Page"	default(1)
//	@Success		200				{object}	model.Pagination{items=[]report.Report,page=int,limit=int,total=int}
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/reports [get]
func (rh *ReportHandler) GetAllReports(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_all_reports")
	defer span.End()

	headerParams, err := http.ValidateParameters(c.Queries())
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to validate query parameters", err)

		logger.Errorf("Failed to validate query parameters, Error: %s", err.Error())

		return http.WithError(c, err)
	}

	pagination := model.Pagination{
		Limit: headerParams.Limit,
		Page:  headerParams.Page,
	}

	logger.Infof("Initiating retrieval all reports")

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err = libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.query_params", headerParams)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert query params to JSON string", err)
	}

	reports, err := rh.Service.GetAllReports(ctx, *headerParams)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to retrieve all Reports on query", err)

		logger.Errorf("Failed to retrieve all Reports, Error: %s", err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully retrieved all Reports")

	pagination.SetItems(reports)
	pagination.SetTotal(len(reports))

	return http.OK(c, pagination)
}
