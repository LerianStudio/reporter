// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"strings"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/model"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
)

// GetDataSourceInformation getting all data sources information connected on reporter
func (uc *UseCase) GetDataSourceInformation(ctx context.Context) []*model.DataSourceInformation {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.datasource.get_information")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	allDataSources := uc.ExternalDataSources.GetAll()

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.external_data_sources", allDataSources)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert external data sources to JSON string", err)
	}

	logger.Infof("Getting data source information")

	result := make([]*model.DataSourceInformation, 0)

	for key, dataSource := range allDataSources {
		if !pkg.IsValidDataSourceID(key) {
			logger.Warnf("Skipping datasource '%s' from listing - not in immutable registry (possible corruption)", key)
			continue
		}

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

		if dataSourceInformation != nil && strings.TrimSpace(dataSourceInformation.Id) != "" {
			// Add note for plugin_crm about field filtering
			if key == "plugin_crm" {
				logger.Infof("Note: plugin_crm data source filters out encrypted fields and only shows non-encrypted fields and search fields for security")
			}

			result = append(result, dataSourceInformation)
		}
	}

	return result
}
