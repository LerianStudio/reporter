// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/model"
	mongodb2 "github.com/LerianStudio/reporter/pkg/mongodb"
	reportData "github.com/LerianStudio/reporter/pkg/mongodb/report"
	postgres2 "github.com/LerianStudio/reporter/pkg/postgres"
	"github.com/LerianStudio/reporter/pkg/seaweedfs/report"
	"github.com/LerianStudio/reporter/pkg/seaweedfs/template"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libCrypto "github.com/LerianStudio/lib-commons/v2/commons/crypto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		extension    string
		expectedType string
	}{
		{
			name:         "Success - existing mime type",
			extension:    "html",
			expectedType: "text/html",
		},
		{
			name:         "Success - unknown mime type",
			extension:    "unknown",
			expectedType: "text/plain",
		},
		{
			name:         "Success - empty extension",
			extension:    "",
			expectedType: "text/plain",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getContentType(tt.extension)
			assert.Equal(t, tt.expectedType, got, "getContentType(%q)", tt.extension)
		})
	}
}

func TestGenerateReport_Success(t *testing.T) {
	t.Parallel()

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
		Filters: map[string]map[string]map[string]model.FilterCondition{
			"onboarding": {
				"organization": {
					"id": {
						Equals: []any{1, 2, 3},
					},
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	mockReportDataRepo.
		EXPECT().
		FindByID(gomock.Any(), reportID).
		Return(&reportData.Report{
			ID:     reportID,
			Status: "processing",
		}, nil)

	mockTemplateRepo.
		EXPECT().
		Get(gomock.Any(), templateID.String()).
		Return([]byte("Hello {{ onboarding.organization.0.name }}"), nil)

	mockPostgresRepo.
		EXPECT().
		GetDatabaseSchema(gomock.Any(), gomock.Any()).
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
		QueryWithAdvancedFilters(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(), // schemaName
			"organization",
			[]string{"name"},
			gomock.Any(),
		).
		Return([]map[string]any{{"name": "World"}}, nil)

	mockReportRepo.
		EXPECT().
		Put(gomock.Any(), gomock.Any(), "text/plain", gomock.Any(), "").
		Return(nil)

	mockReportDataRepo.
		EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), nil).
		Return(nil)

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	circuitBreakerManager := pkg.NewCircuitBreakerManager(logger)

	useCase := &UseCase{
		TemplateSeaweedFS:     mockTemplateRepo,
		ReportSeaweedFS:       mockReportRepo,
		ReportDataRepo:        mockReportDataRepo,
		CircuitBreakerManager: circuitBreakerManager,
		ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
			"onboarding": {
				Initialized:        true,
				DatabaseType:       "postgresql",
				PostgresRepository: mockPostgresRepo,
			},
		}),
	}

	err := useCase.GenerateReport(context.Background(), bodyBytes)
	require.NoError(t, err)
}

func TestGenerateReport_TemplateRepoError(t *testing.T) {
	t.Parallel()

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

	mockReportDataRepo.
		EXPECT().
		FindByID(gomock.Any(), reportID).
		Return(&reportData.Report{
			ID:     reportID,
			Status: "processing",
		}, nil)

	mockTemplateRepo.
		EXPECT().
		Get(gomock.Any(), templateID.String()).
		Return(nil, errors.New("failed to get file"))

	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := &UseCase{
		TemplateSeaweedFS:   mockTemplateRepo,
		ReportDataRepo:      mockReportDataRepo,
		ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{}),
	}

	err := useCase.GenerateReport(context.Background(), bodyBytes)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get file")
}

func TestSaveReport_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)

	useCase := &UseCase{
		ReportSeaweedFS: mockReportRepo,
	}

	reportID := uuid.New()
	message := GenerateReportMessage{
		ReportID:     reportID,
		OutputFormat: "csv",
	}
	renderedOutput := "id,name\n1,Jane"

	mockReportRepo.
		EXPECT().
		Put(gomock.Any(), gomock.Any(), "text/csv", []byte(renderedOutput), "").
		Return(nil)

	ctx := context.Background()

	logger, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)

	err := useCase.saveReport(ctx, tracer, message, renderedOutput, logger)
	require.NoError(t, err)
}

func TestSaveReport_ErrorOnPut(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)
	mockReportDataRepo := reportData.NewMockRepository(ctrl)

	useCase := &UseCase{
		ReportSeaweedFS: mockReportRepo,
		ReportDataRepo:  mockReportDataRepo,
	}

	reportID := uuid.New()
	message := GenerateReportMessage{
		ReportID:     reportID,
		OutputFormat: "html",
	}
	output := "<html></html>"

	mockReportRepo.
		EXPECT().
		Put(gomock.Any(), gomock.Any(), "text/html", gomock.Any(), "").
		Return(errors.New("failed to put file"))

	ctx := context.Background()

	logger, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)

	err := useCase.saveReport(ctx, tracer, message, output, logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to put file")
}

