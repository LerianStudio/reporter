package services

import (
	"plugin-template-engine/components/manager/internal/adapters/mongodb/template"
)

// UseCase is a struct to implement the services methods
type UseCase struct {
	// TemplateRepo provides an abstraction on top of the template data source.
	TemplateRepo template.Repository
}
