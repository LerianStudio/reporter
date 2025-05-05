package services

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/google/uuid"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/constant"
	"plugin-template-engine/pkg/model"
	"plugin-template-engine/pkg/mongodb/report"
	"plugin-template-engine/pkg/mongodb/template"
	"reflect"
)

// CreateReport create a new report
func (uc *UseCase) CreateReport(ctx context.Context, reportInput *model.CreateReportInput, organizationID uuid.UUID) (*report.Report, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "services.create_report")
	defer span.End()

	logger.Infof("Creating report")

	// Validate ledgerID list if all values is uuid
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

	// Find template to generate report
	tOutputFormat, tMappedFields, err := uc.TemplateRepo.FindMappedFieldsAndOutputFormatByID(ctx, reflect.TypeOf(template.Template{}).Name(), templateId, organizationID)
	if err != nil {
		logger.Errorf("Error to find template by id, Error: %v", err)
		return nil, err
	}

	// Build the report model
	reportModel := &report.Report{
		ID:         commons.GenerateUUIDv7(),
		TemplateID: templateId,
		LedgerID:   ledgerIDConverted,
		Filters:    reportInput.Filters,
		Status:     "processing",
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
