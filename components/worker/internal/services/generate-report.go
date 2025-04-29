package services

import (
	"context"
	"encoding/json"
	"fmt"
	libCommons "github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/log"
	libOtel "github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"plugin-template-engine/components/worker/internal/adapters/mongodb"
	"plugin-template-engine/components/worker/internal/adapters/postgres"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/pongo"
	"strings"
	"time"
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
		libOtel.HandleSpanError(&span, "Error unmarshalling message.", err)

		logger.Errorf("Error unmarshalling message: %s", err.Error())

		return err
	}

	ctx, spanTemplate := tracer.Start(ctx, "service.generate_report.get_template")

	fileBytes, err := uc.TemplateFileRepo.Get(ctx, message.TemplateID.String())
	if err != nil {
		libOtel.HandleSpanError(&spanTemplate, "Error getting file from template bucket.", err)

		logger.Errorf("Error getting file from template bucket: %s", err.Error())

		return err
	}

	logger.Infof("Template found: %s", string(fileBytes))

	spanTemplate.End()

	result := make(map[string]map[string][]map[string]any)

	err = uc.queryExternalData(ctx, message, result)
	if err != nil {
		logger.Errorf("Error querying external data: %s", err.Error())

		return err
	}

	ctx, spanRender := tracer.Start(ctx, "service.generate_report.render_template")

	renderer := pongo.NewTemplateRenderer()

	out, err := renderer.RenderFromBytes(ctx, fileBytes, result)
	if err != nil {
		libOtel.HandleSpanError(&spanRender, "Error rendering template.", err)

		logger.Errorf("Error rendering template: %s", err.Error())

		return err
	}

	spanRender.End()

	err = uc.saveReport(ctx, tracer, message, out, logger)
	if err != nil {
		libOtel.HandleSpanError(&span, "Error saving report.", err)

		logger.Errorf("Error saving report: %s", err.Error())

		return err
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
	objectName := message.ReportID.String() + "/" + time.Now().Format("20060102_150405") + "." + outputFormat

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

		return nil // Continue with next database
	}

	if !dataSource.Initialized {
		if err := uc.connectToDataSource(databaseName, &dataSource, logger); err != nil {
			libOtel.HandleSpanError(&dbSpan, "Error initializing database connection.", err)
			return err
		}
	}

	// Prepare results map for this database
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

// queryPostgresDatabase handles querying PostgreSQL databases
func (uc *UseCase) queryPostgresDatabase(
	ctx context.Context,
	dataSource *DataSource,
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
	dataSource *DataSource,
	databaseName string,
	collections map[string][]string,
	databaseFilters map[string]map[string][]any,
	result map[string]map[string][]map[string]any,
	logger log.Logger,
) error {
	for collection, fields := range collections {
		filter := getTableFilters(databaseFilters, collection)

		collectionResult, err := dataSource.MongoDBRepository.Query(ctx, collection, fields, filter)
		if err != nil {
			logger.Errorf("Error querying collection %s: %s", collection, err.Error())
			return err
		}

		// Add the query results to the result map
		result[databaseName][collection] = collectionResult
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

// connectToDataSource establishes a connection to a data source if not already initialized.
func (uc *UseCase) connectToDataSource(databaseName string, dataSource *DataSource, logger log.Logger) error {
	switch dataSource.DatabaseType {
	case pkg.PostgreSQLType:
		dataSource.PostgresRepository = postgres.NewDataSourceRepository(dataSource.DatabaseConfig)

		logger.Infof("Established PostgreSQL connection to %s database", databaseName)

	case pkg.MongoDBType:
		dataSource.MongoDBRepository = mongodb.NewDataSourceRepository(dataSource.MongoURI, dataSource.MongoDBName)

		logger.Infof("Established MongoDB connection to %s database", databaseName)

	default:
		return fmt.Errorf("unsupported database type: %s for database: %s", dataSource.DatabaseType, databaseName)
	}

	dataSource.Initialized = true
	uc.ExternalDataSources[databaseName] = *dataSource

	return nil
}

// getContentType returns the MIME type for a given file extension.
// If the extension is not recognized, it returns "text/plain".
func getContentType(ext string) string {
	if contentType, ok := mimeTypes[ext]; ok {
		return contentType
	}

	return "text/plain"
}
