// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"context"
	"errors"

	"github.com/LerianStudio/reporter/components/manager/internal/services"
	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/model"
	_ "github.com/LerianStudio/reporter/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libConstants "github.com/LerianStudio/lib-commons/v2/commons/constants"
	commonsHttp "github.com/LerianStudio/lib-commons/v2/commons/net/http"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)


// TemplateHandler handles HTTP requests for template operations.
type TemplateHandler struct {
	service *services.UseCase
}

// NewTemplateHandler creates a new TemplateHandler with the given service dependency.
// It returns an error if service is nil.
func NewTemplateHandler(service *services.UseCase) (*TemplateHandler, error) {
	if service == nil {
		return nil, errors.New("service must not be nil for TemplateHandler")
	}

	return &TemplateHandler{service: service}, nil
}

// CreateTemplate is a method that creates a template.
//
//	@Summary		Create a Template for reports
//	@Description	Create a Template for reports with the input payload
//	@Tags			Templates
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			X-Idempotency		header		string	false	"Client-provided idempotency key to prevent duplicate template creation"
//	@Param			templateFile		formData	file	true	"Template file (.tpl)"
//	@Param			outputFormat		formData	string	true	"Output format (e.g., pdf, html)"
//	@Param			description			formData	string	true	"Description of the template"
//	@Success		201					{object}	template.Template
//	@Failure		400					{object}	pkg.HTTPError
//	@Failure		401					{object}	pkg.HTTPError
//	@Failure		403					{object}	pkg.HTTPError
//	@Failure		409					{object}	pkg.HTTPError
//	@Failure		500					{object}	pkg.HTTPError
//	@Router			/v1/templates [post]
func (th *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.template.create")
	defer span.End()

	// Extract X-Idempotency header and inject into context for the service layer
	if idempotencyKey := c.Get(libConstants.IdempotencyKey); idempotencyKey != "" {
		ctx = context.WithValue(ctx, constant.IdempotencyKeyCtx, idempotencyKey)
	}

	replayed := false
	ctx = context.WithValue(ctx, constant.IdempotencyReplayedCtx, &replayed)

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
		libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to get template file from form", err)

		return http.WithError(c, pkg.ValidateBusinessError(constant.ErrInvalidFileUploaded, "", err))
	}

	err = libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", fileHeader)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to set span attributes from struct", err)
	}

	templateFile, errFile := http.GetFileFromHeader(fileHeader)
	if errFile != nil {
		libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to get file from header", errFile)

		return http.WithError(c, errFile)
	}

	// Validate if form fields data is valid
	if errValidate := pkg.ValidateFormDataFields(&outputFormat, &description); errValidate != nil {
		libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to validate form data fields", errValidate)

		logger.Errorf("Error to validate form data fields, Error: %v", errValidate)

		return http.WithError(c, errValidate)
	}

	// Validate if the file content matches the outputFormat
	if errValidateFile := pkg.ValidateFileFormat(outputFormat, templateFile); errValidateFile != nil {
		libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to validate file format", errValidateFile)

		logger.Errorf("Error to validate file format, Error: %v", errValidateFile)

		return http.WithError(c, errValidateFile)
	}

	templateOut, err := th.service.CreateTemplate(ctx, templateFile, outputFormat, description, fileHeader)
	if err != nil {
		if http.IsBusinessError(err) {
			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to create template", err)
		} else {
			libOpentelemetry.HandleSpanError(&span, "Failed to create template", err)
		}

		return http.WithError(c, err)
	}

	logger.Infof("Successfully created template %v", templateOut)

	if replayed {
		c.Set(libConstants.IdempotencyReplayed, "true")
	}

	return commonsHttp.Created(c, templateOut)
}

