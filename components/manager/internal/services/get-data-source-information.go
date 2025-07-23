package services

import (
	"context"
	"plugin-smart-templates/pkg/model"

	"github.com/LerianStudio/lib-commons/commons"
	"go.opentelemetry.io/otel/attribute"
)

// GetDataSourceInformation getting all data sources information connected on plugin smart templates
func (uc *UseCase) GetDataSourceInformation(ctx context.Context) []*model.DataSourceInformation {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "get_data_source_information")
	defer span.End()

	logger.Infof("Getting data source information")

	var result = make([]*model.DataSourceInformation, 0)

	for key, dataSource := range uc.ExternalDataSources {
		var dataSourceInformation *model.DataSourceInformation

		switch dataSource.DatabaseType {
		case "postgresql":
			span.SetAttributes(
				attribute.String("data_source_type", dataSource.DatabaseType),
			)

			dataSourceInformation = &model.DataSourceInformation{
				Id:           key,
				ExternalName: dataSource.DatabaseConfig.DBName,
				Type:         dataSource.DatabaseType,
			}
		case "mongodb":
			span.SetAttributes(
				attribute.String("data_source_type", dataSource.DatabaseType),
			)

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
