// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/model"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
)

// Repository defines an interface for querying data from a specified table and fields.
//
//go:generate mockgen --destination=datasource.postgres.mock.go --package=postgres . Repository
type Repository interface {
	Query(ctx context.Context, schema []TableSchema, schemaName string, table string, fields []string, filter map[string][]any) ([]map[string]any, error)
	QueryWithAdvancedFilters(ctx context.Context, schema []TableSchema, schemaName string, table string, fields []string, filter map[string]model.FilterCondition) ([]map[string]any, error)
	GetDatabaseSchema(ctx context.Context, schemas []string) ([]TableSchema, error)
	CloseConnection() error
}

// qualifyTableName returns a qualified table name with schema if provided.
// If schemaName is empty, returns just the table name.
// If schemaName is provided, returns "schema"."table" format.
func qualifyTableName(schemaName, tableName string) string {
	if schemaName == "" {
		return tableName
	}

	return fmt.Sprintf(`"%s"."%s"`, schemaName, tableName)
}

// TableSchema represents the structure of a database table
type TableSchema struct {
	SchemaName string              `json:"schema_name"`
	TableName  string              `json:"table_name"`
	Columns    []ColumnInformation `json:"columns"`
}

// QualifiedName returns the fully qualified table name in the format "schema.table"
func (t TableSchema) QualifiedName() string {
	return fmt.Sprintf("%s.%s", t.SchemaName, t.TableName)
}

// ColumnInformation contains the details of a database column
type ColumnInformation struct {
	Name         string `json:"name"`
	DataType     string `json:"data_type"`
	IsNullable   bool   `json:"is_nullable"`
	IsPrimaryKey bool   `json:"is_primary_key"`
}

// ExternalDataSource provides an interface for interacting with a PostgreSQL database connection.
type ExternalDataSource struct {
	connection *Connection
}

// Compile-time interface satisfaction check.
var _ Repository = (*ExternalDataSource)(nil)

// NewDataSourceRepository creates a new ExternalDataSource instance using the provided postgres.Connection, initializing the database connection.
// Returns nil and error if connection fails.
func NewDataSourceRepository(pc *Connection) (*ExternalDataSource, error) {
	c := &ExternalDataSource{
		connection: pc,
	}

	_, err := c.connection.GetDB()
	if err != nil {
		pc.Logger.Errorf("Failed to establish PostgreSQL connection: %v", err)
		return nil, fmt.Errorf("failed to establish PostgreSQL connection: %w", err)
	}

	return c, nil
}

// CloseConnection closing the connection with PostgreSQL.
func (ds *ExternalDataSource) CloseConnection() error {
	if ds.connection.ConnectionDB != nil {
		ds.connection.Logger.Info("Closing connection to PostgreSQL...")

		err := ds.connection.ConnectionDB.Close()
		if err != nil {
			ds.connection.Logger.Errorf("Error closing PostgreSQL connection: %v", err)
			return err
		}

		ds.connection.Connected = false
		ds.connection.ConnectionDB = nil
		ds.connection.Logger.Info("PostgreSQL connection closed successfully.")
	}

	return nil
}

// Query executes a SELECT SQL query on the specified table with the given fields and filter criteria.
// The schemaName parameter specifies the database schema to query from (e.g., "public", "payment").
// If schemaName is empty, the table name is used without schema qualification.
// It returns the query results as a slice of maps or an error in case of failure.
func (ds *ExternalDataSource) Query(ctx context.Context, schema []TableSchema, schemaName string, table string, fields []string, filter map[string][]any) ([]map[string]any, error) {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.datasource.query")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.repository_filter", map[string]any{
		"schema": schemaName,
		"table":  table,
		"fields": fields,
		"filter": filter,
	})
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert repository filter to JSON string", err)
	}

	qualifiedTable := qualifyTableName(schemaName, table)
	logger.Infof("Querying %s table with fields %v", qualifiedTable, fields)

	// Validate requested table and fields
	queriedFields, err := ds.ValidateTableAndFields(ctx, table, fields, schema)
	if err != nil {
		return nil, err
	}

	// Transform nested JSONB fields to proper PostgreSQL accessor syntax
	selectFields := transformFieldsForSelect(queriedFields)

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select(selectFields...).From(qualifiedTable)

	// Apply filters, but only if they correspond to valid columns
	queryBuilder = buildDynamicFilters(queryBuilder, schema, table, filter)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error generating SQL: %w", err)
	}

	logger.Infof("Executing SQL: %s with args: %v", query, args)

	// Create timeout context for query execution
	queryCtx, cancel := context.WithTimeout(ctx, constant.QueryTimeoutMedium)
	defer cancel()

	rows, err := ds.connection.ConnectionDB.QueryContext(queryCtx, query, args...)
	if err != nil {
		if queryCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("query execution timeout after %v: %w", constant.QueryTimeoutMedium, err)
		}

		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	return scanRows(rows, logger)
}

