package query

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	example "k8s-golang-addons-boilerplate/mocks/postgres"
	"k8s-golang-addons-boilerplate/pkg/constant"
	"k8s-golang-addons-boilerplate/pkg/example_model/model"
	"testing"
)

func TestGetExampleByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExampleRepo := example.NewMockRepository(ctrl)

	exampleCase := &ExampleQuery{
		ExampleRepo: mockExampleRepo,
	}

	tests := []struct {
		name           string
		exampleId      uuid.UUID
		mockSetup      func()
		expectErr      bool
		expectedResult *model.ExampleOutput
	}{
		{
			name:      "Success - Get an example by id",
			exampleId: uuid.New(),
			mockSetup: func() {
				validUUID := uuid.New()
				mockExampleRepo.EXPECT().
					Find(gomock.Any(), gomock.Any()).
					Return(&model.ExampleOutput{ID: validUUID.String(), Name: "Test Example", Age: 12}, nil)
			},
			expectErr:      false,
			expectedResult: &model.ExampleOutput{ID: "valid-uuid", Name: "Test Example", Age: 12},
		},
		{
			name:      "Error - Get an example by id",
			exampleId: uuid.New(),
			mockSetup: func() {
				mockExampleRepo.EXPECT().
					Find(gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrBadRequest)
			},
			expectErr:      true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := exampleCase.GetExampleByID(ctx, tt.exampleId)

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
