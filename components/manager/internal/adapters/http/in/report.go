package in

import (
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"plugin-template-engine/components/manager/internal/services"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/model"
	"plugin-template-engine/pkg/net/http"
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
//	@Param			Authorization		header		string	true	"The authorization token in the 'Bearer	access_token' format."
//	@Param			X-Organization-Id	header		string						true	"Organization ID"
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
