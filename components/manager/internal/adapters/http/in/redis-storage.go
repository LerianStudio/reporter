// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"context"
	"time"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libRedis "github.com/LerianStudio/lib-commons/v2/commons/redis"
	"github.com/gofiber/fiber/v2"
)

// RateLimitStorage is an alias for fiber.Storage, used to declare the
// dependency explicitly in RateLimitConfig without coupling callers to
// the fiber package.
type RateLimitStorage = fiber.Storage

// RedisStorage adapts a lib-commons RedisConnection to the fiber.Storage
// interface required by the limiter middleware. It provides graceful
// degradation: any Redis error returns (nil, nil) for Get and nil for
// Set/Delete/Reset, allowing traffic through when Redis is unavailable
// instead of blocking all requests.
type RedisStorage struct {
	conn   *libRedis.RedisConnection
	logger log.Logger
}

// Compile-time interface satisfaction check.
var _ fiber.Storage = (*RedisStorage)(nil)

// NewRedisStorage creates a new RedisStorage wrapping the given connection.
func NewRedisStorage(conn *libRedis.RedisConnection, logger log.Logger) *RedisStorage {
	return &RedisStorage{
		conn:   conn,
		logger: logger,
	}
}

// Get retrieves the value for the given key from Redis.
// Returns (nil, nil) when the key does not exist or on Redis errors
// (graceful degradation: allows traffic through on failure).
func (s *RedisStorage) Get(key string) ([]byte, error) {
	if s.conn == nil {
		return nil, nil //nolint:nilnil // graceful degradation: no Redis connection
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := s.conn.GetClient(ctx)
	if err != nil {
		s.logger.Errorf("rate-limit redis storage: failed to get client: %v", err)
		return nil, nil //nolint:nilerr // graceful degradation
	}

	val, err := client.Get(ctx, key).Bytes()
	if err != nil {
		// redis.Nil means key doesn't exist -- not an error, return nil
		// Any other error is also treated as nil for graceful degradation
		return nil, nil //nolint:nilerr // graceful degradation
	}

	return val, nil
}

// Set stores the given value with an expiration in Redis.
// Returns nil on Redis errors (graceful degradation).
func (s *RedisStorage) Set(key string, val []byte, exp time.Duration) error {
	if s.conn == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := s.conn.GetClient(ctx)
	if err != nil {
		s.logger.Errorf("rate-limit redis storage: failed to get client: %v", err)
		return nil //nolint:nilerr // graceful degradation
	}

	if err := client.Set(ctx, key, val, exp).Err(); err != nil {
		s.logger.Errorf("rate-limit redis storage: failed to set key %s: %v", key, err)
		return nil //nolint:nilerr // graceful degradation
	}

	return nil
}

// Delete removes the given key from Redis.
// Returns nil on Redis errors (graceful degradation).
func (s *RedisStorage) Delete(key string) error {
	if s.conn == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := s.conn.GetClient(ctx)
	if err != nil {
		s.logger.Errorf("rate-limit redis storage: failed to get client: %v", err)
		return nil //nolint:nilerr // graceful degradation
	}

	if err := client.Del(ctx, key).Err(); err != nil {
		s.logger.Errorf("rate-limit redis storage: failed to delete key %s: %v", key, err)
		return nil //nolint:nilerr // graceful degradation
	}

	return nil
}

// Reset is a no-op for Redis storage. Rate limit keys expire naturally
// via TTL. A full FLUSHDB would be destructive to other Redis data.
func (s *RedisStorage) Reset() error {
	return nil
}

// Close is a no-op. The RedisConnection lifecycle is managed by the
// bootstrap cleanup stack, not by the storage adapter.
func (s *RedisStorage) Close() error {
	return nil
}
