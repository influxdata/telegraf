package models

import (
	"time"

	"github.com/influxdata/telegraf"
)

type RunningAggregator struct {
	a      telegraf.Aggregator
	Config *AggregatorConfig

	metrics chan telegraf.Metric
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
		r.Config.Filter,
		false,
		false,
		mType,
		t,
	)

	m.SetAggregate(true)

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

		in, _ = telegraf.NewMetric(name, tags, fields, t)
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

func (r *RunningAggregator) Run(
	acc telegraf.Accumulator,
	shutdown chan struct{},
) {
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
			r.add(m)
		case <-periodT.C:
			r.push(acc)
			r.reset()
		}
	}
}
