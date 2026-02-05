// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/model"
	mongodb2 "github.com/LerianStudio/reporter/v4/pkg/mongodb"
	reportData "github.com/LerianStudio/reporter/v4/pkg/mongodb/report"
	postgres2 "github.com/LerianStudio/reporter/v4/pkg/postgres"
	"github.com/LerianStudio/reporter/v4/pkg/seaweedfs/report"
	"github.com/LerianStudio/reporter/v4/pkg/seaweedfs/template"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libCrypto "github.com/LerianStudio/lib-commons/v2/commons/crypto"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
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
	if err == nil || !strings.Contains(err.Error(), "failed to put file") {
		t.Errorf("expected error on Put, got: %v", err)
	}
}

func TestGenerateReport_PluginCRMWithEncryptedData(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	os.Setenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM", hashKey)
	os.Setenv("CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM", encryptKey)
	defer func() {
		os.Unsetenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")
		os.Unsetenv("CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM")
	}()

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
	if err != nil {
		t.Fatalf("Failed to initialize cipher: %v", err)
	}

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
				"organization": {organizationID},
				"holders":      {"name", "document", "contact.primary_email", "banking_details.account"},
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
		Query(
			gomock.Any(),
			"organization",
			[]string{organizationID},
			nil,
		).
		Return([]map[string]any{{"id": organizationID}}, nil)

	mockMongoRepo.
		EXPECT().
		QueryWithAdvancedFilters(
			gomock.Any(),
			"holders_"+organizationID,
			[]string{"name", "document", "contact.primary_email", "banking_details.account"},
			gomock.Any(),
		).
		DoAndReturn(func(ctx context.Context, collection string, fields []string, filters map[string]model.FilterCondition) ([]map[string]any, error) {
			if searchDocFilter, exists := filters["search.document"]; exists {
				if len(searchDocFilter.Equals) > 0 {
					if searchDocFilter.Equals[0] != hashedDocument {
						t.Errorf("Expected hashed document %s, got %s", hashedDocument, searchDocFilter.Equals[0])
					}
				}
			} else {
				t.Error("Expected search.document filter to be present")
			}
			return mockMongoData, nil
		})

	mockReportRepo.
		EXPECT().
		Put(gomock.Any(), gomock.Any(), "text/html", gomock.Any(), "").
		DoAndReturn(func(ctx context.Context, objectName, contentType string, data []byte, ttl string) error {
			// Verificar se o conteúdo foi renderizado com dados descriptografados
			content := string(data)
			if !strings.Contains(content, "João Silva") {
				t.Error("Expected decrypted name 'João Silva' in rendered content")
			}
			if !strings.Contains(content, testDocument) {
				t.Error("Expected decrypted document in rendered content")
			}
			if !strings.Contains(content, "joao@example.com") {
				t.Error("Expected decrypted email in rendered content")
			}
			if !strings.Contains(content, "12345-6") {
				t.Error("Expected decrypted account in rendered content")
			}
			return nil
		})

	mockReportDataRepo.
		EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), nil).
		Return(nil)

	circuitBreakerManager := pkg.NewCircuitBreakerManager(logger)

	useCase := &UseCase{
		TemplateSeaweedFS:     mockTemplateRepo,
		ReportSeaweedFS:       mockReportRepo,
		ReportDataRepo:        mockReportDataRepo,
		CircuitBreakerManager: circuitBreakerManager,
		ExternalDataSources: map[string]pkg.DataSource{
			"plugin_crm": {
				Initialized:       true,
				DatabaseType:      "mongodb",
				MongoDBRepository: mockMongoRepo,
			},
		},
	}

	err = useCase.GenerateReport(context.Background(), bodyBytes)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDecryptRegulatoryFieldsFields(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	if err != nil {
		t.Fatalf("Failed to initialize cipher: %v", err)
	}

	useCase := &UseCase{}

	tests := []struct {
		name           string
		record         map[string]any
		expectedDoc    string
		expectNoChange bool
	}{
		{
			name: "decrypt regulatory_fields.participant_document",
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
			name: "no regulatory_fields present",
			record: map[string]any{
				"id": "test-id",
			},
			expectNoChange: true,
		},
		{
			name: "regulatory_fields without participant_document",
			record: map[string]any{
				"regulatory_fields": map[string]any{
					"other_field": "value",
				},
			},
			expectNoChange: true,
		},
		{
			name: "regulatory_fields with nil participant_document",
			record: map[string]any{
				"regulatory_fields": map[string]any{
					"participant_document": nil,
				},
			},
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := useCase.decryptRegulatoryFieldsFields(tt.record, crypto)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectNoChange {
				regulatoryFields, ok := tt.record["regulatory_fields"].(map[string]any)
				if !ok {
					t.Fatal("regulatory_fields not found or wrong type")
				}
				if regulatoryFields["participant_document"] != tt.expectedDoc {
					t.Errorf("expected participant_document = %q, got %q", tt.expectedDoc, regulatoryFields["participant_document"])
				}
			}
		})
	}
}