// tableInfo holds schema and table name for internal processing
type tableInfo struct {
	schemaName string
	tableName  string
}

// GetDatabaseSchema retrieves all tables and their column details from the specified schemas.
// The schemas parameter specifies which database schemas to query (e.g., ["public", "payment", "transfer"]).
// It returns a slice of TableSchema objects with SchemaName populated, or an error if the operation fails.
func (ds *ExternalDataSource) GetDatabaseSchema(ctx context.Context, schemas []string) ([]TableSchema, error) {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.datasource.get_database_schema")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	logger.Infof("Retrieving database schema information for schemas: %v", schemas)

	// Create timeout context for schema discovery (longer timeout for this operation)
	schemaCtx, cancel := context.WithTimeout(ctx, constant.SchemaDiscoveryTimeout)
	defer cancel()

	tables, err := ds.queryTables(schemaCtx, schemas)
	if err != nil {
		return nil, err
	}

	primaryKeys, err := ds.queryPrimaryKeys(schemaCtx, schemas)
	if err != nil {
		return nil, err
	}

	result, err := ds.buildTableSchemas(schemaCtx, tables, primaryKeys)
	if err != nil {
		return nil, err
	}

	logger.Infof("Retrieved schema for %d tables across %d schemas", len(result), len(schemas))

	return result, nil
}

// queryTables retrieves all user tables from the specified schemas.
func (ds *ExternalDataSource) queryTables(ctx context.Context, schemas []string) ([]tableInfo, error) {
	tableQuery := `
		SELECT table_schema, table_name
		FROM information_schema.tables
		WHERE table_schema = ANY($1)
		AND table_type = 'BASE TABLE'
		ORDER BY table_schema, table_name
	`

	rows, err := ds.connection.ConnectionDB.QueryContext(ctx, tableQuery, pq.Array(schemas))
	if err != nil {
		return nil, fmt.Errorf("error querying tables: %w", err)
	}
	defer rows.Close()

	var tables []tableInfo

	for rows.Next() {
		var info tableInfo
		if err := rows.Scan(&info.schemaName, &info.tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %w", err)
		}

		tables = append(tables, info)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %w", err)
	}

	return tables, nil
}

// queryPrimaryKeys retrieves primary key information for all tables in the specified schemas.
func (ds *ExternalDataSource) queryPrimaryKeys(ctx context.Context, schemas []string) (map[string]map[string]bool, error) {
	pkQuery := `
		SELECT tc.table_schema, tc.table_name, kc.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kc
			ON kc.table_name = tc.table_name
			AND kc.table_schema = tc.table_schema
			AND kc.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'PRIMARY KEY'
		AND tc.table_schema = ANY($1)
	`

	pkRows, err := ds.connection.ConnectionDB.QueryContext(ctx, pkQuery, pq.Array(schemas))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("schema discovery timeout after %v while querying primary keys: %w", constant.SchemaDiscoveryTimeout, err)
		}

		return nil, fmt.Errorf("error querying primary keys: %w", err)
	}
	defer pkRows.Close()

	primaryKeys := make(map[string]map[string]bool)

	for pkRows.Next() {
		var schemaName, tableName, columnName string
		if err := pkRows.Scan(&schemaName, &tableName, &columnName); err != nil {
			return nil, fmt.Errorf("error scanning primary key info: %w", err)
		}

		key := schemaName + "." + tableName
		if primaryKeys[key] == nil {
			primaryKeys[key] = make(map[string]bool)
		}

		primaryKeys[key][columnName] = true
	}

	if err := pkRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating primary keys: %w", err)
	}

	return primaryKeys, nil
}

// buildTableSchemas builds the complete schema information for all tables.
func (ds *ExternalDataSource) buildTableSchemas(ctx context.Context, tables []tableInfo, primaryKeys map[string]map[string]bool) ([]TableSchema, error) {
	result := make([]TableSchema, 0, len(tables))

	for _, tbl := range tables {
		columns, err := ds.queryTableColumns(ctx, tbl, primaryKeys)
		if err != nil {
			return nil, err
		}

		result = append(result, TableSchema{
			SchemaName: tbl.schemaName,
			TableName:  tbl.tableName,
			Columns:    columns,
		})
	}

	return result, nil
}

