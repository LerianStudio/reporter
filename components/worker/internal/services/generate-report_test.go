package services

import (
	"context"
	"encoding/json"
	"errors"
	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/minio/report"
	"plugin-smart-templates/pkg/minio/template"
	reportData "plugin-smart-templates/pkg/mongodb/report"
	postgres2 "plugin-smart-templates/pkg/postgres"
	"strings"
	"testing"
)

func Test_getContentType(t *testing.T) {
	tests := []struct {
		name         string
		extension    string
		expectedType string
	}{
		{
			name:         "existing mime type",
			extension:    "html",
			expectedType: "text/html",
		},
		{
			name:         "unknown mime type",
			extension:    "unknown",
			expectedType: "text/plain",
		},
		{
			name:         "empty extension",
			extension:    "",
			expectedType: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getContentType(tt.extension)
			if got != tt.expectedType {
				t.Errorf("getContentType(%q) = %q; want %q", tt.extension, got, tt.expectedType)
			}
		})
	}
}

func TestGenerateReport_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTemplateRepo := template.NewMockRepository(ctrl)
	mockReportRepo := report.NewMockRepository(ctrl)
	mockPostgresRepo := postgres2.NewMockRepository(ctrl)
	mockReportDataRepo := reportData.NewMockRepository(ctrl)

	templateID := uuid.New()
	reportID := uuid.New()

	body := GenerateReportMessage{
		TemplateID:   templateID,
		ReportID:     reportID,
		OutputFormat: "txt",
		DataQueries: map[string]map[string][]string{
			"onboarding": {"organization": {"name"}},
		},
		Filters: map[string]map[string]map[string][]any{
			"onboarding": {
				"organization": {
					"id": {1, 2, 3},
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	mockTemplateRepo.
		EXPECT().
		Get(gomock.Any(), templateID.String()).
		Return([]byte("Hello {{ onboarding.organization.0.name }}"), nil)

	mockPostgresRepo.
		EXPECT().
		GetDatabaseSchema(gomock.Any()).
		Return([]postgres2.TableSchema{
			{
				TableName: "organization",
				Columns: []postgres2.ColumnInformation{
					{Name: "name", DataType: "text"},
					{Name: "id", DataType: "integer", IsPrimaryKey: true},
				},
			},
		}, nil)

	mockPostgresRepo.
		EXPECT().
		Query(
			gomock.Any(),
			gomock.Any(),
			"organization",
			[]string{"name"},
			gomock.Any(),
		).
		Return([]map[string]any{{"name": "World"}}, nil)

	mockReportRepo.
		EXPECT().
		Put(gomock.Any(), gomock.Any(), "text/plain", gomock.Any()).
		Return(nil)

	mockReportDataRepo.
		EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), nil).
		Return(nil)

	useCase := &UseCase{
		TemplateFileRepo: mockTemplateRepo,
		ReportFileRepo:   mockReportRepo,
		ReportDataRepo:   mockReportDataRepo,
		ExternalDataSources: map[string]pkg.DataSource{
			"onboarding": {
				Initialized:        true,
				DatabaseType:       "postgresql",
				PostgresRepository: mockPostgresRepo,
			},
		},
	}

	err := useCase.GenerateReport(context.Background(), bodyBytes)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateReport_TemplateRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTemplateRepo := template.NewMockRepository(ctrl)
	mockReportDataRepo := reportData.NewMockRepository(ctrl)

	templateID := uuid.New()
	reportID := uuid.New()

	body := GenerateReportMessage{
		TemplateID:   templateID,
		ReportID:     reportID,
		OutputFormat: "txt",
		DataQueries:  map[string]map[string][]string{},
	}
	bodyBytes, _ := json.Marshal(body)

	mockTemplateRepo.
		EXPECT().
		Get(gomock.Any(), templateID.String()).
		Return(nil, errors.New("failed to get file"))

	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := &UseCase{
		TemplateFileRepo:    mockTemplateRepo,
		ReportDataRepo:      mockReportDataRepo,
		ExternalDataSources: map[string]pkg.DataSource{},
	}

	err := useCase.GenerateReport(context.Background(), bodyBytes)
	if err == nil || !strings.Contains(err.Error(), "failed to get file") {
		t.Errorf("expected template get error, got: %v", err)
	}
}

func TestSaveReport_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)

	useCase := &UseCase{
		ReportFileRepo: mockReportRepo,
	}

	reportID := uuid.New()
	message := GenerateReportMessage{
		ReportID:     reportID,
		OutputFormat: "csv",
	}
	renderedOutput := "id,name\n1,Jane"

	mockReportRepo.
		EXPECT().
		Put(gomock.Any(), gomock.Any(), "text/csv", []byte(renderedOutput)).
		Return(nil)

	ctx := context.Background()

	logger := libCommons.NewLoggerFromContext(ctx)
	tracer := libCommons.NewTracerFromContext(ctx)

	err := useCase.saveReport(ctx, tracer, message, renderedOutput, logger)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSaveReport_ErrorOnPut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)
	mockReportDataRepo := reportData.NewMockRepository(ctrl)

	useCase := &UseCase{
		ReportFileRepo: mockReportRepo,
		ReportDataRepo: mockReportDataRepo,
	}

	reportID := uuid.New()
	message := GenerateReportMessage{
		ReportID:     reportID,
		OutputFormat: "html",
	}
	output := "<html></html>"

	mockReportRepo.
		EXPECT().
		Put(gomock.Any(), gomock.Any(), "text/html", gomock.Any()).
		Return(errors.New("failed to put file"))

	ctx := context.Background()

	logger := libCommons.NewLoggerFromContext(ctx)
	tracer := libCommons.NewTracerFromContext(ctx)

	err := useCase.saveReport(ctx, tracer, message, output, logger)
	if err == nil || !strings.Contains(err.Error(), "failed to put file") {
		t.Errorf("expected error on Put, got: %v", err)
	}
}
