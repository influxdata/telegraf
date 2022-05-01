package merge

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type Merge struct {
	grouper *metric.SeriesGrouper
}

func (a *Merge) Init() error {
	a.grouper = metric.NewSeriesGrouper()
	return nil
}

func (a *Merge) Add(m telegraf.Metric) {
	a.grouper.AddMetric(m)
}

func (a *Merge) Push(acc telegraf.Accumulator) {
	// Always use nanosecond precision to avoid rounding metrics that were
	// produced at a precision higher than the agent default.
	acc.SetPrecision(time.Nanosecond)

	for _, m := range a.grouper.Metrics() {
		acc.AddMetric(m)
	}
}

func (a *Merge) Reset() {
	a.grouper = metric.NewSeriesGrouper()
}

func init() {
	aggregators.Add("merge", func() telegraf.Aggregator {
		return &Merge{}
	})
}
