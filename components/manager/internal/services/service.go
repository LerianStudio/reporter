package services

import (
	"plugin-template-engine/components/manager/internal/adapters/rabbitmq"
	pkgConfig "plugin-template-engine/pkg"
	reportMinio "plugin-template-engine/pkg/minio/report"
	templateMinio "plugin-template-engine/pkg/minio/template"
	"plugin-template-engine/pkg/mongodb/report"
	"plugin-template-engine/pkg/mongodb/template"
)

// UseCase is a struct to implement the services methods
type UseCase struct {
	// TemplateRepo provides an abstraction on top of the template data source.
	TemplateRepo template.Repository

	// TemplateMinio is a repository interface for storing template files in MinIO.
	TemplateMinio templateMinio.Repository

	// ReportRepo provides an abstraction on top of the report data source.
	ReportRepo report.Repository

	// ReportMinio is a repository interface for storing report files in MinIO.
	ReportMinio reportMinio.Repository

	// RabbitMQRepo provides an abstraction on top of the producer rabbitmq.
	RabbitMQRepo rabbitmq.ProducerRepository

	// ExternalDataSources holds a map of external data sources identified by their names, each mapped to a DataSource object.
	ExternalDataSources map[string]pkgConfig.DataSource
}
