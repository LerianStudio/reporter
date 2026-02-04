package postgres

import (
	"testing"
)

func TestTableSchema_QualifiedName(t *testing.T) {
	tests := []struct {
		name       string
		schema     TableSchema
		wantResult string
	}{
		{
			name: "returns qualified name with schema and table",
			schema: TableSchema{
				SchemaName: "payment",
				TableName:  "transactions",
			},
			wantResult: "payment.transactions",
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
		t.Run(tt.name, func(t *testing.T) {
			got := tt.schema.QualifiedName()
			if got != tt.wantResult {
				t.Errorf("QualifiedName() = %q, want %q", got, tt.wantResult)
			}
		})
	}
}

func TestTableSchema_SchemaNameField(t *testing.T) {
	// Test that SchemaName field exists and can be set
	ts := TableSchema{
		SchemaName: "payment",
		TableName:  "transactions",
		Columns:    []ColumnInformation{},
	}

	if ts.SchemaName != "payment" {
		t.Errorf("SchemaName = %q, want %q", ts.SchemaName, "payment")
	}

	if ts.TableName != "transactions" {
		t.Errorf("TableName = %q, want %q", ts.TableName, "transactions")
	}
}

func TestQualifyTableName(t *testing.T) {
	tests := []struct {
		name       string
		schemaName string
		tableName  string
		want       string
	}{
		{
			name:       "with schema name",
			schemaName: "payment",
			tableName:  "transactions",
			want:       `"payment"."transactions"`,
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
		t.Run(tt.name, func(t *testing.T) {
			got := qualifyTableName(tt.schemaName, tt.tableName)
			if got != tt.want {
				t.Errorf("qualifyTableName(%q, %q) = %q, want %q", tt.schemaName, tt.tableName, got, tt.want)
			}
		})
	}
}