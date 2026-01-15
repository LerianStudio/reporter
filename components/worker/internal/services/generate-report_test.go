package services

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libCrypto "github.com/LerianStudio/lib-commons/v2/commons/crypto"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/model"
	mongodb2 "github.com/LerianStudio/reporter/v4/pkg/mongodb"
	reportData "github.com/LerianStudio/reporter/v4/pkg/mongodb/report"
	"github.com/LerianStudio/reporter/v4/pkg/pdf"
	postgres2 "github.com/LerianStudio/reporter/v4/pkg/postgres"
	"github.com/LerianStudio/reporter/v4/pkg/seaweedfs/report"
	"github.com/LerianStudio/reporter/v4/pkg/seaweedfs/template"
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
		FindByID(gomock.Any(), reportID, gomock.Any()).
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
		QueryWithAdvancedFilters(
			gomock.Any(),
			gomock.Any(),
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
		FindByID(gomock.Any(), reportID, gomock.Any()).
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
		FindByID(gomock.Any(), reportID, gomock.Any()).
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

func TestParseMessage_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	// Expect UpdateReportStatusById to be called for error
	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	ctx := context.Background()
	_, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "test")
	defer span.End()

	logger, _, _, _ := libCommons.NewTrackingFromContext(ctx)

	invalidJSON := []byte(`{invalid json}`)
	_, err := useCase.parseMessage(ctx, invalidJSON, &span, logger)

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseMessage_UpdateErrorFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	// UpdateReportStatusById fails
	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("database error"))

	ctx := context.Background()
	_, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "test")
	defer span.End()

	logger, _, _, _ := libCommons.NewTrackingFromContext(ctx)

	invalidJSON := []byte(`{invalid json}`)
	_, err := useCase.parseMessage(ctx, invalidJSON, &span, logger)

	if err == nil || !strings.Contains(err.Error(), "database error") {
		t.Errorf("expected database error, got: %v", err)
	}
}

func TestShouldSkipProcessing_FinishedStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	reportID := uuid.New()

	mockReportDataRepo.EXPECT().
		FindByID(gomock.Any(), reportID, gomock.Any()).
		Return(&reportData.Report{
			ID:     reportID,
			Status: "Finished",
		}, nil)

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	skip := useCase.shouldSkipProcessing(context.Background(), reportID, logger)

	if !skip {
		t.Error("expected shouldSkipProcessing to return true for finished status")
	}
}

func TestShouldSkipProcessing_ErrorStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	reportID := uuid.New()

	mockReportDataRepo.EXPECT().
		FindByID(gomock.Any(), reportID, gomock.Any()).
		Return(&reportData.Report{
			ID:     reportID,
			Status: "Error",
		}, nil)

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	skip := useCase.shouldSkipProcessing(context.Background(), reportID, logger)

	if !skip {
		t.Error("expected shouldSkipProcessing to return true for error status")
	}
}

func TestShouldSkipProcessing_CheckStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	reportID := uuid.New()

	mockReportDataRepo.EXPECT().
		FindByID(gomock.Any(), reportID, gomock.Any()).
		Return(nil, errors.New("database error"))

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	skip := useCase.shouldSkipProcessing(context.Background(), reportID, logger)

	if skip {
		t.Error("expected shouldSkipProcessing to return false when checkReportStatus fails")
	}
}

func TestHandleErrorWithUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	reportID := uuid.New()

	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), reportID, gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	ctx := context.Background()
	_, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "test")
	defer span.End()

	logger, _, _, _ := libCommons.NewTrackingFromContext(ctx)

	originalErr := errors.New("original error")
	err := useCase.handleErrorWithUpdate(ctx, reportID, &span, "Test error message", originalErr, logger)

	if err == nil || !strings.Contains(err.Error(), "original error") {
		t.Errorf("expected original error to be returned, got: %v", err)
	}
}

func TestHandleErrorWithUpdate_UpdateFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	reportID := uuid.New()

	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), reportID, gomock.Any(), gomock.Any()).
		Return(errors.New("update failed"))

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	ctx := context.Background()
	_, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "test")
	defer span.End()

	logger, _, _, _ := libCommons.NewTrackingFromContext(ctx)

	originalErr := errors.New("original error")
	err := useCase.handleErrorWithUpdate(ctx, reportID, &span, "Test error message", originalErr, logger)

	if err == nil || !strings.Contains(err.Error(), "update failed") {
		t.Errorf("expected update error to be returned, got: %v", err)
	}
}

