// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pongo

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := RegisterAll(); err != nil {
		log.Fatalf("Failed to register pongo2 filters and tags: %v", err)
	}

	os.Exit(m.Run())
}
