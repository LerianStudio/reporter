package main

import (
	libCommons "github.com/LerianStudio/lib-commons/commons"
	"plugin-template-engine/components/worker/internal/bootstrap"
)

func main() {
	libCommons.InitLocalEnvConfig()
	bootstrap.InitWorker().Run()
}
