package services

import (
	"context"
	"fmt"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/log"
	"github.com/google/uuid"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/mongodb"
	"plugin-smart-templates/pkg/mongodb/template"
	"plugin-smart-templates/pkg/postgres"
	templateUtils "plugin-smart-templates/pkg/template_utils"
	"reflect"
	"time"
)

// CreateTemplate creates a new template with specified parameters and stores it in the repository.
func (uc *UseCase) CreateTemplate(ctx context.Context, templateFile, outFormat, description string, organizationID uuid.UUID) (*template.Template, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "services.create_template")
	defer span.End()

	logger.Infof("Creating template")

	mappedFields := templateUtils.MappedFieldsOfTemplate(templateFile)
	logger.Infof("Mapped Fields is valid to continue %v", mappedFields)

	if errValidateFields := uc.validateIfFieldsExistOnTables(ctx, logger, mappedFields); errValidateFields != nil {
		logger.Errorf("Error to validate fields existence on tables, Error: %v", errValidateFields)
		return nil, errValidateFields
	}

	templateId := commons.GenerateUUIDv7()
	fileName := fmt.Sprintf("%s.tpl", templateId.String())

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

// validateIfFieldsExistOnTables Validate all fields mapped from a template file if exist on table schema
func (uc *UseCase) validateIfFieldsExistOnTables(ctx context.Context, logger log.Logger, mappedFields map[string]map[string][]string) error {
	for databaseName := range mappedFields {
		dataSource, exists := uc.ExternalDataSources[databaseName]
		if !exists {
			logger.Errorf("Unknown data source: %s", databaseName)
			return pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", databaseName)
		}

		if !dataSource.Initialized {
			if err := pkg.ConnectToDataSource("transaction", &dataSource, logger, uc.ExternalDataSources); err != nil {
				logger.Errorf("Error initializing database connection, Err: %s", err)
				return err
			}
		}

		switch dataSource.DatabaseType {
		case pkg.PostgreSQLType:
			errValidate := validateSchemasPostgresOfMappedFields(ctx, databaseName, dataSource, mappedFields)
			if errValidate != nil {
				logger.Errorf("Error to validate schemas of postgres: %s", errValidate.Error())
				return errValidate
			}
		case pkg.MongoDBType:
			errValidate := validateSchemasMongoOfMappedFields(ctx, databaseName, dataSource, mappedFields)
			if errValidate != nil {
				logger.Errorf("Error to validate collections of mongo: %s", errValidate.Error())
				return errValidate
			}
		default:
			return fmt.Errorf("unsupported database type: %s for database: %s", dataSource.DatabaseType, databaseName)
		}
	}

	return nil
}

// validateSchemasPostgresOfMappedFields validate if mapped fields exist on schemas tables columns
func validateSchemasPostgresOfMappedFields(ctx context.Context, databaseName string, dataSource pkg.DataSource, mappedFields map[string]map[string][]string) error {
	var countIfTableExist = int32(0)

	schema, err := dataSource.PostgresRepository.GetDatabaseSchema(ctx)
	if err != nil {
		return err
	}

	for _, s := range schema {
		fieldsMissing := postgres.ValidateFieldsInSchemaPostgres(mappedFields[databaseName][s.TableName], s, &countIfTableExist)
		if len(fieldsMissing) > 0 {
			return pkg.ValidateBusinessError(constant.ErrMissingTableFields, "", fieldsMissing)
		}
	}

	if countIfTableExist == 0 {
		return pkg.ValidateBusinessError(constant.ErrMissingSchemaTable, "", databaseName)
	}

	errClose := dataSource.PostgresRepository.CloseConnection()
	if errClose != nil {
		return errClose
	}

	return nil
}

// validateSchemasMongoOfMappedFields validate if mapped fields exist on schemas tables fields of MongoDB
func validateSchemasMongoOfMappedFields(ctx context.Context, databaseName string, dataSource pkg.DataSource, mappedFields map[string]map[string][]string) error {
	schema, err := dataSource.MongoDBRepository.GetDatabaseSchema(ctx)
	if err != nil {
		return err
	}

	for _, s := range schema {
		fieldsMissing := mongodb.ValidateFieldsInSchemaMongo(mappedFields[databaseName][s.CollectionName], s)
		if len(fieldsMissing) > 0 {
			return pkg.ValidateBusinessError(constant.ErrMissingTableFields, "", fieldsMissing)
		}
	}

	errClose := dataSource.MongoDBRepository.CloseConnection(ctx)
	if errClose != nil {
		return errClose
	}

	return nil
}
