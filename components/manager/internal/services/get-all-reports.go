// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"

	"github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v3/commons"
	"github.com/LerianStudio/lib-commons/v3/commons/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
)

// GetAllReports fetch all Reports from the repository
func (uc *UseCase) GetAllReports(ctx context.Context, filters http.QueryHeader) ([]*report.Report, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.report.get_all")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err := opentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", filters)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert filters to JSON string", err)
	}

	logger.Infof("Retrieving reports")

	reports, err := uc.ReportRepo.FindList(ctx, filters)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get all reports on repo", err)

		return nil, err
	}

	// Return empty slice if no reports found instead of error (consistent with templates)
	if reports == nil {
		return []*report.Report{}, nil
	}

	return reports, nil
}
