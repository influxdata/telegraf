package parsers

import (
	"fmt"

	"github.com/influxdata/telegraf"

	"github.com/influxdata/telegraf/plugins/parsers/graphite"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/plugins/parsers/nagios"
	"github.com/influxdata/telegraf/plugins/parsers/value"
)

// ParserInput is an interface for input plugins that are able to parse
// arbitrary data formats.
type ParserInput interface {
	// SetParser sets the parser function for the interface
	SetParser(parser Parser)
}

// Parser is an interface defining functions that a parser plugin must satisfy.
type Parser interface {
	// Parse takes a byte buffer separated by newlines
	// ie, `cpu.usage.idle 90\ncpu.usage.busy 10`
	// and parses it into telegraf metrics
	Parse(buf []byte) ([]telegraf.Metric, error)

	// ParseLine takes a single string metric
	// ie, "cpu.usage.idle 90"
	// and parses it into a telegraf metric.
	ParseLine(line string) (telegraf.Metric, error)

	// SetDefaultTags tells the parser to add all of the given tags
	// to each parsed metric.
	// NOTE: do _not_ modify the map after you've passed it here!!
	SetDefaultTags(tags map[string]string)
}

// Config is a struct that covers the data types needed for all parser types,
// and can be used to instantiate _any_ of the parsers.
type Config struct {
	// Dataformat can be one of: json, influx, graphite, value, nagios
	DataFormat string

	// Separator only applied to Graphite data.
	Separator string
	// Templates only apply to Graphite data.
	Templates []string

	// TagKeys only apply to JSON data
	TagKeys []string
	// MetricName applies to JSON & value. This will be the name of the measurement.
	MetricName string

	// DataType only applies to value, this will be the type to parse value to
	DataType string

	// DefaultTags are the default tags that will be added to all parsed metrics.
	DefaultTags map[string]string
}

// NewParser returns a Parser interface based on the given config.
func NewParser(config *Config) (Parser, error) {
	var err error
	var parser Parser
	switch config.DataFormat {
	case "json":
		parser, err = NewJSONParser(config.MetricName,
			config.TagKeys, config.DefaultTags)
	case "value":
		parser, err = NewValueParser(config.MetricName,
			config.DataType, config.DefaultTags)
	case "influx":
		parser, err = NewInfluxParser()
	case "nagios":
		parser, err = NewNagiosParser()
	case "graphite":
		parser, err = NewGraphiteParser(config.Separator,
			config.Templates, config.DefaultTags)
	default:
		err = fmt.Errorf("Invalid data format: %s", config.DataFormat)
	}
	return parser, err
}

func NewJSONParser(
	metricName string,
	tagKeys []string,
	defaultTags map[string]string,
) (Parser, error) {
	parser := &json.JSONParser{
		MetricName:  metricName,
		TagKeys:     tagKeys,
		DefaultTags: defaultTags,
	}
	return parser, nil
}

func NewNagiosParser() (Parser, error) {
	return &nagios.NagiosParser{}, nil
}

func NewInfluxParser() (Parser, error) {
	return &influx.InfluxParser{}, nil
}

func NewGraphiteParser(
	separator string,
	templates []string,
	defaultTags map[string]string,
) (Parser, error) {
	return graphite.NewGraphiteParser(separator, templates, defaultTags)
}

func NewValueParser(
	metricName string,
	dataType string,
	defaultTags map[string]string,
) (Parser, error) {
	return &value.ValueParser{
		MetricName:  metricName,
		DataType:    dataType,
		DefaultTags: defaultTags,
	}, nil
}
