package models

import (
	"time"

	"github.com/influxdata/telegraf"
)

type RunningAggregator struct {
	Aggregator telegraf.Aggregator
	Config     *AggregatorConfig
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

// Apply applies the given metric to the aggregator.
// Before applying to the plugin, it will run any defined filters on the metric.
// Apply returns true if the original metric should be dropped.
func (r *RunningAggregator) Apply(in telegraf.Metric) bool {
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

	r.Aggregator.Apply(in)
	return r.Config.DropOriginal
}
