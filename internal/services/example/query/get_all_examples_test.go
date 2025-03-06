package query

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	example "k8s-golang-addons-boilerplate/mocks/postgres"
	"k8s-golang-addons-boilerplate/pkg/constant"
	"k8s-golang-addons-boilerplate/pkg/example_model/model"
	"k8s-golang-addons-boilerplate/pkg/net/http"
	"testing"
)

func TestGetAllExample(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExampleRepo := example.NewMockRepository(ctrl)

	filter := http.QueryHeader{
		Limit: 10,
		Page:  1,
	}

	exampleCase := &ExampleQuery{
		ExampleRepo: mockExampleRepo,
	}

	tests := []struct {
		name           string
		filter         http.QueryHeader
		mockSetup      func()
		expectErr      bool
		expectedResult []*model.ExampleOutput
	}{
		{
			name:   "Success - Get all examples",
			filter: filter,
			mockSetup: func() {
				validUUID := uuid.New()
				mockExampleRepo.EXPECT().
					FindAll(gomock.Any(), filter.ToOffsetPagination()).
					Return([]*model.ExampleOutput{
						{ID: validUUID.String(), Name: "Test Example", Age: 12},
					}, nil)
			},
			expectErr: false,
			expectedResult: []*model.ExampleOutput{
				{ID: "valid-uuid", Name: "Test Example", Age: 12},
			},
		},
		{
			name:   "Error - Get all examples",
			filter: filter,
			mockSetup: func() {
				mockExampleRepo.EXPECT().
					FindAll(gomock.Any(), filter.ToOffsetPagination()).
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
			result, err := exampleCase.GetAllExample(ctx, tt.filter)

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
