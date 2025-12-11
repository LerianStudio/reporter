package services

import (
	"context"
	"errors"
	"testing"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb"
	"github.com/LerianStudio/reporter/v4/pkg/postgres"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestValidateIfFieldsExistOnTables(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockMongoRepo := mongodb.NewMockRepository(ctrl)
	mockPostgresRepo := postgres.NewMockRepository(ctrl)

	// Allow any logging calls
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	tests := []struct {
		name         string
		setupSvc     func() *UseCase
		orgID        string
		mappedFields map[string]map[string][]string
		mockSetup    func()
		expectErr    bool
	}{
		{
			name: "Success - PostgreSQL validation passes",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"pg_ds": {
							DatabaseType:       pkg.PostgreSQLType,
							PostgresRepository: mockPostgresRepo,
							DatabaseConfig:     &postgres.Connection{Connected: true},
							Initialized:        true,
						},
					},
				}
			},
			orgID: "org-123",
			mappedFields: map[string]map[string][]string{
				"pg_ds": {
					"users": {"id", "name"},
				},
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return([]postgres.TableSchema{
					{
						TableName: "users",
						Columns: []postgres.ColumnInformation{
							{Name: "id", DataType: "uuid"},
							{Name: "name", DataType: "varchar"},
							{Name: "email", DataType: "varchar"},
						},
					},
				}, nil)
				mockPostgresRepo.EXPECT().CloseConnection().Return(nil)
			},
			expectErr: false,
		},
		{
			name: "Success - MongoDB validation passes",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"mongo_ds": {
							DatabaseType:      pkg.MongoDBType,
							MongoDBRepository: mockMongoRepo,
							Initialized:       true,
						},
					},
				}
			},
			orgID: "org-123",
			mappedFields: map[string]map[string][]string{
				"mongo_ds": {
					"users": {"_id", "name"},
				},
			},
			mockSetup: func() {
				mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return([]mongodb.CollectionSchema{
					{
						CollectionName: "users",
						Fields: []mongodb.FieldInformation{
							{Name: "_id", DataType: "ObjectId"},
							{Name: "name", DataType: "string"},
							{Name: "email", DataType: "string"},
						},
					},
				}, nil)
				mockMongoRepo.EXPECT().CloseConnection(gomock.Any()).Return(nil)
			},
			expectErr: false,
		},
		{
			name: "Error - Unknown data source",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{},
				}
			},
			orgID: "org-123",
			mappedFields: map[string]map[string][]string{
				"unknown_ds": {
					"users": {"id"},
				},
			},
			mockSetup: func() {},
			expectErr: true,
		},
		{
			name: "Error - Unsupported database type",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"unknown_type": {
							DatabaseType: "unsupported",
							Initialized:  true,
						},
					},
				}
			},
			orgID: "org-123",
			mappedFields: map[string]map[string][]string{
				"unknown_type": {
					"users": {"id"},
				},
			},
			mockSetup: func() {},
			expectErr: true,
		},
		{
			name: "Error - PostgreSQL GetDatabaseSchema fails",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"pg_ds": {
							DatabaseType:       pkg.PostgreSQLType,
							PostgresRepository: mockPostgresRepo,
							DatabaseConfig:     &postgres.Connection{Connected: true},
							Initialized:        true,
						},
					},
				}
			},
			orgID: "org-123",
			mappedFields: map[string]map[string][]string{
				"pg_ds": {
					"users": {"id"},
				},
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(nil, errors.New("db error"))
			},
			expectErr: true,
		},
		{
			name: "Error - MongoDB GetDatabaseSchema fails",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"mongo_ds": {
							DatabaseType:      pkg.MongoDBType,
							MongoDBRepository: mockMongoRepo,
							Initialized:       true,
						},
					},
				}
			},
			orgID: "org-123",
			mappedFields: map[string]map[string][]string{
				"mongo_ds": {
					"users": {"_id"},
				},
			},
			mockSetup: func() {
				mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(nil, errors.New("db error"))
			},
			expectErr: true,
		},
		{
			name: "Error - PostgreSQL table not found",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"pg_ds": {
							DatabaseType:       pkg.PostgreSQLType,
							PostgresRepository: mockPostgresRepo,
							DatabaseConfig:     &postgres.Connection{Connected: true},
							Initialized:        true,
						},
					},
				}
			},
			orgID: "org-123",
			mappedFields: map[string]map[string][]string{
				"pg_ds": {
					"nonexistent_table": {"id"},
				},
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return([]postgres.TableSchema{
					{
						TableName: "users",
						Columns: []postgres.ColumnInformation{
							{Name: "id", DataType: "uuid"},
						},
					},
				}, nil)
			},
			expectErr: true,
		},
		{
			name: "Error - MongoDB collection not found",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"mongo_ds": {
							DatabaseType:      pkg.MongoDBType,
							MongoDBRepository: mockMongoRepo,
							Initialized:       true,
						},
					},
				}
			},
			orgID: "org-123",
			mappedFields: map[string]map[string][]string{
				"mongo_ds": {
					"nonexistent_collection": {"_id"},
				},
			},
			mockSetup: func() {
				mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return([]mongodb.CollectionSchema{
					{
						CollectionName: "users",
						Fields: []mongodb.FieldInformation{
							{Name: "_id", DataType: "ObjectId"},
						},
					},
				}, nil)
			},
			expectErr: true,
		},
		{
			name: "Error - PostgreSQL CloseConnection fails",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"pg_ds": {
							DatabaseType:       pkg.PostgreSQLType,
							PostgresRepository: mockPostgresRepo,
							DatabaseConfig:     &postgres.Connection{Connected: true},
							Initialized:        true,
						},
					},
				}
			},
			orgID: "org-123",
			mappedFields: map[string]map[string][]string{
				"pg_ds": {
					"users": {"id"},
				},
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return([]postgres.TableSchema{
					{
						TableName: "users",
						Columns: []postgres.ColumnInformation{
							{Name: "id", DataType: "uuid"},
						},
					},
				}, nil)
				mockPostgresRepo.EXPECT().CloseConnection().Return(errors.New("close error"))
			},
			expectErr: true,
		},
		{
			name: "Error - MongoDB CloseConnection fails",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"mongo_ds": {
							DatabaseType:      pkg.MongoDBType,
							MongoDBRepository: mockMongoRepo,
							Initialized:       true,
						},
					},
				}
			},
			orgID: "org-123",
			mappedFields: map[string]map[string][]string{
				"mongo_ds": {
					"users": {"_id"},
				},
			},
			mockSetup: func() {
				mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return([]mongodb.CollectionSchema{
					{
						CollectionName: "users",
						Fields: []mongodb.FieldInformation{
							{Name: "_id", DataType: "ObjectId"},
						},
					},
				}, nil)
				mockMongoRepo.EXPECT().CloseConnection(gomock.Any()).Return(errors.New("close error"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setupSvc()
			tt.mockSetup()

			err := svc.ValidateIfFieldsExistOnTables(ctx, tt.orgID, mockLogger, tt.mappedFields)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateCopyOfMappedFields(t *testing.T) {
	tests := []struct {
		name           string
		input          map[string]map[string][]string
		organizationID string
		expected       map[string]map[string][]string
	}{
		{
			name: "Regular database - no transformation",
			input: map[string]map[string][]string{
				"regular_db": {
					"users": {"id", "name"},
				},
			},
			organizationID: "org-123",
			expected: map[string]map[string][]string{
				"regular_db": {
					"users": {"id", "name"},
				},
			},
		},
		{
			name: "plugin_crm database - appends organizationID to table names",
			input: map[string]map[string][]string{
				"plugin_crm": {
					"holders": {"id", "name"},
					"aliases": {"id", "account_id"},
				},
			},
			organizationID: "org-123",
			expected: map[string]map[string][]string{
				"plugin_crm": {
					"holders_org-123": {"id", "name"},
					"aliases_org-123": {"id", "account_id"},
				},
			},
		},
		{
			name: "Mixed databases",
			input: map[string]map[string][]string{
				"plugin_crm": {
					"holders": {"id"},
				},
				"regular_db": {
					"users": {"id"},
				},
			},
			organizationID: "org-456",
			expected: map[string]map[string][]string{
				"plugin_crm": {
					"holders_org-456": {"id"},
				},
				"regular_db": {
					"users": {"id"},
				},
			},
		},
		{
			name:           "Empty input",
			input:          map[string]map[string][]string{},
			organizationID: "org-123",
			expected:       map[string]map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateCopyOfMappedFields(tt.input, tt.organizationID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTransformMappedFieldsForStorage(t *testing.T) {
	tests := []struct {
		name           string
		input          map[string]map[string][]string
		organizationID string
		expected       map[string]map[string][]string
	}{
		{
			name: "Regular database - no transformation",
			input: map[string]map[string][]string{
				"regular_db": {
					"users": {"id", "name"},
				},
			},
			organizationID: "org-123",
			expected: map[string]map[string][]string{
				"regular_db": {
					"users": {"id", "name"},
				},
			},
		},
		{
			name: "plugin_crm database - adds organization mapping",
			input: map[string]map[string][]string{
				"plugin_crm": {
					"holders": {"id", "name"},
				},
			},
			organizationID: "org-123",
			expected: map[string]map[string][]string{
				"plugin_crm": {
					"holders":      {"id", "name"},
					"organization": {"org-123"},
				},
			},
		},
		{
			name: "plugin_crm with multiple tables",
			input: map[string]map[string][]string{
				"plugin_crm": {
					"holders": {"id"},
					"aliases": {"account_id"},
				},
			},
			organizationID: "org-456",
			expected: map[string]map[string][]string{
				"plugin_crm": {
					"holders":      {"id"},
					"aliases":      {"account_id"},
					"organization": {"org-456"},
				},
			},
		},
		{
			name: "Mixed databases",
			input: map[string]map[string][]string{
				"plugin_crm": {
					"holders": {"id"},
				},
				"regular_db": {
					"users": {"id"},
				},
			},
			organizationID: "org-789",
			expected: map[string]map[string][]string{
				"plugin_crm": {
					"holders":      {"id"},
					"organization": {"org-789"},
				},
				"regular_db": {
					"users": {"id"},
				},
			},
		},
		{
			name:           "Empty input",
			input:          map[string]map[string][]string{},
			organizationID: "org-123",
			expected:       map[string]map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformMappedFieldsForStorage(tt.input, tt.organizationID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
