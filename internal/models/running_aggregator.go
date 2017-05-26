package models

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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

// AggregatorConfig containing configuration parameters for the running
// aggregator plugin.
type AggregatorConfig struct {
	Name string

	DropOriginal      bool
	NameOverride      string
	MeasurementPrefix string
	MeasurementSuffix string
	Tags              map[string]string
	Filter            Filter

	Period time.Duration
	Delay  time.Duration
}

func (r *RunningAggregator) Name() string {
	return "aggregators." + r.Config.Name
}

func (r *RunningAggregator) MakeMetric(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	mType telegraf.ValueType,
	t time.Time,
) telegraf.Metric {
	m := makemetric(
		measurement,
		fields,
		tags,
		r.Config.NameOverride,
		r.Config.MeasurementPrefix,
		r.Config.MeasurementSuffix,
		r.Config.Tags,
		nil,
		nil,
		r.Config.Filter,
		false,
		mType,
		t,
	)

	if m != nil {
		m.SetAggregate(true)
	}

	return m
}

// Add applies the given metric to the aggregator.
// Before applying to the plugin, it will run any defined filters on the metric.
// Apply returns true if the original metric should be dropped.
func (r *RunningAggregator) Add(in telegraf.Metric) bool {
	if r.Config.Filter.IsActive() {
		// check if the aggregator should apply this metric
		name := in.Name()
		fields := in.Fields()
		tags := in.Tags()
		t := in.Time()
		if ok := r.Config.Filter.Apply(name, fields, tags); !ok {
			// aggregator should not apply this metric
			return false
		}

		in, _ = metric.New(name, tags, fields, t)
	}

	r.metrics <- in
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
