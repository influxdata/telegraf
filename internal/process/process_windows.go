//go:build windows

package process

import (
	"context"
	"os/exec"
	"time"
)

func (p *Process) gracefulStop(ctx context.Context, cmd *exec.Cmd, timeout time.Duration) {
	select {
	case <-time.After(timeout):
		if err := cmd.Process.Kill(); err != nil {
			p.Log.Errorf("Error after killing process: %v", err)
		}
	case <-ctx.Done():
	}
}
