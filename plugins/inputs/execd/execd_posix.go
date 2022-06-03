//go:build !windows
// +build !windows

package execd

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/influxdata/telegraf"
)

func (e *Execd) Gather(_ telegraf.Accumulator) error {
	if e.process == nil || e.process.Cmd == nil {
		return nil
	}

	osProcess := e.process.Cmd.Process
	if osProcess == nil {
		return nil
	}
	switch e.Signal {
	case "SIGHUP":
		return osProcess.Signal(syscall.SIGHUP)
	case "SIGUSR1":
		return osProcess.Signal(syscall.SIGUSR1)
	case "SIGUSR2":
		return osProcess.Signal(syscall.SIGUSR2)
	case "STDIN":
		if osStdin, ok := e.process.Stdin.(*os.File); ok {
			if err := osStdin.SetWriteDeadline(time.Now().Add(1 * time.Second)); err != nil {
				return fmt.Errorf("setting write deadline failed: %s", err)
			}
		}
		if _, err := io.WriteString(e.process.Stdin, "\n"); err != nil {
			return fmt.Errorf("writing to stdin failed: %s", err)
		}
	case "none":
	default:
		return fmt.Errorf("invalid signal: %s", e.Signal)
	}

	return nil
}
