package in

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/reporter/v4/components/manager/internal/adapters/redis"
	"github.com/LerianStudio/reporter/v4/components/manager/internal/services"
	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/model"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb"
	"github.com/LerianStudio/reporter/v4/pkg/postgres"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
}

func TestDataSourceHandler_GetDataSourceInformation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pgConfig := &postgres.Connection{DBName: "pg_db"}
	orgID := uuid.New()

	tests := []struct {
		name           string
		setupService   func() *services.UseCase
		expectedStatus int
		expectedLen    int
	}{
		{
			name: "Success - Returns multiple data sources",
			setupService: func() *services.UseCase {
				return &services.UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"mongo_ds": {
							DatabaseType:      pkg.MongoDBType,
							MongoDBName:       "mongo_db",
							MongoDBRepository: mongodb.NewMockRepository(nil),
						},
						"pg_ds": {
							DatabaseType:       pkg.PostgreSQLType,
							DatabaseConfig:     pgConfig,
							PostgresRepository: postgres.NewMockRepository(nil),
						},
					},
				}
			},
			expectedStatus: fiber.StatusOK,
			expectedLen:    2,
		},
		{
			name: "Success - Returns empty list when no data sources",
			setupService: func() *services.UseCase {
				return &services.UseCase{
					ExternalDataSources: map[string]pkg.DataSource{},
				}
			},
			expectedStatus: fiber.StatusOK,
			expectedLen:    0,
		},
		{
			name: "Success - Returns only MongoDB data source",
			setupService: func() *services.UseCase {
				return &services.UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"mongo_ds": {
							DatabaseType:      pkg.MongoDBType,
							MongoDBName:       "mongo_db",
							MongoDBRepository: mongodb.NewMockRepository(nil),
						},
					},
				}
			},
			expectedStatus: fiber.StatusOK,
			expectedLen:    1,
		},
		{
			name: "Success - Returns only PostgreSQL data source",
			setupService: func() *services.UseCase {
				return &services.UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"pg_ds": {
							DatabaseType:       pkg.PostgreSQLType,
							DatabaseConfig:     pgConfig,
							PostgresRepository: postgres.NewMockRepository(nil),
						},
					},
				}
			},
			expectedStatus: fiber.StatusOK,
			expectedLen:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestApp()

			handler := &DataSourceHandler{
				Service: tt.setupService(),
			}

			app.Get("/v1/data-sources", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.SetUserContext(context.Background())
				return handler.GetDataSourceInformation(c)
			})

			req := httptest.NewRequest("GET", "/v1/data-sources", nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var result []*model.DataSourceInformation
			err = json.Unmarshal(body, &result)
			require.NoError(t, err)

			assert.Len(t, result, tt.expectedLen)
		})
	}
}

func TestDataSourceHandler_GetDataSourceInformationByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMongoRepo := mongodb.NewMockRepository(ctrl)
	mockRedisRepo := redis.NewMockRedisRepository(ctrl)
	orgID := uuid.New()

	mongoSchema := []mongodb.CollectionSchema{
		{
			CollectionName: "collection1",
			Fields: []mongodb.FieldInformation{
				{Name: "field1", DataType: "string"},
				{Name: "field2", DataType: "int"},
			},
		},
	}

	tests := []struct {
		name           string
		dataSourceID   string
		setupService   func() *services.UseCase
		mockSetup      func()
		expectedStatus int
		expectError    bool
	}{
		{
			name:         "Success - Returns MongoDB data source details",
			dataSourceID: "mongo_ds",
			setupService: func() *services.UseCase {
				return &services.UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"mongo_ds": {
							DatabaseType:      pkg.MongoDBType,
							MongoDBName:       "mongo_db",
							MongoDBRepository: mockMongoRepo,
							Initialized:       true,
						},
					},
					RedisRepo: mockRedisRepo,
				}
			},
			mockSetup: func() {
				mockRedisRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return("", nil)
				mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(mongoSchema, nil)
				mockMongoRepo.EXPECT().CloseConnection(gomock.Any()).Return(nil)
				mockRedisRepo.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: fiber.StatusOK,
			expectError:    false,
		},
		{
			name:         "Error - Data source not found",
			dataSourceID: "not_found",
			setupService: func() *services.UseCase {
				return &services.UseCase{
					ExternalDataSources: map[string]pkg.DataSource{},
					RedisRepo:           mockRedisRepo,
				}
			},
			mockSetup: func() {
				mockRedisRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return("", nil)
			},
			expectedStatus: fiber.StatusBadRequest, // ErrMissingDataSource returns ValidationError (400)
			expectError:    true,
		},
		{
			name:         "Error - Database schema retrieval fails",
			dataSourceID: "mongo_ds",
			setupService: func() *services.UseCase {
				return &services.UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"mongo_ds": {
							DatabaseType:      pkg.MongoDBType,
							MongoDBName:       "mongo_db",
							MongoDBRepository: mockMongoRepo,
							Initialized:       true,
						},
					},
					RedisRepo: mockRedisRepo,
				}
			},
			mockSetup: func() {
				mockRedisRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return("", nil)
				mockMongoRepo.EXPECT().GetDatabaseSchema(gomock.Any()).Return(nil, errors.New("db error"))
				mockMongoRepo.EXPECT().CloseConnection(gomock.Any()).Return(nil)
			},
			expectedStatus: fiber.StatusBadRequest, // ErrMissingDataSource returns ValidationError (400)
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			app := setupTestApp()

			handler := &DataSourceHandler{
				Service: tt.setupService(),
			}

			app.Get("/v1/data-sources/:dataSourceId", func(c *fiber.Ctx) error {
				c.Locals("X-Organization-Id", orgID)
				c.SetUserContext(context.Background())
				return handler.GetDataSourceInformationByID(c)
			})

			req := httptest.NewRequest("GET", "/v1/data-sources/"+tt.dataSourceID, nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if !tt.expectError {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var result model.DataSourceDetails
				err = json.Unmarshal(body, &result)
				require.NoError(t, err)

				assert.Equal(t, tt.dataSourceID, result.Id)
			}
		})
	}
}
