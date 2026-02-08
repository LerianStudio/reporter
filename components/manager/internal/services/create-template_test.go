// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/components/manager/internal/adapters/redis"
	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/pkg/postgres"
	templateSeaweedFS "github.com/LerianStudio/reporter/pkg/seaweedfs/template"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUseCase_CreateTemplate(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because ResetRegisteredDataSourceIDsForTesting
	// mutates global state (registered datasource IDs). Running this test in parallel
	// with other tests that call Reset/Register would cause data races on the shared map.
	//
	// Register datasource IDs once at the top level before subtests start.
	// Subtests only READ this global state, so they can safely run in parallel.
	pkg.ResetRegisteredDataSourceIDsForTesting()
	pkg.RegisterDataSourceIDsForTesting([]string{"midaz_organization", "midaz_onboarding"})

	mongoSchemas := []mongodb.CollectionSchema{
		{
			CollectionName: "organization",
			Fields: []mongodb.FieldInformation{
				{Name: "legal_document", DataType: "string"},
				{Name: "legal_name", DataType: "string"},
				{Name: "doing_business_as", DataType: "string"},
				{Name: "address", DataType: "array"},
			},
		},
	}

	postgresSchemas := []postgres.TableSchema{
		{
			TableName: "ledger",
			Columns: []postgres.ColumnInformation{
				{Name: "name", DataType: "string"},
				{Name: "status", DataType: "string"},
			},
		},
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
		mockSetup      func(ctrl *gomock.Controller) *UseCase
		expectErr      bool
		errContains    string
		expectedResult bool
	}{
		{
			name:         "Success - Create a template",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			fileHeader:   templateTestFileHeader,
			mockSetup: func(ctrl *gomock.Controller) *UseCase {
				mockTempRepo := template.NewMockRepository(ctrl)
				mockDataSourceMongo := mongodb.NewMockRepository(ctrl)
				mockDataSourcePostgres := postgres.NewMockRepository(ctrl)
				mockTemplateStorage := templateSeaweedFS.NewMockRepository(ctrl)

				tempId := uuid.New()
				timestamp := time.Now().Unix()

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
					Return(&template.Template{
						ID:           tempId,
						OutputFormat: "xml",
						Description:  "Template Financeiro",
						FileName:     fmt.Sprintf("%s_%d.tpl", tempId.String(), timestamp),
						CreatedAt:    time.Time{},
					}, nil)
				mockTemplateStorage.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				return &UseCase{
					TemplateRepo:      mockTempRepo,
					TemplateSeaweedFS: mockTemplateStorage,
					ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
						"midaz_organization": {
							DatabaseType: "mongodb", MongoDBRepository: mockDataSourceMongo,
							PostgresRepository: mockDataSourcePostgres, MongoDBName: "organization", Initialized: true,
						},
						"midaz_onboarding": {
							DatabaseType: "postgresql", PostgresRepository: mockDataSourcePostgres,
							MongoDBRepository: mockDataSourceMongo, MongoDBName: "ledger", Initialized: true,
							DatabaseConfig: &postgres.Connection{Connected: true},
						},
					}),
				}
			},
			expectErr:      false,
			expectedResult: true,
		},
		{
			name:         "Error - Create a template",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			fileHeader:   templateTestFileHeader,
			mockSetup: func(ctrl *gomock.Controller) *UseCase {
				mockTempRepo := template.NewMockRepository(ctrl)
				mockDataSourceMongo := mongodb.NewMockRepository(ctrl)
				mockDataSourcePostgres := postgres.NewMockRepository(ctrl)
				mockTemplateStorage := templateSeaweedFS.NewMockRepository(ctrl)

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

				return &UseCase{
					TemplateRepo:      mockTempRepo,
					TemplateSeaweedFS: mockTemplateStorage,
					ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
						"midaz_organization": {
							DatabaseType: "mongodb", MongoDBRepository: mockDataSourceMongo,
							PostgresRepository: mockDataSourcePostgres, MongoDBName: "organization", Initialized: true,
						},
						"midaz_onboarding": {
							DatabaseType: "postgresql", PostgresRepository: mockDataSourcePostgres,
							MongoDBRepository: mockDataSourceMongo, MongoDBName: "ledger", Initialized: true,
							DatabaseConfig: &postgres.Connection{Connected: true},
						},
					}),
				}
			},
			expectErr:   true,
			errContains: constant.ErrInternalServer.Error(),
		},
		{
			name:         "Error - Create a template with <script> tag",
			templateFile: `<html><script>alert('x')</script></html>`,
			outFormat:    "html",
			description:  "Malicious Template",
			fileHeader:   templateTestFileHeader,
			mockSetup: func(ctrl *gomock.Controller) *UseCase {
				return &UseCase{
					ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{}),
				}
			},
			expectErr:   true,
			errContains: constant.ErrScriptTagDetected.Error(),
		},
		{
			name:         "Error - ReadMultipartFile failure",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			fileHeader:   &multipart.FileHeader{},
			mockSetup: func(ctrl *gomock.Controller) *UseCase {
				mockTempRepo := template.NewMockRepository(ctrl)
				mockDataSourceMongo := mongodb.NewMockRepository(ctrl)
				mockDataSourcePostgres := postgres.NewMockRepository(ctrl)
				mockTemplateStorage := templateSeaweedFS.NewMockRepository(ctrl)

				tempId := uuid.New()
				timestamp := time.Now().Unix()

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
					Return(&template.Template{
						ID:           tempId,
						OutputFormat: "xml",
						Description:  "Template Financeiro",
						FileName:     fmt.Sprintf("%s_%d.tpl", tempId.String(), timestamp),
						CreatedAt:    time.Time{},
					}, nil)

				return &UseCase{
					TemplateRepo:      mockTempRepo,
					TemplateSeaweedFS: mockTemplateStorage,
					ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
						"midaz_organization": {
							DatabaseType: "mongodb", MongoDBRepository: mockDataSourceMongo,
							PostgresRepository: mockDataSourcePostgres, MongoDBName: "organization", Initialized: true,
						},
						"midaz_onboarding": {
							DatabaseType: "postgresql", PostgresRepository: mockDataSourcePostgres,
							MongoDBRepository: mockDataSourceMongo, MongoDBName: "ledger", Initialized: true,
							DatabaseConfig: &postgres.Connection{Connected: true},
						},
					}),
				}
			},
			expectErr:   true,
			errContains: "open",
		},
		{
			name:         "Error - Storage Put failure with successful rollback",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			fileHeader:   templateTestFileHeader,
			mockSetup: func(ctrl *gomock.Controller) *UseCase {
				mockTempRepo := template.NewMockRepository(ctrl)
				mockDataSourceMongo := mongodb.NewMockRepository(ctrl)
				mockDataSourcePostgres := postgres.NewMockRepository(ctrl)
				mockTemplateStorage := templateSeaweedFS.NewMockRepository(ctrl)

				tempId := uuid.New()
				timestamp := time.Now().Unix()

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
					Return(&template.Template{
						ID:           tempId,
						OutputFormat: "xml",
						Description:  "Template Financeiro",
						FileName:     fmt.Sprintf("%s_%d.tpl", tempId.String(), timestamp),
						CreatedAt:    time.Time{},
					}, nil)
				mockTemplateStorage.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("storage unavailable"))
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), gomock.Any(), true).
					Return(nil)

				return &UseCase{
					TemplateRepo:      mockTempRepo,
					TemplateSeaweedFS: mockTemplateStorage,
					ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
						"midaz_organization": {
							DatabaseType: "mongodb", MongoDBRepository: mockDataSourceMongo,
							PostgresRepository: mockDataSourcePostgres, MongoDBName: "organization", Initialized: true,
						},
						"midaz_onboarding": {
							DatabaseType: "postgresql", PostgresRepository: mockDataSourcePostgres,
							MongoDBRepository: mockDataSourceMongo, MongoDBName: "ledger", Initialized: true,
							DatabaseConfig: &postgres.Connection{Connected: true},
						},
					}),
				}
			},
			expectErr:   true,
			errContains: "storage unavailable",
		},
		{
			name:         "Error - Storage Put failure with rollback failure",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			fileHeader:   templateTestFileHeader,
			mockSetup: func(ctrl *gomock.Controller) *UseCase {
				mockTempRepo := template.NewMockRepository(ctrl)
				mockDataSourceMongo := mongodb.NewMockRepository(ctrl)
				mockDataSourcePostgres := postgres.NewMockRepository(ctrl)
				mockTemplateStorage := templateSeaweedFS.NewMockRepository(ctrl)

				tempId := uuid.New()
				timestamp := time.Now().Unix()

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
					Return(&template.Template{
						ID:           tempId,
						OutputFormat: "xml",
						Description:  "Template Financeiro",
						FileName:     fmt.Sprintf("%s_%d.tpl", tempId.String(), timestamp),
						CreatedAt:    time.Time{},
					}, nil)
				mockTemplateStorage.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("storage unavailable"))
				mockTempRepo.EXPECT().
					Delete(gomock.Any(), gomock.Any(), true).
					Return(errors.New("delete failed"))

				return &UseCase{
					TemplateRepo:      mockTempRepo,
					TemplateSeaweedFS: mockTemplateStorage,
					ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
						"midaz_organization": {
							DatabaseType: "mongodb", MongoDBRepository: mockDataSourceMongo,
							PostgresRepository: mockDataSourcePostgres, MongoDBName: "organization", Initialized: true,
						},
						"midaz_onboarding": {
							DatabaseType: "postgresql", PostgresRepository: mockDataSourcePostgres,
							MongoDBRepository: mockDataSourceMongo, MongoDBName: "ledger", Initialized: true,
							DatabaseConfig: &postgres.Connection{Connected: true},
						},
					}),
				}
			},
			expectErr:   true,
			errContains: "storage unavailable",
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
			result, err := tempSvc.CreateTemplate(ctx, tt.templateFile, tt.outFormat, tt.description, tt.fileHeader)

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

