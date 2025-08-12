package services

import (
	"context"
	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"plugin-smart-templates/v2/pkg/mongodb/template"
	"plugin-smart-templates/v2/pkg/net/http"
)

// GetAllTemplates fetch all Templates from the repository
func (uc *UseCase) GetAllTemplates(ctx context.Context, filters http.QueryHeader, organizationID uuid.UUID) ([]*template.Template, error) {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)
	reqId := commons.NewHeaderIDFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.get_all_templates")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.organization_id", organizationID.String()),
	)

	err := opentelemetry.SetSpanAttributesFromStructWithObfuscation(&span, "app.request.payload", filters)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert filters to JSON string", err)
	}

	logger.Infof("Retrieving templates")

	filters.OrganizationID = organizationID

	templates, errFind := uc.TemplateRepo.FindList(ctx, filters)
	if err != nil || templates == nil {
		opentelemetry.HandleSpanError(&span, "Failed to get all templates on repo", errFind)

		return nil, errFind
	}

	return templates, nil
}
