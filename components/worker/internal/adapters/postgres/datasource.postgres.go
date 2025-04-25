package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	libCommons "github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/log"
	"github.com/Masterminds/squirrel"
	"plugin-template-engine/pkg/postgres"
	"strings"
)

var filterableIDs = map[string][]string{
	"organization": {"organization_id", "parent_organization_id"},
	"ledger":       {"organization_id", "ledger_id"},
	"asset":        {"organization_id", "ledger_id", "asset_id"},
	"account":      {"organization_id", "ledger_id", "account_id", "parent_account_id", "portfolio_id", "segment_id"},
	"portfolio":    {"organization_id", "ledger_id", "portfolio_id"},
	"segment":      {"organization_id", "ledger_id", "segment_id"},
	"transaction":  {"organization_id", "ledger_id", "transaction_id", "parent_transaction_id"},
	"operation":    {"organization_id", "ledger_id", "transaction_id", "operation_id", "account_id", "balance_id"},
	"asset_rate":   {"organization_id", "ledger_id"},
	"balance":      {"organization_id", "ledger_id", "account_id", "balance_id"},
}

// Repository defines an interface for querying data from a specified table and fields.
//
//go:generate mockgen --destination=datasource.postgres.mock.go --package=postgres . Repository
type Repository interface {
	Query(ctx context.Context, table string, fields []string, filter map[string][]string) ([]map[string]any, error)
}

// ExternalDataSource provides an interface for interacting with a PostgreSQL database connection.
type ExternalDataSource struct {
	connection *postgres.Connection
}

// NewDataSourceRepository creates a new ExternalDataSource instance using the provided postgres.Connection, initializing the database connection.
func NewDataSourceRepository(pc *postgres.Connection) *ExternalDataSource {
	c := &ExternalDataSource{
		connection: pc,
	}

	_, err := c.connection.GetDB()
	if err != nil {
		panic(err)
	}

	return c
}

// Query executes a SELECT SQL query on the specified table with the given fields and filter criteria.
// It returns the query results as a slice of maps or an error in case of failure.
func (ds *ExternalDataSource) Query(ctx context.Context, table string, fields []string, filter map[string][]string) ([]map[string]any, error) {
	logger := libCommons.NewLoggerFromContext(ctx)
	logger.Infof("Querying %s table with fields %v", table, fields)

	filterableFields, isTableValid := filterableIDs[table]
	if !isTableValid {
		return nil, fmt.Errorf("invalid table: %s", table)
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select(fields...).From(table)

	queryBuilder = buildQueryWithFilters(queryBuilder, filter, filterableFields, table)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error generating SQL: %w", err)
	}

	rows, err := ds.connection.ConnectionDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	return scanRows(rows, logger)
}

// buildQueryWithFilters applies filter criteria to the query builder based on the filterable fields.
func buildQueryWithFilters(queryBuilder squirrel.SelectBuilder, filter map[string][]string, filterableFields []string, table string) squirrel.SelectBuilder {
	for _, field := range filterableFields {
		if ids, exists := filter[field]; exists {
			queryBuilder = applyIdFilter(queryBuilder, field, table, ids)
		}
	}

	return queryBuilder
}

// applyIdFilter adds generic WHERE conditions for any table or field.
func applyIdFilter(queryBuilder squirrel.SelectBuilder, fieldName string, table string, ids []string) squirrel.SelectBuilder {
	if len(ids) == 0 {
		return queryBuilder
	}

	args := toInterfaceSlice(ids)
	placeholder := squirrel.Placeholders(len(ids))

	if strings.HasPrefix(fieldName, table) {
		return queryBuilder.Where("id IN ("+placeholder+")", args...)
	}

	return queryBuilder.Where(fieldName+" IN ("+placeholder+")", args...)
}

// toInterfaceSlice converts a slice of strings to a slice of interface{} for placeholders.
func toInterfaceSlice(input []string) []any {
	result := make([]any, len(input))
	for i, v := range input {
		result[i] = v
	}

	return result
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
