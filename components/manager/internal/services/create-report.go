package services

import (
	"context"
	"errors"
	"plugin-smart-templates/v2/pkg"
	"plugin-smart-templates/v2/pkg/constant"
	"plugin-smart-templates/v2/pkg/model"
	"plugin-smart-templates/v2/pkg/mongodb/report"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
)

// CreateReport create a new report
func (uc *UseCase) CreateReport(ctx context.Context, reportInput *model.CreateReportInput, organizationID uuid.UUID) (*report.Report, error) {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)
	reqId := commons.NewHeaderIDFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.create_report")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.organization_id", organizationID.String()),
		attribute.String("app.request.template_id", reportInput.TemplateID),
	)

	err := libOpentelemetry.SetSpanAttributesFromStructWithObfuscation(&span, "app.request.payload", reportInput)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert report input to JSON string", err)
	}

	logger.Infof("Creating report")

	// Validate templateID is UUID
	templateId, errParseUUID := uuid.Parse(reportInput.TemplateID)
	if errParseUUID != nil {
		errInvalidID := pkg.ValidateBusinessError(constant.ErrInvalidTemplateID, "")

		libOpentelemetry.HandleSpanError(&span, "Invalid template ID format", errParseUUID)

		return nil, errInvalidID
	}

	// Find a template to generate a report
	tOutputFormat, tMappedFields, err := uc.TemplateRepo.FindMappedFieldsAndOutputFormatByID(ctx, templateId, organizationID)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to find template by ID", err)

		logger.Errorf("Error to find template by id, Error: %v", err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", constant.MongoCollectionTemplate)
		}

		return nil, err
	}

	if reportInput.Filters != nil {
		filtersMapped := uc.convertFiltersToMappedFieldsType(reportInput.Filters)
		if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, logger, filtersMapped); errValidateFields != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to validate filter fields existence on tables", errValidateFields)

			return nil, errValidateFields
		}
	}

	// Build the report model
	reportModel := &report.Report{
		ID:         commons.GenerateUUIDv7(),
		TemplateID: templateId,
		Filters:    reportInput.Filters,
		Status:     constant.ProcessingStatus,
	}

	result, err := uc.ReportRepo.Create(ctx, reportModel, organizationID)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to create report in repository", err)

		logger.Errorf("Error creating report in database: %v", err)

		return nil, err
	}

	// Build report message model
	reportMessage := model.ReportMessage{
		TemplateID:   templateId,
		ReportID:     result.ID,
		Filters:      reportInput.Filters,
		OutputFormat: *tOutputFormat,
		MappedFields: tMappedFields,
	}

	logger.Infof("Sending report to reports queue...")
	uc.SendReportQueueReports(ctx, reportMessage)

	return result, nil
}

// convertFiltersToMappedFieldsType transforms a deeply nested filter map into a mapped fields structure with limited keys per level.
func (uc *UseCase) convertFiltersToMappedFieldsType(filters map[string]map[string]map[string]model.FilterCondition) map[string]map[string][]string {
	output := make(map[string]map[string][]string)

	for topKey, nested := range filters {
		output[topKey] = make(map[string][]string)

		for midKey, inner := range nested {
			var keys []string

			count := 0

			for innerKey := range inner {
				keys = append(keys, innerKey)

				count++
				if count == 3 {
					break
				}
			}

			output[topKey][midKey] = keys
		}
	}

	return output
}
