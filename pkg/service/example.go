package service

import (
	"k8s-golang-addons-boilerplate/pkg/models"

	"github.com/go-jose/go-jose/v4/json"
)

func (s *Service) CreateExample(input *models.ExampleInput) (*models.ExampleOutput, error) {
	output, err := s.exampleRepo.Create(input)
	if err != nil {
		return &models.ExampleOutput{}, err
	}

	cacheString, err := json.Marshal(output)
	if err != nil {
		return &models.ExampleOutput{}, err
	}

	s.cache.Set(output.ID, string(cacheString))

	return output, nil
}

func (s *Service) GetExample(id string) (*models.ExampleOutput, error) {
	cached := s.cache.Get(id)
	if cached == "" {
		// If cache is empty, fetch from database
		output, err := s.exampleRepo.Get(id)
		if err != nil {
			return &models.ExampleOutput{}, err
		}
		cacheString, err := json.Marshal(output)

		if err != nil {
			return &models.ExampleOutput{}, err
		}

		s.cache.Set(id, string(cacheString))

		return output, nil
	}

	var output = &models.ExampleOutput{}
	err := json.Unmarshal([]byte(cached), &output)
	if err != nil {
		return &models.ExampleOutput{}, err
	}

	return output, nil
}

func (s *Service) GetAllExample() ([]models.ExampleOutput, error) {
	data, err := s.exampleRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var output []models.ExampleOutput
	for _, d := range data {
		output = append(output, models.ExampleOutput{
			Name:      d.Name,
			Value:     d.Value,
			CreatedAt: d.CreatedAt,
		})
	}

	return output, nil
}

func (s *Service) UpdateExample(id string, input *models.ExampleInput) (*models.ExampleOutput, error) {
	err := s.exampleRepo.Update(id, input)
	if err != nil {
		return &models.ExampleOutput{}, err
	}

	output, err := s.exampleRepo.Get(id)
	if err != nil {
		return &models.ExampleOutput{}, err
	}

	cacheString, err := json.Marshal(output)
	if err != nil {
		return &models.ExampleOutput{}, err
	}

	s.cache.Set(output.ID, string(cacheString))

	return output, nil
}

func (s *Service) DeleteExample(id string) error {
	err := s.exampleRepo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}
