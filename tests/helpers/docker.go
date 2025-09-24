package helpers

import (
	"fmt"
	"os/exec"
	"time"
)

// RestartWithWait restarts a container by name and waits a small delay.
func RestartWithWait(container string, delay time.Duration) error {
	if container == "" {
		return fmt.Errorf("empty container name")
	}

	cmd := exec.Command("docker", "restart", container)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker restart %s: %v, out=%s", container, err, string(out))
	}

	if delay > 0 {
		time.Sleep(delay)
	}

	return nil
}
