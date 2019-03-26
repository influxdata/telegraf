package parsers

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers/collectd"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/plugins/parsers/dropwizard"
	"github.com/influxdata/telegraf/plugins/parsers/graphite"
	"github.com/influxdata/telegraf/plugins/parsers/grok"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/plugins/parsers/logfmt"
	"github.com/influxdata/telegraf/plugins/parsers/nagios"
	"github.com/influxdata/telegraf/plugins/parsers/value"
	"github.com/influxdata/telegraf/plugins/parsers/wavefront"
)

type ParserFunc func() (Parser, error)

// ParserInput is an interface for input plugins that are able to parse
// arbitrary data formats.
type ParserInput interface {
	// SetParser sets the parser function for the interface
	SetParser(parser Parser)
}

// ParserFuncInput is an interface for input plugins that are able to parse
// arbitrary data formats.
type ParserFuncInput interface {
	// GetParser returns a new parser.
	SetParserFunc(fn ParserFunc)
}

// Parser is an interface defining functions that a parser plugin must satisfy.
type Parser interface {
	// Parse takes a byte buffer separated by newlines
	// ie, `cpu.usage.idle 90\ncpu.usage.busy 10`
	// and parses it into telegraf metrics
	//
	// Must be thread-safe.
	Parse(buf []byte) ([]telegraf.Metric, error)

	// ParseLine takes a single string metric
	// ie, "cpu.usage.idle 90"
	// and parses it into a telegraf metric.
	//
	// Must be thread-safe.
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
	DataFormat string `toml:"data_format"`

	// Separator only applied to Graphite data.
	Separator string `toml:"separator"`
	// Templates only apply to Graphite data.
	Templates []string `toml:"templates"`

	// TagKeys only apply to JSON data
	TagKeys []string `toml:"tag_keys"`
	// FieldKeys only apply to JSON
	JSONStringFields []string `toml:"json_string_fields"`

	JSONNameKey string `toml:"json_name_key"`
	// MetricName applies to JSON & value. This will be the name of the measurement.
	MetricName string `toml:"metric_name"`

	// holds a gjson path for json parser
	JSONQuery string `toml:"json_query"`

	// key of time
	JSONTimeKey string `toml:"json_time_key"`

	// time format
	JSONTimeFormat string `toml:"json_time_format"`

	// default timezone
	JSONTimezone string `toml:"json_timezone"`

	// Authentication file for collectd
	CollectdAuthFile string `toml:"collectd_auth_file"`
	// One of none (default), sign, or encrypt
	CollectdSecurityLevel string `toml:"collectd_security_level"`
	// Dataset specification for collectd
	CollectdTypesDB []string `toml:"collectd_types_db"`

	// whether to split or join multivalue metrics
	CollectdSplit string `toml:"collectd_split"`

	// DataType only applies to value, this will be the type to parse value to
	DataType string `toml:"data_type"`

	// DefaultTags are the default tags that will be added to all parsed metrics.
	DefaultTags map[string]string `toml:"default_tags"`

	// an optional json path containing the metric registry object
	// if left empty, the whole json object is parsed as a metric registry
	DropwizardMetricRegistryPath string `toml:"dropwizard_metric_registry_path"`
	// an optional json path containing the default time of the metrics
	// if left empty, the processing time is used
	DropwizardTimePath string `toml:"dropwizard_time_path"`
	// time format to use for parsing the time field
	// defaults to time.RFC3339
	DropwizardTimeFormat string `toml:"dropwizard_time_format"`
	// an optional json path pointing to a json object with tag key/value pairs
	// takes precedence over DropwizardTagPathsMap
	DropwizardTagsPath string `toml:"dropwizard_tags_path"`
	// an optional map containing tag names as keys and json paths to retrieve the tag values from as values
	// used if TagsPath is empty or doesn't return any tags
	DropwizardTagPathsMap map[string]string `toml:"dropwizard_tag_paths_map"`

	//grok patterns
	GrokPatterns           []string `toml:"grok_patterns"`
	GrokNamedPatterns      []string `toml:"grok_named_patterns"`
	GrokCustomPatterns     string   `toml:"grok_custom_patterns"`
	GrokCustomPatternFiles []string `toml:"grok_custom_pattern_files"`
	GrokTimezone           string   `toml:"grok_timezone"`
	GrokUniqueTimestamp    string   `toml:"grok_unique_timestamp"`

	//csv configuration
	CSVColumnNames       []string `toml:"csv_column_names"`
	CSVColumnTypes       []string `toml:"csv_column_types"`
	CSVComment           string   `toml:"csv_comment"`
	CSVDelimiter         string   `toml:"csv_delimiter"`
	CSVHeaderRowCount    int      `toml:"csv_header_row_count"`
	CSVMeasurementColumn string   `toml:"csv_measurement_column"`
	CSVSkipColumns       int      `toml:"csv_skip_columns"`
	CSVSkipRows          int      `toml:"csv_skip_rows"`
	CSVTagColumns        []string `toml:"csv_tag_columns"`
	CSVTimestampColumn   string   `toml:"csv_timestamp_column"`
	CSVTimestampFormat   string   `toml:"csv_timestamp_format"`
	CSVTrimSpace         bool     `toml:"csv_trim_space"`
}

