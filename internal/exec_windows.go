// +build windows

package internal

import (
	"log"
	"os/exec"
	"time"
)

// WaitTimeout waits for the given command to finish with a timeout.
// It assumes the command has already been started.
// If the command times out, it attempts to kill the process.
func WaitTimeout(c *exec.Cmd, timeout time.Duration) error {
	timer := time.AfterFunc(timeout, func() {
		err := c.Process.Kill()
		if err != nil {
			log.Printf("E! [agent] Error killing process: %s", err)
			return
		}
	})

	err := c.Wait()

	// Shutdown all timers
	termSent := !timer.Stop()

	// If the process exited without error treat it as success.  This allows a
	// process to do a clean shutdown on signal.
	if err == nil {
		return nil
	}

	// If SIGTERM was sent then treat any process error as a timeout.
	if termSent {
		return ErrTimeout
	}

	// Otherwise there was an error unrelated to termination.
	return err
}
