package services

import (
	"context"
	"mime/multipart"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/mongodb/template"
	"plugin-smart-templates/pkg/net/http"
	templateUtils "plugin-smart-templates/pkg/template_utils"
	"reflect"
	"strings"
	"time"

	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/log"
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

	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "services.update_template")
	defer span.End()

	span.SetAttributes(
		attribute.String("template_id", id.String()),
		attribute.String("organization_id", organizationID.String()),
	)

	logger.Infof("Updating template")

	if fileHeader != nil {
		var err error

		templateFile, mappedFields, err = uc.processTemplateFile(fileHeader, logger)
		if err != nil {
			return err
		}

		if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, logger, mappedFields); errValidateFields != nil {
			logger.Errorf("Error to validate fields existence on tables, Error: %v", errValidateFields)
			return errValidateFields
		}
	}

	if fileHeader != nil && commons.IsNilOrEmpty(&outputFormat) {
		outputFormatExistentTemplate, err := uc.TemplateRepo.FindOutputFormatByID(ctx, reflect.TypeOf(template.Template{}).Name(), id, organizationID)
		if err != nil {
			logger.Errorf("Error to get outputFormat of template by ID, Error: %v", err)
			return err
		}

		if errFileFormat := pkg.ValidateFileFormat(*outputFormatExistentTemplate, templateFile); errFileFormat != nil {
			logger.Errorf("Error to validate file format, Error: %v", errFileFormat)
			return errFileFormat
		}
	}

	if !commons.IsNilOrEmpty(&outputFormat) {
		if !pkg.IsOutputFormatValuesValid(&outputFormat) {
			logger.Errorf("Error invalid outputFormat value %v", outputFormat)
			return pkg.ValidateBusinessError(constant.ErrInvalidOutputFormat, "")
		}

		if fileHeader == nil {
			logger.Error("Can not update outputFormat without passing the file template")
			return pkg.ValidateBusinessError(constant.ErrOutputFormatWithoutTemplateFile, "")
		}

		if errFileFormat := pkg.ValidateFileFormat(outputFormat, templateFile); errFileFormat != nil {
			logger.Errorf("Error to validate file format, Error: %v", errFileFormat)
			return errFileFormat
		}
	}

	setFields := uc.buildSetFields(description, outputFormat, mappedFields)
	updateFields := bson.M{}

	if len(setFields) > 0 {
		updateFields["$set"] = setFields
	}

	if errUpdate := uc.TemplateRepo.Update(ctx, reflect.TypeOf(template.Template{}).Name(), id, organizationID, &updateFields); errUpdate != nil {
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
