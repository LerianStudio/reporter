// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	templateUtils "github.com/LerianStudio/reporter/pkg/template_utils"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
)

// CreateTemplate creates a new template with specified parameters and stores it in the repository.
func (uc *UseCase) CreateTemplate(ctx context.Context, templateFile, outFormat, description string) (*template.Template, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.create_template")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_file", templateFile),
		attribute.String("app.request.output_format", outFormat),
		attribute.String("app.request.description", description),
	)

	logger.Infof("Creating template")

	// Block <script> tags
	if err := templateUtils.ValidateNoScriptTag(templateFile); err != nil {
		libOpentelemetry.HandleSpanError(&span, "Script tag detected in template", err)

		return nil, pkg.ValidateBusinessError(constant.ErrScriptTagDetected, "")
	}

	mappedFields := templateUtils.MappedFieldsOfTemplate(templateFile)
	logger.Infof("Mapped Fields is valid to continue %v", mappedFields)

	if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, "", logger, mappedFields); errValidateFields != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to validate fields existence on tables", errValidateFields)

		logger.Errorf("Error to validate fields existence on tables, Error: %v", errValidateFields)

		return nil, errValidateFields
	}

	// Transform mapped fields for storage
	// Get MidazOrganizationID from plugin_crm datasource if template uses it
	var midazOrgID string

	if _, hasPluginCRM := mappedFields["plugin_crm"]; hasPluginCRM {
		if ds, exists := uc.ExternalDataSources["plugin_crm"]; exists {
			midazOrgID = ds.MidazOrganizationID
		}
	}

	transformedMappedFields := TransformMappedFieldsForStorage(mappedFields, midazOrgID)
	logger.Infof("Transformed Mapped Fields for storage %v", transformedMappedFields)

	templateId := commons.GenerateUUIDv7()
	fileName := fmt.Sprintf("%s.tpl", templateId.String())

	templateModel := &template.TemplateMongoDBModel{
		ID:           templateId,
		OutputFormat: strings.ToLower(outFormat),
		FileName:     fileName,
		Description:  description,
		MappedFields: transformedMappedFields,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		DeletedAt:    nil,
	}

	resultTemplateModel, err := uc.TemplateRepo.Create(ctx, templateModel)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to create template in repository", err)

		logger.Errorf("Error into creating a template, Error: %v", err)

		return nil, err
	}

	return resultTemplateModel, nil
}
