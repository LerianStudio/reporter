package services

import (
	"context"
	"encoding/json"
	libCommons "github.com/LerianStudio/lib-commons/commons"
	"github.com/google/uuid"
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
		logger.Errorf("Error unmarshalling message: %s", err.Error())

		return err
	}

	fileBytes, err := uc.TemplateFileRepo.Get(ctx, message.TemplateID.String())
	if err != nil {
		logger.Errorf("Error getting file from template bucket: %s", err.Error())

		return err
	}

	logger.Infof("Template found: %s", string(fileBytes))

	result := make(map[string]map[string][]map[string]interface{})

	for databaseName, tables := range message.DataQueries {
		logger.Infof("Querying database %s", databaseName)

		dataSource := uc.ExternalDataSources[databaseName]

		// Initialize inner map for this database
		if _, exists := result[databaseName]; !exists {
			result[databaseName] = make(map[string][]map[string]interface{})
		}

		for table, fields := range tables {
			var tableResult []map[string]interface{}

			var err error

			if dataSource.DatabaseType == "mongodb" {
				// TODO: Implement MongoDB query
				logger.Warnf("MongoDB queries not yet implemented for table: %s", table)
				continue // Skip for now
			} else {
				tableResult, err = dataSource.PostgresRepository.Query(ctx, table, fields)
				if err != nil {
					logger.Errorf("Error querying table %s: %s", table, err.Error())

					return err
				}
			}

			// Set the query results for this table
			result[databaseName][table] = tableResult
		}
	}

	renderer := pongo.NewTemplateRenderer()

	out, err := renderer.RenderFromBytes(ctx, fileBytes, result)
	if err != nil {
		logger.Errorf("Error rendering template: %s", err.Error())

		return err
	}

	templateType := strings.ToLower(message.OutputFormat)
	contentType := getContentType(templateType)
	objectName := message.ReportID.String() + "/" + time.Now().Format("20060102_150405") + "." + templateType

	err = uc.ReportFileRepo.Put(ctx, objectName, contentType, []byte(out))
	if err != nil {
		logger.Errorf("Error putting report file: %s", err.Error())

		return err
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
