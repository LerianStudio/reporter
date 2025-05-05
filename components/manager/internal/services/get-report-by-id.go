package services

import (
	"context"
	"errors"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/constant"
	"plugin-template-engine/pkg/mongodb/report"
	"reflect"
)

// GetReportByID recover a report by ID
func (uc *UseCase) GetReportByID(ctx context.Context, id, organizationID uuid.UUID) (*report.Report, error) {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "get_report_by_id")
	defer span.End()

	logger.Infof("Retrieving report for id %v and organizationId %v.", id, organizationID)

	reportModel, err := uc.ReportRepo.FindByID(ctx, reflect.TypeOf(report.Report{}).Name(), id, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get c on repo by id", err)

		logger.Errorf("Error getting FindByID on repo by id: %v", err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", reflect.TypeOf(report.Report{}).Name())
		}

		return nil, err
	}

	return reportModel, nil
}
