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

// UpdateExampleByID update an example from the repository.
func (ex *ExampleCommand) UpdateExampleByID(ctx context.Context, id uuid.UUID, uex *model.UpdateExampleInput) (*model.ExampleOutput, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "command.update_example_by_id")
	defer span.End()

	logger.Infof("Trying to update example: %v", uex)

	example := &model.Example{
		Name: uex.Name,
		Age:  uex.Age,
	}

	organizationUpdated, err := ex.ExampleRepo.Update(ctx, id, example)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to update organization on repo by id", err)

		logger.Errorf("Error updating organization on repo by id: %v", err)

		if errors.Is(err, services.ErrDatabaseItemNotFound) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, reflect.TypeOf(model.Example{}).Name())
		}

		return nil, err
	}

	return organizationUpdated, nil
}