// queryTableColumns retrieves column information for a specific table.
func (ds *ExternalDataSource) queryTableColumns(ctx context.Context, tbl tableInfo, primaryKeys map[string]map[string]bool) ([]ColumnInformation, error) {
	columnQuery := `
		SELECT column_name, data_type,
		       CASE WHEN is_nullable = 'YES' THEN true ELSE false END as is_nullable
		FROM information_schema.columns
		WHERE table_schema = $1
		AND table_name = $2
		ORDER BY ordinal_position
	`

	colRows, err := ds.connection.ConnectionDB.QueryContext(ctx, columnQuery, tbl.schemaName, tbl.tableName)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("schema discovery timeout after %v while querying columns for table %s.%s: %w", constant.SchemaDiscoveryTimeout, tbl.schemaName, tbl.tableName, err)
		}

		return nil, fmt.Errorf("error querying columns for table %s.%s: %w", tbl.schemaName, tbl.tableName, err)
	}
	defer colRows.Close()

	var columns []ColumnInformation

	pkKey := tbl.schemaName + "." + tbl.tableName

	for colRows.Next() {
		var col ColumnInformation
		if err := colRows.Scan(&col.Name, &col.DataType, &col.IsNullable); err != nil {
			return nil, fmt.Errorf("error scanning column info: %w", err)
		}

		if pkCols, exists := primaryKeys[pkKey]; exists {
			col.IsPrimaryKey = pkCols[col.Name]
		}

		columns = append(columns, col)
	}

	if err := colRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %w", err)
	}

	return columns, nil
}

