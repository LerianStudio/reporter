package rabbitmq

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons"
	constant "github.com/LerianStudio/lib-commons/commons/constants"
	"github.com/LerianStudio/lib-commons/commons/log"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/LerianStudio/lib-commons/commons/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"sync"
	"time"
)

// ConsumerRepository provides an interface for Consumer related to rabbitmq.
//
//go:generate mockgen --destination=consumer.mock.go --package=rabbitmq . ConsumerRepository
type ConsumerRepository interface {
	Register(queueName string, handler QueueHandlerFunc)
	RunConsumers() error
}

// QueueHandlerFunc is a function that process a specific queue.
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
	).WithDefaultMessageTemplate(requestIDStr + " | ")

	ctx := commons.ContextWithLogger(
		commons.ContextWithHeaderID(context.Background(), requestIDStr),
		logWithFields,
	)

	err := handlerFunc(ctx, message.Body)
	if err != nil {
		cr.Errorf("Worker %d: Error processing message from queue %s: %v", workerID, queue, err)

		// Recovery and increment retryCount
		if !cr.retryMessageWithCount(message, workerID, queue) {
			return false
		}

		// Exclude original
		_ = message.Nack(false, false)

		return true
	}

	_ = message.Ack(false)

	return false
}

// retryMessage retries a message with the specified retryCount.
// Returns true if the message was successfully requeued; otherwise, returns false.
func (cr *ConsumerRoutes) retryMessageWithCount(message amqp091.Delivery, workerID int, queue string) bool {
	var retryCount int32
	if val, ok := message.Headers["x-retry-count"].(int); ok {
		retryCount = int32(val)
	}

	retryCount++
	if retryCount >= 3 {
		cr.Warnf("Worker %d: Discarding message from queue %s after %d attempts", workerID, queue, retryCount)

		_ = message.Nack(false, false)

		return false
	}

	// Republish with retryCount + 1
	headers := amqp091.Table{}
	for k, v := range message.Headers {
		headers[k] = v
	}

	headers["x-retry-count"] = retryCount

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
