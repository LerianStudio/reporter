package main

import (
	"plugin-smart-templates/v2/components/manager/internal/bootstrap"
	"plugin-smart-templates/v2/pkg"
)

// @title			Plugin Smart Template
// @version		1.0.0
// @description	This is a swagger documentation for plugin smart template
// @termsOfService	http://swagger.io/terms/
// @host			localhost:4005
// @BasePath		/
func main() {
	pkg.InitLocalEnvConfig()
	bootstrap.InitServers().Run()
}
