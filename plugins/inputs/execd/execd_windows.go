// +build windows

package execd

import (
	"fmt"
	"io"

	"github.com/influxdata/telegraf"
)

func (e *Execd) Gather(acc telegraf.Accumulator) error {
	if e.cmd == nil || e.cmd.Process == nil {
		return nil
	}

	switch e.Signal {
	case "STDIN":
		if _, err := io.WriteString(e.stdin, "\n"); err != nil {
			return fmt.Errorf("Error writing to stdin: %s", err)
		}
	case "none":
	default:
		return fmt.Errorf("invalid signal: %s", e.Signal)
	}

	return nil
}
