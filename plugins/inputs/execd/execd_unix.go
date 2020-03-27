// +build !windows

package execd

import (
	"fmt"
	"io"
	"syscall"

	"github.com/influxdata/telegraf"
)

func (e *Execd) Gather(acc telegraf.Accumulator) error {
	if e.cmd == nil || e.cmd.Process == nil {
		return nil
	}

	switch e.Signal {
	case "SIGHUP":
		e.cmd.Process.Signal(syscall.SIGHUP)
	case "SIGUSR1":
		e.cmd.Process.Signal(syscall.SIGUSR1)
	case "SIGUSR2":
		e.cmd.Process.Signal(syscall.SIGUSR2)
	case "STDIN":
		io.WriteString(e.stdin, "\n")
	case "none":
	default:
		return fmt.Errorf("invalid signal: %s", e.Signal)
	}

	return nil
}
