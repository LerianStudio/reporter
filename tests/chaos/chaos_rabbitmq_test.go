package chaos

import (
	"testing"
	"time"

	h "github.com/LerianStudio/reporter/v3/tests/helpers"
)

// Restarts RabbitMQ container and validates recovery of the system
func TestChaos_RabbitMQ_RestartAndRecover(t *testing.T) {
	env := h.LoadEnvironment()
	name := env.RabbitContainer
	if name == "" {
		name = "reporter-rabbitmq"
	}
	if err := h.RestartWithWait(name, 5*time.Second); err != nil {
		t.Fatalf("failed to restart rabbitmq: %v", err)
	}
}
