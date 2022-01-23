//go:build !windows
// +build !windows

package process

import (
	"context"
	"os/exec"
	"syscall"
	"time"
)

func gracefulStop(ctx context.Context, cmd *exec.Cmd, timeout time.Duration) {
	select {
	case <-time.After(timeout):
		cmd.Process.Signal(syscall.SIGTERM)
	case <-ctx.Done():
	}
	select {
	case <-time.After(timeout):
		cmd.Process.Kill()
	case <-ctx.Done():
	}
}
