// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"testing"

	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	httpUtils "github.com/LerianStudio/reporter/pkg/net/http"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAllTemplates(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tempID := uuid.New()
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
		Limit: 10,
		Page:  1,
	}

	tempSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	tests := []struct {
		name           string
		filter         httpUtils.QueryHeader
		mockSetup      func()
		expectErr      bool
		expectedErr    error
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
			expectedErr:    constant.ErrBadRequest,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := tempSvc.GetAllTemplates(ctx, tt.filter)

			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}
