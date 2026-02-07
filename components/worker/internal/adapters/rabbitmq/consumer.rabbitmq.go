// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package rabbitmq

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/LerianStudio/reporter/pkg"
	pkgConstant "github.com/LerianStudio/reporter/pkg/constant"

	"github.com/LerianStudio/lib-commons/v2/commons"
	constant "github.com/LerianStudio/lib-commons/v2/commons/constants"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
func NewConsumerRoutes(conn *rabbitmq.RabbitMQConnection, numWorkers int, logger log.Logger, telemetry *opentelemetry.Telemetry) (*ConsumerRoutes, error) {
	if numWorkers == 0 {
		numWorkers = pkgConstant.DefaultWorkerCount
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
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	return cr, nil
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
			defer func() {
				if r := recover(); r != nil {
					cr.Errorf("Panic recovered in RabbitMQ worker %d for queue %s: %v\nStack: %s", workerID, queue, r, string(debug.Stack()))
				}
			}()

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

	ctx, spanConsumer := tracer.Start(ctx, "repository.rabbitmq.process_message")
	defer spanConsumer.End()

	spanConsumer.SetAttributes(
		attribute.String("app.request.rabbitmq.consumer.request_id", requestIDStr),
	)

	err := opentelemetry.SetSpanAttributesFromStruct(&spanConsumer, "app.request.rabbitmq.consumer.message", message)
	if err != nil {
		opentelemetry.HandleSpanError(&spanConsumer, "Failed to convert message to JSON string", err)
	}

	retryCount := getRetryCount(message)

	spanConsumer.SetAttributes(
		attribute.Int("app.request.rabbitmq.consumer.retry_count", retryCount),
	)

	cr.Infof("Worker %d: Starting processing for queue %s (attempt %d)", workerID, queue, retryCount+1)

	err = handlerFunc(ctx, message.Body)
	if err != nil {
		cr.Errorf("Worker %d: Error processing message from queue %s: %v", workerID, queue, err)
		opentelemetry.HandleSpanError(&spanConsumer, "Error processing message", err)

		cr.handleFailedMessage(workerID, queue, message, err, retryCount, &spanConsumer)

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
	return cr.conn.Channel.Qos(pkgConstant.DefaultPrefetchCount, 0, false)
}

// handleFailedMessage determines whether a failed message should be retried or sent to the DLQ.
// Non-retryable errors (business validation) are immediately sent to DLQ via Nack.
// Retryable errors are requeued with exponential backoff up to MaxMessageRetries attempts.
func (cr *ConsumerRoutes) handleFailedMessage(workerID int, queue string, message amqp091.Delivery, err error, retryCount int, span *trace.Span) {
	if !isRetryable(err) {
		cr.Infof("Worker %d: Non-retryable error for queue %s, sending to DLQ: %v", workerID, queue, err)
		opentelemetry.HandleSpanBusinessErrorEvent(span, "Non-retryable business error, routing to DLQ", err)

		_ = message.Nack(false, false)

		return
	}

	if retryCount >= pkgConstant.MaxMessageRetries {
		cr.Errorf("Worker %d: Max retries (%d) exceeded for queue %s, sending to DLQ: %v",
			workerID, pkgConstant.MaxMessageRetries, queue, err)
		opentelemetry.HandleSpanError(span, "Max retries exceeded, routing to DLQ", err)

		_ = message.Nack(false, false)

		return
	}

	backoff := calculateBackoff(retryCount)

	cr.Infof("Worker %d: Retryable error for queue %s (attempt %d/%d), backoff %v before requeue: %v",
		workerID, queue, retryCount+1, pkgConstant.MaxMessageRetries, backoff, err)

	time.Sleep(backoff)

	_ = message.Nack(false, true)
}

// isRetryable classifies an error as retryable or non-retryable.
// Business validation errors (TPL-XXXX codes) are non-retryable because retrying
// will not change the outcome. Network, timeout, and unknown errors are retryable
// as transient failures may resolve on subsequent attempts.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Context cancellation and deadline exceeded are non-retryable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Business validation errors with TPL-XXXX codes are non-retryable.
	// These represent input/validation failures that will not succeed on retry.
	errMsg := err.Error()
	if strings.Contains(errMsg, "TPL-") {
		return false
	}

	// Check for known non-retryable error types from the application.
	// Business validation and domain errors will never succeed on retry.
	var validationErr pkg.ValidationError
	if errors.As(err, &validationErr) {
		return false
	}

	var notFoundErr pkg.EntityNotFoundError
	if errors.As(err, &notFoundErr) {
		return false
	}

	var knownFieldsErr pkg.ValidationKnownFieldsError
	if errors.As(err, &knownFieldsErr) {
		return false
	}

	var unknownFieldsErr pkg.ValidationUnknownFieldsError
	if errors.As(err, &unknownFieldsErr) {
		return false
	}

	var unprocessableErr pkg.UnprocessableOperationError
	if errors.As(err, &unprocessableErr) {
		return false
	}

	// Unknown errors default to retryable; DLQ handles exhausted retries
	return true
}

// getRetryCount reads the retry count from the RabbitMQ message headers.
// Returns 0 if the header is missing or cannot be parsed, ensuring safe default behavior
// for messages that have not been retried yet.
func getRetryCount(msg amqp091.Delivery) int {
	if msg.Headers == nil {
		return 0
	}

	val, exists := msg.Headers[pkgConstant.RetryCountHeader]
	if !exists {
		return 0
	}

	// RabbitMQ headers can store values as different numeric types depending on
	// the publisher and serialization. Handle all common variants safely.
	switch v := val.(type) {
	case int:
		if v < 0 {
			return 0
		}

		return v
	case int32:
		if v < 0 {
			return 0
		}

		return int(v)
	case int64:
		if v < 0 {
			return 0
		}

		return int(v)
	case float64:
		if v < 0 {
			return 0
		}

		return int(v)
	default:
		return 0
	}
}

// calculateBackoff computes the delay before the next retry attempt using exponential backoff with jitter.
// Formula: min(initialBackoff * 2^attempt, maxBackoff) + random_jitter(0, RetryJitterMax)
// Jitter prevents thundering herd when multiple consumers retry simultaneously.
func calculateBackoff(attempt int) time.Duration {
	backoff := pkgConstant.RetryInitialBackoff

	for i := 0; i < attempt; i++ {
		backoff *= 2
		if backoff > pkgConstant.RetryMaxBackoff {
			backoff = pkgConstant.RetryMaxBackoff

			break
		}
	}

	// Add cryptographically secure random jitter to prevent synchronized retries
	jitterMax := int64(pkgConstant.RetryJitterMax)
	if jitterMax > 0 {
		n, err := rand.Int(rand.Reader, big.NewInt(jitterMax))
		if err == nil {
			backoff += time.Duration(n.Int64())
		}
	}

	return backoff
}
