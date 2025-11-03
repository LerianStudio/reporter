package template

import (
	"time"

	"github.com/google/uuid"
)

// Template represents the entity model for a template
type Template struct {
	ID           uuid.UUID `json:"id" example:"00000000-0000-0000-0000-000000000000"`
	OutputFormat string    `json:"outputFormat" example:"HTML"`
	Description  string    `json:"description" example:"Template Financeiro"`
	FileName     string    `json:"fileName" example:"0196159b-4f26-7300-b3d9-f4f68a7c85f3_1744119295.tpl"`
	CreatedAt    time.Time `json:"createdAt" example:"2021-01-01T00:00:00Z"`
	UpdatedAt    time.Time `json:"updatedAt" example:"2021-01-01T00:00:00Z"`
}

// TemplateMongoDBModel represents the MongoDB model for a template
type TemplateMongoDBModel struct {
	ID             uuid.UUID                      `bson:"_id"`
	OrganizationID uuid.UUID                      `bson:"organization_id"`
	OutputFormat   string                         `bson:"output_format"`
	Description    string                         `bson:"description"`
	FileName       string                         `bson:"filename"`
	MappedFields   map[string]map[string][]string `bson:"mapped_fields"`
	CreatedAt      time.Time                      `bson:"created_at"`
	UpdatedAt      time.Time                      `bson:"updated_at"`
	DeletedAt      *time.Time                     `bson:"deleted_at"`
}

// ToEntity converts TemplateMongoDBModel to Template
func (tm *TemplateMongoDBModel) ToEntity() *Template {
	return &Template{
		ID:           tm.ID,
		OutputFormat: tm.OutputFormat,
		Description:  tm.Description,
		FileName:     tm.FileName,
		CreatedAt:    tm.CreatedAt,
		UpdatedAt:    tm.UpdatedAt,
	}
}
