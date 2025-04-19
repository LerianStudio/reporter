package postgres

import (
	"context"
	"encoding/json"
	libCommons "github.com/LerianStudio/lib-commons/commons"
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

// NewRepository creates a new MidazDataSource instance using the provided postgres.Connection, initializing the database connection.
func NewRepository(pc *postgres.Connection) *MidazDataSource {
	c := &MidazDataSource{
		connection: pc,
	}

	_, err := c.connection.GetDB()
	if err != nil {
		panic(err)
	}

	return c
}

// Query retrieves data from a specified table and fields, returning a slice of maps with column names as keys.
func (ds *MidazDataSource) Query(ctx context.Context, organizationID uuid.UUID, table string, ledgers, fields []string) ([]map[string]any, error) {
	logger := libCommons.NewLoggerFromContext(ctx)

	logger.Infof("Querying %s table with fields %s", table, fields)

	// Use PostgreSQL-specific builder with $ placeholders
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	queryBuilder := psql.Select(fields...).From(table)

	// TODO: improve this
	if table == "organization" {
		queryBuilder = queryBuilder.Where("id = ?", organizationID.String())
	} else {
		queryBuilder = queryBuilder.Where("organization_id = ?", organizationID.String())
	}

	// Only add the WHERE clause if ledgers is not empty
	if len(ledgers) > 0 && table != "organization" {
		args := make([]any, len(ledgers))
		for i, v := range ledgers {
			args[i] = v
		}
		// TODO: create field name mapping between databases and tables?
		if table == "ledger" {
			queryBuilder = queryBuilder.Where("id IN ("+squirrel.Placeholders(len(ledgers))+")", args...)
		} else {
			queryBuilder = queryBuilder.Where("ledger_id IN ("+squirrel.Placeholders(len(ledgers))+")", args...)
		}
	}

	// Generate the SQL
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		logger.Errorf("Error generating SQL: %s", err.Error())

		return nil, err
	}

	rows, err := ds.connection.ConnectionDB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))

	for i := range vals {
		ptrs[i] = &vals[i]
	}

	result := make([]map[string]any, 0)

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]any)
		for i, col := range cols {
			rowMap[col] = vals[i]
		}

		result = append(result, rowMap)
	}

	return result, nil
}