func TestDecryptRelatedPartiesFields(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	if err != nil {
		t.Fatalf("Failed to initialize cipher: %v", err)
	}

	useCase := &UseCase{}

	tests := []struct {
		name           string
		record         map[string]any
		expectedDocs   []string
		expectNoChange bool
	}{
		{
			name: "decrypt multiple related_parties documents",
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
			name: "no related_parties present",
			record: map[string]any{
				"id": "test-id",
			},
			expectNoChange: true,
		},
		{
			name: "empty related_parties array",
			record: map[string]any{
				"related_parties": []any{},
			},
			expectNoChange: true,
		},
		{
			name: "related_parties with nil document",
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
			name: "related_parties with mixed valid and nil documents",
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
		t.Run(tt.name, func(t *testing.T) {
			err := useCase.decryptRelatedPartiesFields(tt.record, crypto)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectNoChange && len(tt.expectedDocs) > 0 {
				relatedParties, ok := tt.record["related_parties"].([]any)
				if !ok {
					t.Fatal("related_parties not found or wrong type")
				}

				docIndex := 0
				for i, party := range relatedParties {
					partyMap, ok := party.(map[string]any)
					if !ok {
						t.Fatalf("related_parties[%d] is not a map", i)
					}

					if partyMap["document"] != nil && docIndex < len(tt.expectedDocs) {
						if partyMap["document"] != tt.expectedDocs[docIndex] {
							t.Errorf("expected related_parties[%d].document = %q, got %q", i, tt.expectedDocs[docIndex], partyMap["document"])
						}
						docIndex++
					}
				}
			}
		})
	}
}

func TestTransformPluginCRMAdvancedFilters_NewFields(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	os.Setenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM", hashKey)
	defer os.Unsetenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey: hashKey,
		Logger:        logger,
	}

	useCase := &UseCase{}

	tests := []struct {
		name          string
		inputField    string
		expectedField string
		inputValue    string
	}{
		{
			name:          "transform regulatory_fields.participant_document",
			inputField:    "regulatory_fields.participant_document",
			expectedField: "search.regulatory_fields_participant_document",
			inputValue:    "12345678901234",
		},
		{
			name:          "transform related_parties.document",
			inputField:    "related_parties.document",
			expectedField: "search.related_party_documents",
			inputValue:    "11111111111",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := map[string]model.FilterCondition{
				tt.inputField: {
					Equals: []any{tt.inputValue},
				},
			}

			transformedFilter, err := useCase.transformPluginCRMAdvancedFilters(filter, logger)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the field was transformed
			if _, exists := transformedFilter[tt.expectedField]; !exists {
				t.Errorf("expected field %q not found in transformed filter", tt.expectedField)
			}

			// Verify the original field was removed
			if _, exists := transformedFilter[tt.inputField]; exists {
				t.Errorf("original field %q should not exist in transformed filter", tt.inputField)
			}

			// Verify the value was hashed
			expectedHash := crypto.GenerateHash(&tt.inputValue)
			if transformedFilter[tt.expectedField].Equals[0] != expectedHash {
				t.Errorf("expected hashed value %q, got %q", expectedHash, transformedFilter[tt.expectedField].Equals[0])
			}
		})
	}
}

