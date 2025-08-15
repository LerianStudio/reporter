package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"plugin-smart-templates/v2/pkg"
	"plugin-smart-templates/v2/pkg/constant"
	"plugin-smart-templates/v2/pkg/model"
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

	// Filters specify advanced filtering criteria using FilterCondition for complex queries.
	// Format: map[databaseName]map[tableName]map[fieldName]model.FilterCondition
	// Example: {"db": {"table": {"created_at": {"gte": ["2025-06-01"], "lte": ["2025-06-30"]}}}}
	Filters map[string]map[string]map[string]model.FilterCondition `json:"filters"`
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
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)

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

	out, err := renderer.RenderFromBytes(ctx, fileBytes, result, logger)
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
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)

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
	allFilters map[string]map[string]map[string]model.FilterCondition,
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
	databaseFilters map[string]map[string]model.FilterCondition,
	result map[string]map[string][]map[string]any,
	logger log.Logger,
) error {
	schema, err := dataSource.PostgresRepository.GetDatabaseSchema(ctx)
	if err != nil {
		logger.Errorf("Error getting database schema: %s", err.Error())
		return err
	}

	for table, fields := range tables {
		tableFilters := getTableFilters(databaseFilters, table)

		// Use advanced filters
		if len(tableFilters) > 0 {
			tableResult, err := dataSource.PostgresRepository.QueryWithAdvancedFilters(ctx, schema, table, fields, tableFilters)
			if err != nil {
				logger.Errorf("Error querying table %s with advanced filters: %s", table, err.Error())
				return err
			}

			logger.Infof("Successfully queried table %s with advanced filters", table)

			result[databaseName][table] = tableResult
		} else {
			// No filters, use legacy method for now
			tableResult, err := dataSource.PostgresRepository.Query(ctx, schema, table, fields, nil)
			if err != nil {
				logger.Errorf("Error querying table %s: %s", table, err.Error())
				return err
			}

			result[databaseName][table] = tableResult
		}
	}

	return nil
}

