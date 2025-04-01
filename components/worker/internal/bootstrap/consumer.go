package bootstrap

import (
	"context"
	"encoding/json"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"os"
	"os/signal"
	"plugin-template-engine/components/worker/internal/adapters/rabbitmq"
	"sync"
	"syscall"
)

// MultiQueueConsumer represents a multi-queue consumer.
type MultiQueueConsumer struct {
	consumerRoutes *rabbitmq.ConsumerRoutes
}

// GenerateReportMessage message structure for report generation.
type GenerateReportMessage struct {
	ID           uuid.UUID        `json:"id"`
	Type         string           `json:"type"`
	FileURL      string           `json:"fileUrl"`
	MappedFields []map[string]any `json:"mappedFields"`
}

// NewMultiQueueConsumer create a new instance of MultiQueueConsumer.
func NewMultiQueueConsumer(routes *rabbitmq.ConsumerRoutes) *MultiQueueConsumer {
	consumer := &MultiQueueConsumer{
		consumerRoutes: routes,
	}

	// Registry handlers for each queue
	routes.Register(os.Getenv("RABBITMQ_GENERATE_REPORT_QUEUE"), consumer.handlerGenerateReport)

	return consumer
}

// Run starts consumers for all registered queues.
func (mq *MultiQueueConsumer) Run(l *commons.Launcher) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}

	// Interrupt signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()

	if err := mq.consumerRoutes.RunConsumers(ctx, wg); err != nil {
		return err
	}

	wg.Wait()

	return nil
}

// handlerGenerateReport processes messages from the generate report queue.
func (mq *MultiQueueConsumer) handlerGenerateReport(ctx context.Context, body []byte) error {
	logger := commons.NewLoggerFromContext(ctx)
	tracer := commons.NewTracerFromContext(ctx)

	_, span := tracer.Start(ctx, "consumer.handler_generate_report")
	defer span.End()

	logger.Info("Processing message from generate report queue")

	var message GenerateReportMessage

	err := json.Unmarshal(body, &message)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Error unmarshalling message JSON", err)

		logger.Errorf("Error unmarshalling accounts message JSON: %v", err)

		return err
	}

	logger.Infof("Generate report message consumed: %s", message)

	// TODO: generate report use case here

	return nil
}
