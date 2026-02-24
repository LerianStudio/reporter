// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/model"
	pkgRabbitmq "github.com/LerianStudio/reporter/pkg/rabbitmq"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libConstants "github.com/LerianStudio/lib-commons/v2/commons/constants"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	libRabbitmq "github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
)

// sleepFunc is the function used for sleeping between retries.
// Overridable in tests for deterministic behavior.
var sleepFunc = time.Sleep

// ProducerRabbitMQRepository is a rabbitmq implementation of the producer
type ProducerRabbitMQRepository struct {
	conn *libRabbitmq.RabbitMQConnection
}

// Compile-time interface satisfaction check.
var _ pkgRabbitmq.ProducerRepository = (*ProducerRabbitMQRepository)(nil)

// NewProducerRabbitMQ returns a new instance of ProducerRabbitMQRepository using the given rabbitmq connection.
// Connection is established lazily on first use to avoid panic during initialization.
func NewProducerRabbitMQ(c *libRabbitmq.RabbitMQConnection) *ProducerRabbitMQRepository {
	prmq := &ProducerRabbitMQRepository{
		conn: c,
	}

	// Try to connect but don't panic if it fails
	// Connection will be retried on first use via EnsureChannel
	_, err := c.GetNewConnect()
	if err != nil {
		c.Logger.Errorf("Failed to connect to RabbitMQ during initialization: %v", err)
		c.Logger.Warn("RabbitMQ connection will be retried on first message publish")
	} else {
		c.Logger.Info("RabbitMQ producer connected successfully")
	}

	return prmq
}

// ProducerDefault publishes a message to RabbitMQ with midaz-style retry logic.
// On each attempt it calls EnsureChannel() to restore the channel if the connection
// dropped, then publishes. Retries up to ProducerMaxRetries with exponential backoff
// and full jitter to prevent thundering herd after a broker restart.
func (prmq *ProducerRabbitMQRepository) ProducerDefault(ctx context.Context, exchange, key string, queueMessage model.ReportMessage) (*string, error) {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	logger.Infof("Init sent message")

	ctx, spanProducer := tracer.Start(ctx, "repository.rabbitmq.publish_message")
	defer spanProducer.End()

	spanProducer.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.exchange", exchange),
		attribute.String("app.request.key", key),
	)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&spanProducer, "app.request.rabbitmq.message", queueMessage)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanProducer, "Failed to convert queue message to JSON string", err)
	}

	message, err := json.Marshal(queueMessage)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanProducer, "Failed to marshal queue message struct", err)

		logger.Errorf("Failed to marshal queue message struct")

		return nil, err
	}

	retryCount := 0

	headers := amqp.Table{
		libConstants.HeaderID: reqId,
		"x-retry-count":       retryCount,
	}

	libOpentelemetry.InjectTraceHeadersIntoQueue(ctx, (*map[string]any)(&headers))

	// Midaz-style retry loop: EnsureChannel + publish with exponential backoff
	backoff := constant.ProducerInitialBackoff

	var publishErr error

	for attempt := 0; attempt <= constant.ProducerMaxRetries; attempt++ {
		// Ensure channel is available (reconnects if connection dropped)
		if chanErr := prmq.conn.EnsureChannel(); chanErr != nil {
			logger.Errorf("EnsureChannel failed (attempt %d/%d): %v", attempt+1, constant.ProducerMaxRetries+1, chanErr)

			spanProducer.SetAttributes(
				attribute.Int("app.request.rabbitmq.retry_attempt", attempt),
			)

			if attempt == constant.ProducerMaxRetries {
				libOpentelemetry.HandleSpanError(&spanProducer, "Failed to ensure RabbitMQ channel after all retries", chanErr)

				return nil, chanErr
			}

			sleepDuration := pkg.FullJitter(backoff)

			logger.Infof("Retrying EnsureChannel in %v (attempt %d/%d)", sleepDuration, attempt+1, constant.ProducerMaxRetries+1)

			sleepFunc(sleepDuration)

			backoff = pkg.NextBackoff(backoff)

			continue
		}

		// Attempt publish
		publishErr = prmq.conn.Channel.Publish(
			exchange,
			key,
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				Headers:      headers,
				Body:         message,
			})

		// Success - return immediately
		if publishErr == nil {
			logger.Infoln("Messages sent successfully")

			return nil, nil
		}

		// Failure - log and retry with backoff
		logger.Errorf("Publish failed (attempt %d/%d): %v", attempt+1, constant.ProducerMaxRetries+1, publishErr)

		spanProducer.SetAttributes(
			attribute.Int("app.request.rabbitmq.retry_attempt", attempt),
		)

		if attempt == constant.ProducerMaxRetries {
			libOpentelemetry.HandleSpanError(&spanProducer, "Failed to publish message after all retries", publishErr)

			return nil, publishErr
		}

		sleepDuration := pkg.FullJitter(backoff)

		logger.Infof("Retrying publish in %v (attempt %d/%d)", sleepDuration, attempt+1, constant.ProducerMaxRetries+1)

		sleepFunc(sleepDuration)

		backoff = pkg.NextBackoff(backoff)
	}

	return nil, publishErr
}
