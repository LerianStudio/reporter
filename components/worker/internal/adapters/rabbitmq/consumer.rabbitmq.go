// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package rabbitmq

import (
	"context"
	"sync"

	"github.com/LerianStudio/reporter/v4/pkg"

	"github.com/LerianStudio/lib-commons/v2/commons"
	constant "github.com/LerianStudio/lib-commons/v2/commons/constants"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
)

// ConsumerRepository provides an interface for Consumer related to rabbitmq.
//
//go:generate mockgen --destination=consumer.mock.go --package=rabbitmq . ConsumerRepository
type ConsumerRepository interface {
	Register(queueName string, handler QueueHandlerFunc)
	RunConsumers() error
}

// QueueHandlerFunc is a function that processes a specific queue.
type QueueHandlerFunc func(ctx context.Context, body []byte) error

// ConsumerRoutes struct
type ConsumerRoutes struct {
	conn       *rabbitmq.RabbitMQConnection
	routes     map[string]QueueHandlerFunc
	numWorkers int
	log.Logger
	opentelemetry.Telemetry
}

// NewConsumerRoutes creates a new instance of ConsumerRoutes.
func NewConsumerRoutes(conn *rabbitmq.RabbitMQConnection, numWorkers int, logger log.Logger, telemetry *opentelemetry.Telemetry) *ConsumerRoutes {
	if numWorkers == 0 {
		numWorkers = 5
	}

	cr := &ConsumerRoutes{
		conn:       conn,
		routes:     make(map[string]QueueHandlerFunc),
		numWorkers: numWorkers,
		Logger:     logger,
		Telemetry:  *telemetry,
	}

	_, err := conn.GetNewConnect()
	if err != nil {
		panic("Failed to connect rabbitmq")
	}

	return cr
}

// Register add a new queue to handler.
func (cr *ConsumerRoutes) Register(queueName string, handler QueueHandlerFunc) {
	cr.routes[queueName] = handler
}

// RunConsumers  init consume for all registry queues.
func (cr *ConsumerRoutes) RunConsumers(ctx context.Context, wg *sync.WaitGroup) error {
	for queueName, handler := range cr.routes {
		cr.Info("Starting consumer for queue " + queueName)

		if err := cr.setupQos(); err != nil {
			return err
		}

		messages, err := cr.consumeMessages(queueName)
		if err != nil {
			return err
		}

		cr.startWorkers(ctx, wg, messages, queueName, handler)
	}

	return nil
}

func (cr *ConsumerRoutes) startWorkers(ctx context.Context, wg *sync.WaitGroup, messages <-chan amqp091.Delivery, queueName string, handler QueueHandlerFunc) {
	for i := 0; i < cr.numWorkers; i++ {
		wg.Add(1)

		go func(workerID int, queue string, handlerFunc QueueHandlerFunc) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					cr.Infof("Worker %d: Shutting down gracefully", workerID)
					return
				case message, ok := <-messages:
					if !ok {
						cr.Infof("Worker %d: Message channel closed", workerID)
						return
					}

					cr.processMessage(workerID, queue, handlerFunc, message)
				}
			}
		}(i, queueName, handler)
	}
}

// processMessage processes a single message from a specified queue using the provided handler function.
func (cr *ConsumerRoutes) processMessage(workerID int, queue string, handlerFunc QueueHandlerFunc, message amqp091.Delivery) {
	requestID, found := message.Headers[constant.HeaderID]
	if !found {
		requestID = commons.GenerateUUIDv7().String()
	}

	requestIDStr, ok := requestID.(string)
	if !ok {
		requestIDStr = commons.GenerateUUIDv7().String()
	}

	logWithFields := cr.Logger.WithFields(
		constant.HeaderID, requestIDStr,
	).WithDefaultMessageTemplate(requestIDStr + constant.LoggerDefaultSeparator)

	ctx := commons.ContextWithLogger(
		commons.ContextWithHeaderID(context.Background(), requestIDStr),
		logWithFields,
	)

	ctx = opentelemetry.ExtractTraceContextFromQueueHeaders(ctx, message.Headers)

	tracer := pkg.NewTracerFromContext(ctx)
	ctx, spanConsumer := tracer.Start(ctx, "rabbitmq.consumer.process_message")

	spanConsumer.SetAttributes(
		attribute.String("app.request.rabbitmq.consumer.request_id", requestIDStr),
	)

	err := opentelemetry.SetSpanAttributesFromStruct(&spanConsumer, "app.request.rabbitmq.consumer.message", message)
	if err != nil {
		opentelemetry.HandleSpanError(&spanConsumer, "Failed to convert message to JSON string", err)
	}

	cr.Infof("Worker %d: Starting processing for queue %s", workerID, queue)

	err = handlerFunc(ctx, message.Body)
	if err != nil {
		cr.Errorf("Worker %d: Error processing message from queue %s: %v", workerID, queue, err)
		opentelemetry.HandleSpanError(&spanConsumer, "Error processing message", err)

		_ = message.Nack(false, false)

		return
	}

	_ = message.Ack(false)

	cr.Infof("Worker %d: Successfully processed message from queue %s", workerID, queue)
}

// consumeMessages establishes a consumer for the specified queue and returns a channel for message deliveries.
func (cr *ConsumerRoutes) consumeMessages(queueName string) (<-chan amqp091.Delivery, error) {
	return cr.conn.Channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil)
}

// setupQos configures QoS settings for the RabbitMQ channel to limit message prefetch count and improve message processing.
func (cr *ConsumerRoutes) setupQos() error {
	return cr.conn.Channel.Qos(1, 0, false)
}
