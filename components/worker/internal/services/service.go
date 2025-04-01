package services

import (
	"plugin-template-engine/components/worker/internal/adapters/minio/report"
	"plugin-template-engine/components/worker/internal/adapters/minio/template"
)

type UseCase struct {
	TemplateFileRepo template.Repository
	ReportFileRepo   report.Repository
}
