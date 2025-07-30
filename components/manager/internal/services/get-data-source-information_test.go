package services

import (
	"context"
	"testing"

	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/model"
	"plugin-smart-templates/pkg/mongodb"
	"plugin-smart-templates/pkg/postgres"

	"github.com/stretchr/testify/assert"
)

func Test_GetDataSourceInformation(t *testing.T) {
	ctx := context.Background()

	pgConfig := &postgres.Connection{DBName: "pg_db"}

	tests := []struct {
		name         string
		setupSvc     func() *UseCase
		expectResult []*model.DataSourceDetails
	}{
		{
			name: "Success - Both MongoDB and PostgreSQL present",
			setupSvc: func() *UseCase {
				return &UseCase{
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
			expectResult: []*model.DataSourceDetails{
				{
					Id:           "mongo_ds",
					ExternalName: "mongo_db",
					Type:         pkg.MongoDBType,
					Tables:       []model.TableDetails{},
				},
				{
					Id:           "pg_ds",
					ExternalName: "pg_db",
					Type:         pkg.PostgreSQLType,
					Tables:       []model.TableDetails{},
				},
			},
		},
		{
			name: "Success - Only MongoDB present",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"mongo_ds": {
							DatabaseType:      pkg.MongoDBType,
							MongoDBName:       "mongo_db",
							MongoDBRepository: mongodb.NewMockRepository(nil),
						},
					},
				}
			},
			expectResult: []*model.DataSourceDetails{
				{
					Id:           "mongo_ds",
					ExternalName: "mongo_db",
					Type:         pkg.MongoDBType,
					Tables:       []model.TableDetails{},
				},
			},
		},
		{
			name: "Success - Only PostgreSQL present",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"pg_ds": {
							DatabaseType:       pkg.PostgreSQLType,
							DatabaseConfig:     pgConfig,
							PostgresRepository: postgres.NewMockRepository(nil),
						},
					},
				}
			},
			expectResult: []*model.DataSourceDetails{
				{
					Id:           "pg_ds",
					ExternalName: "pg_db",
					Type:         pkg.PostgreSQLType,
					Tables:       []model.TableDetails{},
				},
			},
		},
		{
			name: "Success - No data sources",
			setupSvc: func() *UseCase {
				return &UseCase{ExternalDataSources: map[string]pkg.DataSource{}}
			},
			expectResult: []*model.DataSourceDetails{},
		},
		{
			name: "Unknown type - should return empty result",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"unknown_ds": {
							DatabaseType: "unknown",
						},
					},
				}
			},
			expectResult: []*model.DataSourceDetails{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setupSvc()
			result := svc.GetDataSourceInformation(ctx)
			assert.ElementsMatch(t, tt.expectResult, result)
		})
	}
}