func TestUseCase_CreateTemplateWithPluginCRM(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because ResetRegisteredDataSourceIDsForTesting
	// mutates global state (registered datasource IDs). Running this test in parallel
	// with other tests that call Reset/Register would cause data races on the shared map.
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

// hashTemplateIdempotencyInput computes a SHA256 hash of the JSON-serialized template
// idempotency input. This is a test helper that mirrors the hashing logic in
// buildTemplateIdempotencyKey.
func hashTemplateIdempotencyInput(t *testing.T, templateFile, outFormat, description string) string {
	t.Helper()

	input := templateIdempotencyInput{
		TemplateFile: templateFile,
		OutputFormat: outFormat,
		Description:  description,
	}

	data, err := json.Marshal(input)
	require.NoError(t, err, "failed to marshal template input for hash computation")

	return commons.HashSHA256(string(data))
}

func TestUseCase_CreateTemplate_Idempotency(t *testing.T) {
	t.Parallel()

	// Register datasource IDs once at the top level before subtests start.
	// Subtests only READ this global state, so they can safely run in parallel.
	pkg.ResetRegisteredDataSourceIDsForTesting()
	pkg.RegisterDataSourceIDsForTesting([]string{"midaz_organization", "midaz_onboarding"})

	mongoSchemas := []mongodb.CollectionSchema{
		{
			CollectionName: "organization",
			Fields: []mongodb.FieldInformation{
				{Name: "legal_document", DataType: "string"},
				{Name: "legal_name", DataType: "string"},
				{Name: "doing_business_as", DataType: "string"},
				{Name: "address", DataType: "array"},
			},
		},
	}

	postgresSchemas := []postgres.TableSchema{
		{
			TableName: "ledger",
			Columns: []postgres.ColumnInformation{
				{Name: "name", DataType: "string"},
				{Name: "status", DataType: "string"},
			},
		},
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

	outFormat := "xml"
	description := "Template Financeiro"

	templateID := uuid.New()

	templateEntity := &template.Template{
		ID:           templateID,
		OutputFormat: outFormat,
		Description:  description,
		FileName:     fmt.Sprintf("%s.tpl", templateID.String()),
		CreatedAt:    time.Time{},
	}

	// Pre-compute the expected idempotency key based on the request body hash
	expectedHash := hashTemplateIdempotencyInput(t, templateTest, outFormat, description)
	expectedIdempotencyKey := "idempotency:template:" + expectedHash

	// Pre-compute the expected cached response JSON
	cachedResponseJSON, err := json.Marshal(templateEntity)
	require.NoError(t, err, "failed to marshal template entity for cached response")

	// Idempotency TTL: 24 hours as specified in requirements
	idempotencyTTL := 24 * time.Hour

	tests := []struct {
		name           string
		templateFile   string
		outFormat      string
		description    string
		idempotencyKey string
		mockSetup      func(
			mockRedisRepo *redis.MockRedisRepository,
			mockTempRepo *template.MockRepository,
			mockTemplateStorage *templateSeaweedFS.MockRepository,
			mockDataSourceMongo *mongodb.MockRepository,
			mockDataSourcePostgres *postgres.MockRepository,
		)
		expectErr      bool
		errContains    string
		expectedResult *template.Template
		testDesc       string
	}{
		{
			name:           "Success - First call creates template with idempotency lock",
			templateFile:   templateTest,
			outFormat:      outFormat,
			description:    description,
			idempotencyKey: "",
			mockSetup: func(
				mockRedisRepo *redis.MockRedisRepository,
				mockTempRepo *template.MockRepository,
				mockTemplateStorage *templateSeaweedFS.MockRepository,
				mockDataSourceMongo *mongodb.MockRepository,
				mockDataSourcePostgres *postgres.MockRepository,
			) {
				// Expect SetNX to be called with the hash-based key BEFORE template creation
				mockRedisRepo.EXPECT().
					SetNX(gomock.Any(), expectedIdempotencyKey, gomock.Any(), idempotencyTTL).
					Return(true, nil)

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

				// After successful creation, expect the result to be cached in Redis
				mockRedisRepo.EXPECT().
					Set(gomock.Any(), expectedIdempotencyKey, string(cachedResponseJSON), idempotencyTTL).
					Return(nil)
			},
			expectErr:      false,
			expectedResult: templateEntity,
			testDesc: "The first call must acquire the SetNX lock, create the template, " +
				"upload to storage, and cache the response for future duplicates.",
		},
		{
			name:           "Success - Duplicate request returns cached template (no new template created)",
			templateFile:   templateTest,
			outFormat:      outFormat,
			description:    description,
			idempotencyKey: "",
			mockSetup: func(
				mockRedisRepo *redis.MockRedisRepository,
				mockTempRepo *template.MockRepository,
				mockTemplateStorage *templateSeaweedFS.MockRepository,
				mockDataSourceMongo *mongodb.MockRepository,
				mockDataSourcePostgres *postgres.MockRepository,
			) {
				// SetNX returns false: key already exists (duplicate request)
				mockRedisRepo.EXPECT().
					SetNX(gomock.Any(), expectedIdempotencyKey, gomock.Any(), idempotencyTTL).
					Return(false, nil)

				// Expect Get to retrieve the cached response
				mockRedisRepo.EXPECT().
					Get(gomock.Any(), expectedIdempotencyKey).
					Return(string(cachedResponseJSON), nil)

				// NO calls to TemplateRepo, storage, or datasource should happen
				// (gomock will fail if unexpected calls are made)
			},
			expectErr:      false,
			expectedResult: templateEntity,
			testDesc: "When SetNX returns false, it means a duplicate request. " +
				"The service must return the cached response without creating a new template.",
		},
		{
			name:         "Success - Different file content creates a different template",
			templateFile: `<html>{{ midaz_organization.organization.legal_name }}</html>`,
			outFormat:    "html",
			description:  "Different Template",
			mockSetup: func(
				mockRedisRepo *redis.MockRedisRepository,
				mockTempRepo *template.MockRepository,
				mockTemplateStorage *templateSeaweedFS.MockRepository,
				mockDataSourceMongo *mongodb.MockRepository,
				mockDataSourcePostgres *postgres.MockRepository,
			) {
				// Different body produces a different hash, so SetNX succeeds (new key)
				mockRedisRepo.EXPECT().
					SetNX(gomock.Any(), gomock.Not(gomock.Eq(expectedIdempotencyKey)), gomock.Any(), idempotencyTTL).
					Return(true, nil)

				mockDataSourceMongo.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(mongoSchemas, nil)
				mockDataSourceMongo.EXPECT().
					CloseConnection(gomock.Any()).
					Return(nil)

				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(templateEntity, nil)

				mockTemplateStorage.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				mockRedisRepo.EXPECT().
					Set(gomock.Any(), gomock.Not(gomock.Eq(expectedIdempotencyKey)), gomock.Any(), idempotencyTTL).
					Return(nil)
			},
			expectErr:      false,
			expectedResult: templateEntity,
			testDesc: "A request with different template content produces a different hash, " +
				"so SetNX succeeds and a new template is created normally.",
		},
		{
			name:           "Success - Client-provided Idempotency-Key header is used instead of hash",
			templateFile:   templateTest,
			outFormat:      outFormat,
			description:    description,
			idempotencyKey: "client-provided-unique-key-12345",
			mockSetup: func(
				mockRedisRepo *redis.MockRedisRepository,
				mockTempRepo *template.MockRepository,
				mockTemplateStorage *templateSeaweedFS.MockRepository,
				mockDataSourceMongo *mongodb.MockRepository,
				mockDataSourcePostgres *postgres.MockRepository,
			) {
				// When client provides an explicit key, it is used instead of hashing the body
				clientIdempotencyKey := "idempotency:template:client-provided-unique-key-12345"

				mockRedisRepo.EXPECT().
					SetNX(gomock.Any(), clientIdempotencyKey, gomock.Any(), idempotencyTTL).
					Return(true, nil)

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

				mockRedisRepo.EXPECT().
					Set(gomock.Any(), clientIdempotencyKey, gomock.Any(), idempotencyTTL).
					Return(nil)
			},
			expectErr:      false,
			expectedResult: templateEntity,
			testDesc: "When the client provides an Idempotency-Key header, the service must " +
				"use that key instead of computing a hash from the request body.",
		},
		{
			name:           "Error - Duplicate in-flight request (SetNX false, value still processing)",
			templateFile:   templateTest,
			outFormat:      outFormat,
			description:    description,
			idempotencyKey: "",
			mockSetup: func(
				mockRedisRepo *redis.MockRedisRepository,
				mockTempRepo *template.MockRepository,
				mockTemplateStorage *templateSeaweedFS.MockRepository,
				mockDataSourceMongo *mongodb.MockRepository,
				mockDataSourcePostgres *postgres.MockRepository,
			) {
				// SetNX returns false: key already exists
				mockRedisRepo.EXPECT().
					SetNX(gomock.Any(), expectedIdempotencyKey, gomock.Any(), idempotencyTTL).
					Return(false, nil)

				// Get returns "processing": first request is still in-flight
				mockRedisRepo.EXPECT().
					Get(gomock.Any(), expectedIdempotencyKey).
					Return("processing", nil)
			},
			expectErr:      true,
			errContains:    "A duplicate request is currently being processed",
			expectedResult: nil,
			testDesc: "When SetNX returns false and Get returns 'processing', it means the first request " +
				"is still in-flight. The service must return an error indicating a duplicate in-flight request.",
		},
		{
			name:           "Error - Redis SetNX fails",
			templateFile:   templateTest,
			outFormat:      outFormat,
			description:    description,
			idempotencyKey: "",
			mockSetup: func(
				mockRedisRepo *redis.MockRedisRepository,
				mockTempRepo *template.MockRepository,
				mockTemplateStorage *templateSeaweedFS.MockRepository,
				mockDataSourceMongo *mongodb.MockRepository,
				mockDataSourcePostgres *postgres.MockRepository,
			) {
				// Redis is unavailable
				mockRedisRepo.EXPECT().
					SetNX(gomock.Any(), expectedIdempotencyKey, gomock.Any(), idempotencyTTL).
					Return(false, constant.ErrInternalServer)
			},
			expectErr:      true,
			errContains:    constant.ErrInternalServer.Error(),
			expectedResult: nil,
			testDesc: "When Redis SetNX fails due to infrastructure error, the service must " +
				"return an error rather than proceeding without idempotency protection.",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Each subtest gets its own mock controller to avoid interference
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTempRepo := template.NewMockRepository(ctrl)
			mockTemplateStorage := templateSeaweedFS.NewMockRepository(ctrl)
			mockRedisRepo := redis.NewMockRedisRepository(ctrl)
			mockDataSourceMongo := mongodb.NewMockRepository(ctrl)
			mockDataSourcePostgres := postgres.NewMockRepository(ctrl)

			tt.mockSetup(mockRedisRepo, mockTempRepo, mockTemplateStorage, mockDataSourceMongo, mockDataSourcePostgres)

			tempSvc := &UseCase{
				TemplateRepo:      mockTempRepo,
				TemplateSeaweedFS: mockTemplateStorage,
				RedisRepo:         mockRedisRepo,
				ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
					"midaz_organization": {
						DatabaseType:       "mongodb",
						MongoDBRepository:  mockDataSourceMongo,
						PostgresRepository: mockDataSourcePostgres,
						MongoDBName:        "organization",
						Initialized:        true,
					},
					"midaz_onboarding": {
						DatabaseType:       "postgresql",
						PostgresRepository: mockDataSourcePostgres,
						MongoDBRepository:  mockDataSourceMongo,
						MongoDBName:        "ledger",
						Initialized:        true,
						DatabaseConfig:     &postgres.Connection{Connected: true},
					},
				}),
			}

			ctx := context.Background()

			// If an idempotency key is provided, inject it into context
			if tt.idempotencyKey != "" {
				ctx = context.WithValue(ctx, constant.IdempotencyKeyCtx, tt.idempotencyKey)
			}

			result, err := tempSvc.CreateTemplate(ctx, tt.templateFile, tt.outFormat, tt.description, templateTestFileHeader)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.ID, result.ID)
				assert.Equal(t, tt.expectedResult.OutputFormat, result.OutputFormat)
				assert.Equal(t, tt.expectedResult.Description, result.Description)
			}
		})
	}
}
