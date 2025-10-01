package services

import (
	"context"
	"plugin-smart-templates/v3/pkg/constant"
	"plugin-smart-templates/v3/pkg/mongodb/template"
	httpUtils "plugin-smart-templates/v3/pkg/net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_getAllTemplates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tempID := uuid.New()
	orgId := uuid.New()
	resultEntity := []*template.Template{
		{
			ID:           tempID,
			Description:  "Template Financeiro",
			OutputFormat: "html",
			FileName:     "019672b1-9d50-7360-9df5-5099dd166709_1745680964.tpl",
		},
	}

	mockTempRepo := template.NewMockRepository(ctrl)

	filter := httpUtils.QueryHeader{
		Limit:          10,
		Page:           1,
		OrganizationID: orgId,
	}

	tempSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	tests := []struct {
		name           string
		filter         httpUtils.QueryHeader
		mockSetup      func()
		expectErr      bool
		expectedResult []*template.Template
	}{
		{
			name:   "Success - Get all templates",
			filter: filter,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindList(gomock.Any(), filter).
					Return(resultEntity, nil)
			},
			expectErr: false,
			expectedResult: []*template.Template{
				{
					ID:           tempID,
					Description:  "Template Financeiro",
					OutputFormat: "html",
					FileName:     "019672b1-9d50-7360-9df5-5099dd166709_1745680964.tpl",
				},
			},
		},
		{
			name:   "Error - Get all templates",
			filter: filter,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindList(gomock.Any(), filter).
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
			result, err := tempSvc.GetAllTemplates(ctx, tt.filter, orgId)

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
