package postgres

import (
	"context"
	"github.com/Masterminds/squirrel"
	"plugin-template-engine/pkg/postgres"
)

type Repository interface {
	Query(ctx context.Context, table string, fields []string) ([]map[string]interface{}, error)
}

type DataSource struct {
	connection *postgres.Connection
}

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

func (ds *DataSource) Query(ctx context.Context, table string, fields []string) ([]map[string]interface{}, error) {
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
	vals := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))

	for i := range vals {
		ptrs[i] = &vals[i]
	}

	result := make([]map[string]interface{}, 0)

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, col := range cols {
			rowMap[col] = vals[i]
		}

		result = append(result, rowMap)
	}

	return result, nil
}