func TestGenerateReport_PluginCRMWithEncryptedData(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTemplateRepo := template.NewMockRepository(ctrl)
	mockReportRepo := report.NewMockRepository(ctrl)
	mockMongoRepo := mongodb2.NewMockRepository(ctrl)
	mockReportDataRepo := reportData.NewMockRepository(ctrl)

	templateID := uuid.New()
	reportID := uuid.New()
	organizationID := "01956b69-9102-75b7-8860-3e75c11d231c"

	// Dados de teste - documento que será filtrado
	testDocument := "12345678901"

	// Criar instância de crypto para gerar hash do documento
	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	// Inicializar o cipher para criptografia
	err := crypto.InitializeCipher()
	require.NoError(t, err, "Failed to initialize cipher")

	hashedDocument := crypto.GenerateHash(&testDocument)

	templateContent := `Cliente: {{ plugin_crm.holders.0.name }}
Documento: {{ plugin_crm.holders.0.document }}
Email: {{ plugin_crm.holders.0.contact.primary_email }}
Conta Bancária: {{ plugin_crm.holders.0.banking_details.account }}`

	nameStr := "João Silva"
	emailStr := "joao@example.com"
	accountStr := "12345-6"

	encryptedName, _ := crypto.Encrypt(&nameStr)
	encryptedDocument, _ := crypto.Encrypt(&testDocument)
	encryptedEmail, _ := crypto.Encrypt(&emailStr)
	encryptedAccount, _ := crypto.Encrypt(&accountStr)

	mockMongoData := []map[string]any{
		{
			"_id":      "holder-123",
			"name":     *encryptedName,
			"document": *encryptedDocument,
			"search": map[string]any{
				"document": hashedDocument,
				"name":     crypto.GenerateHash(encryptedName),
			},
			"contact": map[string]any{
				"primary_email": *encryptedEmail,
			},
			"banking_details": map[string]any{
				"account": *encryptedAccount,
			},
		},
	}

	body := GenerateReportMessage{
		TemplateID:   templateID,
		ReportID:     reportID,
		OutputFormat: "html",
		DataQueries: map[string]map[string][]string{
			"plugin_crm": {
				"holders": {"name", "document", "contact.primary_email", "banking_details.account"},
			},
		},
		Filters: map[string]map[string]map[string]model.FilterCondition{
			"plugin_crm": {
				"holders": {
					"document": {
						Equals: []any{testDocument},
					},
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	mockReportDataRepo.
		EXPECT().
		FindByID(gomock.Any(), reportID).
		Return(&reportData.Report{
			ID:     reportID,
			Status: "processing",
		}, nil)

	mockTemplateRepo.
		EXPECT().
		Get(gomock.Any(), templateID.String()).
		Return([]byte(templateContent), nil)

	mockMongoRepo.
		EXPECT().
		QueryWithAdvancedFilters(
			gomock.Any(),
			"holders_"+organizationID,
			[]string{"name", "document", "contact.primary_email", "banking_details.account"},
			gomock.Any(),
		).
		DoAndReturn(func(ctx context.Context, collection string, fields []string, filters map[string]model.FilterCondition) ([]map[string]any, error) {
			searchDocFilter, exists := filters["search.document"]
			assert.True(t, exists, "Expected search.document filter to be present")
			if exists && len(searchDocFilter.Equals) > 0 {
				assert.Equal(t, hashedDocument, searchDocFilter.Equals[0], "Expected hashed document")
			}
			return mockMongoData, nil
		})

	mockReportRepo.
		EXPECT().
		Put(gomock.Any(), gomock.Any(), "text/html", gomock.Any(), "").
		DoAndReturn(func(ctx context.Context, objectName, contentType string, data []byte, ttl string) error {
			content := string(data)
			assert.Contains(t, content, "João Silva", "Expected decrypted name in rendered content")
			assert.Contains(t, content, testDocument, "Expected decrypted document in rendered content")
			assert.Contains(t, content, "joao@example.com", "Expected decrypted email in rendered content")
			assert.Contains(t, content, "12345-6", "Expected decrypted account in rendered content")
			return nil
		})

	mockReportDataRepo.
		EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), nil).
		Return(nil)

	circuitBreakerManager := pkg.NewCircuitBreakerManager(logger)

	useCase := &UseCase{
		TemplateSeaweedFS:               mockTemplateRepo,
		ReportSeaweedFS:                 mockReportRepo,
		ReportDataRepo:                  mockReportDataRepo,
		CircuitBreakerManager:           circuitBreakerManager,
		CryptoHashSecretKeyPluginCRM:    hashKey,
		CryptoEncryptSecretKeyPluginCRM: encryptKey,
		ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
			"plugin_crm": {
				Initialized:         true,
				DatabaseType:        "mongodb",
				MongoDBRepository:   mockMongoRepo,
				MidazOrganizationID: organizationID,
			},
		}),
	}

	err = useCase.GenerateReport(context.Background(), bodyBytes)
	require.NoError(t, err)
}

func TestDecryptRegulatoryFieldsFields(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	require.NoError(t, err, "Failed to initialize cipher")

	useCase := &UseCase{}

	tests := []struct {
		name           string
		record         map[string]any
		expectedDoc    string
		expectNoChange bool
	}{
		{
			name: "Success - decrypt regulatory_fields.participant_document",
			record: func() map[string]any {
				doc := "12345678901234"
				encrypted, _ := crypto.Encrypt(&doc)
				return map[string]any{
					"regulatory_fields": map[string]any{
						"participant_document": *encrypted,
					},
				}
			}(),
			expectedDoc: "12345678901234",
		},
		{
			name: "Success - no regulatory_fields present",
			record: map[string]any{
				"id": "test-id",
			},
			expectNoChange: true,
		},
		{
			name: "Success - regulatory_fields without participant_document",
			record: map[string]any{
				"regulatory_fields": map[string]any{
					"other_field": "value",
				},
			},
			expectNoChange: true,
		},
		{
			name: "Success - regulatory_fields with nil participant_document",
			record: map[string]any{
				"regulatory_fields": map[string]any{
					"participant_document": nil,
				},
			},
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := useCase.decryptRegulatoryFieldsFields(tt.record, crypto)
			require.NoError(t, err)

			if !tt.expectNoChange {
				regulatoryFields, ok := tt.record["regulatory_fields"].(map[string]any)
				require.True(t, ok, "regulatory_fields not found or wrong type")
				assert.Equal(t, tt.expectedDoc, regulatoryFields["participant_document"])
			}
		})
	}
}

func TestDecryptRelatedPartiesFields(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	require.NoError(t, err, "Failed to initialize cipher")

	useCase := &UseCase{}

	tests := []struct {
		name           string
		record         map[string]any
		expectedDocs   []string
		expectNoChange bool
	}{
		{
			name: "Success - decrypt multiple related_parties documents",
			record: func() map[string]any {
				doc1 := "11111111111"
				doc2 := "22222222222"
				encrypted1, _ := crypto.Encrypt(&doc1)
				encrypted2, _ := crypto.Encrypt(&doc2)
				return map[string]any{
					"related_parties": []any{
						map[string]any{
							"_id":      "party-1",
							"document": *encrypted1,
							"name":     "Party One",
							"role":     "PRIMARY_HOLDER",
						},
						map[string]any{
							"_id":      "party-2",
							"document": *encrypted2,
							"name":     "Party Two",
							"role":     "LEGAL_REPRESENTATIVE",
						},
					},
				}
			}(),
			expectedDocs: []string{"11111111111", "22222222222"},
		},
		{
			name: "Success - no related_parties present",
			record: map[string]any{
				"id": "test-id",
			},
			expectNoChange: true,
		},
		{
			name: "Success - empty related_parties array",
			record: map[string]any{
				"related_parties": []any{},
			},
			expectNoChange: true,
		},
		{
			name: "Success - related_parties with nil document",
			record: map[string]any{
				"related_parties": []any{
					map[string]any{
						"_id":      "party-1",
						"document": nil,
						"name":     "Party One",
					},
				},
			},
			expectNoChange: true,
		},
		{
			name: "Success - related_parties with mixed valid and nil documents",
			record: func() map[string]any {
				doc1 := "33333333333"
				encrypted1, _ := crypto.Encrypt(&doc1)
				return map[string]any{
					"related_parties": []any{
						map[string]any{
							"_id":      "party-1",
							"document": *encrypted1,
							"name":     "Party One",
						},
						map[string]any{
							"_id":      "party-2",
							"document": nil,
							"name":     "Party Two",
						},
					},
				}
			}(),
			expectedDocs: []string{"33333333333"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := useCase.decryptRelatedPartiesFields(tt.record, crypto)
			require.NoError(t, err)

			if !tt.expectNoChange && len(tt.expectedDocs) > 0 {
				relatedParties, ok := tt.record["related_parties"].([]any)
				require.True(t, ok, "related_parties not found or wrong type")

				docIndex := 0
				for i, party := range relatedParties {
					partyMap, ok := party.(map[string]any)
					require.True(t, ok, "related_parties[%d] is not a map", i)

					if partyMap["document"] != nil && docIndex < len(tt.expectedDocs) {
						assert.Equal(t, tt.expectedDocs[docIndex], partyMap["document"], "related_parties[%d].document", i)
						docIndex++
					}
				}
			}
		})
	}
}

