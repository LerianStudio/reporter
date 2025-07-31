package services

import (
	"context"
	"errors"
	"html/template"
	"plugin-smart-templates/v2/pkg"
	"plugin-smart-templates/v2/pkg/constant"
	"reflect"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
)

// DeleteTemplateByID delete a template from the repository
func (uc *UseCase) DeleteTemplateByID(ctx context.Context, id, organizationID uuid.UUID) error {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)
	reqId := commons.NewHeaderIDFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.delete_template_by_id")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
		attribute.String("app.request.organization_id", organizationID.String()),
	)

	logger.Infof("Remove template for id: %s", id)

	if err := uc.TemplateRepo.SoftDelete(ctx, id, organizationID); err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to delete template on repo by id", err)

		logger.Errorf("Error deleting template on repo by id: %v", err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", reflect.TypeOf(template.Template{}).Name())
		}

		return err
	}

	return nil
}
