package example

import (
	"database/sql"
	"k8s-golang-addons-boilerplate/pkg"
	exampleModel "k8s-golang-addons-boilerplate/pkg/example_model/model"
	"time"
)

// ExamplePostgreSQLModel represents the entity Example into SQL context in Database
type ExamplePostgreSQLModel struct {
	ID        string
	Name      string
	Age       int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
}

// FromEntity converts a request entity Account to AccountPostgreSQLModel
func (expm *ExamplePostgreSQLModel) FromEntity(example *exampleModel.Example) {
	*expm = ExamplePostgreSQLModel{
		ID:        pkg.GenerateUUIDv7().String(),
		Name:      example.Name,
		Age:       example.Age,
		CreatedAt: example.CreatedAt,
		UpdatedAt: example.UpdatedAt,
	}
}

func (expm *ExamplePostgreSQLModel) ToEntity() *exampleModel.ExampleOutput {
	return &exampleModel.ExampleOutput{
		ID:        expm.ID,
		Name:      expm.Name,
		Age:       expm.Age,
		CreatedAt: expm.CreatedAt,
		UpdatedAt: expm.UpdatedAt,
	}
}
