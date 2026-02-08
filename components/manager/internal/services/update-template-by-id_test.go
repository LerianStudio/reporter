// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"testing"

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

func createFileHeaderFromString(content, filename string) (*multipart.FileHeader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	h.Set("Content-Type", "application/tpl")

	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, bytes.NewReader([]byte(content)))
	if err != nil {
		return nil, err
	}

	writer.Close()

	// Parse multipart body to get FileHeader
	r := multipart.NewReader(body, writer.Boundary())
	form, err := r.ReadForm(int64(body.Len()))
	if err != nil {
		return nil, err
	}

	files := form.File["file"]
	if len(files) == 0 {
		return nil, errors.New("no file found in form")
	}

	return files[0], nil
}

func TestUseCase_UpdateTemplateByID(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because ResetRegisteredDataSourceIDsForTesting mutates global state
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Register datasource IDs for testing
	pkg.ResetRegisteredDataSourceIDsForTesting()
	pkg.RegisterDataSourceIDsForTesting([]string{"midaz_organization", "midaz_onboarding"})

	mockTempRepo := template.NewMockRepository(ctrl)
	mockTempSeaweedFS := templateSeaweedFS.NewMockRepository(ctrl)
	mockDataSourceMongo := mongodb.NewMockRepository(ctrl)
	mockDataSourcePostgres := postgres.NewMockRepository(ctrl)
	htmlType := "html"

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
		DatabaseConfig:     nil,
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
		TemplateSeaweedFS:   mockTempSeaweedFS,
		ExternalDataSources: pkg.NewSafeDataSources(externalDataSourcesMap),
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
	templateTestXMLFileHeader, _ := createFileHeaderFromString(templateTest, "teste_template_XML.tpl")

	tests := []struct {
		name         string
		templateFile *multipart.FileHeader
		outFormat    string
		description  string
		tempId       uuid.UUID
		mockSetup    func()
		expectErr    bool
		errContains  string
	}{
		{
			name:         "Success - Update outputFormat template",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "xml",
			description:  "Template Atualizado",
			tempId:       uuid.New(),
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

				// First FindByID to get current template (before update)
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(&template.Template{
						FileName:     "test-template.tpl",
						OutputFormat: "xml",
					}, nil)

				mockTempSeaweedFS.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				mockTempRepo.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				// Second FindByID to get updated template (after update)
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(&template.Template{
						FileName:     "test-template.tpl",
						OutputFormat: "xml",
					}, nil)
			},
			expectErr: false,
		},
		{
			name:         "Error - Update all template fail to find ouputFormat",
			templateFile: templateTestXMLFileHeader,
			description:  "Template Financeiro",
			tempId:       uuid.New(),
			errContains:  constant.ErrInternalServer.Error(),
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
					FindOutputFormatByID(gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr: true,
		},
		{
			name:         "Error - Update all template fail to outputFormat is not equal update file content",
			templateFile: templateTestXMLFileHeader,
			description:  "Template Financeiro",
			tempId:       uuid.New(),
			errContains:  constant.ErrFileContentInvalid.Error(),
			mockSetup: func() {
				htmlTypeP := &htmlType
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
					FindOutputFormatByID(gomock.Any(), gomock.Any()).
					Return(htmlTypeP, nil)
			},
			expectErr: true,
		},
		{
			name:         "Error - Update outputFormat template invalid",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "json",
			description:  "Template Financeiro",
			tempId:       uuid.New(),
			errContains:  constant.ErrInvalidOutputFormat.Error(),
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
			},
			expectErr: true,
		},
		{
			name:         "Error - Update outputFormat template where template file content invalid",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "html",
			description:  "Template Financeiro",
			tempId:       uuid.New(),
			errContains:  constant.ErrFileContentInvalid.Error(),
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
			},
			expectErr: true,
		},
		{
			name:         "Error - Update template error",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "xml",
			description:  "Template Atualizado",
			tempId:       uuid.New(),
			errContains:  constant.ErrInternalServer.Error(),
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

				// FindByID to get current template (before update)
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(&template.Template{
						FileName:     "test-template.tpl",
						OutputFormat: "xml",
					}, nil)

				mockTempSeaweedFS.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				mockTempRepo.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(constant.ErrInternalServer)
			},
			expectErr: true,
		},
		{
			name: "Error - Update template with <script> tag",
			templateFile: func() *multipart.FileHeader {
				fh, _ := createFileHeaderFromString(`<html><script>alert('x')</script></html>`, "malicious.tpl")
				return fh
			}(),
			outFormat:   "html",
			description: "Malicious Template",
			tempId:      uuid.New(),
			mockSetup:   func() {},
			expectErr:   true,
			errContains: constant.ErrScriptTagDetected.Error(),
		},
		{
			name:         "Error - GetTemplateByID after update fails",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "xml",
			description:  "Template Atualizado",
			tempId:       uuid.New(),
			errContains:  "template not found after update",
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

				// First FindByID to get current template (before update)
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(&template.Template{
						FileName:     "test-template.tpl",
						OutputFormat: "xml",
					}, nil)

				mockTempSeaweedFS.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				mockTempRepo.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				// Second FindByID (after update) fails
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("template not found after update"))
			},
			expectErr: true,
		},
		{
			name:         "Error - ReadMultipartFile fails",
			templateFile: &multipart.FileHeader{Filename: "broken.tpl", Size: 1},
			outFormat:    "xml",
			description:  "Template Atualizado",
			tempId:       uuid.New(),
			mockSetup:    func() {},
			expectErr:    true,
			errContains:  "open",
		},
		{
			name:         "Error - Storage Put fails",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "xml",
			description:  "Template Atualizado",
			tempId:       uuid.New(),
			errContains:  "storage unavailable",
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

				// FindByID to get current template (before storage upload)
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(&template.Template{
						FileName:     "test-template.tpl",
						OutputFormat: "xml",
					}, nil)

				mockTempSeaweedFS.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("storage unavailable"))
			},
			expectErr: true,
		},
		{
			name:         "Error - FindByID fails before update",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "xml",
			description:  "Template Atualizado",
			tempId:       uuid.New(),
			errContains:  "template not found",
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

				// FindByID fails before update
				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("template not found"))
			},
			expectErr: true,
		},
		{
			name:         "Success - Description only update (no file)",
			templateFile: nil,
			description:  "Updated Description Only",
			tempId:       uuid.New(),
			mockSetup: func() {
				// No FindByID before update when no file is provided
				mockTempRepo.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				mockTempRepo.EXPECT().
					FindByID(gomock.Any(), gomock.Any()).
					Return(&template.Template{
						FileName:     "test-template.tpl",
						OutputFormat: "xml",
						Description:  "Updated Description Only",
					}, nil)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			_, err := tempSvc.UpdateTemplateByID(ctx, tt.outFormat, tt.description, tt.tempId, tt.templateFile)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