func TestTransformPluginCRMAdvancedFilters_NewFields(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey: hashKey,
		Logger:        logger,
	}

	useCase := &UseCase{
		CryptoHashSecretKeyPluginCRM: hashKey,
	}

	tests := []struct {
		name          string
		inputField    string
		expectedField string
		inputValue    string
	}{
		{
			name:          "Success - transform regulatory_fields.participant_document",
			inputField:    "regulatory_fields.participant_document",
			expectedField: "search.regulatory_fields_participant_document",
			inputValue:    "12345678901234",
		},
		{
			name:          "Success - transform related_parties.document",
			inputField:    "related_parties.document",
			expectedField: "search.related_party_documents",
			inputValue:    "11111111111",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filter := map[string]model.FilterCondition{
				tt.inputField: {
					Equals: []any{tt.inputValue},
				},
			}

			transformedFilter, err := useCase.transformPluginCRMAdvancedFilters(filter, logger)
			require.NoError(t, err)

			// Verify the field was transformed
			assert.Contains(t, transformedFilter, tt.expectedField, "expected field not found in transformed filter")

			// Verify the original field was removed
			assert.NotContains(t, transformedFilter, tt.inputField, "original field should not exist in transformed filter")

			// Verify the value was hashed
			expectedHash := crypto.GenerateHash(&tt.inputValue)
			assert.Equal(t, expectedHash, transformedFilter[tt.expectedField].Equals[0], "expected hashed value")
		})
	}
}

func TestShouldSkipProcessing(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	tests := []struct {
		name         string
		reportID     uuid.UUID
		mockSetup    func(reportID uuid.UUID)
		expectedSkip bool
	}{
		{
			name:     "Success - Skip report already finished",
			reportID: uuid.New(),
			mockSetup: func(reportID uuid.UUID) {
				mockReportDataRepo.EXPECT().
					FindByID(gomock.Any(), reportID).
					Return(&reportData.Report{
						ID:     reportID,
						Status: "Finished",
					}, nil)
			},
			expectedSkip: true,
		},
		{
			name:     "Success - Skip report in error state",
			reportID: uuid.New(),
			mockSetup: func(reportID uuid.UUID) {
				mockReportDataRepo.EXPECT().
					FindByID(gomock.Any(), reportID).
					Return(&reportData.Report{
						ID:     reportID,
						Status: "Error",
					}, nil)
			},
			expectedSkip: true,
		},
		{
			name:     "Success - Don't skip report still processing",
			reportID: uuid.New(),
			mockSetup: func(reportID uuid.UUID) {
				mockReportDataRepo.EXPECT().
					FindByID(gomock.Any(), reportID).
					Return(&reportData.Report{
						ID:     reportID,
						Status: "Processing",
					}, nil)
			},
			expectedSkip: false,
		},
		{
			name:     "Success - Don't skip report not found (first attempt)",
			reportID: uuid.New(),
			mockSetup: func(reportID uuid.UUID) {
				mockReportDataRepo.EXPECT().
					FindByID(gomock.Any(), reportID).
					Return(nil, errors.New("not found"))
			},
			expectedSkip: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.reportID)

			result := useCase.shouldSkipProcessing(context.Background(), tt.reportID, logger)
			assert.Equal(t, tt.expectedSkip, result, "shouldSkipProcessing()")
		})
	}
}

func TestParseMessage_InvalidJSON(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())
	_, span := tracer.Start(context.Background(), "test")

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	// Test with invalid JSON
	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	_, err := useCase.parseMessage(context.Background(), []byte("invalid json"), &span, logger)
	require.Error(t, err, "expected error for invalid JSON")
}

func TestParseMessage_ValidJSON(t *testing.T) {
	t.Parallel()

	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())
	_, span := tracer.Start(context.Background(), "test")

	useCase := &UseCase{}

	templateID := uuid.New()
	reportID := uuid.New()

	body := GenerateReportMessage{
		TemplateID:   templateID,
		ReportID:     reportID,
		OutputFormat: "pdf",
		DataQueries:  map[string]map[string][]string{},
	}
	bodyBytes, _ := json.Marshal(body)

	message, err := useCase.parseMessage(context.Background(), bodyBytes, &span, logger)
	require.NoError(t, err)
	assert.Equal(t, templateID, message.TemplateID)
	assert.Equal(t, reportID, message.ReportID)
}

func TestGetTableFilters(t *testing.T) {
	t.Parallel()

	baseFilter := map[string]model.FilterCondition{
		"id": {Equals: []any{1, 2, 3}},
	}

	tests := []struct {
		name            string
		databaseFilters map[string]map[string]model.FilterCondition
		tableName       string
		expectNil       bool
	}{
		{
			name:            "Success - Nil database filters",
			databaseFilters: nil,
			tableName:       "users",
			expectNil:       true,
		},
		{
			name:            "Success - Table not found in filters",
			databaseFilters: map[string]map[string]model.FilterCondition{},
			tableName:       "users",
			expectNil:       true,
		},
		{
			name: "Success - Table found in filters exact match",
			databaseFilters: map[string]map[string]model.FilterCondition{
				"users": baseFilter,
			},
			tableName: "users",
			expectNil: false,
		},
		{
			name: "Success - Exact match Pongo2 format",
			databaseFilters: map[string]map[string]model.FilterCondition{
				"analytics__transfers": baseFilter,
			},
			tableName: "analytics__transfers",
			expectNil: false,
		},
		{
			name: "Success - Exact match qualified format",
			databaseFilters: map[string]map[string]model.FilterCondition{
				"analytics.transfers": baseFilter,
			},
			tableName: "analytics.transfers",
			expectNil: false,
		},
		{
			name: "Success - Cross-format match filter has dot table has Pongo2",
			databaseFilters: map[string]map[string]model.FilterCondition{
				"analytics.transfers": baseFilter,
			},
			tableName: "analytics__transfers",
			expectNil: false,
		},
		{
			name: "Success - Cross-format match filter has Pongo2 table has dot",
			databaseFilters: map[string]map[string]model.FilterCondition{
				"analytics__transfers": baseFilter,
			},
			tableName: "analytics.transfers",
			expectNil: false,
		},
		{
			name: "Success - No match different table names",
			databaseFilters: map[string]map[string]model.FilterCondition{
				"other_table": baseFilter,
			},
			tableName: "transfers",
			expectNil: true,
		},
		{
			name: "Success - Cross-format match filter has public.table template has just table",
			databaseFilters: map[string]map[string]model.FilterCondition{
				"public.organization": baseFilter,
			},
			tableName: "organization",
			expectNil: false,
		},
		{
			name: "Success - Cross-format match filter has public__table template has just table",
			databaseFilters: map[string]map[string]model.FilterCondition{
				"public__account": baseFilter,
			},
			tableName: "account",
			expectNil: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := getTableFilters(tt.databaseFilters, tt.tableName)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result, "expected non-nil result")
			}
		})
	}
}

