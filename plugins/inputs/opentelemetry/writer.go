package opentelemetry

import (
	"context"
	"time"

	"github.com/influxdata/telegraf"
)

type writeToAccumulator struct {
	accumulator telegraf.Accumulator
}

func (w *writeToAccumulator) WritePoint(_ context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	w.accumulator.AddFields(measurement, fields, tags, ts)
	return nil
}
