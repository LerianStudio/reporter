package in

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/go-playground/validator.v9"
	command "k8s-golang-addons-boilerplate/internal/services/example/command"
	"k8s-golang-addons-boilerplate/internal/services/example/query"
	"k8s-golang-addons-boilerplate/pkg"
	"k8s-golang-addons-boilerplate/pkg/example_model/model"
	"k8s-golang-addons-boilerplate/pkg/net/http"
	"k8s-golang-addons-boilerplate/pkg/opentelemetry"
	proto "k8s-golang-addons-boilerplate/pkg/proto/example"
)

// ExampleProto struct contains an example query for managing example
type ExampleProto struct {
	ExampleQuery   *query.ExampleQuery
	ExampleCommand *command.ExampleCommand
	Validator      *validator.Validate
	proto.UnimplementedExampleServer
}

// CreateExample is a method that creates an Example.
func (exp *ExampleProto) CreateExample(ctx context.Context, input *proto.CreateExampleRequest) (*proto.ExampleResponse, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.CreateExample")
	defer span.End()

	request := &model.CreateExampleInput{
		Name: input.Name,
		Age:  int(input.Age),
	}

	// Validate struct
	if err := exp.Validator.Struct(request); err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to validate an Example", err)

		logger.Errorf("Validation failed: %v", err)

		return nil, status.Errorf(codes.InvalidArgument, "Validation failed: %v", err)
	}

	// Create example
	example, err := exp.ExampleCommand.CreateExample(ctx, request)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to create an Example", err)

		logger.Errorf("Failed to create an Example, Error: %s", err.Error())

		// You can create a function to map the errors
		return nil, status.Errorf(codes.InvalidArgument, "Failed to create an Example, Error: %v", err)
	}

	response := &proto.ExampleResponse{
		Id:        example.ID,
		Name:      example.Name,
		Age:       int64(example.Age),
		CratedAt:  example.CreatedAt.String(),
		UpdatedAt: example.UpdatedAt.String(),
	}

	return response, nil
}

// GetExampleByID is a method that retrieves Example information by id.
func (exp *ExampleProto) GetExampleByID(ctx context.Context, id *proto.ExampleID) (*proto.ExampleResponse, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.GetExampleByID")
	defer span.End()

	exampleUUID, err := uuid.Parse(id.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Error to parse example ID to UUID: %v", err)
	}

	example, err := exp.ExampleQuery.GetExampleByID(ctx, exampleUUID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve Example by ids for grpc", err)

		logger.Errorf("Failed to retrieve Example by id for grpc, Error: %s", err.Error())

		return nil, status.Errorf(codes.InvalidArgument, "Failed to retrieve Example by id for grpc, Error: %v", err)
	}

	response := &proto.ExampleResponse{
		Id:        example.ID,
		Name:      example.Name,
		Age:       int64(example.Age),
		CratedAt:  example.CreatedAt.String(),
		UpdatedAt: example.UpdatedAt.String(),
	}

	return response, nil
}

