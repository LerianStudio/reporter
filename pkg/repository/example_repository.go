package repository

import (
	"k8s-golang-addons-boilerplate/pkg/models"
)

//go:generate mockgen --destination=../../mocks/repository/example_mock.go --package=mock . ExampleRepository
type ExampleRepository interface {
	Create(input *models.ExampleInput) (*models.ExampleOutput, error)
	Update(id string, input *models.ExampleInput) error
	Get(id string) (*models.ExampleOutput, error)
	GetAll() ([]models.ExampleOutput, error)
	Delete(id string) error
}
