// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	pkgRedis "github.com/LerianStudio/reporter/pkg/redis"

	libCommons "github.com/LerianStudio/lib-commons/v3/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v3/commons/opentelemetry"
	libRedis "github.com/LerianStudio/lib-commons/v3/commons/redis"
	tmValkey "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/valkey"
	goRedis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

// RedisConsumerRepository is a Redis implementation of the Redis consumer.
type RedisConsumerRepository struct {
	conn *libRedis.RedisConnection
}

// Compile-time interface satisfaction check.
var _ pkgRedis.RedisRepository = (*RedisConsumerRepository)(nil)

// NewConsumerRedis returns a new instance of RedisRepository using the given Redis connection.
func NewConsumerRedis(rc *libRedis.RedisConnection) (*RedisConsumerRepository, error) {
	r := &RedisConsumerRepository{
		conn: rc,
	}
	if _, err := r.conn.GetClient(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return r, nil
}

// Set sets a key in the redis
func (rc *RedisConsumerRepository) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.redis.set")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.key", key),
		attribute.String("app.request.ttl", ttl.String()),
	)

	rds, err := rc.conn.GetClient(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get redis", err)

		return err
	}

	tenantKey := tmValkey.GetKeyFromContext(ctx, key)

	logger.Infof("value of ttl: %v", ttl)

	err = rds.Set(ctx, tenantKey, value, ttl).Err()
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to set on redis", err)

		return err
	}

	return nil
}

// SetNX sets a key in redis only if it does not already exist (atomic compare-and-set).
// Returns true if the key was set (first request), false if it already existed (duplicate).
func (rc *RedisConsumerRepository) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.redis.set_nx")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.key", key),
		attribute.String("app.request.ttl", ttl.String()),
	)

	rds, err := rc.conn.GetClient(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get redis client", err)

		return false, err
	}

	tenantKey := tmValkey.GetKeyFromContext(ctx, key)

	logger.Infof("SetNX key: %s, ttl: %v", tenantKey, ttl)

	result, err := rds.SetNX(ctx, tenantKey, value, ttl).Result()
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to set_nx on redis", err)

		return false, err
	}

	span.SetAttributes(
		attribute.Bool("app.response.was_set", result),
	)

	return result, nil
}

// Get recovers a key from the redis
func (rc *RedisConsumerRepository) Get(ctx context.Context, key string) (string, error) {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.redis.get")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.key", key),
	)

	rds, err := rc.conn.GetClient(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get redis", err)

		return "", err
	}

	tenantKey := tmValkey.GetKeyFromContext(ctx, key)

	val, err := rds.Get(ctx, tenantKey).Result()
	if err != nil {
		if errors.Is(err, goRedis.Nil) {
			span.SetAttributes(attribute.Bool("app.cache.hit", false))

			return "", err
		}

		libOpentelemetry.HandleSpanError(&span, "Failed to get on redis", err)

		return "", err
	}

	span.SetAttributes(attribute.Bool("app.cache.hit", true))

	logger.Infof("value : %v", val)

	return val, nil
}

// Del deletes a key from the redis
func (rc *RedisConsumerRepository) Del(ctx context.Context, key string) error {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.redis.del")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.key", key),
	)

	rds, err := rc.conn.GetClient(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to del redis", err)

		return err
	}

	tenantKey := tmValkey.GetKeyFromContext(ctx, key)

	val, err := rds.Del(ctx, tenantKey).Result()
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to del on redis", err)

		return err
	}

	logger.Infof("value : %v", val)

	return nil
}
