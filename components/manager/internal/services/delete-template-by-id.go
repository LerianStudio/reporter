// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"

	pkgHTTP "github.com/LerianStudio/reporter/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v3/commons"
	"github.com/LerianStudio/lib-commons/v3/commons/opentelemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// DeleteTemplateByID delete a template from the repository
func (uc *UseCase) DeleteTemplateByID(ctx context.Context, id uuid.UUID, hardDelete bool) error {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.template.delete")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
	)

	logger.Infof("Remove template for id: %s", id)

	if err := uc.TemplateRepo.Delete(ctx, id, hardDelete); err != nil {
		if pkgHTTP.IsBusinessError(err) {
			opentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to delete template on repo by id", err)
		} else {
			opentelemetry.HandleSpanError(&span, "Failed to delete template on repo by id", err)
		}

		logger.Errorf("Error deleting template on repo by id: %v", err)

		return err
	}

	return nil
}
