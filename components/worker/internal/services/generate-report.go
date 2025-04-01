package services

import (
	"context"
	"encoding/json"
	libCommons "github.com/LerianStudio/lib-commons/commons"
	"github.com/google/uuid"
	"strings"
	"time"
)

// GenerateReportMessage message structure for report generation.
type GenerateReportMessage struct {
	ID           uuid.UUID        `json:"id"`
	Type         string           `json:"type"`
	FileURL      string           `json:"fileUrl"`
	MappedFields []map[string]any `json:"mappedFields"`
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

	fileBytes, err := uc.TemplateFileRepo.Get(ctx, message.ID.String())
	if err != nil {
		logger.Errorf("Error getting file from template bucket: %s", err.Error())

		return err
	}

	logger.Infof("Template found: %s", string(fileBytes))

	// TODO: midaz querying
	// TODO: pongo2 template processing here

	templateType := strings.ToLower(message.Type)
	contentType := getContentType(templateType)
	objectName := message.ID.String() + "/" + time.Now().Format("20060102_150405") + "." + templateType

	err = uc.ReportFileRepo.Put(ctx, objectName, contentType, fileBytes)
	if err != nil {
		logger.Errorf("Error putting report file: %s", err.Error())

		return err
	}

	return nil
}

// getContentType returns the MIME type for a given file extension.
func getContentType(ext string) string {
	switch ext {
	case "txt":
		return "text/plain"
	case "html":
		return "text/html"
	case "json":
		return "application/json"
	case "csv":
		return "text/csv"
	default:
		return "application/octet-stream" // fallback genérico
	}
}
