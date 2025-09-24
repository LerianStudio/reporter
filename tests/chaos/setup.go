package chaos

import (
	"os"
	"testing"
)

func shouldRunChaos(t *testing.T) bool {
	t.Helper()

	if v := os.Getenv("RUN_CHAOS"); v != "true" {
		t.Skip("chaos tests disabled; set RUN_CHAOS=true to enable")
		return false
	}

	return true
}

// noop reference to satisfy unused linters when analyzing files in isolation
var _ = shouldRunChaos
