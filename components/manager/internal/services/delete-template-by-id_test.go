package services

import (
	"context"
	"plugin-smart-templates/v3/pkg/constant"
	"plugin-smart-templates/v3/pkg/mongodb/template"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"
)

func Test_deleteTemplateByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	tempID := uuid.New()
	orgId := uuid.New()
	tempSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	tests := []struct {
		name           string
		tempID         uuid.UUID
		orgId          uuid.UUID
		hardDelete     bool
		mockSetup      func()
		expectErr      bool
		expectedResult error
	}{
		{
			name:       "Success - Delete a template",
			tempID:     tempID,
			orgId:      orgId,
			hardDelete: true,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectErr:      false,
			expectedResult: nil,
		},
		{
			name:       "Error Bad Request - Delete a template",
			tempID:     tempID,
			orgId:      orgId,
			hardDelete: true,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(constant.ErrBadRequest)
			},
			expectErr:      true,
			expectedResult: constant.ErrBadRequest,
		},
		{
			name:       "Error Document Not found - Delete a template",
			tempID:     tempID,
			orgId:      orgId,
			hardDelete: true,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mongo.ErrNoDocuments)
			},
			expectErr:      true,
			expectedResult: mongo.ErrNoDocuments,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			err := tempSvc.DeleteTemplateByID(ctx, tt.tempID, tt.orgId, tt.hardDelete)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
