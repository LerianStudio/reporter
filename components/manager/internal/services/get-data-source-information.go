package services

import (
	"context"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/model"

	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/log"
	"go.opentelemetry.io/otel/attribute"
)

// GetDataSourceInformation getting all data sources information with detailed schema connected on plugin smart templates
func (uc *UseCase) GetDataSourceInformation(ctx context.Context) []*model.DataSourceDetails {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "get_data_source_information")
	defer span.End()

	logger.Infof("Retrieving data source details for all entries")

	var result = make([]*model.DataSourceDetails, 0)

	// Add defensive check for nil ExternalDataSources
	if uc.ExternalDataSources == nil {
		logger.Warnf("ExternalDataSources is nil, returning empty result")
		return result
	}

	// Process each data source with panic recovery
	for dataSourceID, dataSource := range uc.ExternalDataSources {
		func() {
			// Add panic recovery for each data source
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("PANIC recovered while processing data source %s: %v", dataSourceID, r)
				}
			}()

			// Skip data sources with empty or invalid IDs
			if dataSourceID == "" {
				logger.Warnf("Skipping data source with empty ID")
				return
			}

			span.SetAttributes(
				attribute.String("data_source_id", dataSourceID),
				attribute.String("data_source_type", dataSource.DatabaseType),
			)

			// Try to get from cache first
			cacheKey := constant.DataSourceDetailsKeyPrefix + ":" + dataSourceID

			if cached, ok := uc.getDataSourceDetailsFromCache(ctx, cacheKey); ok {
				logger.Infof("Cache hit for data source details id %v", dataSourceID)
				result = append(result, cached)
				return
			}

			logger.Infof("Cache miss for %s, fetching from database", dataSourceID)

			// Get detailed information for this data source
			details, err := uc.getDataSourceDetailsWithErrorHandling(ctx, logger, dataSourceID, dataSource)
			if err != nil {
				logger.Errorf("Failed to get details for data source %s: %v", dataSourceID, err)
				return
			}

			// Cache the result
			if errSet := uc.setDataSourceDetailsToCache(ctx, cacheKey, details); errSet != nil {
				logger.Errorf("Error setting data source details to cache for %s: %v", dataSourceID, errSet)
			}

			result = append(result, details)
		}()
	}

	logger.Infof("Successfully retrieved %d detailed data source details", len(result))
	return result
}

// getDataSourceDetailsWithErrorHandling safely retrieves data source details with proper connection management
func (uc *UseCase) getDataSourceDetailsWithErrorHandling(ctx context.Context, logger log.Logger, dataSourceID string, dataSource pkg.DataSource) (result *model.DataSourceDetails, err error) {
	// Add panic recovery for this method
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("PANIC in getDataSourceDetailsWithErrorHandling for %s: %v", dataSourceID, r)
			err = pkg.ValidateBusinessError(constant.ErrMissingDataSource, "Panic occurred while processing data source", dataSourceID)
		}
	}()

	// Validate input parameters
	if dataSourceID == "" {
		return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "Data source ID cannot be empty", "")
	}

	if dataSource.DatabaseType == "" {
		logger.Errorf("Data source %s has empty database type", dataSourceID)
		return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "Database type cannot be empty", dataSourceID)
	}

	// Ensure the data source is connected
	if err := uc.ensureDataSourceConnected(logger, &dataSource); err != nil {
		logger.Errorf("Error initializing database connection for %s: %s", dataSourceID, err)
		return nil, err
	}

	var (
		errGetDataSource error
	)

	// Get detailed schema information based on database type
	switch dataSource.DatabaseType {
	case pkg.PostgreSQLType:

		// Validate PostgreSQL configuration
		if dataSource.PostgresRepository == nil {
			logger.Errorf("PostgreSQL repository is nil for data source %s", dataSourceID)
			return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "PostgreSQL repository not initialized", dataSourceID)
		}
		if dataSource.DatabaseConfig == nil {
			logger.Errorf("Database config is nil for PostgreSQL data source %s", dataSourceID)
			return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "Database config not initialized", dataSourceID)
		}

		result, errGetDataSource = uc.getDataSourceDetailsOfPostgresDatabase(ctx, logger, dataSourceID, dataSource)

		// Always close PostgreSQL connection
		if dataSource.PostgresRepository != nil {
			if errClose := dataSource.PostgresRepository.CloseConnection(); errClose != nil {
				logger.Errorf("Error closing postgres connection for %s: %s", dataSourceID, errClose)
			}
		}

	case pkg.MongoDBType:
		// Validate MongoDB configuration
		if dataSource.MongoDBRepository == nil {
			logger.Errorf("MongoDB repository is nil for data source %s", dataSourceID)
			return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "MongoDB repository not initialized", dataSourceID)
		}
		if dataSource.MongoDBName == "" {
			logger.Errorf("MongoDB name is empty for data source %s", dataSourceID)
			return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "MongoDB name not configured", dataSourceID)
		}

		result, errGetDataSource = uc.getDataSourceDetailsOfMongoDBDatabase(ctx, logger, dataSourceID, dataSource)

		// Always close MongoDB connection
		if dataSource.MongoDBRepository != nil {
			if errClose := dataSource.MongoDBRepository.CloseConnection(ctx); errClose != nil {
				logger.Errorf("Error closing mongodb connection for %s: %s", dataSourceID, errClose)
			}
		}

	default:
		logger.Errorf("Unsupported database type %s for data source %s", dataSource.DatabaseType, dataSourceID)
		return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "Unsupported database type", dataSourceID)
	}

	if errGetDataSource != nil {
		logger.Errorf("Error getting data source details for %s: %s", dataSourceID, errGetDataSource)
		return nil, errGetDataSource
	}

	if result == nil {
		logger.Errorf("Result is nil for data source %s", dataSourceID)
		return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "Failed to retrieve data source details", dataSourceID)
	}

	return result, nil
}
