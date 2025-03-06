package bootstrap

import (
	"k8s-golang-addons-boilerplate/pkg"
	"k8s-golang-addons-boilerplate/pkg/log"
)

// Service is the application glue where we put all top level components to be used.
type Service struct {
	*Server
	*ServerGRPC
	log.Logger
}

// Run starts the application.
// This is the only necessary code to run an app in main.go
func (app *Service) Run() {
	pkg.NewLauncher(
		pkg.WithLogger(app.Logger),
		pkg.RunApp("HTTP Service", app.Server),
		pkg.RunApp("gRPC server", app.ServerGRPC),
	).Run()
}
