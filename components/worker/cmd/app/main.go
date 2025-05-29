package main

import (
	libCommons "github.com/LerianStudio/lib-commons/commons"
	"plugin-smart-templates/components/worker/internal/bootstrap"
)

func main() {
	libCommons.InitLocalEnvConfig()
	bootstrap.InitWorker().Run()
}
