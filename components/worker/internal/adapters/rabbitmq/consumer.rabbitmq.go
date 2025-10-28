package rabbitmq

import (
	"context"
	"sync"
	"time"

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
// Returns true if an error occurred during processing and the message was requeued; otherwise, returns false.
func (cr *ConsumerRoutes) processMessage(workerID int, queue string, handlerFunc QueueHandlerFunc, message amqp091.Delivery) bool {
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

	err = handlerFunc(ctx, message.Body)
	if err != nil {
		cr.Errorf("Worker %d: Error processing message from queue %s: %v", workerID, queue, err)

		// Always retry any error 3 times with backoff
		if !cr.retryMessageWithCount(message, workerID, queue, err) {
			_ = message.Ack(false)
			return false
		}

		// Nack to exclude original (will be republished with retry_count++)
		_ = message.Nack(false, false)

		return true
	}

	_ = message.Ack(false)

	return false
}

// retryMessage retries a message with the specified retryCount.
// Returns true if the message was successfully requeued; otherwise, returns false.
func (cr *ConsumerRoutes) retryMessageWithCount(message amqp091.Delivery, workerID int, queue string, processErr error) bool {
	// Safely extract current retry count
	var retryCount int32

	if raw, ok := message.Headers["x-retry-count"]; ok {
		switch v := raw.(type) {
		case int32:
			retryCount = v
		case int64:
			// Check for overflow before converting
			if v > int64(^int32(0)) || v < int64(^int32(0)>>1) {
				retryCount = 0 // Overflow - reset to 0
			} else {
				retryCount = int32(v)
			}
		case int:
			// Check for overflow before converting
			if v > int(^int32(0)) || v < int(^int32(0)>>1) {
				retryCount = 0 // Overflow - reset to 0
			} else {
				retryCount = int32(v)
			}
		case float32:
			retryCount = int32(v)
		case float64:
			retryCount = int32(v)
		case string:
			// fallback to zero if string cannot be parsed (avoid panic)
			retryCount = 0
		default:
			retryCount = 0
		}
	} else {
		retryCount = 0
	}

	if retryCount >= 3 {
		cr.Warnf("Worker %d: Max retries reached for message from queue %s after %d attempts", workerID, queue, retryCount)
		cr.Warnf("Worker %d: Report status should have been updated to Error by handler", workerID)

		// Don't retry anymore - handler should have updated report status to "Error"
		// Return false to signal caller to ACK the message
		return false
	}

	// Backoff before republishing: 2^n seconds where n is current retryCount
	backoff := time.Duration(1<<retryCount) * time.Second
	cr.Infof("Worker %d: Applying backoff of %v before retrying message from queue %s", workerID, backoff, queue)
	time.Sleep(backoff)

	// Republish with retryCount + 1
	retryCount++

	headers := amqp091.Table{}
	for k, v := range message.Headers {
		headers[k] = v
	}

	headers["x-retry-count"] = retryCount
	if processErr != nil {
		headers["x-failure-reason"] = processErr.Error()
	}

	errPub := cr.conn.Channel.Publish(
		message.Exchange,
		message.RoutingKey,
		false,
		false,
		amqp091.Publishing{
			Headers:      headers,
			ContentType:  message.ContentType,
			DeliveryMode: message.DeliveryMode,
			Body:         message.Body,
			Timestamp:    time.Now(),
		},
	)
	if errPub != nil {
		cr.Errorf("Worker %d: Failed to republish message: %v", workerID, errPub)
	} else {
		cr.Infof("Worker %d: Republished message with retryCount=%d", workerID, retryCount)
	}

	return true
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
