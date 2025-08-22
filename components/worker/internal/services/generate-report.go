package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"plugin-smart-templates/v2/pkg"
	"plugin-smart-templates/v2/pkg/constant"
	"plugin-smart-templates/v2/pkg/pongo"
	"strings"
	"time"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libCrypto "github.com/LerianStudio/lib-commons/v2/commons/crypto"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libOtel "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GenerateReportMessage contains the information needed to generate a report.
type GenerateReportMessage struct {
	// TemplateID is the unique identifier of the template to be used for report generation.
	TemplateID uuid.UUID `json:"templateId"`

	// ReportID uniquely identifies this report generation request
	ReportID uuid.UUID `json:"reportId"`

	// OutputFormat specifies the format of the generated report (e.g., html, csv, json).
	OutputFormat string `json:"outputFormat"`

	// DataQueries maps database names to tables and their fields.
	// Format: map[databaseName]map[tableName][]fieldName.
	// Example: {"onboarding": {"organization": ["name"], "ledger": ["id"]}}.
	DataQueries map[string]map[string][]string `json:"mappedFields"`

	// Filters specify the filtering criteria for the data queries, mapping filter keys to their respective values.
	Filters map[string]map[string]map[string][]any `json:"filters"`
}

// mimeTypes maps file extensions to their corresponding MIME content types
var mimeTypes = map[string]string{
	"txt":  "text/plain",
	"html": "text/html",
	"json": "application/json",
	"csv":  "text/csv",
	// Add more MIME types as needed
}

// GenerateReport handles a report generation request by loading a template file,
// processing it, and storing the final report in the report repository.
func (uc *UseCase) GenerateReport(ctx context.Context, body []byte) error {
	logger := libCommons.NewLoggerFromContext(ctx)
	tracer := libCommons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.generate_report")
	defer span.End()

	var message GenerateReportMessage

	err := json.Unmarshal(body, &message)
	if err != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, message.ReportID, err.Error()); errUpdate != nil {
			libOtel.HandleSpanError(&span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return errUpdate
		}

		libOtel.HandleSpanError(&span, "Error unmarshalling message.", err)

		logger.Errorf("Error unmarshalling message: %s", err.Error())

		return err
	}

	ctx, spanTemplate := tracer.Start(ctx, "service.generate_report.get_template")

	fileBytes, err := uc.TemplateFileRepo.Get(ctx, message.TemplateID.String())
	if err != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, message.ReportID, err.Error()); errUpdate != nil {
			libOtel.HandleSpanError(&span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return errUpdate
		}

		libOtel.HandleSpanError(&spanTemplate, "Error getting file from template bucket.", err)

		logger.Errorf("Error getting file from template bucket: %s", err.Error())

		return err
	}

	logger.Infof("Template found: %s", string(fileBytes))

	spanTemplate.End()

	result := make(map[string]map[string][]map[string]any)

	err = uc.queryExternalData(ctx, message, result)
	if err != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, message.ReportID, err.Error()); errUpdate != nil {
			libOtel.HandleSpanError(&span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return errUpdate
		}

		logger.Errorf("Error querying external data: %s", err.Error())

		return err
	}

	ctx, spanRender := tracer.Start(ctx, "service.generate_report.render_template")

	renderer := pongo.NewTemplateRenderer()

	out, err := renderer.RenderFromBytes(ctx, fileBytes, result)
	if err != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, message.ReportID, err.Error()); errUpdate != nil {
			libOtel.HandleSpanError(&span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return errUpdate
		}

		libOtel.HandleSpanError(&spanRender, "Error rendering template.", err)

		logger.Errorf("Error rendering template: %s", err.Error())

		return err
	}

	spanRender.End()

	err = uc.saveReport(ctx, tracer, message, out, logger)
	if err != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, message.ReportID, err.Error()); errUpdate != nil {
			libOtel.HandleSpanError(&span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return errUpdate
		}

		libOtel.HandleSpanError(&span, "Error saving report.", err)

		logger.Errorf("Error saving report: %s", err.Error())

		return err
	}

	errUpdateStatus := uc.ReportDataRepo.UpdateReportStatusById(ctx,
		constant.FinishedStatus, message.ReportID, time.Now(), nil)
	if errUpdateStatus != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, message.ReportID, errUpdateStatus.Error()); errUpdate != nil {
			libOtel.HandleSpanError(&span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return errUpdate
		}

		libOtel.HandleSpanError(&span, "Error to update report status.", errUpdateStatus)

		logger.Errorf("Error saving report: %s", errUpdateStatus.Error())

		return errUpdateStatus
	}

	return nil
}

