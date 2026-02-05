// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/model"
	"github.com/LerianStudio/reporter/pkg/pongo"
	"github.com/LerianStudio/reporter/pkg/postgres"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libConstants "github.com/LerianStudio/lib-commons/v2/commons/constants"
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
}

// GenerateReport handles a report generation request by loading a template file,
// processing it, and storing the final report in the report repository.
func (uc *UseCase) GenerateReport(ctx context.Context, body []byte) error {
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.generate_report")
	defer span.End()

	message, err := uc.parseMessage(ctx, body, &span, logger)
	if err != nil {
		return err
	}

	if skip := uc.shouldSkipProcessing(ctx, message.ReportID, logger); skip {
		return nil
	}

	templateBytes, err := uc.loadTemplate(ctx, tracer, message, &span, logger)
	if err != nil {
		return err
	}

	result := make(map[string]map[string][]map[string]any)

	if err := uc.queryExternalData(ctx, message, result); err != nil {
		return uc.handleErrorWithUpdate(ctx, message.ReportID, &span, "Error querying external data", err, logger)
	}

	renderedOutput, err := uc.renderTemplate(ctx, tracer, templateBytes, result, message, &span, logger)
	if err != nil {
		return err
	}

	finalOutput, err := uc.convertToPDFIfNeeded(ctx, tracer, message, renderedOutput, &span, logger)
	if err != nil {
		return err
	}

	if err := uc.saveReport(ctx, tracer, message, finalOutput, logger); err != nil {
		return uc.handleErrorWithUpdate(ctx, message.ReportID, &span, "Error saving report", err, logger)
	}

	if err := uc.markReportAsFinished(ctx, message.ReportID, &span, logger); err != nil {
		return err
	}

	return nil
}

// parseMessage parses the RabbitMQ message body into GenerateReportMessage struct.
func (uc *UseCase) parseMessage(ctx context.Context, body []byte, span *trace.Span, logger log.Logger) (GenerateReportMessage, error) {
	var message GenerateReportMessage

	err := json.Unmarshal(body, &message)
	if err != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, message.ReportID, err.Error()); errUpdate != nil {
			libOtel.HandleSpanError(span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return message, errUpdate
		}

		libOtel.HandleSpanError(span, "Error unmarshalling message.", err)
		logger.Errorf("Error unmarshalling message: %s", err.Error())

		return message, err
	}

	return message, nil
}

// shouldSkipProcessing checks if report should be skipped due to idempotency.
func (uc *UseCase) shouldSkipProcessing(ctx context.Context, reportID uuid.UUID, logger log.Logger) bool {
	reportStatus, err := uc.checkReportStatus(ctx, reportID, logger)
	if err == nil {
		if reportStatus == constant.FinishedStatus {
			logger.Infof("Report %s is already finished, skipping reprocessing", reportID)
			return true
		}

		if reportStatus == constant.ErrorStatus {
			logger.Warnf("Report %s is in error state, skipping reprocessing", reportID)
			return true
		}
	}

	return false
}

// loadTemplate loads template file from SeaweedFS.
func (uc *UseCase) loadTemplate(ctx context.Context, tracer trace.Tracer, message GenerateReportMessage, span *trace.Span, logger log.Logger) ([]byte, error) {
	ctx, spanTemplate := tracer.Start(ctx, "service.generate_report.get_template")
	defer spanTemplate.End()

	fileBytes, err := uc.TemplateSeaweedFS.Get(ctx, message.TemplateID.String())
	if err != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, message.ReportID, err.Error()); errUpdate != nil {
			libOtel.HandleSpanError(span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return nil, errUpdate
		}

		libOtel.HandleSpanError(&spanTemplate, "Error getting file from template bucket.", err)
		logger.Errorf("Error getting file from template bucket: %s", err.Error())

		return nil, err
	}

	logger.Infof("Template found: %s", string(fileBytes))

	return fileBytes, nil
}

