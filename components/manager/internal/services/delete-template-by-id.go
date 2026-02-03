// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// DeleteTemplateByID delete a template from the repository
func (uc *UseCase) DeleteTemplateByID(ctx context.Context, id, organizationID uuid.UUID, hardDelete bool) error {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.delete_template_by_id")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
		attribute.String("app.request.organization_id", organizationID.String()),
	)

	logger.Infof("Remove template for id: %s", id)

	if err := uc.TemplateRepo.Delete(ctx, id, organizationID, hardDelete); err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to delete template on repo by id", err)
		logger.Errorf("Error deleting template on repo by id: %v", err)

		return err
	}

	return nil
}
