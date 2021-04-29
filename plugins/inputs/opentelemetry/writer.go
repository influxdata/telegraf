package opentelemetry

import (
	"context"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type writeToAccumulator struct {
	accumulator telegraf.Accumulator
}

func (w *writeToAccumulator) WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	w.accumulator.AddMetric(metric.New(measurement, tags, fields, ts))
	return nil
}