func TestShouldSkipProcessing(t *testing.T) {
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
			name:     "Skip - Report already finished",
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
			name:     "Skip - Report in error state",
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
			name:     "Don't skip - Report still processing",
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
			name:     "Don't skip - Report not found (first attempt)",
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
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.reportID)

			result := useCase.shouldSkipProcessing(context.Background(), tt.reportID, logger)
			if result != tt.expectedSkip {
				t.Errorf("shouldSkipProcessing() = %v, want %v", result, tt.expectedSkip)
			}
		})
	}
}

func TestParseMessage_InvalidJSON(t *testing.T) {
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
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseMessage_ValidJSON(t *testing.T) {
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
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if message.TemplateID != templateID {
		t.Errorf("expected templateID %s, got %s", templateID, message.TemplateID)
	}
	if message.ReportID != reportID {
		t.Errorf("expected reportID %s, got %s", reportID, message.ReportID)
	}
}

func TestGetTableFilters(t *testing.T) {
	tests := []struct {
		name            string
		databaseFilters map[string]map[string]model.FilterCondition
		tableName       string
		expectNil       bool
	}{
		{
			name:            "Nil database filters",
			databaseFilters: nil,
			tableName:       "users",
			expectNil:       true,
		},
		{
			name:            "Table not found in filters",
			databaseFilters: map[string]map[string]model.FilterCondition{},
			tableName:       "users",
			expectNil:       true,
		},
		{
			name: "Table found in filters",
			databaseFilters: map[string]map[string]model.FilterCondition{
				"users": {
					"id": {
						Equals: []any{1, 2, 3},
					},
				},
			},
			tableName: "users",
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTableFilters(tt.databaseFilters, tt.tableName)
			if tt.expectNil && result != nil {
				t.Errorf("expected nil, got %v", result)
			}
			if !tt.expectNil && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestIsEncryptedField(t *testing.T) {
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
		t.Run(tt.field, func(t *testing.T) {
			result := isEncryptedField(tt.field)
			if result != tt.expected {
				t.Errorf("isEncryptedField(%q) = %v, want %v", tt.field, result, tt.expected)
			}
		})
	}
}

func TestHashFilterValues(t *testing.T) {
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
			name:   "Hash string values",
			values: []any{"value1", "value2"},
		},
		{
			name:   "Keep non-string values",
			values: []any{123, 456.78, true},
		},
		{
			name:   "Mixed values",
			values: []any{"string", 123, "another", nil},
		},
		{
			name:   "Empty string value",
			values: []any{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := useCase.hashFilterValues(tt.values, crypto)
			if len(result) != len(tt.values) {
				t.Errorf("expected %d values, got %d", len(tt.values), len(result))
			}

			for i, v := range tt.values {
				if strVal, ok := v.(string); ok && strVal != "" {
					expectedHash := crypto.GenerateHash(&strVal)
					if result[i] != expectedHash {
						t.Errorf("value[%d]: expected hashed value, got %v", i, result[i])
					}
				} else {
					if result[i] != v {
						t.Errorf("value[%d]: expected unchanged value %v, got %v", i, v, result[i])
					}
				}
			}
		})
	}
}

func TestDecryptContactFields(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	if err != nil {
		t.Fatalf("Failed to initialize cipher: %v", err)
	}

	useCase := &UseCase{}

	tests := []struct {
		name           string
		record         map[string]any
		expectedEmails []string
		expectNoChange bool
	}{
		{
			name: "Decrypt contact fields",
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
			name: "No contact field present",
			record: map[string]any{
				"id": "test-id",
			},
			expectNoChange: true,
		},
		{
			name: "Contact field is not a map",
			record: map[string]any{
				"contact": "not a map",
			},
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := useCase.decryptContactFields(tt.record, crypto)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectNoChange && len(tt.expectedEmails) > 0 {
				contact, ok := tt.record["contact"].(map[string]any)
				if !ok {
					t.Fatal("contact not found or wrong type")
				}
				if contact["primary_email"] != tt.expectedEmails[0] {
					t.Errorf("expected primary_email = %q, got %q", tt.expectedEmails[0], contact["primary_email"])
				}
			}
		})
	}
}

