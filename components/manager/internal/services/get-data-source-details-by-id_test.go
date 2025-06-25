package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"plugin-smart-templates/pkg"
	_ "plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/model"
	"plugin-smart-templates/pkg/mongodb"
	"plugin-smart-templates/pkg/postgres"
)

func Test_GetDataSourceDetailsByID(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMongoRepo := mongodb.NewMockRepository(ctrl)
	mockPostgresRepo := postgres.NewMockRepository(ctrl)

	mongoSchema := []mongodb.CollectionSchema{
		{
			CollectionName: "collection1",
			Fields: []mongodb.FieldInformation{
				{Name: "field1", DataType: "string"},
				{Name: "field2", DataType: "int"},
			},
		},
	}
	postgresSchema := []postgres.TableSchema{
		{
			TableName: "table1",
			Columns: []postgres.ColumnInformation{
				{Name: "col1", DataType: "string"},
				{Name: "col2", DataType: "int"},
			},
		},
	}

	tests := []struct {
		name         string
		setupSvc     func() *UseCase
		dataSourceID string
		mockSetup    func()
		expectErr    bool
		expectResult *model.DataSourceDetails
	}{
		{
			name:         "Success - MongoDB initialized",
			dataSourceID: "mongo_ds",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"mongo_ds": {
							DatabaseType:      pkg.MongoDBType,
							MongoDBRepository: mockMongoRepo,
							MongoDBName:       "mongo_db",
							Initialized:       true,
						},
					},
				}
			},
			mockSetup: func() {
				mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(mongoSchema, nil)
			},
			expectErr: false,
			expectResult: &model.DataSourceDetails{
				Id:           "mongo_ds",
				ExternalName: "mongo_db",
				Type:         pkg.MongoDBType,
				Tables: []model.TableDetails{{
					Name:   "collection1",
					Fields: []string{"field1", "field2"},
				}},
			},
		},
		{
			name:         "Success - PostgreSQL initialized",
			dataSourceID: "pg_ds",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"pg_ds": {
							DatabaseType:       pkg.PostgreSQLType,
							PostgresRepository: mockPostgresRepo,
							DatabaseConfig:     &postgres.Connection{Connected: true, DBName: "pg_db"},
							MongoDBName:        "pg_db",
							Initialized:        true,
						},
					},
				}
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(postgresSchema, nil)
			},
			expectErr: false,
			expectResult: &model.DataSourceDetails{
				Id:           "pg_ds",
				ExternalName: "pg_db",
				Type:         pkg.PostgreSQLType,
				Tables: []model.TableDetails{{
					Name:   "table1",
					Fields: []string{"col1", "col2"},
				}},
			},
		},
		{
			name:         "Error - Data source not found",
			dataSourceID: "not_found",
			setupSvc: func() *UseCase {
				return &UseCase{ExternalDataSources: map[string]pkg.DataSource{}}
			},
			mockSetup:    func() {},
			expectErr:    true,
			expectResult: nil,
		},
		{
			name:         "Error - MongoDB repo returns error",
			dataSourceID: "mongo_ds",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"mongo_ds": {
							DatabaseType:      pkg.MongoDBType,
							MongoDBRepository: mockMongoRepo,
							MongoDBName:       "mongo_db",
							Initialized:       true,
						},
					},
				}
			},
			mockSetup: func() {
				mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(nil, errors.New("db error"))
			},
			expectErr:    true,
			expectResult: nil,
		},
		{
			name:         "Error - PostgreSQL repo returns error",
			dataSourceID: "pg_ds",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"pg_ds": {
							DatabaseType:       pkg.PostgreSQLType,
							PostgresRepository: mockPostgresRepo,
							DatabaseConfig:     &postgres.Connection{Connected: true, DBName: "pg_db"},
							MongoDBName:        "pg_db",
							Initialized:        true,
						},
					},
				}
			},
			mockSetup: func() {
				mockPostgresRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(nil, errors.New("db error"))
			},
			expectErr:    true,
			expectResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setupSvc()
			tt.mockSetup()
			result, err := svc.GetDataSourceDetailsByID(ctx, tt.dataSourceID)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectResult, result)
			}
		})
	}
}
