package cgroup

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type CGroup struct {
	Paths []string `toml:"paths"`
	Files []string `toml:"files"`
}

var sampleConfig = `
  ## Directories in which to look for files, globs are supported.
  ## Consider restricting paths to the set of cgroups you really
  ## want to monitor if you have a large number of cgroups, to avoid
  ## any cardinality issues.
  # paths = [
  #   "/cgroup/memory",
  #   "/cgroup/memory/child1",
  #   "/cgroup/memory/child2/*",
  # ]
  ## cgroup stat fields, as file names, globs are supported.
  ## these file names are appended to each path from above.
  # files = ["memory.*usage*", "memory.limit_in_bytes"]
`

func (g *CGroup) SampleConfig() string {
	return sampleConfig
}

func (g *CGroup) Description() string {
	return "Read specific statistics per cgroup"
}

func init() {
	inputs.Add("cgroup", func() telegraf.Input { return &CGroup{} })
}
