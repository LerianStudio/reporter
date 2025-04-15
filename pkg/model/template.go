package model

// CreateTemplateInput is a struct designed to encapsulate request create payload data.
//
// swagger:model CreateTemplateInput
// @Description CreateTemplateInput is the input payload to create a template.
type CreateTemplateInput struct {
	Name string `json:"name" validate:"required" example:"Template name"`
	Age  int    `json:"age" example:"12"`
} // @name CreateTemplateInput
