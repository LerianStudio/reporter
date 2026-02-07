// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/LerianStudio/reporter/pkg/model"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libConstants "github.com/LerianStudio/lib-commons/v2/commons/constants"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	libRabbitmq "github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
)

// ProducerRepository provides an interface for Producer related to rabbitmq.
//
//go:generate mockgen --destination=producer.mock.go --package=rabbitmq . ProducerRepository
type ProducerRepository interface {
	ProducerDefault(ctx context.Context, exchange, key string, message model.ReportMessage) (*string, error)
}

// ProducerRabbitMQRepository is a rabbitmq implementation of the producer
type ProducerRabbitMQRepository struct {
	conn *libRabbitmq.RabbitMQConnection
}

// NewProducerRabbitMQ returns a new instance of ProducerRabbitMQRepository using the given rabbitmq connection.
// Connection is established lazily on first use to avoid panic during initialization.
func NewProducerRabbitMQ(c *libRabbitmq.RabbitMQConnection) *ProducerRabbitMQRepository {
	prmq := &ProducerRabbitMQRepository{
		conn: c,
	}

	// Try to connect but don't panic if it fails
	// Connection will be retried on first use
	_, err := c.GetNewConnect()
	if err != nil {
		c.Logger.Errorf("Failed to connect to RabbitMQ during initialization: %v", err)
		c.Logger.Warn("RabbitMQ connection will be retried on first message publish")
	} else {
		c.Logger.Info("RabbitMQ producer connected successfully")
	}

	return prmq
}

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

	// Ensure connection is established before publishing
	if prmq.conn.Channel == nil || prmq.conn.Channel.IsClosed() {
		logger.Warn("RabbitMQ channel not initialized - attempting to connect...")

		_, err := prmq.conn.GetNewConnect()
		if err != nil {
			libOpentelemetry.HandleSpanError(&spanProducer, "Failed to establish RabbitMQ connection", err)
			logger.Errorf("Failed to establish RabbitMQ connection: %v", err)

			return nil, err
		}

		logger.Info("RabbitMQ connection established on-demand")
	}

	retryCount := 0

	headers := amqp.Table{
		libConstants.HeaderID: reqId,
		"x-retry-count":       retryCount,
	}

	libOpentelemetry.InjectTraceHeadersIntoQueue(ctx, (*map[string]any)(&headers))

	err = prmq.conn.Channel.Publish(
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
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanProducer, "Failed to marshal queue message struct", err)

		logger.Errorf("Failed to publish message: %s", err)

		return nil, err
	}

	logger.Infoln("Messages sent successfully")

	return nil, nil
}