// scanRows processes the query rows and creates the resulting slice of maps.
func scanRows(rows *sql.Rows, logger log.Logger) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting column names: %w", err)
	}

	values := make([]any, len(columns))
	pointers := make([]any, len(columns))

	for i := range values {
		pointers[i] = &values[i]
	}

	var result []map[string]any

	for rows.Next() {
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}

		rowMap := createRowMap(columns, values, logger)
		result = append(result, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

// createRowMap maps column names to their respective values.
func createRowMap(columns []string, values []any, logger log.Logger) map[string]any {
	rowMap := make(map[string]any)

	for i, column := range columns {
		// Attempt to parse any value that could be JSONB
		rowMap[column] = parseJSONBField(values[i], logger)
	}

	return rowMap
}

// parseJSONBField unmarshals any field that might be a JSONB type
func parseJSONBField(value any, logger log.Logger) any {
	if value == nil {
		return nil
	}

	// Check if the value is []uint8, which is how the PostgreSQL driver
	// represents JSONB and JSON fields
	if byteData, ok := value.([]uint8); ok {
		// Try to deserialize as a generic map[string]any
		var jsonMap map[string]any
		if err := json.Unmarshal(byteData, &jsonMap); err == nil {
			return jsonMap
		}

		// If object parsing fails, try as array
		var jsonArray []any
		if err := json.Unmarshal(byteData, &jsonArray); err == nil {
			return jsonArray
		}

		// Try as string in case it's a JSON string format
		var jsonString string
		if err := json.Unmarshal(byteData, &jsonString); err == nil {
			return jsonString
		}

		// If all attempts fail, log a warning and return the original value
		logger.Warnf("Failed to unmarshal potential JSONB data for value: %v", string(byteData))
	}

	return value
}

// extractRootColumn extracts the root column name from a potentially nested JSONB field path.
// For example: "fee_charge.totalAmount" returns "fee_charge"
// Simple fields (without dots) are returned as-is.
// This allows the entire JSONB column to be selected, which is then parsed by parseJSONBField
// into a nested map structure that the template engine can traverse.
func extractRootColumn(field string) string {
	if dotIdx := strings.Index(field, "."); dotIdx != -1 {
		return field[:dotIdx]
	}

	return field
}

// transformFieldsForSelect converts a list of fields to SQL-safe column names.
// For nested JSONB field paths (e.g., "fee_charge.totalAmount"), only the root column
// is included in the SELECT. The JSONB column will be parsed into a nested map by
// parseJSONBField, allowing the template engine to access nested values.
// This also deduplicates columns to avoid selecting the same column multiple times.
func transformFieldsForSelect(fields []string) []string {
	seen := make(map[string]bool)

	var result []string

	for _, field := range fields {
		rootColumn := extractRootColumn(field)
		if !seen[rootColumn] {
			seen[rootColumn] = true
			result = append(result, rootColumn)
		}
	}

	return result
}

// ValidateTableAndFields checks if the specified table exists and validates that
// all requested fields exist in that table.
// It returns a list of valid fields and an error if the table doesn't exist or fields are invalid.
func (ds *ExternalDataSource) ValidateTableAndFields(ctx context.Context, tableName string, requestedFields []string, schema []TableSchema) ([]string, error) {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	_, span := tracer.Start(ctx, "repository.datasource.validate_table_and_fields")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.repository_filter", map[string]any{
		"table":  tableName,
		"fields": requestedFields,
		"schema": schema,
	})
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert repository filter to JSON string", err)
	}

	logger.Infof("Validating table '%s' and fields %v", tableName, requestedFields)

	// Check if table exists
	var tableFound bool

	var tableColumns []ColumnInformation

	for _, table := range schema {
		if table.TableName == tableName {
			tableFound = true
			tableColumns = table.Columns

			break
		}
	}

	if !tableFound {
		return nil, fmt.Errorf("table '%s' does not exist in the database", tableName)
	}

	// Create a map of valid column names for efficient lookup
	validColumns := make(map[string]bool)
	for _, col := range tableColumns {
		validColumns[col.Name] = true
	}

	// Special case: if "*" is in the fields, return all columns
	if len(requestedFields) == 1 && requestedFields[0] == "*" {
		allFields := make([]string, len(tableColumns))
		for i, col := range tableColumns {
			allFields[i] = col.Name
		}

		return allFields, nil
	}

	// Validate each requested field
	var validFields []string

	var invalidFields []string

	for _, field := range requestedFields {
		// Handle nested JSONB field paths (e.g., "fee_charge.totalAmount")
		// For nested paths, validate that the root column exists
		fieldToCheck := field
		if dotIdx := strings.Index(field, "."); dotIdx != -1 {
			fieldToCheck = field[:dotIdx]
		}

		if validColumns[fieldToCheck] {
			validFields = append(validFields, field)
		} else {
			invalidFields = append(invalidFields, field)
		}
	}

	if len(invalidFields) > 0 {
		return nil, fmt.Errorf("invalid fields for table '%s': %v", tableName, invalidFields)
	}

	if len(validFields) == 0 {
		return nil, fmt.Errorf("no valid fields specified for table '%s'", tableName)
	}

	logger.Infof("Successfully validated table '%s' and fields %v", tableName, validFields)

	return validFields, nil
}

// buildDynamicFilters applies filter criteria to the query builder based on valid columns.
func buildDynamicFilters(queryBuilder squirrel.SelectBuilder, schema []TableSchema, table string, filter map[string][]any) squirrel.SelectBuilder {
	// Find the table's column information
	var tableColumns []ColumnInformation

	for _, t := range schema {
		if t.TableName == table {
			tableColumns = t.Columns
			break
		}
	}

	// Create a map of valid column names for efficient lookup
	validColumns := make(map[string]bool)
	for _, col := range tableColumns {
		validColumns[col.Name] = true
	}

	for field, values := range filter {
		// Only apply filters for valid columns
		if validColumns[field] && len(values) > 0 {
			queryBuilder = applyFilter(queryBuilder, field, values)
		}
	}

	return queryBuilder
}

// applyFilter adds a WHERE condition for a field with multiple possible values.
func applyFilter(queryBuilder squirrel.SelectBuilder, fieldName string, values []any) squirrel.SelectBuilder {
	if len(values) == 0 {
		return queryBuilder
	}

	// No need for conversion since values is already []any
	placeholder := squirrel.Placeholders(len(values))

	return queryBuilder.Where(fieldName+" IN ("+placeholder+")", values...)
}

