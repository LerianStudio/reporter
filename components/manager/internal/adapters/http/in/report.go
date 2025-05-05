package in

import (
	"bytes"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"plugin-template-engine/components/manager/internal/services"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/model"
	"plugin-template-engine/pkg/net/http"
	templateUtils "plugin-template-engine/pkg/template_utils"
	"time"
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
//	@Param			Authorization		header		string					true	"The authorization token in the 'Bearer	access_token' format."
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

	ctx, span = tracer.Start(ctx, "handler.create_template")
	defer span.End()

	organizationID := c.Locals("X-Organization-Id").(uuid.UUID)
	payload := p.(*model.CreateReportInput)
	logger.Infof("Request to create a report with details: %#v", payload)

	templateOut, err := rh.Service.CreateReport(ctx, payload, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to create pack on command", err)

		return http.WithError(c, err)
	}

	logger.Infof("Successfully created create report %v", templateOut)

	return http.OK(c, templateOut)
}

// GetDownloadReport is a method to make download of a report.
//
//	@Summary		Download a Report
//	@Description	Make a download of a Report passing the ID
//	@Tags			Reports
//	@Accept			json
//	@Produce		json
//	@Param			Authorization		header	string	true	"The authorization token in the 'Bearer	access_token' format."
//	@Param			X-Organization-Id	header	string	true	"Organization ID"
//	@Success		200					{file}	any
//	@Router			/v1/reports/:id/download [get]
func (rh *ReportHandler) GetDownloadReport(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_template")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating get a Template with ID: %s", id)

	organizationID := c.Locals("X-Organization-Id").(uuid.UUID)

	reportModel, err := rh.Service.GetReportByID(ctx, id, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve template on query", err)

		logger.Errorf("Failed to retrieve Template with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	templateModel, err := rh.Service.GetTemplateByID(ctx, reportModel.TemplateID, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve template on query", err)

		logger.Errorf("Failed to retrieve Template with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	location, errLocation := time.LoadLocation("America/Sao_Paulo")
	if errLocation != nil {
		logger.Errorf("Failed load location time: %s", errLocation.Error())
		return http.WithError(c, errLocation)
	}

	completedAtInBR := reportModel.CompletedAt.In(location)
	objectName := reportModel.ID.String() + "/" + completedAtInBR.Format("20060102_150405") + "." + templateModel.OutputFormat

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
