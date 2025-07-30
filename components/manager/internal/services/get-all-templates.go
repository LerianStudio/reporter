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
	reqId := commons.NewHeaderIDFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.get_all_templates")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.organization_id", organizationID.String()),
	)

	err := opentelemetry.SetSpanAttributesFromStructWithObfuscation(&span, "app.request.query_params", filters)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert filters to JSON string", err)
	}

	logger.Infof("Retrieving templates")

	filters.OrganizationID = organizationID

	packs, err := uc.TemplateRepo.FindList(ctx, filters)
	if err != nil || packs == nil {
		opentelemetry.HandleSpanError(&span, "Failed to get all templates on repo", err)

		return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", reflect.TypeOf(template.Template{}).Name())
	}

	return packs, nil
}