// QueryWithAdvancedFilters executes a SELECT SQL query with advanced FilterCondition support.
// The schemaName parameter specifies the database schema to query from (e.g., "public", "payment").
// If schemaName is empty, the table name is used without schema qualification.
func (ds *ExternalDataSource) QueryWithAdvancedFilters(ctx context.Context, schema []TableSchema, schemaName string, table string, fields []string, filter map[string]model.FilterCondition) ([]map[string]any, error) {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.datasource.query_with_advanced_filters")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.repository_filter", map[string]any{
		"schema": schemaName,
		"table":  table,
		"fields": fields,
		"filter": filter,
	})
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert repository filter to JSON string", err)
	}

	qualifiedTable := qualifyTableName(schemaName, table)
	logger.Infof("Querying %s table with advanced filters on fields %v", qualifiedTable, fields)

	// Validate requested table and fields
	queriedFields, err := ds.ValidateTableAndFields(ctx, table, fields, schema)
	if err != nil {
		return nil, err
	}

	// Transform nested JSONB fields to proper PostgreSQL accessor syntax
	selectFields := transformFieldsForSelect(queriedFields)

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select(selectFields...).From(qualifiedTable)

	// Apply advanced filters
	queryBuilder, err = ds.buildAdvancedFilters(queryBuilder, schema, table, filter)
	if err != nil {
		return nil, fmt.Errorf("error building advanced filters: %w", err)
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error generating SQL: %w", err)
	}

	logger.Infof("[DEBUG] Executing advanced filter SQL: %s", query)
	logger.Infof("[DEBUG] SQL args: %v", args)
	logger.Infof("[DEBUG] Original filter conditions: %+v", filter)

	// Create timeout context for query execution (slower timeout for advanced filters)
	queryCtx, cancel := context.WithTimeout(ctx, constant.QueryTimeoutSlow)
	defer cancel()

	rows, err := ds.connection.ConnectionDB.QueryContext(queryCtx, query, args...)
	if err != nil {
		if queryCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("advanced filter query timeout after %v: %w", constant.QueryTimeoutSlow, err)
		}

		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	return scanRows(rows, logger)
}

// buildAdvancedFilters applies FilterCondition criteria to the query builder
func (ds *ExternalDataSource) buildAdvancedFilters(queryBuilder squirrel.SelectBuilder, schema []TableSchema, table string, filter map[string]model.FilterCondition) (squirrel.SelectBuilder, error) {
	var tableColumns []ColumnInformation

	for _, t := range schema {
		if t.TableName == table {
			tableColumns = t.Columns
			break
		}
	}

	// Create a map of valid column names for efficient lookup
	validColumns := make(map[string]bool)
	for _, col := range tableColumns {
		validColumns[col.Name] = true
	}

	for field, condition := range filter {
		// Only apply filters for valid columns
		if !validColumns[field] {
			continue
		}

		if isFilterConditionEmpty(condition) {
			continue
		}

		// Validate the condition
		if err := validateFilterCondition(field, condition); err != nil {
			return queryBuilder, err
		}

		// Apply each filter operator
		queryBuilder = ds.applyAdvancedFilter(queryBuilder, field, condition)
	}

	return queryBuilder, nil
}

// applyAdvancedFilter applies a single FilterCondition to the query builder
func (ds *ExternalDataSource) applyAdvancedFilter(queryBuilder squirrel.SelectBuilder, field string, condition model.FilterCondition) squirrel.SelectBuilder {
	// Handle equals (IN clause for multiple values, = for single value)
	if len(condition.Equals) > 0 {
		if len(condition.Equals) == 1 {
			queryBuilder = queryBuilder.Where(squirrel.Eq{field: condition.Equals[0]})
		} else {
			queryBuilder = queryBuilder.Where(squirrel.Eq{field: condition.Equals})
		}
	}

	// Handle greater than
	if len(condition.GreaterThan) > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Gt{field: condition.GreaterThan[0]})
	}

	// Handle greater than or equal
	if len(condition.GreaterOrEqual) > 0 {
		queryBuilder = queryBuilder.Where(squirrel.GtOrEq{field: condition.GreaterOrEqual[0]})
	}

	// Handle less than
	if len(condition.LessThan) > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{field: condition.LessThan[0]})
	}

	// Handle less than or equal
	if len(condition.LessOrEqual) > 0 {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{field: condition.LessOrEqual[0]})
	}

	// Handle between (using AND with >= and <=)
	if len(condition.Between) == constant.BetweenOperatorValues {
		// For date fields, ensure proper date range handling
		startValue := condition.Between[0]
		endValue := condition.Between[1]

		// If it looks like a date field and we have date strings, adjust the end date to include the full day
		if isDateField(field) && isDateString(startValue) && isDateString(endValue) {
			// Convert end date to end of day (23:59:59.999)
			if endStr, ok := endValue.(string); ok {
				// If it's just a date (YYYY-MM-DD), add time to make it end of day
				if len(endStr) == constant.DateOnlyStringLength { // YYYY-MM-DD format
					endValue = endStr + "T23:59:59.999Z"
				}
			}
		}

		queryBuilder = queryBuilder.Where(squirrel.GtOrEq{field: startValue}).Where(squirrel.LtOrEq{field: endValue})
	}

	// Handle in
	if len(condition.In) > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Eq{field: condition.In})
	}

	// Handle not in
	if len(condition.NotIn) > 0 {
		queryBuilder = queryBuilder.Where(squirrel.NotEq{field: condition.NotIn})
	}

	return queryBuilder
}

