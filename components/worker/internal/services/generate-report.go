package services

import (
	"context"
	"encoding/json"
	libCommons "github.com/LerianStudio/lib-commons/commons"
	libOtel "github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"plugin-template-engine/components/worker/internal/adapters/postgres"
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

	OrganizationID uuid.UUID `json:"organizationId"`

	Ledgers []string `json:"ledgerIds"`

	// OutputFormat specifies the format of the generated report (e.g., html, csv, json)
	OutputFormat string `json:"outputFormat"`

	// DataQueries maps database names to tables and their fields
	// Format: map[databaseName]map[tableName][]fieldName
	// Example: {"onboarding": {"organization": ["name"], "ledger": ["id"]}}
	DataQueries map[string]map[string][]string `json:"mappedFields"`
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

	ctx, spanSaveReport := tracer.Start(ctx, "service.generate_report.save_report")

	outputFormat := strings.ToLower(message.OutputFormat)
	contentType := getContentType(outputFormat)
	objectName := message.ReportID.String() + "/" + time.Now().Format("20060102_150405") + "." + outputFormat

	err = uc.ReportFileRepo.Put(ctx, objectName, contentType, []byte(out))
	if err != nil {
		libOtel.HandleSpanError(&spanSaveReport, "Error putting report file.", err)

		logger.Errorf("Error putting report file: %s", err.Error())

		return err
	}

	spanSaveReport.End()

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
		ctx, spanDatabase := tracer.Start(ctx, "service.generate_report.query_external_data.database")

		logger.Infof("Querying database %s", databaseName)

		dataSource, exists := uc.ExternalDataSources[databaseName]
		if !exists {
			libOtel.HandleSpanError(&spanDatabase, "Unknown data source.", nil)

			logger.Errorf("Unknown data source: %s", databaseName)

			continue
		}

		// Initialize the database connection if not already initialized
		if !dataSource.Initialized && dataSource.DatabaseType == "postgres" {
			// Create repository and update the data source
			dataSource.PostgresRepository = postgres.NewRepository(dataSource.DatabaseConfig)
			dataSource.Initialized = true

			// Update the data source in the map
			uc.ExternalDataSources[databaseName] = dataSource

			logger.Infof("Established connection to %s database on demand", databaseName)
		}

		// Initialize inner map for this database
		if _, exists := result[databaseName]; !exists {
			result[databaseName] = make(map[string][]map[string]any)
		}

		for table, fields := range tables {
			ctx, spanTable := tracer.Start(ctx, "service.generate_report.query_external_data.table")

			var tableResult []map[string]any

			var err error

			if dataSource.DatabaseType == "postgres" && dataSource.PostgresRepository != nil {
				tableResult, err = dataSource.PostgresRepository.Query(ctx, message.OrganizationID, table, message.Ledgers, fields)
				if err != nil {
					libOtel.HandleSpanError(&spanTable, "Error querying table.", err)

					logger.Errorf("Error querying table %s: %s", table, err.Error())

					return err
				}
			} else if dataSource.DatabaseType == "mongodb" {
				libOtel.HandleSpanError(&spanTable, "MongoDB queries not yet implemented.", nil)

				// TODO: Implement MongoDB query
				logger.Warnf("MongoDB queries not yet implemented for table: %s", table)

				continue // Skip for now
			}

			// Set the query results for this table
			result[databaseName][table] = tableResult

			spanTable.End()
		}

		spanDatabase.End()
	}

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