func TestDecryptBankingDetailsFields(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	if err != nil {
		t.Fatalf("Failed to initialize cipher: %v", err)
	}

	useCase := &UseCase{}

	tests := []struct {
		name            string
		record          map[string]any
		expectedAccount string
		expectNoChange  bool
	}{
		{
			name: "Decrypt banking details fields",
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
			name: "No banking_details field present",
			record: map[string]any{
				"id": "test-id",
			},
			expectNoChange: true,
		},
		{
			name: "banking_details field is not a map",
			record: map[string]any{
				"banking_details": "not a map",
			},
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := useCase.decryptBankingDetailsFields(tt.record, crypto)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectNoChange && tt.expectedAccount != "" {
				bankingDetails, ok := tt.record["banking_details"].(map[string]any)
				if !ok {
					t.Fatal("banking_details not found or wrong type")
				}
				if bankingDetails["account"] != tt.expectedAccount {
					t.Errorf("expected account = %q, got %q", tt.expectedAccount, bankingDetails["account"])
				}
			}
		})
	}
}

func TestDecryptLegalPersonFields(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	if err != nil {
		t.Fatalf("Failed to initialize cipher: %v", err)
	}

	useCase := &UseCase{}

	tests := []struct {
		name           string
		record         map[string]any
		expectedName   string
		expectNoChange bool
	}{
		{
			name: "Decrypt legal person representative fields",
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
			name: "No legal_person field present",
			record: map[string]any{
				"id": "test-id",
			},
			expectNoChange: true,
		},
		{
			name: "legal_person without representative",
			record: map[string]any{
				"legal_person": map[string]any{
					"company_name": "Test Company",
				},
			},
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := useCase.decryptLegalPersonFields(tt.record, crypto)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectNoChange && tt.expectedName != "" {
				legalPerson, ok := tt.record["legal_person"].(map[string]any)
				if !ok {
					t.Fatal("legal_person not found or wrong type")
				}
				representative, ok := legalPerson["representative"].(map[string]any)
				if !ok {
					t.Fatal("representative not found or wrong type")
				}
				if representative["name"] != tt.expectedName {
					t.Errorf("expected name = %q, got %q", tt.expectedName, representative["name"])
				}
			}
		})
	}
}

func TestDecryptNaturalPersonFields(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	if err != nil {
		t.Fatalf("Failed to initialize cipher: %v", err)
	}

	useCase := &UseCase{}

	tests := []struct {
		name               string
		record             map[string]any
		expectedMotherName string
		expectNoChange     bool
	}{
		{
			name: "Decrypt natural person fields",
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
			name: "No natural_person field present",
			record: map[string]any{
				"id": "test-id",
			},
			expectNoChange: true,
		},
		{
			name: "natural_person field is not a map",
			record: map[string]any{
				"natural_person": "not a map",
			},
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := useCase.decryptNaturalPersonFields(tt.record, crypto)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectNoChange && tt.expectedMotherName != "" {
				naturalPerson, ok := tt.record["natural_person"].(map[string]any)
				if !ok {
					t.Fatal("natural_person not found or wrong type")
				}
				if naturalPerson["mother_name"] != tt.expectedMotherName {
					t.Errorf("expected mother_name = %q, got %q", tt.expectedMotherName, naturalPerson["mother_name"])
				}
			}
		})
	}
}

