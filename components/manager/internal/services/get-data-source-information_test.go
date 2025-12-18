package services

import (
	"context"
	"testing"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/model"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb"
	"github.com/LerianStudio/reporter/v4/pkg/postgres"

	"github.com/stretchr/testify/assert"
)

func Test_GetDataSourceInformation(t *testing.T) {
	ctx := context.Background()

	// Register datasource IDs for testing
	pkg.ResetRegisteredDataSourceIDsForTesting()
	pkg.RegisterDataSourceIDsForTesting([]string{"mongo_ds", "pg_ds"})

	pgConfig := &postgres.Connection{DBName: "pg_db"}

	tests := []struct {
		name         string
		setupSvc     func() *UseCase
		expectResult []*model.DataSourceInformation
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
			expectResult: []*model.DataSourceInformation{
				{
					Id:           "mongo_ds",
					ExternalName: "mongo_db",
					Type:         pkg.MongoDBType,
				},
				{
					Id:           "pg_ds",
					ExternalName: "pg_db",
					Type:         pkg.PostgreSQLType,
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
			expectResult: []*model.DataSourceInformation{
				{
					Id:           "mongo_ds",
					ExternalName: "mongo_db",
					Type:         pkg.MongoDBType,
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
			expectResult: []*model.DataSourceInformation{
				{
					Id:           "pg_ds",
					ExternalName: "pg_db",
					Type:         pkg.PostgreSQLType,
				},
			},
		},
		{
			name: "Success - No data sources",
			setupSvc: func() *UseCase {
				return &UseCase{ExternalDataSources: map[string]pkg.DataSource{}}
			},
			expectResult: []*model.DataSourceInformation{},
		},
		{
			name: "Unknown type - should return empty slice",
			setupSvc: func() *UseCase {
				return &UseCase{
					ExternalDataSources: map[string]pkg.DataSource{
						"unknown_ds": {
							DatabaseType: "unknown",
						},
					},
				}
			},
			expectResult: []*model.DataSourceInformation{},
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
