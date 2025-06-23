package services

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"plugin-smart-templates/pkg/mongodb/report"
	"plugin-smart-templates/pkg/net/http"
	"reflect"
)

// GetAllReports fetch all Reports from the repository
func (uc *UseCase) GetAllReports(ctx context.Context, filters http.QueryHeader, organizationID uuid.UUID) ([]*report.Report, error) {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "get_all_reports")
	defer span.End()

	logger.Infof("Retrieving reports")

	filters.OrganizationID = organizationID

	reports, err := uc.ReportRepo.FindList(ctx, reflect.TypeOf(report.Report{}).Name(), filters)
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