// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"github.com/LerianStudio/reporter/v4/pkg"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
)

// Service is the application glue where we put all top level components to be used.
type Service struct {
	*MultiQueueConsumer
	log.Logger
	healthChecker *pkg.HealthChecker
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

	app.Info("Graceful shutdown complete")
}
