package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/v4/pkg/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_createTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockDataSourceMongo := mongodb.NewMockRepository(ctrl)
	mockDataSourcePostgres := postgres.NewMockRepository(ctrl)
	tempId := uuid.New()
	orgId := uuid.New()

	mongoSchemas := []mongodb.CollectionSchema{
		{
			CollectionName: "organization",
			Fields: []mongodb.FieldInformation{
				{
					Name:     "legal_document",
					DataType: "string",
				},
				{
					Name:     "legal_name",
					DataType: "string",
				},
				{
					Name:     "doing_business_as",
					DataType: "string",
				},
				{
					Name:     "address",
					DataType: "array",
				},
			},
		},
	}

	postgresSchemas := []postgres.TableSchema{
		{
			TableName: "ledger",
			Columns: []postgres.ColumnInformation{
				{
					Name:     "name",
					DataType: "string",
				},
				{
					Name:     "status",
					DataType: "string",
				},
			},
		},
	}

	externalDataSources := map[string]pkg.DataSource{}
	externalDataSources["midaz_organization"] = pkg.DataSource{
		DatabaseType:       "mongodb",
		PostgresRepository: mockDataSourcePostgres,
		MongoDBRepository:  mockDataSourceMongo,
		MongoURI:           "",
		MongoDBName:        "organization",
		Connection:         nil,
		Initialized:        true,
	}

	externalDataSources["midaz_onboarding"] = pkg.DataSource{
		DatabaseType:       "postgresql",
		PostgresRepository: mockDataSourcePostgres,
		MongoDBRepository:  mockDataSourceMongo,
		DatabaseConfig: &postgres.Connection{
			ConnectionString:   "",
			DBName:             "",
			ConnectionDB:       nil,
			Connected:          true,
			Logger:             nil,
			MaxOpenConnections: 0,
			MaxIdleConnections: 0,
		},
		MongoURI:    "",
		MongoDBName: "ledger",
		Connection:  nil,
		Initialized: true,
	}

	tempSvc := &UseCase{
		TemplateRepo:        mockTempRepo,
		ExternalDataSources: externalDataSources,
	}

	timestamp := time.Now().Unix()
	templateEntity := &template.Template{
		ID:           tempId,
		OutputFormat: "xml",
		Description:  "Template Financeiro",
		FileName:     fmt.Sprintf("%s_%d.tpl", tempId.String(), timestamp),
		CreatedAt:    time.Time{},
	}

	templateTest := `
		<?xml version="1.0" encoding="UTF-8"?>
		{% for org in midaz_organization.organization %}
		<Organizacao>
			<CNPJ>{{ org.legal_document }}</CNPJ>
			<NomeLegal>{{ org.legal_name }}</NomeLegal>
			<NomeFantasia>{{ org.doing_business_as }}</NomeFantasia>
			<Endereco>{{ org.address.line1 }}, {{ org.address.city }} - {{ org.address.state }}</Endereco>
		</Organizacao>
		{% endfor %}
	
		{% for l in midaz_onboarding.ledger %}
		<Ledger>
			<Nome>{{ l.name }}</Nome>
			<Status>{{ l.status }}</Status>
		</Ledger>
		{% endfor %}
	`

	tests := []struct {
		name           string
		templateFile   string
		outFormat      string
		description    string
		orgId          uuid.UUID
		mockSetup      func()
		expectErr      bool
		expectedResult *template.Template
	}{
		{
			name:         "Success - Create a template",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			orgId:        orgId,
			mockSetup: func() {

				mockDataSourceMongo.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(mongoSchemas, nil)

				mockDataSourceMongo.EXPECT().
					CloseConnection(gomock.Any()).
					Return(nil)

				mockDataSourcePostgres.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(postgresSchemas, nil)

				mockDataSourcePostgres.EXPECT().
					CloseConnection().
					Return(nil)

				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
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
			name:         "Error - Create a template",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			orgId:        orgId,
			mockSetup: func() {
				mockDataSourceMongo.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(mongoSchemas, nil)

				mockDataSourceMongo.EXPECT().
					CloseConnection(gomock.Any()).
					Return(nil)

				mockDataSourcePostgres.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(postgresSchemas, nil)

				mockDataSourcePostgres.EXPECT().
					CloseConnection().
					Return(nil)

				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			expectedResult: nil,
		},
		{
			name:           "Error - Create a template with <script> tag",
			templateFile:   `<html><script>alert('x')</script></html>`,
			outFormat:      "html",
			description:    "Malicious Template",
			orgId:          orgId,
			mockSetup:      func() {},
			expectErr:      true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := tempSvc.CreateTemplate(ctx, tt.templateFile, tt.outFormat, tt.description, tt.orgId)

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
