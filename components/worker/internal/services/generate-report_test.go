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
