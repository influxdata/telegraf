//go:generate ../../../tools/readme_config_includer/generator
package systemd_units

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
)

// Gather parses systemctl outputs and adds counters to the Accumulator
func parseListUnits(acc telegraf.Accumulator, buffer *bytes.Buffer) {
	scanner := bufio.NewScanner(buffer)
	for scanner.Scan() {
		line := scanner.Text()

		data := strings.Fields(line)
		if len(data) < 4 {
			acc.AddError(fmt.Errorf("parsing line failed (expected at least 4 fields): %s", line))
			continue
		}
		name := data[0]
		load := data[1]
		active := data[2]
		sub := data[3]
		tags := map[string]string{
			"name":   name,
			"load":   load,
			"active": active,
			"sub":    sub,
		}

		var (
			loadCode   int
			activeCode int
			subCode    int
			ok         bool
		)
		if loadCode, ok = loadMap[load]; !ok {
			acc.AddError(fmt.Errorf("parsing field 'load' failed, value not in map: %s", load))
			continue
		}
		if activeCode, ok = activeMap[active]; !ok {
			acc.AddError(fmt.Errorf("parsing field field 'active' failed, value not in map: %s", active))
			continue
		}
		if subCode, ok = subMap[sub]; !ok {
			acc.AddError(fmt.Errorf("parsing field field 'sub' failed, value not in map: %s", sub))
			continue
		}
		fields := map[string]interface{}{
			"load_code":   loadCode,
			"active_code": activeCode,
			"sub_code":    subCode,
		}

		acc.AddFields(measurement, fields, tags)
	}
}

func getListUnitsParameters(s *SystemdUnits) *[]string {
	// build parameters for systemctl call
	params := []string{"list-units"}
	// create patterns parameters if provided in config
	if s.Pattern != "" {
		psplit := strings.SplitN(s.Pattern, " ", -1)
		params = append(params, psplit...)
	}
	params = append(params,
		"--all",
		"--plain",
		"--type="+s.UnitType,
		"--no-legend",
	)

	return &params
}

func initSubcommandListUnits() *subCommandInfo {
	return &subCommandInfo{
		getParameters: getListUnitsParameters,
		parseResult:   parseListUnits,
	}
}
