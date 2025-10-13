package bootstrap

import (
	"plugin-smart-templates/v3/pkg"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libCommonsLicense "github.com/LerianStudio/lib-commons/v2/commons/license"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
)

// Service is the application glue where we put all top level components to be used.
type Service struct {
	*MultiQueueConsumer
	log.Logger
	licenseShutdown *libCommonsLicense.ManagerShutdown
	healthChecker   *pkg.HealthChecker
}

// Run starts the application.
// This is the only necessary code to run an app in main.go
func (app *Service) Run() {
	commons.NewLauncher(
		commons.WithLogger(app.Logger),
		commons.RunApp("RabbitMQ Consumer", app.MultiQueueConsumer),
	).Run()

	// Graceful shutdown
	app.Logger.Info("Starting graceful shutdown...")

	// Stop health checker
	if app.healthChecker != nil {
		app.Logger.Info("Stopping health checker...")
		app.healthChecker.Stop()
	}

	// After all consumers are done, shutdown license
	if app.licenseShutdown != nil {
		app.licenseShutdown.Terminate("Consumers are done.")
	}

	app.Logger.Info("Graceful shutdown complete")
}
