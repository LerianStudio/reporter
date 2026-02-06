// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package main

import (
	libCommons "github.com/LerianStudio/lib-commons/v2/commons"

	"github.com/LerianStudio/reporter/components/manager/internal/bootstrap"
)

// @title			Reporter
// @version		1.0.0
// @description	This is a swagger documentation for Reporter
// @termsOfService	http://swagger.io/terms/
// @host			localhost:4005
// @BasePath		/
func main() {
	libCommons.InitLocalEnvConfig()
	bootstrap.InitServers().Run()
}
