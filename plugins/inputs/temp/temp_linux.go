//go:build linux
// +build linux

package temp

import (
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
)

func (t *Temperature) Gather(acc telegraf.Accumulator) error {
	temps, err := t.ps.Temperature()
	if err != nil {
		if strings.Contains(err.Error(), "not implemented yet") {
			return fmt.Errorf("plugin is not supported on this platform: %v", err)
		}
		return fmt.Errorf("error getting temperatures info: %s", err)
	}
	for _, temp := range temps {
		acc.AddFields(
			"temp",
			map[string]interface{}{"temp": temp.Critical},
			map[string]string{"sensor": temp.SensorKey + "_crit"},
		)
		acc.AddFields(
			"temp",
			map[string]interface{}{"temp": temp.High},
			map[string]string{"sensor": temp.SensorKey + "_max"},
		)
		acc.AddFields(
			"temp",
			map[string]interface{}{"temp": temp.Temperature},
			map[string]string{"sensor": temp.SensorKey + "_input"},
		)
	}
	return nil
}
