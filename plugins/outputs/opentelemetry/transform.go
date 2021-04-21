package opentelemetry

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

func monotonicIntegerPoint(labels []*commonpb.StringKeyValue, start, end int64, value int64) *metricpb.IntSum {
	integer := &metricpb.IntDataPoint{
		Labels:            labels,
		StartTimeUnixNano: uint64(start),
		TimeUnixNano:      uint64(end),
		Value:             value,
	}
	return &metricpb.IntSum{
		IsMonotonic:            true,
		AggregationTemporality: metricpb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
		DataPoints:             []*metricpb.IntDataPoint{integer},
	}
}

func monotonicDoublePoint(labels []*commonpb.StringKeyValue, start, end int64, value float64) *metricpb.DoubleSum {
	double := &metricpb.DoubleDataPoint{
		Labels:            labels,
		StartTimeUnixNano: uint64(time.Duration(start) * time.Nanosecond),
		TimeUnixNano:      uint64(time.Duration(end) * time.Nanosecond),
		Value:             value,
	}
	return &metricpb.DoubleSum{
		IsMonotonic:            true,
		AggregationTemporality: metricpb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
		DataPoints:             []*metricpb.DoubleDataPoint{double},
	}
}

func protoLabel(tag *telegraf.Tag) *commonpb.KeyValue {
	return &commonpb.KeyValue{
		Key: tag.Key,
		Value: &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: tag.Value,
			},
		},
	}
}

func protoStringLabel(tag *telegraf.Tag) *commonpb.StringKeyValue {
	return &commonpb.StringKeyValue{
		Key:   tag.Key,
		Value: tag.Value,
	}
}

func protoResourceAttributes(tags []*telegraf.Tag) []*commonpb.KeyValue {
	ret := make([]*commonpb.KeyValue, len(tags))
	for i := range tags {
		ret[i] = protoLabel(tags[i])
	}
	return ret
}

func protoStringLabels(tags []*telegraf.Tag) []*commonpb.StringKeyValue {
	ret := make([]*commonpb.StringKeyValue, len(tags))
	for i := range tags {
		ret[i] = protoStringLabel(tags[i])
	}
	return ret
}

func protoResource(resourceTags []*telegraf.Tag) *resourcepb.Resource {
	return &resourcepb.Resource{
		Attributes: protoResourceAttributes(resourceTags),
	}
}

func protoTimeseries(resourceTags []*telegraf.Tag, m telegraf.Metric, f *telegraf.Field) (*metricpb.ResourceMetrics, *metricpb.Metric) {
	metric := &metricpb.Metric{
		Name:        fmt.Sprintf("%s.%s", m.Name(), f.Key),
		Description: "", // TODO
		Unit:        "", // TODO
	}
	return &metricpb.ResourceMetrics{
		Resource: protoResource(resourceTags),
		InstrumentationLibraryMetrics: []*metricpb.InstrumentationLibraryMetrics{
			{
				InstrumentationLibrary: &commonpb.InstrumentationLibrary{
					Name:    instrumentationLibraryName,
					Version: internal.Version(),
				},
				Metrics: []*metricpb.Metric{metric},
			},
		},
	}, metric
}

func intGauge(labels []*commonpb.StringKeyValue, ts int64, value int64) *metricpb.IntGauge {
	integer := &metricpb.IntDataPoint{
		Labels:       labels,
		TimeUnixNano: uint64(ts),
		Value:        value,
	}
	return &metricpb.IntGauge{
		DataPoints: []*metricpb.IntDataPoint{integer},
	}
}

func doubleGauge(labels []*commonpb.StringKeyValue, ts int64, value float64) *metricpb.DoubleGauge {
	double := &metricpb.DoubleDataPoint{
		Labels:       labels,
		TimeUnixNano: uint64(ts),
		Value:        value,
	}
	return &metricpb.DoubleGauge{
		DataPoints: []*metricpb.DoubleDataPoint{double},
	}
}
