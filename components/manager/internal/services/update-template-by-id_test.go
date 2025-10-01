package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"plugin-smart-templates/v3/pkg"
	"plugin-smart-templates/v3/pkg/constant"
	"plugin-smart-templates/v3/pkg/mongodb"
	"plugin-smart-templates/v3/pkg/mongodb/template"
	"plugin-smart-templates/v3/pkg/postgres"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

func Test_updateTemplateById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	mockDataSourceMongo := mongodb.NewMockRepository(ctrl)
	mockDataSourcePostgres := postgres.NewMockRepository(ctrl)
	orgId := uuid.New()
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

	externalDataSources := map[string]pkg.DataSource{}
	externalDataSources["midaz_organization"] = pkg.DataSource{
		DatabaseType:       "mongodb",
		PostgresRepository: mockDataSourcePostgres,
		MongoDBRepository:  mockDataSourceMongo,
		DatabaseConfig:     nil,
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
		orgId        uuid.UUID
		tempId       uuid.UUID
		mockSetup    func()
		expectErr    bool
	}{
		{
			name:         "Success - Update outputFormat template",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "xml",
			description:  "Template Atualizado",
			orgId:        uuid.New(),
			tempId:       uuid.New(),
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
					Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name:         "Error - Update all template fail to find ouputFormat",
			templateFile: templateTestXMLFileHeader,
			description:  "Template Financeiro",
			orgId:        orgId,
			tempId:       uuid.New(),
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
					FindOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr: true,
		},
		{
			name:         "Error - Update all template fail to outputFormat is not equal update file content",
			templateFile: templateTestXMLFileHeader,
			description:  "Template Financeiro",
			orgId:        orgId,
			tempId:       uuid.New(),
			mockSetup: func() {
				htmlTypeP := &htmlType
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
					FindOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(htmlTypeP, nil)
			},
			expectErr: true,
		},
		{
			name:         "Error - Update outputFormat template invalid",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "json",
			description:  "Template Financeiro",
			orgId:        orgId,
			tempId:       uuid.New(),
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
			},
			expectErr: true,
		},
		{
			name:         "Error - Update outputFormat template where template file content invalid",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "html",
			description:  "Template Financeiro",
			orgId:        orgId,
			tempId:       uuid.New(),
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
			},
			expectErr: true,
		},
		{
			name:         "Error - Update template error",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "xml",
			description:  "Template Atualizado",
			orgId:        orgId,
			tempId:       uuid.New(),
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
					Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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
			orgId:       orgId,
			tempId:      uuid.New(),
			mockSetup:   func() {},
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			err := tempSvc.UpdateTemplateByID(ctx, tt.outFormat, tt.description, tt.orgId, tt.tempId, tt.templateFile)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
