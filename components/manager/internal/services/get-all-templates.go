package services

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/constant"
	"plugin-template-engine/pkg/mongodb/template"
	"plugin-template-engine/pkg/net/http"
	"reflect"
)

// GetAllTemplates fetch all Templates from the repository
func (uc *UseCase) GetAllTemplates(ctx context.Context, filters http.QueryHeader, organizationID uuid.UUID) ([]*template.Template, error) {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "get_all_templates")
	defer span.End()

	logger.Infof("Retrieving templates")

	filters.OrganizationID = organizationID

	packs, err := uc.TemplateRepo.FindList(ctx, reflect.TypeOf(template.Template{}).Name(), filters)
	if err != nil || packs == nil {
		opentelemetry.HandleSpanError(&span, "Failed to get packages on repo", err)

		return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", reflect.TypeOf(template.Template{}).Name())
	}

	return packs, nil
}
