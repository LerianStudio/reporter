package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/attribute"
)

// Repository defines an interface for querying data from a specified table and fields.
//
//go:generate mockgen --destination=datasource.postgres.mock.go --package=postgres . Repository
type Repository interface {
	Query(ctx context.Context, schema []TableSchema, table string, fields []string, filter map[string][]any) ([]map[string]any, error)
	GetDatabaseSchema(ctx context.Context) ([]TableSchema, error)
	CloseConnection() error
}

// TableSchema represents the structure of a database table
type TableSchema struct {
	TableName string              `json:"table_name"`
	Columns   []ColumnInformation `json:"columns"`
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

// NewDataSourceRepository creates a new ExternalDataSource instance using the provided postgres.Connection, initializing the database connection.
func NewDataSourceRepository(pc *Connection) *ExternalDataSource {
	c := &ExternalDataSource{
		connection: pc,
	}

	_, err := c.connection.GetDB()
	if err != nil {
		panic(err)
	}

	return c
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
// It returns the query results as a slice of maps or an error in case of failure.
func (ds *ExternalDataSource) Query(ctx context.Context, schema []TableSchema, table string, fields []string, filter map[string][]any) ([]map[string]any, error) {
	logger := libCommons.NewLoggerFromContext(ctx)
	tracer := libCommons.NewTracerFromContext(ctx)
	reqId := libCommons.NewHeaderIDFromContext(ctx)

	_, span := tracer.Start(ctx, "postgres.data_source.query")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.table", table),
		attribute.StringSlice("app.request.fields", fields),
	)

	logger.Infof("Querying %s table with fields %v", table, fields)

	// Validate requested table and fields
	queriedFields, err := ds.ValidateTableAndFields(ctx, table, fields, schema)
	if err != nil {
		return nil, err
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select(queriedFields...).From(table)

	// Apply filters, but only if they correspond to valid columns
	queryBuilder = buildDynamicFilters(queryBuilder, schema, table, filter)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error generating SQL: %w", err)
	}

	logger.Infof("Executing SQL: %s with args: %v", query, args)

	rows, err := ds.connection.ConnectionDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	return scanRows(rows, logger)
}

// GetDatabaseSchema retrieves all tables and their column details from the database
// It returns a slice of TableSchema objects or an error if the operation fails
func (ds *ExternalDataSource) GetDatabaseSchema(ctx context.Context) ([]TableSchema, error) {
	logger := libCommons.NewLoggerFromContext(ctx)
	tracer := libCommons.NewTracerFromContext(ctx)
	reqId := libCommons.NewHeaderIDFromContext(ctx)

	_, span := tracer.Start(ctx, "postgres.data_source.get_database_schema")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	logger.Info("Retrieving database schema information")

	// Query to get all user tables in the database
	tableQuery := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := ds.connection.ConnectionDB.Query(tableQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying tables: %w", err)
	}
	defer rows.Close()

	var tables []string

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %w", err)
		}

		tables = append(tables, tableName)
	}

	// Query to get primary key information
	pkQuery := `
		SELECT tc.table_name, kc.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kc 
			ON kc.table_name = tc.table_name 
			AND kc.table_schema = tc.table_schema
			AND kc.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'PRIMARY KEY'
		AND tc.table_schema = 'public'
	`

	pkRows, err := ds.connection.ConnectionDB.Query(pkQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying primary keys: %w", err)
	}
	defer pkRows.Close()

	// Map to store primary key columns by table name
	primaryKeys := make(map[string]map[string]bool)

	for pkRows.Next() {
		var tableName, columnName string
		if err := pkRows.Scan(&tableName, &columnName); err != nil {
			return nil, fmt.Errorf("error scanning primary key info: %w", err)
		}
	}

	// Build the complete schema information
	schema := make([]TableSchema, 0, len(tables))

	for _, tableName := range tables {
		// Query to get column information for the current table
		columnQuery := `
			SELECT column_name, data_type, 
			       CASE WHEN is_nullable = 'YES' THEN true ELSE false END as is_nullable
			FROM information_schema.columns
			WHERE table_schema = 'public'
			AND table_name = $1
			ORDER BY ordinal_position
		`

		colRows, err := ds.connection.ConnectionDB.Query(columnQuery, tableName)
		if err != nil {
			return nil, fmt.Errorf("error querying columns for table %s: %w", tableName, err)
		}

		var columns []ColumnInformation

		for colRows.Next() {
			var col ColumnInformation
			if err := colRows.Scan(&col.Name, &col.DataType, &col.IsNullable); err != nil {
				if closeErr := colRows.Close(); closeErr != nil {
					logger.Warnf("error closing rows after scan error: %v", closeErr)
				}

				return nil, fmt.Errorf("error scanning column info: %w", err)
			}

			// Check if this column is a primary key
			if pkCols, exists := primaryKeys[tableName]; exists {
				col.IsPrimaryKey = pkCols[col.Name]
			}

			columns = append(columns, col)
		}

		if err := colRows.Close(); err != nil {
			logger.Warnf("error closing column rows: %v", err)
		}

		schema = append(schema, TableSchema{
			TableName: tableName,
			Columns:   columns,
		})
	}

	logger.Infof("Retrieved schema for %d tables", len(schema))

	return schema, nil
}

// scanRows processes the query rows and creates the resulting slice of maps.
func scanRows(rows *sql.Rows, logger log.Logger) ([]map[string]any, error) {
	columns, _ := rows.Columns()
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

// ValidateTableAndFields checks if the specified table exists and validates that
// all requested fields exist in that table.
// It returns a list of valid fields and an error if the table doesn't exist or fields are invalid.
func (ds *ExternalDataSource) ValidateTableAndFields(ctx context.Context, tableName string, requestedFields []string, schema []TableSchema) ([]string, error) {
	logger := libCommons.NewLoggerFromContext(ctx)
	tracer := libCommons.NewTracerFromContext(ctx)
	reqId := libCommons.NewHeaderIDFromContext(ctx)

	_, span := tracer.Start(ctx, "postgres.data_source.validate_table_and_fields")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.table", tableName),
		attribute.StringSlice("app.request.fields", requestedFields),
	)

	err := libOpentelemetry.SetSpanAttributesFromStructWithObfuscation(&span, "app.request.data_source.schema", schema)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert schema to JSON string", err)
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
		if validColumns[field] {
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
