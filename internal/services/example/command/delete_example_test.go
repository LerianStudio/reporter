package query

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"k8s-golang-addons-boilerplate/internal/services"
	example "k8s-golang-addons-boilerplate/mocks/postgres"
	"testing"
)

func TestDeleteExample(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExampleRepo := example.NewMockRepository(ctrl)

	exampleCase := &ExampleCommand{
		ExampleRepo: mockExampleRepo,
	}

	tests := []struct {
		name           string
		exampleID      uuid.UUID
		mockSetup      func()
		expectErr      bool
		expectedResult error
	}{
		{
			name:      "Success - Delete example",
			exampleID: uuid.New(),
			mockSetup: func() {
				mockExampleRepo.EXPECT().
					Delete(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectErr:      false,
			expectedResult: nil,
		},
		{
			name:      "Error - Create an example",
			exampleID: uuid.New(),
			mockSetup: func() {
				mockExampleRepo.EXPECT().
					Delete(gomock.Any(), gomock.Any()).
					Return(services.ErrDatabaseItemNotFound)
			},
			expectErr:      true,
			expectedResult: services.ErrDatabaseItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			err := exampleCase.DeleteExampleByID(ctx, tt.exampleID)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
