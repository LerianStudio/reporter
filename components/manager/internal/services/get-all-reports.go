package services

import (
	"context"
	"github.com/LerianStudio/reporter/v3/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/v3/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// GetAllReports fetch all Reports from the repository
func (uc *UseCase) GetAllReports(ctx context.Context, filters http.QueryHeader, organizationID uuid.UUID) ([]*report.Report, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.get_all_reports")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.organization_id", organizationID.String()),
	)

	err := opentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", filters)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert filters to JSON string", err)
	}

	logger.Infof("Retrieving reports")

	filters.OrganizationID = organizationID

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
