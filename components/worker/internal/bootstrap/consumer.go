package bootstrap

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"os"
	"os/signal"
	"plugin-smart-templates/components/worker/internal/adapters/rabbitmq"
	"plugin-smart-templates/components/worker/internal/services"
	"plugin-smart-templates/pkg/mongodb/report"
	"reflect"
	"sync"
	"syscall"
	"time"
)

// MultiQueueConsumer represents a multi-queue consumer.
type MultiQueueConsumer struct {
	consumerRoutes *rabbitmq.ConsumerRoutes
	UseCase        *services.UseCase
}

// NewMultiQueueConsumer create a new instance of MultiQueueConsumer.
func NewMultiQueueConsumer(routes *rabbitmq.ConsumerRoutes, useCase *services.UseCase) *MultiQueueConsumer {
	consumer := &MultiQueueConsumer{
		consumerRoutes: routes,
		UseCase:        useCase,
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

	reportID, err := mq.UseCase.GenerateReport(ctx, body)
	if err != nil {
		metadata := map[string]any{}
		metadata["error"] = err.Error()

		if errUpdate := mq.UseCase.ReportDataRepo.UpdateReportStatusById(ctx, reflect.TypeOf(report.Report{}).Name(), "Error",
			*reportID, time.UnixMicro(0), metadata); errUpdate != nil {
			opentelemetry.HandleSpanError(&span, "Error updating report status", errUpdate)

			logger.Errorf("Error updating report status: %v", errUpdate)
		}

		opentelemetry.HandleSpanError(&span, "Error generating report", err)

		logger.Errorf("Error generating report: %v", err)

		return err
	}

	return nil
}
