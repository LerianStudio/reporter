// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"context"

	"github.com/LerianStudio/reporter/pkg"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libMongo "github.com/LerianStudio/lib-commons/v2/commons/mongo"
)

// Service is the application glue where we put all top level components to be used.
type Service struct {
	*MultiQueueConsumer
	log.Logger
	healthChecker   *pkg.HealthChecker
	mongoConnection *libMongo.MongoConnection
}

// Run starts the application.
// This is the only necessary code to run an app in main.go
func (app *Service) Run() {
	commons.NewLauncher(
		commons.WithLogger(app.Logger),
		commons.RunApp("RabbitMQ Consumer", app.MultiQueueConsumer),
	).Run()

	// Graceful shutdown
	app.Info("Starting graceful shutdown...")

	// Stop health checker
	if app.healthChecker != nil {
		app.Info("Stopping health checker...")
		app.healthChecker.Stop()
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

	app.Info("Graceful shutdown complete")
}
