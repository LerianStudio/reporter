// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"errors"

	"github.com/LerianStudio/reporter/components/manager/internal/services"
	_ "github.com/LerianStudio/reporter/pkg"
	_ "github.com/LerianStudio/reporter/pkg/model"
	"github.com/LerianStudio/reporter/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v2/commons"
	commonsHttp "github.com/LerianStudio/lib-commons/v2/commons/net/http"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// DataSourceHandler handles HTTP requests for data source operations.
type DataSourceHandler struct {
	service *services.UseCase
}

// NewDataSourceHandler creates a new DataSourceHandler with the given service dependency.
// It returns an error if service is nil.
func NewDataSourceHandler(service *services.UseCase) (*DataSourceHandler, error) {
	if service == nil {
		return nil, errors.New("service must not be nil for DataSourceHandler")
	}

	return &DataSourceHandler{service: service}, nil
}

// GetDataSourceInformation retrieves all data sources connected on reporter.
//
//	@Summary		Get all data sources connected on reporter
//	@Description	Retrieves all data sources connected on plugin with all information from the database
//	@Tags			Data source
//	@Produce		json
//	@Param			Authorization	header		string	false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Success		200				{object}	[]model.DataSourceInformation
//	@Failure		401				{object}	pkg.HTTPError
//	@Failure		403				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/data-sources [get]
func (ds *DataSourceHandler) GetDataSourceInformation(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.data_source.get")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	logger.Infof("Initiating retrieval data source information")

	dataSourceInfo := ds.service.GetDataSourceInformation(ctx)

	logger.Infof("Successfully get all data source information.")

	return commonsHttp.OK(c, dataSourceInfo)
}

// GetDataSourceInformationByID retrieves a data sources information with data source id passed
//
//	@Summary		Get a data sources information
//	@Description	Retrieves a data sources information with data source id passed
//	@Tags			Data source
//	@Produce		json
//	@Param			Authorization	header		string	false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Param			dataSourceId	path		string	true	"Data source ID"
//	@Success		200				{object}	model.DataSourceDetails
//	@Failure		400				{object}	pkg.HTTPError
//	@Failure		401				{object}	pkg.HTTPError
//	@Failure		403				{object}	pkg.HTTPError
//	@Failure		404				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/data-sources/{dataSourceId} [get]
func (ds *DataSourceHandler) GetDataSourceInformationByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.data_source.get_details_by_id")
	defer span.End()

	dataSourceID := c.Locals("dataSourceId").(uuid.UUID).String()

	logger.Infof("Initiating retrieval data source information with ID: %s", dataSourceID)

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.data_source_id", dataSourceID),
	)

	dataSourceInfo, err := ds.service.GetDataSourceDetailsByID(ctx, dataSourceID)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to retrieve data source information on query", err)

		logger.Errorf("Failed to retrieve data source information, Error: %s", err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully retrieved all data source information")

	return commonsHttp.OK(c, dataSourceInfo)
}
