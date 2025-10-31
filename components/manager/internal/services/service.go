package services

import (
	"github.com/LerianStudio/reporter/v4/components/manager/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/v4/components/manager/internal/adapters/redis"
	pkgConfig "github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb/template"
	reportSeaweedFS "github.com/LerianStudio/reporter/v4/pkg/seaweedfs/report"
	templateSeaweedFS "github.com/LerianStudio/reporter/v4/pkg/seaweedfs/template"
)

// UseCase is a struct to implement the services methods
type UseCase struct {
	// TemplateRepo provides an abstraction on top of the template data source.
	TemplateRepo template.Repository

	// TemplateSeaweedFS is a repository interface for storing template files in SeaweedFS.
	TemplateSeaweedFS templateSeaweedFS.Repository

	// ReportRepo provides an abstraction on top of the report data source.
	ReportRepo report.Repository

	// ReportSeaweed is a repository interface for storing report files in SeaweedFS.
	ReportSeaweedFS reportSeaweedFS.Repository

	// RabbitMQRepo provides an abstraction on top of the producer rabbitmq.
	RabbitMQRepo rabbitmq.ProducerRepository

	// ExternalDataSources holds a map of external data sources identified by their names, each mapped to a DataSource object.
	ExternalDataSources map[string]pkgConfig.DataSource

	// RedisRepo provides an abstraction on top of the redis consumer.
	RedisRepo redis.RedisRepository
}
