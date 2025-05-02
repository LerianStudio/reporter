package report

import (
	"github.com/google/uuid"
	"time"
)

// Report represents the entity model for a report
type Report struct {
	ID         uuid.UUID      `json:"id" example:"00000000-0000-0000-0000-000000000000"`
	TemplateID uuid.UUID      `json:"templateId" example:"00000000-0000-0000-0000-000000000000"`
	LedgerID   []uuid.UUID    `json:"ledgerId" example:"['00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000']"`
	Filters    map[string]any `json:"filters"`
	Status     string         `json:"status" example:"processing"`
}

// ReportMongoDBModel represents the MongoDB model for a report
type ReportMongoDBModel struct {
	ID          uuid.UUID      `bson:"_id"`
	TemplateID  uuid.UUID      `bson:"template_id"`
	Status      string         `bson:"status"`
	Metadata    map[string]any `bson:"metadata"`
	CompletedAt *time.Time     `bson:"completed_at"`
	CreatedAt   time.Time      `bson:"created_at"`
	UpdatedAt   time.Time      `bson:"updated_at"`
	DeletedAt   *time.Time     `bson:"deleted_at"`
}

// ToEntity converts ReportMongoDBModel to Report
func (rm *ReportMongoDBModel) ToEntity(ledgerIDs []uuid.UUID, filters map[string]any) *Report {
	return &Report{
		ID:         rm.ID,
		TemplateID: rm.TemplateID,
		Status:     rm.Status,
		LedgerID:   ledgerIDs,
		Filters:    filters,
	}
}

// FromEntity converts Report to ReportMongoDBModel
func (rm *ReportMongoDBModel) FromEntity(r *Report) error {
	dateNow := time.Now()
	rm.ID = r.ID
	rm.TemplateID = r.TemplateID
	rm.Metadata = nil
	rm.Status = r.Status
	rm.CompletedAt = nil
	rm.CreatedAt = dateNow
	rm.UpdatedAt = dateNow
	rm.DeletedAt = nil

	return nil
}
