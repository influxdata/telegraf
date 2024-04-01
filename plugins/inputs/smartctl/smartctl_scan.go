package smartctl

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/influxdata/telegraf/internal"
)

// This is here so we can override it during testing
var scanArgs = []string{"--json", "--scan"}

type scanDevice struct {
	Name string
	Type string
}

func (s *Smartctl) scan() ([]scanDevice, error) {
	cmd := execCommand(s.Path, scanArgs...)
	if s.UseSudo {
		cmd = execCommand("sudo", append([]string{"-n", s.Path}, scanArgs...)...)
	}
	out, err := internal.CombinedOutputTimeout(cmd, time.Duration(s.Timeout))
	if err != nil {
		return nil, fmt.Errorf("error running smartctl with %s: %w", scanArgs, err)
	}

	var scan smartctlScanJSON
	if err := json.Unmarshal(out, &scan); err != nil {
		return nil, fmt.Errorf("error unmarshalling smartctl scan output: %w", err)
	}

	devices := make([]scanDevice, 0)
	for _, device := range scan.Devices {
		if s.deviceFilter.Match(device.Name) {
			device := scanDevice{
				Name: device.Name,
				Type: device.Type,
			}
			devices = append(devices, device)
		}
	}

	return devices, nil
}
