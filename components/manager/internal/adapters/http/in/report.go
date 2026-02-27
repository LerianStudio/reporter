// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"bytes"
	"context"
	"errors"

	"github.com/LerianStudio/reporter/components/manager/internal/services"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/model"
	_ "github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v3/commons"
	libConstants "github.com/LerianStudio/lib-commons/v3/commons/constants"
	commonsHttp "github.com/LerianStudio/lib-commons/v3/commons/net/http"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v3/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// ReportHandler handles HTTP requests for report operations.
type ReportHandler struct {
	service *services.UseCase
}

// NewReportHandler creates a new ReportHandler with the given service dependency.
// It returns an error if service is nil.
func NewReportHandler(service *services.UseCase) (*ReportHandler, error) {
	if service == nil {
		return nil, errors.New("service must not be nil for ReportHandler")
	}

	return &ReportHandler{service: service}, nil
}

// CreateReport is a method that creates a report.
//
//	@Summary		Create a Report
//	@Description	Create a Report of existent template with the input payload
//	@Tags			Reports
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			X-Idempotency	header		string					false	"Client-provided idempotency key to prevent duplicate report creation"
//	@Param			reports			body		model.CreateReportInput	true	"Report Input"
//	@Success		201				{object}	report.Report
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		401				{object}	pkg.HTTPError
//	@Failure		403				{object}	pkg.HTTPError
//	@Failure		404				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/reports [post]
func (rh *ReportHandler) CreateReport(p any, c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.report.create")
	defer span.End()

	// Extract X-Idempotency header and inject into context for the service layer
	if idempotencyKey := c.Get(libConstants.IdempotencyKey); idempotencyKey != "" {
		ctx = context.WithValue(ctx, constant.IdempotencyKeyCtx, idempotencyKey)
	}

	replayed := false
	ctx = context.WithValue(ctx, constant.IdempotencyReplayedCtx, &replayed)

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

	reportOut, err := rh.service.CreateReport(ctx, payload)
	if err != nil {
		if http.IsBusinessError(err) {
			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to create report", err)
		} else {
			libOpentelemetry.HandleSpanError(&span, "Failed to create report", err)
		}

		return http.WithError(c, err)
	}

	logger.Infof("Successfully created a report %v", reportOut)

	if replayed {
		c.Set(libConstants.IdempotencyReplayed, "true")
	}

	return commonsHttp.Created(c, reportOut)
}

// GetDownloadReport is a method to make download of a report.
//
//	@Summary		Download a Report
//	@Description	Make a download of a Report passing the ID
//	@Tags			Reports
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id				path		string	true	"Report ID"
//	@Success		200				{file}		any
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		401				{object}	pkg.HTTPError
//	@Failure		403				{object}	pkg.HTTPError
//	@Failure		404				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/reports/{id}/download [get]
func (rh *ReportHandler) GetDownloadReport(c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.report.get_download")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating download of Report with ID: %s", id)

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.report_id", id.String()),
	)

	fileBytes, fileName, contentType, err := rh.service.DownloadReport(ctx, id)
	if err != nil {
		if http.IsBusinessError(err) {
			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to download report", err)
		} else {
			libOpentelemetry.HandleSpanError(&span, "Failed to download report", err)
		}

		logger.Errorf("Failed to download Report with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")

	logger.Infof("Successfully downloaded Report with ID: %s", id)

	return c.SendStream(bytes.NewReader(fileBytes))
}

// GetReport is a method to get a report information.
//
//	@Summary		Get a Report
//	@Description	Get information of a Report passing the ID
//	@Tags			Reports
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id					path		string	true	"Report ID"
//	@Success		200					{object}	report.Report
//	@Failure		400					{object}	pkg.HTTPError
//	@Failure		401					{object}	pkg.HTTPError
//	@Failure		403					{object}	pkg.HTTPError
//	@Failure		404					{object}	pkg.HTTPError
//	@Failure		500					{object}	pkg.HTTPError
//	@Router			/v1/reports/{id}																					 [get]
func (rh *ReportHandler) GetReport(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.report.get")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating get a Report with ID: %s", id)

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.report_id", id.String()),
	)

	reportModel, err := rh.service.GetReportByID(ctx, id)
	if err != nil {
		if http.IsBusinessError(err) {
			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to retrieve report on query", err)
		} else {
			libOpentelemetry.HandleSpanError(&span, "Failed to retrieve report on query", err)
		}

		logger.Errorf("Failed to retrieve Report with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	return commonsHttp.OK(c, reportModel)
}

// GetAllReports is a method that recovery all Reports information.
//
//	@Summary		Get all reports
//	@Description	List all the reports
//	@Tags			Reports
//	@Produce		json
//	@Security		BearerAuth
//	@Param			status			query		string	false	"Report status (processing, finished, error)"
//	@Param			template_id		query		string	false	"Template ID (also accepts templateId)"
//	@Param			created_at		query		string	false	"Created at date, YYYY-MM-DD (also accepts createdAt)"
//	@Param			limit			query		int		false	"Limit"	default(10)
//	@Param			page			query		int		false	"Page"	default(1)
//	@Success		200				{object}	model.Pagination{items=[]report.Report,page=int,limit=int,total=int}
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		401				{object}	pkg.HTTPError
//	@Failure		403				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/reports [get]
func (rh *ReportHandler) GetAllReports(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.report.get_all")
	defer span.End()

	headerParams, err := http.ValidateParameters(c.Queries())
	if err != nil {
		libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to validate query parameters", err)

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

	reports, err := rh.service.GetAllReports(ctx, *headerParams)
	if err != nil {
		if http.IsBusinessError(err) {
			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to retrieve all Reports on query", err)
		} else {
			libOpentelemetry.HandleSpanError(&span, "Failed to retrieve all Reports on query", err)
		}

		logger.Errorf("Failed to retrieve all Reports, Error: %s", err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully retrieved all Reports")

	pagination.SetItems(reports)
	pagination.SetTotal(len(reports))

	return commonsHttp.OK(c, pagination)
}
