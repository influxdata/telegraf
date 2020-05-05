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
		if osStdin, ok := e.stdin.(*os.File); ok {
			osStdin.SetWriteDeadline(time.Now().Add(1 * time.Second))
		}
		if _, err := io.WriteString(e.stdin, "\n"); err != nil {
			return fmt.Errorf("Error writing to stdin: %s", err)
		}
	case "none":
	default:
		return fmt.Errorf("invalid signal: %s", e.Signal)
	}

	return nil
}
