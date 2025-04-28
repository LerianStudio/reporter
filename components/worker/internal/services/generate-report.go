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
// It supports querying PostgreSQL databases and skips MongoDB queries with a warning logger until implemented.
// Returns an error if there is an issue with querying any of the data sources.
func (uc *UseCase) queryExternalData(ctx context.Context, message GenerateReportMessage, result map[string]map[string][]map[string]any) error {
	logger := libCommons.NewLoggerFromContext(ctx)
	tracer := libCommons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.generate_report.query_external_data")
	defer span.End()

	for databaseName, tables := range message.DataQueries {
		_, dbSpan := tracer.Start(ctx, "service.generate_report.query_external_data.database")

		logger.Infof("Querying database %s", databaseName)

		dataSource, exists := uc.ExternalDataSources[databaseName]
		if !exists {
			libOtel.HandleSpanError(&dbSpan, "Unknown data source.", nil)
			logger.Errorf("Unknown data source: %s", databaseName)

			continue
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

		switch dataSource.DatabaseType {
		case pkg.PostgreSQLType:
			schema, err := dataSource.PostgresRepository.GetDatabaseSchema(ctx)
			if err != nil {
				logger.Errorf("Error getting database schema: %s", err.Error())
				return err
			}

			for table, fields := range tables {
				var tableResult []map[string]any

				var err error

				var filter map[string][]any
				if databaseFilters, filtersExist := message.Filters[databaseName]; filtersExist {
					filter = databaseFilters[table] // Get filters specific to this table
				}

				// Pass the table-specific filters to the Query method
				tableResult, err = dataSource.PostgresRepository.Query(ctx, schema, table, fields, filter)
				if err != nil {
					logger.Errorf("Error querying table %s: %s", table, err.Error())
					return err
				}

				// Add the query results to the result map
				result[databaseName][table] = tableResult
			}
		case "mongodb":
			logger.Warnf("MongoDB queries not yet implemented.")

			continue
		}

		dbSpan.End()
	}

	return nil
}

// connectToDataSource establishes a connection to a data source if not already initialized.
func (uc *UseCase) connectToDataSource(databaseName string, dataSource *DataSource, logger log.Logger) error {
	switch dataSource.DatabaseType {
	case pkg.PostgreSQLType:
		dataSource.PostgresRepository = postgres.NewDataSourceRepository(dataSource.DatabaseConfig)

		logger.Infof("Established PostgreSQL connection to %s database", databaseName)
	case "mongodb":
		return fmt.Errorf("MongoDB connections not yet implemented")
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
