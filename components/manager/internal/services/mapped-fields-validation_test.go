// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"errors"
	"testing"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/mongodb"
	"github.com/LerianStudio/reporter/pkg/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUseCase_ValidateIfFieldsExistOnTables_PostgreSQL(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because ResetRegisteredDataSourceIDsForTesting mutates global state
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostgresRepo := postgres.NewMockRepository(ctrl)

	// Register datasource IDs for testing
	pkg.ResetRegisteredDataSourceIDsForTesting()
	pkg.RegisterDataSourceIDsForTesting([]string{"test_postgres_db"})

	postgresSchema := []postgres.TableSchema{
		{
			SchemaName: "public",
			TableName:  "users",
			Columns: []postgres.ColumnInformation{
				{Name: "id", DataType: "uuid"},
				{Name: "name", DataType: "varchar"},
				{Name: "email", DataType: "varchar"},
			},
		},
		{
			SchemaName: "public",
			TableName:  "accounts",
			Columns: []postgres.ColumnInformation{
				{Name: "id", DataType: "uuid"},
				{Name: "balance", DataType: "numeric"},
			},
		},
	}

	tests := []struct {
		name         string
		mappedFields map[string]map[string][]string
		mockSetup    func()
		expectErr    bool
		errContains  string
	}{
		{
			name: "Success - All fields exist",
			mappedFields: map[string]map[string][]string{
				"test_postgres_db": {
					"users": {"id", "name", "email"},
				},
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().
					GetDatabaseSchema(gomock.Any(), []string{"public"}).
					Return(postgresSchema, nil)
				mockPostgresRepo.EXPECT().
					CloseConnection().
					Return(nil)
			},
			expectErr: false,
		},
		{
			name: "Error - Missing fields in table",
			mappedFields: map[string]map[string][]string{
				"test_postgres_db": {
					"users": {"id", "name", "nonexistent_field"},
				},
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().
					GetDatabaseSchema(gomock.Any(), []string{"public"}).
					Return(postgresSchema, nil)
				// CloseConnection is NOT called because validation fails before reaching it
			},
			expectErr:   true,
			errContains: "nonexistent_field",
		},
		{
			name: "Error - Table does not exist",
			mappedFields: map[string]map[string][]string{
				"test_postgres_db": {
					"nonexistent_table": {"id", "name"},
				},
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().
					GetDatabaseSchema(gomock.Any(), []string{"public"}).
					Return(postgresSchema, nil)
				// CloseConnection is NOT called because table doesn't exist
			},
			expectErr:   true,
			errContains: "nonexistent_table",
		},
		{
			name: "Error - Database schema retrieval fails",
			mappedFields: map[string]map[string][]string{
				"test_postgres_db": {
					"users": {"id", "name"},
				},
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().
					GetDatabaseSchema(gomock.Any(), []string{"public"}).
					Return(nil, errors.New("database connection error"))
				// CloseConnection is NOT called because schema retrieval fails
			},
			expectErr:   true,
			errContains: "database connection error",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &UseCase{
				ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
					"test_postgres_db": {
						DatabaseType:       pkg.PostgreSQLType,
						PostgresRepository: mockPostgresRepo,
						Initialized:        true,
						Schemas:            []string{"public"},
						DatabaseConfig: &postgres.Connection{
							Connected: true,
						},
					},
				}),
			}

			ctx := context.Background()
			err := svc.ValidateIfFieldsExistOnTables(ctx, tt.mappedFields)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUseCase_ValidateIfFieldsExistOnTables_MongoDB(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because ResetRegisteredDataSourceIDsForTesting mutates global state
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMongoRepo := mongodb.NewMockRepository(ctrl)

	// Register datasource IDs for testing
	pkg.ResetRegisteredDataSourceIDsForTesting()
	pkg.RegisterDataSourceIDsForTesting([]string{"test_mongo_db"})

	mongoSchema := []mongodb.CollectionSchema{
		{
			CollectionName: "transactions",
			Fields: []mongodb.FieldInformation{
				{Name: "_id", DataType: "objectId"},
				{Name: "amount", DataType: "number"},
				{Name: "status", DataType: "string"},
			},
		},
		{
			CollectionName: "accounts",
			Fields: []mongodb.FieldInformation{
				{Name: "_id", DataType: "objectId"},
				{Name: "balance", DataType: "number"},
			},
		},
	}

	tests := []struct {
		name         string
		mappedFields map[string]map[string][]string
		mockSetup    func()
		expectErr    bool
		errContains  string
	}{
		{
			name: "Success - All fields exist",
			mappedFields: map[string]map[string][]string{
				"test_mongo_db": {
					"transactions": {"_id", "amount", "status"},
				},
			},
			mockSetup: func() {
				mockMongoRepo.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(mongoSchema, nil)
				mockMongoRepo.EXPECT().
					CloseConnection(gomock.Any()).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name: "Error - Missing fields in collection",
			mappedFields: map[string]map[string][]string{
				"test_mongo_db": {
					"transactions": {"_id", "amount", "nonexistent_field"},
				},
			},
			mockSetup: func() {
				mockMongoRepo.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(mongoSchema, nil)
				// CloseConnection is NOT called because validation fails
			},
			expectErr:   true,
			errContains: "nonexistent_field",
		},
		{
			name: "Error - Collection does not exist",
			mappedFields: map[string]map[string][]string{
				"test_mongo_db": {
					"nonexistent_collection": {"_id"},
				},
			},
			mockSetup: func() {
				mockMongoRepo.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(mongoSchema, nil)
				// CloseConnection is NOT called because collection doesn't exist
			},
			expectErr:   true,
			errContains: "nonexistent_collection",
		},
		{
			name: "Error - Database schema retrieval fails",
			mappedFields: map[string]map[string][]string{
				"test_mongo_db": {
					"transactions": {"_id"},
				},
			},
			mockSetup: func() {
				mockMongoRepo.EXPECT().
					GetDatabaseSchema(gomock.Any()).
					Return(nil, errors.New("mongodb connection error"))
				// CloseConnection is NOT called because schema retrieval fails
			},
			expectErr:   true,
			errContains: "mongodb connection error",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &UseCase{
				ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
					"test_mongo_db": {
						DatabaseType:      pkg.MongoDBType,
						MongoDBRepository: mockMongoRepo,
						Initialized:       true,
					},
				}),
			}

			ctx := context.Background()
			err := svc.ValidateIfFieldsExistOnTables(ctx, tt.mappedFields)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUseCase_ValidateIfFieldsExistOnTables_InvalidDataSource(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because ResetRegisteredDataSourceIDsForTesting mutates global state

	// Register datasource IDs for testing - NOT including "unregistered_db"
	pkg.ResetRegisteredDataSourceIDsForTesting()
	pkg.RegisterDataSourceIDsForTesting([]string{"registered_db"})

	tests := []struct {
		name         string
		mappedFields map[string]map[string][]string
		dataSources  *pkg.SafeDataSources
		expectErr    bool
		errContains  string
	}{
		{
			name: "Error - Unregistered data source ID",
			mappedFields: map[string]map[string][]string{
				"unregistered_db": {
					"users": {"id"},
				},
			},
			dataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{}),
			expectErr:   true,
			errContains: "unregistered_db",
		},
		{
			name: "Error - Data source ID registered but not in runtime map",
			mappedFields: map[string]map[string][]string{
				"registered_db": {
					"users": {"id"},
				},
			},
			dataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
				// Data source is not in the map
			}),
			expectErr:   true,
			errContains: "registered_db",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			svc := &UseCase{
				ExternalDataSources: tt.dataSources,
			}

			ctx := context.Background()
			err := svc.ValidateIfFieldsExistOnTables(ctx, tt.mappedFields)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUseCase_ValidateIfFieldsExistOnTables_UnsupportedDatabaseType(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because ResetRegisteredDataSourceIDsForTesting mutates global state

	// Register datasource IDs for testing
	pkg.ResetRegisteredDataSourceIDsForTesting()
	pkg.RegisterDataSourceIDsForTesting([]string{"unknown_db"})

	svc := &UseCase{
		ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
			"unknown_db": {
				DatabaseType: "unsupported_type",
				Initialized:  true,
			},
		}),
	}

	mappedFields := map[string]map[string][]string{
		"unknown_db": {
			"users": {"id"},
		},
	}

	ctx := context.Background()
	err := svc.ValidateIfFieldsExistOnTables(ctx, mappedFields)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported database type")
}

