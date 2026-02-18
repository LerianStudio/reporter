// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"github.com/LerianStudio/reporter/pkg"
	reportData "github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/pdf"
	reportSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/report"
	templateSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/template"
)

// UseCase is a struct that coordinates the handling of template files, report storage, external data sources, and report data.
type UseCase struct {
	// TemplateSeaweedFS is a repository used to retrieve template files from SeaweedFS storage.
	TemplateSeaweedFS templateSeaweedFS.Repository

	// ReportSeaweedFS is a repository interface for storing report files in SeaweedFS.
	ReportSeaweedFS reportSeaweedFS.Repository

	// ExternalDataSources holds a thread-safe map of external data sources identified by their names.
	ExternalDataSources *pkg.SafeDataSources

	// ReportDataRepo is an interface for operations related to report data storage used in the reporting use case
	ReportDataRepo reportData.Repository

	// CircuitBreakerManager manages circuit breakers for external datasources
	CircuitBreakerManager pkg.CircuitBreakerExecutor

	// HealthChecker performs periodic health checks and reconnection attempts
	HealthChecker pkg.HealthCheckRunner

	// ReportTTL defines the Time To Live for reports (e.g., "1m", "1h", "7d", "30d"). Empty means no TTL.
	ReportTTL string

	// PdfPool provides PDF generation capabilities using Chrome headless
	PdfPool pdf.PDFGenerator

	// CryptoHashSecretKeyPluginCRM is the hash secret key for plugin_crm data operations.
	CryptoHashSecretKeyPluginCRM string

	// CryptoEncryptSecretKeyPluginCRM is the encryption secret key for plugin_crm data operations.
	CryptoEncryptSecretKeyPluginCRM string
}
