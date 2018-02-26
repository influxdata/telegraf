package regex

import (
	"regexp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Regex struct {
	Tags   []converter
	Fields []converter
}

type converter struct {
	Key         string
	Pattern     string
	Replacement string
	ResultKey   string
}

const sampleConfig = `
  ## Tag and field conversions defined in a separate sub-tables
  # [[processors.regex.tags]]
  #   ## Tag to change
  #   key = "resp_code"
  #   ## Regular expression to match on a tag value
  #   pattern = "^(\\d)\\d\\d$"
  #   ## Pattern for constructing a new value (${1} represents first subgroup)
  #   replacement = "${1}xx"

  # [[processors.regex.fields]]
  #   key = "request"
  #   ## All the power of the Go regular expressions available here
  #   ## For example, named subgroups
  #   pattern = "^/api(?P<method>/[\\w/]+)\\S*" 
  #   replacement = "${method}"
  #   ## If result_key is present, a new field will be created
  #   ## instead of changing existing field
  #   result_key = "method"

  ## Multiple conversions may be applied for one field sequentially
  ## Let's extract one more value
  # [[processors.regex.fields]]
  #   key = "request"
  #   pattern = ".*category=(\\w+).*"
  #   replacement = "${1}"
  #   result_key = "search_category"
`

func (r *Regex) SampleConfig() string {
	return sampleConfig
}

func (r *Regex) Description() string {
	return "Transforms tag and field values with regex pattern"
}

func (r *Regex) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		for _, converter := range r.Tags {
			if value, ok := metric.Tags()[converter.Key]; ok {
				metric.AddTag(
					getKey(converter),
					getValue(converter, value),
				)
			}
		}

		for _, converter := range r.Fields {
			if fieldValue, ok := metric.Fields()[converter.Key]; ok {
				switch fieldValue := fieldValue.(type) {
				case string:
					metric.AddField(
						getKey(converter),
						getValue(converter, fieldValue),
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

func getValue(c converter, value string) string {
	regex := regexp.MustCompile(c.Pattern)
	if c.ResultKey != "" && !regex.MatchString(value) {
		return ""
	}
	return regex.ReplaceAllString(value, c.Replacement)
}

func init() {
	processors.Add("regex", func() telegraf.Processor {
		return &Regex{}
	})
}