func TestUseCase_TransformMappedFieldsForStorage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mappedFields   map[string]map[string][]string
		organizationID string
		expected       map[string]map[string][]string
	}{
		{
			name: "Success - Regular database no transformation",
			mappedFields: map[string]map[string][]string{
				"midaz_onboarding": {
					"users":    {"id", "name"},
					"accounts": {"id", "balance"},
				},
			},
			organizationID: "org-123",
			expected: map[string]map[string][]string{
				"midaz_onboarding": {
					"users":    {"id", "name"},
					"accounts": {"id", "balance"},
				},
			},
		},
		{
			name: "Success - Plugin CRM database adds organization mapping",
			mappedFields: map[string]map[string][]string{
				"plugin_crm": {
					"contacts": {"id", "name"},
				},
			},
			organizationID: "org-456",
			expected: map[string]map[string][]string{
				"plugin_crm": {
					"contacts":     {"id", "name"},
					"organization": {"org-456"},
				},
			},
		},
		{
			name: "Success - Multiple databases mixed transformation",
			mappedFields: map[string]map[string][]string{
				"midaz_onboarding": {
					"users": {"id"},
				},
				"plugin_crm": {
					"leads": {"name", "email"},
				},
			},
			organizationID: "org-789",
			expected: map[string]map[string][]string{
				"midaz_onboarding": {
					"users": {"id"},
				},
				"plugin_crm": {
					"leads":        {"name", "email"},
					"organization": {"org-789"},
				},
			},
		},
		{
			name:           "Success - Empty mapped fields",
			mappedFields:   map[string]map[string][]string{},
			organizationID: "org-123",
			expected:       map[string]map[string][]string{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := TransformMappedFieldsForStorage(tt.mappedFields, tt.organizationID)

			// Check database count
			assert.Equal(t, len(tt.expected), len(result))

			// Check each database's tables
			for dbName, expectedTables := range tt.expected {
				resultTables, exists := result[dbName]
				require.True(t, exists, "Expected database %s to exist in result", dbName)
				assert.Equal(t, len(expectedTables), len(resultTables), "Table count mismatch for database %s", dbName)

				for tableName, expectedFields := range expectedTables {
					resultFields, tableExists := resultTables[tableName]
					require.True(t, tableExists, "Expected table %s to exist in database %s", tableName, dbName)
					assert.ElementsMatch(t, expectedFields, resultFields, "Fields mismatch for table %s in database %s", tableName, dbName)
				}
			}
		})
	}
}

func TestUseCase_GenerateCopyOfMappedFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		original      map[string]map[string][]string
		dataSources   map[string]pkg.DataSource
		checkModified func(t *testing.T, original, copy map[string]map[string][]string)
	}{
		{
			name: "Success - Creates deep copy regular database",
			original: map[string]map[string][]string{
				"midaz_onboarding": {
					"users": {"id", "name"},
				},
			},
			dataSources: map[string]pkg.DataSource{
				"midaz_onboarding": {},
			},
			checkModified: func(t *testing.T, original, copiedFields map[string]map[string][]string) {
				// Verify copy has same structure
				assert.Contains(t, copiedFields, "midaz_onboarding")
				assert.Contains(t, copiedFields["midaz_onboarding"], "users")
				assert.ElementsMatch(t, []string{"id", "name"}, copiedFields["midaz_onboarding"]["users"])

				// Modify copy and verify original is unchanged
				copiedFields["midaz_onboarding"]["users"] = append(copiedFields["midaz_onboarding"]["users"], "modified")
				assert.NotContains(t, original["midaz_onboarding"]["users"], "modified")
			},
		},
		{
			name: "Success - Plugin CRM appends MidazOrganizationID to table names",
			original: map[string]map[string][]string{
				"plugin_crm": {
					"contacts": {"id", "name"},
				},
			},
			dataSources: map[string]pkg.DataSource{
				"plugin_crm": {MidazOrganizationID: "org-456"},
			},
			checkModified: func(t *testing.T, original, copiedFields map[string]map[string][]string) {
				// Verify copy has modified table name with organization ID
				assert.Contains(t, copiedFields, "plugin_crm")
				assert.Contains(t, copiedFields["plugin_crm"], "contacts_org-456")
				assert.ElementsMatch(t, []string{"id", "name"}, copiedFields["plugin_crm"]["contacts_org-456"])

				// Verify original is unchanged
				assert.Contains(t, original["plugin_crm"], "contacts")
				assert.NotContains(t, original["plugin_crm"], "contacts_org-456")
			},
		},
		{
			name: "Success - Plugin CRM no MidazOrganizationID keeps original table names",
			original: map[string]map[string][]string{
				"plugin_crm": {
					"contacts": {"id", "name"},
				},
			},
			dataSources: map[string]pkg.DataSource{
				"plugin_crm": {},
			},
			checkModified: func(t *testing.T, original, copiedFields map[string]map[string][]string) {
				// Verify copy keeps original table name when no MidazOrganizationID
				assert.Contains(t, copiedFields, "plugin_crm")
				assert.Contains(t, copiedFields["plugin_crm"], "contacts")
				assert.ElementsMatch(t, []string{"id", "name"}, copiedFields["plugin_crm"]["contacts"])
			},
		},
		{
			name:        "Success - Empty mapped fields",
			original:    map[string]map[string][]string{},
			dataSources: map[string]pkg.DataSource{},
			checkModified: func(t *testing.T, original, copiedFields map[string]map[string][]string) {
				assert.Empty(t, copiedFields)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			copiedFields := generateCopyOfMappedFields(tt.original, tt.dataSources)
			tt.checkModified(t, tt.original, copiedFields)
		})
	}
}

