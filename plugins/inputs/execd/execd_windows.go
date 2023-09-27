//go:build windows

package execd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/influxdata/telegraf"
)

func (e *Execd) Gather(_ telegraf.Accumulator) error {
	if e.process == nil {
		return nil
	}

	switch e.Signal {
	case "STDIN":
		if osStdin, ok := e.process.Stdin.(*os.File); ok {
			if err := osStdin.SetWriteDeadline(time.Now().Add(1 * time.Second)); err != nil {
				if !errors.Is(err, os.ErrNoDeadline) {
					return fmt.Errorf("setting write deadline failed: %w", err)
				}
			}
		}
		if _, err := io.WriteString(e.process.Stdin, "\n"); err != nil {
			return fmt.Errorf("error writing to stdin: %w", err)
		}
	case "none":
	default:
		return fmt.Errorf("invalid signal: %s", e.Signal)
	}

	return nil
}