// isFilterConditionEmpty checks if a FilterCondition has no active filters
func isFilterConditionEmpty(condition model.FilterCondition) bool {
	return len(condition.Equals) == 0 &&
		len(condition.GreaterThan) == 0 &&
		len(condition.GreaterOrEqual) == 0 &&
		len(condition.LessThan) == 0 &&
		len(condition.LessOrEqual) == 0 &&
		len(condition.Between) == 0 &&
		len(condition.In) == 0 &&
		len(condition.NotIn) == 0
}

// validateFilterCondition validates that a FilterCondition has proper values for each operator
func validateFilterCondition(fieldName string, condition model.FilterCondition) error {
	// Validate between operator has exactly 2 values
	if len(condition.Between) > 0 && len(condition.Between) != 2 {
		return fmt.Errorf("between operator for field '%s' must have exactly 2 values, got %d", fieldName, len(condition.Between))
	}

	// Validate single-value operators have exactly 1 value
	singleValueOps := map[string][]any{
		"gt":  condition.GreaterThan,
		"gte": condition.GreaterOrEqual,
		"lt":  condition.LessThan,
		"lte": condition.LessOrEqual,
	}

	for opName, values := range singleValueOps {
		if len(values) > 0 && len(values) != 1 {
			return fmt.Errorf("%s operator for field '%s' must have exactly 1 value, got %d", opName, fieldName, len(values))
		}
	}

	// Validate field name patterns for common UUID fields
	if isLikelyUUIDField(fieldName) {
		if err := validateUUIDFieldValues(fieldName, condition); err != nil {
			return err
		}
	}

	return nil
}

// isLikelyUUIDField checks if a field name suggests it contains UUID values
func isLikelyUUIDField(fieldName string) bool {
	uuidPatterns := []string{"id", "_id", "uuid", "template_id", "organization_id", "user_id", "account_id"}
	fieldLower := strings.ToLower(fieldName)

	for _, pattern := range uuidPatterns {
		if strings.Contains(fieldLower, pattern) {
			return true
		}
	}

	return false
}

// validateUUIDFieldValues validates that values for UUID fields are valid UUIDs
func validateUUIDFieldValues(fieldName string, condition model.FilterCondition) error {
	allValues := [][]any{
		condition.Equals,
		condition.GreaterThan,
		condition.GreaterOrEqual,
		condition.LessThan,
		condition.LessOrEqual,
		condition.Between,
		condition.In,
		condition.NotIn,
	}

	for _, values := range allValues {
		for _, value := range values {
			if str, ok := value.(string); ok {
				if !isValidUUIDFormat(str) {
					return fmt.Errorf("field '%s' appears to be a UUID field but received non-UUID value '%s'. UUID fields require valid UUID format (e.g., '550e8400-e29b-41d4-a716-446655440000') or use a date field for date filtering", fieldName, str)
				}
			}
		}
	}

	return nil
}

// isValidUUIDFormat checks if a string is a valid UUID format
func isValidUUIDFormat(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// isDateField checks if a field name suggests it contains date/timestamp values
func isDateField(fieldName string) bool {
	datePatterns := []string{"created_at", "updated_at", "deleted_at", "completed_at", "date", "time", "_at", "_date", "_time"}
	fieldLower := strings.ToLower(fieldName)

	for _, pattern := range datePatterns {
		if strings.Contains(fieldLower, pattern) {
			return true
		}
	}

	return false
}

// isDateString checks if a value looks like a date string
func isDateString(value any) bool {
	if str, ok := value.(string); ok {
		// Check for common date formats: YYYY-MM-DD, YYYY-MM-DDTHH:MM:SS, etc.
		return len(str) >= 10 && strings.Contains(str, "-") && (len(str) == 10 || strings.Contains(str, "T"))
	}

	return false
}
