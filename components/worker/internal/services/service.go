package services

import (
	"github.com/LerianStudio/reporter/v4/pkg"
	reportData "github.com/LerianStudio/reporter/v4/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/v4/pkg/pdf"
	"github.com/LerianStudio/reporter/v4/pkg/storage"
)

// UseCase is a struct that coordinates the handling of template files, report storage, external data sources, and report data.
type UseCase struct {
	// TemplateStorage is a repository used to retrieve template files from storage (S3, SeaweedFS, etc).
	TemplateStorage storage.TemplateRepository

	// ReportStorage is a repository interface for storing report files (S3, SeaweedFS, etc).
	ReportStorage storage.ReportRepository

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

	// PdfPool provides PDF generation capabilities using Chrome headless
	PdfPool *pdf.WorkerPool
}
