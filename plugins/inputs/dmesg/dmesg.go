//go:build linux
// +build linux

package dmesg

import (
	"fmt"
	"os/exec"
	"regexp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type regStrMap struct {
	Filter string `toml:"filter"`
	Field  string `toml:"field"`
}

type realRegMap struct {
	Filter *regexp.Regexp
	Field  string
}

type DmesgConf struct {
	Filters []regStrMap `toml:"filters"`
	Binary  string      `toml:"dmesg_binary"`
	Options []string    `toml:"options"`
}

func (k *DmesgConf) Description() string {
	return "Return counts of specific regexes against dmesg"
}

var dmesgSampleConfig = `
	## some basic dmesg regexes
	filters = [{"filter": ".*oom_reaper.*|.*Out of memory.*", "field": "oom.count"},
			   {"filter": ".*Power-on or device reset occurred.*", "field": "device.reset"},
			   {"filter": ".*I/O error.*", "field": "io.error"},
			   {"filter": ".*MCE MEMORY.*", "field": "mce.memory.errors"}]
	dmesg_binary = "/usr/bin/dmesg"
	## CLI options for the dmesg binary (-T, -H, etc.)
	options = []
`

func (k *DmesgConf) SampleConfig() string {
	return dmesgSampleConfig
}

func (k *DmesgConf) Gather(acc telegraf.Accumulator) error {
	var realRegexes []realRegMap
	for _, re := range k.Filters {
		realRegexes = append(realRegexes, realRegMap{Filter: regexp.MustCompile(re.Filter),
			Field: re.Field})
	}
	output, err := exec.Command(k.Binary, k.Options...).Output()
	if err != nil {
		fmt.Errorf("Execution of dmesg binary failed: %s", k.Binary)
		return err
	}
	fields := make(map[string]interface{})
	for _, re := range realRegexes {
		results := re.Filter.FindAll(output, -1)
		fields[re.Field] = len(results)
	}
	acc.AddFields("dmesg", fields, map[string]string{})
	return nil
}

func init() {
	inputs.Add("dmesg", func() telegraf.Input { return &DmesgConf{} })
}
