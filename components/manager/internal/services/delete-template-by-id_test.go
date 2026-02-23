// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"testing"

	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"
)

func TestUseCase_DeleteTemplateByID(t *testing.T) {
	t.Parallel()

	tempID := uuid.New()

	tests := []struct {
		name           string
		tempID         uuid.UUID
		hardDelete     bool
		mockSetup      func(ctrl *gomock.Controller) *UseCase
		expectErr      bool
		expectedResult error
	}{
		{
			name:       "Success - Delete a template",
			tempID:     tempID,
			hardDelete: true,
			mockSetup: func(ctrl *gomock.Controller) *UseCase {
				mockTempRepo := template.NewMockRepository(ctrl)
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				return &UseCase{TemplateRepo: mockTempRepo}
			},
			expectErr:      false,
			expectedResult: nil,
		},
		{
			name:       "Error Bad Request - Delete a template",
			tempID:     tempID,
			hardDelete: true,
			mockSetup: func(ctrl *gomock.Controller) *UseCase {
				mockTempRepo := template.NewMockRepository(ctrl)
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(constant.ErrBadRequest)
				return &UseCase{TemplateRepo: mockTempRepo}
			},
			expectErr:      true,
			expectedResult: constant.ErrBadRequest,
		},
		{
			name:       "Error Document Not found - Delete a template",
			tempID:     tempID,
			hardDelete: true,
			mockSetup: func(ctrl *gomock.Controller) *UseCase {
				mockTempRepo := template.NewMockRepository(ctrl)
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mongo.ErrNoDocuments)
				return &UseCase{TemplateRepo: mockTempRepo}
			},
			expectErr:      true,
			expectedResult: mongo.ErrNoDocuments,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tempSvc := tt.mockSetup(ctrl)

			ctx := context.Background()
			err := tempSvc.DeleteTemplateByID(ctx, tt.tempID, tt.hardDelete)

			if tt.expectErr {
				assert.ErrorIs(t, err, tt.expectedResult)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
