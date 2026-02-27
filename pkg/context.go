// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/LerianStudio/lib-commons/v3/commons/log"
	// otel/trace is a structural dependency: this project-level wrapper returns trace.Tracer
	// directly; no lib-commons abstraction wraps the Tracer interface itself.
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
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

	return noop.Tracer{}
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
