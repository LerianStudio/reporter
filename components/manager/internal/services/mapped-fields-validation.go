// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb"
	"github.com/LerianStudio/reporter/pkg/postgres"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
)

// ValidateIfFieldsExistOnTables Validate all fields mapped from a template file if exist on table schema
func (uc *UseCase) ValidateIfFieldsExistOnTables(ctx context.Context, mappedFields map[string]map[string][]string) error {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.report.validate_mapped_fields")
	defer span.End()

	span.SetAttributes(attribute.String("app.request.request_id", reqId))

	logger.Infof("Validating if mapped fields exist on tables")

	allDataSources := uc.ExternalDataSources.GetAll()

	for databaseName := range mappedFields {
		if !pkg.IsValidDataSourceID(databaseName) {
			logger.Errorf("Unknown data source: %s - not in immutable registry, rejecting request", databaseName)
			return pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", databaseName)
		}

		if _, exists := allDataSources[databaseName]; !exists {
			logger.Errorf("Datasource %s is registered but not in runtime map - possible corruption", databaseName)
			return pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", databaseName)
		}
	}

	mappedFieldsToValidate := generateCopyOfMappedFields(mappedFields, allDataSources)

	for databaseName := range mappedFields {
		dataSource := allDataSources[databaseName]

		switch dataSource.DatabaseType {
		case pkg.PostgreSQLType:
			if !dataSource.Initialized || !dataSource.DatabaseConfig.Connected {
				if err := uc.ExternalDataSources.ConnectDataSource(databaseName, &dataSource, logger); err != nil {
					libOpentelemetry.HandleSpanError(&span, "Failed to initialize PostgreSQL connection", err)
					logger.Errorf("Error initializing database connection, Err: %s", err)

					return err
				}
			}

			errValidate := validateSchemasPostgresOfMappedFields(ctx, databaseName, dataSource, mappedFieldsToValidate)
			if errValidate != nil {
				libOpentelemetry.HandleSpanError(&span, "Failed to validate schemas of postgres", errValidate)
				logger.Errorf("Error to validate schemas of postgres: %s", errValidate.Error())

				return errValidate
			}
		case pkg.MongoDBType:
			if !dataSource.Initialized {
				if err := uc.ExternalDataSources.ConnectDataSource(databaseName, &dataSource, logger); err != nil {
					libOpentelemetry.HandleSpanError(&span, "Failed to initialize MongoDB connection", err)
					logger.Errorf("Error initializing database connection, Err: %s", err)

					return err
				}
			}

			errValidate := validateSchemasMongoOfMappedFields(ctx, databaseName, dataSource, mappedFieldsToValidate)
			if errValidate != nil {
				libOpentelemetry.HandleSpanError(&span, "Failed to validate collections of mongo", errValidate)
				logger.Errorf("Error to validate collections of mongo: %s", errValidate.Error())

				return errValidate
			}
		default:
			err := fmt.Errorf("unsupported database type: %s for database: %s", dataSource.DatabaseType, databaseName)
			libOpentelemetry.HandleSpanError(&span, "Unsupported database type", err)

			return err
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

	// Early-fail validation: check for schema ambiguity on tables without explicit schema
	// This prevents templates from being saved with ambiguous references that would fail at report generation
	if err := validateSchemaAmbiguity(databaseName, schema, mappedFields); err != nil {
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
	var (
		schema []mongodb.CollectionSchema
		err    error
	)

	// For plugin_crm with MidazOrganizationID, fetch only collections for that organization
	if dataSource.MidazOrganizationID != "" {
		schema, err = dataSource.MongoDBRepository.GetDatabaseSchemaForOrganization(ctx, dataSource.MidazOrganizationID)
	} else {
		schema, err = dataSource.MongoDBRepository.GetDatabaseSchema(ctx)
	}

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
// For plugin_crm database, table names are appended with MidazOrganizationID from datasource config
func generateCopyOfMappedFields(orig map[string]map[string][]string, dataSources map[string]pkg.DataSource) map[string]map[string][]string {
	copyMappedFields := make(map[string]map[string][]string)

	for k, v := range orig {
		sub := make(map[string][]string)

		for subK, subV := range v {
			newSlice := make([]string, len(subV))
			copy(newSlice, subV)

			// For plugin_crm database, append MidazOrganizationID to table names
			if k == "plugin_crm" {
				if ds, exists := dataSources[k]; exists && ds.MidazOrganizationID != "" {
					newTableName := subK + "_" + ds.MidazOrganizationID
					sub[newTableName] = newSlice
				} else {
					sub[subK] = newSlice
				}
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

		// For plugin_crm database, add organization mapping only if organizationID is provided
		if databaseName == "plugin_crm" && organizationID != "" {
			transformedTables["organization"] = []string{organizationID}
		}

		transformedFields[databaseName] = transformedTables
	}

	return transformedFields
}

// validateSchemaAmbiguity checks for schema ambiguity on tables without explicit schema reference.
// This is an early-fail validation that prevents templates with ambiguous table references from being saved.
// A table reference is ambiguous when:
//   - The table name has no explicit schema (no "__" separator in Pongo2 format)
//   - The table exists in multiple schemas
//   - None of those schemas is "public" (which would be used as default)
func validateSchemaAmbiguity(databaseName string, schema []postgres.TableSchema, mappedFields map[string]map[string][]string) error {
	// Build a map of table name -> list of schemas where it exists
	tableSchemas := make(map[string][]string)
	for _, s := range schema {
		tableSchemas[s.TableName] = append(tableSchemas[s.TableName], s.SchemaName)
	}

	// Check each table in mappedFields for this database
	for tableKey := range mappedFields[databaseName] {
		// Skip tables with explicit schema (Pongo2 format: schema__table)
		if strings.Contains(tableKey, "__") {
			continue
		}

		// Skip tables with qualified format (schema.table)
		if strings.Contains(tableKey, ".") {
			continue
		}

		// This is a table without explicit schema - check for ambiguity
		schemas, exists := tableSchemas[tableKey]
		if !exists {
			// Table doesn't exist - will be caught by existing validation
			continue
		}

		if len(schemas) > 1 {
			// Table exists in multiple schemas - check if "public" is one of them
			hasPublic := false

			for _, s := range schemas {
				if s == "public" {
					hasPublic = true
					break
				}
			}

			if !hasPublic {
				// Ambiguous reference: table exists in multiple schemas, none is "public"
				return pkg.ValidateBusinessError(constant.ErrSchemaAmbiguous, "", tableKey, schemas)
			}
		}
	}

	return nil
}