// NewParser returns a Parser interface based on the given config.
func NewParser(config *Config) (Parser, error) {
	var err error
	var parser Parser
	switch config.DataFormat {
	case "json":
		parser = newJSONParser(config.MetricName,
			config.TagKeys,
			config.JSONNameKey,
			config.JSONStringFields,
			config.JSONQuery,
			config.JSONTimeKey,
			config.JSONTimeFormat,
			config.JSONTimezone,
			config.DefaultTags)
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
	case "collectd":
		parser, err = NewCollectdParser(config.CollectdAuthFile,
			config.CollectdSecurityLevel, config.CollectdTypesDB, config.CollectdSplit)
	case "dropwizard":
		parser, err = NewDropwizardParser(
			config.DropwizardMetricRegistryPath,
			config.DropwizardTimePath,
			config.DropwizardTimeFormat,
			config.DropwizardTagsPath,
			config.DropwizardTagPathsMap,
			config.DefaultTags,
			config.Separator,
			config.Templates)
	case "wavefront":
		parser, err = NewWavefrontParser(config.DefaultTags)
	case "grok":
		parser, err = newGrokParser(
			config.MetricName,
			config.GrokPatterns,
			config.GrokNamedPatterns,
			config.GrokCustomPatterns,
			config.GrokCustomPatternFiles,
			config.GrokTimezone,
			config.GrokUniqueTimestamp)
	case "csv":
		parser, err = newCSVParser(config.MetricName,
			config.CSVHeaderRowCount,
			config.CSVSkipRows,
			config.CSVSkipColumns,
			config.CSVDelimiter,
			config.CSVComment,
			config.CSVTrimSpace,
			config.CSVColumnNames,
			config.CSVColumnTypes,
			config.CSVTagColumns,
			config.CSVMeasurementColumn,
			config.CSVTimestampColumn,
			config.CSVTimestampFormat,
			config.DefaultTags)
	case "logfmt":
		parser, err = NewLogFmtParser(config.MetricName, config.DefaultTags)
	default:
		err = fmt.Errorf("Invalid data format: %s", config.DataFormat)
	}
	return parser, err
}

