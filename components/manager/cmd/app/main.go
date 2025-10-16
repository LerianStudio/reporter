package main

import (
	"github.com/LerianStudio/reporter/v3/components/manager/internal/bootstrap"
	"github.com/LerianStudio/reporter/v3/pkg"
)

// @title			Reporter
// @version		3.0.0
// @description	This is a swagger documentation for Reporter
// @termsOfService	http://swagger.io/terms/
// @host			localhost:4005
// @BasePath		/
func main() {
	pkg.InitLocalEnvConfig()
	bootstrap.InitServers().Run()
}
