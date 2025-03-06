package query

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"k8s-golang-addons-boilerplate/internal/services"
	example "k8s-golang-addons-boilerplate/mocks/postgres"
	"k8s-golang-addons-boilerplate/pkg/example_model/model"
	"testing"
)

func TestUpdateExample(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExampleRepo := example.NewMockRepository(ctrl)

	exampleToUpdate := &model.UpdateExampleInput{
		Name: "Update test example",
		Age:  22,
	}

	exampleCase := &ExampleCommand{
		ExampleRepo: mockExampleRepo,
	}

	tests := []struct {
		name           string
		exampleId      uuid.UUID
		exampleInput   *model.UpdateExampleInput
		mockSetup      func()
		expectErr      bool
		expectedResult *model.ExampleOutput
	}{
		{
			name:         "Success - Update example by id",
			exampleId:    uuid.New(),
			exampleInput: exampleToUpdate,
			mockSetup: func() {
				validUUID := uuid.New()
				mockExampleRepo.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&model.ExampleOutput{ID: validUUID.String(), Name: "Update test Example", Age: 22}, nil)
			},
			expectErr:      false,
			expectedResult: &model.ExampleOutput{ID: "valid-uuid", Name: "Update test Example", Age: 22},
		},
		{
			name:         "Error - Update example by id",
			exampleId:    uuid.New(),
			exampleInput: exampleToUpdate,
			mockSetup: func() {
				mockExampleRepo.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, services.ErrDatabaseItemNotFound)
			},
			expectErr:      true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := exampleCase.UpdateExampleByID(ctx, tt.exampleId, tt.exampleInput)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
