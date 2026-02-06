// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/pkg/net/http"
	templateUtils "github.com/LerianStudio/reporter/pkg/template_utils"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
)

// CreateTemplate creates a new template with specified parameters, stores it in the repository,
// uploads the file to object storage, and performs a compensating transaction on storage failure.
func (uc *UseCase) CreateTemplate(ctx context.Context, templateFile, outFormat, description string, fileHeader *multipart.FileHeader) (*template.Template, error) {
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
		return nil, pkg.ValidateBusinessError(constant.ErrScriptTagDetected, "")
	}

	mappedFields := templateUtils.MappedFieldsOfTemplate(templateFile)
	logger.Infof("Mapped Fields is valid to continue %v", mappedFields)

	if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, mappedFields); errValidateFields != nil {
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

	// Read file bytes and upload to object storage
	fileBytes, err := http.ReadMultipartFile(fileHeader)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to read multipart file", err)

		logger.Errorf("Error to get the file content: %v", err)

		return nil, err
	}

	errPutStorage := uc.TemplateSeaweedFS.Put(ctx, resultTemplateModel.FileName, outFormat, fileBytes)
	if errPutStorage != nil {
		libOpentelemetry.HandleSpanError(&span, "Error putting template file on storage", errPutStorage)

		// Compensating transaction: Attempt to roll back the database change to prevent an orphaned record.
		if errDelete := uc.DeleteTemplateByID(ctx, resultTemplateModel.ID, true); errDelete != nil {
			logger.Errorf("Failed to roll back template creation for ID %s after storage failure. Error: %s", resultTemplateModel.ID.String(), errDelete.Error())
		}

		logger.Errorf("Error putting template file on storage: %s", errPutStorage.Error())

		return nil, errPutStorage
	}

	return resultTemplateModel, nil
}
