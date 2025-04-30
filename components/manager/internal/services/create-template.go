package services

import (
	"context"
	"fmt"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/google/uuid"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/mongodb/template"
	templateUtils "plugin-template-engine/pkg/template_utils"
	"reflect"
	"time"
)

// CreateTemplate create a new template
func (uc *UseCase) CreateTemplate(ctx context.Context, templateFile, outFormat, description string, organizationID uuid.UUID) (*template.Template, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "services.create_template")
	defer span.End()

	logger.Infof("Creating template")

	mappedFields := templateUtils.MappedFieldsOfTemplate(templateFile)
	logger.Infof("Mapped Fields is valid to continue %v", mappedFields)

	templateId := commons.GenerateUUIDv7()
	fileName := fmt.Sprintf("%s.tpl", templateId.String())

	templateModel := &template.TemplateMongoDBModel{
		ID:             templateId,
		OutputFormat:   outFormat,
		OrganizationID: organizationID,
		FileName:       fileName,
		Description:    description,
		MappedFields:   mappedFields,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      nil,
	}

	resultTemplateModel, err := uc.TemplateRepo.Create(ctx, reflect.TypeOf(template.Template{}).Name(), templateModel)
	if err != nil {
		logger.Errorf("Error into creating a template, Error: %v", err)
		return nil, err
	}

	return resultTemplateModel, nil
}
