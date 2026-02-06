// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package redis

import (
	"context"
	"fmt"
	"time"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	libRedis "github.com/LerianStudio/lib-commons/v2/commons/redis"
	"go.opentelemetry.io/otel/attribute"
)

// RedisRepository provides an interface for redis.
// It defines methods for setting, getting, deleting keys, and incrementing values.
//
//go:generate mockgen --destination=consumer.redis.mock.go --package=redis . RedisRepository
type RedisRepository interface {
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

// RedisConsumerRepository is a Redis implementation of the Redis consumer.
type RedisConsumerRepository struct {
	conn *libRedis.RedisConnection
}

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
		attribute.String("app.request.value", value),
		attribute.String("app.request.ttl", ttl.String()),
	)

	rds, err := rc.conn.GetClient(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get redis", err)

		return err
	}

	logger.Infof("value of ttl: %v", ttl)

	err = rds.Set(ctx, key, value, ttl).Err()
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to set on redis", err)

		return err
	}

	return nil
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

	val, err := rds.Get(ctx, key).Result()
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get on redis", err)

		return "", err
	}

	span.SetAttributes(
		attribute.String("app.response.value", val),
	)

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

	val, err := rds.Del(ctx, key).Result()
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to del on redis", err)

		return err
	}

	logger.Infof("value : %v", val)

	return nil
}
