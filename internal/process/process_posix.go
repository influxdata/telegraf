// +build !windows

package process

import (
	"os/exec"
	"syscall"
	"time"
)

func gracefulStop(cmd *exec.Cmd, timeout time.Duration, processQuit chan struct{}) {
	select {
	case <-processQuit:
		return
	case <-time.After(timeout):
	}

	cmd.Process.Signal(syscall.SIGTERM)
	select {
	case <-processQuit:
		return
	case <-time.After(timeout):
	}
	cmd.Process.Kill()
}