func TestUseCase_ValidateSchemaAmbiguity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		databaseName string
		schema       []postgres.TableSchema
		mappedFields map[string]map[string][]string
		expectErr    bool
		errContains  string
	}{
		{
			name:         "Success - No ambiguity table exists in single schema",
			databaseName: "test_db",
			schema: []postgres.TableSchema{
				{SchemaName: "public", TableName: "users"},
				{SchemaName: "public", TableName: "accounts"},
			},
			mappedFields: map[string]map[string][]string{
				"test_db": {
					"users": {"id", "name"},
				},
			},
			expectErr: false,
		},
		{
			name:         "Success - No ambiguity table in multiple schemas but has public",
			databaseName: "test_db",
			schema: []postgres.TableSchema{
				{SchemaName: "public", TableName: "users"},
				{SchemaName: "billing", TableName: "users"},
			},
			mappedFields: map[string]map[string][]string{
				"test_db": {
					"users": {"id", "name"},
				},
			},
			expectErr: false, // public exists, so no ambiguity
		},
		{
			name:         "Error - Ambiguity table in multiple schemas without public",
			databaseName: "test_db",
			schema: []postgres.TableSchema{
				{SchemaName: "billing", TableName: "users"},
				{SchemaName: "sales", TableName: "users"},
			},
			mappedFields: map[string]map[string][]string{
				"test_db": {
					"users": {"id", "name"},
				},
			},
			expectErr:   true,
			errContains: "users",
		},
		{
			name:         "Success - No ambiguity explicit schema with Pongo2 format",
			databaseName: "test_db",
			schema: []postgres.TableSchema{
				{SchemaName: "billing", TableName: "users"},
				{SchemaName: "sales", TableName: "users"},
			},
			mappedFields: map[string]map[string][]string{
				"test_db": {
					"billing__users": {"id", "name"}, // Explicit schema
				},
			},
			expectErr: false, // explicit schema, no ambiguity check needed
		},
		{
			name:         "Success - No ambiguity explicit schema with dot format",
			databaseName: "test_db",
			schema: []postgres.TableSchema{
				{SchemaName: "billing", TableName: "users"},
				{SchemaName: "sales", TableName: "users"},
			},
			mappedFields: map[string]map[string][]string{
				"test_db": {
					"billing.users": {"id", "name"}, // Explicit schema
				},
			},
			expectErr: false, // explicit schema, no ambiguity check needed
		},
		{
			name:         "Success - Table does not exist (caught by other validation)",
			databaseName: "test_db",
			schema: []postgres.TableSchema{
				{SchemaName: "public", TableName: "accounts"},
			},
			mappedFields: map[string]map[string][]string{
				"test_db": {
					"nonexistent": {"id"},
				},
			},
			expectErr: false, // this function doesn't validate existence
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateSchemaAmbiguity(tt.databaseName, tt.schema, tt.mappedFields)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUseCase_ValidateIfFieldsExistOnTables_PostgresWithSchemaFormats(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because ResetRegisteredDataSourceIDsForTesting mutates global state
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostgresRepo := postgres.NewMockRepository(ctrl)

	// Register datasource IDs for testing
	pkg.ResetRegisteredDataSourceIDsForTesting()
	pkg.RegisterDataSourceIDsForTesting([]string{"test_postgres_db"})

	postgresSchema := []postgres.TableSchema{
		{
			SchemaName: "analytics",
			TableName:  "transfers",
			Columns: []postgres.ColumnInformation{
				{Name: "id", DataType: "uuid"},
				{Name: "amount", DataType: "numeric"},
				{Name: "status", DataType: "varchar"},
			},
		},
	}

	tests := []struct {
		name         string
		mappedFields map[string]map[string][]string
		mockSetup    func()
		expectErr    bool
	}{
		{
			name: "Success - Pongo2 format schema__table",
			mappedFields: map[string]map[string][]string{
				"test_postgres_db": {
					"analytics__transfers": {"id", "amount"},
				},
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().
					GetDatabaseSchema(gomock.Any(), []string{"public", "analytics"}).
					Return(postgresSchema, nil)
				mockPostgresRepo.EXPECT().
					CloseConnection().
					Return(nil)
			},
			expectErr: false,
		},
		{
			name: "Success - Qualified format schema.table",
			mappedFields: map[string]map[string][]string{
				"test_postgres_db": {
					"analytics.transfers": {"id", "amount"},
				},
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().
					GetDatabaseSchema(gomock.Any(), []string{"public", "analytics"}).
					Return(postgresSchema, nil)
				mockPostgresRepo.EXPECT().
					CloseConnection().
					Return(nil)
			},
			expectErr: false,
		},
		{
			name: "Success - Legacy format just table name",
			mappedFields: map[string]map[string][]string{
				"test_postgres_db": {
					"transfers": {"id", "amount"},
				},
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().
					GetDatabaseSchema(gomock.Any(), []string{"public", "analytics"}).
					Return(postgresSchema, nil)
				mockPostgresRepo.EXPECT().
					CloseConnection().
					Return(nil)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			svc := &UseCase{
				ExternalDataSources: pkg.NewSafeDataSources(map[string]pkg.DataSource{
					"test_postgres_db": {
						DatabaseType:       pkg.PostgreSQLType,
						PostgresRepository: mockPostgresRepo,
						Initialized:        true,
						Schemas:            []string{"public", "analytics"},
						DatabaseConfig: &postgres.Connection{
							Connected: true,
						},
					},
				}),
			}

			ctx := context.Background()
			err := svc.ValidateIfFieldsExistOnTables(ctx, tt.mappedFields)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
