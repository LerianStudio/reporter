package in

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	command "k8s-golang-addons-boilerplate/internal/services/example/command"
	"k8s-golang-addons-boilerplate/internal/services/example/query"
	"k8s-golang-addons-boilerplate/pkg"
	"k8s-golang-addons-boilerplate/pkg/constant"
	exampleModel "k8s-golang-addons-boilerplate/pkg/example_model/model"
	"k8s-golang-addons-boilerplate/pkg/net/http"
	"k8s-golang-addons-boilerplate/pkg/opentelemetry"
	"k8s-golang-addons-boilerplate/pkg/postgres"
	"os"
	"reflect"
)

type ExampleHandler struct {
	ExampleQuery   *query.ExampleQuery
	ExampleCommand *command.ExampleCommand
}

// CreateExample is a method that creates an example.
//
//	@Summary		Create an Example
//	@Description	Create an Example with the input payload
//	@Tags			Example
//	@Accept			json
//	@Produce		json
//	@Param			example	body		exampleModel.CreateExampleInput	true	"Example Input"
//	@Success		200					{object}	exampleModel.ExampleOutput
//	@Router			/v1/example [post]
func (ex *ExampleHandler) CreateExample(p any, c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.create_example")
	defer span.End()

	c.SetUserContext(ctx)

	payload := p.(*exampleModel.CreateExampleInput)
	logger.Infof("Request to create an transaction with details: %#v", payload)

	err := opentelemetry.SetSpanAttributesFromStruct(&span, "payload", payload)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert payload to JSON string", err)

		return http.WithError(c, err)
	}

	response, err := ex.ExampleCommand.CreateExample(ctx, payload)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to create example on services", err)

		return http.WithError(c, err)
	}

	logger.Infof("Successfully created Example")

	return http.Created(c, response)
}

// GetExampleByID is a method that retrieves Example information by a given id.
//
//	@Summary		Get an Example by ID
//	@Description	Get an Example with the input ID
//	@Tags			Example
//	@Produce		json
//	@Param			id				path		string	true	"Example ID"
//	@Success		200				{object}	exampleModel.ExampleOutput
//	@Router			/v1/example/{id} [get]
func (ex *ExampleHandler) GetExampleByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_example_by_id")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating retrieval of Example with ID: %s", id.String())

	example, err := ex.ExampleQuery.GetExampleByID(ctx, id)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve example on query", err)

		logger.Errorf("Failed to retrieve Example with ID: %s, Error: %s", id.String(), err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully retrieved Example with ID: %s", id.String())

	return http.OK(c, example)
}

// GetAllExample is a method that retrieves all Example.
//
//	@Summary		Get all Example
//	@Description	Get all Example with the input metadata or without metadata
//	@Tags			Example
//	@Produce		json
//	@Param			limit			query		int		false	"Limit"			default(10)
//	@Param			page			query		int		false	"Page"			default(1)
//	@Param			start_date		query		string	false	"Start Date"	example "2021-01-01"
//	@Param			end_date		query		string	false	"End Date"		example "2021-01-01"
//	@Param			sort_order		query		string	false	"Sort Order"	Enums(asc,desc)
//	@Success		200				{object}	postgres.Pagination{items=[]exampleModel.ExampleOutput,page=int,limit=int}
//	@Router			/v1/example [get]
func (ex *ExampleHandler) GetAllExample(c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.get_all_examples")
	defer span.End()

	headerParams, err := http.ValidateParameters(c.Queries())
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to validate query parameters", err)

		logger.Errorf("Failed to validate query parameters, Error: %s", err.Error())

		return http.WithError(c, err)
	}

	pagination := postgres.Pagination{
		Limit:     headerParams.Limit,
		Page:      headerParams.Page,
		SortOrder: headerParams.SortOrder,
		StartDate: headerParams.StartDate,
		EndDate:   headerParams.EndDate,
	}

	logger.Infof("Initiating retrieval of all Examples ")
	logger.Infof("Headers values: %v", headerParams)

	examples, err := ex.ExampleQuery.GetAllExample(ctx, *headerParams)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve all examples", err)

		logger.Errorf("Failed to retrieve all Examples, Error: %s", err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully retrieved all Examples")

	pagination.SetItems(examples)

	return http.OK(c, pagination)
}

// UpdateExample is a method that updates Example information.
//
//	@Summary		Update an Example
//	@Description	Update an Example with the input payload
//	@Tags			Example
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string							true	"Example ID"
//	@Param			organization	body		exampleModel.UpdateExampleInput	true	"Example Input"
//	@Success		200				{object}	exampleModel.ExampleOutput
//	@Router			/v1/example/{id} [patch]
func (ex *ExampleHandler) UpdateExample(p any, c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.update_example")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating update of Example with ID: %s", id.String())

	payload := p.(*exampleModel.UpdateExampleInput)
	logger.Infof("Request to update an example with details: %#v", payload)

	err := opentelemetry.SetSpanAttributesFromStruct(&span, "payload", payload)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert payload to JSON string", err)

		return http.WithError(c, err)
	}

	_, err = ex.ExampleCommand.UpdateExampleByID(ctx, id, payload)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to update example", err)

		logger.Errorf("Failed to update Example with ID: %s, Error: %s", id.String(), err.Error())

		return http.WithError(c, err)
	}

	example, err := ex.ExampleQuery.GetExampleByID(ctx, id)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve example on query", err)

		logger.Errorf("Failed to retrieve Example with ID: %s, Error: %s", id.String(), err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully updated Example with ID: %s", id.String())

	return http.OK(c, example)
}

// DeleteExampleByID is a method that removes Example information by a given id.
//
//	@Summary		Delete an Example by ID
//	@Description	Delete an Example with the input ID
//	@Tags			Example
//	@Param			id				path	string	true	"Example ID"
//	@Success		204
//	@Router			/v1/example/{id} [delete]
func (ex *ExampleHandler) DeleteExampleByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.delete_example_by_id")
	defer span.End()

	id := c.Locals("id").(uuid.UUID)
	logger.Infof("Initiating removal of Example with ID: %s", id.String())

	if os.Getenv("ENV_NAME") == "production" {
		opentelemetry.HandleSpanError(&span, "Failed to remove example: "+constant.ErrActionNotPermitted.Error(), constant.ErrActionNotPermitted)

		logger.Errorf("Failed to remove Example with ID: %s in ", id.String())

		err := pkg.ValidateBusinessError(constant.ErrActionNotPermitted, reflect.TypeOf(exampleModel.Example{}).Name())

		return http.WithError(c, err)
	}

	if err := ex.ExampleCommand.DeleteExampleByID(ctx, id); err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to remove example on database", err)

		logger.Errorf("Failed to remove Example with ID: %s, Error: %s", id.String(), err.Error())

		return http.WithError(c, err)
	}

	logger.Infof("Successfully removed Example with ID: %s", id.String())

	return http.NoContent(c)
}