// updateReportWithErrors updates the status of a report to "Error" with metadata containing the provided error message.
func (uc *UseCase) updateReportWithErrors(ctx context.Context, reportId uuid.UUID, errorMessage string) error {
	metadata := make(map[string]any)
	metadata["error"] = errorMessage

	errUpdate := uc.ReportDataRepo.UpdateReportStatusById(ctx, constant.ErrorStatus,
		reportId, time.Now(), metadata)
	if errUpdate != nil {
		return errUpdate
	}

	return nil
}

// saveReport handles saving the generated report file to the report repository and logs any encountered errors.
// It determines the object name, content type, and stores the file using the ReportFileRepo interface.
// Returns an error if the file storage operation fails.
func (uc *UseCase) saveReport(ctx context.Context, tracer trace.Tracer, message GenerateReportMessage, out string, logger log.Logger) error {
	ctx, spanSaveReport := tracer.Start(ctx, "service.generate_report.save_report")
	defer spanSaveReport.End()

	outputFormat := strings.ToLower(message.OutputFormat)
	contentType := getContentType(outputFormat)
	objectName := message.TemplateID.String() + "/" + message.ReportID.String() + "." + outputFormat

	err := uc.ReportFileRepo.Put(ctx, objectName, contentType, []byte(out))
	if err != nil {
		libOtel.HandleSpanError(&spanSaveReport, "Error putting report file.", err)

		logger.Errorf("Error putting report file: %s", err.Error())

		return err
	}

	return nil
}

// queryExternalData retrieves data from external data sources specified in the message and populates the result map.
func (uc *UseCase) queryExternalData(ctx context.Context, message GenerateReportMessage, result map[string]map[string][]map[string]any) error {
	logger := libCommons.NewLoggerFromContext(ctx)
	tracer := libCommons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.generate_report.query_external_data")
	defer span.End()

	for databaseName, tables := range message.DataQueries {
		if err := uc.queryDatabase(ctx, databaseName, tables, message.Filters, result, logger, tracer); err != nil {
			return err
		}
	}

	return nil
}

// queryDatabase handles data retrieval for a specific database
func (uc *UseCase) queryDatabase(
	ctx context.Context,
	databaseName string,
	tables map[string][]string,
	allFilters map[string]map[string]map[string][]any,
	result map[string]map[string][]map[string]any,
	logger log.Logger,
	tracer trace.Tracer,
) error {
	ctx, dbSpan := tracer.Start(ctx, "service.generate_report.query_external_data.database")
	defer dbSpan.End()

	logger.Infof("Querying database %s", databaseName)

	dataSource, exists := uc.ExternalDataSources[databaseName]
	if !exists {
		libOtel.HandleSpanError(&dbSpan, "Unknown data source.", nil)
		logger.Errorf("Unknown data source: %s", databaseName)

		return nil // Continue with the next database
	}

	if !dataSource.Initialized {
		if err := pkg.ConnectToDataSource(databaseName, &dataSource, logger, uc.ExternalDataSources); err != nil {
			libOtel.HandleSpanError(&dbSpan, "Error initializing database connection.", err)
			return err
		}
	}

	// Prepare a result map for this database
	if _, databaseExists := result[databaseName]; !databaseExists {
		result[databaseName] = make(map[string][]map[string]any)
	}

	// Get filters for this database
	databaseFilters := allFilters[databaseName]

	switch dataSource.DatabaseType {
	case pkg.PostgreSQLType:
		return uc.queryPostgresDatabase(ctx, &dataSource, databaseName, tables, databaseFilters, result, logger)
	case pkg.MongoDBType:
		return uc.queryMongoDatabase(ctx, &dataSource, databaseName, tables, databaseFilters, result, logger)
	default:
		return fmt.Errorf("unsupported database type: %s for database: %s", dataSource.DatabaseType, databaseName)
	}
}

// queryPostgresDatabase handles querying PostgresSQL databases
func (uc *UseCase) queryPostgresDatabase(
	ctx context.Context,
	dataSource *pkg.DataSource,
	databaseName string,
	tables map[string][]string,
	databaseFilters map[string]map[string][]any,
	result map[string]map[string][]map[string]any,
	logger log.Logger,
) error {
	schema, err := dataSource.PostgresRepository.GetDatabaseSchema(ctx)
	if err != nil {
		logger.Errorf("Error getting database schema: %s", err.Error())
		return err
	}

	for table, fields := range tables {
		filter := getTableFilters(databaseFilters, table)

		tableResult, err := dataSource.PostgresRepository.Query(ctx, schema, table, fields, filter)
		if err != nil {
			logger.Errorf("Error querying table %s: %s", table, err.Error())
			return err
		}

		// Add the query results to the result map
		result[databaseName][table] = tableResult
	}

	return nil
}

