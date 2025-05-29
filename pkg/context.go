package pkg

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type customContextKey string

var CustomContextKey = customContextKey("custom_context")

type CustomContextKeyValue struct {
	Tracer trace.Tracer
	Logger log.Logger
}

// NewLoggerFromContext extract the Logger from "logger" value inside context
func NewLoggerFromContext(ctx context.Context) log.Logger {
	if customContext, ok := ctx.Value(CustomContextKey).(*CustomContextKeyValue); ok &&
		customContext.Logger != nil {
		return customContext.Logger
	}

	return &log.NoneLogger{}
}

// NewTracerFromContext returns a new tracer from the context.
func NewTracerFromContext(ctx context.Context) trace.Tracer {
	if customContext, ok := ctx.Value(CustomContextKey).(*CustomContextKeyValue); ok &&
		customContext.Tracer != nil {
		return customContext.Tracer
	}

	return otel.Tracer("default")
}

// ContextWithLogger returns a context within a Logger in "logger" value.
func ContextWithLogger(ctx context.Context, logger log.Logger) context.Context {
	values, _ := ctx.Value(CustomContextKey).(*CustomContextKeyValue)
	if values == nil {
		values = &CustomContextKeyValue{}
	}

	values.Logger = logger

	return context.WithValue(ctx, CustomContextKey, values)
}

// ContextWithTracer returns a context within a trace.Tracer in "tracer" value.
func ContextWithTracer(ctx context.Context, tracer trace.Tracer) context.Context {
	values, _ := ctx.Value(CustomContextKey).(*CustomContextKeyValue)
	if values == nil {
		values = &CustomContextKeyValue{}
	}

	values.Tracer = tracer

	return context.WithValue(ctx, CustomContextKey, values)
}
