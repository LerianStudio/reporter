package services

import (
	"context"
	"plugin-smart-templates/pkg/model"

	"github.com/LerianStudio/lib-commons/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
)

// GetDataSourceInformation getting all data sources information connected on plugin smart templates
func (uc *UseCase) GetDataSourceInformation(ctx context.Context) []*model.DataSourceInformation {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)
	reqId := commons.NewHeaderIDFromContext(ctx)

	_, span := tracer.Start(ctx, "get_data_source_information")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err := libOpentelemetry.SetSpanAttributesFromStructWithObfuscation(&span, "app.request.external_data_sources", uc.ExternalDataSources)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert external data sources to JSON string", err)
	}

	logger.Infof("Getting data source information")

	var result = make([]*model.DataSourceInformation, 0)

	for key, dataSource := range uc.ExternalDataSources {
		var dataSourceInformation *model.DataSourceInformation

		switch dataSource.DatabaseType {
		case "postgresql":
			dataSourceInformation = &model.DataSourceInformation{
				Id:           key,
				ExternalName: dataSource.DatabaseConfig.DBName,
				Type:         dataSource.DatabaseType,
			}
		case "mongodb":
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