// queryMongoDatabase handles querying MongoDB databases
func (uc *UseCase) queryMongoDatabase(
	ctx context.Context,
	dataSource *pkg.DataSource,
	databaseName string,
	collections map[string][]string,
	databaseFilters map[string]map[string][]any,
	result map[string]map[string][]map[string]any,
	logger log.Logger,
) error {
	tracer := libCommons.NewTracerFromContext(ctx)
	reqId := libCommons.NewHeaderIDFromContext(ctx)

	_, span := tracer.Start(ctx, "service.generate_report.query_mongo_database")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.database_name", databaseName),
	)

	for collection, fields := range collections {
		if (databaseName == "plugin_crm" && collection != "organization") || databaseName != "plugin_crm" {
			newCollection := collection
			if databaseName == "plugin_crm" {
				newCollection = collection + "_" + collections["organization"][0]
			}

			filter := getTableFilters(databaseFilters, collection)

			// Transform filters for plugin_crm to use search fields
			if databaseName == "plugin_crm" {
				transformedFilter, err := uc.transformPluginCRMFilters(filter, logger)
				if err != nil {
					logger.Errorf("Error transforming filters for collection %s: %s", collection, err.Error())
					return err
				}

				filter = transformedFilter
			}

			collectionResult, err := dataSource.MongoDBRepository.Query(ctx, newCollection, fields, filter)
			if err != nil {
				logger.Errorf("Error querying collection %s: %s", collection, err.Error())
				return err
			}

			if databaseName == "plugin_crm" {
				decryptedResult, err := uc.decryptPluginCRMData(logger, collectionResult, fields)
				if err != nil {
					logger.Errorf("Error decrypting data for collection %s: %s", collection, err.Error())
					return err
				}

				// Add the query results to the result map
				result[databaseName][collection] = decryptedResult
			} else {
				// Add the query results to the result map
				result[databaseName][collection] = collectionResult
			}
		}
	}

	return nil
}

// getTableFilters extracts filters for a specific table/collection
func getTableFilters(databaseFilters map[string]map[string][]any, tableName string) map[string][]any {
	if databaseFilters == nil {
		return nil
	}

	return databaseFilters[tableName]
}

// transformPluginCRMFilters transforms filters for plugin_crm to use search fields instead of encrypted fields
func (uc *UseCase) transformPluginCRMFilters(filter map[string][]any, logger log.Logger) (map[string][]any, error) {
	if filter == nil {
		return nil, nil
	}

	// Initialize crypto instance for hashing
	hashSecretKey := os.Getenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")
	if hashSecretKey == "" {
		return nil, fmt.Errorf("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM environment variable not set")
	}

	crypto := &libCrypto.Crypto{
		HashSecretKey: hashSecretKey,
		Logger:        logger,
	}

	transformedFilter := make(map[string][]any)

	// Define field mappings: encrypted field -> search field
	fieldMappings := map[string]string{
		"document":                "search.document",
		"name":                    "search.name",
		"banking_details.account": "search.banking_details_account",
		"banking_details.iban":    "search.banking_details_iban",
		"contact.primary_email":   "search.contact_primary_email",
		"contact.secondary_email": "search.contact_secondary_email",
		"contact.mobile_phone":    "search.contact_mobile_phone",
		"contact.other_phone":     "search.contact_other_phone",
	}

	for fieldName, values := range filter {
		if searchField, exists := fieldMappings[fieldName]; exists {
			// Transform values to hashes
			hashedValues := make([]any, len(values))

			for i, value := range values {
				if strValue, ok := value.(string); ok && strValue != "" {
					hash := crypto.GenerateHash(&strValue)
					hashedValues[i] = hash
					logger.Infof("Transformed filter: %s = %s -> %s = %s", fieldName, strValue, searchField, hash)
				} else {
					hashedValues[i] = value // Keep non-string values as-is
				}
			}

			transformedFilter[searchField] = hashedValues
		} else {
			// Keep non-mapped fields as-is
			transformedFilter[fieldName] = values
		}
	}

	return transformedFilter, nil
}

// getContentType returns the MIME type for a given file extension.
// If the extension is not recognized, it returns "text/plain".
func getContentType(ext string) string {
	if contentType, ok := mimeTypes[ext]; ok {
		return contentType
	}

	return "text/plain"
}

