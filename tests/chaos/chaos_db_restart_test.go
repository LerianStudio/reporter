// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package chaos

import (
	"testing"
	"time"

	h "github.com/LerianStudio/reporter/v4/tests/helpers"
)

// Restarts MongoDB and Redis containers and validates recovery of the system
func TestChaos_Datastores_RestartAndRecover(t *testing.T) {
	env := h.LoadEnvironment()
	if err := h.RestartWithWait(env.MongoContainer, 5*time.Second); err != nil {
		t.Fatalf("restart mongo: %v", err)
	}
	if err := h.RestartWithWait(env.RedisContainer, 5*time.Second); err != nil {
		t.Fatalf("restart redis: %v", err)
	}
}
