// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"mime/multipart"
	"strings"
	"time"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/pkg/net/http"
	templateUtils "github.com/LerianStudio/reporter/pkg/template_utils"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	// otel/attribute is used for span attribute types (no lib-commons wrapper available)
	"go.opentelemetry.io/otel/attribute"
	// otel/trace is used for trace.Span parameter type in validateOutputFormatAndFile
	"go.opentelemetry.io/otel/trace"
)

// UpdateTemplateByID updates an existing template, optionally uploading a new file to storage,
// and returns the updated template.
func (uc *UseCase) UpdateTemplateByID(ctx context.Context, outputFormat, description string, id uuid.UUID, fileHeader *multipart.FileHeader) (*template.Template, error) {
	var (
		templateFile string
		mappedFields map[string]map[string][]string
	)

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.template.update")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
	)

	logger.Infof("Updating template")

	if fileHeader != nil {
		var err error

		templateFile, mappedFields, err = uc.processTemplateFile(fileHeader, logger)
		if err != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to process template file", err)

			return nil, err
		}

		if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, mappedFields); errValidateFields != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to validate fields existence on tables", errValidateFields)

			logger.Errorf("Error to validate fields existence on tables, Error: %v", errValidateFields)

			return nil, errValidateFields
		}
	}

	// Validate output format and file format compatibility
	if err := uc.validateOutputFormatAndFile(ctx, &span, id, fileHeader, outputFormat, templateFile, logger); err != nil {
		return nil, err
	}

	// If a new file was provided, upload it to object storage FIRST (before DB update)
	if fileHeader != nil {
		// Fetch the current template BEFORE updating to get the FileName for storage upload
		currentTemplate, err := uc.TemplateRepo.FindByID(ctx, id)
		if err != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to retrieve current template", err)

			logger.Errorf("Failed to retrieve Template with ID: %s, Error: %s", id, err.Error())

			return nil, err
		}

		fileBytes, errRead := http.ReadMultipartFile(fileHeader)
		if errRead != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to read multipart file", errRead)

			logger.Errorf("Error to get file content: %v", errRead)

			return nil, errRead
		}

		// Determine the contentType for storage: use the new outputFormat if provided, otherwise use existing
		storageContentType := outputFormat
		if commons.IsNilOrEmpty(&storageContentType) {
			storageContentType = currentTemplate.OutputFormat
		}

		errPutStorage := uc.TemplateSeaweedFS.Put(ctx, currentTemplate.FileName, storageContentType, fileBytes)
		if errPutStorage != nil {
			libOpentelemetry.HandleSpanError(&span, "Error putting template file on storage", errPutStorage)

			logger.Errorf("Error putting template file on storage: %s", errPutStorage.Error())

			return nil, errPutStorage
		}
	}

	// Now update the database
	setFields := uc.buildSetFields(description, outputFormat, mappedFields)
	updateFields := bson.M{}

	if len(setFields) > 0 {
		updateFields["$set"] = setFields
	}

	if errUpdate := uc.TemplateRepo.Update(ctx, id, &updateFields); errUpdate != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to update template in repository", errUpdate)

		logger.Errorf("Error into updating a template, Error: %v", errUpdate)

		// Note: storage has already been updated at this point (idempotent operation)
		logger.Warnf("Storage was updated but DB update failed - storage operations are idempotent")

		return nil, errUpdate
	}

	// Fetch the updated template to return
	templateUpdated, err := uc.GetTemplateByID(ctx, id)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to retrieve updated template", err)

		logger.Errorf("Failed to retrieve Template with ID: %s, Error: %s", id, err.Error())

		return nil, err
	}

	return templateUpdated, nil
}

// processTemplateFile handles file extraction, script tag validation, and mapped fields extraction.
func (uc *UseCase) processTemplateFile(fileHeader *multipart.FileHeader, logger log.Logger) (string, map[string]map[string][]string, error) {
	templateFile, errFile := http.GetFileFromHeader(fileHeader)
	if errFile != nil {
		return "", nil, errFile
	}

	if err := templateUtils.ValidateNoScriptTag(templateFile); err != nil {
		return "", nil, pkg.ValidateBusinessError(constant.ErrScriptTagDetected, "")
	}

	mappedFields := templateUtils.MappedFieldsOfTemplate(templateFile)
	logger.Infof("Mapped Fields is valid to continue %v", mappedFields)

	return templateFile, mappedFields, nil
}

// buildSetFields builds the setFields map for the update operation.
func (uc *UseCase) buildSetFields(description, outputFormat string, mappedFields map[string]map[string][]string) bson.M {
	setFields := bson.M{}
	if !commons.IsNilOrEmpty(&description) {
		setFields["description"] = description
	}

	if !commons.IsNilOrEmpty(&outputFormat) {
		setFields["output_format"] = strings.ToLower(outputFormat)
	}

	if mappedFields != nil {
		setFields["mapped_fields"] = mappedFields
	}

	setFields["updated_at"] = time.Now()

	return setFields
}

// validateOutputFormatAndFile validates output format and file format compatibility.
func (uc *UseCase) validateOutputFormatAndFile(ctx context.Context, span *trace.Span, id uuid.UUID, fileHeader *multipart.FileHeader, outputFormat, templateFile string, logger log.Logger) error {
	// If file is provided without explicit outputFormat, validate against existing template's outputFormat
	if fileHeader != nil && commons.IsNilOrEmpty(&outputFormat) {
		outputFormatExistentTemplate, err := uc.TemplateRepo.FindOutputFormatByID(ctx, id)
		if err != nil {
			libOpentelemetry.HandleSpanError(span, "Failed to get outputFormat of template by ID", err)
			logger.Errorf("Error to get outputFormat of template by ID, Error: %v", err)

			return err
		}

		if errFileFormat := pkg.ValidateFileFormat(*outputFormatExistentTemplate, templateFile); errFileFormat != nil {
			logger.Errorf("Error to validate file format, Error: %v", errFileFormat)
			return errFileFormat
		}
	}

	// If outputFormat is explicitly provided, validate it
	if !commons.IsNilOrEmpty(&outputFormat) {
		if !pkg.IsOutputFormatValuesValid(&outputFormat) {
			errInvalidFormat := pkg.ValidateBusinessError(constant.ErrInvalidOutputFormat, "")

			logger.Errorf("Error invalid outputFormat value %v", outputFormat)

			return errInvalidFormat
		}

		if fileHeader == nil {
			errMissingFile := pkg.ValidateBusinessError(constant.ErrOutputFormatWithoutTemplateFile, "")

			logger.Error("Can not update outputFormat without passing the file template")

			return errMissingFile
		}

		if errFileFormat := pkg.ValidateFileFormat(outputFormat, templateFile); errFileFormat != nil {
			logger.Errorf("Error to validate file format, Error: %v", errFileFormat)
			return errFileFormat
		}
	}

	return nil
}
