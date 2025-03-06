package model

import (
	"time"
)

type ExampleInput struct {
	Name  string `json:"name" validate:"required"`
	Value string `json:"value" validate:"required"`
}

// CreateExampleInput is a struct design to encapsulate payload data.
//
// swagger:model CreateExampleInput
// @Description CreateExampleInput is the input payload to create an example_model.
type CreateExampleInput struct {
	Name string `json:"name" validate:"required,max=256"`
	Age  int    `json:"age" validate:"required"`
} // @name CreateExampleInput

// ExampleOutput is a struct design to output payload data.
//
// swagger:model ExampleOutput
// @Description ExampleOutput is the output payload.
type ExampleOutput struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateExampleInput is a struct design to encapsulate request update payload data.
//
// // swagger:model UpdateExampleInput
// @Description UpdateExampleInput is the input payload to update an example.
type UpdateExampleInput struct {
	Name string `json:"name" validate:"max=256" example:"Example test"`
	Age  int    `json:"age" example:"12"`
} // @name UpdateExampleInput

// Example structure for marshaling/unmarshalling JSON.
//
// swagger:model Example
// @Description Example is a struct designed to store example_model data.
type Example struct {
	Name      string    `json:"name" example:"example name"`
	Age       int       `json:"age" example:"18"`
	CreatedAt time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}
