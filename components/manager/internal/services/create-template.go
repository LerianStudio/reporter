package services

import (
	"context"
	"fmt"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/mongodb/template"
	templateUtils "plugin-smart-templates/pkg/template_utils"
	"strings"
	"time"

	"github.com/LerianStudio/lib-commons/commons"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// CreateTemplate creates a new template with specified parameters and stores it in the repository.
func (uc *UseCase) CreateTemplate(ctx context.Context, templateFile, outFormat, description string, organizationID uuid.UUID) (*template.Template, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "services.create_template")
	defer span.End()

	span.SetAttributes(
		attribute.String("template_file", templateFile),
		attribute.String("output_format", outFormat),
		attribute.String("description", description),
		attribute.String("organization_id", organizationID.String()),
	)

	logger.Infof("Creating template")

	// Block <script> tags
	if err := templateUtils.ValidateNoScriptTag(templateFile); err != nil {
		return nil, pkg.ValidateBusinessError(constant.ErrScriptTagDetected, "")
	}

	mappedFields := templateUtils.MappedFieldsOfTemplate(templateFile)
	logger.Infof("Mapped Fields is valid to continue %v", mappedFields)

	if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, logger, mappedFields); errValidateFields != nil {
		logger.Errorf("Error to validate fields existence on tables, Error: %v", errValidateFields)
		return nil, errValidateFields
	}

	templateId := commons.GenerateUUIDv7()
	fileName := fmt.Sprintf("%s.tpl", templateId.String())

	templateModel := &template.TemplateMongoDBModel{
		ID:             templateId,
		OutputFormat:   strings.ToLower(outFormat),
		OrganizationID: organizationID,
		FileName:       fileName,
		Description:    description,
		MappedFields:   mappedFields,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      nil,
	}

	resultTemplateModel, err := uc.TemplateRepo.Create(ctx, templateModel)
	if err != nil {
		logger.Errorf("Error into creating a template, Error: %v", err)
		return nil, err
	}

	return resultTemplateModel, nil
}
