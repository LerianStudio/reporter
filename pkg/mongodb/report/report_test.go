// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package report

import (
	"errors"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportMongoDBModel_ToEntity(t *testing.T) {
	now := time.Now()
	completedAt := now.Add(time.Hour)
	id := uuid.New()
	templateID := uuid.New()

	mongoModel := &ReportMongoDBModel{
		ID:          id,
		TemplateID:  templateID,
		Status:      "completed",
		Filters:     nil,
		Metadata:    map[string]any{"key": "value"},
		CompletedAt: &completedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
	}

	customFilters := map[string]map[string]map[string]model.FilterCondition{
		"table1": {
			"column1": {
				"filter1": {
					Equals: []any{"test"},
				},
			},
		},
	}

	entity := mongoModel.ToEntity(customFilters)

	assert.Equal(t, id, entity.ID)
	assert.Equal(t, templateID, entity.TemplateID)
	assert.Equal(t, "completed", entity.Status)
	assert.Equal(t, customFilters, entity.Filters)
	assert.Equal(t, &completedAt, entity.CompletedAt)
	assert.Equal(t, now, entity.CreatedAt)
	assert.Equal(t, now, entity.UpdatedAt)
	assert.Nil(t, entity.DeletedAt)
}

func TestReportMongoDBModel_ToEntity_NilFilters(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	templateID := uuid.New()

	mongoModel := &ReportMongoDBModel{
		ID:         id,
		TemplateID: templateID,
		Status:     "processing",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	entity := mongoModel.ToEntity(nil)

	assert.Equal(t, id, entity.ID)
	assert.Equal(t, templateID, entity.TemplateID)
	assert.Equal(t, "processing", entity.Status)
	assert.Nil(t, entity.Filters)
}

func TestReportMongoDBModel_ToEntityFindByID(t *testing.T) {
	now := time.Now()
	completedAt := now.Add(time.Hour)
	deletedAt := now.Add(2 * time.Hour)
	id := uuid.New()
	templateID := uuid.New()

	filters := map[string]map[string]map[string]model.FilterCondition{
		"users": {
			"status": {
				"active": {
					Equals: []any{true},
				},
			},
		},
	}

	metadata := map[string]any{
		"createdBy": "admin",
		"priority":  1,
	}

	mongoModel := &ReportMongoDBModel{
		ID:          id,
		TemplateID:  templateID,
		Status:      "completed",
		Filters:     filters,
		Metadata:    metadata,
		CompletedAt: &completedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   &deletedAt,
	}

	entity := mongoModel.ToEntityFindByID()

	assert.Equal(t, id, entity.ID)
	assert.Equal(t, templateID, entity.TemplateID)
	assert.Equal(t, "completed", entity.Status)
	assert.Equal(t, filters, entity.Filters)
	assert.Equal(t, metadata, entity.Metadata)
	assert.Equal(t, &completedAt, entity.CompletedAt)
	assert.Equal(t, now, entity.CreatedAt)
	assert.Equal(t, now, entity.UpdatedAt)
	assert.Equal(t, &deletedAt, entity.DeletedAt)
}

func TestReportMongoDBModel_ToEntityFindByID_EmptyFields(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	templateID := uuid.New()

	mongoModel := &ReportMongoDBModel{
		ID:         id,
		TemplateID: templateID,
		Status:     "pending",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	entity := mongoModel.ToEntityFindByID()

	assert.Equal(t, id, entity.ID)
	assert.Equal(t, templateID, entity.TemplateID)
	assert.Equal(t, "pending", entity.Status)
	assert.Nil(t, entity.Filters)
	assert.Nil(t, entity.Metadata)
	assert.Nil(t, entity.CompletedAt)
	assert.Nil(t, entity.DeletedAt)
}

func TestReportMongoDBModel_FromEntity(t *testing.T) {
	id := uuid.New()
	templateID := uuid.New()
	completedAt := time.Now()

	filters := map[string]map[string]map[string]model.FilterCondition{
		"orders": {
			"amount": {
				"gt": {
					GreaterThan: []any{1000},
				},
			},
		},
	}

	report := &Report{
		ID:          id,
		TemplateID:  templateID,
		Filters:     filters,
		Status:      "completed",
		Metadata:    map[string]any{"key": "value"},
		CompletedAt: &completedAt,
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now(),
	}

	mongoModel := &ReportMongoDBModel{}
	err := mongoModel.FromEntity(report)

	assert.NoError(t, err)
	assert.Equal(t, id, mongoModel.ID)
	assert.Equal(t, templateID, mongoModel.TemplateID)
	assert.Equal(t, "completed", mongoModel.Status)
	assert.Equal(t, filters, mongoModel.Filters)
	assert.Equal(t, report.Metadata, mongoModel.Metadata)       // FromEntity preserves metadata
	assert.Equal(t, report.CompletedAt, mongoModel.CompletedAt) // FromEntity preserves completedAt
	assert.Nil(t, mongoModel.DeletedAt)
	assert.False(t, mongoModel.CreatedAt.IsZero())
	assert.False(t, mongoModel.UpdatedAt.IsZero())
}

func TestReportMongoDBModel_FromEntity_EmptyReport(t *testing.T) {
	report := &Report{
		ID:         uuid.New(),
		TemplateID: uuid.New(),
		Status:     "pending",
	}

	mongoModel := &ReportMongoDBModel{}
	err := mongoModel.FromEntity(report)

	assert.NoError(t, err)
	assert.Equal(t, report.ID, mongoModel.ID)
	assert.Equal(t, report.TemplateID, mongoModel.TemplateID)
	assert.Equal(t, "pending", mongoModel.Status)
	assert.Nil(t, mongoModel.Filters)
}

func TestReport_Struct(t *testing.T) {
	now := time.Now()
	completedAt := now.Add(time.Hour)
	deletedAt := now.Add(2 * time.Hour)
	id := uuid.New()
	templateID := uuid.New()

	filters := map[string]map[string]map[string]model.FilterCondition{
		"table": {
			"column": {
				"condition": {
					Equals: []any{"test"},
				},
			},
		},
	}

	metadata := map[string]any{
		"key1": "value1",
		"key2": 123,
	}

	report := Report{
		ID:          id,
		TemplateID:  templateID,
		Filters:     filters,
		Status:      "completed",
		Metadata:    metadata,
		CompletedAt: &completedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   &deletedAt,
	}

	assert.Equal(t, id, report.ID)
	assert.Equal(t, templateID, report.TemplateID)
	assert.Equal(t, filters, report.Filters)
	assert.Equal(t, "completed", report.Status)
	assert.Equal(t, metadata, report.Metadata)
	assert.Equal(t, &completedAt, report.CompletedAt)
	assert.Equal(t, now, report.CreatedAt)
	assert.Equal(t, now, report.UpdatedAt)
	assert.Equal(t, &deletedAt, report.DeletedAt)
}

func TestReportMongoDBModel_Struct(t *testing.T) {
	now := time.Now()
	completedAt := now.Add(time.Hour)
	deletedAt := now.Add(2 * time.Hour)
	id := uuid.New()
	templateID := uuid.New()

	filters := map[string]map[string]map[string]model.FilterCondition{
		"table": {
			"column": {
				"condition": {
					In: []any{"a", "b", "c"},
				},
			},
		},
	}

	metadata := map[string]any{
		"source": "api",
	}

	mongoModel := ReportMongoDBModel{
		ID:          id,
		TemplateID:  templateID,
		Status:      "processing",
		Filters:     filters,
		Metadata:    metadata,
		CompletedAt: &completedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   &deletedAt,
	}

	assert.Equal(t, id, mongoModel.ID)
	assert.Equal(t, templateID, mongoModel.TemplateID)
	assert.Equal(t, "processing", mongoModel.Status)
	assert.Equal(t, filters, mongoModel.Filters)
	assert.Equal(t, metadata, mongoModel.Metadata)
	assert.Equal(t, &completedAt, mongoModel.CompletedAt)
	assert.Equal(t, now, mongoModel.CreatedAt)
	assert.Equal(t, now, mongoModel.UpdatedAt)
	assert.Equal(t, &deletedAt, mongoModel.DeletedAt)
}

func TestReportStatuses(t *testing.T) {
	statuses := []string{"pending", "processing", "completed", "failed"}

	for _, status := range statuses {
		t.Run("Status_"+status, func(t *testing.T) {
			report := Report{
				ID:         uuid.New(),
				TemplateID: uuid.New(),
				Status:     status,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			assert.Equal(t, status, report.Status)
		})
	}
}

func TestReportMongoDBModel_ToEntity_WithComplexFilters(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	templateID := uuid.New()

	mongoModel := &ReportMongoDBModel{
		ID:         id,
		TemplateID: templateID,
		Status:     "completed",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	complexFilters := map[string]map[string]map[string]model.FilterCondition{
		"transactions": {
			"amount": {
				"range": {
					Between: []any{100.0, 500.0},
				},
			},
			"status": {
				"list": {
					In: []any{"pending", "approved"},
				},
			},
		},
		"users": {
			"created_at": {
				"after": {
					GreaterOrEqual: []any{"2024-01-01"},
				},
			},
		},
	}

	entity := mongoModel.ToEntity(complexFilters)

	assert.Equal(t, complexFilters, entity.Filters)
	assert.Len(t, entity.Filters, 2)
	assert.Len(t, entity.Filters["transactions"], 2)
	assert.Len(t, entity.Filters["users"], 1)
}

func TestNewReport(t *testing.T) {
	t.Parallel()

	validTemplateID := uuid.New()
	validFilters := map[string]map[string]map[string]model.FilterCondition{
		"transactions": {
			"amount": {
				"range": {Equals: []any{"100"}},
			},
		},
	}

	tests := []struct {
		name        string
		id          uuid.UUID
		templateID  uuid.UUID
		status      string
		filters     map[string]map[string]map[string]model.FilterCondition
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "valid report with all fields",
			id:         uuid.New(),
			templateID: validTemplateID,
			status:     constant.ProcessingStatus,
			filters:    validFilters,
			wantErr:    false,
		},
		{
			name:       "valid report with nil filters",
			id:         uuid.New(),
			templateID: validTemplateID,
			status:     constant.ProcessingStatus,
			filters:    nil,
			wantErr:    false,
		},
		{
			name:        "nil ID returns error",
			id:          uuid.Nil,
			templateID:  validTemplateID,
			status:      constant.ProcessingStatus,
			filters:     validFilters,
			wantErr:     true,
			expectedErr: constant.ErrMissingRequiredFields,
		},
		{
			name:        "nil templateID returns error",
			id:          uuid.New(),
			templateID:  uuid.Nil,
			status:      constant.ProcessingStatus,
			filters:     validFilters,
			wantErr:     true,
			expectedErr: constant.ErrMissingRequiredFields,
		},
		{
			name:        "empty status returns error",
			id:          uuid.New(),
			templateID:  validTemplateID,
			status:      "",
			filters:     validFilters,
			wantErr:     true,
			expectedErr: constant.ErrMissingRequiredFields,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewReport(tt.id, tt.templateID, tt.status, tt.filters)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)

				if tt.expectedErr != nil {
					assert.True(t, errors.Is(err, tt.expectedErr), "expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.id, got.ID)
				assert.Equal(t, tt.templateID, got.TemplateID)
				assert.Equal(t, tt.status, got.Status)
				assert.Equal(t, tt.filters, got.Filters)
			}
		})
	}
}

func TestFilterCondition_AllOperators(t *testing.T) {
	tests := []struct {
		name   string
		filter model.FilterCondition
	}{
		{
			name:   "Equals operator",
			filter: model.FilterCondition{Equals: []any{"value1", "value2"}},
		},
		{
			name:   "GreaterThan operator",
			filter: model.FilterCondition{GreaterThan: []any{100}},
		},
		{
			name:   "GreaterOrEqual operator",
			filter: model.FilterCondition{GreaterOrEqual: []any{"2025-01-01"}},
		},
		{
			name:   "LessThan operator",
			filter: model.FilterCondition{LessThan: []any{1000}},
		},
		{
			name:   "LessOrEqual operator",
			filter: model.FilterCondition{LessOrEqual: []any{"2025-12-31"}},
		},
		{
			name:   "Between operator",
			filter: model.FilterCondition{Between: []any{100, 500}},
		},
		{
			name:   "In operator",
			filter: model.FilterCondition{In: []any{"a", "b", "c"}},
		},
		{
			name:   "NotIn operator",
			filter: model.FilterCondition{NotIn: []any{"deleted", "archived"}},
		},
		{
			name: "Combined operators",
			filter: model.FilterCondition{
				GreaterOrEqual: []any{100},
				LessThan:       []any{1000},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := map[string]map[string]map[string]model.FilterCondition{
				"table": {
					"column": {
						"test": tt.filter,
					},
				},
			}

			mongoModel := &ReportMongoDBModel{
				ID:         uuid.New(),
				TemplateID: uuid.New(),
				Status:     "pending",
				Filters:    filters,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			entity := mongoModel.ToEntityFindByID()
			assert.Equal(t, filters, entity.Filters)
		})
	}
}
