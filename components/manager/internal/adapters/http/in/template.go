package in

import (
	"github.com/LerianStudio/lib-commons/commons"
	commonsHttp "github.com/LerianStudio/lib-commons/commons/net/http"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"plugin-template-engine/components/manager/internal/services"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/constant"
	"plugin-template-engine/pkg/model"
	_ "plugin-template-engine/pkg/mongodb/template"
	"plugin-template-engine/pkg/net/http"
)

const errorFileAccepted = "there is no uploaded file associated with the given key"

type TemplateHandler struct {
	Service *services.UseCase
}

// CreateTemplate is a method that creates a template.
//
//	@Summary		Create a Template for reports
//	@Description	Create a Template for reports with the input payload
//	@Tags			Templates
//	@Accept			mpfd
//	@Produce		json
//	@Param			Authorization		header		string	true	"The authorization token in the 'Bearer	access_token' format."
//	@Param			templateFile		formData	file	true	"Template file (.tpl)"
//	@Param			outputFormat	 	formData	string	true	"Output format (e.g., pdf, html)"
//	@Param			description			formData	string	true	"Description of the template"
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
		opentelemetry.HandleSpanError(&span, "Error putting template file.", err)

		logger.Errorf("Error putting template file: %s", err.Error())

		return http.WithError(c, errPutMinio)
	}

	logger.Infof("Successfully created create template %v", templateOut)

	return http.OK(c, templateOut)
}

// UpdateTemplateByID is a method that update a Template by a given id.
//
//	@Summary		Update a template
//	@Description	Update a template with the input payload
//	@Tags			Templates
//	@Accept			mpfd
//	@Produce		json
//	@Param			Authorization		header		string	true	"The authorization token in the 'Bearer	access_token' format."
//	@Param			templateFile		formData	file	true	"Template file (.tpl)"
//	@Param			outputFormat	 	formData	string	true	"Output format (e.g., pdf, html)"
//	@Param			description			formData	string	true	"Description of the template"
//	@Success		200					{object}	template.Template
//	@Router			/v1/templates/{id} [patch]
func (th *TemplateHandler) UpdateTemplateByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.update_template")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating update of Package with ID: %s", id)

	organizationID := c.Locals("X-Organization-Id").(uuid.UUID)

	outputFormat := c.FormValue("outputFormat")
	description := c.FormValue("description")

	fileHeader, err := c.FormFile("template")
	if err != nil && err.Error() != errorFileAccepted {
		return http.WithError(c, pkg.ValidateBusinessError(constant.ErrInvalidFileUploaded, "", err))
	}

	if errUpdate := th.Service.UpdateTemplateByID(ctx, outputFormat, description, organizationID, id, fileHeader); errUpdate != nil {
		opentelemetry.HandleSpanError(&span, "Failed to update package", errUpdate)

		logger.Errorf("Failed to update Package with ID: %s, Error: %s", id, errUpdate.Error())

		return http.WithError(c, errUpdate)
	}

	templateUpdated, err := th.Service.GetTemplateByID(ctx, id, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve template on query", err)

		logger.Errorf("Failed to retrieve Template with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	if fileHeader != nil {
		// Get file in bytes
		fileBytes, err := http.ReadMultipartFile(fileHeader)
		if err != nil {
			logger.Errorf("Erro ao ler conteúdo do arquivo: %v", err)
			return http.WithError(c, err)
		}

		errPutMinio := th.Service.TemplateMinio.Put(ctx, templateUpdated.FileName, outputFormat, fileBytes)
		if errPutMinio != nil {
			opentelemetry.HandleSpanError(&span, "Error putting template file.", err)

			logger.Errorf("Error putting template file: %s", err.Error())

			return http.WithError(c, errPutMinio)
		}
	}

	logger.Infof("Successfully updated Template with ID: %s", id)

	return commonsHttp.OK(c, templateUpdated)
}

// GetTemplateByID is a method that retrieves a Template information by a given id.
//
//	@Summary		Get template
//	@Description	Get a template by id
//	@Tags			Templates
//	@Accept			mpfd
//	@Produce		json
//	@Param			Authorization		header		string	true	"The authorization token in the 'Bearer	access_token' format."
//	@Param			templateFile		formData	file	true	"Template file (.tpl)"
//	@Param			outputFormat	 	formData	string	true	"Output format (e.g., pdf, html)"
//	@Param			description			formData	string	true	"Description of the template"
//	@Success		200					{object}	template.Template
//	@Router			/v1/templates/{id} [get]
func (th *TemplateHandler) GetTemplateByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_template")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating get a Template with ID: %s", id)

	organizationID := c.Locals("X-Organization-Id").(uuid.UUID)

	templateModel, err := th.Service.GetTemplateByID(ctx, id, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve template on query", err)

		logger.Errorf("Failed to retrieve Template with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully retrieve Template with ID: %s", id)

	return commonsHttp.OK(c, templateModel)
}

// GetAllTemplates is a method that recovery all Templates information.
//
//	@Summary		Get all templates
//	@Description	List all the templates
//	@Tags			Templates
//	@Produce		json
//	@Param			Authorization		header		string	true	"The authorization token in the 'Bearer	access_token' format."
//	@Param			X-Organization-Id	header		string	true	"Organization ID"
//	@Param			limit				query		int		false	"Limit"	default(10)
//	@Param			page				query		int		false	"Page"	default(1)
//	@Success		200					{object}	model.Pagination{items=[]template.Template,page=int,limit=int,total=int}
//	@Router			/v1/templates [get]
func (th *TemplateHandler) GetAllTemplates(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_all_template")
	defer span.End()

	headerParams, err := http.ValidateParameters(c.Queries())
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to validate query parameters", err)

		logger.Errorf("Failed to validate query parameters, Error: %s", err.Error())

		return http.WithError(c, err)
	}

	pagination := model.Pagination{
		Limit: headerParams.Limit,
		Page:  headerParams.Page,
	}

	logger.Infof("Initiating retrieval all templates")

	organizationID := c.Locals("X-Organization-Id").(uuid.UUID)

	templates, err := th.Service.GetAllTemplates(ctx, *headerParams, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve all Templates on query", err)

		logger.Errorf("Failed to retrieve all Templates, Error: %s", err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully retrieved all Templates")

	pagination.SetItems(templates)
	pagination.SetTotal(len(templates))

	return commonsHttp.OK(c, pagination)
}

// DeleteTemplateByID is a method that removes a template information by a given id.
//
//	@Summary		SoftDelete a Template by ID
//	@Description	SoftDelete a Template with the input ID
//	@Tags			Templates
//	@Param			Authorization		header	string	true	"The authorization token in the 'Bearer	access_token' format."
//	@Param			X-Organization-Id	header	string	true	"Organization ID"
//	@Param			id					path	string	true	"Template ID"
//	@Success		204
//	@Router			/v1/templates/{id} [delete]
func (th *TemplateHandler) DeleteTemplateByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.delete_template_by_id")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating removal of Template with ID: %s", id.String())

	organizationID := c.Locals("X-Organization-Id").(uuid.UUID)
	if err := th.Service.DeleteTemplateByID(ctx, id, organizationID); err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to remove template on database", err)

		logger.Errorf("Failed to remove Template with ID: %s, Error: %s", id.String(), err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully removed Template with ID: %s", id.String())

	return commonsHttp.NoContent(c)
}
