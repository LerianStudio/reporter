package services

import (
	pkgConfig "plugin-smart-templates/v2/pkg"
	reportData "plugin-smart-templates/v2/pkg/mongodb/report"
	reportSeaweedFS "plugin-smart-templates/v2/pkg/seaweedfs/report"
	templateSeaweedFS "plugin-smart-templates/v2/pkg/seaweedfs/template"
)

// UseCase is a struct that coordinates the handling of template files, report storage, external data sources, and report data.
type UseCase struct {
	// TemplateSeaweedFS is a repository interface for storing template files in SeaweedFS.
	TemplateSeaweedFS templateSeaweedFS.Repository

	// ReportSeaweed is a repository interface for storing report files in SeaweedFS.
	ReportSeaweedFS reportSeaweedFS.Repository

	// ExternalDataSources holds a map of external data sources identified by their names, each mapped to a DataSource object.
	ExternalDataSources map[string]pkgConfig.DataSource

	// ReportDataRepo is an interface for operations related to report data storage used in the reporting use case.
	ReportDataRepo reportData.Repository
}