func TestQueryDatabase_UnknownDataSource(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())
	circuitBreakerManager := pkg.NewCircuitBreakerManager(logger)

	useCase := &UseCase{
		ExternalDataSources:   map[string]pkg.DataSource{},
		CircuitBreakerManager: circuitBreakerManager,
	}

	result := make(map[string]map[string][]map[string]any)

	err := useCase.queryDatabase(
		context.Background(),
		"unknown_database",
		map[string][]string{"table1": {"field1"}},
		nil,
		result,
		logger,
		tracer,
	)

	// Should return nil (continue with next database)
	if err != nil {
		t.Errorf("expected nil error for unknown datasource, got: %v", err)
	}
}

func TestQueryDatabase_UnsupportedDatabaseType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, tracer, _, _ := libCommons.NewTrackingFromContext(context.Background())
	circuitBreakerManager := pkg.NewCircuitBreakerManager(logger)

	useCase := &UseCase{
		ExternalDataSources: map[string]pkg.DataSource{
			"test_db": {
				Initialized:  true,
				DatabaseType: "unsupported_type",
			},
		},
		CircuitBreakerManager: circuitBreakerManager,
	}

	result := make(map[string]map[string][]map[string]any)

	err := useCase.queryDatabase(
		context.Background(),
		"test_db",
		map[string][]string{"table1": {"field1"}},
		nil,
		result,
		logger,
		tracer,
	)

	if err == nil || !strings.Contains(err.Error(), "unsupported database type") {
		t.Errorf("expected unsupported database type error, got: %v", err)
	}
}

func TestMarkReportAsFinished_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	reportID := uuid.New()

	// First call fails, second call (for error update) succeeds
	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), "Finished", reportID, gomock.Any(), nil).
		Return(errors.New("update failed"))

	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), "Error", reportID, gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	ctx := context.Background()
	_, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "test")
	defer span.End()

	logger, _, _, _ := libCommons.NewTrackingFromContext(ctx)

	err := useCase.markReportAsFinished(ctx, reportID, &span, logger)

	if err == nil || !strings.Contains(err.Error(), "update failed") {
		t.Errorf("expected update failed error, got: %v", err)
	}
}

func TestGetTableFilters_NilFilters(t *testing.T) {
	result := getTableFilters(nil, "table1")

	if result != nil {
		t.Error("expected nil result for nil filters")
	}
}

func TestGetTableFilters_TableNotFound(t *testing.T) {
	filters := map[string]map[string]model.FilterCondition{
		"other_table": {
			"field1": {Equals: []any{"value1"}},
		},
	}

	result := getTableFilters(filters, "table1")

	if result != nil {
		t.Error("expected nil result when table not found")
	}
}

func TestGetTableFilters_TableFound(t *testing.T) {
	filters := map[string]map[string]model.FilterCondition{
		"table1": {
			"field1": {Equals: []any{"value1"}},
		},
	}

	result := getTableFilters(filters, "table1")

	if result == nil {
		t.Error("expected non-nil result when table found")
	}
	if _, exists := result["field1"]; !exists {
		t.Error("expected field1 to be present in result")
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
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := isEncryptedField(tt.field)
			if result != tt.expected {
				t.Errorf("isEncryptedField(%q) = %v; want %v", tt.field, result, tt.expected)
			}
		})
	}
}

