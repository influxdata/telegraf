//go:generate ../../../tools/readme_config_includer/generator
package systemd_units

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
)

type valueDef struct {
	valueName string
	valueMap  *map[string]int
}

// The following two maps configure the mapping of systemd properties to
// tags or values. The properties are automatically requested from
// `systemctl show`.
// If a vlaue has no valueMap, `atoi` is called on the value to convert it to
// an integer.
var tagMap = map[string]string{
	"Id":             "name",
	"LoadState":      "load",
	"ActiveState":    "active",
	"SubState":       "sub",
	"UnitFileState":  "uf_state",
	"UnitFilePreset": "uf_preset",
}

var valueMap = map[string]valueDef{
	"LoadState":         {valueName: "load_code", valueMap: &loadMap},
	"ActiveState":       {valueName: "active_code", valueMap: &activeMap},
	"SubState":          {valueName: "sub_code", valueMap: &subMap},
	"StatusErrno":       {valueName: "status_errno", valueMap: nil},
	"NRestarts":         {valueName: "restarts", valueMap: nil},
	"MemoryCurrent":     {valueName: "mem_current", valueMap: nil},
	"MemoryPeak":        {valueName: "mem_peak", valueMap: nil},
	"MemorySwapCurrent": {valueName: "swap_current", valueMap: nil},
	"MemorySwapPeak":    {valueName: "swap_peak", valueMap: nil},
	"MemoryAvailable":   {valueName: "mem_avail", valueMap: nil},
	"MainPID":           {valueName: "pid", valueMap: nil},
}

// Gather parses systemctl outputs and adds counters to the Accumulator
func parseShow(acc telegraf.Accumulator, buffer *bytes.Buffer) {
	scanner := bufio.NewScanner(buffer)

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	for scanner.Scan() {
		line := scanner.Text()

		// An empty line signals the start of the next unit
		if len(line) == 0 {
			// We need at least a "name" field. This prevents values from the
			// global information block (enabled by the --all switch) to be
			// shown as a unit.
			if _, ok := tags["name"]; ok {
				acc.AddFields(measurement, fields, tags)
			}

			tags = make(map[string]string)
			fields = make(map[string]interface{})

			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			acc.AddError(fmt.Errorf("error parsing line (expected key=value): %s", line))
			continue
		}

		// Map the tags
		if tagName, isTag := tagMap[key]; isTag {
			tags[tagName] = value
		}

		// Map the values
		if valueDef, isValue := valueMap[key]; isValue {
			// If a value map is set use it. If not, just try to convert the
			// value into an integer.
			if valueDef.valueMap != nil {
				code, ok := (*valueDef.valueMap)[value]
				if !ok {
					acc.AddError(fmt.Errorf("error parsing field '%s', value '%s' not in map", key, value))
					continue
				}

				fields[valueDef.valueName] = code
			} else {
				if value != "[not set]" {
					intVal, err := strconv.Atoi(value)
					if err != nil {
						acc.AddError(fmt.Errorf("error '%w' parsing field '%s'. Not an integer value", err, key))
						continue
					}
					fields[valueDef.valueName] = intVal
				}
			}
		}
	}

	// Add the last unit because the output does not contain a newline for this
	if _, ok := tags["name"]; ok {
		acc.AddFields(measurement, fields, tags)
	}
}

func getShowParameters(s *SystemdUnits) *[]string {
	// build parameters for systemctl call
	params := []string{"show"}
	// create patterns parameters if provided in config
	if s.Pattern == "" {
		params = append(params, "*")
	} else {
		psplit := strings.SplitN(s.Pattern, " ", -1)
		params = append(params, psplit...)
	}

	params = append(params, "--all", "--type="+s.UnitType)

	// add the fields we're interested in to the command line
	for property := range tagMap {
		params = append(params, fmt.Sprintf("--property=%s", property))
	}
	for property := range valueMap {
		// If a property exists within the tagMap it was already added. Do not add it again to
		// keep the command line short.
		if _, exists := tagMap[property]; !exists {
			params = append(params, fmt.Sprintf("--property=%s", property))
		}
	}

	return &params
}

func initSubcommandShow() *subCommandInfo {
	return &subCommandInfo{
		getParameters: getShowParameters,
		parseResult:   parseShow,
	}
}
