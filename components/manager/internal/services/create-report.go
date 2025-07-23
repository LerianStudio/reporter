package services

import (
	"context"
	"errors"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/model"
	"plugin-smart-templates/pkg/mongodb/report"
	"plugin-smart-templates/pkg/mongodb/template"
	"reflect"

	"github.com/LerianStudio/lib-commons/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
)

// CreateReport create a new report
func (uc *UseCase) CreateReport(ctx context.Context, reportInput *model.CreateReportInput, organizationID uuid.UUID) (*report.Report, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "services.create_report")
	defer span.End()

	span.SetAttributes(
		attribute.String("organization_id", organizationID.String()),
		attribute.String("template_id", reportInput.TemplateID),
	)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "report_input", reportInput)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert report input to JSON string", err)
	}

	logger.Infof("Creating report")

	// Validate the ledgerID list if all values are uuid
	ledgerIDConverted := make([]uuid.UUID, 0, len(reportInput.LedgerID))

	for _, ledgerId := range reportInput.LedgerID {
		ledgerConverted, errParseLedgerID := uuid.Parse(ledgerId)
		if errParseLedgerID != nil {
			return nil, pkg.ValidateBusinessError(constant.ErrInvalidLedgerIDList, "", ledgerId)
		}

		ledgerIDConverted = append(ledgerIDConverted, ledgerConverted)
	}

	// Validate templateID is UUID
	templateId, errParseUUID := uuid.Parse(reportInput.TemplateID)
	if errParseUUID != nil {
		return nil, pkg.ValidateBusinessError(constant.ErrInvalidTemplateID, "")
	}

	// Find a template to generate a report
	tOutputFormat, tMappedFields, err := uc.TemplateRepo.FindMappedFieldsAndOutputFormatByID(ctx, reflect.TypeOf(template.Template{}).Name(), templateId, organizationID)
	if err != nil {
		logger.Errorf("Error to find template by id, Error: %v", err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", reflect.TypeOf(template.Template{}).Name())
		}

		return nil, err
	}

	if reportInput.Filters != nil {
		filtersMapped := uc.convertFiltersToMappedFieldsType(reportInput.Filters)
		if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, logger, filtersMapped); errValidateFields != nil {
			return nil, errValidateFields
		}
	}

	// Build the report model
	reportModel := &report.Report{
		ID:         commons.GenerateUUIDv7(),
		TemplateID: templateId,
		LedgerID:   ledgerIDConverted,
		Filters:    reportInput.Filters,
		Status:     constant.ProcessingStatus,
	}

	result, err := uc.ReportRepo.Create(ctx, reflect.TypeOf(report.Report{}).Name(), reportModel, organizationID)
	if err != nil {
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
func (uc *UseCase) convertFiltersToMappedFieldsType(filters map[string]map[string]map[string][]string) map[string]map[string][]string {
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