func TestDecryptPluginCRMData_NoDecryptionNeeded(t *testing.T) {
	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	useCase := &UseCase{}

	data := []map[string]any{
		{"id": "123", "status": "active"},
	}

	result, err := useCase.decryptPluginCRMData(logger, data, []string{"id", "status"})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

func TestDecryptPluginCRMData_MissingEncryptKey(t *testing.T) {
	os.Setenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM", "test_hash_key")
	os.Unsetenv("CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM")
	defer os.Unsetenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	useCase := &UseCase{}

	data := []map[string]any{
		{"name": "encrypted_value"},
	}

	_, err := useCase.decryptPluginCRMData(logger, data, []string{"name"})

	if err == nil || !strings.Contains(err.Error(), "CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM") {
		t.Errorf("expected missing encrypt key error, got: %v", err)
	}
}

func TestDecryptPluginCRMData_MissingHashKey(t *testing.T) {
	os.Unsetenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")
	os.Setenv("CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM", "test_encrypt_key")
	defer os.Unsetenv("CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM")

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	useCase := &UseCase{}

	data := []map[string]any{
		{"name": "encrypted_value"},
	}

	_, err := useCase.decryptPluginCRMData(logger, data, []string{"name"})

	if err == nil || !strings.Contains(err.Error(), "CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM") {
		t.Errorf("expected missing hash key error, got: %v", err)
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

	// Encrypt test values
	nameStr := "John Doe"
	documentStr := "123456789"
	emailStr := "john@example.com"

	encryptedName, _ := crypto.Encrypt(&nameStr)
	encryptedDocument, _ := crypto.Encrypt(&documentStr)
	encryptedEmail, _ := crypto.Encrypt(&emailStr)

	useCase := &UseCase{}

	record := map[string]any{
		"legal_person": map[string]any{
			"representative": map[string]any{
				"name":     *encryptedName,
				"document": *encryptedDocument,
				"email":    *encryptedEmail,
			},
		},
	}

	err = useCase.decryptLegalPersonFields(record, crypto)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify decryption
	legalPerson := record["legal_person"].(map[string]any)
	representative := legalPerson["representative"].(map[string]any)

	if representative["name"] != nameStr {
		t.Errorf("expected name %q, got %q", nameStr, representative["name"])
	}
	if representative["document"] != documentStr {
		t.Errorf("expected document %q, got %q", documentStr, representative["document"])
	}
	if representative["email"] != emailStr {
		t.Errorf("expected email %q, got %q", emailStr, representative["email"])
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

	motherNameStr := "Jane Doe"
	fatherNameStr := "John Doe Sr"

	encryptedMotherName, _ := crypto.Encrypt(&motherNameStr)
	encryptedFatherName, _ := crypto.Encrypt(&fatherNameStr)

	useCase := &UseCase{}

	record := map[string]any{
		"natural_person": map[string]any{
			"mother_name": *encryptedMotherName,
			"father_name": *encryptedFatherName,
		},
	}

	err = useCase.decryptNaturalPersonFields(record, crypto)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	naturalPerson := record["natural_person"].(map[string]any)

	if naturalPerson["mother_name"] != motherNameStr {
		t.Errorf("expected mother_name %q, got %q", motherNameStr, naturalPerson["mother_name"])
	}
	if naturalPerson["father_name"] != fatherNameStr {
		t.Errorf("expected father_name %q, got %q", fatherNameStr, naturalPerson["father_name"])
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

	primaryEmail := "primary@example.com"
	secondaryEmail := "secondary@example.com"
	mobilePhone := "1234567890"
	otherPhone := "0987654321"

	encryptedPrimaryEmail, _ := crypto.Encrypt(&primaryEmail)
	encryptedSecondaryEmail, _ := crypto.Encrypt(&secondaryEmail)
	encryptedMobilePhone, _ := crypto.Encrypt(&mobilePhone)
	encryptedOtherPhone, _ := crypto.Encrypt(&otherPhone)

	useCase := &UseCase{}

	record := map[string]any{
		"contact": map[string]any{
			"primary_email":   *encryptedPrimaryEmail,
			"secondary_email": *encryptedSecondaryEmail,
			"mobile_phone":    *encryptedMobilePhone,
			"other_phone":     *encryptedOtherPhone,
		},
	}

	err = useCase.decryptContactFields(record, crypto)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	contact := record["contact"].(map[string]any)

	if contact["primary_email"] != primaryEmail {
		t.Errorf("expected primary_email %q, got %q", primaryEmail, contact["primary_email"])
	}
	if contact["secondary_email"] != secondaryEmail {
		t.Errorf("expected secondary_email %q, got %q", secondaryEmail, contact["secondary_email"])
	}
	if contact["mobile_phone"] != mobilePhone {
		t.Errorf("expected mobile_phone %q, got %q", mobilePhone, contact["mobile_phone"])
	}
	if contact["other_phone"] != otherPhone {
		t.Errorf("expected other_phone %q, got %q", otherPhone, contact["other_phone"])
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

	accountStr := "12345-6"
	ibanStr := "BR1234567890123456789012345"

	encryptedAccount, _ := crypto.Encrypt(&accountStr)
	encryptedIban, _ := crypto.Encrypt(&ibanStr)

	useCase := &UseCase{}

	record := map[string]any{
		"banking_details": map[string]any{
			"account": *encryptedAccount,
			"iban":    *encryptedIban,
		},
	}

	err = useCase.decryptBankingDetailsFields(record, crypto)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	bankingDetails := record["banking_details"].(map[string]any)

	if bankingDetails["account"] != accountStr {
		t.Errorf("expected account %q, got %q", accountStr, bankingDetails["account"])
	}
	if bankingDetails["iban"] != ibanStr {
		t.Errorf("expected iban %q, got %q", ibanStr, bankingDetails["iban"])
	}
}

func TestDecryptFieldValue_NonString(t *testing.T) {
	useCase := &UseCase{}

	container := map[string]any{
		"field1": 12345, // not a string
	}

	err := useCase.decryptFieldValue(container, "field1", 12345, nil)
	if err != nil {
		t.Errorf("unexpected error for non-string value: %v", err)
	}
	// Value should remain unchanged
	if container["field1"] != 12345 {
		t.Errorf("expected value to remain unchanged")
	}
}

func TestDecryptFieldValue_EmptyString(t *testing.T) {
	useCase := &UseCase{}

	container := map[string]any{
		"field1": "",
	}

	err := useCase.decryptFieldValue(container, "field1", "", nil)
	if err != nil {
		t.Errorf("unexpected error for empty string: %v", err)
	}
}

func TestTransformPluginCRMAdvancedFilters_NilFilter(t *testing.T) {
	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	useCase := &UseCase{}

	result, err := useCase.transformPluginCRMAdvancedFilters(nil, logger)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for nil filter")
	}
}

func TestTransformPluginCRMAdvancedFilters_MissingHashKey(t *testing.T) {
	os.Unsetenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	useCase := &UseCase{}

	filter := map[string]model.FilterCondition{
		"document": {Equals: []any{"123456789"}},
	}

	_, err := useCase.transformPluginCRMAdvancedFilters(filter, logger)

	if err == nil || !strings.Contains(err.Error(), "CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM") {
		t.Errorf("expected missing hash key error, got: %v", err)
	}
}

func TestTransformPluginCRMAdvancedFilters_AllFilterTypes(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	os.Setenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM", hashKey)
	defer os.Unsetenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	useCase := &UseCase{}

	filter := map[string]model.FilterCondition{
		"document": {
			Equals:        []any{"value1"},
			GreaterThan:   []any{"value2"},
			GreaterOrEqual: []any{"value3"},
			LessThan:      []any{"value4"},
			LessOrEqual:   []any{"value5"},
			Between:       []any{"value6", "value7"},
			In:            []any{"value8", "value9"},
			NotIn:         []any{"value10"},
		},
		"unknown_field": {
			Equals: []any{"keep_as_is"},
		},
	}

	result, err := useCase.transformPluginCRMAdvancedFilters(filter, logger)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check that document was transformed to search.document
	if _, exists := result["search.document"]; !exists {
		t.Error("expected search.document to be present")
	}

	// Check that unknown_field was kept as-is
	if _, exists := result["unknown_field"]; !exists {
		t.Error("expected unknown_field to be present")
	}
}

func TestHashFilterValues_MixedTypes(t *testing.T) {
	hashKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	crypto := &libCrypto.Crypto{
		HashSecretKey: hashKey,
		Logger:        logger,
	}

	useCase := &UseCase{}

	values := []any{"string_value", 12345, "", nil}

	result := useCase.hashFilterValues(values, crypto)

	if len(result) != 4 {
		t.Errorf("expected 4 results, got %d", len(result))
	}

	// First value should be hashed (non-empty string)
	if result[0] == "string_value" {
		t.Error("expected first value to be hashed")
	}

	// Second value should remain as integer
	if result[1] != 12345 {
		t.Errorf("expected second value to remain as integer, got: %v", result[1])
	}

	// Third value should remain as empty string (not hashed)
	if result[2] != "" {
		t.Errorf("expected third value to remain as empty string, got: %v", result[2])
	}

	// Fourth value should remain as nil
	if result[3] != nil {
		t.Errorf("expected fourth value to remain as nil, got: %v", result[3])
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
	templateID := uuid.New()
	message := GenerateReportMessage{
		TemplateID:   templateID,
		ReportID:     reportID,
		OutputFormat: "json",
	}
	renderedOutput := `{"data": "test"}`

	mockReportRepo.
		EXPECT().
		Put(gomock.Any(), templateID.String()+"/"+reportID.String()+".json", "application/json", []byte(renderedOutput), "30d").
		Return(nil)

	ctx := context.Background()
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)

	err := useCase.saveReport(ctx, tracer, message, renderedOutput, logger)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConvertToPDFIfNeeded_NotPDF(t *testing.T) {
	useCase := &UseCase{}

	ctx := context.Background()
	_, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "test")
	defer span.End()

	logger, _, _, _ := libCommons.NewTrackingFromContext(ctx)

	message := GenerateReportMessage{
		OutputFormat: "html",
	}

	result, err := useCase.convertToPDFIfNeeded(ctx, tracer, message, "<html>test</html>", &span, logger)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "<html>test</html>" {
		t.Errorf("expected output unchanged for non-PDF format")
	}
}

func TestQueryPostgresDatabase_NoFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostgresRepo := postgres2.NewMockRepository(ctrl)
	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())
	circuitBreakerManager := pkg.NewCircuitBreakerManager(logger)

	mockPostgresRepo.EXPECT().
		GetDatabaseSchema(gomock.Any()).
		Return([]postgres2.TableSchema{
			{
				TableName: "users",
				Columns: []postgres2.ColumnInformation{
					{Name: "name", DataType: "text"},
				},
			},
		}, nil)

	mockPostgresRepo.EXPECT().
		Query(gomock.Any(), gomock.Any(), "users", []string{"name"}, nil).
		Return([]map[string]any{{"name": "John"}}, nil)

	useCase := &UseCase{
		CircuitBreakerManager: circuitBreakerManager,
	}

	dataSource := &pkg.DataSource{
		Initialized:        true,
		DatabaseType:       "postgresql",
		PostgresRepository: mockPostgresRepo,
	}

	result := make(map[string]map[string][]map[string]any)
	result["test_db"] = make(map[string][]map[string]any)

	err := useCase.queryPostgresDatabase(
		context.Background(),
		dataSource,
		"test_db",
		map[string][]string{"users": {"name"}},
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

func TestUpdateReportWithErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	reportID := uuid.New()

	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), "Error", reportID, gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	err := useCase.updateReportWithErrors(context.Background(), reportID, "test error message")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateReportWithErrors_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	reportID := uuid.New()

	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), "Error", reportID, gomock.Any(), gomock.Any()).
		Return(errors.New("database error"))

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	err := useCase.updateReportWithErrors(context.Background(), reportID, "test error message")

	if err == nil || !strings.Contains(err.Error(), "database error") {
		t.Errorf("expected database error, got: %v", err)
	}
}

