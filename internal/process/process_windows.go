// +build windows

package process

import (
	"os/exec"
	"time"
)

func gracefulStop(cmd *exec.Cmd, timeout time.Duration, processQuit chan struct{}) {
	select {
	case <-processQuit:
		return
	case <-time.After(timeout):
	}

	cmd.Process.Kill()
}
