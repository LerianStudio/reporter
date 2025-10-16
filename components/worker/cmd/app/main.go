package main

import (
	"github.com/LerianStudio/reporter/v3/components/worker/internal/bootstrap"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
)

func main() {
	libCommons.InitLocalEnvConfig()
	bootstrap.InitWorker().Run()
}
