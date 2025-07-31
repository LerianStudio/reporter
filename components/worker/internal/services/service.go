package services

import (
	pkgConfig "plugin-smart-templates/v2/pkg"
	reportFile "plugin-smart-templates/v2/pkg/minio/report"
	templateFile "plugin-smart-templates/v2/pkg/minio/template"
	reportData "plugin-smart-templates/v2/pkg/mongodb/report"
)

// UseCase is a struct that coordinates the handling of template files, report storage, external data sources, and report data.
type UseCase struct {
	// TemplateFileRepo is a repository used to retrieve template files from MinIO storage.
	TemplateFileRepo templateFile.Repository

	// ReportFileRepo is a repository interface for storing report files in MinIO.
	ReportFileRepo reportFile.Repository

	// ExternalDataSources holds a map of external data sources identified by their names, each mapped to a DataSource object.
	ExternalDataSources map[string]pkgConfig.DataSource

	// ReportDataRepo is an interface for operations related to report data storage used in the reporting use case.
	ReportDataRepo reportData.Repository
}
