//go:generate ../../../tools/readme_config_includer/generator
package cgroup

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type CGroup struct {
	Paths []string `toml:"paths"`
	Files []string `toml:"files"`
}

func (*CGroup) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("cgroup", func() telegraf.Input { return &CGroup{} })
}
