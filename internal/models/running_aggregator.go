package models

import (
	"log"
	"time"

	"github.com/influxdata/telegraf"
)

type RunningAggregator struct {
	a      telegraf.Aggregator
	Config *AggregatorConfig

	metrics chan telegraf.Metric

	periodStart time.Time
	periodEnd   time.Time
}

func NewRunningAggregator(
	a telegraf.Aggregator,
	conf *AggregatorConfig,
) *RunningAggregator {
	return &RunningAggregator{
		a:       a,
		Config:  conf,
		metrics: make(chan telegraf.Metric, 100),
	}
}

// AggregatorConfig is the common config for all aggregators.
type AggregatorConfig struct {
	Name         string
	DropOriginal bool
	Period       time.Duration
	Delay        time.Duration

	NameOverride      string
	MeasurementPrefix string
	MeasurementSuffix string
	Tags              map[string]string
	Filter            Filter
}

func (r *RunningAggregator) Name() string {
	return "aggregators." + r.Config.Name
}

func (r *RunningAggregator) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	m := makemetric(
		metric,
		r.Config.NameOverride,
		r.Config.MeasurementPrefix,
		r.Config.MeasurementSuffix,
		r.Config.Tags,
		nil)

	if m != nil {
		m.SetAggregate(true)
	}

	return m
}

// Add a metric to the aggregator and return true if the original metric
// should be dropped.
func (r *RunningAggregator) Add(metric telegraf.Metric) bool {
	if ok := r.Config.Filter.Select(metric); !ok {
		return false
	}

	r.Config.Filter.Modify(metric)
	if len(metric.FieldList()) == 0 {
		return r.Config.DropOriginal
	}

	r.metrics <- metric

	return r.Config.DropOriginal
}

func (r *RunningAggregator) add(in telegraf.Metric) {
	r.a.Add(in)
}

func (r *RunningAggregator) push(acc telegraf.Accumulator) {
	r.a.Push(acc)
}

func (r *RunningAggregator) reset() {
	r.a.Reset()
}

// Run runs the running aggregator, listens for incoming metrics, and waits
// for period ticks to tell it when to push and reset the aggregator.
func (r *RunningAggregator) Run(
	acc telegraf.Accumulator,
	shutdown chan struct{},
) {
	// The start of the period is truncated to the nearest second.
	//
	// Every metric then gets it's timestamp checked and is dropped if it
	// is not within:
	//
	//   start < t < end + truncation + delay
	//
	// So if we start at now = 00:00.2 with a 10s period and 0.3s delay:
	//   now = 00:00.2
	//   start = 00:00
	//   truncation = 00:00.2
	//   end = 00:10
	// 1st interval: 00:00 - 00:10.5
	// 2nd interval: 00:10 - 00:20.5
	// etc.
	//
	now := time.Now()
	r.periodStart = now.Truncate(time.Second)
	truncation := now.Sub(r.periodStart)
	r.periodEnd = r.periodStart.Add(r.Config.Period)
	time.Sleep(r.Config.Delay)
	periodT := time.NewTicker(r.Config.Period)
	defer periodT.Stop()

	for {
		select {
		case <-shutdown:
			if len(r.metrics) > 0 {
				// wait until metrics are flushed before exiting
				continue
			}
			return
		case m := <-r.metrics:
			if m.Time().Before(r.periodStart) ||
				m.Time().After(r.periodEnd.Add(truncation).Add(r.Config.Delay)) {
				// the metric is outside the current aggregation period, so
				// skip it.
				log.Printf("D! aggregator: metric \"%s\" is not in the current timewindow, skipping", m.Name())
				continue
			}
			r.add(m)
		case <-periodT.C:
			r.periodStart = r.periodEnd
			r.periodEnd = r.periodStart.Add(r.Config.Period)
			r.push(acc)
			r.reset()
		}
	}
}
