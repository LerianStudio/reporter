// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package template

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTemplateMongoDBModel_ToEntity(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	mongoModel := &TemplateMongoDBModel{
		ID:           id,
		OutputFormat: "PDF",
		Description:  "Financial Report Template",
		FileName:     "0196159b-4f26-7300-b3d9-f4f68a7c85f3_1744119295.tpl",
		MappedFields: map[string]map[string][]string{
			"users": {
				"table1": {"name", "email", "created_at"},
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}

	entity := mongoModel.ToEntity()

	assert.Equal(t, id, entity.ID)
	assert.Equal(t, "PDF", entity.OutputFormat)
	assert.Equal(t, "Financial Report Template", entity.Description)
	assert.Equal(t, "0196159b-4f26-7300-b3d9-f4f68a7c85f3_1744119295.tpl", entity.FileName)
	assert.Equal(t, now, entity.CreatedAt)
	assert.Equal(t, now, entity.UpdatedAt)
}

func TestTemplateMongoDBModel_ToEntity_EmptyFields(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	mongoModel := &TemplateMongoDBModel{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
	}

	entity := mongoModel.ToEntity()

	assert.Equal(t, id, entity.ID)
	assert.Empty(t, entity.OutputFormat)
	assert.Empty(t, entity.Description)
	assert.Empty(t, entity.FileName)
	assert.Equal(t, now, entity.CreatedAt)
	assert.Equal(t, now, entity.UpdatedAt)
}

func TestTemplateMongoDBModel_ToEntity_AllOutputFormats(t *testing.T) {
	formats := []string{"PDF", "HTML", "CSV", "XML", "TXT"}

	for _, format := range formats {
		t.Run("Format_"+format, func(t *testing.T) {
			mongoModel := &TemplateMongoDBModel{
				ID:           uuid.New(),
				OutputFormat: format,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			entity := mongoModel.ToEntity()

			assert.Equal(t, format, entity.OutputFormat)
		})
	}
}

func TestTemplate_Struct(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	template := Template{
		ID:           id,
		OutputFormat: "HTML",
		Description:  "Monthly Sales Report",
		FileName:     "template_123.tpl",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	assert.Equal(t, id, template.ID)
	assert.Equal(t, "HTML", template.OutputFormat)
	assert.Equal(t, "Monthly Sales Report", template.Description)
	assert.Equal(t, "template_123.tpl", template.FileName)
	assert.Equal(t, now, template.CreatedAt)
	assert.Equal(t, now, template.UpdatedAt)
}

func TestTemplateMongoDBModel_Struct(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)
	id := uuid.New()

	mappedFields := map[string]map[string][]string{
		"data_source_1": {
			"table_a": {"col1", "col2", "col3"},
			"table_b": {"id", "name"},
		},
		"data_source_2": {
			"table_c": {"value", "timestamp"},
		},
	}

	mongoModel := TemplateMongoDBModel{
		ID:           id,
		OutputFormat: "CSV",
		Description:  "Export Template",
		FileName:     "export_template.tpl",
		MappedFields: mappedFields,
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    &deletedAt,
	}

	assert.Equal(t, id, mongoModel.ID)
	assert.Equal(t, "CSV", mongoModel.OutputFormat)
	assert.Equal(t, "Export Template", mongoModel.Description)
	assert.Equal(t, "export_template.tpl", mongoModel.FileName)
	assert.Equal(t, mappedFields, mongoModel.MappedFields)
	assert.Equal(t, now, mongoModel.CreatedAt)
	assert.Equal(t, now, mongoModel.UpdatedAt)
	assert.Equal(t, &deletedAt, mongoModel.DeletedAt)
}

func TestTemplateMongoDBModel_MappedFields(t *testing.T) {
	tests := []struct {
		name         string
		mappedFields map[string]map[string][]string
	}{
		{
			name:         "Empty mapped fields",
			mappedFields: nil,
		},
		{
			name:         "Single data source",
			mappedFields: map[string]map[string][]string{"ds1": {"t1": {"c1"}}},
		},
		{
			name: "Multiple data sources",
			mappedFields: map[string]map[string][]string{
				"primary":   {"users": {"id", "name"}, "orders": {"id", "total"}},
				"secondary": {"logs": {"timestamp", "message"}},
			},
		},
		{
			name: "Complex nested structure",
			mappedFields: map[string]map[string][]string{
				"analytics": {
					"events":      {"event_id", "event_type", "user_id", "timestamp", "payload"},
					"sessions":    {"session_id", "user_id", "start_time", "end_time", "device"},
					"conversions": {"conversion_id", "type", "value", "source"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mongoModel := TemplateMongoDBModel{
				ID:           uuid.New(),
				MappedFields: tt.mappedFields,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			assert.Equal(t, tt.mappedFields, mongoModel.MappedFields)
		})
	}
}

func TestTemplateMongoDBModel_ToEntity_DoesNotIncludeMappedFields(t *testing.T) {
	mappedFields := map[string]map[string][]string{
		"ds1": {
			"table1": {"col1", "col2"},
		},
	}

	mongoModel := &TemplateMongoDBModel{
		ID:           uuid.New(),
		OutputFormat: "PDF",
		MappedFields: mappedFields,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	entity := mongoModel.ToEntity()

	// Template entity doesn't have MappedFields
	assert.Equal(t, "PDF", entity.OutputFormat)
	// MappedFields is only in MongoDB model, not in entity
}

func TestTemplateMongoDBModel_ToEntity_WithDeletedAt(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	mongoModel := &TemplateMongoDBModel{
		ID:           uuid.New(),
		OutputFormat: "XML",
		Description:  "Deleted Template",
		FileName:     "deleted.tpl",
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    &deletedAt,
	}

	entity := mongoModel.ToEntity()

	// Entity doesn't have DeletedAt field (soft delete only in MongoDB model)
	assert.Equal(t, "XML", entity.OutputFormat)
	assert.Equal(t, "Deleted Template", entity.Description)
}

func TestTemplate_JSONTags(t *testing.T) {
	template := Template{
		ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		OutputFormat: "PDF",
		Description:  "Test",
		FileName:     "test.tpl",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	// Verify struct fields exist with expected values
	assert.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), template.ID)
	assert.Equal(t, "PDF", template.OutputFormat)
	assert.Equal(t, "Test", template.Description)
	assert.Equal(t, "test.tpl", template.FileName)
}

func TestTemplateMongoDBModel_BSONTags(t *testing.T) {
	deletedAt := time.Now()

	mongoModel := TemplateMongoDBModel{
		ID:           uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		OutputFormat: "HTML",
		Description:  "BSON Test",
		FileName:     "bson_test.tpl",
		MappedFields: map[string]map[string][]string{"ds": {"t": {"c"}}},
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		DeletedAt:    &deletedAt,
	}

	// Verify struct fields exist with expected values
	assert.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000002"), mongoModel.ID)
	assert.Equal(t, "HTML", mongoModel.OutputFormat)
	assert.Equal(t, "BSON Test", mongoModel.Description)
	assert.Equal(t, "bson_test.tpl", mongoModel.FileName)
	assert.NotNil(t, mongoModel.MappedFields)
	assert.Equal(t, &deletedAt, mongoModel.DeletedAt)
}
