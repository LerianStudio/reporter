package services

import (
	"context"
	"plugin-smart-templates/pkg/mongodb/report"
	"plugin-smart-templates/pkg/net/http"

	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// GetAllReports fetch all Reports from the repository
func (uc *UseCase) GetAllReports(ctx context.Context, filters http.QueryHeader, organizationID uuid.UUID) ([]*report.Report, error) {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)
	reqId := commons.NewHeaderIDFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.get_all_reports")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.organization_id", organizationID.String()),
	)

	err := opentelemetry.SetSpanAttributesFromStructWithObfuscation(&span, "app.request.payload", filters)
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
