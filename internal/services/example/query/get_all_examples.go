package query

import (
	"context"
	"errors"
	servicesExample "k8s-golang-addons-boilerplate/internal/services"
	"k8s-golang-addons-boilerplate/pkg"
	"k8s-golang-addons-boilerplate/pkg/constant"
	"k8s-golang-addons-boilerplate/pkg/example_model/model"
	"k8s-golang-addons-boilerplate/pkg/net/http"
	"k8s-golang-addons-boilerplate/pkg/opentelemetry"
	"reflect"
)

// GetAllExample fetch all Examples from the repository
func (ex *ExampleQuery) GetAllExample(ctx context.Context, filter http.QueryHeader) ([]*model.ExampleOutput, error) {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "query.get_all_examples")
	defer span.End()

	logger.Infof("Retrieving examples")

	examples, err := ex.ExampleRepo.FindAll(ctx, filter.ToOffsetPagination())
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get examples on repo", err)

		logger.Errorf("Error getting examples on repo: %v", err)

		if errors.Is(err, servicesExample.ErrDatabaseItemNotFound) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, reflect.TypeOf(model.Example{}).Name())
		}

		return nil, err
	}

	return examples, nil
}
