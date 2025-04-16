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

type TemplateHandler struct {
	Service *services.UseCase
}

// CreateTemplate is a method that creates a template.
//
//	@Summary		Create a Template for reports
//	@Description	Create a Template for reports with the input payload
//	@Tags			Template
//	@Accept			json
//	@Produce		json
//	@Param			pack				body		model.CreateTemplateInput	true	"Template Input"
//	@Success		201					{object}	template.Template
//	@Router			/v1/template [post]
func (th *TemplateHandler) CreateTemplate(t any, c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.create_template")
	defer span.End()

	c.SetUserContext(ctx)

	logger.Info("Request to create template")

	organizationID := c.Locals("X-Organization-Id").(uuid.UUID)
	payload := t.(*model.CreateTemplateInput)
	logger.Infof("Request to create a pack with details: %#v", payload)

	err := opentelemetry.SetSpanAttributesFromStruct(&span, "payload", payload)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert payload to JSON string", err)

		return http.WithError(c, err)
	}

	templateOut, err := th.Service.CreateTemplate(ctx, payload, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to create pack on command", err)

		return http.WithError(c, err)
	}

	logger.Infof("Successfully created create template %v", templateOut)

	return http.OK(c, templateOut)
}
