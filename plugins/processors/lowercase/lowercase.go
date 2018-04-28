package lowercase

import (
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Lowercase struct {
	Tags   []converter
	Fields []converter
}

type converter struct {
	Key         string
	ResultKey   string
}

const sampleConfig = `
  ## Tag and field conversions defined in a separate sub-tables
  # [[processors.lowercase.tags]]
  #   ## Tag to change
  #   key = "method"

  # [[processors.lowercase.fields]]
  #   key = "uri_stem"
  #   result_key = "uri_stem_normalised"
`

func (r *Lowercase) SampleConfig() string {
	return sampleConfig
}

func (r *Lowercase) Description() string {
	return "Transforms tag and field values to lower case"
}

func (r *Lowercase) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		for _, converter := range r.Tags {
			if value, ok := metric.Tags()[converter.Key]; ok {
				metric.AddTag(
					getKey(converter),
                    strings.ToLower(value),
				)
			}
		}

		for _, converter := range r.Fields {
			if value, ok := metric.Fields()[converter.Key]; ok {
				switch value := value.(type) {
				case string:
					metric.AddField(
						getKey(converter),
                        strings.ToLower(value),
					)
				}
			}
		}
	}

	return in
}

func getKey(c converter) string {
	if c.ResultKey != "" {
		return c.ResultKey
	}
	return c.Key
}

func init() {
	processors.Add("lowercase", func() telegraf.Processor {
		return &Lowercase{}
	})
}
