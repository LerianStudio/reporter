// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pkg

import (
	"context"
	"testing"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestNewLoggerFromContext(t *testing.T) {
	t.Parallel()

	// Create a shared logger instance for the "with logger" test case
	sharedLogger := &log.NoneLogger{}

	tests := []struct {
		name         string
		setupCtx     func() context.Context
		expectSame   log.Logger // If non-nil, expect this exact logger instance
		expectIsNone bool       // If true, expect result to be *log.NoneLogger type
	}{
		{
			name: "Context with logger",
			setupCtx: func() context.Context {
				return ContextWithLogger(context.Background(), sharedLogger)
			},
			expectSame: sharedLogger,
		},
		{
			name: "Empty context - returns NoneLogger",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectIsNone: true,
		},
		{
			name: "Context with CustomContextKeyValue but nil logger",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), CustomContextKey, &CustomContextKeyValue{
					Logger: nil,
				})
			},
			expectIsNone: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := tt.setupCtx()
			logger := NewLoggerFromContext(ctx)

			assert.NotNil(t, logger)

			if tt.expectSame != nil {
				assert.Equal(t, tt.expectSame, logger, "Expected the exact logger instance that was set")
			} else if tt.expectIsNone {
				_, isNoneLogger := logger.(*log.NoneLogger)
				assert.True(t, isNoneLogger, "Expected logger to be *log.NoneLogger type")
			}
		})
	}
}

func TestNewTracerFromContext(t *testing.T) {
	t.Parallel()

	// Create a shared tracer instance for the "with tracer" test case
	sharedTracer := noop.Tracer{}

	tests := []struct {
		name       string
		setupCtx   func() context.Context
		expectSame trace.Tracer // If non-nil, expect this exact tracer instance
	}{
		{
			name: "Context with tracer",
			setupCtx: func() context.Context {
				return ContextWithTracer(context.Background(), sharedTracer)
			},
			expectSame: sharedTracer,
		},
		{
			name: "Empty context - returns noop tracer",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectSame: noop.Tracer{},
		},
		{
			name: "Context with CustomContextKeyValue but nil tracer - returns noop tracer",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), CustomContextKey, &CustomContextKeyValue{
					Tracer: nil,
				})
			},
			expectSame: noop.Tracer{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := tt.setupCtx()
			tracer := NewTracerFromContext(ctx)

			// Tracer should never be nil
			assert.NotNil(t, tracer)

			if tt.expectSame != nil {
				assert.Equal(t, tt.expectSame, tracer, "Expected the exact tracer instance that was set")
			}
		})
	}
}

func TestContextWithLogger(t *testing.T) {
	t.Parallel()

	t.Run("Success - Add logger to empty context", func(t *testing.T) {
		t.Parallel()

		logger := &log.NoneLogger{}
		ctx := ContextWithLogger(context.Background(), logger)

		assert.NotNil(t, ctx)

		// Retrieve and verify
		retrievedLogger := NewLoggerFromContext(ctx)
		assert.Equal(t, logger, retrievedLogger)
	})

	t.Run("Success - Add logger to context with existing tracer", func(t *testing.T) {
		t.Parallel()

		tracer := noop.Tracer{}
		ctx := ContextWithTracer(context.Background(), tracer)

		logger := &log.NoneLogger{}
		ctx = ContextWithLogger(ctx, logger)

		// Both should be retrievable
		retrievedLogger := NewLoggerFromContext(ctx)
		retrievedTracer := NewTracerFromContext(ctx)

		assert.Equal(t, logger, retrievedLogger)
		assert.NotNil(t, retrievedTracer)
	})

	t.Run("Success - Replace existing logger", func(t *testing.T) {
		t.Parallel()

		logger1 := &log.NoneLogger{}
		ctx := ContextWithLogger(context.Background(), logger1)

		logger2 := &log.NoneLogger{}
		ctx = ContextWithLogger(ctx, logger2)

		retrievedLogger := NewLoggerFromContext(ctx)
		assert.Equal(t, logger2, retrievedLogger)
	})
}

func TestContextWithTracer(t *testing.T) {
	t.Parallel()

	t.Run("Success - Add tracer to empty context", func(t *testing.T) {
		t.Parallel()

		tracer := noop.Tracer{}
		ctx := ContextWithTracer(context.Background(), tracer)

		assert.NotNil(t, ctx)

		// Retrieve and verify
		retrievedTracer := NewTracerFromContext(ctx)
		assert.NotNil(t, retrievedTracer)
	})

	t.Run("Success - Add tracer to context with existing logger", func(t *testing.T) {
		t.Parallel()

		logger := &log.NoneLogger{}
		ctx := ContextWithLogger(context.Background(), logger)

		tracer := noop.Tracer{}
		ctx = ContextWithTracer(ctx, tracer)

		// Both should be retrievable
		retrievedLogger := NewLoggerFromContext(ctx)
		retrievedTracer := NewTracerFromContext(ctx)

		assert.Equal(t, logger, retrievedLogger)
		assert.NotNil(t, retrievedTracer)
	})
}

func TestCustomContextKey(t *testing.T) {
	t.Parallel()

	// Verify the key is defined and is the expected type
	assert.Equal(t, customContextKey("custom_context"), CustomContextKey)
}

func TestCustomContextKeyValue(t *testing.T) {
	t.Parallel()

	t.Run("Success - Create with both values", func(t *testing.T) {
		t.Parallel()

		logger := &log.NoneLogger{}
		tracer := noop.Tracer{}

		value := &CustomContextKeyValue{
			Logger: logger,
			Tracer: tracer,
		}

		assert.Equal(t, logger, value.Logger)
		assert.Equal(t, tracer, value.Tracer)
	})

	t.Run("Success - Create with nil values", func(t *testing.T) {
		t.Parallel()

		value := &CustomContextKeyValue{}

		assert.Nil(t, value.Logger)
		assert.True(t, value.Tracer == nil || value.Tracer == trace.Tracer(nil))
	})
}
