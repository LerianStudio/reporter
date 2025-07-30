package in

import (
	"github.com/LerianStudio/lib-commons/commons"
	commonsHttp "github.com/LerianStudio/lib-commons/commons/net/http"
	libOpentelemetry "github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"plugin-smart-templates/components/manager/internal/services"
	_ "plugin-smart-templates/pkg"
	_ "plugin-smart-templates/pkg/model"
	"plugin-smart-templates/pkg/net/http"
)

type DataSourceHandler struct {
	Service *services.UseCase
}

// GetDataSourceInformation retrieves all data sources connected on plugin smart templates
//
//	@Summary		Get all data sources connected on plugin smart templates
//	@Description	Retrieves all data sources connected on plugin with all information from the database
//	@Tags			Data source
//	@Produce		json
//	@Param			Authorization	header		string	false	"The authorization token in the 'Bearer	access_token' format. Only required when auth plugin is enabled."
//	@Success		200				{object}	[]model.DataSourceInformation
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/data-sources [get]
func (ds *DataSourceHandler) GetDataSourceInformation(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)
	reqId := commons.NewHeaderIDFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_data_source")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	logger.Infof("Initiating retrieval data source information")

	dataSourceInfo := ds.Service.GetDataSourceInformation(ctx)

	logger.Infof("Successfully get all data source information")

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
//	@Failure		404				{object}	pkg.HTTPError
//	@Failure		500				{object}	pkg.HTTPError
//	@Router			/v1/data-sources/{dataSourceId} [get]
func (ds *DataSourceHandler) GetDataSourceInformationByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)
	reqId := commons.NewHeaderIDFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_data_source_details_by_id")
	defer span.End()

	dataSourceID := c.Params("dataSourceId")
	logger.Infof("Initiating retrieval data source information with ID: %s", dataSourceID)

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.data_source_id", dataSourceID),
	)

	dataSourceInfo, err := ds.Service.GetDataSourceDetailsByID(ctx, dataSourceID)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to retrieve data source information on query", err)

		logger.Errorf("Failed to retrieve data source information, Error: %s", err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully retrieved all data source information")

	return commonsHttp.OK(c, dataSourceInfo)
}
