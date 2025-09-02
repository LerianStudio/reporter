package services

import (
	"context"
	"fmt"
	"plugin-smart-templates/v2/pkg"
	"plugin-smart-templates/v2/pkg/constant"
	"plugin-smart-templates/v2/pkg/mongodb"
	"plugin-smart-templates/v2/pkg/postgres"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
)

// ValidateIfFieldsExistOnTables Validate all fields mapped from a template file if exist on table schema
func (uc *UseCase) ValidateIfFieldsExistOnTables(ctx context.Context, organizationID string, logger log.Logger, mappedFields map[string]map[string][]string) error {
	mappedFieldsToValidate := generateCopyOfMappedFields(mappedFields, organizationID)

	for databaseName := range mappedFields {
		dataSource, exists := uc.ExternalDataSources[databaseName]
		if !exists {
			logger.Errorf("Unknown data source: %s", databaseName)
			return pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", databaseName)
		}

		switch dataSource.DatabaseType {
		case pkg.PostgreSQLType:
			if !dataSource.Initialized || !dataSource.DatabaseConfig.Connected {
				if err := pkg.ConnectToDataSource(databaseName, &dataSource, logger, uc.ExternalDataSources); err != nil {
					logger.Errorf("Error initializing database connection, Err: %s", err)
					return err
				}
			}

			errValidate := validateSchemasPostgresOfMappedFields(ctx, databaseName, dataSource, mappedFieldsToValidate)
			if errValidate != nil {
				logger.Errorf("Error to validate schemas of postgres: %s", errValidate.Error())

				return errValidate
			}
		case pkg.MongoDBType:
			if !dataSource.Initialized {
				if err := pkg.ConnectToDataSource(databaseName, &dataSource, logger, uc.ExternalDataSources); err != nil {
					logger.Errorf("Error initializing database connection, Err: %s", err)
					return err
				}
			}

			errValidate := validateSchemasMongoOfMappedFields(ctx, databaseName, dataSource, mappedFieldsToValidate)
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
	schema, err := dataSource.PostgresRepository.GetDatabaseSchema(ctx)
	if err != nil {
		return err
	}

	for _, s := range schema {
		countIfTableExist := int32(0)
		fieldsMissing := postgres.ValidateFieldsInSchemaPostgres(mappedFields[databaseName][s.TableName], s, &countIfTableExist)
		// Remove of mappedFields copies the table if exist on a schema list
		if countIfTableExist > 0 {
			if mt, ok := mappedFields[databaseName]; ok {
				delete(mt, s.TableName)
			}
		}

		if len(fieldsMissing) > 0 {
			return pkg.ValidateBusinessError(constant.ErrMissingTableFields, "", fieldsMissing)
		}
	}

	// Create an array of tables that does not exist for a database passed
	errorTables := make([]string, 0, len(mappedFields[databaseName]))
	for key := range mappedFields[databaseName] {
		errorTables = append(errorTables, key)
	}

	if len(mappedFields[databaseName]) > 0 {
		return pkg.ValidateBusinessError(constant.ErrMissingSchemaTable, "", errorTables, databaseName)
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
		countIfTableExist := int32(0)
		fieldsMissing := mongodb.ValidateFieldsInSchemaMongo(mappedFields[databaseName][s.CollectionName], s, &countIfTableExist)
		// Remove of mappedFields copies the table if exist on a schema list
		if countIfTableExist > 0 {
			if mt, ok := mappedFields[databaseName]; ok {
				delete(mt, s.CollectionName)
			}
		}

		if len(fieldsMissing) > 0 {
			return pkg.ValidateBusinessError(constant.ErrMissingTableFields, "", fieldsMissing)
		}
	}

	// Create an array of tables that does not exist for a database passed
	errorTables := make([]string, 0, len(mappedFields[databaseName]))
	for key := range mappedFields[databaseName] {
		errorTables = append(errorTables, key)
	}

	if len(mappedFields[databaseName]) > 0 {
		return pkg.ValidateBusinessError(constant.ErrMissingSchemaTable, "", errorTables, databaseName)
	}

	errClose := dataSource.MongoDBRepository.CloseConnection(ctx)
	if errClose != nil {
		return errClose
	}

	return nil
}

// generateCopyOfMappedFields generate a copy of mapped fields to make a deep copy of the original
// For plugin_crm database, table names are appended with organizationID
func generateCopyOfMappedFields(orig map[string]map[string][]string, organizationID string) map[string]map[string][]string {
	copyMappedFields := make(map[string]map[string][]string)

	for k, v := range orig {
		sub := make(map[string][]string)

		for subK, subV := range v {
			newSlice := make([]string, len(subV))
			copy(newSlice, subV)

			// For plugin_crm database, append organizationID to table names
			if k == "plugin_crm" {
				newTableName := subK + "_" + organizationID
				sub[newTableName] = newSlice
			} else {
				sub[subK] = newSlice
			}
		}

		copyMappedFields[k] = sub
	}

	return copyMappedFields
}

// TransformMappedFieldsForStorage transforms mapped fields for storage in the database
// For plugin_crm database, table names are appended with organizationID and organization mapping is added
func TransformMappedFieldsForStorage(mappedFields map[string]map[string][]string, organizationID string) map[string]map[string][]string {
	transformedFields := make(map[string]map[string][]string)

	for databaseName, tables := range mappedFields {
		transformedTables := make(map[string][]string)

		for tableName, fields := range tables {
			// For plugin_crm database, append organizationID to table names
			if databaseName == "plugin_crm" {
				newTableName := tableName
				transformedTables[newTableName] = fields
			} else {
				transformedTables[tableName] = fields
			}
		}

		// For plugin_crm database, always add organization mapping
		if databaseName == "plugin_crm" {
			transformedTables["organization"] = []string{organizationID}
		}

		transformedFields[databaseName] = transformedTables
	}

	return transformedFields
}
