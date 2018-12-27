package strings

import (
	"strings"
	"unicode"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Strings struct {
	Lowercase  []converter `toml:"lowercase"`
	Uppercase  []converter `toml:"uppercase"`
	Trim       []converter `toml:"trim"`
	TrimLeft   []converter `toml:"trim_left"`
	TrimRight  []converter `toml:"trim_right"`
	TrimPrefix []converter `toml:"trim_prefix"`
	TrimSuffix []converter `toml:"trim_suffix"`
	Replace    []converter `toml:"replace"`

	converters []converter
	init       bool
}

type ConvertFunc func(s string) string

type converter struct {
	Field       string
	Tag         string
	Measurement string
	Dest        string
	Cutset      string
	Suffix      string
	Prefix      string
	Old         string
	New         string

	fn ConvertFunc
}

const sampleConfig = `
  ## Convert a tag value to uppercase
  # [[processors.strings.uppercase]]
  #   tag = "method"

  ## Convert a field value to lowercase and store in a new field
  # [[processors.strings.lowercase]]
  #   field = "uri_stem"
  #   dest = "uri_stem_normalised"

  ## Trim leading and trailing whitespace using the default cutset
  # [[processors.strings.trim]]
  #   field = "message"

  ## Trim leading characters in cutset
  # [[processors.strings.trim_left]]
  #   field = "message"
  #   cutset = "\t"

  ## Trim trailing characters in cutset
  # [[processors.strings.trim_right]]
  #   field = "message"
  #   cutset = "\r\n"

  ## Trim the given prefix from the field
  # [[processors.strings.trim_prefix]]
  #   field = "my_value"
  #   prefix = "my_"

  ## Trim the given suffix from the field
  # [[processors.strings.trim_suffix]]
  #   field = "read_count"
  #   suffix = "_count"

  ## Replace all non-overlapping instances of old with new
  # [[processors.strings.replace]]
  #   measurement = "*"
  #   old = ":"
  #   new = "_"
`

func (s *Strings) SampleConfig() string {
	return sampleConfig
}

func (s *Strings) Description() string {
	return "Perform string processing on tags, fields, and measurements"
}

func (c *converter) convertTag(metric telegraf.Metric) {
	var tags map[string]string
	if c.Tag == "*" {
		tags = metric.Tags()
	} else {
		tags = make(map[string]string)
		tv, ok := metric.GetTag(c.Tag)
		if !ok {
			return
		}
		tags[c.Tag] = tv
	}

	for key, value := range tags {
		dest := key
		if c.Tag != "*" && c.Dest != "" {
			dest = c.Dest
		}
		metric.AddTag(dest, c.fn(value))
	}
}

func (c *converter) convertField(metric telegraf.Metric) {
	var fields map[string]interface{}
	if c.Field == "*" {
		fields = metric.Fields()
	} else {
		fields = make(map[string]interface{})
		fv, ok := metric.GetField(c.Field)
		if !ok {
			return
		}
		fields[c.Field] = fv
	}

	for key, value := range fields {
		dest := key
		if c.Field != "*" && c.Dest != "" {
			dest = c.Dest
		}
		if fv, ok := value.(string); ok {
			metric.AddField(dest, c.fn(fv))
		}
	}
}

func (c *converter) convertMeasurement(metric telegraf.Metric) {
	if metric.Name() != c.Measurement && c.Measurement != "*" {
		return
	}

	metric.SetName(c.fn(metric.Name()))
}

func (c *converter) convert(metric telegraf.Metric) {
	if c.Field != "" {
		c.convertField(metric)
	}

	if c.Tag != "" {
		c.convertTag(metric)
	}

	if c.Measurement != "" {
		c.convertMeasurement(metric)
	}
}

func (s *Strings) initOnce() {
	if s.init {
		return
	}

	s.converters = make([]converter, 0)
	for _, c := range s.Lowercase {
		c.fn = strings.ToLower
		s.converters = append(s.converters, c)
	}
	for _, c := range s.Uppercase {
		c.fn = strings.ToUpper
		s.converters = append(s.converters, c)
	}
	for _, c := range s.Trim {
		c := c
		if c.Cutset != "" {
			c.fn = func(s string) string { return strings.Trim(s, c.Cutset) }
		} else {
			c.fn = func(s string) string { return strings.TrimFunc(s, unicode.IsSpace) }
		}
		s.converters = append(s.converters, c)
	}
	for _, c := range s.TrimLeft {
		c := c
		if c.Cutset != "" {
			c.fn = func(s string) string { return strings.TrimLeft(s, c.Cutset) }
		} else {
			c.fn = func(s string) string { return strings.TrimLeftFunc(s, unicode.IsSpace) }
		}
		s.converters = append(s.converters, c)
	}
	for _, c := range s.TrimRight {
		c := c
		if c.Cutset != "" {
			c.fn = func(s string) string { return strings.TrimRight(s, c.Cutset) }
		} else {
			c.fn = func(s string) string { return strings.TrimRightFunc(s, unicode.IsSpace) }
		}
		s.converters = append(s.converters, c)
	}
	for _, c := range s.TrimPrefix {
		c := c
		c.fn = func(s string) string { return strings.TrimPrefix(s, c.Prefix) }
		s.converters = append(s.converters, c)
	}
	for _, c := range s.TrimSuffix {
		c := c
		c.fn = func(s string) string { return strings.TrimSuffix(s, c.Suffix) }
		s.converters = append(s.converters, c)
	}
	for _, c := range s.Replace {
		c := c
		c.fn = func(s string) string {
			newString := strings.Replace(s, c.Old, c.New, -1)
			if newString == "" {
				return s
			} else {
				return newString
			}
		}
		s.converters = append(s.converters, c)
	}

	s.init = true
}

func (s *Strings) Apply(in ...telegraf.Metric) []telegraf.Metric {
	s.initOnce()

	for _, metric := range in {
		for _, converter := range s.converters {
			converter.convert(metric)
		}
	}

	return in
}

func init() {
	processors.Add("strings", func() telegraf.Processor {
		return &Strings{}
	})
}
