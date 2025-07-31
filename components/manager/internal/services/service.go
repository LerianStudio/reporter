package services

import (
	"plugin-smart-templates/v2/components/manager/internal/adapters/rabbitmq"
	"plugin-smart-templates/v2/components/manager/internal/adapters/redis"
	pkgConfig "plugin-smart-templates/v2/pkg"
	reportMinio "plugin-smart-templates/v2/pkg/minio/report"
	templateMinio "plugin-smart-templates/v2/pkg/minio/template"
	"plugin-smart-templates/v2/pkg/mongodb/report"
	"plugin-smart-templates/v2/pkg/mongodb/template"
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

	// RedisRepo provides an abstraction on top of the redis consumer.
	RedisRepo redis.RedisRepository
}
