package services

import (
	"plugin-template-engine/components/manager/internal/adapters/rabbitmq"
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

	// RabbitMQRepo provides an abstraction on top of the producer rabbitmq.
	RabbitMQRepo rabbitmq.ProducerRepository
}
