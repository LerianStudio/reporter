package in

import (
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"plugin-template-engine/components/manager/internal/services"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/constant"
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
//	@Router			/v1/templates [post]
func (th *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.create_template")
	defer span.End()

	c.SetUserContext(ctx)

	logger.Info("Request to create template")

	organizationID := c.Locals("X-Organization-Id").(uuid.UUID)

	outputFormat := c.FormValue("outputFormat")
	description := c.FormValue("description")

	fileHeader, err := c.FormFile("template")
	if err != nil {
		return http.WithError(c, pkg.ValidateBusinessError(constant.ErrInvalidFileUploaded, "", err))
	}

	templateFile, errFile := http.GetFileFromHeader(fileHeader)
	if errFile != nil {
		return http.WithError(c, errFile)
	}

	// Validate if form fields data is valid
	if errValidate := pkg.ValidateFormDataFields(&outputFormat, &description); errValidate != nil {
		logger.Errorf("Error to validate form data fields, Error: %v", errValidate)
		return http.WithError(c, errValidate)
	}

	// Validate if the file content matches the outputFormat
	if errValidateFile := pkg.ValidateFileFormat(outputFormat, templateFile); errValidateFile != nil {
		logger.Errorf("Error to validate file format, Error: %v", errValidateFile)
		return http.WithError(c, errValidateFile)
	}

	templateOut, err := th.Service.CreateTemplate(ctx, templateFile, outputFormat, description, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to create pack on command", err)

		return http.WithError(c, err)
	}

	// Get file in bytes
	fileBytes, err := http.ReadMultipartFile(fileHeader)
	if err != nil {
		logger.Errorf("Erro ao ler conteúdo do arquivo: %v", err)
		return http.WithError(c, err)
	}

	errPutMinio := th.Service.TemplateMinio.Put(ctx, templateOut.FileName, outputFormat, fileBytes)
	if errPutMinio != nil {
		opentelemetry.HandleSpanError(&span, "Error putting report file.", err)

		logger.Errorf("Error putting report file: %s", err.Error())

		return http.WithError(c, errPutMinio)
	}

	logger.Infof("Successfully created create template %v", templateOut)

	return http.OK(c, templateOut)
}
