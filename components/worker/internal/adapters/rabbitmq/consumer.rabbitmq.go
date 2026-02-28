// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package rabbitmq

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"maps"
	"math/big"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/LerianStudio/reporter/pkg"
	pkgConstant "github.com/LerianStudio/reporter/pkg/constant"
	pkgRabbitmq "github.com/LerianStudio/reporter/pkg/rabbitmq"

	"github.com/LerianStudio/lib-commons/v3/commons"
	constant "github.com/LerianStudio/lib-commons/v3/commons/constants"
	"github.com/LerianStudio/lib-commons/v3/commons/log"
	"github.com/LerianStudio/lib-commons/v3/commons/opentelemetry"
	"github.com/LerianStudio/lib-commons/v3/commons/rabbitmq"
	tmcore "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/core"
	tmmongo "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/mongo"
	mongoRepository "github.com/LerianStudio/reporter/pkg/mongodb/report"
	"github.com/rabbitmq/amqp091-go"

	// otel/attribute is used for span attribute types (no lib-commons wrapper available)
	"go.opentelemetry.io/otel/attribute"
	// otel/trace is used for trace.Span parameter type in handleFailedMessage
	"go.opentelemetry.io/otel/trace"
)

// RabbitMQConnectionChannel is an interface for the AMQP channel used by the consumer.
// This allows mocking the channel in tests for multi-tenant scenarios.
type RabbitMQConnectionChannel interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error
}

// RabbitMQManagerConsumerInterface is an interface for the tenant-manager RabbitMQ manager
// used by the consumer for per-tenant vhost isolation during message republishing.
type RabbitMQManagerConsumerInterface interface {
	GetConnection(ctx context.Context, tenantID string) (RabbitMQConnectionChannel, error)
}

// ConsumerRoutes struct
type ConsumerRoutes struct {
	conn            *rabbitmq.RabbitMQConnection
	routes          map[string]pkgRabbitmq.QueueHandlerFunc
	numWorkers      int
	sleepFunc       func(time.Duration)
	mongoManager    *tmmongo.Manager                 // nil in single-tenant mode
	rabbitMQManager RabbitMQManagerConsumerInterface // nil in single-tenant mode
	multiTenantMode bool
	mongoRepository *mongoRepository.ReportMongoDBRepository
	log.Logger
	opentelemetry.Telemetry
}

// Compile-time interface satisfaction check.
var _ pkgRabbitmq.ConsumerRepository = (*ConsumerRoutes)(nil)

// NewConsumerRoutes creates a new instance of ConsumerRoutes for single-tenant mode.
// mongoManager is optional: pass nil for single-tenant mode. When non-nil, the
// consumer will resolve per-tenant MongoDB connections from message headers.
func NewConsumerRoutes(conn *rabbitmq.RabbitMQConnection, numWorkers int, logger log.Logger, telemetry *opentelemetry.Telemetry, mongoManager *tmmongo.Manager, reportMongoDBRepository *mongoRepository.ReportMongoDBRepository) (*ConsumerRoutes, error) {
	if telemetry == nil {
		return nil, fmt.Errorf("telemetry must not be nil")
	}

	if numWorkers == 0 {
		numWorkers = pkgConstant.DefaultWorkerCount
	}

	cr := &ConsumerRoutes{
		conn:            conn,
		routes:          make(map[string]pkgRabbitmq.QueueHandlerFunc),
		numWorkers:      numWorkers,
		sleepFunc:       time.Sleep,
		mongoManager:    mongoManager,
		multiTenantMode: false,
		Logger:          logger,
		Telemetry:       *telemetry,
		mongoRepository: reportMongoDBRepository,
	}

	_, err := conn.GetNewConnect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	return cr, nil
}

