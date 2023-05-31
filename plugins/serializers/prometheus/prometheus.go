package prometheus

import (
	"bytes"
	"time"

	"github.com/prometheus/common/expfmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type FormatConfig struct {
	ExportTimestamp bool `toml:"prometheus_export_timestamp"`
	SortMetrics     bool `toml:"prometheus_sort_metrics"`
	StringAsLabel   bool `toml:"prometheus_string_as_label"`
	// CompactEncoding defines whether to include
	// HELP metadata in Prometheus payload. Setting to true
	// helps to reduce payload size.
	CompactEncoding bool `toml:"prometheus_compact_encoding"`
}

type Serializer struct {
	FormatConfig
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
		enc := expfmt.NewEncoder(&buf, expfmt.FmtText)
		err := enc.Encode(mf)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func init() {
	serializers.Add("prometheus",
		func() serializers.Serializer {
			return &Serializer{}
		},
	)
}

// InitFromConfig is a compatibility function to construct the parser the old way
func (s *Serializer) InitFromConfig(cfg *serializers.Config) error {
	s.FormatConfig.CompactEncoding = cfg.PrometheusCompactEncoding
	s.FormatConfig.SortMetrics = cfg.PrometheusSortMetrics
	s.FormatConfig.StringAsLabel = cfg.PrometheusStringAsLabel
	s.FormatConfig.ExportTimestamp = cfg.PrometheusExportTimestamp

	return nil
}
