package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/v4/components/manager/internal/adapters/redis"
	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/model"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb"
	"github.com/LerianStudio/reporter/v4/pkg/postgres"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libConstants "github.com/LerianStudio/lib-commons/v2/commons/constants"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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

func Test_getBaseCollectionName(t *testing.T) {
	uc := &UseCase{}

	tests := []struct {
		name           string
		collectionName string
		expected       string
	}{
		{
			name:           "Collection with organization suffix",
			collectionName: "holders_org-123",
			expected:       "holders",
		},
		{
			name:           "Collection without underscore",
			collectionName: "users",
			expected:       "users",
		},
		{
			name:           "Collection with multiple underscores",
			collectionName: "some_long_name_org-456",
			expected:       "some_long_name",
		},
		{
			name:           "Empty string",
			collectionName: "",
			expected:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uc.getBaseCollectionName(tt.collectionName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_getDisplayNameForCollection(t *testing.T) {
	uc := &UseCase{}

	tests := []struct {
		name           string
		collectionName string
		dataSourceID   string
		expected       string
	}{
		{
			name:           "plugin_crm - extracts base name",
			collectionName: "holders_org-123",
			dataSourceID:   "plugin_crm",
			expected:       "holders",
		},
		{
			name:           "Other database - returns original name",
			collectionName: "users",
			dataSourceID:   "regular_db",
			expected:       "users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uc.getDisplayNameForCollection(tt.collectionName, tt.dataSourceID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_shouldIncludeFieldForPluginCRM(t *testing.T) {
	uc := &UseCase{}

	tests := []struct {
		name           string
		fieldName      string
		collectionName string
		expected       bool
	}{
		{
			name:           "Search field - included",
			fieldName:      "search.document",
			collectionName: "holders",
			expected:       true,
		},
		{
			name:           "Encrypted field - excluded",
			fieldName:      "document",
			collectionName: "holders",
			expected:       false,
		},
		{
			name:           "Name field - excluded",
			fieldName:      "name",
			collectionName: "holders",
			expected:       false,
		},
		{
			name:           "Nested encrypted field - excluded",
			fieldName:      "contact.primary_email",
			collectionName: "holders",
			expected:       false,
		},
		{
			name:           "Regular field in holders - included",
			fieldName:      "external_id",
			collectionName: "holders",
			expected:       true,
		},
		{
			name:           "Regular field in aliases - included",
			fieldName:      "account_id",
			collectionName: "aliases",
			expected:       true,
		},
		{
			name:           "Regular field in other collection - included",
			fieldName:      "some_field",
			collectionName: "other_collection",
			expected:       true,
		},
		{
			name:           "Search object - included",
			fieldName:      "search",
			collectionName: "holders",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uc.shouldIncludeFieldForPluginCRM(tt.fieldName, tt.collectionName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_getExpandedFieldsForPluginCRM(t *testing.T) {
	uc := &UseCase{}

	tests := []struct {
		name           string
		collectionName string
		expectNil      bool
		expectFields   []string
	}{
		{
			name:           "holders collection - returns expanded fields",
			collectionName: "holders",
			expectNil:      false,
			expectFields: []string{
				"_id",
				"external_id",
				"type",
				"addresses",
				"created_at",
				"updated_at",
				"deleted_at",
				"metadata",
				"search.document",
				"natural_person.favorite_name",
				"natural_person.social_name",
				"natural_person.gender",
				"natural_person.birth_date",
				"natural_person.civil_status",
				"natural_person.nationality",
				"natural_person.status",
				"legal_person.trade_name",
				"legal_person.activity",
				"legal_person.type",
				"legal_person.founding_date",
				"legal_person.size",
				"legal_person.status",
				"legal_person.representative.role",
			},
		},
		{
			name:           "aliases collection - returns expanded fields",
			collectionName: "aliases",
			expectNil:      false,
			expectFields: []string{
				"_id",
				"account_id",
				"holder_id",
				"ledger_id",
				"type",
				"created_at",
				"updated_at",
				"deleted_at",
				"metadata",
				"search.document",
				"search.banking_details_account",
				"search.banking_details_iban",
				"banking_details.branch",
				"banking_details.type",
				"banking_details.opening_date",
				"banking_details.country_code",
				"banking_details.bank_id",
			},
		},
		{
			name:           "unknown collection - returns nil",
			collectionName: "unknown",
			expectNil:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uc.getExpandedFieldsForPluginCRM(tt.collectionName)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expectFields, result)
			}
		})
	}
}

func Test_getFieldsForCollection(t *testing.T) {
	uc := &UseCase{}

	tests := []struct {
		name         string
		collection   mongodb.CollectionSchema
		dataSourceID string
		expectLen    int
	}{
		{
			name: "plugin_crm with holders - uses expanded fields",
			collection: mongodb.CollectionSchema{
				CollectionName: "holders_org-123",
				Fields: []mongodb.FieldInformation{
					{Name: "_id", DataType: "ObjectId"},
					{Name: "document", DataType: "string"},
				},
			},
			dataSourceID: "plugin_crm",
			expectLen:    23, // holders expanded fields count
		},
		{
			name: "plugin_crm with aliases - uses expanded fields",
			collection: mongodb.CollectionSchema{
				CollectionName: "aliases_org-123",
				Fields: []mongodb.FieldInformation{
					{Name: "_id", DataType: "ObjectId"},
				},
			},
			dataSourceID: "plugin_crm",
			expectLen:    17, // aliases expanded fields count
		},
		{
			name: "plugin_crm with unknown collection - filters fields",
			collection: mongodb.CollectionSchema{
				CollectionName: "unknown_org-123",
				Fields: []mongodb.FieldInformation{
					{Name: "_id", DataType: "ObjectId"},
					{Name: "external_id", DataType: "string"},
					{Name: "document", DataType: "string"}, // encrypted - excluded
				},
			},
			dataSourceID: "plugin_crm",
			expectLen:    2, // Only _id and external_id
		},
		{
			name: "regular database - includes all fields",
			collection: mongodb.CollectionSchema{
				CollectionName: "users",
				Fields: []mongodb.FieldInformation{
					{Name: "_id", DataType: "ObjectId"},
					{Name: "name", DataType: "string"},
					{Name: "email", DataType: "string"},
				},
			},
			dataSourceID: "regular_db",
			expectLen:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uc.getFieldsForCollection(tt.collection, tt.dataSourceID)
			assert.Equal(t, tt.expectLen, len(result))
		})
	}
}

func Test_ensureDataSourceConnected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	tests := []struct {
		name         string
		dataSourceID string
		dataSource   *pkg.DataSource
		expectError  bool
	}{
		{
			name:         "PostgreSQL already connected",
			dataSourceID: "pg_ds",
			dataSource: &pkg.DataSource{
				DatabaseType:   pkg.PostgreSQLType,
				Initialized:    true,
				DatabaseConfig: &postgres.Connection{Connected: true},
			},
			expectError: false,
		},
		{
			name:         "MongoDB already initialized",
			dataSourceID: "mongo_ds",
			dataSource: &pkg.DataSource{
				DatabaseType: pkg.MongoDBType,
				Initialized:  true,
			},
			expectError: false,
		},
		{
			name:         "Unavailable datasource logs warning but continues",
			dataSourceID: "unavailable_ds",
			dataSource: &pkg.DataSource{
				DatabaseType:   pkg.PostgreSQLType,
				Initialized:    true,
				DatabaseConfig: &postgres.Connection{Connected: true},
				Status:         libConstants.DataSourceStatusUnavailable,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &UseCase{
				ExternalDataSources: map[string]pkg.DataSource{
					tt.dataSourceID: *tt.dataSource,
				},
			}

			err := uc.ensureDataSourceConnected(mockLogger, tt.dataSourceID, tt.dataSource)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_getDataSourceDetailsFromCache_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedisRepo := redis.NewMockRedisRepository(ctrl)

	uc := &UseCase{
		RedisRepo: mockRedisRepo,
	}

	cacheKey := "test-key"

	// Return invalid JSON
	mockRedisRepo.EXPECT().Get(gomock.Any(), cacheKey).Return("invalid-json{}", nil)

	result, ok := uc.getDataSourceDetailsFromCache(ctx, cacheKey)

	assert.False(t, ok)
	assert.Nil(t, result)
}

func Test_getDataSourceDetailsFromCache_NilRedisRepo(t *testing.T) {
	ctx := context.Background()

	uc := &UseCase{
		RedisRepo: nil,
	}

	result, ok := uc.getDataSourceDetailsFromCache(ctx, "any-key")

	assert.False(t, ok)
	assert.Nil(t, result)
}

func Test_setDataSourceDetailsToCache_NilDetails(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedisRepo := redis.NewMockRedisRepository(ctrl)

	uc := &UseCase{
		RedisRepo: mockRedisRepo,
	}

	// Should return nil without calling Redis
	err := uc.setDataSourceDetailsToCache(ctx, "any-key", nil)
	assert.NoError(t, err)
}

func Test_setDataSourceDetailsToCache_NilRedisRepo(t *testing.T) {
	ctx := context.Background()

	uc := &UseCase{
		RedisRepo: nil,
	}

	details := &model.DataSourceDetails{
		Id: "test",
	}

	err := uc.setDataSourceDetailsToCache(ctx, "any-key", details)
	assert.NoError(t, err)
}

func Test_GetDataSourceDetailsByID_UnsupportedDatabaseType(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedisRepo := redis.NewMockRedisRepository(ctrl)

	uc := &UseCase{
		ExternalDataSources: map[string]pkg.DataSource{
			"unsupported_ds": {
				DatabaseType: "unsupported",
				Initialized:  true,
			},
		},
		RedisRepo: mockRedisRepo,
	}

	mockRedisRepo.EXPECT().Get(gomock.Any(), constant.DataSourceDetailsKeyPrefix+":unsupported_ds").Return("", nil)

	result, err := uc.GetDataSourceDetailsByID(ctx, "unsupported_ds")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func Test_GetDataSourceDetailsByID_CacheSetError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMongoRepo := mongodb.NewMockRepository(ctrl)
	mockRedisRepo := redis.NewMockRedisRepository(ctrl)

	mongoSchema := []mongodb.CollectionSchema{
		{
			CollectionName: "collection1",
			Fields: []mongodb.FieldInformation{
				{Name: "field1", DataType: "string"},
			},
		},
	}

	mongoResult := &model.DataSourceDetails{
		Id:           "mongo_ds",
		ExternalName: "mongo_db",
		Type:         pkg.MongoDBType,
		Tables: []model.TableDetails{{
			Name:   "collection1",
			Fields: []string{"field1"},
		}},
	}
	mongoResultJSON, _ := json.Marshal(mongoResult)
	cacheKey := constant.DataSourceDetailsKeyPrefix + ":mongo_ds"

	uc := &UseCase{
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

	mockRedisRepo.EXPECT().Get(gomock.Any(), cacheKey).Return("", nil)
	mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(mongoSchema, nil)
	mockMongoRepo.EXPECT().CloseConnection(gomock.Any()).Return(nil)
	mockRedisRepo.EXPECT().Set(gomock.Any(), cacheKey, string(mongoResultJSON), time.Second*time.Duration(constant.RedisTTL)).Return(errors.New("cache error"))

	result, err := uc.GetDataSourceDetailsByID(ctx, "mongo_ds")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func Test_GetDataSourceDetailsByID_MongoCloseError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMongoRepo := mongodb.NewMockRepository(ctrl)
	mockRedisRepo := redis.NewMockRedisRepository(ctrl)

	mongoSchema := []mongodb.CollectionSchema{
		{
			CollectionName: "collection1",
			Fields: []mongodb.FieldInformation{
				{Name: "field1", DataType: "string"},
			},
		},
	}

	cacheKey := constant.DataSourceDetailsKeyPrefix + ":mongo_ds"

	uc := &UseCase{
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

	mockRedisRepo.EXPECT().Get(gomock.Any(), cacheKey).Return("", nil)
	mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(mongoSchema, nil)
	mockMongoRepo.EXPECT().CloseConnection(gomock.Any()).Return(errors.New("close error"))

	result, err := uc.GetDataSourceDetailsByID(ctx, "mongo_ds")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func Test_GetDataSourceDetailsByID_PostgresCloseError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostgresRepo := postgres.NewMockRepository(ctrl)
	mockRedisRepo := redis.NewMockRedisRepository(ctrl)

	postgresSchema := []postgres.TableSchema{
		{
			TableName: "table1",
			Columns: []postgres.ColumnInformation{
				{Name: "col1", DataType: "string"},
			},
		},
	}

	cacheKey := constant.DataSourceDetailsKeyPrefix + ":pg_ds"

	uc := &UseCase{
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

	mockRedisRepo.EXPECT().Get(gomock.Any(), cacheKey).Return("", nil)
	mockPostgresRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(postgresSchema, nil)
	mockPostgresRepo.EXPECT().CloseConnection().Return(errors.New("close error"))

	result, err := uc.GetDataSourceDetailsByID(ctx, "pg_ds")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func Test_GetDataSourceDetailsByID_PluginCRM(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMongoRepo := mongodb.NewMockRepository(ctrl)
	mockRedisRepo := redis.NewMockRedisRepository(ctrl)

	mongoSchema := []mongodb.CollectionSchema{
		{
			CollectionName: "holders_org-123",
			Fields: []mongodb.FieldInformation{
				{Name: "_id", DataType: "ObjectId"},
				{Name: "external_id", DataType: "string"},
				{Name: "document", DataType: "string"},
			},
		},
	}

	cacheKey := constant.DataSourceDetailsKeyPrefix + ":plugin_crm"

	uc := &UseCase{
		ExternalDataSources: map[string]pkg.DataSource{
			"plugin_crm": {
				DatabaseType:      pkg.MongoDBType,
				MongoDBRepository: mockMongoRepo,
				MongoDBName:       "crm_db",
				Initialized:       true,
			},
		},
		RedisRepo: mockRedisRepo,
	}

	mockRedisRepo.EXPECT().Get(gomock.Any(), cacheKey).Return("", nil)
	mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(mongoSchema, nil)
	mockMongoRepo.EXPECT().CloseConnection(gomock.Any()).Return(nil)
	mockRedisRepo.EXPECT().Set(gomock.Any(), cacheKey, gomock.Any(), time.Second*time.Duration(constant.RedisTTL)).Return(nil)

	result, err := uc.GetDataSourceDetailsByID(ctx, "plugin_crm")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "plugin_crm", result.Id)
	// The table name should be "holders" (base name extracted)
	assert.Equal(t, "holders", result.Tables[0].Name)
	// Should use expanded fields for holders
	assert.Equal(t, 23, len(result.Tables[0].Fields))
}
