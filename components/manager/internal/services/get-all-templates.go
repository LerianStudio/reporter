package services

import (
	"context"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/mongodb/template"
	"plugin-smart-templates/pkg/net/http"
	"reflect"

	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// GetAllTemplates fetch all Templates from the repository
func (uc *UseCase) GetAllTemplates(ctx context.Context, filters http.QueryHeader, organizationID uuid.UUID) ([]*template.Template, error) {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "get_all_templates")
	defer span.End()

	span.SetAttributes(
		attribute.String("organization_id", organizationID.String()),
	)

	err := opentelemetry.SetSpanAttributesFromStruct(&span, "filters", filters)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert filters to JSON string", err)
	}

	logger.Infof("Retrieving templates")

	filters.OrganizationID = organizationID

	packs, err := uc.TemplateRepo.FindList(ctx, reflect.TypeOf(template.Template{}).Name(), filters)
	if err != nil || packs == nil {
		opentelemetry.HandleSpanError(&span, "Failed to get all templates on repo", err)

		return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", reflect.TypeOf(template.Template{}).Name())
	}

	return packs, nil
}
