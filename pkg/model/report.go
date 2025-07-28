package model

import "github.com/google/uuid"

// CreateReportInput is a struct designed to encapsulate request create payload data.
//
// swagger:model CreateReportInput
// @Description CreateReportInput is the input payload to create a report.
type CreateReportInput struct {
	TemplateID string                                    `json:"templateId" validate:"required" example:"00000000-0000-0000-0000-000000000000"`
	Filters    map[string]map[string]map[string][]string `json:"filters" validate:"required"`
} // @name CreateReportInput

// ReportMessage is a struct designed to encapsulate response payload data.
//
// swagger:model ReportMessage
//
// @Description ReportMessage represents a report struct of message sent it in RabbitMQ
type ReportMessage struct {
	TemplateID   uuid.UUID                                 `json:"templateId" example:"00000000-0000-0000-0000-000000000000"`
	ReportID     uuid.UUID                                 `json:"reportId" example:"00000000-0000-0000-0000-000000000000"`
	OutputFormat string                                    `json:"outputFormat" example:"html"`
	Filters      map[string]map[string]map[string][]string `json:"filters"`
	MappedFields map[string]map[string][]string            `json:"mappedFields"`
} // @name ReportMessage