// NewConsumerRoutesMultiTenant creates a new instance of ConsumerRoutes for multi-tenant mode.
// In multi-tenant mode:
// - Layer 1: Uses rabbitMQManager for per-tenant vhost isolation during message republishing
// - Layer 2: X-Tenant-ID header extraction is already implemented (preserved in both modes)
// - MongoDB: Uses mongoManager for per-tenant database connections
func NewConsumerRoutesMultiTenant(
	conn *rabbitmq.RabbitMQConnection,
	numWorkers int,
	logger log.Logger,
	telemetry *opentelemetry.Telemetry,
	mongoManager *tmmongo.Manager,
	rabbitMQManager RabbitMQManagerConsumerInterface,
	reportMongoDBRepository *mongoRepository.ReportMongoDBRepository,
) (*ConsumerRoutes, error) {
	if telemetry == nil {
		return nil, fmt.Errorf("telemetry must not be nil")
	}

	if numWorkers == 0 {
		numWorkers = pkgConstant.DefaultWorkerCount
	}

	cr := &ConsumerRoutes{
		conn:            conn,
		routes:          make(map[string]pkgRabbitmq.QueueHandlerFunc),
		numWorkers:      numWorkers,
		sleepFunc:       time.Sleep,
		mongoManager:    mongoManager,
		rabbitMQManager: rabbitMQManager,
		multiTenantMode: true,
		Logger:          logger,
		Telemetry:       *telemetry,
		mongoRepository: reportMongoDBRepository,
	}

	_, err := conn.GetNewConnect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	return cr, nil
}

