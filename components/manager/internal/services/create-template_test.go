// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/pkg/postgres"
	templateSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/template"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateTemplate(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because ResetRegisteredDataSourceIDsForTesting mutates global state
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Register datasource IDs for testing
	pkg.ResetRegisteredDataSourceIDsForTesting()
	pkg.RegisterDataSourceIDsForTesting([]string{"midaz_organization", "midaz_onboarding"})

	mockTempRepo := template.NewMockRepository(ctrl)
	mockDataSourceMongo := mongodb.NewMockRepository(ctrl)
	mockDataSourcePostgres := postgres.NewMockRepository(ctrl)
	mockTemplateStorage := templateSeaweedFS.NewMockRepository(ctrl)
	tempId := uuid.New()

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

	externalDataSourcesMap := map[string]pkg.DataSource{}
	externalDataSourcesMap["midaz_organization"] = pkg.DataSource{
		DatabaseType:       "mongodb",
		PostgresRepository: mockDataSourcePostgres,
		MongoDBRepository:  mockDataSourceMongo,
		MongoURI:           "",
		MongoDBName:        "organization",
		Connection:         nil,
		Initialized:        true,
	}

	externalDataSourcesMap["midaz_onboarding"] = pkg.DataSource{
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
		TemplateSeaweedFS:   mockTemplateStorage,
		ExternalDataSources: pkg.NewSafeDataSources(externalDataSourcesMap),
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
	templateTestFileHeader, _ := createFileHeaderFromString(templateTest, "teste_template_XML.tpl")

	tests := []struct {
		name           string
		templateFile   string
		outFormat      string
		description    string
		fileHeader     *multipart.FileHeader
		mockSetup      func()
		expectErr      bool
		expectedResult *template.Template
	}{
		{
			name:         "Success - Create a template",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			fileHeader:   templateTestFileHeader,
			mockSetup: func() {
				mockDataSourceMongo.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(mongoSchemas, nil)

				mockDataSourceMongo.EXPECT().
					CloseConnection(gomock.Any()).
					Return(nil)

				mockDataSourcePostgres.EXPECT().
					GetDatabaseSchema(gomock.Any(), gomock.Any()).
					Return(postgresSchemas, nil)

				mockDataSourcePostgres.EXPECT().
					CloseConnection().
					Return(nil)

				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(templateEntity, nil)

				mockTemplateStorage.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
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
			fileHeader:   templateTestFileHeader,
			mockSetup: func() {
				mockDataSourceMongo.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(mongoSchemas, nil)

				mockDataSourceMongo.EXPECT().
					CloseConnection(gomock.Any()).
					Return(nil)

				mockDataSourcePostgres.EXPECT().
					GetDatabaseSchema(gomock.Any(), gomock.Any()).
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
			fileHeader:     templateTestFileHeader,
			mockSetup:      func() {},
			expectErr:      true,
			expectedResult: nil,
		},
		{
			name:         "Error - ReadMultipartFile failure",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			fileHeader:   &multipart.FileHeader{},
			mockSetup: func() {
				mockDataSourceMongo.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(mongoSchemas, nil)

				mockDataSourceMongo.EXPECT().
					CloseConnection(gomock.Any()).
					Return(nil)

				mockDataSourcePostgres.EXPECT().
					GetDatabaseSchema(gomock.Any(), gomock.Any()).
					Return(postgresSchemas, nil)

				mockDataSourcePostgres.EXPECT().
					CloseConnection().
					Return(nil)

				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(templateEntity, nil)
			},
			expectErr:      true,
			expectedResult: nil,
		},
		{
			name:         "Error - Storage Put failure with successful rollback",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			fileHeader:   templateTestFileHeader,
			mockSetup: func() {
				mockDataSourceMongo.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(mongoSchemas, nil)

				mockDataSourceMongo.EXPECT().
					CloseConnection(gomock.Any()).
					Return(nil)

				mockDataSourcePostgres.EXPECT().
					GetDatabaseSchema(gomock.Any(), gomock.Any()).
					Return(postgresSchemas, nil)

				mockDataSourcePostgres.EXPECT().
					CloseConnection().
					Return(nil)

				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(templateEntity, nil)

				mockTemplateStorage.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("storage unavailable"))

				// Rollback: DeleteTemplateByID calls TemplateRepo.Delete
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), gomock.Any(), true).
					Return(nil)
			},
			expectErr:      true,
			expectedResult: nil,
		},
		{
			name:         "Error - Storage Put failure with rollback failure",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			fileHeader:   templateTestFileHeader,
			mockSetup: func() {
				mockDataSourceMongo.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(mongoSchemas, nil)

				mockDataSourceMongo.EXPECT().
					CloseConnection(gomock.Any()).
					Return(nil)

				mockDataSourcePostgres.EXPECT().
					GetDatabaseSchema(gomock.Any(), gomock.Any()).
					Return(postgresSchemas, nil)

				mockDataSourcePostgres.EXPECT().
					CloseConnection().
					Return(nil)

				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(templateEntity, nil)

				mockTemplateStorage.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("storage unavailable"))

				// Rollback fails: DeleteTemplateByID calls TemplateRepo.Delete which also fails
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), gomock.Any(), true).
					Return(errors.New("delete failed"))
			},
			expectErr:      true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := tempSvc.CreateTemplate(ctx, tt.templateFile, tt.outFormat, tt.description, tt.fileHeader)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

