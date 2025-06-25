package in

import (
	"bytes"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"plugin-smart-templates/components/manager/internal/services"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/model"
	_ "plugin-smart-templates/pkg/mongodb/report"
	"plugin-smart-templates/pkg/net/http"
	templateUtils "plugin-smart-templates/pkg/template_utils"
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
//	@Param			Authorization		header		string					false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Param			X-Organization-Id	header		string					true	"Organization ID"
//	@Param			reports				body		model.CreateReportInput	true	"Report Input"
//	@Success		201					{object}	report.Report
//	@Router			/v1/reports [post]
func (rh *ReportHandler) CreateReport(p any, c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.create_report")
	defer span.End()

	c.SetUserContext(ctx)

	ctx, span = tracer.Start(ctx, "handler.create_report")
	defer span.End()

	organizationID := c.Locals("X-Organization-Id").(uuid.UUID)
	payload := p.(*model.CreateReportInput)
	logger.Infof("Request to create a report with details: %#v", payload)

	reportOut, err := rh.Service.CreateReport(ctx, payload, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to create report", err)

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
//	@Param			Authorization		header	string	false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Param			X-Organization-Id	header	string	true	"Organization ID"
//	@Param			id					path	string	true	"Report ID"
//	@Success		200					{file}	any
//	@Router			/v1/reports/{id}/download [get]
func (rh *ReportHandler) GetDownloadReport(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_report_download")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating get a Report with ID: %s", id)

	organizationID := c.Locals("X-Organization-Id").(uuid.UUID)

	reportModel, err := rh.Service.GetReportByID(ctx, id, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve report on query", err)

		logger.Errorf("Failed to retrieve Report with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	if reportModel.Status != constant.FinishedStatus {
		errStatus := pkg.ValidateBusinessError(constant.ErrReportStatusNotFinished, "")
		opentelemetry.HandleSpanError(&span, "Report is not Finished", errStatus)

		logger.Errorf("Report with ID %s is not Finished", id)

		return http.WithError(c, errStatus)
	}

	templateModel, err := rh.Service.GetTemplateByID(ctx, reportModel.TemplateID, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve template on query", err)

		logger.Errorf("Failed to retrieve Template with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	objectName := templateModel.ID.String() + "/" + reportModel.ID.String() + "." + templateModel.OutputFormat

	fileBytes, errFile := rh.Service.ReportMinio.Get(ctx, objectName)
	if errFile != nil {
		logger.Errorf("Failed to download file from MinIO: %s", errFile.Error())
		return http.WithError(c, errFile)
	}

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
//	@Param			X-Organization-Id	header		string	true	"Organization ID"
//	@Param			id					path		string	true	"Report ID"
//	@Success		200					{object}	report.Report
//	@Router			/v1/reports/{id}													 [get]
func (rh *ReportHandler) GetReport(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_report")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating get a Report with ID: %s", id)

	organizationID := c.Locals("X-Organization-Id").(uuid.UUID)

	reportModel, err := rh.Service.GetReportByID(ctx, id, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve report on query", err)

		logger.Errorf("Failed to retrieve Report with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	return http.OK(c, reportModel)
}
