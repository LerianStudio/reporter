package main

import (
	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	"plugin-smart-templates/v2/components/worker/internal/bootstrap"
)

func main() {
	libCommons.InitLocalEnvConfig()
	bootstrap.InitWorker().Run()
}
