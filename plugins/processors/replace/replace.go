package replace

import (
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Replace struct {
	Old string
	New string
}

var sampleConfig = `
  ## This plugin is used to replace substrings within field names to allow for
  ## different conventions between various input and output plugins. Some
  ## example usages are eliminating disallowed characters in field names or
  ## replacing separators between different separators
  #
  # [[processors.replace]]
  #   old = "_"
  #   new = "-"

  # [[processors.replace]]
  #   old = ":"
  #   new = "_"
`

func (r *Replace) SampleConfig() string {
	return sampleConfig
}

func (r *Replace) Description() string {
	return "Do a plain string replace for all tag and field names passing through."
}

func (r *Replace) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		newName := strings.Replace(metric.Name(), r.Old, r.New, -1)
		if metric.Name() != newName {
			metric.SetName(newName)
		}
	}
	return in
}

func init() {
	processors.Add("replace", func() telegraf.Processor {
		return &Replace{}
	})
}
