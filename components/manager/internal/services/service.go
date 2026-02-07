// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"github.com/LerianStudio/reporter/components/manager/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/components/manager/internal/adapters/redis"
	pkgConfig "github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	reportSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/report"
	templateSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/template"
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

	// ExternalDataSources holds a thread-safe map of external data sources identified by their names.
	ExternalDataSources *pkgConfig.SafeDataSources

	// RedisRepo provides an abstraction on top of the redis consumer.
	RedisRepo redis.RedisRepository

	// RabbitMQExchange is the exchange name for publishing report generation messages.
	RabbitMQExchange string

	// RabbitMQGenerateReportKey is the routing key for report generation messages.
	RabbitMQGenerateReportKey string
}
