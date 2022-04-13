package cgroup

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type CGroup struct {
	Paths []string `toml:"paths"`
	Files []string `toml:"files"`
}

func init() {
	inputs.Add("cgroup", func() telegraf.Input { return &CGroup{} })
}
