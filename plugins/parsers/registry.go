package parsers

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/influxdata/telegraf/plugins/parsers/graphite"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/plugins/parsers/ltsv"
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
	//
	// NOTE: For the LTSV parser, you need to call an additional `Parse(nil)`
	// if the last data does not end with the newline `\n`.
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
	// Dataformat can be one of: json, influx, graphite, ltsv
	DataFormat string

	// Separator only applied to Graphite data.
	Separator string
	// Templates only apply to Graphite data.
	Templates []string

	// TagKeys only apply to JSON data
	TagKeys []string
	// MetricName only applies to JSON data and LTSV data. This will be the name of the measurement.
	MetricName string

	// TimeLabel only applies to LTSV data. This will be the label of the timestamp.
	// If this label is not found in the measurement, the current time will be used.
	TimeLabel string
	// TimeFormat only applies to LTSV data. This will be the format of the timestamp.
	// Please see https://golang.org/pkg/time/#Parse for the format string.
	TimeFormat string
	// StrFieldLabels only applies to LTSV data. This will be the labels of string fields.
	StrFieldLabels []string
	// IntFieldLabels only applies to LTSV data. This will be the labels of integer fields.
	IntFieldLabels []string
	// FloatFieldLabels only applies to LTSV data. This will be the labels of float fields.
	FloatFieldLabels []string
	// BoolFieldLabels only applies to LTSV data. This will be the labels of boolean fields.
	BoolFieldLabels []string
	// TagLabels only applies to LTSV data. This will be the labels of tags.
	TagLabels []string
	// DuplicatePointsModifierMethod only applies to LTSV data.
	// Must be one of "add_uniq_tag", "increment_time", "no_op".
	// This will be used to modify duplicated points.
	// For detail, please see https://docs.influxdata.com/influxdb/v0.10/troubleshooting/frequently_encountered_issues/#writing-duplicate-points
	// NOTE: For modifier methods other than "no_op" to work correctly, the log lines
	// MUST be sorted by timestamps in ascending order.
	DuplicatePointsModifierMethod string
	// DuplicatePointsIncrementDuration only applies to LTSV data.
	// When duplicate_points_modifier_method is "increment_time",
	// this will be added to the time of the previous measurement
	// if the time of current time is equal to or less than the
	// time of the previous measurement.
	//
	// NOTE: You need to set this value equal to or greater than
	// precisions of your output plugins. Otherwise the times will
	// become the same value!
	// For the precision of the InfluxDB plugin, please see
	// https://github.com/influxdata/telegraf/blob/v0.10.1/plugins/outputs/influxdb/influxdb.go#L40-L42
	DuplicatePointsIncrementDuration time.Duration
	// DuplicatePointsModifierUniqTag only applies to LTSV data.
	// When DuplicatePointsModifierMethod is one of "add_uniq_tag",
	// this will be the label of the tag to be added to ensure uniqueness of points.
	// NOTE: The uniq tag will be only added to the successive points of duplicated
	// points, it will not be added to the first point of duplicated points.
	// If you want to always add the uniq tag, add a tag with the same name as
	// DuplicatePointsModifierUniqTag and the string value "0" to DefaultTags.
	DuplicatePointsModifierUniqTag string

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
	case "influx":
		parser, err = NewInfluxParser()
	case "graphite":
		parser, err = NewGraphiteParser(config.Separator,
			config.Templates, config.DefaultTags)
	case "ltsv":
		parser, err = NewLTSVParser(
			config.MetricName,
			config.TimeLabel,
			config.TimeFormat,
			config.StrFieldLabels,
			config.IntFieldLabels,
			config.FloatFieldLabels,
			config.BoolFieldLabels,
			config.TagLabels,
			config.DuplicatePointsModifierMethod,
			config.DuplicatePointsIncrementDuration,
			config.DuplicatePointsModifierUniqTag,
			config.DefaultTags,
		)
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

func NewLTSVParser(
	metricName string,
	timeLabel string,
	timeFormat string,
	strFieldLabels []string,
	intFieldLabels []string,
	floatFieldLabels []string,
	boolFieldLabels []string,
	tagLabels []string,
	duplicatePointsModifierMethod string,
	duplicatePointsIncrementDuration time.Duration,
	duplicatePointsModifierUniqTag string,
	defaultTags map[string]string,
) (Parser, error) {
	parser := &ltsv.LTSVParser{
		MetricName:                       metricName,
		TimeLabel:                        timeLabel,
		TimeFormat:                       timeFormat,
		StrFieldLabels:                   strFieldLabels,
		IntFieldLabels:                   intFieldLabels,
		FloatFieldLabels:                 floatFieldLabels,
		BoolFieldLabels:                  boolFieldLabels,
		TagLabels:                        tagLabels,
		DuplicatePointsModifierMethod:    duplicatePointsModifierMethod,
		DuplicatePointsIncrementDuration: duplicatePointsIncrementDuration,
		DuplicatePointsModifierUniqTag:   duplicatePointsModifierUniqTag,
		DefaultTags:                      defaultTags,
	}
	return parser, nil
}
