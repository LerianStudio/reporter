// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package chaos

import (
	"testing"
	"time"
)

// Restarts MongoDB and Redis containers and validates recovery of the system
func TestChaos_Datastores_RestartAndRecover(t *testing.T) {
	if err := RestartMongoDB(5 * time.Second); err != nil {
		t.Fatalf("restart mongo: %v", err)
	}
	if err := RestartValkey(5 * time.Second); err != nil {
		t.Fatalf("restart redis/valkey: %v", err)
	}
}
