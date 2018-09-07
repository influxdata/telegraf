package parsers

import (
	"fmt"

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
	DataFormat string

	// Separator only applied to Graphite data.
	Separator string
	// Templates only apply to Graphite data.
	Templates []string

	// TagKeys only apply to JSON data
	TagKeys []string
	// FieldKeys only apply to JSON
	JSONStringFields []string

	JSONNameKey string
	// MetricName applies to JSON & value. This will be the name of the measurement.
	MetricName string

	// holds a gjson path for json parser
	JSONQuery string

	// key of time
	JSONTimeKey string

	// time format
	JSONTimeFormat string

	// Authentication file for collectd
	CollectdAuthFile string
	// One of none (default), sign, or encrypt
	CollectdSecurityLevel string
	// Dataset specification for collectd
	CollectdTypesDB []string

	// whether to split or join multivalue metrics
	CollectdSplit string

	// DataType only applies to value, this will be the type to parse value to
	DataType string

	// DefaultTags are the default tags that will be added to all parsed metrics.
	DefaultTags map[string]string

	// an optional json path containing the metric registry object
	// if left empty, the whole json object is parsed as a metric registry
	DropwizardMetricRegistryPath string
	// an optional json path containing the default time of the metrics
	// if left empty, the processing time is used
	DropwizardTimePath string
	// time format to use for parsing the time field
	// defaults to time.RFC3339
	DropwizardTimeFormat string
	// an optional json path pointing to a json object with tag key/value pairs
	// takes precedence over DropwizardTagPathsMap
	DropwizardTagsPath string
	// an optional map containing tag names as keys and json paths to retrieve the tag values from as values
	// used if TagsPath is empty or doesn't return any tags
	DropwizardTagPathsMap map[string]string

	//grok patterns
	GrokPatterns           []string
	GrokNamedPatterns      []string
	GrokCustomPatterns     string
	GrokCustomPatternFiles []string
	GrokTimeZone           string

	//csv configuration
	CSVDelimiter         string
	CSVComment           string
	CSVTrimSpace         bool
	CSVColumnNames       []string
	CSVTagColumns        []string
	CSVMeasurementColumn string
	CSVTimestampColumn   string
	CSVTimestampFormat   string
	CSVHeaderRowCount    int
	CSVSkipRows          int
	CSVSkipColumns       int
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
			config.GrokTimeZone)
	case "csv":
		parser, err = newCSVParser(config.MetricName,
			config.CSVHeaderRowCount,
			config.CSVSkipRows,
			config.CSVSkipColumns,
			config.CSVDelimiter,
			config.CSVComment,
			config.CSVTrimSpace,
			config.CSVColumnNames,
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
	header int,
	skipRows int,
	skipColumns int,
	delimiter string,
	comment string,
	trimSpace bool,
	dataColumns []string,
	tagColumns []string,
	nameColumn string,
	timestampColumn string,
	timestampFormat string,
	defaultTags map[string]string) (Parser, error) {

	if header == 0 && len(dataColumns) == 0 {
		// if there is no header and no DataColumns, that's an error
		return nil, fmt.Errorf("there must be a header if `csv_data_columns` is not specified")
	}

	if delimiter != "" {
		runeStr := []rune(delimiter)
		if len(runeStr) > 1 {
			return nil, fmt.Errorf("delimiter must be a single character, got: %s", delimiter)
		}
		delimiter = fmt.Sprintf("%v", runeStr[0])
	}

	if comment != "" {
		runeStr := []rune(comment)
		if len(runeStr) > 1 {
			return nil, fmt.Errorf("delimiter must be a single character, got: %s", comment)
		}
		comment = fmt.Sprintf("%v", runeStr[0])
	}

	parser := &csv.Parser{
		MetricName:        metricName,
		HeaderRowCount:    header,
		SkipRows:          skipRows,
		SkipColumns:       skipColumns,
		Delimiter:         delimiter,
		Comment:           comment,
		TrimSpace:         trimSpace,
		ColumnNames:       dataColumns,
		TagColumns:        tagColumns,
		MeasurementColumn: nameColumn,
		TimestampColumn:   timestampColumn,
		TimestampFormat:   timestampFormat,
		DefaultTags:       defaultTags,
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
		DefaultTags:    defaultTags,
	}
	return parser
}

//Deprecated: Use NewParser to get a JSONParser object
func newGrokParser(metricName string,
	patterns []string,
	nPatterns []string,
	cPatterns string,
	cPatternFiles []string, tZone string) (Parser, error) {
	parser := grok.Parser{
		Measurement:        metricName,
		Patterns:           patterns,
		NamedPatterns:      nPatterns,
		CustomPatterns:     cPatterns,
		CustomPatternFiles: cPatternFiles,
		Timezone:           tZone,
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
