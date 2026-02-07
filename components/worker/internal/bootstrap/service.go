// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"context"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/pdf"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libMongo "github.com/LerianStudio/lib-commons/v2/commons/mongo"
	libOtel "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	libRabbitMQ "github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
)

// Service is the application glue where we put all top level components to be used.
type Service struct {
	*MultiQueueConsumer
	log.Logger
	healthChecker      *pkg.HealthChecker
	mongoConnection    *libMongo.MongoConnection
	rabbitMQConnection *libRabbitMQ.RabbitMQConnection
	pdfPool            *pdf.WorkerPool
	telemetry          *libOtel.Telemetry
}

// Run starts the application.
// This is the only necessary code to run an app in main.go
func (app *Service) Run() {
	commons.NewLauncher(
		commons.WithLogger(app.Logger),
		commons.RunApp("RabbitMQ Consumer", app.MultiQueueConsumer),
	).Run()

	// Graceful shutdown - close resources in reverse initialization order
	app.Info("Starting graceful shutdown...")

	// Stop health checker
	if app.healthChecker != nil {
		app.Info("Stopping health checker...")
		app.healthChecker.Stop()
	}

	// Close PDF worker pool (waits for in-progress tasks to complete)
	if app.pdfPool != nil {
		app.Info("Closing PDF worker pool...")
		app.pdfPool.Close()
		app.Info("PDF worker pool closed")
	}

	// Close RabbitMQ connection
	if app.rabbitMQConnection != nil {
		app.Info("Closing RabbitMQ connection...")

		if app.rabbitMQConnection.Channel != nil {
			if err := app.rabbitMQConnection.Channel.Close(); err != nil {
				app.Errorf("Failed to close RabbitMQ channel: %v", err)
			}
		}

		if app.rabbitMQConnection.Connection != nil && !app.rabbitMQConnection.Connection.IsClosed() {
			if err := app.rabbitMQConnection.Connection.Close(); err != nil {
				app.Errorf("Failed to close RabbitMQ connection: %v", err)
			} else {
				app.Info("RabbitMQ connection closed")
			}
		}
	}

	// Close MongoDB connection
	if app.mongoConnection != nil && app.mongoConnection.DB != nil {
		app.Info("Closing MongoDB connection...")

		if err := app.mongoConnection.DB.Disconnect(context.Background()); err != nil {
			app.Errorf("Failed to close MongoDB connection: %v", err)
		} else {
			app.Info("MongoDB connection closed")
		}
	}

	// Flush telemetry (must be last to capture shutdown spans)
	if app.telemetry != nil {
		app.Info("Flushing telemetry...")
		app.telemetry.ShutdownTelemetry()
		app.Info("Telemetry flushed")
	}

	app.Info("Graceful shutdown complete")
}
