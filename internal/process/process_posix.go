// +build !windows

package process

import (
	"os/exec"
	"syscall"
	"time"
)

func gracefulStop(cmd *exec.Cmd, timeout time.Duration) {
	time.AfterFunc(timeout, func() {
		if cmd == nil || cmd.ProcessState == nil {
			return
		}
		if !cmd.ProcessState.Exited() {
			cmd.Process.Signal(syscall.SIGTERM)
			time.AfterFunc(timeout, func() {
				if cmd == nil || cmd.ProcessState == nil {
					return
				}
				if !cmd.ProcessState.Exited() {
					cmd.Process.Kill()
				}
			})
		}
	})
}
