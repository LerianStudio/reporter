package bootstrap

import (
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/LerianStudio/lib-commons/commons/log"
)

// Service is the application glue where we put all top-level components to be used.
type Service struct {
	*Server
	log.Logger
}

// Run starts the application.
// This is the only necessary code to run an app in the main.go
func (app *Service) Run() {
	commons.NewLauncher(
		commons.WithLogger(app.Logger),
		commons.RunApp("HTTP Service", app.Server),
	).Run()
}
