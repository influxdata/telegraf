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
	// the same name as those are controlled by the sender.
	if certTags := clientCertTags(ctx, w.clientCertTags); len(certTags) > 0 {
		tags = maps.Clone(tags)
		if tags == nil {
			tags = certTags
		} else {
			maps.Copy(tags, certTags)
		}
	}

	switch vType {
	case common.InfluxMetricValueTypeUntyped:
		w.accumulator.AddFields(measurement, fields, tags, ts)
	case common.InfluxMetricValueTypeGauge:
		w.accumulator.AddGauge(measurement, fields, tags, ts)
	case common.InfluxMetricValueTypeSum:
		w.accumulator.AddCounter(measurement, fields, tags, ts)
	case common.InfluxMetricValueTypeHistogram:
		w.accumulator.AddHistogram(measurement, fields, tags, ts)
	case common.InfluxMetricValueTypeSummary:
		w.accumulator.AddSummary(measurement, fields, tags, ts)
	default:
		return fmt.Errorf("unrecognized InfluxMetricValueType %q", vType)
	}
	return nil
}

// WriteBatch does nothing.
func (*writeToAccumulator) WriteBatch(context.Context) error {
	return nil
}
