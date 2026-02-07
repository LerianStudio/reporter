// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"
)

func TestGetTemplateByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	tempId := uuid.New()

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
		tempId         uuid.UUID
		mockSetup      func()
		expectErr      bool
		errContains    string
		expectedResult *template.Template
	}{
		{
			name:   "Success - Get a template by id",
			tempId: tempId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
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
			tempId: tempId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			errContains:    constant.ErrInternalServer.Error(),
			expectedResult: nil,
		},
		{
			name:   "Error - Get a template by id not found",
			tempId: tempId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(nil, mongo.ErrNoDocuments)
			},
			expectErr:      true,
			errContains:    "No template entity was found",
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := tempSvc.GetTemplateByID(ctx, tt.tempId)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}
