package opentelemetry_listener

import (
	"context"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type writer struct {
	accumulator telegraf.Accumulator
}

func (w *writer) WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	m, err := metric.New(measurement, tags, fields, ts)
	if err != nil {
		return err
	}
	w.accumulator.AddMetric(m)
	return nil
}
