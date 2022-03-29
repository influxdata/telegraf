//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
package cgroup

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type CGroup struct {
	Paths []string `toml:"paths"`
	Files []string `toml:"files"`
}

func (g *CGroup) SampleConfig() string {
	return `{{ .SampleConfig }}`
}

func init() {
	inputs.Add("cgroup", func() telegraf.Input { return &CGroup{} })
}
