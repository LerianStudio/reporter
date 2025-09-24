package chaos

import (
	"testing"
	"time"

	h "plugin-smart-templates/v2/tests/helpers"
)

// Restarts RabbitMQ container and validates recovery of the system
func TestChaos_RabbitMQ_RestartAndRecover(t *testing.T) {
	shouldRunChaos(t)
	env := h.LoadEnvironment()
	name := env.RabbitContainer
	if name == "" {
		name = "smart-templates-rabbitmq"
	}
	if err := h.RestartWithWait(name, 5*time.Second); err != nil {
		t.Fatalf("failed to restart rabbitmq: %v", err)
	}
}
