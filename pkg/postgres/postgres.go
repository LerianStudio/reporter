package postgres

import (
	"database/sql"
	"github.com/LerianStudio/lib-commons/commons/log"
	_ "github.com/jackc/pgx/v5/stdlib" // Registers the "pgx" driver with database/sql via init() – required for sql.Open("pgx", ...)
	"strings"
	"time"
)

// Connection is a hub which deals with postgres connections.
type Connection struct {
	ConnectionString   string
	DBName             string
	ConnectionDB       *sql.DB
	Connected          bool
	Logger             log.Logger
	MaxOpenConnections int
	MaxIdleConnections int
}

// Connect initializes the connection with the PostgreSQL DB.
func (c *Connection) Connect() error {
	c.Logger.Info("Connecting to PostgreSQL...")

	db, err := sql.Open("pgx", c.ConnectionString)
	if err != nil {
		c.Logger.Errorf("Error opening connection: %v", err)
		return err
	}

	if err := db.Ping(); err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			c.Logger.Errorf("Error closing connection: %v", closeErr)
		}

		c.Logger.Errorf("Error pinging PostgreSQL: %v", err)

		return err
	}

	db.SetMaxOpenConns(c.MaxOpenConnections)
	db.SetMaxIdleConns(c.MaxIdleConnections)
	db.SetConnMaxLifetime(10 * time.Minute)

	c.ConnectionDB = db
	c.Connected = true

	c.Logger.Infof("Connected to PostgreSQL [%s]", c.DBName)

	return nil
}

// GetDB returns a pointer to the postgres connection, initializing it if necessary.
func (pc *Connection) GetDB() (*sql.DB, error) {
	if pc.ConnectionDB == nil {
		if err := pc.Connect(); err != nil {
			pc.Logger.Infof("ERRCONECT %s", err)
			return nil, err
		}
	}

	return pc.ConnectionDB, nil
}

// ValidateFieldsInSchemaPostgres validate if all fields exist on postgres schema table
func ValidateFieldsInSchemaPostgres(expectedFields []string, schema TableSchema, countIfTableExist *int32) (missing []string) {
	columnSet := make(map[string]struct{}, len(schema.Columns))
	for _, col := range schema.Columns {
		columnSet[strings.ToLower(col.Name)] = struct{}{}
	}

	for _, field := range expectedFields {
		*countIfTableExist++ // variable to count if a table exists on a schema list

		if _, exists := columnSet[strings.ToLower(field)]; !exists {
			missing = append(missing, field)
		}
	}

	return
}
