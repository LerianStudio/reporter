package services

import (
	"context"
	"errors"

	"github.com/LerianStudio/reporter/v3/pkg"
	"github.com/LerianStudio/reporter/v3/pkg/constant"
	"github.com/LerianStudio/reporter/v3/pkg/mongodb/report"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
)

// GetReportByID recover a report by ID
func (uc *UseCase) GetReportByID(ctx context.Context, id, organizationID uuid.UUID) (*report.Report, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.get_report_by_id")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.report_id", id.String()),
		attribute.String("app.request.organization_id", organizationID.String()),
	)

	logger.Infof("Retrieving report for id %v and organizationId %v.", id, organizationID)

	reportModel, err := uc.ReportRepo.FindByID(ctx, id, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get report on repo by id", err)

		logger.Errorf("Error getting report on repo by id: %v", err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", constant.MongoCollectionReport)
		}

		return nil, err
	}

	return reportModel, nil
}
