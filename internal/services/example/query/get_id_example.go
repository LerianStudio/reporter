package query

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"k8s-golang-addons-boilerplate/internal/services"
	"k8s-golang-addons-boilerplate/pkg"
	"k8s-golang-addons-boilerplate/pkg/constant"
	"k8s-golang-addons-boilerplate/pkg/example_model/model"
	"k8s-golang-addons-boilerplate/pkg/opentelemetry"
	"reflect"
)

// GetExampleByID fetch a new example from the repository
func (ex *ExampleQuery) GetExampleByID(ctx context.Context, id uuid.UUID) (*model.ExampleOutput, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "query.get_example_by_id")
	defer span.End()

	logger.Infof("Retrieving example for id: %s", id.String())

	example, err := ex.ExampleRepo.Find(ctx, id)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get example on repo by id", err)

		logger.Errorf("Error getting example on repo by id: %v", err)

		if errors.Is(err, services.ErrDatabaseItemNotFound) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, reflect.TypeOf(model.Example{}).Name())
		}

		return nil, err
	}

	return example, nil
}
