// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/LerianStudio/reporter/components/worker/internal/adapters/rabbitmq"
	"github.com/LerianStudio/reporter/components/worker/internal/services"
	"github.com/LerianStudio/reporter/pkg"
	pkgHTTP "github.com/LerianStudio/reporter/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
)

// MultiQueueConsumer represents a multi-queue consumer.
type MultiQueueConsumer struct {
	consumerRoutes *rabbitmq.ConsumerRoutes
	UseCase        *services.UseCase
	logger         log.Logger
}

// NewMultiQueueConsumer create a new instance of MultiQueueConsumer.
func NewMultiQueueConsumer(routes *rabbitmq.ConsumerRoutes, useCase *services.UseCase, queueName string, logger log.Logger) *MultiQueueConsumer {
	consumer := &MultiQueueConsumer{
		consumerRoutes: routes,
		UseCase:        useCase,
		logger:         logger,
	}

	// Registry handlers for each queue
	if routes != nil {
		routes.Register(queueName, consumer.handlerGenerateReport)
	}

	return consumer
}

// Run starts consumers for all registered queues.
func (mq *MultiQueueConsumer) Run(l *commons.Launcher) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	pkg.GoWithCleanup(mq.logger, func() {
		<-sigs
		cancel()
	}, func(_ any) {
		cancel()
	})

	if err := mq.consumerRoutes.RunConsumers(ctx, wg); err != nil {
		return err
	}

	wg.Wait()

	return nil
}

// handlerGenerateReport processes messages from the generate report queue.
func (mq *MultiQueueConsumer) handlerGenerateReport(ctx context.Context, body []byte) error {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.report.generate")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	logger.Info("Processing message from generate report queue")

	err := mq.UseCase.GenerateReport(ctx, body)
	if err != nil {
		if pkgHTTP.IsBusinessError(err) {
			opentelemetry.HandleSpanBusinessErrorEvent(&span, "Error generating report.", err)
		} else {
			opentelemetry.HandleSpanError(&span, "Error generating report.", err)
		}

		logger.Errorf("Error generating report: %v", err)

		return err
	}

	return nil
}
