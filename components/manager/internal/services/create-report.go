// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"errors"
	"strings"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/model"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb/report"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
)

// CreateReport create a new report
func (uc *UseCase) CreateReport(ctx context.Context, reportInput *model.CreateReportInput) (*report.Report, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.create_report")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", reportInput.TemplateID),
	)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", reportInput)
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
	tOutputFormat, tMappedFields, err := uc.TemplateRepo.FindMappedFieldsAndOutputFormatByID(ctx, templateId)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to find template by ID", err)

		logger.Errorf("Error to find template by id, Error: %v", err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", constant.MongoCollectionTemplate)
		}

		return nil, err
	}

	// Normalize filters to include schema (default to "public" if not specified)
	normalizedFilters := normalizeFiltersWithSchema(reportInput.Filters)

	if normalizedFilters != nil {
		filtersMapped := uc.convertFiltersToMappedFieldsType(normalizedFilters)
		if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, "", logger, filtersMapped); errValidateFields != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to validate filter fields existence on tables", errValidateFields)

			return nil, errValidateFields
		}
	}

	// Build the report model
	reportModel := &report.Report{
		ID:         commons.GenerateUUIDv7(),
		TemplateID: templateId,
		Filters:    normalizedFilters,
		Status:     constant.ProcessingStatus,
	}

	result, err := uc.ReportRepo.Create(ctx, reportModel)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to create report in repository", err)

		logger.Errorf("Error creating report in database: %v", err)

		return nil, err
	}

	// Build report message model
	reportMessage := model.ReportMessage{
		TemplateID:   templateId,
		ReportID:     result.ID,
		Filters:      normalizedFilters,
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

// normalizeFiltersWithSchema normalizes filter table names to include schema.
// If a table name doesn't have a schema prefix (no "." or "__"), it defaults to "public.table".
// This ensures consistent handling of filters in both validation and worker processing.
//
// Examples:
//   - "organization" -> "public.organization"
//   - "public.ledger" -> "public.ledger" (unchanged)
//   - "payment.account" -> "payment.account" (unchanged)
//   - "payment__transfers" -> "payment__transfers" (unchanged, Pongo2 format)
func normalizeFiltersWithSchema(filters map[string]map[string]map[string]model.FilterCondition) map[string]map[string]map[string]model.FilterCondition {
	if filters == nil {
		return nil
	}

	normalized := make(map[string]map[string]map[string]model.FilterCondition)

	for datasource, tables := range filters {
		normalized[datasource] = make(map[string]map[string]model.FilterCondition)

		for tableName, fields := range tables {
			// Check if table name already has a schema
			// Formats with schema: "schema.table" or "schema__table" (Pongo2)
			hasSchema := strings.Contains(tableName, ".") || strings.Contains(tableName, "__")

			normalizedTableName := tableName
			if !hasSchema {
				// Default to public schema
				normalizedTableName = "public." + tableName
			}

			normalized[datasource][normalizedTableName] = fields
		}
	}

	return normalized
}
