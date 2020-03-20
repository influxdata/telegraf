package byteconvert

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"strings"
)

const sampleConfig = `
  ## field to convert
  field_src = "total_net_usage_bytes"

  ## new field to create
  field_name = "total_net_usage_mb"

  # format to convert to
  format = "mb"
`


type byteConvert struct {
	FieldSrc     string            `toml:"field_src"`
	FieldName     string            `toml:"field_name"`
	Format string            `toml:"format"`

}

func (d *byteConvert) SampleConfig() string {
	return sampleConfig
}

func (d *byteConvert) Description() string {
	return "Dates measurements, tags, and fields that pass through this filter."
}

func (d *byteConvert) Init() error {
	return nil
}

func (d *byteConvert) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		if field, ok := point.GetField(d.FieldSrc); ok {
			v := d.convert(toFloat64(field))
			point.AddField(d.FieldName, v)
		}

	}

	return in
}

func toFloat64(v interface{}) float64 {
	switch i := v.(type) {
	case float64:
		return i
	case float32:
		return float64(i)
	case int64:
		return float64(i)
	case int32:
		return float64(i)
	case int:
		return float64(i)
	}
	return 0
}

func (d *byteConvert) convert(bytes float64) float64{
	switch strings.ToLower(d.Format) {
	case "kb":
		return bytes/1024
	case "mb":
		return bytes/1024/1024
	case "gb":
		return bytes/1024/1024
	}
	return 0
}

func init() {
	processors.Add("byteconvert", func() telegraf.Processor {
		return &byteConvert{}
	})
}
