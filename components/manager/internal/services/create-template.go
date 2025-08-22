package services

import (
	"context"
	"fmt"
	"plugin-smart-templates/v2/pkg"
	"plugin-smart-templates/v2/pkg/constant"
	"plugin-smart-templates/v2/pkg/mongodb/template"
	templateUtils "plugin-smart-templates/v2/pkg/template_utils"
	"strings"
	"time"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// CreateTemplate creates a new template with specified parameters and stores it in the repository.
func (uc *UseCase) CreateTemplate(ctx context.Context, templateFile, outFormat, description string, organizationID uuid.UUID) (*template.Template, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.create_template")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_file", templateFile),
		attribute.String("app.request.output_format", outFormat),
		attribute.String("app.request.description", description),
		attribute.String("app.request.organization_id", organizationID.String()),
	)

	logger.Infof("Creating template")

	// Block <script> tags
	if err := templateUtils.ValidateNoScriptTag(templateFile); err != nil {
		libOpentelemetry.HandleSpanError(&span, "Script tag detected in template", err)

		return nil, pkg.ValidateBusinessError(constant.ErrScriptTagDetected, "")
	}

	mappedFields := templateUtils.MappedFieldsOfTemplate(templateFile)
	logger.Infof("Mapped Fields is valid to continue %v", mappedFields)

	if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, organizationID.String(), logger, mappedFields); errValidateFields != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to validate fields existence on tables", errValidateFields)

		logger.Errorf("Error to validate fields existence on tables, Error: %v", errValidateFields)

		return nil, errValidateFields
	}

	// Transform mapped fields for storage (append organizationID to plugin_crm table names)
	transformedMappedFields := TransformMappedFieldsForStorage(mappedFields, organizationID.String())
	logger.Infof("Transformed Mapped Fields for storage %v", transformedMappedFields)

	templateId := commons.GenerateUUIDv7()
	fileName := fmt.Sprintf("%s.tpl", templateId.String())

	templateModel := &template.TemplateMongoDBModel{
		ID:             templateId,
		OutputFormat:   strings.ToLower(outFormat),
		OrganizationID: organizationID,
		FileName:       fileName,
		Description:    description,
		MappedFields:   transformedMappedFields,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      nil,
	}

	resultTemplateModel, err := uc.TemplateRepo.Create(ctx, templateModel)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to create template in repository", err)

		logger.Errorf("Error into creating a template, Error: %v", err)

		return nil, err
	}

	return resultTemplateModel, nil
}
