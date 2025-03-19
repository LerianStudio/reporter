package main

import (
	"plugin-template-engine/components/template-reports/internal/bootstrap"
	"plugin-template-engine/pkg"
)

// @title					K8s addons boilerplate
// @version					1.0.0
// @description				This is a swagger documentation for K8s addons boilerplate
// @termsOfService			http://swagger.io/terms/
// @host					localhost:4009
// @BasePath					/
func main() {
	pkg.InitLocalEnvConfig()
	bootstrap.InitServers().Run()
}
