package serializers

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
)

// Creator is the function to create a new serializer
type Creator func() Serializer

// Serializers contains the registry of all known serializers (following the new style)
var Serializers = map[string]Creator{}

// Add adds a serializer to the registry. Usually this function is called in the plugin's init function
func Add(name string, creator Creator) {
	Serializers[name] = creator
}

// SerializerOutput is an interface for output plugins that are able to
// serialize telegraf metrics into arbitrary data formats.
type SerializerOutput interface {
	// SetSerializer sets the serializer function for the interface.
	SetSerializer(serializer Serializer)
}

// Serializer is an interface defining functions that a serializer plugin must
// satisfy.
//
// Implementations of this interface should be reentrant but are not required
// to be thread-safe.
type Serializer interface {
	// Serialize takes a single telegraf metric and turns it into a byte buffer.
	// separate metrics should be separated by a newline, and there should be
	// a newline at the end of the buffer.
	//
	// New plugins should use SerializeBatch instead to allow for non-line
	// delimited metrics.
	Serialize(metric telegraf.Metric) ([]byte, error)

	// SerializeBatch takes an array of telegraf metric and serializes it into
	// a byte buffer.  This method is not required to be suitable for use with
	// line oriented framing.
	SerializeBatch(metrics []telegraf.Metric) ([]byte, error)
}

// SerializerCompatibility is an interface for backward-compatible initialization of serializers
type SerializerCompatibility interface {
	// InitFromConfig sets the serializers internal variables from the old-style config
	InitFromConfig(config *Config) error
}

// Config is a struct that covers the data types needed for all serializer types,
// and can be used to instantiate _any_ of the serializers.
type Config struct {
	// DataFormat can be one of the serializer types listed in NewSerializer.
	DataFormat string `toml:"data_format"`

	// Carbon2 metric format.
	Carbon2Format string `toml:"carbon2_format"`

	// Character used for metric name sanitization in Carbon2.
	Carbon2SanitizeReplaceChar string `toml:"carbon2_sanitize_replace_char"`

	// Separator for CSV
	CSVSeparator string `toml:"csv_separator"`

	// Output a CSV header for naming the columns
	CSVHeader bool `toml:"csv_header"`

	// Prefix the tag and field columns for CSV format
	CSVPrefix bool `toml:"csv_column_prefix"`

	// Support tags in graphite protocol
	GraphiteTagSupport bool `toml:"graphite_tag_support"`

	// Support tags which follow the spec
	GraphiteTagSanitizeMode string `toml:"graphite_tag_sanitize_mode"`

	// Character for separating metric name and field for Graphite tags
	GraphiteSeparator string `toml:"graphite_separator"`

	// Regex string
	GraphiteStrictRegex string `toml:"graphite_strict_sanitize_regex"`

	// Maximum line length in bytes; influx format only
	InfluxMaxLineBytes int `toml:"influx_max_line_bytes"`

	// Sort field keys, set to true only when debugging as it less performant
	// than unsorted fields; influx format only
	InfluxSortFields bool `toml:"influx_sort_fields"`

	// Support unsigned integer output; influx format only
	InfluxUintSupport bool `toml:"influx_uint_support"`

	// Prefix to add to all measurements, only supports Graphite
	Prefix string `toml:"prefix"`

	// Template for converting telegraf metrics into Graphite
	// only supports Graphite
	Template string `toml:"template"`

	// Templates same Template, but multiple
	Templates []string `toml:"templates"`

	// Timestamp units to use for JSON formatted output
	TimestampUnits time.Duration `toml:"timestamp_units"`

	// Timestamp format to use for JSON and CSV formatted output
	TimestampFormat string `toml:"timestamp_format"`

	// Transformation as JSONata expression to use for JSON formatted output
	Transformation string `toml:"transformation"`

	// Field filter for interpreting data as nested JSON for JSON serializer
	JSONNestedFieldInclude []string `toml:"json_nested_fields_include"`
	JSONNestedFieldExclude []string `toml:"json_nested_fields_exclude"`

	// Include HEC routing fields for splunkmetric output
	HecRouting bool `toml:"hec_routing"`

	// Enable Splunk MultiMetric output (Splunk 8.0+)
	SplunkmetricMultiMetric bool `toml:"splunkmetric_multi_metric"`

	// Omit the Splunk Event "metric" tag
	SplunkmetricOmitEventTag bool `toml:"splunkmetric_omit_event_tag"`

	// Point tags to use as the source name for Wavefront (if none found, host will be used).
	WavefrontSourceOverride []string `toml:"wavefront_source_override"`

	// Use Strict rules to sanitize metric and tag names from invalid characters for Wavefront
	// When enabled forward slash (/) and comma (,) will be accepted
	WavefrontUseStrict bool `toml:"wavefront_use_strict"`

	// Convert "_" in prefixes to "." for Wavefront
	WavefrontDisablePrefixConversion bool `toml:"wavefront_disable_prefix_conversion"`

	// Include the metric timestamp on each sample.
	PrometheusExportTimestamp bool `toml:"prometheus_export_timestamp"`

	// Sort prometheus metric families and metric samples.  Useful for
	// debugging.
	PrometheusSortMetrics bool `toml:"prometheus_sort_metrics"`

	// Output string fields as metric labels; when false string fields are
	// discarded.
	PrometheusStringAsLabel bool `toml:"prometheus_string_as_label"`

	// Encode metrics without HELP metadata. This helps reduce the payload size.
	PrometheusCompactEncoding bool `toml:"prometheus_compact_encoding"`
}

// NewSerializer a Serializer interface based on the given config.
func NewSerializer(config *Config) (Serializer, error) {
	creator, found := Serializers[config.DataFormat]
	if !found {
		return nil, fmt.Errorf("invalid data format: %s", config.DataFormat)
	}

	// Try to create new-style serializers the old way...
	serializer := creator()
	p, ok := serializer.(SerializerCompatibility)
	if !ok {
		return nil, fmt.Errorf("serializer for %q cannot be created the old way", config.DataFormat)
	}
	err := p.InitFromConfig(config)

	return serializer, err
}
