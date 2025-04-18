package services

import (
	reportFile "plugin-template-engine/components/worker/internal/adapters/minio/report"
	templateFile "plugin-template-engine/components/worker/internal/adapters/minio/template"
	reportData "plugin-template-engine/components/worker/internal/adapters/mongodb/report"
	"plugin-template-engine/components/worker/internal/adapters/postgres"
	pg "plugin-template-engine/pkg/postgres"
)

// DataSource represents a configuration for an external data source, specifying the database type and repository used.
type DataSource struct {
	// DatabaseType specifies the type of database being used, such as "postgres" or "mongodb".
	DatabaseType string

	// PostgresRepository is an interface for querying PostgreSQL tables and fields in an external data source.
	PostgresRepository postgres.Repository

	// DatabaseConfig holds the configuration needed to establish a connection
	DatabaseConfig *pg.Connection

	// Connection holds the actual database connection that can be closed
	Connection *pg.Connection

	// Initialized indicates if the connection has been established
	Initialized bool
}

// UseCase is a struct that coordinates the handling of template files, report storage, external data sources, and report data.
type UseCase struct {
	// TemplateFileRepo is a repository used to retrieve template files from MinIO storage.
	TemplateFileRepo templateFile.Repository

	// ReportFileRepo is a repository interface for storing report files in MinIO.
	ReportFileRepo reportFile.Repository

	// ExternalDataSources holds a map of external data sources identified by their names, each mapped to a DataSource object.
	ExternalDataSources map[string]DataSource

	// ReportDataRepo is an interface for operations related to report data storage used in the reporting use case.
	ReportDataRepo reportData.Repository
}