func TestCheckReportStatus_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	reportID := uuid.New()

	mockReportDataRepo.EXPECT().
		FindByID(gomock.Any(), reportID, gomock.Any()).
		Return(&reportData.Report{
			ID:     reportID,
			Status: "processing",
		}, nil)

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	status, err := useCase.checkReportStatus(context.Background(), reportID, logger)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if status != "processing" {
		t.Errorf("expected status 'processing', got %q", status)
	}
}

func TestCheckReportStatus_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)
	reportID := uuid.New()

	mockReportDataRepo.EXPECT().
		FindByID(gomock.Any(), reportID, gomock.Any()).
		Return(nil, errors.New("not found"))

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	logger, _, _, _ := libCommons.NewTrackingFromContext(context.Background())

	status, err := useCase.checkReportStatus(context.Background(), reportID, logger)

	if err == nil {
		t.Error("expected error for not found report")
	}
	if status != "" {
		t.Errorf("expected empty status, got %q", status)
	}
}

func TestDecryptNestedFields_NoNestedData(t *testing.T) {
	useCase := &UseCase{}

	record := map[string]any{
		"id":   "123",
		"name": "test",
	}

	err := useCase.decryptNestedFields(record, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoadTemplate_UpdateErrorFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTemplateRepo := template.NewMockRepository(ctrl)
	mockReportDataRepo := reportData.NewMockRepository(ctrl)

	templateID := uuid.New()
	reportID := uuid.New()

	mockTemplateRepo.EXPECT().
		Get(gomock.Any(), templateID.String()).
		Return(nil, errors.New("template not found"))

	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), reportID, gomock.Any(), gomock.Any()).
		Return(errors.New("update failed"))

	useCase := &UseCase{
		TemplateSeaweedFS: mockTemplateRepo,
		ReportDataRepo:    mockReportDataRepo,
	}

	ctx := context.Background()
	_, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "test")
	defer span.End()

	logger, _, _, _ := libCommons.NewTrackingFromContext(ctx)

	message := GenerateReportMessage{
		TemplateID: templateID,
		ReportID:   reportID,
	}

	_, err := useCase.loadTemplate(ctx, tracer, message, &span, logger)

	if err == nil || !strings.Contains(err.Error(), "update failed") {
		t.Errorf("expected update failed error, got: %v", err)
	}
}

