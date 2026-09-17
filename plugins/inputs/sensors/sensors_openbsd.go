//go:build openbsd

package sensors

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
)

const cmd = "sysctl"

func (s *Sensors) Init() error {
	if s.path == "" {
		path, err := exec.LookPath(cmd)
		if err != nil {
			return fmt.Errorf("looking up %q failed: %w", cmd, err)
		}
		s.path = path
	}

	if s.run == nil {
		s.run = sysctlRunner
	}
	return s.initOpenBSD()
}

func (s *Sensors) Gather(acc telegraf.Accumulator) error {
	return s.gatherOpenBSD(acc)
}

func sysctlRunner(binary string, timeout config.Duration) (*bytes.Buffer, error) {
	cmd := exec.Command(binary, "hw.sensors")

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := internal.RunTimeout(cmd, time.Duration(timeout)); err != nil {
		return nil, fmt.Errorf("error running sysctl: %w", err)
	}

	return &out, nil
}
