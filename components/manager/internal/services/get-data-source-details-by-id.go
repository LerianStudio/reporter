package services

import (
	"context"
	"encoding/json"
	"plugin-smart-templates/v2/pkg"
	"plugin-smart-templates/v2/pkg/constant"
	"plugin-smart-templates/v2/pkg/model"
	"time"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"go.opentelemetry.io/otel/attribute"
)

// GetDataSourceDetailsByID retrieves the data source information by data source id
func (uc *UseCase) GetDataSourceDetailsByID(ctx context.Context, dataSourceID string) (*model.DataSourceDetails, error) {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)
	reqId := commons.NewHeaderIDFromContext(ctx)

	ctx, span := tracer.Start(ctx, "get_data_source_details_by_id")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.data_source_id", dataSourceID),
	)

	logger.Infof("Retrieving data source details for id %v", dataSourceID)

	cacheKey := constant.DataSourceDetailsKeyPrefix + ":" + dataSourceID
	if cached, ok := uc.getDataSourceDetailsFromCache(ctx, cacheKey); ok {
		logger.Infof("Cache hit for data source details id %v", dataSourceID)
		return cached, nil
	}

	dataSource, ok := uc.ExternalDataSources[dataSourceID]
	if !ok {
		return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", dataSourceID)
	}

	if err := uc.ensureDataSourceConnected(logger, &dataSource); err != nil {
		logger.Errorf("Error initializing database connection, Err: %s", err)
		return nil, err
	}

	var (
		result           *model.DataSourceDetails
		errGetDataSource error
	)

	switch dataSource.DatabaseType {
	case pkg.PostgreSQLType:
		result, errGetDataSource = uc.getDataSourceDetailsOfPostgresDatabase(ctx, logger, dataSourceID, dataSource)

		errClose := dataSource.PostgresRepository.CloseConnection()
		if errClose != nil {
			logger.Errorf("Error to close postgres connection, Err: %s", errClose)
			return nil, errClose
		}
	case pkg.MongoDBType:
		result, errGetDataSource = uc.getDataSourceDetailsOfMongoDBDatabase(ctx, logger, dataSourceID, dataSource)

		errClose := dataSource.MongoDBRepository.CloseConnection(ctx)
		if errClose != nil {
			return nil, errClose
		}
	default:
		return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", dataSourceID)
	}

	if errGetDataSource != nil {
		logger.Errorf("Error to get data source details, Err: %s", errGetDataSource)
		return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", dataSourceID)
	}

	logger.Info("Close the connection to the database.")

	errSet := uc.setDataSourceDetailsToCache(ctx, cacheKey, result)
	if errSet != nil {
		logger.Errorf("Error to set data source details to cache, Err: %s", errSet)
		return nil, errSet
	}

	return result, nil
}

// getDataSourceDetailsFromCache tries to get and unmarshal DataSourceDetails from Redis
func (uc *UseCase) getDataSourceDetailsFromCache(ctx context.Context, cacheKey string) (*model.DataSourceDetails, bool) {
	if uc.RedisRepo == nil {
		return nil, false
	}

	cached, err := uc.RedisRepo.Get(ctx, cacheKey)
	if err != nil || cached == "" {
		return nil, false
	}

	var details model.DataSourceDetails
	if err := json.Unmarshal([]byte(cached), &details); err != nil {
		return nil, false
	}

	return &details, true
}

// setDataSourceDetailsToCache marshals and sets DataSourceDetails in Redis
func (uc *UseCase) setDataSourceDetailsToCache(ctx context.Context, cacheKey string, details *model.DataSourceDetails) error {
	if uc.RedisRepo == nil || details == nil {
		return nil
	}

	if marshaled, err := json.Marshal(details); err == nil {
		if errCache := uc.RedisRepo.Set(ctx, cacheKey, string(marshaled), time.Second*constant.RedisTTL); errCache != nil {
			return errCache
		}
	}

	return nil
}

// ensureDataSourceConnected ensures the data source is initialized/connected
func (uc *UseCase) ensureDataSourceConnected(logger log.Logger, dataSource *pkg.DataSource) error {
	switch dataSource.DatabaseType {
	case pkg.PostgreSQLType:
		if !dataSource.Initialized || !dataSource.DatabaseConfig.Connected {
			return pkg.ConnectToDataSource(dataSource.MongoDBName, dataSource, logger, uc.ExternalDataSources)
		}
	case pkg.MongoDBType:
		if !dataSource.Initialized {
			return pkg.ConnectToDataSource(dataSource.MongoDBName, dataSource, logger, uc.ExternalDataSources)
		}
	}

	return nil
}

// getDataSourceDetailsOfMongoDBDatabase retrieves the data source information of a MongoDB database
func (uc *UseCase) getDataSourceDetailsOfMongoDBDatabase(ctx context.Context, logger log.Logger, dataSourceID string, dataSource pkg.DataSource) (*model.DataSourceDetails, error) {
	schema, err := dataSource.MongoDBRepository.GetDatabaseSchema(ctx)
	if err != nil {
		logger.Errorf("Error get schemas of mongo db: %s", err.Error())

		return nil, err
	}

	tableDetails := make([]model.TableDetails, 0)

	for _, collection := range schema {
		fields := make([]string, 0)
		for _, collectionField := range collection.Fields {
			fields = append(fields, collectionField.Name)
		}

		tableSchema := model.TableDetails{
			Name:   collection.CollectionName,
			Fields: fields,
		}

		tableDetails = append(tableDetails, tableSchema)
	}

	result := &model.DataSourceDetails{
		Id:           dataSourceID,
		ExternalName: dataSource.MongoDBName,
		Type:         dataSource.DatabaseType,
		Tables:       tableDetails,
	}

	return result, nil
}

// getDataSourceDetailsOfPostgresDatabase retrieves the data source information of a PostgresSQL database
func (uc *UseCase) getDataSourceDetailsOfPostgresDatabase(ctx context.Context, logger log.Logger, dataSourceID string, dataSource pkg.DataSource) (*model.DataSourceDetails, error) {
	schemas, err := dataSource.PostgresRepository.GetDatabaseSchema(ctx)
	if err != nil {
		logger.Errorf("Error get schemas of postgres: %s", err.Error())

		return nil, err
	}

	tableDetails := make([]model.TableDetails, 0)

	for _, tableSchema := range schemas {
		fields := make([]string, 0)
		for _, field := range tableSchema.Columns {
			fields = append(fields, field.Name)
		}

		tableDetail := model.TableDetails{
			Name:   tableSchema.TableName,
			Fields: fields,
		}

		tableDetails = append(tableDetails, tableDetail)
	}

	result := &model.DataSourceDetails{
		Id:           dataSourceID,
		ExternalName: dataSource.MongoDBName,
		Type:         dataSource.DatabaseType,
		Tables:       tableDetails,
	}

	return result, nil
}
