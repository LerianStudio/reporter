package services

import (
	"context"
	"errors"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/mongodb/report"
	"reflect"

	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
)

// GetReportByID recover a report by ID
func (uc *UseCase) GetReportByID(ctx context.Context, id, organizationID uuid.UUID) (*report.Report, error) {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "get_report_by_id")
	defer span.End()

	span.SetAttributes(
		attribute.String("report_id", id.String()),
		attribute.String("organization_id", organizationID.String()),
	)

	logger.Infof("Retrieving report for id %v and organizationId %v.", id, organizationID)

	reportModel, err := uc.ReportRepo.FindByID(ctx, id, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get report on repo by id", err)

		logger.Errorf("Error getting report on repo by id: %v", err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", reflect.TypeOf(report.Report{}).Name())
		}

		return nil, err
	}

	return reportModel, nil
}
