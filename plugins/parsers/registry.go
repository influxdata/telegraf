package parsers

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers/temporary/json_v2"
	"github.com/influxdata/telegraf/plugins/parsers/temporary/xpath"
)

// Creator is the function to create a new parser
type Creator func(defaultMetricName string) telegraf.Parser

// Parsers contains the registry of all known parsers (following the new style)
var Parsers = map[string]Creator{}

// Add adds a parser to the registry. Usually this function is called in the plugin's init function
func Add(name string, creator Creator) {
	Parsers[name] = creator
}

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
	// SetParserFunc returns a new parser.
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
	// This function is only called by plugins that expect line based protocols
	// Doesn't need to be implemented by non-linebased parsers (e.g. json, xml)
	ParseLine(line string) (telegraf.Metric, error)

	// SetDefaultTags tells the parser to add all of the given tags
	// to each parsed metric.
	// NOTE: do _not_ modify the map after you've passed it here!!
	SetDefaultTags(tags map[string]string)
}

// ParserCompatibility is an interface for backward-compatible initialization of new parsers
type ParserCompatibility interface {
	// InitFromConfig sets the parser internal variables from the old-style config
	InitFromConfig(config *Config) error
}

// Config is a struct that covers the data types needed for all parser types,
// and can be used to instantiate _any_ of the parsers.
type Config struct {
	// DataFormat can be one of: json, influx, graphite, value, nagios
	DataFormat string `toml:"data_format"`

	// Separator only applied to Graphite data.
	Separator string `toml:"separator"`
	// Templates only apply to Graphite data.
	Templates []string `toml:"templates"`

	// TagKeys only apply to JSON data
	TagKeys []string `toml:"tag_keys"`
	// Array of glob pattern strings keys that should be added as string fields.
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

	// Whether to continue if a JSON object can't be coerced
	JSONStrict bool `toml:"json_strict"`

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
	CSVColumnNames        []string `toml:"csv_column_names"`
	CSVColumnTypes        []string `toml:"csv_column_types"`
	CSVComment            string   `toml:"csv_comment"`
	CSVDelimiter          string   `toml:"csv_delimiter"`
	CSVHeaderRowCount     int      `toml:"csv_header_row_count"`
	CSVMeasurementColumn  string   `toml:"csv_measurement_column"`
	CSVSkipColumns        int      `toml:"csv_skip_columns"`
	CSVSkipRows           int      `toml:"csv_skip_rows"`
	CSVTagColumns         []string `toml:"csv_tag_columns"`
	CSVTimestampColumn    string   `toml:"csv_timestamp_column"`
	CSVTimestampFormat    string   `toml:"csv_timestamp_format"`
	CSVTimezone           string   `toml:"csv_timezone"`
	CSVTrimSpace          bool     `toml:"csv_trim_space"`
	CSVSkipValues         []string `toml:"csv_skip_values"`
	CSVSkipErrors         bool     `toml:"csv_skip_errors"`
	CSVMetadataRows       int      `toml:"csv_metadata_rows"`
	CSVMetadataSeparators []string `toml:"csv_metadata_separators"`
	CSVMetadataTrimSet    string   `toml:"csv_metadata_trim_set"`

	// FormData configuration
	FormUrlencodedTagKeys []string `toml:"form_urlencoded_tag_keys"`

	// Prometheus configuration
	PrometheusIgnoreTimestamp bool `toml:"prometheus_ignore_timestamp"`

	// Value configuration
	ValueFieldName string `toml:"value_field_name"`

	// XPath configuration
	XPathPrintDocument       bool           `toml:"xpath_print_document"`
	XPathProtobufFile        string         `toml:"xpath_protobuf_file"`
	XPathProtobufType        string         `toml:"xpath_protobuf_type"`
	XPathProtobufImportPaths []string       `toml:"xpath_protobuf_import_paths"`
	XPathAllowEmptySelection bool           `toml:"xpath_allow_empty_selection"`
	XPathConfig              []xpath.Config `toml:"xpath"`

	// JSONPath configuration
	JSONV2Config []json_v2.Config `toml:"json_v2"`

	// Influx configuration
	InfluxParserType string `toml:"influx_parser_type"`

	// LogFmt configuration
	LogFmtTagKeys []string `toml:"logfmt_tag_keys"`
}

// NewParser returns a Parser interface based on the given config.
// DEPRECATED: Please instantiate the parser directly instead of using this function.
func NewParser(config *Config) (Parser, error) {
	creator, found := Parsers[config.DataFormat]
	if !found {
		return nil, fmt.Errorf("invalid data format: %s", config.DataFormat)
	}

	// Try to create new-style parsers the old way...
	parser := creator(config.MetricName)
	p, ok := parser.(ParserCompatibility)
	if !ok {
		return nil, fmt.Errorf("parser for %q cannot be created the old way", config.DataFormat)
	}
	err := p.InitFromConfig(config)

	return parser, err
}
