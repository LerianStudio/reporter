// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"github.com/LerianStudio/reporter/v4/components/manager/internal/services"
	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/model"
	_ "github.com/LerianStudio/reporter/v4/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/v4/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v2/commons"
	commonsHttp "github.com/LerianStudio/lib-commons/v2/commons/net/http"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
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
//	@Param			Authorization	header		string	false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Param			templateFile	formData	file	true	"Template file (.tpl)"
//	@Param			outputFormat	formData	string	true	"Output format (e.g., pdf, html)"
//	@Param			description		formData	string	true	"Description of the template"
//	@Success		201				{object}	template.Template
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/templates [post]
func (th *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.create_template")
	defer span.End()

	c.SetUserContext(ctx)

	logger.Info("Request to create template")

	outputFormat := c.FormValue("outputFormat")
	description := c.FormValue("description")

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.output_format", outputFormat),
		attribute.String("app.request.description", description),
	)

	fileHeader, err := c.FormFile("template")
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get template file from form", err)

		return http.WithError(c, pkg.ValidateBusinessError(constant.ErrInvalidFileUploaded, "", err))
	}

	err = libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", fileHeader)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to set span attributes from struct", err)
	}

	templateFile, errFile := http.GetFileFromHeader(fileHeader)
	if errFile != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get file from header", errFile)

		return http.WithError(c, errFile)
	}

	// Validate if form fields data is valid
	if errValidate := pkg.ValidateFormDataFields(&outputFormat, &description); errValidate != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to validate form data fields", errValidate)

		logger.Errorf("Error to validate form data fields, Error: %v", errValidate)

		return http.WithError(c, errValidate)
	}

	// Validate if the file content matches the outputFormat
	if errValidateFile := pkg.ValidateFileFormat(outputFormat, templateFile); errValidateFile != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to validate file format", errValidateFile)

		logger.Errorf("Error to validate file format, Error: %v", errValidateFile)

		return http.WithError(c, errValidateFile)
	}

	templateOut, err := th.Service.CreateTemplate(ctx, templateFile, outputFormat, description)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to create template", err)

		return http.WithError(c, err)
	}

	// Get a file in bytes
	fileBytes, err := http.ReadMultipartFile(fileHeader)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to read multipart file", err)

		logger.Errorf("Error to get the file content: %v", err)

		return http.WithError(c, err)
	}

	errPutSeaweedFS := th.Service.TemplateSeaweedFS.Put(ctx, templateOut.FileName, outputFormat, fileBytes)
	if errPutSeaweedFS != nil {
		libOpentelemetry.HandleSpanError(&span, "Error putting template file on SeaweedFS.", errPutSeaweedFS)

		// Compensating transaction: Attempt to roll back the database change to prevent an orphaned record.
		if errDelete := th.Service.DeleteTemplateByID(ctx, templateOut.ID, true); errDelete != nil {
			logger.Errorf("Failed to roll back template creation for ID %s after SeaweedFS failure. Error: %s", templateOut.ID.String(), errDelete.Error())
		}

		logger.Errorf("Error putting template file on SeaweedFS: %s", errPutSeaweedFS.Error())

		return http.WithError(c, errPutSeaweedFS)
	}

	logger.Infof("Successfully created create template %v", templateOut)

	return http.Created(c, templateOut)
}

