// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/LerianStudio/reporter/components/worker/internal/bootstrap"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
)

func main() {
	libCommons.InitLocalEnvConfig()

	svc, err := bootstrap.InitWorker()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize worker: %v\n", err)
		os.Exit(1)
	}

	svc.Run()
}
