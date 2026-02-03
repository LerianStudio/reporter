// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package report

import (
	"time"

	"github.com/LerianStudio/reporter/v4/pkg/model"

	"github.com/google/uuid"
)

// Report represents the entity model for a report
type Report struct {
	ID          uuid.UUID                                              `json:"id" example:"00000000-0000-0000-0000-000000000000"`
	TemplateID  uuid.UUID                                              `json:"templateId" example:"00000000-0000-0000-0000-000000000000"`
	Filters     map[string]map[string]map[string]model.FilterCondition `json:"filters"`
	Status      string                                                 `json:"status" example:"processing"`
	Metadata    map[string]any                                         `json:"metadata"`
	CompletedAt *time.Time                                             `json:"completedAt"`
	CreatedAt   time.Time                                              `json:"createdAt"`
	UpdatedAt   time.Time                                              `json:"updatedAt"`
	DeletedAt   *time.Time                                             `json:"deletedAt"`
}

// ReportMongoDBModel represents the MongoDB model for a report
type ReportMongoDBModel struct {
	ID             uuid.UUID                                              `bson:"_id"`
	TemplateID     uuid.UUID                                              `bson:"template_id"`
	OrganizationID uuid.UUID                                              `bson:"organization_id"`
	Status         string                                                 `bson:"status"`
	Filters        map[string]map[string]map[string]model.FilterCondition `bson:"filters"`
	Metadata       map[string]any                                         `bson:"metadata"`
	CompletedAt    *time.Time                                             `bson:"completed_at"`
	CreatedAt      time.Time                                              `bson:"created_at"`
	UpdatedAt      time.Time                                              `bson:"updated_at"`
	DeletedAt      *time.Time                                             `bson:"deleted_at"`
}

// ToEntity converts ReportMongoDBModel to Report
func (rm *ReportMongoDBModel) ToEntity(filters map[string]map[string]map[string]model.FilterCondition) *Report {
	return &Report{
		ID:          rm.ID,
		TemplateID:  rm.TemplateID,
		Status:      rm.Status,
		Filters:     filters,
		CompletedAt: rm.CompletedAt,
		CreatedAt:   rm.CreatedAt,
		UpdatedAt:   rm.UpdatedAt,
		DeletedAt:   rm.DeletedAt,
	}
}

// ToEntityFindByID converts ReportMongoDBModel to Report
func (rm *ReportMongoDBModel) ToEntityFindByID() *Report {
	return &Report{
		ID:          rm.ID,
		TemplateID:  rm.TemplateID,
		Status:      rm.Status,
		Filters:     rm.Filters,
		Metadata:    rm.Metadata,
		CompletedAt: rm.CompletedAt,
		CreatedAt:   rm.CreatedAt,
		UpdatedAt:   rm.UpdatedAt,
		DeletedAt:   rm.DeletedAt,
	}
}

// FromEntity converts Report to ReportMongoDBModel
func (rm *ReportMongoDBModel) FromEntity(r *Report, organizationID uuid.UUID) error {
	dateNow := time.Now()
	rm.ID = r.ID
	rm.TemplateID = r.TemplateID
	rm.OrganizationID = organizationID
	rm.Metadata = nil
	rm.Status = r.Status
	rm.Filters = r.Filters
	rm.CompletedAt = nil
	rm.CreatedAt = dateNow
	rm.UpdatedAt = dateNow
	rm.DeletedAt = nil

	return nil
}
