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

// DeleteExampleByID fetch a new example from the repository
func (ex *ExampleCommand) DeleteExampleByID(ctx context.Context, id uuid.UUID) error {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "example_command.delete_example_by_id")
	defer span.End()

	logger.Infof("Remove example for id: %s", id)

	if err := ex.ExampleRepo.Delete(ctx, id); err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to delete example on repo by id", err)

		logger.Errorf("Error deleting example on repo by id: %v", err)

		if errors.Is(err, services.ErrDatabaseItemNotFound) {
			return pkg.ValidateBusinessError(constant.ErrEntityNotFound, reflect.TypeOf(model.Example{}).Name())
		}

		return err
	}

	return nil
}