// Register add a new queue to handler.
func (cr *ConsumerRoutes) Register(queueName string, handler pkgRabbitmq.QueueHandlerFunc) {
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

func (cr *ConsumerRoutes) startWorkers(ctx context.Context, wg *sync.WaitGroup, messages <-chan amqp091.Delivery, queueName string, handler pkgRabbitmq.QueueHandlerFunc) {
	for i := range cr.numWorkers {
		wg.Add(1)

		go func(workerID int, queue string, handlerFunc pkgRabbitmq.QueueHandlerFunc) {
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
func (cr *ConsumerRoutes) processMessage(workerID int, queue string, handlerFunc pkgRabbitmq.QueueHandlerFunc, message amqp091.Delivery) {
	if message.Headers == nil {
		message.Headers = amqp091.Table{}
	}

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

	tracer := pkg.NewTracerFromContext(ctx)

	ctx, spanConsumer := tracer.Start(ctx, "repository.rabbitmq.process_message")
	defer spanConsumer.End()

	ctx = opentelemetry.ExtractTraceContextFromQueueHeaders(ctx, message.Headers)
	ctx = extractTenantIDFromHeaders(ctx, message.Headers)

	// When a tenant MongoDB manager is available and a tenant ID was extracted
	// from the message headers, resolve the per-tenant MongoDB connection and
	// inject it into context. Downstream repository calls (getCollection) then
	// find tmcore.GetMongoForTenant succeeding and use the correct tenant DB.
	//
	// When mongoManager is nil (single-tenant mode) or the tenant ID is absent
	// (legacy messages), this block is skipped entirely and the static MongoDB
	// fallback behaviour is preserved — no change to existing single-tenant paths.
	if cr.mongoManager != nil {
		if tenantID := tmcore.GetTenantIDFromContext(ctx); tenantID != "" {
			tenantDB, tenantDBErr := cr.mongoManager.GetDatabaseForTenant(ctx, tenantID)
			if tenantDBErr != nil {
				cr.Errorf("Worker %d: failed to resolve tenant MongoDB for tenant %s: %v",
					workerID, tenantID, tenantDBErr)

				retryCount := getRetryCount(message)

				cr.handleFailedMessage(workerID, queue, message, tenantDBErr, retryCount, &spanConsumer)

				return
			}

			ctx = tmcore.ContextWithTenantMongo(ctx, tenantDB)

			logWithFields.Info("Ensuring MongoDB indexes exist for reports...")

			if indexErr := cr.mongoRepository.EnsureIndexes(ctx); indexErr != nil {
				cr.Errorf("Worker %d: failed to ensure MongoDB indexes for tenant %s: %v",
					workerID, tmcore.GetTenantIDFromContext(ctx), indexErr)
			}
		}
	}

	spanConsumer.SetAttributes(
		attribute.String("app.request.request_id", requestIDStr),
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
	if cr.conn.Channel == nil {
		return nil, fmt.Errorf("rabbitmq channel is nil, cannot consume from queue %s", queueName)
	}

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
	if cr.conn.Channel == nil {
		return fmt.Errorf("rabbitmq channel is nil, cannot setup QoS")
	}

	return cr.conn.Channel.Qos(pkgConstant.DefaultPrefetchCount, 0, false)
}

// handleFailedMessage determines whether a failed message should be retried or sent to the DLQ.
// Non-retryable errors (business validation) are immediately sent to DLQ via Nack.
// Retryable errors are republished with an incremented retry counter and exponential backoff,
// up to MaxMessageRetries attempts. When retries are exhausted, the message is Nack'd without
// requeue so the broker routes it to the configured dead-letter exchange.
func (cr *ConsumerRoutes) handleFailedMessage(workerID int, queue string, message amqp091.Delivery, err error, retryCount int, span *trace.Span) {
	if !isRetryable(err) {
		cr.Infof("Worker %d: Non-retryable error for queue %s, sending to DLQ: %v", workerID, queue, err)
		opentelemetry.HandleSpanBusinessErrorEvent(span, "Non-retryable business error, routing to DLQ", err)

		if nackErr := message.Nack(false, false); nackErr != nil {
			cr.Errorf("Worker %d: Nack failed for queue %s: %v", workerID, queue, nackErr)
		}

		return
	}

	if retryCount >= pkgConstant.MaxMessageRetries {
		cr.Errorf("Worker %d: Max retries (%d) exceeded for queue %s, sending to DLQ: %v",
			workerID, pkgConstant.MaxMessageRetries, queue, err)
		opentelemetry.HandleSpanError(span, "Max retries exceeded, routing to DLQ", err)

		if nackErr := message.Nack(false, false); nackErr != nil {
			cr.Errorf("Worker %d: Nack failed for queue %s: %v", workerID, queue, nackErr)
		}

		return
	}

	backoff := calculateBackoff(retryCount)

	cr.Infof("Worker %d: Retryable error for queue %s (attempt %d/%d), backoff %v before republish: %v",
		workerID, queue, retryCount+1, pkgConstant.MaxMessageRetries, backoff, err)

	cr.sleepFunc(backoff)

	// Build new headers with incremented retry count
	retryHeaders := buildRetryHeaders(message.Headers, retryCount, err)

	// Republish message with updated headers to the same exchange/routing-key.
	// We use the Exchange and RoutingKey from the delivery metadata, which the
	// broker populates with the original publish destination.
	if cr.conn.Channel == nil {
		cr.Errorf("Worker %d: Channel is nil, cannot republish for retry on queue %s. Sending to DLQ.", workerID, queue)
		opentelemetry.HandleSpanError(span, "Channel nil, cannot republish for retry, routing to DLQ", fmt.Errorf("rabbitmq channel is nil"))

		if nackErr := message.Nack(false, false); nackErr != nil {
			cr.Errorf("Worker %d: Nack failed for queue %s: %v", workerID, queue, nackErr)
		}

		return
	}

	publishErr := cr.conn.Channel.Publish(
		message.Exchange,
		message.RoutingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  message.ContentType,
			DeliveryMode: amqp091.Persistent,
			Headers:      retryHeaders,
			Body:         message.Body,
		},
	)
	if publishErr != nil {
		cr.Errorf("Worker %d: Failed to republish message for retry on queue %s: %v. Sending to DLQ.",
			workerID, queue, publishErr)
		opentelemetry.HandleSpanError(span, "Failed to republish for retry, routing to DLQ", publishErr)

		// If we can't republish, Nack without requeue to send to DLQ rather than
		// losing the message or creating an infinite loop.
		if nackErr := message.Nack(false, false); nackErr != nil {
			cr.Errorf("Worker %d: Nack failed for queue %s: %v", workerID, queue, nackErr)
		}

		return
	}

	// Ack the original message since we successfully republished with updated headers.
	// This removes the old message from the queue, preventing duplicates.
	if ackErr := message.Ack(false); ackErr != nil {
		cr.Errorf("Worker %d: Ack failed after republish for queue %s: %v (message may be redelivered)",
			workerID, queue, ackErr)
	}

	cr.Infof("Worker %d: Message republished for retry %d/%d on queue %s",
		workerID, retryCount+1, pkgConstant.MaxMessageRetries, queue)
}

// isRetryable classifies an error as retryable or non-retryable.
// Business validation errors (TPL-XXXX codes) are non-retryable because retrying
// will not change the outcome. Network, timeout, and unknown errors are retryable
// as transient failures may resolve on subsequent attempts.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Context cancellation and deadline exceeded are non-retryable.
	// Deadline exceeded means the operation's time budget is exhausted; retrying would
	// start with the same expired context. The message will route to DLQ for inspection.
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

	var conflictErr pkg.EntityConflictError
	if errors.As(err, &conflictErr) {
		return false
	}

	var forbiddenErr pkg.ForbiddenError
	if errors.As(err, &forbiddenErr) {
		return false
	}

	var unauthorizedErr pkg.UnauthorizedError
	if errors.As(err, &unauthorizedErr) {
		return false
	}

	var preconditionErr pkg.FailedPreconditionError
	if errors.As(err, &preconditionErr) {
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

// buildRetryHeaders creates a new header table for a retry republish.
// It copies all original headers, then overwrites the retry count (incremented)
// and failure reason (truncated to RetryFailureReasonMaxLen). This ensures
// tracing headers (e.g., traceparent) and request IDs survive across retries.
func buildRetryHeaders(original amqp091.Table, currentRetryCount int, lastErr error) amqp091.Table {
	headers := make(amqp091.Table, len(original)+2)
	maps.Copy(headers, original) // safe with nil source (no-op)

	headers[pkgConstant.RetryCountHeader] = currentRetryCount + 1
	headers[pkgConstant.RetryFailureReasonHeader] = sanitizeFailureReason(lastErr.Error())

	return headers
}

// sanitizeFailureReason truncates the error message to RetryFailureReasonMaxLen characters
// to prevent leaking internal infrastructure details (e.g., connection strings from DB driver errors)
// into message headers.
func sanitizeFailureReason(reason string) string {
	if len(reason) <= pkgConstant.RetryFailureReasonMaxLen {
		return reason
	}

	return reason[:pkgConstant.RetryFailureReasonMaxLen]
}

// extractTenantIDFromHeaders reads the X-Tenant-ID header from an AMQP message
// and, if present and non-empty, stores the tenant ID in the returned context
// using the lib-commons tenant-manager API.
//
// When the header is absent or not a string (e.g. legacy single-tenant messages),
// the context is returned unchanged, preserving full backward compatibility.
// All downstream repository calls that depend on tenant context will fall back
// to their single-tenant code path, exactly as before multi-tenant support was added.
func extractTenantIDFromHeaders(ctx context.Context, headers amqp091.Table) context.Context {
	if headers == nil {
		return ctx
	}

	if tenantID, ok := headers[pkgConstant.HeaderXTenantID].(string); ok && tenantID != "" {
		return tmcore.SetTenantIDInContext(ctx, tenantID)
	}

	return ctx
}

// calculateBackoff computes the delay before the next retry attempt using exponential backoff with jitter.
// Formula: min(initialBackoff * 2^attempt, maxBackoff) + random_jitter(0, RetryJitterMax)
// Jitter prevents thundering herd when multiple consumers retry simultaneously.
func calculateBackoff(attempt int) time.Duration {
	backoff := pkgConstant.RetryInitialBackoff

	for range attempt {
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