// renderTemplate renders the template with data from external sources.
func (uc *UseCase) renderTemplate(ctx context.Context, tracer trace.Tracer, templateBytes []byte, result map[string]map[string][]map[string]any, message GenerateReportMessage, span *trace.Span, logger log.Logger) (string, error) {
	ctx, spanRender := tracer.Start(ctx, "service.generate_report.render_template")
	defer spanRender.End()

	renderer := pongo.NewTemplateRenderer()

	out, err := renderer.RenderFromBytes(ctx, templateBytes, result, logger)
	if err != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, message.ReportID, err.Error()); errUpdate != nil {
			libOtel.HandleSpanError(span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return "", errUpdate
		}

		libOtel.HandleSpanError(&spanRender, "Error rendering template.", err)
		logger.Errorf("Error rendering template: %s", err.Error())

		return "", err
	}

	return out, nil
}

// convertToPDFIfNeeded converts HTML to PDF if output format is PDF.
func (uc *UseCase) convertToPDFIfNeeded(ctx context.Context, tracer trace.Tracer, message GenerateReportMessage, htmlOutput string, span *trace.Span, logger log.Logger) (string, error) {
	if strings.ToLower(message.OutputFormat) != "pdf" {
		return htmlOutput, nil
	}

	_, spanPDF := tracer.Start(ctx, "service.generate_report.convert_to_pdf")
	defer spanPDF.End()

	logger.Infof("Converting HTML to PDF for report %s (HTML size: %d bytes)", message.ReportID, len(htmlOutput))

	pdfBytes, err := uc.convertHTMLToPDF(htmlOutput, logger)
	if err != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, message.ReportID, err.Error()); errUpdate != nil {
			libOtel.HandleSpanError(span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return "", errUpdate
		}

		libOtel.HandleSpanError(&spanPDF, "Error converting HTML to PDF.", err)
		logger.Errorf("Error converting HTML to PDF: %s", err.Error())

		return "", err
	}

	logger.Infof("PDF generated successfully (PDF size: %d bytes)", len(pdfBytes))

	return string(pdfBytes), nil
}

// markReportAsFinished updates report status to finished.
func (uc *UseCase) markReportAsFinished(ctx context.Context, reportID uuid.UUID, span *trace.Span, logger log.Logger) error {
	err := uc.ReportDataRepo.UpdateReportStatusById(ctx, constant.FinishedStatus, reportID, time.Now(), nil)
	if err != nil {
		if errUpdate := uc.updateReportWithErrors(ctx, reportID, err.Error()); errUpdate != nil {
			libOtel.HandleSpanError(span, "Error to update report status with error.", errUpdate)
			logger.Errorf("Error update report status with error: %s", errUpdate.Error())

			return errUpdate
		}

		libOtel.HandleSpanError(span, "Error to update report status.", err)
		logger.Errorf("Error saving report: %s", err.Error())

		return err
	}

	return nil
}

