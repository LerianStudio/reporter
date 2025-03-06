package models

import (
	"time"

	"github.com/google/uuid"
)

type ExampleInput struct {
	Name  string `json:"name" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type ExampleOutput struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

type ExamplePostgreSQLModel struct {
	ID        string
	Name      string
	Value     string
	CreatedAt time.Time
	DeletedAt time.Time
}

func (e *ExamplePostgreSQLModel) ToEntity() *ExampleOutput {
	return &ExampleOutput{
		ID:        e.ID,
		Name:      e.Name,
		Value:     e.Value,
		CreatedAt: e.CreatedAt,
	}
}

func (e *ExamplePostgreSQLModel) FromEntity(input *ExampleInput) {
	*e = ExamplePostgreSQLModel{
		ID:        uuid.New().String(),
		Name:      input.Name,
		Value:     input.Value,
		CreatedAt: time.Now(),
	}
}
