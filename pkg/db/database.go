package db

//go:generate mockgen --destination=../../mocks/db/database_mock.go --package=mock . Database
type Database interface {
	Connect() error
	GetDB() (DBConn, error)
}
