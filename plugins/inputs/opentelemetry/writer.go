package opentelemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/influxdb-observability/otel2influx"
	"github.com/influxdata/telegraf"
)

type writeToAccumulator struct {
	accumulator telegraf.Accumulator
}

func (w *writeToAccumulator) WritePoint(_ context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time, vType otel2influx.InfluxWriterValueType) error {
	switch vType {
	case otel2influx.InfluxWriterValueTypeUntyped:
		w.accumulator.AddFields(measurement, fields, tags, ts)
	case otel2influx.InfluxWriterValueTypeGauge:
		w.accumulator.AddGauge(measurement, fields, tags, ts)
	case otel2influx.InfluxWriterValueTypeSum:
		w.accumulator.AddCounter(measurement, fields, tags, ts)
	case otel2influx.InfluxWriterValueTypeHistogram:
		w.accumulator.AddHistogram(measurement, fields, tags, ts)
	case otel2influx.InfluxWriterValueTypeSummary:
		w.accumulator.AddSummary(measurement, fields, tags, ts)
	default:
		return fmt.Errorf("unrecognized InfluxWriterValueType %q", vType)
	}
	return nil
}
