package chaos

import (
	"testing"
	"time"

	h "plugin-smart-templates/v2/tests/helpers"
)

// Restarts MongoDB and Redis containers and validates recovery of the system
func TestChaos_Datastores_RestartAndRecover(t *testing.T) {
	shouldRunChaos(t)
	env := h.LoadEnvironment()
	if err := h.RestartWithWait(env.MongoContainer, 5*time.Second); err != nil {
		t.Fatalf("restart mongo: %v", err)
	}
	if err := h.RestartWithWait(env.RedisContainer, 5*time.Second); err != nil {
		t.Fatalf("restart redis: %v", err)
	}
}
