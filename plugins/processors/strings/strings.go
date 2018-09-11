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
`

func (s *Strings) SampleConfig() string {
	return sampleConfig
}

func (s *Strings) Description() string {
	return "Perform string processing on tags, fields, and measurements"
}

func (c *converter) convertTag(metric telegraf.Metric) {
	tv, ok := metric.GetTag(c.Tag)
	if !ok {
		return
	}

	dest := c.Tag
	if c.Dest != "" {
		dest = c.Dest
	}

	metric.AddTag(dest, c.fn(tv))
}

func (c *converter) convertField(metric telegraf.Metric) {
	fv, ok := metric.GetField(c.Field)
	if !ok {
		return
	}

	dest := c.Field
	if c.Dest != "" {
		dest = c.Dest
	}

	if fv, ok := fv.(string); ok {
		metric.AddField(dest, c.fn(fv))
	}
}

func (c *converter) convertMeasurement(metric telegraf.Metric) {
	if metric.Name() != c.Measurement {
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
		if c.Cutset != "" {
			c.fn = func(s string) string { return strings.Trim(s, c.Cutset) }
		} else {
			c.fn = func(s string) string { return strings.TrimFunc(s, unicode.IsSpace) }
		}
		s.converters = append(s.converters, c)
	}
	for _, c := range s.TrimLeft {
		if c.Cutset != "" {
			c.fn = func(s string) string { return strings.TrimLeft(s, c.Cutset) }
		} else {
			c.fn = func(s string) string { return strings.TrimLeftFunc(s, unicode.IsSpace) }
		}
		s.converters = append(s.converters, c)
	}
	for _, c := range s.TrimRight {
		if c.Cutset != "" {
			c.fn = func(s string) string { return strings.TrimRight(s, c.Cutset) }
		} else {
			c.fn = func(s string) string { return strings.TrimRightFunc(s, unicode.IsSpace) }
		}
		s.converters = append(s.converters, c)
	}
	for _, c := range s.TrimPrefix {
		c.fn = func(s string) string { return strings.TrimPrefix(s, c.Prefix) }
		s.converters = append(s.converters, c)
	}
	for _, c := range s.TrimSuffix {
		c.fn = func(s string) string { return strings.TrimSuffix(s, c.Suffix) }
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
