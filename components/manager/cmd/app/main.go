// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/LerianStudio/reporter/v4/components/manager/internal/bootstrap"
	"github.com/LerianStudio/reporter/v4/pkg"
)

// @title			Reporter
// @version		4.0.0
// @description	This is a swagger documentation for Reporter
// @termsOfService	http://swagger.io/terms/
// @host			localhost:4005
// @BasePath		/
func main() {
	pkg.InitLocalEnvConfig()
	bootstrap.InitServers().Run()
}
