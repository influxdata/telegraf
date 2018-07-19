package regex

import (
	"regexp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Regex struct {
	Tags       []converter
	Fields     []converter
	regexCache map[string]*regexp.Regexp
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

func NewRegex() *Regex {
	return &Regex{
		regexCache: make(map[string]*regexp.Regexp),
	}
}

func (r *Regex) SampleConfig() string {
	return sampleConfig
}

func (r *Regex) Description() string {
	return "Transforms tag and field values with regex pattern"
}

func (r *Regex) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		for _, converter := range r.Tags {
			if value, ok := metric.GetTag(converter.Key); ok {
				if key, newValue := r.convert(converter, value); newValue != "" {
					metric.AddTag(key, newValue)
				}
			}
		}

		for _, converter := range r.Fields {
			if value, ok := metric.GetField(converter.Key); ok {
				switch value := value.(type) {
				case string:
					if key, newValue := r.convert(converter, value); newValue != "" {
						metric.AddField(key, newValue)
					}
				}
			}
		}
	}

	return in
}

func (r *Regex) convert(c converter, src string) (string, string) {
	regex, compiled := r.regexCache[c.Pattern]
	if !compiled {
		regex = regexp.MustCompile(c.Pattern)
		r.regexCache[c.Pattern] = regex
	}

	value := ""
	if c.ResultKey == "" || regex.MatchString(src) {
		value = regex.ReplaceAllString(src, c.Replacement)
	}

	if c.ResultKey != "" {
		return c.ResultKey, value
	}

	return c.Key, value
}

func init() {
	processors.Add("regex", func() telegraf.Processor {
		return NewRegex()
	})
}
