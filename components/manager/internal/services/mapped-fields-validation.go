// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb"
	"github.com/LerianStudio/reporter/v4/pkg/postgres"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
)

// ValidateIfFieldsExistOnTables Validate all fields mapped from a template file if exist on table schema
func (uc *UseCase) ValidateIfFieldsExistOnTables(ctx context.Context, organizationID string, logger log.Logger, mappedFields map[string]map[string][]string) error {
	for databaseName := range mappedFields {
		if !pkg.IsValidDataSourceID(databaseName) {
			logger.Errorf("Unknown data source: %s - not in immutable registry, rejecting request", databaseName)
			return pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", databaseName)
		}

		if _, exists := uc.ExternalDataSources[databaseName]; !exists {
			logger.Errorf("Datasource %s is registered but not in runtime map - possible corruption", databaseName)
			return pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", databaseName)
		}
	}

	mappedFieldsToValidate := generateCopyOfMappedFields(mappedFields, organizationID)

	for databaseName := range mappedFields {
		dataSource := uc.ExternalDataSources[databaseName]

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
	// Use configured schemas or default to public
	configuredSchemas := dataSource.Schemas
	if len(configuredSchemas) == 0 {
		configuredSchemas = []string{"public"}
	}

	schema, err := dataSource.PostgresRepository.GetDatabaseSchema(ctx, configuredSchemas)
	if err != nil {
		return err
	}

	for _, s := range schema {
		countIfTableExist := int32(0)

		// Support multiple formats:
		// - Legacy: mappedFields[database][table] (e.g., "transfers")
		// - Qualified with dot: mappedFields[database][schema.table] (e.g., "payment.transfers")
		// - Pongo2 compatible: mappedFields[database][schema__table] (e.g., "payment__transfers")
		qualifiedTableName := s.QualifiedName()                              // "schema.table" format
		pongo2TableName := strings.Replace(qualifiedTableName, ".", "__", 1) // "schema__table" format for Pongo2
		tableKey := s.TableName                                              // legacy format

		// Check if fields exist for any of the supported formats
		var (
			fieldsToValidate []string
			keyToDelete      string
		)

		switch {
		case mappedFields[databaseName][pongo2TableName] != nil:
			fieldsToValidate = mappedFields[databaseName][pongo2TableName]
			keyToDelete = pongo2TableName
		case mappedFields[databaseName][qualifiedTableName] != nil:
			fieldsToValidate = mappedFields[databaseName][qualifiedTableName]
			keyToDelete = qualifiedTableName
		case mappedFields[databaseName][tableKey] != nil:
			fieldsToValidate = mappedFields[databaseName][tableKey]
			keyToDelete = tableKey
		}

		if len(fieldsToValidate) > 0 {
			fieldsMissing := postgres.ValidateFieldsInSchemaPostgres(fieldsToValidate, s, &countIfTableExist)
			// Remove of mappedFields copies the table if exist on a schema list
			if countIfTableExist > 0 {
				if mt, ok := mappedFields[databaseName]; ok {
					delete(mt, keyToDelete)
				}
			}

			if len(fieldsMissing) > 0 {
				return pkg.ValidateBusinessError(constant.ErrMissingTableFields, "", fieldsMissing)
			}
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