func TestIsEncryptedField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		field    string
		expected bool
	}{
		{"document", true},
		{"name", true},
		{"email", false},
		{"id", false},
		{"contact", false},
		{"", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.field, func(t *testing.T) {
			t.Parallel()

			result := isEncryptedField(tt.field)
			assert.Equal(t, tt.expected, result, "isEncryptedField(%q)", tt.field)
		})
	}
}

func TestHashFilterValues(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey: hashKey,
		Logger:        logger,
	}

	useCase := &UseCase{}

	tests := []struct {
		name   string
		values []any
	}{
		{
			name:   "Success - Hash string values",
			values: []any{"value1", "value2"},
		},
		{
			name:   "Success - Keep non-string values",
			values: []any{123, 456.78, true},
		},
		{
			name:   "Success - Mixed values",
			values: []any{"string", 123, "another", nil},
		},
		{
			name:   "Success - Empty string value",
			values: []any{""},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := useCase.hashFilterValues(tt.values, crypto)
			require.Len(t, result, len(tt.values))

			for i, v := range tt.values {
				if strVal, ok := v.(string); ok && strVal != "" {
					expectedHash := crypto.GenerateHash(&strVal)
					assert.Equal(t, expectedHash, result[i], "value[%d]: expected hashed value", i)
				} else {
					assert.Equal(t, v, result[i], "value[%d]: expected unchanged value", i)
				}
			}
		})
	}
}

func TestDecryptContactFields(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	require.NoError(t, err, "Failed to initialize cipher")

	useCase := &UseCase{}

	tests := []struct {
		name           string
		record         map[string]any
		expectedEmails []string
		expectNoChange bool
	}{
		{
			name: "Success - Decrypt contact fields",
			record: func() map[string]any {
				email := "test@example.com"
				phone := "+1234567890"
				encrypted1, _ := crypto.Encrypt(&email)
				encrypted2, _ := crypto.Encrypt(&phone)
				return map[string]any{
					"contact": map[string]any{
						"primary_email": *encrypted1,
						"mobile_phone":  *encrypted2,
					},
				}
			}(),
			expectedEmails: []string{"test@example.com", "+1234567890"},
		},
		{
			name: "Success - No contact field present",
			record: map[string]any{
				"id": "test-id",
			},
			expectNoChange: true,
		},
		{
			name: "Success - Contact field is not a map",
			record: map[string]any{
				"contact": "not a map",
			},
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := useCase.decryptContactFields(tt.record, crypto)
			require.NoError(t, err)

			if !tt.expectNoChange && len(tt.expectedEmails) > 0 {
				contact, ok := tt.record["contact"].(map[string]any)
				require.True(t, ok, "contact not found or wrong type")
				assert.Equal(t, tt.expectedEmails[0], contact["primary_email"])
			}
		})
	}
}

func TestDecryptBankingDetailsFields(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	require.NoError(t, err, "Failed to initialize cipher")

	useCase := &UseCase{}

	tests := []struct {
		name            string
		record          map[string]any
		expectedAccount string
		expectNoChange  bool
	}{
		{
			name: "Success - Decrypt banking details fields",
			record: func() map[string]any {
				account := "12345-6"
				iban := "BR1234567890"
				encrypted1, _ := crypto.Encrypt(&account)
				encrypted2, _ := crypto.Encrypt(&iban)
				return map[string]any{
					"banking_details": map[string]any{
						"account": *encrypted1,
						"iban":    *encrypted2,
					},
				}
			}(),
			expectedAccount: "12345-6",
		},
		{
			name: "Success - No banking_details field present",
			record: map[string]any{
				"id": "test-id",
			},
			expectNoChange: true,
		},
		{
			name: "Success - banking_details field is not a map",
			record: map[string]any{
				"banking_details": "not a map",
			},
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := useCase.decryptBankingDetailsFields(tt.record, crypto)
			require.NoError(t, err)

			if !tt.expectNoChange && tt.expectedAccount != "" {
				bankingDetails, ok := tt.record["banking_details"].(map[string]any)
				require.True(t, ok, "banking_details not found or wrong type")
				assert.Equal(t, tt.expectedAccount, bankingDetails["account"])
			}
		})
	}
}

func TestDecryptLegalPersonFields(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	require.NoError(t, err, "Failed to initialize cipher")

	useCase := &UseCase{}

	tests := []struct {
		name           string
		record         map[string]any
		expectedName   string
		expectNoChange bool
	}{
		{
			name: "Success - Decrypt legal person representative fields",
			record: func() map[string]any {
				name := "John Doe"
				doc := "12345678901"
				encrypted1, _ := crypto.Encrypt(&name)
				encrypted2, _ := crypto.Encrypt(&doc)
				return map[string]any{
					"legal_person": map[string]any{
						"representative": map[string]any{
							"name":     *encrypted1,
							"document": *encrypted2,
						},
					},
				}
			}(),
			expectedName: "John Doe",
		},
		{
			name: "Success - No legal_person field present",
			record: map[string]any{
				"id": "test-id",
			},
			expectNoChange: true,
		},
		{
			name: "Success - legal_person without representative",
			record: map[string]any{
				"legal_person": map[string]any{
					"company_name": "Test Company",
				},
			},
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := useCase.decryptLegalPersonFields(tt.record, crypto)
			require.NoError(t, err)

			if !tt.expectNoChange && tt.expectedName != "" {
				legalPerson, ok := tt.record["legal_person"].(map[string]any)
				require.True(t, ok, "legal_person not found or wrong type")
				representative, ok := legalPerson["representative"].(map[string]any)
				require.True(t, ok, "representative not found or wrong type")
				assert.Equal(t, tt.expectedName, representative["name"])
			}
		})
	}
}

func TestDecryptNaturalPersonFields(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	require.NoError(t, err, "Failed to initialize cipher")

	useCase := &UseCase{}

	tests := []struct {
		name               string
		record             map[string]any
		expectedMotherName string
		expectNoChange     bool
	}{
		{
			name: "Success - Decrypt natural person fields",
			record: func() map[string]any {
				motherName := "Maria Silva"
				fatherName := "Jose Silva"
				encrypted1, _ := crypto.Encrypt(&motherName)
				encrypted2, _ := crypto.Encrypt(&fatherName)
				return map[string]any{
					"natural_person": map[string]any{
						"mother_name": *encrypted1,
						"father_name": *encrypted2,
					},
				}
			}(),
			expectedMotherName: "Maria Silva",
		},
		{
			name: "Success - No natural_person field present",
			record: map[string]any{
				"id": "test-id",
			},
			expectNoChange: true,
		},
		{
			name: "Success - natural_person field is not a map",
			record: map[string]any{
				"natural_person": "not a map",
			},
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := useCase.decryptNaturalPersonFields(tt.record, crypto)
			require.NoError(t, err)

			if !tt.expectNoChange && tt.expectedMotherName != "" {
				naturalPerson, ok := tt.record["natural_person"].(map[string]any)
				require.True(t, ok, "natural_person not found or wrong type")
				assert.Equal(t, tt.expectedMotherName, naturalPerson["mother_name"])
			}
		})
	}
}

