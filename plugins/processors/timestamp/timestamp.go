package timestamp

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
  ## New tag to create
  field_key = "timestamp"
`

type Timestamp struct {
	FieldKey string `toml:"tag_key"`
}

func (d *Timestamp) SampleConfig() string {
	return sampleConfig
}

func (d *Timestamp) Description() string {
	return "Add unix nano timestamp to metrics."
}

func (d *Timestamp) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		point.AddField(d.FieldKey, point.Time().UnixNano())
	}

	return in
}

func init() {
	processors.Add("timestamp", func() telegraf.Processor {
		return &Timestamp{}
	})
}
