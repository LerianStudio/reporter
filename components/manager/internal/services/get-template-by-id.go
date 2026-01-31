// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"errors"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb/template"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
)

// GetTemplateByID recover a package by ID
func (uc *UseCase) GetTemplateByID(ctx context.Context, id, organizationID uuid.UUID) (*template.Template, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.get_template_by_id")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
		attribute.String("app.request.organization_id", organizationID.String()),
	)

	logger.Infof("Retrieving template for id %v and organizationId %v.", id, organizationID)

	templateModel, err := uc.TemplateRepo.FindByID(ctx, id, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get template on repo by id", err)

		logger.Errorf("Error getting template on repo by id: %v", err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", constant.MongoCollectionTemplate)
		}

		return nil, err
	}

	return templateModel, nil
}