func TestDecryptFieldValue(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	require.NoError(t, err, "Failed to initialize cipher")

	useCase := &UseCase{}

	tests := []struct {
		name        string
		container   map[string]any
		fieldName   string
		fieldValue  any
		expectError bool
	}{
		{
			name:       "Success - Decrypt valid string",
			container:  map[string]any{},
			fieldName:  "test_field",
			fieldValue: func() string { v := "test"; e, _ := crypto.Encrypt(&v); return *e }(),
		},
		{
			name:       "Success - Skip non-string value",
			container:  map[string]any{},
			fieldName:  "test_field",
			fieldValue: 123,
		},
		{
			name:       "Success - Skip empty string",
			container:  map[string]any{},
			fieldName:  "test_field",
			fieldValue: "",
		},
		{
			name:        "Error - invalid encrypted value",
			container:   map[string]any{},
			fieldName:   "test_field",
			fieldValue:  "not-encrypted-data",
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := useCase.decryptFieldValue(tt.container, tt.fieldName, tt.fieldValue, crypto)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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

func TestQueryDatabase_UnknownDataSource(t *testing.T) {
	t.Parallel()

	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())

	useCase := &UseCase{
		ExternalDataSources:   pkg.NewSafeDataSources(map[string]pkg.DataSource{}),
		CircuitBreakerManager: pkg.NewCircuitBreakerManager(logger),
	}

	result := make(map[string]map[string][]map[string]any)

	err := useCase.queryDatabase(
		context.Background(),
		"unknown_db",
		map[string][]string{"table": {"field"}},
		nil,
		result,
		logger,
		tracer,
	)
	// Unknown data source should not return error, just skip
	require.NoError(t, err, "expected nil error for unknown data source")
}

func TestQueryDatabase_CircuitBreakerUnhealthy(t *testing.T) {
	t.Parallel()

	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())

	cbManager := pkg.NewCircuitBreakerManager(logger)

	// Force circuit breaker to open state by recording failures
	for i := 0; i < 10; i++ {
		_, _ = cbManager.Execute("test_db", func() (any, error) {
			return nil, errors.New("simulated failure")
		})
	}

	useCase := &UseCase{
		ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
			"test_db": {
				Initialized:  true,
				DatabaseType: "postgresql",
			},
		}),
		CircuitBreakerManager: cbManager,
	}

	result := make(map[string]map[string][]map[string]any)

	err := useCase.queryDatabase(
		context.Background(),
		"test_db",
		map[string][]string{"table": {"field"}},
		nil,
		result,
		logger,
		tracer,
	)

	require.Error(t, err, "expected error when circuit breaker is unhealthy")
	assert.Contains(t, err.Error(), "circuit breaker")
}

func TestQueryDatabase_UnsupportedDatabaseType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())

	useCase := &UseCase{
		ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
			"test_db": {
				Initialized:  true,
				DatabaseType: "unsupported_type",
			},
		}),
		CircuitBreakerManager: pkg.NewCircuitBreakerManager(logger),
	}

	result := make(map[string]map[string][]map[string]any)

	err := useCase.queryDatabase(
		context.Background(),
		"test_db",
		map[string][]string{"table": {"field"}},
		nil,
		result,
		logger,
		tracer,
	)

	require.Error(t, err, "expected error for unsupported database type")
	assert.Contains(t, err.Error(), "unsupported database type")
}

func TestTransformPluginCRMAdvancedFilters_NilFilter(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	useCase := &UseCase{}

	result, err := useCase.transformPluginCRMAdvancedFilters(nil, logger)
	require.NoError(t, err)
	assert.Nil(t, result, "expected nil result for nil input")
}

func TestTransformPluginCRMAdvancedFilters_MissingEnvVar(t *testing.T) {
	// NOTE: t.Setenv is incompatible with t.Parallel()
	t.Setenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM", "")

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	useCase := &UseCase{}

	filter := map[string]model.FilterCondition{
		"document": {
			Equals: []any{"12345678901"},
		},
	}

	_, err := useCase.transformPluginCRMAdvancedFilters(filter, logger)
	require.Error(t, err, "expected error when env var is missing")
	assert.Contains(t, err.Error(), "CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")
}

func TestTransformPluginCRMAdvancedFilters_AllFilterConditions(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	useCase := &UseCase{
		CryptoHashSecretKeyPluginCRM: hashKey,
	}

	filter := map[string]model.FilterCondition{
		"document": {
			Equals:         []any{"value1"},
			GreaterThan:    []any{"value2"},
			GreaterOrEqual: []any{"value3"},
			LessThan:       []any{"value4"},
			LessOrEqual:    []any{"value5"},
			Between:        []any{"value6", "value7"},
			In:             []any{"value8"},
			NotIn:          []any{"value9"},
		},
	}

	result, err := useCase.transformPluginCRMAdvancedFilters(filter, logger)
	require.NoError(t, err)

	assert.Contains(t, result, "search.document", "expected search.document field in result")

	// Verify all conditions were transformed
	searchDoc := result["search.document"]
	assert.NotEmpty(t, searchDoc.Equals, "expected Equals to be transformed")
	assert.NotEmpty(t, searchDoc.GreaterThan, "expected GreaterThan to be transformed")
	assert.NotEmpty(t, searchDoc.GreaterOrEqual, "expected GreaterOrEqual to be transformed")
	assert.NotEmpty(t, searchDoc.LessThan, "expected LessThan to be transformed")
	assert.NotEmpty(t, searchDoc.LessOrEqual, "expected LessOrEqual to be transformed")
	assert.NotEmpty(t, searchDoc.Between, "expected Between to be transformed")
	assert.NotEmpty(t, searchDoc.In, "expected In to be transformed")
	assert.NotEmpty(t, searchDoc.NotIn, "expected NotIn to be transformed")
}

func TestTransformPluginCRMAdvancedFilters_NonMappedField(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	useCase := &UseCase{
		CryptoHashSecretKeyPluginCRM: hashKey,
	}

	filter := map[string]model.FilterCondition{
		"unmapped_field": {
			Equals: []any{"value1"},
		},
	}

	result, err := useCase.transformPluginCRMAdvancedFilters(filter, logger)
	require.NoError(t, err)

	// Non-mapped fields should be kept as-is
	assert.Contains(t, result, "unmapped_field", "expected unmapped_field to be preserved")
}

