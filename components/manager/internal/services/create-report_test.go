package services

import (
	"context"
	"testing"

	"github.com/LerianStudio/reporter/v4/components/manager/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/model"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/v4/pkg/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"
)

func Test_createReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockReportRepo := report.NewMockRepository(ctrl)
	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)
	reportId := uuid.New()
	tempId := uuid.New()
	orgId := uuid.New()

	reportSvc := &UseCase{
		TemplateRepo: mockTempRepo,
		ReportRepo:   mockReportRepo,
		RabbitMQRepo: mockRabbitMQ,
	}

	mappedFields := map[string]map[string][]string{
		"midaz_transaction_metadata": {
			"transaction": {"metadata"},
		},
		"midaz_onboarding": {
			"asset":        {"name", "type", "code"},
			"organization": {"legal_document", "legal_name", "doing_business_as", "address"},
			"ledger":       {"name", "status"},
		},
	}

	outputFormat := "xml"

	reportInput := &model.CreateReportInput{
		TemplateID: tempId.String(),
		Filters:    nil,
	}

	reportEntity := &report.Report{
		ID:         reportId,
		TemplateID: tempId,
		Filters:    nil,
		Status:     "processing",
	}

	tests := []struct {
		name           string
		reportInput    *model.CreateReportInput
		orgId          uuid.UUID
		mockSetup      func()
		expectErr      bool
		expectedResult *report.Report
	}{
		{
			name:        "Success - Create a report",
			reportInput: reportInput,
			orgId:       orgId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(reportEntity, nil)

				mockRabbitMQ.EXPECT().
					ProducerDefault(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectErr: false,
			expectedResult: &report.Report{
				ID:         reportId,
				TemplateID: tempId,
				Filters:    nil,
				Status:     "processing",
			},
		},
		{
			name:        "Error - Find mapped fields and output format",
			reportInput: reportInput,
			orgId:       orgId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			expectedResult: nil,
		},
		{
			name:        "Error - Create report",
			reportInput: reportInput,
			orgId:       orgId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			expectedResult: nil,
		},

		{
			name:        "Error - Send message on RabbitMQ",
			reportInput: reportInput,
			orgId:       orgId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&outputFormat, mappedFields, nil)

				mockReportRepo.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(reportEntity, nil)

				mockRabbitMQ.EXPECT().
					ProducerDefault(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr: false,
			expectedResult: &report.Report{
				ID:         reportId,
				TemplateID: tempId,
				Filters:    nil,
				Status:     "processing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := reportSvc.CreateReport(ctx, tt.reportInput, tt.orgId)

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

func Test_createReport_InvalidTemplateID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reportSvc := &UseCase{}

	reportInput := &model.CreateReportInput{
		TemplateID: "invalid-uuid",
		Filters:    nil,
	}

	ctx := context.Background()
	result, err := reportSvc.CreateReport(ctx, reportInput, uuid.New())

	assert.Error(t, err)
	assert.Nil(t, result)
}

func Test_createReport_TemplateNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)

	reportSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	tempId := uuid.New()
	orgId := uuid.New()

	reportInput := &model.CreateReportInput{
		TemplateID: tempId.String(),
		Filters:    nil,
	}

	mockTempRepo.EXPECT().
		FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil, mongo.ErrNoDocuments)

	ctx := context.Background()
	result, err := reportSvc.CreateReport(ctx, reportInput, orgId)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func Test_createReport_WithFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockReportRepo := report.NewMockRepository(ctrl)
	mockRabbitMQ := rabbitmq.NewMockProducerRepository(ctrl)
	mockPostgresRepo := postgres.NewMockRepository(ctrl)

	reportId := uuid.New()
	tempId := uuid.New()
	orgId := uuid.New()

	mappedFields := map[string]map[string][]string{
		"pg_ds": {
			"users": {"id", "name", "email"},
		},
	}

	outputFormat := "pdf"

	filters := map[string]map[string]map[string]model.FilterCondition{
		"pg_ds": {
			"users": {
				"status": {
					Equals: []any{"active"},
				},
			},
		},
	}

	reportInput := &model.CreateReportInput{
		TemplateID: tempId.String(),
		Filters:    filters,
	}

	reportEntity := &report.Report{
		ID:         reportId,
		TemplateID: tempId,
		Filters:    filters,
		Status:     "processing",
	}

	reportSvc := &UseCase{
		TemplateRepo: mockTempRepo,
		ReportRepo:   mockReportRepo,
		RabbitMQRepo: mockRabbitMQ,
		ExternalDataSources: map[string]pkg.DataSource{
			"pg_ds": {
				DatabaseType:       pkg.PostgreSQLType,
				PostgresRepository: mockPostgresRepo,
				DatabaseConfig:     &postgres.Connection{Connected: true},
				Initialized:        true,
			},
		},
	}

	mockTempRepo.EXPECT().
		FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&outputFormat, mappedFields, nil)

	mockPostgresRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return([]postgres.TableSchema{
		{
			TableName: "users",
			Columns: []postgres.ColumnInformation{
				{Name: "id", DataType: "uuid"},
				{Name: "name", DataType: "varchar"},
				{Name: "email", DataType: "varchar"},
				{Name: "status", DataType: "varchar"},
			},
		},
	}, nil)
	mockPostgresRepo.EXPECT().CloseConnection().Return(nil)

	mockReportRepo.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(reportEntity, nil)

	mockRabbitMQ.EXPECT().
		ProducerDefault(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil)

	ctx := context.Background()
	result, err := reportSvc.CreateReport(ctx, reportInput, orgId)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func Test_createReport_FiltersValidationFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)

	tempId := uuid.New()
	orgId := uuid.New()

	mappedFields := map[string]map[string][]string{
		"pg_ds": {
			"users": {"id", "name"},
		},
	}

	outputFormat := "pdf"

	filters := map[string]map[string]map[string]model.FilterCondition{
		"unknown_ds": {
			"users": {
				"status": {
					Equals: []any{"active"},
				},
			},
		},
	}

	reportInput := &model.CreateReportInput{
		TemplateID: tempId.String(),
		Filters:    filters,
	}

	reportSvc := &UseCase{
		TemplateRepo:        mockTempRepo,
		ExternalDataSources: map[string]pkg.DataSource{},
	}

	mockTempRepo.EXPECT().
		FindMappedFieldsAndOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&outputFormat, mappedFields, nil)

	ctx := context.Background()
	result, err := reportSvc.CreateReport(ctx, reportInput, orgId)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func Test_convertFiltersToMappedFieldsType(t *testing.T) {
	uc := &UseCase{}

	tests := []struct {
		name     string
		filters  map[string]map[string]map[string]model.FilterCondition
		expected map[string]map[string][]string
	}{
		{
			name: "Single filter",
			filters: map[string]map[string]map[string]model.FilterCondition{
				"database": {
					"table": {
						"field1": {Equals: []any{"value"}},
					},
				},
			},
			expected: map[string]map[string][]string{
				"database": {
					"table": {"field1"},
				},
			},
		},
		{
			name: "Multiple filters - limited to 3 keys",
			filters: map[string]map[string]map[string]model.FilterCondition{
				"database": {
					"table": {
						"field1": {Equals: []any{"value1"}},
						"field2": {Equals: []any{"value2"}},
						"field3": {Equals: []any{"value3"}},
						"field4": {Equals: []any{"value4"}},
						"field5": {Equals: []any{"value5"}},
					},
				},
			},
			expected: map[string]map[string][]string{
				"database": {
					"table": {"field1", "field2", "field3"},
				},
			},
		},
		{
			name:     "Empty filters",
			filters:  map[string]map[string]map[string]model.FilterCondition{},
			expected: map[string]map[string][]string{},
		},
		{
			name: "Multiple tables",
			filters: map[string]map[string]map[string]model.FilterCondition{
				"database": {
					"table1": {
						"field1": {Equals: []any{"value"}},
					},
					"table2": {
						"field2": {Equals: []any{"value"}},
					},
				},
			},
			expected: map[string]map[string][]string{
				"database": {
					"table1": {"field1"},
					"table2": {"field2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uc.convertFiltersToMappedFieldsType(tt.filters)

			// Check that the result has the same structure
			assert.Equal(t, len(tt.expected), len(result))

			for db, tables := range tt.expected {
				assert.Contains(t, result, db)
				assert.Equal(t, len(tables), len(result[db]))

				for table, fields := range tables {
					assert.Contains(t, result[db], table)
					assert.Equal(t, len(fields), len(result[db][table]))
				}
			}
		})
	}
}
