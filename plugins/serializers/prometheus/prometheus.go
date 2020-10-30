package prometheus

import (
	"bytes"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/prometheus/prompb"
	"math"
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

// ProtobufFormating defines how to format body request.
type Formating string

const (
	FormatRemoteWrite Formating = "remote_write"
	FormatText        Formating = "text"
	Format            Formating = FormatText
)

type FormatConfig struct {
	TimestampExport TimestampExport
	MetricSortOrder MetricSortOrder
	StringHandling  StringHandling
	Format          Formating
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

	if s.config.Format == FormatRemoteWrite {
		var promTS []*prompb.TimeSeries
		for _, mf := range coll.GetProto() {
			for _, m := range mf.GetMetric() {
				commonLabels := make([]*prompb.Label, len(m.Label))
				for j, label := range m.Label {
					commonLabels[j] = &prompb.Label{Name: label.GetName(), Value: label.GetValue()}
				}
				switch mf.GetType() {
				case dto.MetricType_COUNTER:
					promTS = addPromTS(mf.GetName(), commonLabels, m.Counter.GetValue(), m.GetTimestampMs(), promTS)
				case dto.MetricType_GAUGE:
					promTS = addPromTS(mf.GetName(), commonLabels, m.Gauge.GetValue(), m.GetTimestampMs(), promTS)
				case dto.MetricType_SUMMARY:
					if len(m.Summary.Quantile) > 0 {
						for _, q := range m.Summary.GetQuantile() {
							labels := make([]*prompb.Label, len(commonLabels), len(commonLabels)+1)
							copy(labels, commonLabels)
							labels = append(labels, &prompb.Label{
								Name:  "quantile",
								Value: fmt.Sprint(q.GetQuantile()),
							})
							promTS = addPromTS(
								mf.GetName(),
								labels,
								q.GetValue(),
								m.GetTimestampMs(),
								promTS,
							)
						}
					}
					if m.Summary.SampleSum != nil {
						promTS = addPromTS(
							fmt.Sprintf("%s_sum", mf.GetName()),
							commonLabels,
							m.Summary.GetSampleSum(),
							m.GetTimestampMs(),
							promTS,
						)
					}
					if m.Summary.SampleCount != nil {
						promTS = addPromTS(
							fmt.Sprintf("%s_count", mf.GetName()),
							commonLabels,
							float64(m.Summary.GetSampleCount()),
							m.GetTimestampMs(),
							promTS,
						)
					}
				case dto.MetricType_UNTYPED:
					promTS = addPromTS(mf.GetName(), commonLabels, m.Untyped.GetValue(), m.GetTimestampMs(), promTS)
				case dto.MetricType_HISTOGRAM:
					var isInfPresent bool = false

					if len(m.Histogram.Bucket) > 0 {
						for _, b := range m.Histogram.Bucket {
							labels := make([]*prompb.Label, len(commonLabels), len(commonLabels)+1)
							copy(labels, commonLabels)
							labels = append(labels, &prompb.Label{
								Name:  "le",
								Value: fmt.Sprint(b.GetUpperBound()),
							})
							promTS = addPromTS(
								fmt.Sprintf("%s_bucket", mf.GetName()),
								labels,
								float64(*b.CumulativeCount),
								m.GetTimestampMs(),
								promTS,
							)
							if b.GetUpperBound() == math.Inf(1) {
								isInfPresent = true
							}
						}
					}
					if !isInfPresent {
						labels := make([]*prompb.Label, len(commonLabels), len(commonLabels)+1)
						copy(labels, commonLabels)
						labels = append(labels, &prompb.Label{
							Name:  "le",
							Value: "+Inf",
						})
						promTS = addPromTS(
							fmt.Sprintf("%s_bucket", mf.GetName()),
							labels,
							float64(m.Histogram.GetSampleCount()),
							m.GetTimestampMs(),
							promTS,
						)
					}
					if m.Histogram.SampleSum != nil {
						promTS = addPromTS(
							fmt.Sprintf("%s_sum", mf.GetName()),
							commonLabels,
							m.Histogram.GetSampleSum(),
							m.GetTimestampMs(),
							promTS,
						)
					}
					if m.Histogram.SampleCount != nil {
						promTS = addPromTS(
							fmt.Sprintf("%s_count", mf.GetName()),
							commonLabels,
							float64(m.Histogram.GetSampleCount()),
							m.GetTimestampMs(),
							promTS,
						)
					}
				default:
					return nil, fmt.Errorf("Unknown type %v", mf.Type)
				}
			}

		}
		data, err := proto.Marshal(&prompb.WriteRequest{Timeseries: promTS})
		if err != nil {
			return nil, fmt.Errorf("unable to marshal protobuf: %v", err)
		}
		encoded := snappy.Encode(nil, data)
		buf.Write(encoded)
	}

	if s.config.Format == FormatText {
		for _, mf := range coll.GetProto() {
			enc := expfmt.NewEncoder(&buf, expfmt.FmtText)
			err := enc.Encode(mf)
			if err != nil {
				return nil, err
			}
		}
	}

	return buf.Bytes(), nil
}

func addPromTS(name string, labels []*prompb.Label, value float64, ts int64, series []*prompb.TimeSeries) []*prompb.TimeSeries {
	sample := []prompb.Sample{{
		// Timestamp is int milliseconds for remote write.
		Timestamp: ts,
		Value:     value,
	}}
	labelscopy := make([]*prompb.Label, len(labels), len(labels)+1)
	copy(labelscopy, labels)
	labels = append(labelscopy, &prompb.Label{
		Name:  "__name__",
		Value: name,
	})
	return append(series, &prompb.TimeSeries{Labels: labels, Samples: sample})
}
