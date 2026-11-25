package opentelemetry

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/influxdata/influxdb-observability/common"
	"github.com/influxdata/influxdb-observability/otel2influx"

	"github.com/influxdata/telegraf"
)

var (
	_ otel2influx.InfluxWriter      = (*writeToAccumulator)(nil)
	_ otel2influx.InfluxWriterBatch = (*writeToAccumulator)(nil)
)

type writeToAccumulator struct {
	accumulator    telegraf.Accumulator
	clientCertTags []string
}

// NewBatch creates a new batch for writing telemetry data.
func (w *writeToAccumulator) NewBatch() otel2influx.InfluxWriterBatch {
	return w
}

// EnqueuePoint adds a telemetry data point to the accumulator.
func (w *writeToAccumulator) EnqueuePoint(
	ctx context.Context,
	measurement string,
	tags map[string]string,
	fields map[string]any,
	ts time.Time,
	vType common.InfluxMetricValueType,
) error {
	// Tags derived from the certificate take precedence over attributes with
	// the same name as those are controlled by the sender. Attributes that
	// the certificate does not provide must not be settable by the sender
	// either, so remove all potential certificate tags first.
	mergedTags := tags
	if len(w.clientCertTags) > 0 {
		mergedTags = make(map[string]string, len(tags)+len(w.clientCertTags))
		maps.Copy(mergedTags, tags)
		removeClientCertTags(mergedTags, w.clientCertTags)
		maps.Copy(mergedTags, clientCertTags(ctx, w.clientCertTags))
	}

	switch vType {
	case common.InfluxMetricValueTypeUntyped:
		w.accumulator.AddFields(measurement, fields, mergedTags, ts)
	case common.InfluxMetricValueTypeGauge:
		w.accumulator.AddGauge(measurement, fields, mergedTags, ts)
	case common.InfluxMetricValueTypeSum:
		w.accumulator.AddCounter(measurement, fields, mergedTags, ts)
	case common.InfluxMetricValueTypeHistogram:
		w.accumulator.AddHistogram(measurement, fields, mergedTags, ts)
	case common.InfluxMetricValueTypeSummary:
		w.accumulator.AddSummary(measurement, fields, mergedTags, ts)
	default:
		return fmt.Errorf("unrecognized InfluxMetricValueType %q", vType)
	}
	return nil
}

// WriteBatch does nothing.
func (*writeToAccumulator) WriteBatch(context.Context) error {
	return nil
}
