package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	libCommons "github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/log"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"plugin-template-engine/pkg/postgres"
)

// Repository defines an interface for querying data from a specified table and fields.
//
//go:generate mockgen --destination=midaz.postgres.mock.go --package=postgres . Repository
type Repository interface {
	Query(ctx context.Context, organizationID uuid.UUID, table string, ledgers, fields []string) ([]map[string]any, error)
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

func (ds *MidazDataSource) Query(ctx context.Context, organizationID uuid.UUID, table string, ledgers, fields []string) ([]map[string]any, error) {
	logger := libCommons.NewLoggerFromContext(ctx)

	logger.Infof("Querying %s table with fields %s", table, fields)

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select(fields...).From(table)

	queryBuilder = applyOrganizationFilter(queryBuilder, table, organizationID)

	if len(ledgers) > 0 && table != "organization" {
		queryBuilder = applyLedgerFilter(queryBuilder, table, ledgers)
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		logger.Errorf("Error generating SQL: %s", err.Error())

		return nil, err
	}

	rows, err := ds.connection.ConnectionDB.Query(query, args...)
	if err != nil {
		logger.Errorf("Error executing query: %s", err.Error())

		return nil, err
	}
	defer rows.Close()

	return scanRows(rows, logger)
}

// applyOrganizationFilter adds the correct `WHERE` clause for organization filtering.
func applyOrganizationFilter(queryBuilder squirrel.SelectBuilder, table string, organizationID uuid.UUID) squirrel.SelectBuilder {
	if table == "organization" {
		return queryBuilder.Where("id = ?", organizationID.String())
	}

	return queryBuilder.Where("organization_id = ?", organizationID.String())
}

// applyLedgerFilter adds `WHERE` clauses for ledgers, depending on the table.
func applyLedgerFilter(queryBuilder squirrel.SelectBuilder, table string, ledgers []string) squirrel.SelectBuilder {
	args := toInterfaceSlice(ledgers)
	placeholder := squirrel.Placeholders(len(ledgers))

	if table == "ledger" {
		return queryBuilder.Where("id IN ("+placeholder+")", args...)
	}

	return queryBuilder.Where("ledger_id IN ("+placeholder+")", args...)
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
