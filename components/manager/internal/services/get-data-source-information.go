package services

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons"
	"plugin-smart-templates/pkg/model"
)

// GetDataSourceInformation getting all data sources information connected on plugin smart templates
func (uc *UseCase) GetDataSourceInformation(ctx context.Context) []*model.DataSourceInformation {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "get_data_source_information")
	defer span.End()

	logger.Infof("Getting data source information")

	var result = make([]*model.DataSourceInformation, 0)
	for key, dataSource := range uc.ExternalDataSources {
		var dataSourceInformation *model.DataSourceInformation
		if dataSource.DatabaseType == "postgresql" {
			dataSourceInformation = &model.DataSourceInformation{
				Id:           key,
				ExternalName: dataSource.DatabaseConfig.DBName,
				Type:         dataSource.DatabaseType,
			}
		} else if dataSource.DatabaseType == "mongodb" {
			dataSourceInformation = &model.DataSourceInformation{
				Id:           key,
				ExternalName: dataSource.MongoDBName,
				Type:         dataSource.DatabaseType,
			}
		}

		result = append(result, dataSourceInformation)
	}

	return result
}
