// +build !windows

package procval

import (
	"fmt"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Procval struct {
	// Files represents a map where the value is a path in the proc
	// file system and key represents the the field name for the value
	// read from that file.
	Files map[string]string `toml:"files"`
}

func (_ *Procval) Description() string {
	return "Read integer values from proc files"
}

var sampleConfig = `
  [inputs.procval.files]
    ## specify list of proc files to read
	# fieldName = /proc/path/to/procfile
	#
	# for example if you want to measure the available
	# entropy on the system:
	# entropy = "/proc/sys/kernel/random/entropy_avail"
`

func (_ *Procval) SampleConfig() string {
	return sampleConfig
}

func (p *Procval) Gather(acc telegraf.Accumulator) error {
	fields := map[string]interface{}{}
	for fieldName, procFile := range p.Files {
		lines, err := internal.ReadLines(procFile)
		if err != nil {
			return err
		}
		if len(lines) == 0 {
			return fmt.Errorf("could not read enought lines in %s", procFile)
		}
		value, err := strconv.Atoi(lines[0])
		if err != nil {
			return err
		}
		fields[fieldName] = value
	}

	acc.AddFields("procval", fields, map[string]string{})
	return nil
}

func init() {
	inputs.Add("procval", func() telegraf.Input {
		return &Procval{}
	})
}
