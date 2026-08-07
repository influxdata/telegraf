package prometheus

import (
	"bytes"
	"fmt"
	"time"

	"github.com/prometheus/common/expfmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Serializer struct {
	FormatConfig
}

// FormatConfig contains the configuration for the Prometheus serializer.
type FormatConfig struct {
	ExportTimestamp bool `toml:"prometheus_export_timestamp"`
	SortMetrics     bool `toml:"prometheus_sort_metrics"`
	StringAsLabel   bool `toml:"prometheus_string_as_label"`
	// CompactEncoding defines whether to include
	// HELP metadata in Prometheus payload. Setting to true
	// helps to reduce payload size.
	CompactEncoding bool        `toml:"prometheus_compact_encoding"`
	TypeMappings    MetricTypes `toml:"prometheus_metric_types"`
	// NameSanitization controls how metric names and label names are sanitized.
	// Valid values: "legacy" (ASCII-only rules), "utf8" (allows UTF-8 names).
	NameSanitization string `toml:"prometheus_name_sanitization"`
}

// MetricTypes defines the mapping of metric names to their types.
type MetricTypes struct {
	Counter []string `toml:"counter"`
	Gauge   []string `toml:"gauge"`

	filterCounter filter.Filter
	filterGauge   filter.Filter
}

// Init initializes the MetricTypes by compiling the filters for counter and gauge metrics.
func (mt *MetricTypes) Init() error {
	// Setup the explicit type mappings
	var err error
	mt.filterCounter, err = filter.Compile(mt.Counter)
	if err != nil {
		return fmt.Errorf("creating counter filter failed: %w", err)
	}
	mt.filterGauge, err = filter.Compile(mt.Gauge)
	if err != nil {
		return fmt.Errorf("creating gauge filter failed: %w", err)
	}
	return nil
}

// DetermineType determines the type of the metric based on its name and the configured filters.
func (mt *MetricTypes) DetermineType(name string, m telegraf.Metric) telegraf.ValueType {
	metricType := m.Type()
	if mt.filterCounter != nil && mt.filterCounter.Match(name) {
		metricType = telegraf.Counter
	}
	if mt.filterGauge != nil && mt.filterGauge.Match(name) {
		metricType = telegraf.Gauge
	}
	return metricType
}

func (s *Serializer) Init() error {
	switch s.NameSanitization {
	case "":
		s.NameSanitization = "legacy"
	case "legacy", "utf8":
		// Valid sanitization modes.
	default:
		return fmt.Errorf("invalid prometheus_name_sanitization %q: must be \"legacy\" or \"utf8\"", s.NameSanitization)
	}

	return s.FormatConfig.TypeMappings.Init()
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	return s.SerializeBatch([]telegraf.Metric{metric})
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	coll := NewCollection(s.FormatConfig)
	for _, metric := range metrics {
		coll.Add(metric, time.Now())
	}

	var buf bytes.Buffer
	for _, mf := range coll.GetProto() {
		enc := expfmt.NewEncoder(&buf, expfmt.NewFormat(expfmt.TypeTextPlain))
		err := enc.Encode(mf)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func init() {
	serializers.Add("prometheus",
		func() telegraf.Serializer {
			return &Serializer{}
		},
	)
}
