package template

import (
	"github.com/google/uuid"
)

// Template represents the entity model for a template
type Template struct {
	ID   uuid.UUID `json:"id" example:"00000000-0000-0000-0000-000000000000"`
	Name string    `json:"name" example:"Template name"`
	Age  int       `json:"age" example:"23"`
}

// TemplateMongoDBModel represents the MongoDB model for a template
type TemplateMongoDBModel struct {
	ID             uuid.UUID `bson:"_id"`
	Name           string    `bson:"name"`
	Age            int       `bson:"age"`
	OrganizationID uuid.UUID `bson:"organization_id"`
}

// ToEntity converts PackageMongoDBModel to Package
func (tm *TemplateMongoDBModel) ToEntity() *Template {
	return &Template{
		ID:   tm.ID,
		Name: tm.Name,
		Age:  tm.Age,
	}
}

// FromEntity converts Package to PackageMongoDBModel
func (tm *TemplateMongoDBModel) FromEntity(t *Template, organizationID uuid.UUID) error {
	tm.ID = t.ID
	tm.Name = t.Name
	tm.Age = t.Age
	tm.OrganizationID = organizationID

	return nil
}