func TestRenderTemplate_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDataRepo := reportData.NewMockRepository(ctrl)

	reportID := uuid.New()

	mockReportDataRepo.EXPECT().
		UpdateReportStatusById(gomock.Any(), gomock.Any(), reportID, gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := &UseCase{
		ReportDataRepo: mockReportDataRepo,
	}

	ctx := context.Background()
	_, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "test")
	defer span.End()

	logger, _, _, _ := libCommons.NewTrackingFromContext(ctx)

	message := GenerateReportMessage{
		ReportID: reportID,
	}

	// Invalid template syntax
	templateBytes := []byte("{% invalid syntax %}")
	result := make(map[string]map[string][]map[string]any)

	_, err := useCase.renderTemplate(ctx, tracer, templateBytes, result, message, &span, logger)

	if err == nil {
		t.Error("expected error for invalid template syntax")
	}
}

func TestConvertHTMLToPDF_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPdfPool := pdf.NewMockPool(ctrl)
	mockLogger := log.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

	htmlContent := "<html><body><h1>Test PDF</h1></body></html>"

	// Mock Submit to write a fake PDF content to the file
	mockPdfPool.EXPECT().
		Submit(htmlContent, gomock.Any()).
		DoAndReturn(func(html, filename string) error {
			// Write fake PDF content to the file
			return os.WriteFile(filename, []byte("%PDF-1.4 fake pdf content"), 0600)
		})

	useCase := &UseCase{
		PdfPool: mockPdfPool,
	}

	pdfBytes, err := useCase.convertHTMLToPDF(htmlContent, mockLogger)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Error("expected non-empty PDF bytes")
	}
	if !strings.HasPrefix(string(pdfBytes), "%PDF") {
		t.Error("expected PDF content to start with %PDF")
	}
}

