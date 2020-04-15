package byteconvert

import (
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
  ## Name of the field to source data from.
  ##
  ## The given field should contain a measurement represented in bytes.
  field_src = "total_net_usage_bytes"

  ## Name of the new field to contain the converted value
  field_name = "total_net_usage_mb"

  ## Unit to convert the source value into.
  ##
  ## Allowed values: KiB, MiB, GiB 
  convert_unit = "MiB"
`

type ByteConvert struct {
	FieldSrc    string `toml:"field_src"`
	FieldName   string `toml:"field_name"`
	ConvertUnit string `toml:"convert_unit"`
}

func (d *ByteConvert) SampleConfig() string {
	return sampleConfig
}

func (d *ByteConvert) Description() string {
	return "Convert a value in bytes to a configured unit."
}

func (d *ByteConvert) Init() error {
	return nil
}

func (d *ByteConvert) Apply(in ...telegraf.Metric) []telegraf.Metric {
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

func (d *ByteConvert) convert(bytes float64) float64 {
	switch strings.ToLower(d.ConvertUnit) {
	case "kib":
		return bytes / 1024
	case "mib":
		return bytes / 1024 / 1024
	case "gib":
		return bytes / 1024 / 1024 / 1024
	}
	return 0
}

func init() {
	processors.Add("byteconvert", func() telegraf.Processor {
		return &ByteConvert{}
	})
}
