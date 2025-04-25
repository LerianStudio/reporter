package services

import (
	"context"
	"fmt"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/google/uuid"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/mongodb/template"
	templateUtils "plugin-template-engine/pkg/template_utils"
	"reflect"
	"regexp"
	"time"
)

// CreateTemplate create a new template
func (uc *UseCase) CreateTemplate(ctx context.Context, templateFile, outFormat, description string, organizationID uuid.UUID) (*template.Template, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "services.create_template")
	defer span.End()

	logger.Infof("Creating template")

	mappedFields := uc.mappedFieldsOfTemplate(templateFile)
	logger.Infof("Mapped Fields is valid to continue %v", mappedFields)

	templateId := commons.GenerateUUIDv7()
	timestamp := time.Now().Unix()
	fileName := fmt.Sprintf("%s_%d.tpl", templateId.String(), timestamp)

	templateModel := &template.TemplateMongoDBModel{
		ID:             templateId,
		OutputFormat:   outFormat,
		OrganizationID: organizationID,
		FileName:       fileName,
		Description:    description,
		MappedFields:   mappedFields,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      nil,
	}

	resultTemplateModel, err := uc.TemplateRepo.Create(ctx, reflect.TypeOf(template.Template{}).Name(), templateModel)
	if err != nil {
		logger.Errorf("Error into creating a template, Error: %v", err)
		return nil, err
	}

	return resultTemplateModel, nil
}

func (uc *UseCase) mappedFieldsOfTemplate(templateFile string) map[string]any {
	// Variable present on for loops of template
	variableMap := map[string][]string{}

	// Process for loops of template
	forRegex := regexp.MustCompile(`{%-?\s*for\s+(\w+)\s+in\s+([^\s%]+)\s*-?%}`)

	forMatches := forRegex.FindAllStringSubmatch(templateFile, -1)
	for _, match := range forMatches {
		variable := match[1]
		path := templateUtils.CleanPath(match[2])
		variableMap[variable] = path
	}

	// Process {{ ... }}
	fieldRegex := regexp.MustCompile(`{{\s*([\w.\[\]_]+)\s*}}`)
	fieldMatches := fieldRegex.FindAllStringSubmatch(templateFile, -1)

	result := map[string]any{}

	for _, match := range fieldMatches {
		expr := match[1]
		parts := templateUtils.CleanPath(expr)

		if len(parts) < 2 {
			continue
		}

		if loopPath, ok := variableMap[parts[0]]; ok {
			// ex: t.id → loopPath = organization.transaction.account
			last := loopPath[len(loopPath)-1]
			templateUtils.InsertField(result, append(loopPath[:len(loopPath)-1], last), parts[1])
		} else {
			// ex: organization.user.legal_name
			templateUtils.InsertField(result, parts[:len(parts)-1], parts[len(parts)-1])
		}
	}

	return result
}
