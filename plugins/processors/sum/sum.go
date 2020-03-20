package sum

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
  ## New field to create
  field_key = "total_net_usage"

  # fields to sum
  field_sum = ["bytes_recv", "bytes_sent"]
`


type Sum struct {
	FieldKey     string            `toml:"field_key"`
	FieldSum []string `toml:"field_sum"`

}

func (d *Sum) SampleConfig() string {
	return sampleConfig
}

func (d *Sum) Description() string {
	return "Dates measurements, tags, and fields that pass through this filter."
}

func (d *Sum) Init() error {
	return nil
}

func (d *Sum) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		var s int64

		for _, f := range d.FieldSum {
			if field, ok := point.GetField(f); ok {
				s += int64(field.(float64))
			}
		}

		point.AddField(d.FieldKey, s)
	}

	return in
}

func init() {
	processors.Add("sum", func() telegraf.Processor {
		return &Sum{}
	})
}
