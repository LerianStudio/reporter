package services

import (
	example "plugin-template-engine/internal/adapters/mongodb/templates"
)

// UseCase is a struct to implement the services methods
type UseCase struct {
	// TemplateRepo provides an abstraction on top of the template data source.
	TemplateRepo example.Repository
}
