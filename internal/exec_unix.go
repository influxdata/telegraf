//go:build !windows

package internal

import (
	"log"
	"os/exec"
	"syscall"
	"time"
)

// KillGrace is the amount of time we allow a process to shutdown before
// sending a SIGKILL.
const KillGrace = 5 * time.Second

// WaitTimeout waits for the given command to finish with a timeout.
// It assumes the command has already been started.
// If the command times out, it attempts to kill the process.
func WaitTimeout(c *exec.Cmd, timeout time.Duration) error {
	var kill *time.Timer
	term := time.AfterFunc(timeout, func() {
		err := syscall.Kill(-c.Process.Pid, syscall.SIGTERM)
		if err != nil {
			log.Printf("E! [agent] Error terminating process children: %s", err)
		}
		err = c.Process.Signal(syscall.SIGTERM)
		if err != nil {
			log.Printf("E! [agent] Error terminating process: %s", err)
			return
		}

		kill = time.AfterFunc(KillGrace, func() {
			err := syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
			if err != nil {
				log.Printf("E! [agent] Error terminating process children: %s", err)
			}
			err = c.Process.Kill()
			if err != nil {
				log.Printf("E! [agent] Error killing process: %s", err)
				return
			}
		})
	})

	err := c.Wait()

	// Shutdown all timers
	if kill != nil {
		kill.Stop()
	}
	termSent := !term.Stop()

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
