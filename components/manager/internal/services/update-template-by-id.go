package services

import (
	"context"
	"mime/multipart"
	"strings"
	"time"

	"github.com/LerianStudio/reporter/v3/pkg"
	"github.com/LerianStudio/reporter/v3/pkg/constant"
	"github.com/LerianStudio/reporter/v3/pkg/net/http"
	templateUtils "github.com/LerianStudio/reporter/v3/pkg/template_utils"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.opentelemetry.io/otel/attribute"
)

// UpdateTemplateByID update a existent template
func (uc *UseCase) UpdateTemplateByID(ctx context.Context, outputFormat, description string, organizationID, id uuid.UUID, fileHeader *multipart.FileHeader) error {
	var (
		templateFile string
		mappedFields map[string]map[string][]string
	)

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.update_template_by_id")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
		attribute.String("app.request.organization_id", organizationID.String()),
	)

	logger.Infof("Updating template")

	if fileHeader != nil {
		var err error

		templateFile, mappedFields, err = uc.processTemplateFile(fileHeader, logger)
		if err != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to process template file", err)

			return err
		}

		if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, organizationID.String(), logger, mappedFields); errValidateFields != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to validate fields existence on tables", errValidateFields)

			logger.Errorf("Error to validate fields existence on tables, Error: %v", errValidateFields)

			return errValidateFields
		}
	}

	if fileHeader != nil && commons.IsNilOrEmpty(&outputFormat) {
		outputFormatExistentTemplate, err := uc.TemplateRepo.FindOutputFormatByID(ctx, id, organizationID)
		if err != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to get outputFormat of template by ID", err)

			logger.Errorf("Error to get outputFormat of template by ID, Error: %v", err)

			return err
		}

		if errFileFormat := pkg.ValidateFileFormat(*outputFormatExistentTemplate, templateFile); errFileFormat != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to validate file format", errFileFormat)

			logger.Errorf("Error to validate file format, Error: %v", errFileFormat)

			return errFileFormat
		}
	}

	if !commons.IsNilOrEmpty(&outputFormat) {
		if !pkg.IsOutputFormatValuesValid(&outputFormat) {
			errInvalidFormat := pkg.ValidateBusinessError(constant.ErrInvalidOutputFormat, "")
			libOpentelemetry.HandleSpanError(&span, "Invalid outputFormat value", errInvalidFormat)

			logger.Errorf("Error invalid outputFormat value %v", outputFormat)

			return errInvalidFormat
		}

		if fileHeader == nil {
			errMissingFile := pkg.ValidateBusinessError(constant.ErrOutputFormatWithoutTemplateFile, "")
			libOpentelemetry.HandleSpanError(&span, "Cannot update outputFormat without template file", errMissingFile)

			logger.Error("Can not update outputFormat without passing the file template")

			return errMissingFile
		}

		if errFileFormat := pkg.ValidateFileFormat(outputFormat, templateFile); errFileFormat != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to validate file format", errFileFormat)

			logger.Errorf("Error to validate file format, Error: %v", errFileFormat)

			return errFileFormat
		}
	}

	setFields := uc.buildSetFields(description, outputFormat, mappedFields)
	updateFields := bson.M{}

	if len(setFields) > 0 {
		updateFields["$set"] = setFields
	}

	if errUpdate := uc.TemplateRepo.Update(ctx, id, organizationID, &updateFields); errUpdate != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to update template in repository", errUpdate)

		logger.Errorf("Error into creating a template, Error: %v", errUpdate)

		return errUpdate
	}

	return nil
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
