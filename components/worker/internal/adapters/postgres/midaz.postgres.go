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
//go:generate mockgen --destination=midaz.postgres.mock.go --package=postgres . Repository
type Repository interface {
	Query(ctx context.Context, table string, fields []string, filter map[string][]string) ([]map[string]any, error)
}

// MidazDataSource provides an interface for interacting with a PostgreSQL database connection.
type MidazDataSource struct {
	connection *postgres.Connection
}

// NewMidazRepository creates a new MidazDataSource instance using the provided postgres.Connection, initializing the database connection.
func NewMidazRepository(pc *postgres.Connection) *MidazDataSource {
	c := &MidazDataSource{
		connection: pc,
	}

	_, err := c.connection.GetDB()
	if err != nil {
		panic(err)
	}

	return c
}

func (ds *MidazDataSource) Query(ctx context.Context, table string, fields []string, filter map[string][]string) ([]map[string]any, error) {
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

// applyIdFilter adiciona condições WHERE genéricas para qualquer tabela ou campo.
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
		if column == "address" {
			rowMap[column] = parseAddress(values[i], logger)
		} else {
			rowMap[column] = values[i]
		}
	}

	return rowMap
}

// parseAddress unmarshals the address column if it contains JSON.
func parseAddress(value any, logger log.Logger) any {
	if value == nil {
		return nil
	}

	if byteData, ok := value.([]uint8); ok {
		addressMap := make(map[string]string)
		if err := json.Unmarshal(byteData, &addressMap); err == nil {
			return addressMap
		}

		logger.Warnf("Failed to unmarshal address.")
	}

	return value
}
