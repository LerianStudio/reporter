package services

import (
	"context"
	"errors"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/mongodb/template"
	"reflect"

	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
)

// GetTemplateByID recover a package by ID
func (uc *UseCase) GetTemplateByID(ctx context.Context, id, organizationID uuid.UUID) (*template.Template, error) {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "get_template_by_id")
	defer span.End()

	span.SetAttributes(
		attribute.String("template_id", id.String()),
		attribute.String("organization_id", organizationID.String()),
	)

	logger.Infof("Retrieving template for id %v and organizationId %v.", id, organizationID)

	templateModel, err := uc.TemplateRepo.FindByID(ctx, reflect.TypeOf(template.Template{}).Name(), id, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get template on repo by id", err)

		logger.Errorf("Error getting template on repo by id: %v", err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", reflect.TypeOf(template.Template{}).Name())
		}

		return nil, err
	}

	return templateModel, nil
}
