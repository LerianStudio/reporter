package main

import (
	"k8s-golang-addons-boilerplate/internal/bootstrap"
	"k8s-golang-addons-boilerplate/pkg"
)

// @title					K8s addons boilerplate
// @version					1.0.0
// @description				This is a swagger documentation for K8s addons boilerplate
// @termsOfService			http://swagger.io/terms/
// @host					localhost:4000
// @BasePath					/
func main() {
	pkg.InitLocalEnvConfig()
	bootstrap.InitServers().Run()
}
