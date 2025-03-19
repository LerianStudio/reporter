package services

import (
	"context"
	"plugin-template-engine/pkg"
)

// CreateTemplate create a new template
func (ex *UseCase) CreateTemplate(ctx context.Context) error {
	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "services.create_example")
	defer span.End()

	logger.Infof("Creating template")

	return nil
}
