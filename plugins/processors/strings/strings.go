package strings

import (
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Strings struct {
    Lowercase  []converter
    Uppercase  []converter
    Trim       []converter
    TrimLeft   []converter
    TrimRight  []converter
    TrimPrefix []converter
    TrimSuffix []converter
}

type converter struct {
	Tag         string
    Field       string
	ResultKey   string
    Argument    string
}

const sampleConfig = `
  ## Tag and field conversions defined in a separate sub-tables

  # [[processors.strings.uppercase]]
  #   tag = "method"

  # [[processors.strings.lowercase]]
  #   field = "uri_stem"
  #   result_key = "uri_stem_normalised"
`

func (r *Strings) SampleConfig() string {
	return sampleConfig
}

func (r *Strings) Description() string {
	return "Transforms tag and field values to lower case"
}

func ApplyFunction(
        metric telegraf.Metric,
        c converter,
        fn func(string) string) {

    if value, ok := metric.Tags()[c.Tag]; ok {
        metric.AddTag(
            getKey(c),
            fn(value),
        )
    } else if value, ok := metric.Fields()[c.Field]; ok {
        switch value := value.(type) {
        case string:
            metric.AddField(
                getKey(c),
                fn(value),
            )
        }
    }
}

func (r *Strings) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		for _, converter := range r.Lowercase {
            ApplyFunction(metric, converter, strings.ToLower)
		}
		for _, converter := range r.Uppercase {
            ApplyFunction(metric, converter, strings.ToUpper)
		}
        for _, converter := range r.Trim {
            ApplyFunction(metric, converter,
                func(s string) string { return strings.Trim(s, converter.Argument) })
        }
        for _, converter := range r.TrimPrefix {
            ApplyFunction(metric, converter,
                func(s string) string { return strings.TrimPrefix(s, converter.Argument) })
        }
        for _, converter := range r.TrimSuffix{
            ApplyFunction(metric, converter,
                func(s string) string { return strings.TrimSuffix(s, converter.Argument) })
        }
        for _, converter := range r.TrimRight {
            ApplyFunction(metric, converter,
                func(s string) string { return strings.TrimRight(s, converter.Argument) })
        }
        for _, converter := range r.TrimLeft {
            ApplyFunction(metric, converter,
                func(s string) string { return strings.TrimLeft(s, converter.Argument) })
        }
	}

	return in
}

func getKey(c converter) string {
	if c.ResultKey != "" {
		return c.ResultKey
	} else if c.Field != "" {
        return c.Field
    } else {
	    return c.Tag
    }
}

func init() {
	processors.Add("strings", func() telegraf.Processor {
		return &Strings{}
	})
}