func TestCreateTemplateWithPluginCRM(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because ResetRegisteredDataSourceIDsForTesting mutates global state
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Register datasource IDs including plugin_crm
	pkg.ResetRegisteredDataSourceIDsForTesting()
	pkg.RegisterDataSourceIDsForTesting([]string{"plugin_crm"})

	mockTempRepo := template.NewMockRepository(ctrl)
	mockDataSourceMongo := mongodb.NewMockRepository(ctrl)
	mockTemplateStorage := templateSeaweedFS.NewMockRepository(ctrl)
	tempId := uuid.New()

	crmSchemas := []mongodb.CollectionSchema{
		{
			CollectionName: "holder_org-123-abc",
			Fields: []mongodb.FieldInformation{
				{
					Name:     "name",
					DataType: "string",
				},
				{
					Name:     "document",
					DataType: "string",
				},
			},
		},
	}

	externalDataSourcesMap := map[string]pkg.DataSource{}
	externalDataSourcesMap["plugin_crm"] = pkg.DataSource{
		DatabaseType:        "mongodb",
		MongoDBRepository:   mockDataSourceMongo,
		MongoURI:            "",
		MongoDBName:         "plugin_crm",
		Connection:          nil,
		Initialized:         true,
		MidazOrganizationID: "org-123-abc",
	}

	tempSvc := &UseCase{
		TemplateRepo:        mockTempRepo,
		TemplateSeaweedFS:   mockTemplateStorage,
		ExternalDataSources: pkg.NewSafeDataSources(externalDataSourcesMap),
	}

	templateEntity := &template.Template{
		ID:           tempId,
		OutputFormat: "xml",
		Description:  "CRM Template",
		FileName:     fmt.Sprintf("%s.tpl", tempId.String()),
		CreatedAt:    time.Time{},
	}

	templateCRM := `
		<?xml version="1.0" encoding="UTF-8"?>
		{% for h in plugin_crm.holder %}
		<Holder>
			<Name>{{ h.name }}</Name>
			<Document>{{ h.document }}</Document>
		</Holder>
		{% endfor %}
	`
	templateCRMFileHeader, _ := createFileHeaderFromString(templateCRM, "crm_template.tpl")

	t.Run("Success - Template with plugin_crm datasource", func(t *testing.T) {
		mockDataSourceMongo.EXPECT().
			GetDatabaseSchemaForOrganization(gomock.Any(), "org-123-abc").
			Return(crmSchemas, nil)

		mockDataSourceMongo.EXPECT().
			CloseConnection(gomock.Any()).
			Return(nil)

		mockTempRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(templateEntity, nil)

		mockTemplateStorage.EXPECT().
			Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		ctx := context.Background()
		result, err := tempSvc.CreateTemplate(ctx, templateCRM, "xml", "CRM Template", templateCRMFileHeader)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, tempId, result.ID)
	})
}