func TestDecryptFieldValue(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	if err != nil {
		t.Fatalf("Failed to initialize cipher: %v", err)
	}

	useCase := &UseCase{}

	tests := []struct {
		name        string
		container   map[string]any
		fieldName   string
		fieldValue  any
		expectError bool
	}{
		{
			name:       "Decrypt valid string",
			container:  map[string]any{},
			fieldName:  "test_field",
			fieldValue: func() string { v := "test"; e, _ := crypto.Encrypt(&v); return *e }(),
		},
		{
			name:       "Skip non-string value",
			container:  map[string]any{},
			fieldName:  "test_field",
			fieldValue: 123,
		},
		{
			name:       "Skip empty string",
			container:  map[string]any{},
			fieldName:  "test_field",
			fieldValue: "",
		},
		{
			name:        "Error on invalid encrypted value",
			container:   map[string]any{},
			fieldName:   "test_field",
			fieldValue:  "not-encrypted-data",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := useCase.decryptFieldValue(tt.container, tt.fieldName, tt.fieldValue, crypto)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConvertToPDFIfNeeded_NonPDFFormat(t *testing.T) {
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())
	_, span := tracer.Start(context.Background(), "test")

	useCase := &UseCase{}

	message := GenerateReportMessage{
		ReportID:     uuid.New(),
		OutputFormat: "html",
	}

	htmlContent := "<html><body>Test</body></html>"

	result, err := useCase.convertToPDFIfNeeded(context.Background(), tracer, message, htmlContent, &span, logger)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != htmlContent {
		t.Errorf("expected unchanged content for non-PDF format")
	}
}

func TestQueryDatabase_UnknownDataSource(t *testing.T) {
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())

	useCase := &UseCase{
		ExternalDataSources:   map[string]pkg.DataSource{},
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
	if err != nil {
		t.Errorf("expected nil error for unknown data source, got: %v", err)
	}
}

func TestQueryDatabase_CircuitBreakerUnhealthy(t *testing.T) {
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())

	cbManager := pkg.NewCircuitBreakerManager(logger)

	// Force circuit breaker to open state by recording failures
	for i := 0; i < 10; i++ {
		_, _ = cbManager.Execute("test_db", func() (any, error) {
			return nil, errors.New("simulated failure")
		})
	}

	useCase := &UseCase{
		ExternalDataSources: map[string]pkg.DataSource{
			"test_db": {
				Initialized:  true,
				DatabaseType: "postgresql",
			},
		},
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

	if err == nil {
		t.Error("expected error when circuit breaker is unhealthy")
	}
	if !strings.Contains(err.Error(), "circuit breaker") {
		t.Errorf("expected circuit breaker error, got: %v", err)
	}
}

func TestQueryDatabase_UnsupportedDatabaseType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())

	useCase := &UseCase{
		ExternalDataSources: map[string]pkg.DataSource{
			"test_db": {
				Initialized:  true,
				DatabaseType: "unsupported_type",
			},
		},
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

	if err == nil {
		t.Error("expected error for unsupported database type")
	}
	if !strings.Contains(err.Error(), "unsupported database type") {
		t.Errorf("expected 'unsupported database type' error, got: %v", err)
	}
}

func TestTransformPluginCRMAdvancedFilters_NilFilter(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	os.Setenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM", hashKey)
	defer os.Unsetenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	useCase := &UseCase{}

	result, err := useCase.transformPluginCRMAdvancedFilters(nil, logger)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for nil input, got: %v", result)
	}
}

func TestTransformPluginCRMAdvancedFilters_MissingEnvVar(t *testing.T) {
	os.Unsetenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	useCase := &UseCase{}

	filter := map[string]model.FilterCondition{
		"document": {
			Equals: []any{"12345678901"},
		},
	}

	_, err := useCase.transformPluginCRMAdvancedFilters(filter, logger)
	if err == nil {
		t.Error("expected error when env var is missing")
	}
	if !strings.Contains(err.Error(), "CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM") {
		t.Errorf("expected env var error, got: %v", err)
	}
}

func TestTransformPluginCRMAdvancedFilters_AllFilterConditions(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	os.Setenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM", hashKey)
	defer os.Unsetenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	useCase := &UseCase{}

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
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if _, exists := result["search.document"]; !exists {
		t.Error("expected search.document field in result")
	}

	// Verify all conditions were transformed
	searchDoc := result["search.document"]
	if len(searchDoc.Equals) == 0 {
		t.Error("expected Equals to be transformed")
	}
	if len(searchDoc.GreaterThan) == 0 {
		t.Error("expected GreaterThan to be transformed")
	}
	if len(searchDoc.GreaterOrEqual) == 0 {
		t.Error("expected GreaterOrEqual to be transformed")
	}
	if len(searchDoc.LessThan) == 0 {
		t.Error("expected LessThan to be transformed")
	}
	if len(searchDoc.LessOrEqual) == 0 {
		t.Error("expected LessOrEqual to be transformed")
	}
	if len(searchDoc.Between) == 0 {
		t.Error("expected Between to be transformed")
	}
	if len(searchDoc.In) == 0 {
		t.Error("expected In to be transformed")
	}
	if len(searchDoc.NotIn) == 0 {
		t.Error("expected NotIn to be transformed")
	}
}

