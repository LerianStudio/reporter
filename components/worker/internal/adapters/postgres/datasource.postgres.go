package postgres

import (
	"context"
	"github.com/Masterminds/squirrel"
	"plugin-template-engine/pkg/postgres"
)

// Repository defines an interface for querying data from a specified table and fields.
//
//go:generate mockgen --destination=datasource.postgres.mock.go --package=postgres . Repository
type Repository interface {
	Query(ctx context.Context, table string, fields []string) ([]map[string]any, error)
}

// DataSource provides an interface for interacting with a PostgreSQL database connection.
type DataSource struct {
	connection *postgres.Connection
}

// NewRepository creates a new DataSource instance using the provided postgres.Connection, initializing the database connection.
func NewRepository(pc *postgres.Connection) *DataSource {
	c := &DataSource{
		connection: pc,
	}

	_, err := c.connection.GetDB()
	if err != nil {
		panic(err)
	}

	return c
}

// Query retrieves data from a specified table and fields, returning a slice of maps with column names as keys.
func (ds *DataSource) Query(ctx context.Context, table string, fields []string) ([]map[string]any, error) {
	query, args, err := squirrel.Select(fields...).From(table).ToSql()
	if err != nil {
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