func TestGenerateReport_ReportAlreadyFinished(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Report is already finished - should skip processing
	mockReportDataRepo.
		EXPECT().
		FindByID(gomock.Any(), reportID).
		Return(&reportData.Report{
			ID:     reportID,
			Status: "Finished",
		}, nil)

	useCase := &UseCase{
		ReportDataRepo:      mockReportDataRepo,
		ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{}),
	}

	err := useCase.GenerateReport(context.Background(), bodyBytes)
	require.NoError(t, err, "expected no error for already finished report")
}

func TestUpdateReportWithErrors(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)

	tests := []struct {
		name         string
		reportID     uuid.UUID
		errorMessage string
		mockSetup    func(reportID uuid.UUID)
		expectError  bool
	}{
		{
			name:         "Success - Update report with error",
			reportID:     uuid.New(),
			errorMessage: "test error message",
			mockSetup: func(reportID uuid.UUID) {
				mockReportDataRepo.EXPECT().
					UpdateReportStatusById(gomock.Any(), "Error", reportID, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:         "Error - Failed to update report",
			reportID:     uuid.New(),
			errorMessage: "test error message",
			mockSetup: func(reportID uuid.UUID) {
				mockReportDataRepo.EXPECT().
					UpdateReportStatusById(gomock.Any(), "Error", reportID, gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.reportID)

			useCase := &UseCase{
				ReportDataRepo: mockReportDataRepo,
			}

			err := useCase.updateReportWithErrors(context.Background(), tt.reportID, tt.errorMessage)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMarkReportAsFinished(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())
	_, span := tracer.Start(context.Background(), "test")

	tests := []struct {
		name        string
		reportID    uuid.UUID
		mockSetup   func(reportID uuid.UUID)
		expectError bool
	}{
		{
			name:     "Success - Mark report as finished",
			reportID: uuid.New(),
			mockSetup: func(reportID uuid.UUID) {
				mockReportDataRepo.EXPECT().
					UpdateReportStatusById(gomock.Any(), "Finished", reportID, gomock.Any(), nil).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:     "Error - Failed to mark as finished",
			reportID: uuid.New(),
			mockSetup: func(reportID uuid.UUID) {
				mockReportDataRepo.EXPECT().
					UpdateReportStatusById(gomock.Any(), "Finished", reportID, gomock.Any(), nil).
					Return(errors.New("database error"))
				mockReportDataRepo.EXPECT().
					UpdateReportStatusById(gomock.Any(), "Error", reportID, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.reportID)

			useCase := &UseCase{
				ReportDataRepo: mockReportDataRepo,
			}

			err := useCase.markReportAsFinished(context.Background(), tt.reportID, &span, logger)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckReportStatus(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	tests := []struct {
		name           string
		reportID       uuid.UUID
		mockSetup      func(reportID uuid.UUID)
		expectedStatus string
		expectError    bool
	}{
		{
			name:     "Success - Get report status",
			reportID: uuid.New(),
			mockSetup: func(reportID uuid.UUID) {
				mockReportDataRepo.EXPECT().
					FindByID(gomock.Any(), reportID).
					Return(&reportData.Report{
						ID:     reportID,
						Status: "Processing",
					}, nil)
			},
			expectedStatus: "Processing",
			expectError:    false,
		},
		{
			name:     "Error - Report not found",
			reportID: uuid.New(),
			mockSetup: func(reportID uuid.UUID) {
				mockReportDataRepo.EXPECT().
					FindByID(gomock.Any(), reportID).
					Return(nil, errors.New("not found"))
			},
			expectedStatus: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.reportID)

			useCase := &UseCase{
				ReportDataRepo: mockReportDataRepo,
			}

			status, err := useCase.checkReportStatus(context.Background(), tt.reportID, logger)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestSaveReport_WithTTL(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportRepo := report.NewMockRepository(ctrl)

	useCase := &UseCase{
		ReportSeaweedFS: mockReportRepo,
		ReportTTL:       "30d",
	}

	reportID := uuid.New()
	message := GenerateReportMessage{
		ReportID:     reportID,
		TemplateID:   uuid.New(),
		OutputFormat: "json",
	}
	renderedOutput := `{"data": "test"}`

	mockReportRepo.
		EXPECT().
		Put(gomock.Any(), gomock.Any(), "application/json", []byte(renderedOutput), "30d").
		Return(nil)

	ctx := context.Background()
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)

	err := useCase.saveReport(ctx, tracer, message, renderedOutput, logger)
	require.NoError(t, err)
}

func TestQueryPostgresDatabase_SchemaFormats(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostgresRepo := postgres2.NewMockRepository(ctrl)
	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	tests := []struct {
		name      string
		tableKey  string
		mockSetup func()
	}{
		{
			name:     "Success - Pongo2 format schema__table",
			tableKey: "custom_schema__users",
			mockSetup: func() {
				mockPostgresRepo.EXPECT().
					GetDatabaseSchema(gomock.Any(), []string{"custom_schema"}).
					Return([]postgres2.TableSchema{
						{
							SchemaName: "custom_schema",
							TableName:  "users",
							Columns: []postgres2.ColumnInformation{
								{Name: "id", DataType: "integer"},
								{Name: "name", DataType: "text"},
							},
						},
					}, nil)

				mockPostgresRepo.EXPECT().
					Query(gomock.Any(), gomock.Any(), "custom_schema", "users", []string{"name"}, nil).
					Return([]map[string]any{{"name": "John"}}, nil)
			},
		},
		{
			name:     "Success - Qualified format schema.table",
			tableKey: "other_schema.products",
			mockSetup: func() {
				mockPostgresRepo.EXPECT().
					GetDatabaseSchema(gomock.Any(), []string{"other_schema"}).
					Return([]postgres2.TableSchema{
						{
							SchemaName: "other_schema",
							TableName:  "products",
							Columns: []postgres2.ColumnInformation{
								{Name: "id", DataType: "integer"},
								{Name: "name", DataType: "text"},
							},
						},
					}, nil)

				mockPostgresRepo.EXPECT().
					Query(gomock.Any(), gomock.Any(), "other_schema", "products", []string{"name"}, nil).
					Return([]map[string]any{{"name": "Product1"}}, nil)
			},
		},
		{
			name:     "Success - Legacy format table only (autodiscovery)",
			tableKey: "orders",
			mockSetup: func() {
				mockPostgresRepo.EXPECT().
					GetDatabaseSchema(gomock.Any(), []string{"public"}).
					Return([]postgres2.TableSchema{
						{
							SchemaName: "public",
							TableName:  "orders",
							Columns: []postgres2.ColumnInformation{
								{Name: "id", DataType: "integer"},
								{Name: "total", DataType: "numeric"},
							},
						},
					}, nil)

				mockPostgresRepo.EXPECT().
					Query(gomock.Any(), gomock.Any(), "public", "orders", []string{"total"}, nil).
					Return([]map[string]any{{"total": 100.50}}, nil)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			cbManager := pkg.NewCircuitBreakerManager(logger)

			dataSource := &pkg.DataSource{
				Initialized:        true,
				DatabaseType:       "postgresql",
				PostgresRepository: mockPostgresRepo,
			}

			// Extract schema from tableKey for configuring schemas
			var schemas []string
			if strings.Contains(tt.tableKey, "__") {
				parts := strings.SplitN(tt.tableKey, "__", 2)
				schemas = []string{parts[0]}
			} else if strings.Contains(tt.tableKey, ".") {
				parts := strings.SplitN(tt.tableKey, ".", 2)
				schemas = []string{parts[0]}
			} else {
				schemas = []string{"public"}
			}
			dataSource.Schemas = schemas

			useCase := &UseCase{
				CircuitBreakerManager: cbManager,
			}

			result := make(map[string]map[string][]map[string]any)
			result["test_db"] = make(map[string][]map[string]any)

			tables := map[string][]string{
				tt.tableKey: {"name"},
			}
			if tt.tableKey == "orders" {
				tables = map[string][]string{
					tt.tableKey: {"total"},
				}
			}

			err := useCase.queryPostgresDatabase(
				context.Background(),
				dataSource,
				"test_db",
				tables,
				nil,
				result,
				logger,
			)
			require.NoError(t, err)
		})
	}
}

func TestDecryptPluginCRMData_MissingEnvVars(t *testing.T) {
	// NOTE: t.Setenv is incompatible with t.Parallel()
	t.Setenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM", "")
	t.Setenv("CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM", "")

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	useCase := &UseCase{}

	collectionResult := []map[string]any{
		{"document": "encrypted_value"},
	}

	_, err := useCase.decryptPluginCRMData(logger, collectionResult, []string{"document"})
	require.Error(t, err, "expected error when env vars are missing")
}

func TestDecryptPluginCRMData_NoDecryptionNeeded(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	useCase := &UseCase{}

	collectionResult := []map[string]any{
		{"id": "123", "status": "active"},
	}

	result, err := useCase.decryptPluginCRMData(logger, collectionResult, []string{"id", "status"})
	require.NoError(t, err)
	require.Len(t, result, 1)
}

func TestHandleErrorWithUpdate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())
	_, span := tracer.Start(context.Background(), "test")

	tests := []struct {
		name        string
		reportID    uuid.UUID
		errorMsg    string
		inputErr    error
		mockSetup   func(reportID uuid.UUID)
		expectError bool
	}{
		{
			name:     "Success - Log error and update report",
			reportID: uuid.New(),
			errorMsg: "Test error message",
			inputErr: errors.New("original error"),
			mockSetup: func(reportID uuid.UUID) {
				mockReportDataRepo.EXPECT().
					UpdateReportStatusById(gomock.Any(), "Error", reportID, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: true, // Returns the original error
		},
		{
			name:     "Error - Failed to update report status",
			reportID: uuid.New(),
			errorMsg: "Test error message",
			inputErr: errors.New("original error"),
			mockSetup: func(reportID uuid.UUID) {
				mockReportDataRepo.EXPECT().
					UpdateReportStatusById(gomock.Any(), "Error", reportID, gomock.Any(), gomock.Any()).
					Return(errors.New("update failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.reportID)

			useCase := &UseCase{
				ReportDataRepo: mockReportDataRepo,
			}

			err := useCase.handleErrorWithUpdate(context.Background(), tt.reportID, &span, tt.errorMsg, tt.inputErr, logger)
			if tt.expectError {
				require.Error(t, err)
			}
		})
	}
}

func TestDecryptNestedFields_AllTypes(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	require.NoError(t, err, "Failed to initialize cipher")

	useCase := &UseCase{}

	// Create a record with all nested field types
	email := "test@example.com"
	account := "12345-6"
	repName := "John Doe"
	motherName := "Maria"
	participantDoc := "12345678901234"
	partyDoc := "11111111111"

	encEmail, _ := crypto.Encrypt(&email)
	encAccount, _ := crypto.Encrypt(&account)
	encRepName, _ := crypto.Encrypt(&repName)
	encMotherName, _ := crypto.Encrypt(&motherName)
	encParticipantDoc, _ := crypto.Encrypt(&participantDoc)
	encPartyDoc, _ := crypto.Encrypt(&partyDoc)

	record := map[string]any{
		"contact": map[string]any{
			"primary_email": *encEmail,
		},
		"banking_details": map[string]any{
			"account": *encAccount,
		},
		"legal_person": map[string]any{
			"representative": map[string]any{
				"name": *encRepName,
			},
		},
		"natural_person": map[string]any{
			"mother_name": *encMotherName,
		},
		"regulatory_fields": map[string]any{
			"participant_document": *encParticipantDoc,
		},
		"related_parties": []any{
			map[string]any{
				"document": *encPartyDoc,
			},
		},
	}

	err = useCase.decryptNestedFields(record, crypto)
	require.NoError(t, err)

	// Verify all fields were decrypted
	contact := record["contact"].(map[string]any)
	assert.Equal(t, email, contact["primary_email"])

	bankingDetails := record["banking_details"].(map[string]any)
	assert.Equal(t, account, bankingDetails["account"])

	legalPerson := record["legal_person"].(map[string]any)
	representative := legalPerson["representative"].(map[string]any)
	assert.Equal(t, repName, representative["name"])

	naturalPerson := record["natural_person"].(map[string]any)
	assert.Equal(t, motherName, naturalPerson["mother_name"])

	regulatoryFields := record["regulatory_fields"].(map[string]any)
	assert.Equal(t, participantDoc, regulatoryFields["participant_document"])

	relatedParties := record["related_parties"].([]any)
	party := relatedParties[0].(map[string]any)
	assert.Equal(t, partyDoc, party["document"])
}

func TestDecryptRecord(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	require.NoError(t, err, "Failed to initialize cipher")

	useCase := &UseCase{}

	tests := []struct {
		name           string
		record         map[string]any
		expectedFields map[string]any
	}{
		{
			name: "Success - Decrypt record with top-level encrypted fields",
			record: func() map[string]any {
				doc := "12345678901"
				name := "John Doe"
				encDoc, _ := crypto.Encrypt(&doc)
				encName, _ := crypto.Encrypt(&name)
				return map[string]any{
					"document": *encDoc,
					"name":     *encName,
					"id":       "123",
				}
			}(),
			expectedFields: map[string]any{
				"document": "12345678901",
				"name":     "John Doe",
				"id":       "123",
			},
		},
		{
			name: "Success - Decrypt record with no encrypted fields",
			record: map[string]any{
				"id":     "123",
				"status": "active",
			},
			expectedFields: map[string]any{
				"id":     "123",
				"status": "active",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := useCase.decryptRecord(tt.record, crypto)
			require.NoError(t, err)

			for key, expected := range tt.expectedFields {
				assert.Equal(t, expected, result[key], "field %s", key)
			}
		})
	}
}

func TestDecryptTopLevelFields(t *testing.T) {
	t.Parallel()

	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	require.NoError(t, err, "Failed to initialize cipher")

	useCase := &UseCase{}

	tests := []struct {
		name           string
		record         map[string]any
		expectedDoc    string
		expectNoChange bool
	}{
		{
			name: "Success - Decrypt document and name fields",
			record: func() map[string]any {
				doc := "12345678901"
				name := "John Doe"
				encDoc, _ := crypto.Encrypt(&doc)
				encName, _ := crypto.Encrypt(&name)
				return map[string]any{
					"document": *encDoc,
					"name":     *encName,
				}
			}(),
			expectedDoc: "12345678901",
		},
		{
			name: "Success - No encrypted fields present",
			record: map[string]any{
				"id":     "123",
				"status": "active",
			},
			expectNoChange: true,
		},
		{
			name: "Success - Encrypted field with nil value",
			record: map[string]any{
				"document": nil,
				"name":     nil,
			},
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := useCase.decryptTopLevelFields(tt.record, crypto)
			require.NoError(t, err)

			if !tt.expectNoChange && tt.expectedDoc != "" {
				assert.Equal(t, tt.expectedDoc, tt.record["document"])
			}
		})
	}
}

func TestGenerateReport_ReportInErrorState(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Report is in error state - should skip processing
	mockReportDataRepo.
		EXPECT().
		FindByID(gomock.Any(), reportID).
		Return(&reportData.Report{
			ID:     reportID,
			Status: "Error",
		}, nil)

	useCase := &UseCase{
		ReportDataRepo:      mockReportDataRepo,
		ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{}),
	}

	err := useCase.GenerateReport(context.Background(), bodyBytes)
	require.NoError(t, err, "expected no error for report in error state")
}

func TestQueryMongoDatabase_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMongoRepo := mongodb2.NewMockRepository(ctrl)
	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	cbManager := pkg.NewCircuitBreakerManager(logger)

	mockMongoRepo.EXPECT().
		Query(gomock.Any(), "users", []string{"name", "email"}, nil).
		Return([]map[string]any{
			{"name": "John", "email": "john@example.com"},
		}, nil)

	dataSource := &pkg.DataSource{
		Initialized:       true,
		DatabaseType:      "mongodb",
		MongoDBRepository: mockMongoRepo,
	}

	useCase := &UseCase{
		CircuitBreakerManager: cbManager,
	}

	result := make(map[string]map[string][]map[string]any)
	result["test_db"] = make(map[string][]map[string]any)

	err := useCase.queryMongoDatabase(
		context.Background(),
		dataSource,
		"test_db",
		map[string][]string{"users": {"name", "email"}},
		nil,
		result,
		logger,
	)
	require.NoError(t, err)
	require.Len(t, result["test_db"]["users"], 1)
}

func TestQueryMongoDatabase_WithFilters(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMongoRepo := mongodb2.NewMockRepository(ctrl)
	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	cbManager := pkg.NewCircuitBreakerManager(logger)

	filters := map[string]map[string]model.FilterCondition{
		"users": {
			"status": {
				Equals: []any{"active"},
			},
		},
	}

	mockMongoRepo.EXPECT().
		QueryWithAdvancedFilters(gomock.Any(), "users", []string{"name"}, gomock.Any()).
		Return([]map[string]any{
			{"name": "Active User"},
		}, nil)

	dataSource := &pkg.DataSource{
		Initialized:       true,
		DatabaseType:      "mongodb",
		MongoDBRepository: mockMongoRepo,
	}

	useCase := &UseCase{
		CircuitBreakerManager: cbManager,
	}

	result := make(map[string]map[string][]map[string]any)
	result["test_db"] = make(map[string][]map[string]any)

	err := useCase.queryMongoDatabase(
		context.Background(),
		dataSource,
		"test_db",
		map[string][]string{"users": {"name"}},
		filters,
		result,
		logger,
	)
	require.NoError(t, err)
}

func TestProcessRegularMongoCollection(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMongoRepo := mongodb2.NewMockRepository(ctrl)
	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	cbManager := pkg.NewCircuitBreakerManager(logger)

	mockMongoRepo.EXPECT().
		Query(gomock.Any(), "products", []string{"name", "price"}, nil).
		Return([]map[string]any{
			{"name": "Product 1", "price": 100},
		}, nil)

	dataSource := &pkg.DataSource{
		Initialized:       true,
		DatabaseType:      "mongodb",
		MongoDBRepository: mockMongoRepo,
	}

	useCase := &UseCase{
		CircuitBreakerManager: cbManager,
	}

	result := make(map[string]map[string][]map[string]any)
	result["shop_db"] = make(map[string][]map[string]any)

	err := useCase.processRegularMongoCollection(
		context.Background(),
		dataSource,
		"shop_db",
		"products",
		[]string{"name", "price"},
		nil,
		result,
		logger,
	)
	require.NoError(t, err)
	require.Len(t, result["shop_db"]["products"], 1)
}

func TestLoadTemplate_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTemplateRepo := template.NewMockRepository(ctrl)
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())
	_, span := tracer.Start(context.Background(), "test")

	templateID := uuid.New()
	templateContent := []byte("Hello {{ name }}")

	mockTemplateRepo.EXPECT().
		Get(gomock.Any(), templateID.String()).
		Return(templateContent, nil)

	useCase := &UseCase{
		TemplateSeaweedFS: mockTemplateRepo,
	}

	message := GenerateReportMessage{
		TemplateID: templateID,
		ReportID:   uuid.New(),
	}

	result, err := useCase.loadTemplate(context.Background(), tracer, message, &span, logger)
	require.NoError(t, err)
	assert.Equal(t, string(templateContent), string(result))
}

func TestLoadTemplate_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTemplateRepo := template.NewMockRepository(ctrl)
	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())
	_, span := tracer.Start(context.Background(), "test")

	templateID := uuid.New()
	reportID := uuid.New()

	mockTemplateRepo.EXPECT().
		Get(gomock.Any(), templateID.String()).
		Return(nil, errors.New("template not found"))

	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), "Error", reportID, gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := &UseCase{
		TemplateSeaweedFS: mockTemplateRepo,
		ReportDataRepo:    mockReportDataRepo,
	}

	message := GenerateReportMessage{
		TemplateID: templateID,
		ReportID:   reportID,
	}

	_, err := useCase.loadTemplate(context.Background(), tracer, message, &span, logger)
	require.Error(t, err)
}

func TestQueryExternalData_NoDataSources(t *testing.T) {
	t.Parallel()

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	cbManager := pkg.NewCircuitBreakerManager(logger)

	useCase := &UseCase{
		ExternalDataSources:   pkg.NewSafeDataSources(map[string]pkg.DataSource{}),
		CircuitBreakerManager: cbManager,
	}

	message := GenerateReportMessage{
		TemplateID:   uuid.New(),
		ReportID:     uuid.New(),
		OutputFormat: "txt",
		DataQueries:  map[string]map[string][]string{},
	}

	result := make(map[string]map[string][]map[string]any)

	err := useCase.queryExternalData(context.Background(), message, result)
	require.NoError(t, err)
	assert.Empty(t, result, "expected empty result")
}