func TestTransformPluginCRMAdvancedFilters_NonMappedField(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	os.Setenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM", hashKey)
	defer os.Unsetenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	useCase := &UseCase{}

	filter := map[string]model.FilterCondition{
		"unmapped_field": {
			Equals: []any{"value1"},
		},
	}

	result, err := useCase.transformPluginCRMAdvancedFilters(filter, logger)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Non-mapped fields should be kept as-is
	if _, exists := result["unmapped_field"]; !exists {
		t.Error("expected unmapped_field to be preserved")
	}
}

func TestGenerateReport_ReportAlreadyFinished(t *testing.T) {
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
		ExternalDataSources: map[string]pkg.DataSource{},
	}

	err := useCase.GenerateReport(context.Background(), bodyBytes)
	if err != nil {
		t.Errorf("expected no error for already finished report, got: %v", err)
	}
}

func TestUpdateReportWithErrors(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.reportID)

			useCase := &UseCase{
				ReportDataRepo: mockReportDataRepo,
			}

			err := useCase.updateReportWithErrors(context.Background(), tt.reportID, tt.errorMessage)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMarkReportAsFinished(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.reportID)

			useCase := &UseCase{
				ReportDataRepo: mockReportDataRepo,
			}

			err := useCase.markReportAsFinished(context.Background(), tt.reportID, &span, logger)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCheckReportStatus(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.reportID)

			useCase := &UseCase{
				ReportDataRepo: mockReportDataRepo,
			}

			status, err := useCase.checkReportStatus(context.Background(), tt.reportID, logger)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if status != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, status)
			}
		})
	}
}