// UpdateTemplateByID is a method that updates a Template by a given id.
//
//	@Summary		Update a template
//	@Description	Update a template with the input payload
//	@Tags			Templates
//	@Accept			mpfd
//	@Produce		json
//	@Param			Authorization	header		string	false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Param			templateFile	formData	file	true	"Template file (.tpl)"
//	@Param			outputFormat	formData	string	true	"Output format (e.g., pdf, html)"
//	@Param			description		formData	string	true	"Description of the template"
//	@Param			id				path		string	true	"Template ID"
//	@Success		200				{object}	template.Template
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		404				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/templates/{id} [patch]
func (th *TemplateHandler) UpdateTemplateByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.update_template")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating update of Template with ID: %s", id)

	outputFormat := c.FormValue("outputFormat")
	description := c.FormValue("description")

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
		attribute.String("app.request.output_format", outputFormat),
		attribute.String("app.request.description", description),
	)

	fileHeader, err := c.FormFile("template")
	if err != nil && err.Error() != errorFileAccepted {
		libOpentelemetry.HandleSpanError(&span, "Failed to get template file from form", err)

		return http.WithError(c, pkg.ValidateBusinessError(constant.ErrInvalidFileUploaded, "", err))
	}

	err = libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", fileHeader)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to set span attributes from struct", err)
	}

	if errUpdate := th.Service.UpdateTemplateByID(ctx, outputFormat, description, id, fileHeader); errUpdate != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to update template", errUpdate)

		logger.Errorf("Failed to update Template with ID: %s, Error: %s", id, errUpdate.Error())

		return http.WithError(c, errUpdate)
	}

	templateUpdated, err := th.Service.GetTemplateByID(ctx, id)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to retrieve template on query", err)

		logger.Errorf("Failed to retrieve Template with ID: %s, Error: %s", id, err.Error())

		return http.WithError(c, err)
	}

	if fileHeader != nil {
		// Get a file in bytes
		fileBytes, err := http.ReadMultipartFile(fileHeader)
		if err != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to read multipart file", err)

			logger.Errorf("Error to get file content: %v", err)

			return http.WithError(c, err)
		}

		errPutSeaweedFS := th.Service.TemplateSeaweedFS.Put(ctx, templateUpdated.FileName, outputFormat, fileBytes)
		if errPutSeaweedFS != nil {
			libOpentelemetry.HandleSpanError(&span, "Error putting template file on SeaweedFS.", errPutSeaweedFS)

			logger.Errorf("Error putting template file on SeaweedFS: %s", errPutSeaweedFS.Error())

			return http.WithError(c, errPutSeaweedFS)
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
//	@Produce		json
//	@Param			Authorization	header		string	false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Param			id				path		string	true	"Template ID"
//	@Success		200				{object}	template.Template
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		404				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//
//	@Router			/v1/templates/{id} [get]
func (th *TemplateHandler) GetTemplateByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_template")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating get a Template with ID: %s", id)

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
	)

	templateModel, err := th.Service.GetTemplateByID(ctx, id)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to retrieve template on query", err)

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
//	@Param			Authorization	header		string	false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Param			outputFormat	query		string	false	"XML, HTML, TXT and CSV"
//	@Param			description		query		string	false	"Description of template"
//	@Param			limit			query		int		false	"Limit"	default(10)
//	@Param			page			query		int		false	"Page"	default(1)
//	@Success		200				{object}	model.Pagination{items=[]template.Template,page=int,limit=int,total=int}
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/templates [get]
func (th *TemplateHandler) GetAllTemplates(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_all_template")
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

	logger.Infof("Initiating retrieval all templates")

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err = libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.query_params", headerParams)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert query params to JSON string", err)
	}

	templates, err := th.Service.GetAllTemplates(ctx, *headerParams)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to retrieve all Templates on query", err)

		logger.Errorf("Failed to retrieve all Templates, Error: %s", err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully retrieved all Templates")

	pagination.SetItems(templates)
	pagination.SetTotal(len(templates))

	return commonsHttp.OK(c, pagination)
}

// DeleteTemplateByID is a method that removes template information by a given id.
//
//	@Summary		SoftDelete a Template by ID
//	@Description	SoftDelete a Template with the input ID. Returns 204 with no content on success.
//	@Tags			Templates
//	@Produce		json
//	@Param			Authorization	header	string	false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Param			id				path	string	true	"Template ID"
//	@Success		204				"No content"
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		404				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/templates/{id} [delete]
func (th *TemplateHandler) DeleteTemplateByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.delete_template_by_id")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating removal of Template with ID: %s", id.String())

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
	)

	if err := th.Service.DeleteTemplateByID(ctx, id, false); err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to remove template on database", err)

		logger.Errorf("Failed to remove Template with ID: %s, Error: %s", id.String(), err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully removed Template with ID: %s", id.String())

	return commonsHttp.NoContent(c)
}
