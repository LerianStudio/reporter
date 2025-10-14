package services

import (
	"plugin-smart-templates/v3/pkg"
	reportData "plugin-smart-templates/v3/pkg/mongodb/report"
	reportSeaweedFS "plugin-smart-templates/v3/pkg/seaweedfs/report"
	templateSeaweedFS "plugin-smart-templates/v3/pkg/seaweedfs/template"
)

// UseCase is a struct that coordinates the handling of template files, report storage, external data sources, and report data.
type UseCase struct {
	// TemplateSeaweedFS is a repository used to retrieve template files from SeaweedFS storage.
	TemplateSeaweedFS templateSeaweedFS.Repository

	// ReportSeaweedFS is a repository interface for storing report files in SeaweedFS.
	ReportSeaweedFS reportSeaweedFS.Repository

	// ExternalDataSources holds a map of external data sources identified by their names, each mapped to a DataSource object.
	ExternalDataSources map[string]pkg.DataSource

	// ReportDataRepo is an interface for operations related to report data storage used in the reporting use case
	ReportDataRepo reportData.Repository

	// CircuitBreakerManager manages circuit breakers for external datasources
	CircuitBreakerManager *pkg.CircuitBreakerManager

	// HealthChecker performs periodic health checks and reconnection attempts
	HealthChecker *pkg.HealthChecker

	// ReportTTL defines the Time To Live for reports (e.g., "1m", "1h", "7d", "30d"). Empty means no TTL.
	ReportTTL string
}