func TestConvertHTMLToPDF_SubmitError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPdfPool := pdf.NewMockPool(ctrl)
	mockLogger := log.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	htmlContent := "<html><body><h1>Test</h1></body></html>"

	mockPdfPool.EXPECT().
		Submit(htmlContent, gomock.Any()).
		Return(errors.New("chrome failed"))

	useCase := &UseCase{
		PdfPool: mockPdfPool,
	}

	_, err := useCase.convertHTMLToPDF(htmlContent, mockLogger)

	if err == nil {
		t.Error("expected error when Submit fails")
	}
	if !strings.Contains(err.Error(), "failed to generate PDF") {
		t.Errorf("expected 'failed to generate PDF' error, got: %v", err)
	}
}

func TestConvertHTMLToPDF_ReadFileError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPdfPool := pdf.NewMockPool(ctrl)
	mockLogger := log.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	htmlContent := "<html><body><h1>Test</h1></body></html>"

	// Submit succeeds but doesn't create the file (simulates file not being created)
	mockPdfPool.EXPECT().
		Submit(htmlContent, gomock.Any()).
		DoAndReturn(func(html, filename string) error {
			// Remove the file if it exists (simulates PDF generation failure without error)
			os.Remove(filename)
			return nil
		})

	useCase := &UseCase{
		PdfPool: mockPdfPool,
	}

	_, err := useCase.convertHTMLToPDF(htmlContent, mockLogger)

	if err == nil {
		t.Error("expected error when file cannot be read")
	}
	if !strings.Contains(err.Error(), "failed to read generated PDF") {
		t.Errorf("expected 'failed to read generated PDF' error, got: %v", err)
	}
}
