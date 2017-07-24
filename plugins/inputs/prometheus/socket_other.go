// +build !darwin,!freebsd,!linux,!netbsd,!openbsd

package prometheus

import (
	"os"
	"path/filepath"

	"github.com/influxdata/telegraf"
)

// harvestSocket dummy implementation for non UNIX systems where we dont have any sockets
func (p *Prometheus) harvestSocket(acc telegraf.Accumulator) filepath.WalkFunc {
	return func(file string, fileInfo os.FileInfo, err error) error {
		return nil
	}
}
