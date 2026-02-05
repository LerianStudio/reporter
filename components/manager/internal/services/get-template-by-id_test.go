// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb/template"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"
)

func Test_getTemplateById(t *testing.T) {
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
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := tempSvc.GetTemplateByID(ctx, tt.tempId)

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
