package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"plugin-smart-templates/v2/components/manager/internal/adapters/redis"

	"plugin-smart-templates/v2/pkg"
	"plugin-smart-templates/v2/pkg/constant"
	_ "plugin-smart-templates/v2/pkg/constant"
	"plugin-smart-templates/v2/pkg/model"
	"plugin-smart-templates/v2/pkg/mongodb"
	"plugin-smart-templates/v2/pkg/postgres"
	"testing"
	"time"
)

func Test_GetDataSourceDetailsByID(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMongoRepo := mongodb.NewMockRepository(ctrl)
	mockPostgresRepo := postgres.NewMockRepository(ctrl)
	mockRedisRepo := redis.NewMockRedisRepository(ctrl)

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

	cacheKey := constant.DataSourceDetailsKeyPrefix + ":mongo_ds"
	cacheKeyPG := constant.DataSourceDetailsKeyPrefix + ":pg_ds"

	mongoResult := &model.DataSourceDetails{
		Id:           "mongo_ds",
		ExternalName: "mongo_db",
		Type:         pkg.MongoDBType,
		Tables: []model.TableDetails{{
			Name:   "collection1",
			Fields: []string{"field1", "field2"},
		}},
	}
	pgResult := &model.DataSourceDetails{
		Id:           "pg_ds",
		ExternalName: "pg_db",
		Type:         pkg.PostgreSQLType,
		Tables: []model.TableDetails{{
			Name:   "table1",
			Fields: []string{"col1", "col2"},
		}},
	}
	mongoResultJSON, _ := json.Marshal(mongoResult)
	pgResultJSON, _ := json.Marshal(pgResult)

	tests := []struct {
		name         string
		setupSvc     func() *UseCase
		dataSourceID string
		mockSetup    func()
		expectErr    bool
		expectResult *model.DataSourceDetails
	}{
		{
			name:         "Cache hit - MongoDB",
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
					RedisRepo: mockRedisRepo,
				}
			},
			mockSetup: func() {
				mockRedisRepo.EXPECT().Get(gomock.Any(), cacheKey).Return(string(mongoResultJSON), nil)
			},
			expectErr:    false,
			expectResult: mongoResult,
		},
		{
			name:         "Cache miss - MongoDB, sets cache",
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
					RedisRepo: mockRedisRepo,
				}
			},
			mockSetup: func() {
				mockRedisRepo.EXPECT().Get(gomock.Any(), cacheKey).Return("", nil)
				mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(mongoSchema, nil)
				mockMongoRepo.EXPECT().CloseConnection(gomock.Any()).Return(nil)
				mockRedisRepo.EXPECT().Set(gomock.Any(), cacheKey, string(mongoResultJSON), time.Second*time.Duration(constant.RedisTTL)).Return(nil)
			},
			expectErr:    false,
			expectResult: mongoResult,
		},
		{
			name:         "Cache error - MongoDB, acts as miss",
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
					RedisRepo: mockRedisRepo,
				}
			},
			mockSetup: func() {
				mockRedisRepo.EXPECT().Get(gomock.Any(), cacheKey).Return("", errors.New("redis error"))
				mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(mongoSchema, nil)
				mockMongoRepo.EXPECT().CloseConnection(gomock.Any()).Return(nil)
				mockRedisRepo.EXPECT().Set(gomock.Any(), cacheKey, string(mongoResultJSON), time.Second*time.Duration(constant.RedisTTL)).Return(nil)
			},
			expectErr:    false,
			expectResult: mongoResult,
		},
		{
			name:         "Cache hit - PostgreSQL",
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
					RedisRepo: mockRedisRepo,
				}
			},
			mockSetup: func() {
				mockRedisRepo.EXPECT().Get(gomock.Any(), cacheKeyPG).Return(string(pgResultJSON), nil)
			},
			expectErr:    false,
			expectResult: pgResult,
		},
		{
			name:         "Cache miss - PostgreSQL, sets cache",
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
					RedisRepo: mockRedisRepo,
				}
			},
			mockSetup: func() {
				mockRedisRepo.EXPECT().Get(gomock.Any(), cacheKeyPG).Return("", nil)
				mockPostgresRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(postgresSchema, nil)
				mockPostgresRepo.EXPECT().CloseConnection().Return(nil)
				mockRedisRepo.EXPECT().Set(gomock.Any(), cacheKeyPG, string(pgResultJSON), time.Second*time.Duration(constant.RedisTTL)).Return(nil)
			},
			expectErr:    false,
			expectResult: pgResult,
		},
		{
			name:         "Error - Data source not found",
			dataSourceID: "not_found",
			setupSvc: func() *UseCase {
				return &UseCase{ExternalDataSources: map[string]pkg.DataSource{}, RedisRepo: mockRedisRepo}
			},
			mockSetup: func() {
				mockRedisRepo.EXPECT().Get(gomock.Any(), constant.DataSourceDetailsKeyPrefix+":not_found").Return("", nil)
			},
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
					RedisRepo: mockRedisRepo,
				}
			},
			mockSetup: func() {
				mockRedisRepo.EXPECT().Get(gomock.Any(), cacheKey).Return("", nil)
				mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(nil, errors.New("db error"))
				mockMongoRepo.EXPECT().CloseConnection(gomock.Any()).Return(nil)
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
					RedisRepo: mockRedisRepo,
				}
			},
			mockSetup: func() {
				mockRedisRepo.EXPECT().Get(gomock.Any(), cacheKeyPG).Return("", nil)
				mockPostgresRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(nil, errors.New("db error"))
				mockPostgresRepo.EXPECT().CloseConnection().Return(nil)
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