// handleErrorWithUpdate logs error and updates report status to error.
func (uc *UseCase) handleErrorWithUpdate(ctx context.Context, reportID uuid.UUID, span *trace.Span, errorMsg string, err error, logger log.Logger) error {
	if errUpdate := uc.updateReportWithErrors(ctx, reportID, err.Error()); errUpdate != nil {
		libOtel.HandleSpanError(span, "Error to update report status with error.", errUpdate)
		logger.Errorf("Error update report status with error: %s", errUpdate.Error())

		return errUpdate
	}

	libOtel.HandleSpanError(span, errorMsg, err)
	logger.Errorf("%s: %s", errorMsg, err.Error())

	return err
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
// It determines the object name, content type, and stores the file using the ReportSeaweedFS interface.
// If ReportTTL is configured, the file will be saved with TTL (Time To Live).
// Returns an error if the file storage operation fails.
func (uc *UseCase) saveReport(ctx context.Context, tracer trace.Tracer, message GenerateReportMessage, out string, logger log.Logger) error {
	ctx, spanSaveReport := tracer.Start(ctx, "service.generate_report.save_report")
	defer spanSaveReport.End()

	outputFormat := strings.ToLower(message.OutputFormat)
	contentType := getContentType(outputFormat)
	objectName := message.TemplateID.String() + "/" + message.ReportID.String() + "." + outputFormat

	err := uc.ReportSeaweedFS.Put(ctx, objectName, contentType, []byte(out), uc.ReportTTL)
	if err != nil {
		libOtel.HandleSpanError(&spanSaveReport, "Error putting report file.", err)

		logger.Errorf("Error putting report file: %s", err.Error())

		return err
	}

	if uc.ReportTTL != "" {
		logger.Infof("Saving report with TTL: %s", uc.ReportTTL)
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

	// Check circuit breaker state before attempting query
	if !uc.CircuitBreakerManager.IsHealthy(databaseName) {
		cbState := uc.CircuitBreakerManager.GetState(databaseName)
		err := fmt.Errorf("datasource %s is unhealthy - circuit breaker state: %s", databaseName, cbState)
		libOtel.HandleSpanError(&dbSpan, "Circuit breaker blocking request", err)
		logger.Errorf("⚠️  Circuit breaker blocking request to datasource %s (state: %s)", databaseName, cbState)

		return err
	}

	// Check datasource initialization status
	if !dataSource.Initialized {
		// Check if datasource is marked as unavailable from initialization
		if dataSource.Status == libConstants.DataSourceStatusUnavailable {
			err := fmt.Errorf("datasource %s is unavailable (initialization failed)", databaseName)
			libOtel.HandleSpanError(&dbSpan, "Datasource unavailable", err)
			logger.Errorf("⚠️  Datasource %s is unavailable - last error: %v", databaseName, dataSource.LastError)

			return err
		}

		// Attempt to connect
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
	// Use configured schemas or default to public
	configuredSchemas := dataSource.Schemas
	if len(configuredSchemas) == 0 {
		configuredSchemas = []string{"public"}
	}

	logger.Infof("Querying database %s with schemas: %v", databaseName, configuredSchemas)

	// Execute schema query with circuit breaker protection
	schemaResult, err := uc.CircuitBreakerManager.Execute(databaseName, func() (any, error) {
		return dataSource.PostgresRepository.GetDatabaseSchema(ctx, configuredSchemas)
	})
	if err != nil {
		logger.Errorf("Error getting database schema for %s (circuit breaker): %s", databaseName, err.Error())
		return err
	}

	schema := schemaResult.([]postgres.TableSchema)

	// Initialize SchemaResolver with discovered tables
	resolver := pkg.NewSchemaResolver()
	resolver.RegisterDatabase(databaseName, schema)

	for tableKey, fields := range tables {
		tableFilters := getTableFilters(databaseFilters, tableKey)

		// Parse table key to extract explicit schema if present
		// Supports multiple formats:
		// - "schema__table" (Pongo2 compatible format from CleanPath)
		// - "schema.table" (explicit qualified format)
		// - "table" (autodiscovery)
		var explicitSchema, tableName string

		if strings.Contains(tableKey, "__") {
			// Pongo2 format: schema__table -> split by double underscore
			parts := strings.SplitN(tableKey, "__", 2)
			explicitSchema = parts[0]
			tableName = parts[1]
		} else if strings.Contains(tableKey, ".") {
			// Qualified format: schema.table -> split by dot
			parts := strings.SplitN(tableKey, ".", 2)
			explicitSchema = parts[0]
			tableName = parts[1]
		} else {
			tableName = tableKey
		}

		// Resolve schema name for this table
		schemaName, err := resolver.ResolveSchema(databaseName, explicitSchema, tableName)
		if err != nil {
			// Check if it's an ambiguity error for actionable message
			if ambiguityErr, ok := err.(*pkg.SchemaAmbiguityError); ok {
				logger.Errorf("Schema ambiguity for table %s in %s: %s", tableName, databaseName, ambiguityErr.Error())
			} else {
				logger.Errorf("Error resolving schema for table %s in %s: %s", tableName, databaseName, err.Error())
			}

			return err
		}

		logger.Infof("Resolved schema '%s' for table '%s' in database '%s'", schemaName, tableName, databaseName)

		// Execute query with circuit breaker protection
		var tableResult []map[string]any

		queryResult, err := uc.CircuitBreakerManager.Execute(databaseName, func() (any, error) {
			if len(tableFilters) > 0 {
				return dataSource.PostgresRepository.QueryWithAdvancedFilters(ctx, schema, schemaName, tableName, fields, tableFilters)
			}

			return dataSource.PostgresRepository.Query(ctx, schema, schemaName, tableName, fields, nil)
		})
		if err != nil {
			logger.Errorf("Error querying table %s.%s in %s (circuit breaker): %s", schemaName, tableName, databaseName, err.Error())
			return err
		}

		tableResult = queryResult.([]map[string]any)

		if len(tableFilters) > 0 {
			logger.Infof("Successfully queried table %s.%s with advanced filters (circuit breaker: %s)",
				schemaName, tableName, uc.CircuitBreakerManager.GetState(databaseName))
		} else {
			logger.Infof("Successfully queried table %s.%s (circuit breaker: %s)",
				schemaName, tableName, uc.CircuitBreakerManager.GetState(databaseName))
		}

		// Store result using the original tableKey which is already in Pongo2-compatible format
		// (schema__table from CleanPath) for dot notation access in templates
		result[databaseName][tableKey] = tableResult
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
		collectionFilters := getTableFilters(databaseFilters, collection)

		if err := uc.processMongoCollection(ctx, dataSource, databaseName, collection, fields, collectionFilters, result, logger); err != nil {
			return err
		}
	}

	return nil
}

// processMongoCollection processes a single MongoDB collection
func (uc *UseCase) processMongoCollection(
	ctx context.Context,
	dataSource *pkg.DataSource,
	databaseName, collection string,
	fields []string,
	collectionFilters map[string]model.FilterCondition,
	result map[string]map[string][]map[string]any,
	logger log.Logger,
) error {
	// Handle plugin_crm special cases
	if databaseName == "plugin_crm" {
		// Skip "organization" collection - it's not a real collection, just stores the organizationID for template context
		if collection == "organization" {
			logger.Debugf("Skipping organization collection for plugin_crm - it's a metadata field, not a queryable collection")
			return nil
		}

		return uc.processPluginCRMCollection(ctx, dataSource, collection, fields, collectionFilters, result, logger)
	}

	// Handle regular collections
	return uc.processRegularMongoCollection(ctx, dataSource, collection, fields, collectionFilters, result, logger)
}

// processPluginCRMCollection handles plugin_crm specific collection processing
func (uc *UseCase) processPluginCRMCollection(
	ctx context.Context,
	dataSource *pkg.DataSource,
	collection string,
	fields []string,
	collectionFilters map[string]model.FilterCondition,
	result map[string]map[string][]map[string]any,
	logger log.Logger,
) error {
	// Get Midaz organization ID from datasource configuration (DATASOURCE_CRM_MIDAZ_ORGANIZATION_ID)
	if dataSource.MidazOrganizationID == "" {
		logger.Errorf("Midaz Organization ID not configured for plugin_crm datasource. Set DATASOURCE_CRM_MIDAZ_ORGANIZATION_ID environment variable.")
		return nil
	}

	newCollection := collection + "_" + dataSource.MidazOrganizationID

	// Query the collection
	collectionResult, err := uc.queryMongoCollectionWithFilters(ctx, dataSource, newCollection, fields, collectionFilters, logger, "plugin_crm")
	if err != nil {
		return err
	}

	result["plugin_crm"][collection] = collectionResult

	// Decrypt data for plugin_crm
	decryptedResult, err := uc.decryptPluginCRMData(logger, result["plugin_crm"][collection], fields)
	if err != nil {
		logger.Errorf("Error decrypting data for collection %s: %s", collection, err.Error())
		return pkg.ValidateBusinessError(constant.ErrDecryptionData, "", err)
	}

	result["plugin_crm"][collection] = decryptedResult

	return nil
}

// processRegularMongoCollection handles regular MongoDB collection processing
func (uc *UseCase) processRegularMongoCollection(
	ctx context.Context,
	dataSource *pkg.DataSource,
	collection string,
	fields []string,
	collectionFilters map[string]model.FilterCondition,
	result map[string]map[string][]map[string]any,
	logger log.Logger,
) error {
	// Determine database name from context (assuming it's available in the result map)
	var databaseName string
	for dbName := range result {
		databaseName = dbName
		break
	}

	collectionResult, err := uc.queryMongoCollectionWithFilters(ctx, dataSource, collection, fields, collectionFilters, logger, databaseName)
	if err != nil {
		return err
	}

	result[databaseName][collection] = collectionResult

	return nil
}

// queryMongoCollectionWithFilters queries a MongoDB collection with or without filters
func (uc *UseCase) queryMongoCollectionWithFilters(
	ctx context.Context,
	dataSource *pkg.DataSource,
	collection string,
	fields []string,
	collectionFilters map[string]model.FilterCondition,
	logger log.Logger,
	databaseName string,
) ([]map[string]any, error) {
	// Execute query with circuit breaker protection
	queryResult, err := uc.CircuitBreakerManager.Execute(databaseName, func() (any, error) {
		if len(collectionFilters) > 0 {
			// Check if this is plugin_crm and needs filter transformation
			if strings.Contains(collection, "_") && !strings.Contains(collection, "organization") {
				transformedFilter, err := uc.transformPluginCRMAdvancedFilters(collectionFilters, logger)
				if err != nil {
					return nil, fmt.Errorf("error transforming advanced filters for collection %s: %w", collection, err)
				}

				collectionFilters = transformedFilter
			}

			return dataSource.MongoDBRepository.QueryWithAdvancedFilters(ctx, collection, fields, collectionFilters)
		}

		// No filters, use legacy method
		return dataSource.MongoDBRepository.Query(ctx, collection, fields, nil)
	})
	if err != nil {
		logger.Errorf("Error querying collection %s in %s (circuit breaker): %s", collection, databaseName, err.Error())
		return nil, err
	}

	collectionResult := queryResult.([]map[string]any)

	if len(collectionFilters) > 0 {
		logger.Infof("Successfully queried collection %s with advanced filters (circuit breaker: %s)",
			collection, uc.CircuitBreakerManager.GetState(databaseName))
	} else {
		logger.Infof("Successfully queried collection %s (circuit breaker: %s)",
			collection, uc.CircuitBreakerManager.GetState(databaseName))
	}

	return collectionResult, nil
}

// getTableFilters extracts filters for a specific table/collection
// Supports multiple table name formats:
// - "schema__table" (Pongo2 format)
// - "schema.table" (qualified format)
// - "table" (simple format, will try with "public." prefix)
func getTableFilters(databaseFilters map[string]map[string]model.FilterCondition, tableName string) map[string]model.FilterCondition {
	if databaseFilters == nil {
		return nil
	}

	// Try exact match first
	if filters, ok := databaseFilters[tableName]; ok {
		return filters
	}

	// Try alternative formats
	var alternativeKeys []string

	if strings.Contains(tableName, "__") {
		// Pongo2 format: schema__table -> try schema.table
		alternativeKeys = append(alternativeKeys, strings.Replace(tableName, "__", ".", 1))
	} else if strings.Contains(tableName, ".") {
		// Qualified format: schema.table -> try schema__table
		alternativeKeys = append(alternativeKeys, strings.Replace(tableName, ".", "__", 1))
	} else {
		// Simple table name without schema -> try with public schema
		// This handles the case where template has "organization" but filter has "public.organization"
		alternativeKeys = append(alternativeKeys, "public."+tableName)
		alternativeKeys = append(alternativeKeys, "public__"+tableName)
	}

	for _, altKey := range alternativeKeys {
		if filters, ok := databaseFilters[altKey]; ok {
			return filters
		}
	}

	return nil
}

// transformPluginCRMAdvancedFilters transforms advanced FilterCondition filters for plugin_crm to use search fields
func (uc *UseCase) transformPluginCRMAdvancedFilters(filter map[string]model.FilterCondition, logger log.Logger) (map[string]model.FilterCondition, error) {
	if filter == nil {
		return nil, nil
	}

	hashSecretKey := os.Getenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM")
	if hashSecretKey == "" {
		return nil, fmt.Errorf("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM environment variable not set")
	}

	crypto := &libCrypto.Crypto{
		HashSecretKey: hashSecretKey,
		Logger:        logger,
	}

	transformedFilter := make(map[string]model.FilterCondition)

	// Define field mappings: encrypted field -> search field
	fieldMappings := map[string]string{
		"document":                               "search.document",
		"name":                                   "search.name",
		"banking_details.account":                "search.banking_details_account",
		"banking_details.iban":                   "search.banking_details_iban",
		"contact.primary_email":                  "search.contact_primary_email",
		"contact.secondary_email":                "search.contact_secondary_email",
		"contact.mobile_phone":                   "search.contact_mobile_phone",
		"contact.other_phone":                    "search.contact_other_phone",
		"regulatory_fields.participant_document": "search.regulatory_fields_participant_document",
		"related_parties.document":               "search.related_party_documents",
	}

	for fieldName, condition := range filter {
		if searchField, exists := fieldMappings[fieldName]; exists {
			// Transform the condition by hashing string values
			transformedCondition := model.FilterCondition{}

			// Transform Equals values
			if len(condition.Equals) > 0 {
				transformedCondition.Equals = uc.hashFilterValues(condition.Equals, crypto)
			}

			// Transform GreaterThan values
			if len(condition.GreaterThan) > 0 {
				transformedCondition.GreaterThan = uc.hashFilterValues(condition.GreaterThan, crypto)
			}

			// Transform GreaterOrEqual values
			if len(condition.GreaterOrEqual) > 0 {
				transformedCondition.GreaterOrEqual = uc.hashFilterValues(condition.GreaterOrEqual, crypto)
			}

			// Transform LessThan values
			if len(condition.LessThan) > 0 {
				transformedCondition.LessThan = uc.hashFilterValues(condition.LessThan, crypto)
			}

			// Transform LessOrEqual values
			if len(condition.LessOrEqual) > 0 {
				transformedCondition.LessOrEqual = uc.hashFilterValues(condition.LessOrEqual, crypto)
			}

			// Transform Between values
			if len(condition.Between) > 0 {
				transformedCondition.Between = uc.hashFilterValues(condition.Between, crypto)
			}

			// Transform In values
			if len(condition.In) > 0 {
				transformedCondition.In = uc.hashFilterValues(condition.In, crypto)
			}

			// Transform NotIn values
			if len(condition.NotIn) > 0 {
				transformedCondition.NotIn = uc.hashFilterValues(condition.NotIn, crypto)
			}

			transformedFilter[searchField] = transformedCondition

			logger.Infof("Transformed advanced filter: %s -> %s", fieldName, searchField)
		} else {
			// Keep non-mapped fields as-is
			transformedFilter[fieldName] = condition
		}
	}

	return transformedFilter, nil
}

// hashFilterValues hashes string values in a filter condition array
func (uc *UseCase) hashFilterValues(values []any, crypto *libCrypto.Crypto) []any {
	hashedValues := make([]any, len(values))

	for i, value := range values {
		if strValue, ok := value.(string); ok && strValue != "" {
			hash := crypto.GenerateHash(&strValue)
			hashedValues[i] = hash
		} else {
			hashedValues[i] = value // Keep non-string values as-is
		}
	}

	return hashedValues
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

	if err := uc.decryptRegulatoryFieldsFields(record, crypto); err != nil {
		return err
	}

	if err := uc.decryptRelatedPartiesFields(record, crypto); err != nil {
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

// decryptRegulatoryFieldsFields decrypts fields within the regulatory_fields object
func (uc *UseCase) decryptRegulatoryFieldsFields(record map[string]any, crypto *libCrypto.Crypto) error {
	regulatoryFields, ok := record["regulatory_fields"].(map[string]any)
	if !ok {
		return nil
	}

	regulatoryFieldNames := []string{"participant_document"}
	for _, fieldName := range regulatoryFieldNames {
		if fieldValue, exists := regulatoryFields[fieldName]; exists && fieldValue != nil {
			if err := uc.decryptFieldValue(regulatoryFields, fieldName, fieldValue, crypto); err != nil {
				return fmt.Errorf("failed to decrypt regulatory_fields.%s: %w", fieldName, err)
			}
		}
	}

	record["regulatory_fields"] = regulatoryFields

	return nil
}

// decryptRelatedPartiesFields decrypts the document field within each related_parties array item
func (uc *UseCase) decryptRelatedPartiesFields(record map[string]any, crypto *libCrypto.Crypto) error {
	relatedParties, ok := record["related_parties"].([]any)
	if !ok {
		return nil
	}

	for i, party := range relatedParties {
		partyMap, ok := party.(map[string]any)
		if !ok {
			continue
		}

		if fieldValue, exists := partyMap["document"]; exists && fieldValue != nil {
			if err := uc.decryptFieldValue(partyMap, "document", fieldValue, crypto); err != nil {
				return fmt.Errorf("failed to decrypt related_parties[%d].document: %w", i, err)
			}
		}

		relatedParties[i] = partyMap
	}

	record["related_parties"] = relatedParties

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

// checkReportStatus checks the current status of a report to implement idempotency.
func (uc *UseCase) checkReportStatus(ctx context.Context, reportID uuid.UUID, logger log.Logger) (string, error) {
	report, err := uc.ReportDataRepo.FindByID(ctx, reportID)
	if err != nil {
		logger.Debugf("Could not check report status for %s (may be first attempt): %v", reportID, err)
		return "", err
	}

	logger.Debugf("Report %s current status: %s", reportID, report.Status)

	return report.Status, nil
}

// convertHTMLToPDF converts HTML content to PDF using Chrome headless via PDF pool.
func (uc *UseCase) convertHTMLToPDF(htmlContent string, logger log.Logger) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "pdf-*.pdf")
	if err != nil {
		logger.Errorf("Failed to create temporary PDF file: %v", err)
		return nil, fmt.Errorf("failed to create temporary PDF file: %w", err)
	}

	tmpFileName := tmpFile.Name()
	if closeErr := tmpFile.Close(); closeErr != nil {
		logger.Warnf("Failed to close temporary file %s: %v", tmpFileName, closeErr)
	}

	defer func() {
		if removeErr := os.Remove(tmpFileName); removeErr != nil {
			logger.Warnf("Failed to remove temporary PDF file %s: %v", tmpFileName, removeErr)
		}
	}()

	err = uc.PdfPool.Submit(htmlContent, tmpFileName)
	if err != nil {
		logger.Errorf("Failed to generate PDF from HTML: %v", err)
		return nil, fmt.Errorf("failed to generate PDF from HTML: %w", err)
	}

	// Read generated PDF file - tmpFileName is safe as it comes from os.CreateTemp
	// #nosec G304 -- tmpFileName is generated by os.CreateTemp and is safe
	pdfBytes, err := os.ReadFile(tmpFileName)
	if err != nil {
		logger.Errorf("Failed to read generated PDF: %v", err)
		return nil, fmt.Errorf("failed to read generated PDF: %w", err)
	}

	return pdfBytes, nil
}
