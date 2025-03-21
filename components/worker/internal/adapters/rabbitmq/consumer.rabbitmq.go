package rabbitmq

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons"
	constant "github.com/LerianStudio/lib-commons/commons/constants"
	"github.com/LerianStudio/lib-commons/commons/log"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/LerianStudio/lib-commons/commons/rabbitmq"
	"sync"
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
		cr.Logger.Info("Starting consumer for queue " + queueName)

		err := cr.conn.Channel.Qos(1, 0, false)
		if err != nil {
			return err
		}

		messages, err := cr.conn.Channel.Consume(
			queueName,
			"",
			false,
			false,
			false,
			false,
			nil)
		if err != nil {
			return err
		}

		for i := 0; i < cr.numWorkers; i++ {
			wg.Add(1)

			go func(workerID int, queue string, handlerFunc QueueHandlerFunc) {
				defer wg.Done()

				for {
					select {
					case <-ctx.Done():
						cr.Logger.Infof("Worker %d: Shutting down gracefully", workerID)
						return
					case message, ok := <-messages:
						if !ok {
							cr.Logger.Infof("Worker %d: Message channel closed", workerID)
							return
						}

						requestID, found := message.Headers[constant.HeaderID]
						if !found {
							requestID = commons.GenerateUUIDv7().String()
						}

						logWithFields := cr.Logger.WithFields(
							constant.HeaderID, requestID.(string),
						).WithDefaultMessageTemplate(requestID.(string) + " | ")

						ctx := commons.ContextWithLogger(
							commons.ContextWithHeaderID(context.Background(), requestID.(string)),
							logWithFields,
						)

						err := handlerFunc(ctx, message.Body)
						if err != nil {
							cr.Logger.Errorf("Worker %d: Error processing message from queue %s: %v", workerID, queue, err)

							_ = message.Nack(false, true)

							continue
						}

						_ = message.Ack(false)
					}
				}
			}(i, queueName, handler)
		}
	}

	return nil
}
