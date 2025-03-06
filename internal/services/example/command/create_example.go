package query

import (
	"context"
	"k8s-golang-addons-boilerplate/pkg"
	"k8s-golang-addons-boilerplate/pkg/example_model/model"
	"k8s-golang-addons-boilerplate/pkg/opentelemetry"
	"time"
)

// CreateExample create a new example
func (ex *ExampleCommand) CreateExample(ctx context.Context, ei *model.CreateExampleInput) (*model.ExampleOutput, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "services.create_example")
	defer span.End()

	logger.Infof("Creating example")

	example := &model.Example{
		Name:      ei.Name,
		Age:       ei.Age,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := opentelemetry.SetSpanAttributesFromStruct(&span, "example_repository_input", example)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert example repository input to JSON string", err)

		return nil, err
	}

	out, err := ex.ExampleRepo.Create(ctx, example)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to create example", err)
		logger.Errorf("Failed to create example: %v", err)

		return nil, err
	}

	return out, nil
}
