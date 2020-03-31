package serializers

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/carbon2"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/plugins/serializers/nowmetric"
	"github.com/influxdata/telegraf/plugins/serializers/prometheus"
	"github.com/influxdata/telegraf/plugins/serializers/splunkmetric"
	"github.com/influxdata/telegraf/plugins/serializers/wavefront"
)

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

// Config is a struct that covers the data types needed for all serializer types,
// and can be used to instantiate _any_ of the serializers.
type Config struct {
	// Dataformat can be one of the serializer types listed in NewSerializer.
	DataFormat string `toml:"data_format"`

	// Support tags in graphite protocol
	GraphiteTagSupport bool `toml:"graphite_tag_support"`

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

	// Include HEC routing fields for splunkmetric output
	HecRouting bool `toml:"hec_routing"`

	// Enable Splunk MultiMetric output (Splunk 8.0+)
	SplunkmetricMultiMetric bool `toml:"splunkmetric_multi_metric"`

	// Point tags to use as the source name for Wavefront (if none found, host will be used).
	WavefrontSourceOverride []string `toml:"wavefront_source_override"`

	// Use Strict rules to sanitize metric and tag names from invalid characters for Wavefront
	// When enabled forward slash (/) and comma (,) will be accepted
	WavefrontUseStrict bool `toml:"wavefront_use_strict"`

	// Include the metric timestamp on each sample.
	PrometheusExportTimestamp bool `toml:"prometheus_export_timestamp"`

	// Sort prometheus metric families and metric samples.  Useful for
	// debugging.
	PrometheusSortMetrics bool `toml:"prometheus_sort_metrics"`

	// Output string fields as metric labels; when false string fields are
	// discarded.
	PrometheusStringAsLabel bool `toml:"prometheus_string_as_label"`
}

// NewSerializer a Serializer interface based on the given config.
func NewSerializer(config *Config) (Serializer, error) {
	var err error
	var serializer Serializer
	switch config.DataFormat {
	case "influx":
		serializer, err = NewInfluxSerializerConfig(config)
	case "graphite":
		serializer, err = NewGraphiteSerializer(config.Prefix, config.Template, config.GraphiteTagSupport, config.Templates)
	case "json":
		serializer, err = NewJsonSerializer(config.TimestampUnits)
	case "splunkmetric":
		serializer, err = NewSplunkmetricSerializer(config.HecRouting, config.SplunkmetricMultiMetric)
	case "nowmetric":
		serializer, err = NewNowSerializer()
	case "carbon2":
		serializer, err = NewCarbon2Serializer()
	case "wavefront":
		serializer, err = NewWavefrontSerializer(config.Prefix, config.WavefrontUseStrict, config.WavefrontSourceOverride)
	case "prometheus":
		serializer, err = NewPrometheusSerializer(config)
	default:
		err = fmt.Errorf("Invalid data format: %s", config.DataFormat)
	}
	return serializer, err
}

func NewPrometheusSerializer(config *Config) (Serializer, error) {
	exportTimestamp := prometheus.NoExportTimestamp
	if config.PrometheusExportTimestamp {
		exportTimestamp = prometheus.ExportTimestamp
	}

	sortMetrics := prometheus.NoSortMetrics
	if config.PrometheusExportTimestamp {
		sortMetrics = prometheus.SortMetrics
	}

	stringAsLabels := prometheus.DiscardStrings
	if config.PrometheusStringAsLabel {
		stringAsLabels = prometheus.StringAsLabel
	}

	return prometheus.NewSerializer(prometheus.FormatConfig{
		TimestampExport: exportTimestamp,
		MetricSortOrder: sortMetrics,
		StringHandling:  stringAsLabels,
	})
}

func NewWavefrontSerializer(prefix string, useStrict bool, sourceOverride []string) (Serializer, error) {
	return wavefront.NewSerializer(prefix, useStrict, sourceOverride)
}

func NewJsonSerializer(timestampUnits time.Duration) (Serializer, error) {
	return json.NewSerializer(timestampUnits)
}

func NewCarbon2Serializer() (Serializer, error) {
	return carbon2.NewSerializer()
}

func NewSplunkmetricSerializer(splunkmetric_hec_routing bool, splunkmetric_multimetric bool) (Serializer, error) {
	return splunkmetric.NewSerializer(splunkmetric_hec_routing, splunkmetric_multimetric)
}

func NewNowSerializer() (Serializer, error) {
	return nowmetric.NewSerializer()
}

func NewInfluxSerializerConfig(config *Config) (Serializer, error) {
	var sort influx.FieldSortOrder
	if config.InfluxSortFields {
		sort = influx.SortFields
	}

	var typeSupport influx.FieldTypeSupport
	if config.InfluxUintSupport {
		typeSupport = typeSupport + influx.UintSupport
	}

	s := influx.NewSerializer()
	s.SetMaxLineBytes(config.InfluxMaxLineBytes)
	s.SetFieldSortOrder(sort)
	s.SetFieldTypeSupport(typeSupport)
	return s, nil
}

func NewInfluxSerializer() (Serializer, error) {
	return influx.NewSerializer(), nil
}

func NewGraphiteSerializer(prefix, template string, tag_support bool, templates []string) (Serializer, error) {
	graphiteTemplates, defaultTemplate, err := graphite.InitGraphiteTemplates(templates)

	if err != nil {
		return nil, err
	}

	if defaultTemplate != "" {
		template = defaultTemplate
	}

	return &graphite.GraphiteSerializer{
		Prefix:     prefix,
		Template:   template,
		TagSupport: tag_support,
		Templates:  graphiteTemplates,
	}, nil
}
