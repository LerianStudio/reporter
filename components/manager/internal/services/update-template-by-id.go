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
	pkgHTTP "github.com/LerianStudio/reporter/pkg/net/http"
	templateUtils "github.com/LerianStudio/reporter/pkg/templateutils"

	"github.com/LerianStudio/lib-commons/v3/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v3/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	// otel/attribute is used for span attribute types (no lib-commons wrapper available)
	"go.opentelemetry.io/otel/attribute"
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

		templateFile, mappedFields, err = uc.processTemplateFile(ctx, fileHeader)
		if err != nil {
			if pkgHTTP.IsBusinessError(err) {
				libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to process template file", err)
			} else {
				libOpentelemetry.HandleSpanError(&span, "Failed to process template file", err)
			}

			return nil, err
		}

		if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, mappedFields); errValidateFields != nil {
			if pkgHTTP.IsBusinessError(errValidateFields) {
				libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to validate fields existence on tables", errValidateFields)
			} else {
				libOpentelemetry.HandleSpanError(&span, "Failed to validate fields existence on tables", errValidateFields)
			}

			logger.Errorf("Error to validate fields existence on tables, Error: %v", errValidateFields)

			return nil, errValidateFields
		}
	}

	// Validate output format and file format compatibility
	if err := uc.validateOutputFormatAndFile(ctx, id, fileHeader, outputFormat, templateFile); err != nil {
		return nil, err
	}

	// If a new file was provided, upload it to object storage FIRST (before DB update)
	if fileHeader != nil {
		if err := uc.uploadTemplateFileToStorage(ctx, id, outputFormat, fileHeader, &span); err != nil {
			return nil, err
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
		if pkgHTTP.IsBusinessError(err) {
			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to retrieve updated template", err)
		} else {
			libOpentelemetry.HandleSpanError(&span, "Failed to retrieve updated template", err)
		}

		logger.Errorf("Failed to retrieve Template with ID: %s, Error: %s", id, err.Error())

		return nil, err
	}

	return templateUpdated, nil
}

// uploadTemplateFileToStorage fetches the current template, reads the file bytes, and uploads
// the new file content to object storage.
func (uc *UseCase) uploadTemplateFileToStorage(ctx context.Context, id uuid.UUID, outputFormat string, fileHeader *multipart.FileHeader, span *trace.Span) error {
	logger, _, _, _ := commons.NewTrackingFromContext(ctx) //nolint:dogsled // only logger needed from tracking context

	// Fetch the current template BEFORE updating to get the FileName for storage upload
	currentTemplate, err := uc.TemplateRepo.FindByID(ctx, id)
	if err != nil {
		if pkgHTTP.IsBusinessError(err) {
			libOpentelemetry.HandleSpanBusinessErrorEvent(span, "Failed to retrieve current template", err)
		} else {
			libOpentelemetry.HandleSpanError(span, "Failed to retrieve current template", err)
		}

		logger.Errorf("Failed to retrieve Template with ID: %s, Error: %s", id, err.Error())

		return err
	}

	fileBytes, errRead := pkgHTTP.ReadMultipartFile(fileHeader)
	if errRead != nil {
		libOpentelemetry.HandleSpanError(span, "Failed to read multipart file", errRead)

		logger.Errorf("Error to get file content: %v", errRead)

		return errRead
	}

	// Determine the contentType for storage: use the new outputFormat if provided, otherwise use existing
	storageContentType := outputFormat
	if commons.IsNilOrEmpty(&storageContentType) {
		storageContentType = currentTemplate.OutputFormat
	}

	errPutStorage := uc.TemplateSeaweedFS.Put(ctx, currentTemplate.FileName, storageContentType, fileBytes)
	if errPutStorage != nil {
		libOpentelemetry.HandleSpanError(span, "Error putting template file on storage", errPutStorage)

		logger.Errorf("Error putting template file on storage: %s", errPutStorage.Error())

		return errPutStorage
	}

	return nil
}

// processTemplateFile handles file extraction, script tag validation, and mapped fields extraction.
func (uc *UseCase) processTemplateFile(ctx context.Context, fileHeader *multipart.FileHeader) (string, map[string]map[string][]string, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	_, span := tracer.Start(ctx, "service.template.process_template_file")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	templateFile, errFile := pkgHTTP.GetFileFromHeader(fileHeader)
	if errFile != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get file from header", errFile)
		return "", nil, errFile
	}

	if err := templateUtils.ValidateNoScriptTag(templateFile); err != nil {
		errBusiness := pkg.ValidateBusinessError(constant.ErrScriptTagDetected, "")
		libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Script tag detected in template file", errBusiness)

		return "", nil, errBusiness
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
func (uc *UseCase) validateOutputFormatAndFile(ctx context.Context, id uuid.UUID, fileHeader *multipart.FileHeader, outputFormat, templateFile string) error {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.template.validate_output_format")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
	)

	// If file is provided without explicit outputFormat, validate against existing template's outputFormat
	if fileHeader != nil && commons.IsNilOrEmpty(&outputFormat) {
		outputFormatExistentTemplate, err := uc.TemplateRepo.FindOutputFormatByID(ctx, id)
		if err != nil {
			if pkgHTTP.IsBusinessError(err) {
				libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Failed to get outputFormat of template by ID", err)
			} else {
				libOpentelemetry.HandleSpanError(&span, "Failed to get outputFormat of template by ID", err)
			}

			logger.Errorf("Error to get outputFormat of template by ID, Error: %v", err)

			return err
		}

		if outputFormatExistentTemplate == nil {
			err := fmt.Errorf("output format not found for template %s", id)
			libOpentelemetry.HandleSpanError(&span, "Output format is nil", err)
			logger.Errorf("Output format is nil for template %s", id)

			return err
		}

		if errFileFormat := pkg.ValidateFileFormat(*outputFormatExistentTemplate, templateFile); errFileFormat != nil {
			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "File format validation failed", errFileFormat)

			logger.Errorf("Error to validate file format, Error: %v", errFileFormat)

			return errFileFormat
		}
	}

	// If outputFormat is explicitly provided, validate it
	if !commons.IsNilOrEmpty(&outputFormat) {
		if !pkg.IsOutputFormatValuesValid(&outputFormat) {
			errInvalidFormat := pkg.ValidateBusinessError(constant.ErrInvalidOutputFormat, "")

			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Invalid output format value", errInvalidFormat)

			logger.Errorf("Error invalid outputFormat value %v", outputFormat)

			return errInvalidFormat
		}

		if fileHeader == nil {
			errMissingFile := pkg.ValidateBusinessError(constant.ErrOutputFormatWithoutTemplateFile, "")

			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "Output format provided without template file", errMissingFile)

			logger.Error("Can not update outputFormat without passing the file template")

			return errMissingFile
		}

		if errFileFormat := pkg.ValidateFileFormat(outputFormat, templateFile); errFileFormat != nil {
			libOpentelemetry.HandleSpanBusinessErrorEvent(&span, "File format validation failed", errFileFormat)

			logger.Errorf("Error to validate file format, Error: %v", errFileFormat)

			return errFileFormat
		}
	}

	return nil
}