// queryMongoDatabase handles querying MongoDB databases
func (uc *UseCase) queryMongoDatabase(
	ctx context.Context,
	dataSource *pkg.DataSource,
	databaseName string,
	collections map[string][]string,
	databaseFilters map[string]map[string]model.FilterCondition,
	result map[string]map[string][]map[string]any,
	logger log.Logger,
) error {
	_, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

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

<<<<<<< HEAD
			collectionFilters := getTableFilters(databaseFilters, collection)
			if (databaseName == "plugin_crm" && collection != "organization") || databaseName != "plugin_crm" {
				newCollection := collection
				if databaseName == "plugin_crm" {
					newCollection = collection + "_" + collections["organization"][0]
				}
=======
			logger.Infof("Successfully queried collection %s with advanced filters", collection)

			result[databaseName][collection] = collectionResult
		} else {
			// No filters, use legacy method for now
			collectionResult, err := dataSource.MongoDBRepository.Query(ctx, collection, fields, nil)
			if err != nil {
				logger.Errorf("Error querying collection %s: %s", collection, err.Error())
				return err
			}
>>>>>>> 908f9bb (style: lint fixes :gem:)

				filter := getTableFilters(databaseFilters, collection)

				// Use advanced filters
				if len(collectionFilters) > 0 {
					if databaseName == "plugin_crm" {
						transformedFilter, err := uc.transformPluginCRMFilters(collectionFilters, logger)
						if err != nil {
							logger.Errorf("Error transforming filters for collection %s: %s", collection, err.Error())
							return err
						}

						collectionFilters = transformedFilter
					}

					collectionResult, err := dataSource.MongoDBRepository.QueryWithAdvancedFilters(ctx, collection, fields, collectionFilters)
					if err != nil {
						logger.Errorf("Error querying collection %s with advanced filters: %s", collection, err.Error())
						return err
					}

					logger.Infof("Successfully queried collection %s with advanced filters", collection)
					result[databaseName][collection] = collectionResult
				} else {
					// No filters, use legacy method for now
					collectionResult, err := dataSource.MongoDBRepository.Query(ctx, collection, fields, nil)
					if err != nil {
						logger.Errorf("Error querying collection %s: %s", collection, err.Error())
						return err
					}
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

					result[databaseName][collection] = collectionResult
				}
			}
			if databaseName == "plugin_crm" {
				decryptedResult, err := uc.decryptPluginCRMData(logger, collectionResult, fields)
				if err != nil {
					logger.Errorf("Error decrypting data for collection %s: %s", collection, err.Error())
					return pkg.ValidateBusinessError(constant.ErrDecryptionData, "", err)
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
func getTableFilters(databaseFilters map[string]map[string]model.FilterCondition, tableName string) map[string]model.FilterCondition {
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
	if encryptSecretKey == "" {
		return nil, fmt.Errorf("CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM environment variable not set")
	}

	if hashSecretKey == "" {
		return nil, fmt.Errorf("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM environment variable not set")
	}

	crypto := &libCrypto.Crypto{
		HashSecretKey:    hashSecretKey,
		EncryptSecretKey: encryptSecretKey,
		Logger:           logger,
	}

	err := crypto.InitializeCipher()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cipher: %w", err)
	}

	// Process each record in the collection
	for i, record := range collectionResult {
		decryptedRecord, err := uc.decryptRecord(record, crypto)
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
func (uc *UseCase) decryptRecord(record map[string]any, crypto *libCrypto.Crypto) (map[string]any, error) {
	// Create a copy of the record to avoid modifying the original
	decryptedRecord := make(map[string]any)
	for k, v := range record {
		decryptedRecord[k] = v
	}

	// Decrypt top-level fields
	if err := uc.decryptTopLevelFields(decryptedRecord, crypto); err != nil {
		return nil, err
	}

	// Decrypt nested fields
	if err := uc.decryptNestedFields(decryptedRecord, crypto); err != nil {
		return nil, err
	}

	return decryptedRecord, nil
}

// decryptTopLevelFields decrypts top-level encrypted fields
func (uc *UseCase) decryptTopLevelFields(record map[string]any, crypto *libCrypto.Crypto) error {
	for fieldName, fieldValue := range record {
		if isEncryptedField(fieldName) && fieldValue != nil {
			if err := uc.decryptFieldValue(record, fieldName, fieldValue, crypto); err != nil {
				return fmt.Errorf("failed to decrypt field %s: %w", fieldName, err)
			}
		}
	}

	return nil
}

// decryptNestedFields decrypts nested encrypted fields in the record
func (uc *UseCase) decryptNestedFields(record map[string]any, crypto *libCrypto.Crypto) error {
	if err := uc.decryptContactFields(record, crypto); err != nil {
		return err
	}

	if err := uc.decryptBankingDetailsFields(record, crypto); err != nil {
		return err
	}

	if err := uc.decryptLegalPersonFields(record, crypto); err != nil {
		return err
	}

	if err := uc.decryptNaturalPersonFields(record, crypto); err != nil {
		return err
	}

	return nil
}

// decryptContactFields decrypts fields within the contact object
func (uc *UseCase) decryptContactFields(record map[string]any, crypto *libCrypto.Crypto) error {
	contact, ok := record["contact"].(map[string]any)
	if !ok {
		return nil
	}

	contactFields := []string{"primary_email", "secondary_email", "mobile_phone", "other_phone"}
	for _, fieldName := range contactFields {
		if fieldValue, exists := contact[fieldName]; exists && fieldValue != nil {
			if err := uc.decryptFieldValue(contact, fieldName, fieldValue, crypto); err != nil {
				return fmt.Errorf("failed to decrypt contact.%s: %w", fieldName, err)
			}
		}
	}

	record["contact"] = contact

	return nil
}

// decryptBankingDetailsFields decrypts fields within the banking_details object
func (uc *UseCase) decryptBankingDetailsFields(record map[string]any, crypto *libCrypto.Crypto) error {
	bankingDetails, ok := record["banking_details"].(map[string]any)
	if !ok {
		return nil
	}

	bankingFields := []string{"account", "iban"}
	for _, fieldName := range bankingFields {
		if fieldValue, exists := bankingDetails[fieldName]; exists && fieldValue != nil {
			if err := uc.decryptFieldValue(bankingDetails, fieldName, fieldValue, crypto); err != nil {
				return fmt.Errorf("failed to decrypt banking_details.%s: %w", fieldName, err)
			}
		}
	}

	record["banking_details"] = bankingDetails

	return nil
}

// decryptLegalPersonFields decrypts fields within the legal_person object
func (uc *UseCase) decryptLegalPersonFields(record map[string]any, crypto *libCrypto.Crypto) error {
	legalPerson, ok := record["legal_person"].(map[string]any)
	if !ok {
		return nil
	}

	representative, ok := legalPerson["representative"].(map[string]any)
	if !ok {
		return nil
	}

	representativeFields := []string{"name", "document", "email"}
	for _, fieldName := range representativeFields {
		if fieldValue, exists := representative[fieldName]; exists && fieldValue != nil {
			if err := uc.decryptFieldValue(representative, fieldName, fieldValue, crypto); err != nil {
				return fmt.Errorf("failed to decrypt legal_person.representative.%s: %w", fieldName, err)
			}
		}
	}

	legalPerson["representative"] = representative
	record["legal_person"] = legalPerson

	return nil
}

// decryptNaturalPersonFields decrypts fields within the natural_person object
func (uc *UseCase) decryptNaturalPersonFields(record map[string]any, crypto *libCrypto.Crypto) error {
	naturalPerson, ok := record["natural_person"].(map[string]any)
	if !ok {
		return nil
	}

	naturalPersonFields := []string{"mother_name", "father_name"}
	for _, fieldName := range naturalPersonFields {
		if fieldValue, exists := naturalPerson[fieldName]; exists && fieldValue != nil {
			if err := uc.decryptFieldValue(naturalPerson, fieldName, fieldValue, crypto); err != nil {
				return fmt.Errorf("failed to decrypt natural_person.%s: %w", fieldName, err)
			}
		}
	}

	record["natural_person"] = naturalPerson

	return nil
}

// decryptFieldValue decrypts a single field value if it's a non-empty string
func (uc *UseCase) decryptFieldValue(container map[string]any, fieldName string, fieldValue any, crypto *libCrypto.Crypto) error {
	strValue, ok := fieldValue.(string)
	if !ok || strValue == "" {
		return nil
	}

	decryptedValue, err := crypto.Decrypt(&strValue)
	if err != nil {
		return err
	}

	container[fieldName] = *decryptedValue

	return nil
}
