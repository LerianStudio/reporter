package services

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/log"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/model"
)

// GetDataSourceDetailsByID retrieves the data source information by data source id
func (uc *UseCase) GetDataSourceDetailsByID(ctx context.Context, dataSourceID string) (*model.DataSourceDetails, error) {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "get_data_source_details_by_id")
	defer span.End()

	logger.Infof("Retrieving data source details for id %v", dataSourceID)

	dataSource, ok := uc.ExternalDataSources[dataSourceID]
	if !ok {
		return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", dataSourceID)
	}

	switch dataSource.DatabaseType {
	case pkg.PostgreSQLType:
		if !dataSource.Initialized || !dataSource.DatabaseConfig.Connected {
			if err := pkg.ConnectToDataSource(dataSource.MongoDBName, &dataSource, logger, uc.ExternalDataSources); err != nil {
				logger.Errorf("Error initializing database connection, Err: %s", err)
				return nil, err
			}
		}

		result, errGetDataSource := uc.getDataSourceDetailsOfPostgresDatabase(ctx, logger, dataSourceID, dataSource)
		if errGetDataSource != nil {
			logger.Errorf("Error to get data source details of postgres database, Err: %s", errGetDataSource)
			return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", dataSourceID)
		}

		return result, nil
	case pkg.MongoDBType:
		if !dataSource.Initialized {
			if err := pkg.ConnectToDataSource(dataSource.MongoDBName, &dataSource, logger, uc.ExternalDataSources); err != nil {
				logger.Errorf("Error initializing database connection, Err: %s", err)
				return nil, err
			}
		}

		result, errGetDataSource := uc.getDataSourceDetailsOfMongoDBDatabase(ctx, logger, dataSourceID, dataSource)
		if errGetDataSource != nil {
			logger.Errorf("Error to get data source details of mongoDB database, Err: %s", errGetDataSource)
			return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", dataSourceID)
		}

		return result, nil
	default:
		return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", dataSourceID)
	}
}

// getDataSourceDetailsOfMongoDBDatabase retrieves the data source information of a MongoDB database
func (uc *UseCase) getDataSourceDetailsOfMongoDBDatabase(ctx context.Context, logger log.Logger, dataSourceID string, dataSource pkg.DataSource) (*model.DataSourceDetails, error) {
	schema, err := dataSource.MongoDBRepository.GetDatabaseSchema(ctx)
	if err != nil {
		logger.Errorf("Error get schemas of mongo db: %s", err.Error())

		return nil, err
	}

	var tableDetails []model.TableDetails
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

	var tableDetails []model.TableDetails
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
