package prometheus

import (
	"bytes"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/prometheus/common/expfmt"
)

// TimestampExport controls if the output contains timestamps.
type TimestampExport int

const (
	NoExportTimestamp TimestampExport = iota
	ExportTimestamp
)

// MetricSortOrder controls if the output is sorted.
type MetricSortOrder int

const (
	NoSortMetrics MetricSortOrder = iota
	SortMetrics
)

// StringHandling defines how to process string fields.
type StringHandling int

const (
	DiscardStrings StringHandling = iota
	StringAsLabel
)

type FormatConfig struct {
	TimestampExport TimestampExport
	MetricSortOrder MetricSortOrder
	StringHandling  StringHandling
}

type Serializer struct {
	config FormatConfig
}

func NewSerializer(config FormatConfig) (*Serializer, error) {
	s := &Serializer{config: config}
	return s, nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	return s.SerializeBatch([]telegraf.Metric{metric})
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	coll := NewCollection(s.config)
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
