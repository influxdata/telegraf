package smartctl

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/influxdata/telegraf/internal"
)

func (s *Smartctl) scan() (map[string]string, error) {
	args := []string{"--json", "--scan"}
	cmd := execCommand(s.Path, args...)
	if s.UseSudo {
		cmd = execCommand("sudo", append([]string{"-n", s.Path}, args...)...)
	}
	out, err := internal.CombinedOutputTimeout(cmd, time.Duration(s.Timeout))
	if err != nil {
		return nil, fmt.Errorf("error running smartctl with %s: %w", args, err)
	}

	var scan smartctlScanJSON
	if err := json.Unmarshal(out, &scan); err != nil {
		return nil, fmt.Errorf("error unmarshalling smartctl scan output: %w", err)
	}

	devices := make(map[string]string, len(scan.Devices))
	for _, device := range scan.Devices {
		if s.deviceFilter.Match(device.Name) {
			devices[device.Name] = device.Type
		}
	}

	return devices, nil
}
