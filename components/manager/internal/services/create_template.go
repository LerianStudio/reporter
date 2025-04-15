package services

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/google/uuid"
	"plugin-template-engine/components/manager/internal/adapters/mongodb/template"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/model"
	"reflect"
)

// CreateTemplate create a new template
func (uc *UseCase) CreateTemplate(ctx context.Context, in *model.CreateTemplateInput, organizationID uuid.UUID) (*template.Template, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "services.create_example")
	defer span.End()

	logger.Infof("Creating template")

	templateModel := &template.Template{
		ID:   commons.GenerateUUIDv7(),
		Name: in.Name,
		Age:  in.Age,
	}

	resultTemplateModel, err := uc.TemplateRepo.Create(ctx, reflect.TypeOf(template.Template{}).Name(), templateModel, organizationID)
	if err != nil {
		logger.Errorf("Error into creating a template, Error: %v", err)
		return nil, err
	}

	return resultTemplateModel, nil
}
