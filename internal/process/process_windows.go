//go:build windows
// +build windows

package process

import (
	"context"
	"os/exec"
	"time"
)

func gracefulStop(ctx context.Context, cmd *exec.Cmd, timeout time.Duration) {
	select {
	case <-time.After(timeout):
		cmd.Process.Kill()
	case <-ctx.Done():
	}
}
