package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"
	"plugin-template-engine/pkg/constant"
	"plugin-template-engine/pkg/mongodb/template"
	"testing"
	"time"
)

func Test_getTemplateById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	tempId := uuid.New()
	orgId := uuid.New()

	tempSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	timestamp := time.Now().Unix()
	templateEntity := &template.Template{
		ID:           tempId,
		OutputFormat: "xml",
		Description:  "Template Financeiro",
		FileName:     fmt.Sprintf("%s_%d.tpl", tempId.String(), timestamp),
		CreatedAt:    time.Time{},
	}

	tests := []struct {
		name           string
		orgId          uuid.UUID
		tempId         uuid.UUID
		mockSetup      func()
		expectErr      bool
		expectedResult *template.Template
	}{
		{
			name:   "Success - Get a template by id",
			orgId:  orgId,
			tempId: tempId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(templateEntity, nil)
			},
			expectErr: false,
			expectedResult: &template.Template{
				ID:           tempId,
				OutputFormat: "xml",
				Description:  "Template Financeiro",
				FileName:     fmt.Sprintf("%s_%d.tpl", tempId.String(), timestamp),
				CreatedAt:    time.Time{},
			},
		},
		{
			name:   "Error - Get a template by id",
			orgId:  orgId,
			tempId: tempId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			expectedResult: nil,
		},
		{
			name:   "Error - Get a template by id not found",
			orgId:  orgId,
			tempId: tempId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, mongo.ErrNoDocuments)
			},
			expectErr:      true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := tempSvc.GetTemplateByID(ctx, tt.tempId, tt.orgId)

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