func TestSaveReport_WithTTL(t *testing.T) {
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
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestQueryPostgresDatabase_SchemaFormats(t *testing.T) {
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
			name:     "Pongo2 format - schema__table",
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
			name:     "Qualified format - schema.table",
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
			name:     "Legacy format - table only (autodiscovery)",
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
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDecryptPluginCRMData_MissingEnvVars(t *testing.T) {
	os.Unsetenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")
	os.Unsetenv("CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM")

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	useCase := &UseCase{}

	collectionResult := []map[string]any{
		{"document": "encrypted_value"},
	}

	_, err := useCase.decryptPluginCRMData(logger, collectionResult, []string{"document"})
	if err == nil {
		t.Error("expected error when env vars are missing")
	}
}

func TestDecryptPluginCRMData_NoDecryptionNeeded(t *testing.T) {
	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	useCase := &UseCase{}

	collectionResult := []map[string]any{
		{"id": "123", "status": "active"},
	}

	result, err := useCase.decryptPluginCRMData(logger, collectionResult, []string{"id", "status"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

func TestHandleErrorWithUpdate(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.reportID)

			useCase := &UseCase{
				ReportDataRepo: mockReportDataRepo,
			}

			err := useCase.handleErrorWithUpdate(context.Background(), tt.reportID, &span, tt.errorMsg, tt.inputErr, logger)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
		})
	}
}

func TestDecryptNestedFields_AllTypes(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	if err != nil {
		t.Fatalf("Failed to initialize cipher: %v", err)
	}

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
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify all fields were decrypted
	contact := record["contact"].(map[string]any)
	if contact["primary_email"] != email {
		t.Errorf("expected email %q, got %q", email, contact["primary_email"])
	}

	bankingDetails := record["banking_details"].(map[string]any)
	if bankingDetails["account"] != account {
		t.Errorf("expected account %q, got %q", account, bankingDetails["account"])
	}

	legalPerson := record["legal_person"].(map[string]any)
	representative := legalPerson["representative"].(map[string]any)
	if representative["name"] != repName {
		t.Errorf("expected rep name %q, got %q", repName, representative["name"])
	}

	naturalPerson := record["natural_person"].(map[string]any)
	if naturalPerson["mother_name"] != motherName {
		t.Errorf("expected mother name %q, got %q", motherName, naturalPerson["mother_name"])
	}

	regulatoryFields := record["regulatory_fields"].(map[string]any)
	if regulatoryFields["participant_document"] != participantDoc {
		t.Errorf("expected participant doc %q, got %q", participantDoc, regulatoryFields["participant_document"])
	}

	relatedParties := record["related_parties"].([]any)
	party := relatedParties[0].(map[string]any)
	if party["document"] != partyDoc {
		t.Errorf("expected party doc %q, got %q", partyDoc, party["document"])
	}
}

func TestDecryptRecord(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	if err != nil {
		t.Fatalf("Failed to initialize cipher: %v", err)
	}

	useCase := &UseCase{}

	tests := []struct {
		name           string
		record         map[string]any
		expectedFields map[string]any
	}{
		{
			name: "Decrypt record with top-level encrypted fields",
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
			name: "Decrypt record with no encrypted fields",
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
		t.Run(tt.name, func(t *testing.T) {
			result, err := useCase.decryptRecord(tt.record, crypto)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			for key, expected := range tt.expectedFields {
				if result[key] != expected {
					t.Errorf("expected %s = %v, got %v", key, expected, result[key])
				}
			}
		})
	}
}

func TestDecryptTopLevelFields(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	encryptKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashKey,
		EncryptSecretKey: encryptKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	if err != nil {
		t.Fatalf("Failed to initialize cipher: %v", err)
	}

	useCase := &UseCase{}

	tests := []struct {
		name           string
		record         map[string]any
		expectedDoc    string
		expectNoChange bool
	}{
		{
			name: "Decrypt document and name fields",
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
			name: "No encrypted fields present",
			record: map[string]any{
				"id":     "123",
				"status": "active",
			},
			expectNoChange: true,
		},
		{
			name: "Encrypted field with nil value",
			record: map[string]any{
				"document": nil,
				"name":     nil,
			},
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := useCase.decryptTopLevelFields(tt.record, crypto)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectNoChange && tt.expectedDoc != "" {
				if tt.record["document"] != tt.expectedDoc {
					t.Errorf("expected document = %q, got %q", tt.expectedDoc, tt.record["document"])
				}
			}
		})
	}
}

func TestGenerateReport_ReportInErrorState(t *testing.T) {
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
		ExternalDataSources: map[string]pkg.DataSource{},
	}

	err := useCase.GenerateReport(context.Background(), bodyBytes)
	if err != nil {
		t.Errorf("expected no error for report in error state, got: %v", err)
	}
}

func TestQueryMongoDatabase_Success(t *testing.T) {
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
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result["test_db"]["users"]) != 1 {
		t.Errorf("expected 1 result, got %d", len(result["test_db"]["users"]))
	}
}

func TestQueryMongoDatabase_WithFilters(t *testing.T) {
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
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProcessRegularMongoCollection(t *testing.T) {
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
		"products",
		[]string{"name", "price"},
		nil,
		result,
		logger,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result["shop_db"]["products"]) != 1 {
		t.Errorf("expected 1 product, got %d", len(result["shop_db"]["products"]))
	}
}

func TestLoadTemplate_Success(t *testing.T) {
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
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if string(result) != string(templateContent) {
		t.Errorf("expected template content %q, got %q", templateContent, result)
	}
}

func TestLoadTemplate_Error(t *testing.T) {
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
	if err == nil {
		t.Error("expected error but got none")
	}
}

func TestQueryExternalData_NoDataSources(t *testing.T) {
	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	cbManager := pkg.NewCircuitBreakerManager(logger)

	useCase := &UseCase{
		ExternalDataSources:   map[string]pkg.DataSource{},
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
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got %d entries", len(result))
	}
}
