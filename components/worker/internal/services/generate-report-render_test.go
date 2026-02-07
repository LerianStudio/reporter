// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"errors"
	"testing"

	reportData "github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/pkg/seaweedfs/template"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLoadTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		templateContent []byte
		templateErr     error
		expectError     bool
	}{
		{
			name:            "Success - loads template content",
			templateContent: []byte("Hello {{ name }}"),
			templateErr:     nil,
			expectError:     false,
		},
		{
			name:            "Error - template not found",
			templateContent: nil,
			templateErr:     errors.New("template not found"),
			expectError:     true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTemplateRepo := template.NewMockRepository(ctrl)
			logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())
			_, span := tracer.Start(context.Background(), "test")

			templateID := uuid.New()
			reportID := uuid.New()

			mockTemplateRepo.EXPECT().
				Get(gomock.Any(), templateID.String()).
				Return(tt.templateContent, tt.templateErr)

			useCase := &UseCase{
				TemplateSeaweedFS: mockTemplateRepo,
			}

			if tt.expectError {
				mockReportDataRepo := reportData.NewMockRepository(ctrl)
				mockReportDataRepo.EXPECT().
					UpdateReportStatusById(gomock.Any(), "Error", reportID, gomock.Any(), gomock.Any()).
					Return(nil)
				useCase.ReportDataRepo = mockReportDataRepo
			}

			message := GenerateReportMessage{
				TemplateID: templateID,
				ReportID:   reportID,
			}

			result, err := useCase.loadTemplate(context.Background(), tracer, message, &span, logger)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, string(tt.templateContent), string(result))
			}
		})
	}
}

func TestConvertToPDFIfNeeded_NonPDFFormat(t *testing.T) {
	t.Parallel()

	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())
	_, span := tracer.Start(context.Background(), "test")

	useCase := &UseCase{}

	message := GenerateReportMessage{
		ReportID:     uuid.New(),
		OutputFormat: "html",
	}

	htmlContent := "<html><body>Test</body></html>"

	result, err := useCase.convertToPDFIfNeeded(context.Background(), tracer, message, htmlContent, &span, logger)
	require.NoError(t, err)
	assert.Equal(t, htmlContent, result, "expected unchanged content for non-PDF format")
}
