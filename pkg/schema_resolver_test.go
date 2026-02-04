package pkg

import (
	"testing"

	"github.com/LerianStudio/reporter/v4/pkg/postgres"
)

func TestSchemaResolver_ResolveSchema(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(r *SchemaResolver)
		database       string
		explicitSchema string
		table          string
		wantSchema     string
		wantErr        bool
		wantErrType    string
	}{
		{
			name: "explicit schema - valid",
			setup: func(r *SchemaResolver) {
				r.RegisterDatabase("pix_btg", []postgres.TableSchema{
					{SchemaName: "payment", TableName: "transactions"},
					{SchemaName: "transfer", TableName: "orders"},
				})
			},
			database:       "pix_btg",
			explicitSchema: "payment",
			table:          "transactions",
			wantSchema:     "payment",
			wantErr:        false,
		},
		{
			name: "explicit schema - table not in schema",
			setup: func(r *SchemaResolver) {
				r.RegisterDatabase("pix_btg", []postgres.TableSchema{
					{SchemaName: "payment", TableName: "transactions"},
					{SchemaName: "transfer", TableName: "orders"},
				})
			},
			database:       "pix_btg",
			explicitSchema: "payment",
			table:          "orders",
			wantErr:        true,
			wantErrType:    "not found",
		},
		{
			name: "implicit schema - single match",
			setup: func(r *SchemaResolver) {
				r.RegisterDatabase("pix_btg", []postgres.TableSchema{
					{SchemaName: "payment", TableName: "transactions"},
					{SchemaName: "transfer", TableName: "orders"},
				})
			},
			database:       "pix_btg",
			explicitSchema: "",
			table:          "transactions",
			wantSchema:     "payment",
			wantErr:        false,
		},
		{
			name: "implicit schema - multiple matches with public",
			setup: func(r *SchemaResolver) {
				r.RegisterDatabase("midaz", []postgres.TableSchema{
					{SchemaName: "public", TableName: "users"},
					{SchemaName: "audit", TableName: "users"},
				})
			},
			database:       "midaz",
			explicitSchema: "",
			table:          "users",
			wantSchema:     "public",
			wantErr:        false,
		},
		{
			name: "implicit schema - multiple matches without public",
			setup: func(r *SchemaResolver) {
				r.RegisterDatabase("pix_btg", []postgres.TableSchema{
					{SchemaName: "payment", TableName: "transactions"},
					{SchemaName: "transfer", TableName: "transactions"},
				})
			},
			database:       "pix_btg",
			explicitSchema: "",
			table:          "transactions",
			wantErr:        true,
			wantErrType:    "ambiguous",
		},
		{
			name: "table not found in any schema",
			setup: func(r *SchemaResolver) {
				r.RegisterDatabase("pix_btg", []postgres.TableSchema{
					{SchemaName: "payment", TableName: "transactions"},
				})
			},
			database:       "pix_btg",
			explicitSchema: "",
			table:          "nonexistent",
			wantErr:        true,
			wantErrType:    "not found",
		},
		{
			name: "database not registered",
			setup: func(r *SchemaResolver) {
				// No database registered
			},
			database:       "unknown_db",
			explicitSchema: "",
			table:          "users",
			wantErr:        true,
			wantErrType:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewSchemaResolver()
			tt.setup(r)

			gotSchema, err := r.ResolveSchema(tt.database, tt.explicitSchema, tt.table)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveSchema() expected error, got nil")
					return
				}
				// Check error type if specified
				if tt.wantErrType != "" {
					errStr := err.Error()
					if tt.wantErrType == "ambiguous" {
						if _, ok := err.(*SchemaAmbiguityError); !ok {
							t.Errorf("ResolveSchema() expected SchemaAmbiguityError, got %T: %v", err, err)
						}
					} else if tt.wantErrType == "not found" {
						// Just check it's an error, not ambiguity
						if _, ok := err.(*SchemaAmbiguityError); ok {
							t.Errorf("ResolveSchema() expected not found error, got ambiguity: %v", errStr)
						}
					}
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveSchema() unexpected error: %v", err)
				return
			}

			if gotSchema != tt.wantSchema {
				t.Errorf("ResolveSchema() = %q, want %q", gotSchema, tt.wantSchema)
			}
		})
	}
}

func TestSchemaAmbiguityError_Error(t *testing.T) {
	err := &SchemaAmbiguityError{
		Database: "pix_btg",
		Table:    "transactions",
		Schemas:  []string{"payment", "transfer"},
	}

	errMsg := err.Error()

	// Should contain useful information
	if errMsg == "" {
		t.Error("Error() returned empty string")
	}

	// Should mention the table
	if !contains(errMsg, "transactions") {
		t.Errorf("Error() should mention table name, got: %s", errMsg)
	}

	// Should mention the schemas
	if !contains(errMsg, "payment") || !contains(errMsg, "transfer") {
		t.Errorf("Error() should mention available schemas, got: %s", errMsg)
	}

	// Should suggest explicit syntax
	if !contains(errMsg, ":") {
		t.Errorf("Error() should suggest explicit syntax with ':', got: %s", errMsg)
	}
}

func TestSchemaResolver_RegisterDatabase(t *testing.T) {
	r := NewSchemaResolver()

	tables := []postgres.TableSchema{
		{SchemaName: "public", TableName: "users"},
		{SchemaName: "public", TableName: "accounts"},
	}

	r.RegisterDatabase("midaz", tables)

	// Should be able to resolve tables after registration
	schema, err := r.ResolveSchema("midaz", "", "users")
	if err != nil {
		t.Errorf("ResolveSchema() after RegisterDatabase failed: %v", err)
	}
	if schema != "public" {
		t.Errorf("ResolveSchema() = %q, want %q", schema, "public")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
