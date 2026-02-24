// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package postgres

import (
	"testing"
)

func TestTableSchema_QualifiedName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		schema     TableSchema
		wantResult string
	}{
		{
			name: "returns qualified name with schema and table",
			schema: TableSchema{
				SchemaName: "sales",
				TableName:  "orders",
			},
			wantResult: "sales.orders",
		},
		{
			name: "returns qualified name for public schema",
			schema: TableSchema{
				SchemaName: "public",
				TableName:  "users",
			},
			wantResult: "public.users",
		},
		{
			name: "handles empty schema name",
			schema: TableSchema{
				SchemaName: "",
				TableName:  "accounts",
			},
			wantResult: ".accounts",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.schema.QualifiedName()
			if got != tt.wantResult {
				t.Errorf("QualifiedName() = %q, want %q", got, tt.wantResult)
			}
		})
	}
}

func TestTableSchema_SchemaNameField(t *testing.T) {
	t.Parallel()

	// Test that SchemaName field exists and can be set
	ts := TableSchema{
		SchemaName: "sales",
		TableName:  "orders",
		Columns:    []ColumnInformation{},
	}

	if ts.SchemaName != "sales" {
		t.Errorf("SchemaName = %q, want %q", ts.SchemaName, "sales")
	}

	if ts.TableName != "orders" {
		t.Errorf("TableName = %q, want %q", ts.TableName, "orders")
	}
}

func TestQualifyTableName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		schemaName string
		tableName  string
		want       string
	}{
		{
			name:       "with schema name",
			schemaName: "sales",
			tableName:  "orders",
			want:       `"sales"."orders"`,
		},
		{
			name:       "with public schema",
			schemaName: "public",
			tableName:  "users",
			want:       `"public"."users"`,
		},
		{
			name:       "without schema name - empty string",
			schemaName: "",
			tableName:  "accounts",
			want:       "accounts",
		},
		{
			name:       "handles special characters in names",
			schemaName: "my_schema",
			tableName:  "my_table",
			want:       `"my_schema"."my_table"`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := qualifyTableName(tt.schemaName, tt.tableName)
			if got != tt.want {
				t.Errorf("qualifyTableName(%q, %q) = %q, want %q", tt.schemaName, tt.tableName, got, tt.want)
			}
		})
	}
}
