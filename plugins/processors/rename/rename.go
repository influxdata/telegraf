package rename

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
  ## Measurement, tag, and field renamings are stored in separate sub-tables.
  ## Specify one sub-table per rename operation.
  # [[processors.rename.measurement]]
  #   ## measurement to change
  #   from = "kilobytes_per_second"
  #   to = "kbps"

  # [[processors.rename.tag]]
  #   ## tag to change
  #   from = "host"
  #   to = "hostname"

  # [[processors.rename.field]]
  #   ## field to change
  #   from = "lower"
  #   to = "min"

  # [[processors.rename.field]]
  #   ## field to change
  #   from = "upper"
  #   to = "max"
`

type renamer struct {
	From string
	To   string
}

type Rename struct {
	Measurement []renamer
	Tag         []renamer
	Field       []renamer
}

func (r *Rename) SampleConfig() string {
	return sampleConfig
}

func (r *Rename) Description() string {
	return "Rename measurements, tags, and fields that pass through this filter."
}

func (r *Rename) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		for _, measurementRenamer := range r.Measurement {
			if point.Name() == measurementRenamer.From {
				point.SetName(measurementRenamer.To)
				break
			}
		}

		for _, tagRenamer := range r.Tag {
			if value, ok := point.GetTag(tagRenamer.From); ok {
				point.RemoveTag(tagRenamer.From)
				point.AddTag(tagRenamer.To, value)
			}
		}

		for _, fieldRenamer := range r.Field {
			if value, ok := point.GetField(fieldRenamer.From); ok {
				point.RemoveField(fieldRenamer.From)
				point.AddField(fieldRenamer.To, value)
			}
		}
	}

	return in
}

func init() {
	processors.Add("rename", func() telegraf.Processor {
		return &Rename{}
	})
}