// UpdateTemplateByID is a method that updates a Template by a given id.
//
//	@Summary		Update a template
//	@Description	Update a template with the input payload
//	@Tags			Templates
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			templateFile	formData	file	true	"Template file (.tpl)"
//	@Param			outputFormat	formData	string	true	"Output format (e.g., pdf, html)"
//	@Param			description		formData	string	true	"Description of the template"
//	@Param			id				path		string	true	"Template ID"
//	@Success		200				{object}	template.Template
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		401				{object}	pkg.HTTPError
//	@Failure		403				{object}	pkg.HTTPError
//	@Failure		404				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/templates/{id} [patch]
func (th *TemplateHandler) UpdateTemplateByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.template.update")
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
	if err != nil && err.Error() != constant.ErrFileAccepted {
		libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to get template file from form", err)

		return http.WithError(c, pkg.ValidateBusinessError(constant.ErrInvalidFileUploaded, "", err))
	}

	err = libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", fileHeader)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to set span attributes from struct", err)
	}

	templateUpdated, errUpdate := th.service.UpdateTemplateByID(ctx, outputFormat, description, id, fileHeader)
	if errUpdate != nil {
		if http.IsBusinessError(errUpdate) {
			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to update template", errUpdate)
		} else {
			libOpentelemetry.HandleSpanError(&span, "Failed to update template", errUpdate)
		}

		logger.Errorf("Failed to update Template with ID: %s, Error: %s", id, errUpdate.Error())

		return http.WithError(c, errUpdate)
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
//	@Security		BearerAuth
//	@Param			id				path		string	true	"Template ID"
//	@Success		200				{object}	template.Template
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		401				{object}	pkg.HTTPError
//	@Failure		403				{object}	pkg.HTTPError
//	@Failure		404				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//
//	@Router			/v1/templates/{id} [get]
func (th *TemplateHandler) GetTemplateByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.template.get")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating get a Template with ID: %s", id)

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
	)

	templateModel, err := th.service.GetTemplateByID(ctx, id)
	if err != nil {
		if http.IsBusinessError(err) {
			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to retrieve template on query", err)
		} else {
			libOpentelemetry.HandleSpanError(&span, "Failed to retrieve template on query", err)
		}

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
//	@Security		BearerAuth
//	@Param			output_format	query		string	false	"Output format filter: XML, HTML, TXT, CSV (also accepts outputFormat)"
//	@Param			description		query		string	false	"Description of template"
//	@Param			limit			query		int		false	"Limit"	default(10)
//	@Param			page			query		int		false	"Page"	default(1)
//	@Success		200				{object}	model.Pagination{items=[]template.Template,page=int,limit=int,total=int}
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		401				{object}	pkg.HTTPError
//	@Failure		403				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/templates [get]
func (th *TemplateHandler) GetAllTemplates(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.template.get_all")
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

	logger.Infof("Initiating retrieval all templates")

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err = libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.query_params", headerParams)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert query params to JSON string", err)
	}

	templates, err := th.service.GetAllTemplates(ctx, *headerParams)
	if err != nil {
		if http.IsBusinessError(err) {
			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to retrieve all Templates on query", err)
		} else {
			libOpentelemetry.HandleSpanError(&span, "Failed to retrieve all Templates on query", err)
		}

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
//	@Security		BearerAuth
//	@Param			id				path	string	true	"Template ID"
//	@Success		204				"No content"
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		401				{object}	pkg.HTTPError
//	@Failure		403				{object}	pkg.HTTPError
//	@Failure		404				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/templates/{id} [delete]
func (th *TemplateHandler) DeleteTemplateByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.template.delete")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating removal of Template with ID: %s", id.String())

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
	)

	if err := th.service.DeleteTemplateByID(ctx, id, false); err != nil {
		if http.IsBusinessError(err) {
			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to remove template on database", err)
		} else {
			libOpentelemetry.HandleSpanError(&span, "Failed to remove template on database", err)
		}

		logger.Errorf("Failed to remove Template with ID: %s, Error: %s", id.String(), err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully removed Template with ID: %s", id.String())

	return commonsHttp.NoContent(c)
}
