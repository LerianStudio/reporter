// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"testing"
	"time"

	"github.com/LerianStudio/lib-commons/v3/commons/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRedisStorage_ImplementsFiberStorage verifies at the type level that
// RedisStorage satisfies the fiber.Storage interface.
func TestRedisStorage_ImplementsFiberStorage(t *testing.T) {
	t.Parallel()

	// This compiles only if RedisStorage implements fiber.Storage.
	var _ RateLimitStorage = (*RedisStorage)(nil)
}

// TestRedisStorage_GracefulDegradation_NilConnection verifies that all
// RedisStorage methods return nil (no error, no data) when the underlying
// Redis connection is nil. This tests the graceful degradation path --
// traffic is allowed through when Redis is unavailable.
func TestRedisStorage_GracefulDegradation_NilConnection(t *testing.T) {
	t.Parallel()

	// Pass nil connection to simulate Redis being unavailable.
	s := &RedisStorage{conn: nil, logger: &noopLogger{}}

	t.Run("Get returns nil on nil connection", func(t *testing.T) {
		t.Parallel()

		val, err := s.Get("test-key")
		assert.Nil(t, val, "Get should return nil value on Redis failure")
		assert.NoError(t, err, "Get should return nil error on Redis failure (graceful degradation)")
	})

	t.Run("Set returns nil on nil connection", func(t *testing.T) {
		t.Parallel()

		err := s.Set("test-key", []byte("val"), 60*time.Second)
		assert.NoError(t, err, "Set should return nil error on Redis failure (graceful degradation)")
	})

	t.Run("Delete returns nil on nil connection", func(t *testing.T) {
		t.Parallel()

		err := s.Delete("test-key")
		assert.NoError(t, err, "Delete should return nil error on Redis failure (graceful degradation)")
	})

	t.Run("Reset is a no-op", func(t *testing.T) {
		t.Parallel()

		err := s.Reset()
		require.NoError(t, err, "Reset should always succeed (no-op)")
	})

	t.Run("Close is a no-op", func(t *testing.T) {
		t.Parallel()

		err := s.Close()
		require.NoError(t, err, "Close should always succeed (no-op)")
	})
}

// noopLogger is a minimal logger implementation for tests that satisfies
// the log.Logger interface without producing any output.
type noopLogger struct{}

// Compile-time interface check.
var _ log.Logger = (*noopLogger)(nil)

func (n *noopLogger) Info(_ ...any)                                  {}
func (n *noopLogger) Infof(_ string, _ ...any)                       {}
func (n *noopLogger) Infoln(_ ...any)                                {}
func (n *noopLogger) Error(_ ...any)                                 {}
func (n *noopLogger) Errorf(_ string, _ ...any)                      {}
func (n *noopLogger) Errorln(_ ...any)                               {}
func (n *noopLogger) Warn(_ ...any)                                  {}
func (n *noopLogger) Warnf(_ string, _ ...any)                       {}
func (n *noopLogger) Warnln(_ ...any)                                {}
func (n *noopLogger) Debug(_ ...any)                                 {}
func (n *noopLogger) Debugf(_ string, _ ...any)                      {}
func (n *noopLogger) Debugln(_ ...any)                               {}
func (n *noopLogger) Fatal(_ ...any)                                 {}
func (n *noopLogger) Fatalf(_ string, _ ...any)                      {}
func (n *noopLogger) Fatalln(_ ...any)                               {}
func (n *noopLogger) WithFields(_ ...any) log.Logger                 { return n }
func (n *noopLogger) WithDefaultMessageTemplate(_ string) log.Logger { return n }
func (n *noopLogger) Sync() error                                    { return nil }
