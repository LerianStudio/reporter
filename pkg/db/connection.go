package db

import "database/sql"

//go:generate mockgen --destination=../../mocks/db/connection_mock.go --package=mock . DBConn
type DBConn interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}
