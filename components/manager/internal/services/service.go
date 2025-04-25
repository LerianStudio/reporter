package services

import (
	templateMinio "plugin-template-engine/pkg/minio/template"
	"plugin-template-engine/pkg/mongodb/template"
)

// UseCase is a struct to implement the services methods
type UseCase struct {
	// TemplateRepo provides an abstraction on top of the template data source.
	TemplateRepo template.Repository

	// TemplateMinio is a repository interface for storing template files in MinIO.
	TemplateMinio templateMinio.Repository
}
