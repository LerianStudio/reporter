// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"

	"github.com/LerianStudio/reporter/v4/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/v4/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
)

// GetAllTemplates fetch all Templates from the repository
func (uc *UseCase) GetAllTemplates(ctx context.Context, filters http.QueryHeader) ([]*template.Template, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.get_all_templates")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err := opentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", filters)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert filters to JSON string", err)
	}

	logger.Infof("Retrieving templates")

	templates, errFind := uc.TemplateRepo.FindList(ctx, filters)
	if errFind != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get all templates on repo", errFind)

		return nil, errFind
	}

	return templates, nil
}
