package services

import (
	"plugin-template-engine/components/worker/internal/adapters/minio/report"
	"plugin-template-engine/components/worker/internal/adapters/minio/template"
	"plugin-template-engine/components/worker/internal/adapters/postgres"
)

type DataSource struct {
	DatabaseType       string
	PostgresRepository postgres.Repository
}

type UseCase struct {
	TemplateFileRepo    template.Repository
	ReportFileRepo      report.Repository
	ExternalDataSources map[string]DataSource
}