func newCSVParser(metricName string,
	headerRowCount int,
	skipRows int,
	skipColumns int,
	delimiter string,
	comment string,
	trimSpace bool,
	columnNames []string,
	columnTypes []string,
	tagColumns []string,
	nameColumn string,
	timestampColumn string,
	timestampFormat string,
	defaultTags map[string]string) (Parser, error) {

	if headerRowCount == 0 && len(columnNames) == 0 {
		return nil, fmt.Errorf("`csv_header_row_count` must be defined if `csv_column_names` is not specified")
	}

	if delimiter != "" {
		runeStr := []rune(delimiter)
		if len(runeStr) > 1 {
			return nil, fmt.Errorf("csv_delimiter must be a single character, got: %s", delimiter)
		}
	}

	if comment != "" {
		runeStr := []rune(comment)
		if len(runeStr) > 1 {
			return nil, fmt.Errorf("csv_delimiter must be a single character, got: %s", comment)
		}
	}

	if len(columnNames) > 0 && len(columnTypes) > 0 && len(columnNames) != len(columnTypes) {
		return nil, fmt.Errorf("csv_column_names field count doesn't match with csv_column_types")
	}

	parser := &csv.Parser{
		MetricName:        metricName,
		HeaderRowCount:    headerRowCount,
		SkipRows:          skipRows,
		SkipColumns:       skipColumns,
		Delimiter:         delimiter,
		Comment:           comment,
		TrimSpace:         trimSpace,
		ColumnNames:       columnNames,
		ColumnTypes:       columnTypes,
		TagColumns:        tagColumns,
		MeasurementColumn: nameColumn,
		TimestampColumn:   timestampColumn,
		TimestampFormat:   timestampFormat,
		DefaultTags:       defaultTags,
		TimeFunc:          time.Now,
	}

	return parser, nil
}

func newJSONParser(
	metricName string,
	tagKeys []string,
	jsonNameKey string,
	stringFields []string,
	jsonQuery string,
	timeKey string,
	timeFormat string,
	timezone string,
	defaultTags map[string]string,
) Parser {
	parser := &json.JSONParser{
		MetricName:     metricName,
		TagKeys:        tagKeys,
		StringFields:   stringFields,
		JSONNameKey:    jsonNameKey,
		JSONQuery:      jsonQuery,
		JSONTimeKey:    timeKey,
		JSONTimeFormat: timeFormat,
		JSONTimezone:   timezone,
		DefaultTags:    defaultTags,
	}
	return parser
}

func newGrokParser(metricName string,
	patterns []string, nPatterns []string,
	cPatterns string, cPatternFiles []string,
	tZone string, uniqueTimestamp string) (Parser, error) {
	parser := grok.Parser{
		Measurement:        metricName,
		Patterns:           patterns,
		NamedPatterns:      nPatterns,
		CustomPatterns:     cPatterns,
		CustomPatternFiles: cPatternFiles,
		Timezone:           tZone,
		UniqueTimestamp:    uniqueTimestamp,
	}

	err := parser.Compile()
	return &parser, err
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
	handler := influx.NewMetricHandler()
	return influx.NewParser(handler), nil
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

func NewCollectdParser(
	authFile string,
	securityLevel string,
	typesDB []string,
	split string,
) (Parser, error) {
	return collectd.NewCollectdParser(authFile, securityLevel, typesDB, split)
}

func NewDropwizardParser(
	metricRegistryPath string,
	timePath string,
	timeFormat string,
	tagsPath string,
	tagPathsMap map[string]string,
	defaultTags map[string]string,
	separator string,
	templates []string,

) (Parser, error) {
	parser := dropwizard.NewParser()
	parser.MetricRegistryPath = metricRegistryPath
	parser.TimePath = timePath
	parser.TimeFormat = timeFormat
	parser.TagsPath = tagsPath
	parser.TagPathsMap = tagPathsMap
	parser.DefaultTags = defaultTags
	err := parser.SetTemplates(separator, templates)
	if err != nil {
		return nil, err
	}
	return parser, err
}

// NewLogFmtParser returns a logfmt parser with the default options.
func NewLogFmtParser(metricName string, defaultTags map[string]string) (Parser, error) {
	return logfmt.NewParser(metricName, defaultTags), nil
}

func NewWavefrontParser(defaultTags map[string]string) (Parser, error) {
	return wavefront.NewWavefrontParser(defaultTags), nil
}