// GetAllExamples is a method that retrieves all Examples
func (exp *ExampleProto) GetAllExamples(ctx context.Context, req *proto.ExampleQuery) (*proto.PaginationExample, error) {
	// Creating logger and tracer
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	// Start span for tracing
	ctx, span := tracer.Start(ctx, "handler.GetAllExamples")
	defer span.End()

	headerParams, err := http.ValidateParameters(req.Parameters)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to validate query parameters", err)

		logger.Errorf("Failed to validate query parameters, Error: %s", err.Error())

		return nil, status.Errorf(codes.InvalidArgument, "Failed to validate query parameters, Error: %v", err)
	}

	// Logging the query parameters
	logger.Infof("Retrieving Examples with parameters: %+v", headerParams)

	// Fetching the examples using the QueryExample
	examples, err := exp.ExampleQuery.GetAllExample(ctx, *headerParams)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve all examples", err)

		logger.Errorf("Failed to retrieve all Examples, Error: %s", err.Error())

		return nil, status.Errorf(codes.InvalidArgument, "Failed to retrieve all Examples, Error: %v", err)
	}

	logger.Infof("Successfully retrieved Examples")

	limit, errSafeIntLimit := pkg.SafeIntToInt32(headerParams.Limit)
	if errSafeIntLimit != nil {
		opentelemetry.HandleSpanError(&span, "Value out of range for int32", errSafeIntLimit)

		logger.Errorf("Value out of range for int32, Error: %s", errSafeIntLimit.Error())

		return nil, status.Errorf(codes.InvalidArgument, "Value out of range for int32, Error: %v", errSafeIntLimit)
	}

	page, errSafeIntPage := pkg.SafeIntToInt32(headerParams.Page)
	if errSafeIntPage != nil {
		opentelemetry.HandleSpanError(&span, "Value out of range for int32", errSafeIntPage)

		logger.Errorf("Value out of range for int32, Error: %s", errSafeIntPage.Error())

		return nil, status.Errorf(codes.InvalidArgument, "Value out of range for int32, Error: %v", errSafeIntPage)
	}

	// Creating the response
	response := &proto.PaginationExample{
		Limit:     limit,
		Page:      page,
		SortOrder: headerParams.SortOrder,
		StartDate: headerParams.StartDate.String(),
		EndDate:   headerParams.EndDate.String(),
	}

	// Prevents relocations while you add elements with attachment
	exampleResponses := make([]*proto.ExampleResponse, 0, len(examples))
	for _, example := range examples {
		exampleResponses = append(exampleResponses, &proto.ExampleResponse{
			Id:        example.ID,
			Name:      example.Name,
			Age:       int64(example.Age),
			CratedAt:  example.CreatedAt.String(),
			UpdatedAt: example.UpdatedAt.String(),
		})
	}

	response.Items = exampleResponses

	return response, nil
}

// UpdateExample is a method that update Example by ID.
func (exp *ExampleProto) UpdateExample(ctx context.Context, update *proto.ExampleRequest) (*proto.ExampleResponse, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.UpdateExample")
	defer span.End()

	if pkg.IsNilOrEmpty(&update.Id) {
		opentelemetry.HandleSpanError(&span, "Failed to update Example because id is empty", nil)

		logger.Errorf("Failed to update Example because id is empty")

		return nil, status.Errorf(codes.InvalidArgument, "Failed to update Example because id is empty")
	}

	updateExIn := &model.UpdateExampleInput{
		Name: update.Name,
		Age:  int(update.Age),
	}

	exampleUUID, err := uuid.Parse(update.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Error to parse example ID to UUID: %v", err)
	}

	_, err = exp.ExampleCommand.UpdateExampleByID(ctx, exampleUUID, updateExIn)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to update balance in Example by id", err)

		logger.Errorf("Failed to update Example by id, Error: %s", err.Error())

		return nil, status.Errorf(codes.InvalidArgument, "Failed to update Example by id, Error: %v", err)
	}

	example, err := exp.ExampleQuery.GetExampleByID(ctx, exampleUUID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to retrieve Example by ids for grpc", err)

		logger.Errorf("Failed to update Example by id, Error: %s", err.Error())

		return nil, status.Errorf(codes.InvalidArgument, "Failed to update Example by id, Error: %v", err)
	}

	response := &proto.ExampleResponse{
		Id:        example.ID,
		Name:      example.Name,
		Age:       int64(example.Age),
		CratedAt:  example.CreatedAt.String(),
		UpdatedAt: example.UpdatedAt.String(),
	}

	return response, nil
}

// DeleteExampleByID is a method that delete Example information by id.
func (exp *ExampleProto) DeleteExampleByID(ctx context.Context, id *proto.ExampleID) (*proto.DeleteExampleResponse, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.DeleteExampleByID")
	defer span.End()

	exampleUUID, err := uuid.Parse(id.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Error to parse example ID to UUID: %v", err)
	}

	errDeleted := exp.ExampleCommand.DeleteExampleByID(ctx, exampleUUID)
	if errDeleted != nil {
		opentelemetry.HandleSpanError(&span, "Failed to delete Example by ids for grpc", errDeleted)

		logger.Errorf("Failed to delete Example by id for grpc, Error: %s", errDeleted.Error())

		return nil, status.Errorf(codes.InvalidArgument, "Failed to delete Example by id for grpc, Error: %v", errDeleted)
	}

	response := &proto.DeleteExampleResponse{
		Code:    int32(201),
		Message: "Example Deleted",
	}

	return response, nil
}
