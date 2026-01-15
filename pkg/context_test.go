package pkg

import (
	"context"
	"testing"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
)

func TestNewLoggerFromContext_WithLogger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	ctx := context.WithValue(context.Background(), CustomContextKey, &CustomContextKeyValue{
		Logger: mockLogger,
	})

	result := NewLoggerFromContext(ctx)

	assert.Equal(t, mockLogger, result)
}

func TestNewLoggerFromContext_WithoutLogger(t *testing.T) {
	ctx := context.Background()

	result := NewLoggerFromContext(ctx)

	assert.IsType(t, &log.NoneLogger{}, result)
}

func TestNewLoggerFromContext_NilLogger(t *testing.T) {
	ctx := context.WithValue(context.Background(), CustomContextKey, &CustomContextKeyValue{
		Logger: nil,
	})

	result := NewLoggerFromContext(ctx)

	assert.IsType(t, &log.NoneLogger{}, result)
}

func TestNewTracerFromContext_WithTracer(t *testing.T) {
	tracer := otel.Tracer("test-tracer")

	ctx := context.WithValue(context.Background(), CustomContextKey, &CustomContextKeyValue{
		Tracer: tracer,
	})

	result := NewTracerFromContext(ctx)

	assert.NotNil(t, result)
}

func TestNewTracerFromContext_WithoutTracer(t *testing.T) {
	ctx := context.Background()

	result := NewTracerFromContext(ctx)

	assert.NotNil(t, result) // Returns default tracer
}

func TestContextWithLogger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	ctx := ContextWithLogger(context.Background(), mockLogger)

	// Verify logger was set
	result := NewLoggerFromContext(ctx)
	assert.Equal(t, mockLogger, result)
}

func TestContextWithLogger_ExistingValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	tracer := otel.Tracer("test-tracer")

	// Set tracer first
	ctx := context.WithValue(context.Background(), CustomContextKey, &CustomContextKeyValue{
		Tracer: tracer,
	})

	// Add logger
	ctx = ContextWithLogger(ctx, mockLogger)

	// Both should be set
	resultLogger := NewLoggerFromContext(ctx)
	resultTracer := NewTracerFromContext(ctx)

	assert.Equal(t, mockLogger, resultLogger)
	assert.NotNil(t, resultTracer)
}

func TestContextWithTracer(t *testing.T) {
	tracer := otel.Tracer("test-tracer")

	ctx := ContextWithTracer(context.Background(), tracer)

	// Verify tracer was set
	result := NewTracerFromContext(ctx)
	assert.NotNil(t, result)
}

func TestContextWithTracer_ExistingValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	tracer := otel.Tracer("test-tracer")

	// Set logger first
	ctx := context.WithValue(context.Background(), CustomContextKey, &CustomContextKeyValue{
		Logger: mockLogger,
	})

	// Add tracer
	ctx = ContextWithTracer(ctx, tracer)

	// Both should be set
	resultLogger := NewLoggerFromContext(ctx)
	resultTracer := NewTracerFromContext(ctx)

	assert.Equal(t, mockLogger, resultLogger)
	assert.NotNil(t, resultTracer)
}
