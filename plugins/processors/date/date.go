package date

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
  ## New tag to create
  tag_key = "month"

  ## Date format string, must be a representation of the Go "reference time"
  ## which is "Mon Jan 2 15:04:05 -0700 MST 2006".
  date_format = "Jan"
`

type Date struct {
	TagKey     string `toml:"tag_key"`
	DateFormat string `toml:"date_format"`
}

func (d *Date) SampleConfig() string {
	return sampleConfig
}

func (d *Date) Description() string {
	return "Dates measurements, tags, and fields that pass through this filter."
}

func (d *Date) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		point.AddTag(d.TagKey, point.Time().Format(d.DateFormat))
	}

	return in
}

func init() {
	processors.Add("date", func() telegraf.Processor {
		return &Date{}
	})
}
