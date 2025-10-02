package services

import (
	"context"
	"plugin-smart-templates/v3/pkg/mongodb/template"
	"plugin-smart-templates/v3/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// GetAllTemplates fetch all Templates from the repository
func (uc *UseCase) GetAllTemplates(ctx context.Context, filters http.QueryHeader, organizationID uuid.UUID) ([]*template.Template, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.get_all_templates")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.organization_id", organizationID.String()),
	)

	err := opentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", filters)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert filters to JSON string", err)
	}

	logger.Infof("Retrieving templates")

	filters.OrganizationID = organizationID

	templates, errFind := uc.TemplateRepo.FindList(ctx, filters)
	if errFind != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get all templates on repo", errFind)

		return nil, errFind
	}

	return templates, nil
}