// decryptPluginCRMData decrypts sensitive fields for plugin_crm database
func (uc *UseCase) decryptPluginCRMData(logger log.Logger, collectionResult []map[string]any, fields []string) ([]map[string]any, error) {
	// Check if we need to decrypt any fields
	needsDecryption := false

	for _, field := range fields {
		// Check for top-level encrypted fields
		if isEncryptedField(field) {
			needsDecryption = true
			break
		}
		// Check for nested fields that might need decryption
		if strings.Contains(field, ".") {
			needsDecryption = true
			break
		}
	}

	if !needsDecryption {
		return collectionResult, nil
	}

	// Initialize crypto instance
	hashSecretKey := os.Getenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")
	encryptSecretKey := os.Getenv("CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM")
	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashSecretKey,
		EncryptSecretKey: encryptSecretKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	if err != nil {
		panic(err)
	}

	// Process each record in the collection
	for i, record := range collectionResult {
		decryptedRecord, err := uc.decryptRecord(record, fields, crypto)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt record %d: %w", i, err)
		}

		collectionResult[i] = decryptedRecord
	}

	return collectionResult, nil
}

// isEncryptedField checks if a field is known to be encrypted in plugin_crm
func isEncryptedField(field string) bool {
	encryptedFields := map[string]bool{
		"document": true,
		"name":     true,
	}

	return encryptedFields[field]
}

// decryptRecord decrypts a single record's encrypted fields
func (uc *UseCase) decryptRecord(record map[string]any, fields []string, crypto *libCrypto.Crypto) (map[string]any, error) {
	// Create a copy of the record to avoid modifying the original
	decryptedRecord := make(map[string]any)
	for k, v := range record {
		decryptedRecord[k] = v
	}

	// Decrypt top-level fields
	for fieldName, fieldValue := range decryptedRecord {
		if isEncryptedField(fieldName) && fieldValue != nil {
			if strValue, ok := fieldValue.(string); ok && strValue != "" {
				decryptedValue, err := crypto.Decrypt(&strValue)
				if err != nil {
					return nil, fmt.Errorf("failed to decrypt field %s: %w", fieldName, err)
				}

				decryptedRecord[fieldName] = *decryptedValue
			}
		}
	}

	// Decrypt nested fields in contact
	if contact, ok := decryptedRecord["contact"].(map[string]any); ok {
		contactFields := []string{"primary_email", "secondary_email", "mobile_phone", "other_phone"}
		for _, fieldName := range contactFields {
			if fieldValue, exists := contact[fieldName]; exists && fieldValue != nil {
				if strValue, ok := fieldValue.(string); ok && strValue != "" {
					decryptedValue, err := crypto.Decrypt(&strValue)
					if err != nil {
						return nil, fmt.Errorf("failed to decrypt contact.%s: %w", fieldName, err)
					}

					contact[fieldName] = *decryptedValue
				}
			}
		}

		decryptedRecord["contact"] = contact
	}

	// Decrypt nested fields in banking_details
	if bankingDetails, ok := decryptedRecord["banking_details"].(map[string]any); ok {
		bankingFields := []string{"account", "iban"}
		for _, fieldName := range bankingFields {
			if fieldValue, exists := bankingDetails[fieldName]; exists && fieldValue != nil {
				if strValue, ok := fieldValue.(string); ok && strValue != "" {
					decryptedValue, err := crypto.Decrypt(&strValue)
					if err != nil {
						return nil, fmt.Errorf("failed to decrypt banking_details.%s: %w", fieldName, err)
					}

					bankingDetails[fieldName] = *decryptedValue
				}
			}
		}

		decryptedRecord["banking_details"] = bankingDetails
	}

	// Decrypt nested fields in legal_person.representative
	if legalPerson, ok := decryptedRecord["legal_person"].(map[string]any); ok {
		if representative, ok := legalPerson["representative"].(map[string]any); ok {
			representativeFields := []string{"name", "document", "email"}
			for _, fieldName := range representativeFields {
				if fieldValue, exists := representative[fieldName]; exists && fieldValue != nil {
					if strValue, ok := fieldValue.(string); ok && strValue != "" {
						decryptedValue, err := crypto.Decrypt(&strValue)
						if err != nil {
							return nil, fmt.Errorf("failed to decrypt legal_person.representative.%s: %w", fieldName, err)
						}

						representative[fieldName] = *decryptedValue
					}
				}
			}

			legalPerson["representative"] = representative
		}

		decryptedRecord["legal_person"] = legalPerson
	}

	// Decrypt nested fields in natural_person
	if naturalPerson, ok := decryptedRecord["natural_person"].(map[string]any); ok {
		naturalPersonFields := []string{"mother_name", "father_name"}
		for _, fieldName := range naturalPersonFields {
			if fieldValue, exists := naturalPerson[fieldName]; exists && fieldValue != nil {
				if strValue, ok := fieldValue.(string); ok && strValue != "" {
					decryptedValue, err := crypto.Decrypt(&strValue)
					if err != nil {
						return nil, fmt.Errorf("failed to decrypt natural_person.%s: %w", fieldName, err)
					}

					naturalPerson[fieldName] = *decryptedValue
				}
			}
		}

		decryptedRecord["natural_person"] = naturalPerson
	}

	return decryptedRecord, nil
}
