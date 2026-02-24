// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
)

// Service is the application glue where we put all top-level components to be used.
type Service struct {
	*Server
	log.Logger
	cleanup func()
}

// Run starts the application.
// This is the only necessary code to run an app in the main.go
func (app *Service) Run() {
	commons.NewLauncher(
		commons.WithLogger(app.Logger),
		commons.RunApp("HTTP Service", app.Server),
	).Run()

	// Graceful shutdown
	app.Info("Starting graceful shutdown...")

	if app.cleanup != nil {
		app.cleanup()
	}

	app.Info("Graceful shutdown complete")
}
