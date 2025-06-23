package bootstrap

import (
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/log"
	"github.com/LerianStudio/lib-commons/commons/shutdown"
)

// Service is the application glue where we put all top level components to be used.
type Service struct {
	*MultiQueueConsumer
	log.Logger
	licenseShutdown *shutdown.LicenseManagerShutdown
}

// Run starts the application.
// This is the only necessary code to run an app in main.go
func (app *Service) Run() {
	commons.NewLauncher(
		commons.WithLogger(app.Logger),
		commons.RunApp("RabbitMQ Consumer", app.MultiQueueConsumer),
	).Run()

	// After all consumers are done, shutdown license
	if app.licenseShutdown != nil {
		app.licenseShutdown.Terminate("Consumers are done.")
	}
}
