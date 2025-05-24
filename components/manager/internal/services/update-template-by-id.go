package services

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"mime/multipart"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/mongodb/template"
	"plugin-smart-templates/pkg/net/http"
	templateUtils "plugin-smart-templates/pkg/template_utils"
	"reflect"
	"time"
)

// UpdateTemplateByID update a existent template
func (uc *UseCase) UpdateTemplateByID(ctx context.Context, outputFormat, description string, organizationID, id uuid.UUID, fileHeader *multipart.FileHeader) error {
	var (
		templateFile string
		errFile      error
	)

	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "services.update_template")
	defer span.End()

	logger.Infof("Updating template")

	setFields := bson.M{}

	if !commons.IsNilOrEmpty(&description) {
		setFields["description"] = description
	}

	if fileHeader != nil {
		templateFile, errFile = http.GetFileFromHeader(fileHeader)
		if errFile != nil {
			return errFile
		}
	}

	// Validate if updated template content is the same as existent template outputFormat
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
		// Validate if outputFormat is valid
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

		setFields["output_format"] = outputFormat
	}

	if fileHeader != nil {
		mappedFields := templateUtils.MappedFieldsOfTemplate(templateFile)
		logger.Infof("Mapped Fields is valid to continue %v", mappedFields)

		setFields["mapped_fields"] = mappedFields
	}

	setFields["updated_at"] = time.Now()

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
