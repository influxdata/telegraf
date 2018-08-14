package models

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

type RunningAggregator struct {
	sync.Mutex
	Aggregator  telegraf.Aggregator
	Config      *AggregatorConfig
	periodStart time.Time
	periodEnd   time.Time

	MetricsPushed   selfstat.Stat
	MetricsFiltered selfstat.Stat
	MetricsDropped  selfstat.Stat
	PushTime        selfstat.Stat
}

func NewRunningAggregator(
	aggregator telegraf.Aggregator,
	config *AggregatorConfig,
) *RunningAggregator {
	return &RunningAggregator{
		Aggregator: aggregator,
		Config:     config,
		MetricsPushed: selfstat.Register(
			"aggregate",
			"metrics_pushed",
			map[string]string{"aggregator": config.Name},
		),
		MetricsFiltered: selfstat.Register(
			"aggregate",
			"metrics_filtered",
			map[string]string{"aggregator": config.Name},
		),
		MetricsDropped: selfstat.Register(
			"aggregate",
			"metrics_dropped",
			map[string]string{"aggregator": config.Name},
		),
		PushTime: selfstat.Register(
			"aggregate",
			"push_time_ns",
			map[string]string{"aggregator": config.Name},
		),
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

func (r *RunningAggregator) Period() time.Duration {
	return r.Config.Period
}

func (r *RunningAggregator) SetPeriodStart(start time.Time) {
	r.periodStart = start
	r.periodEnd = r.periodStart.Add(r.Config.Period).Add(r.Config.Delay)
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

	r.MetricsPushed.Incr(1)

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

	r.Lock()
	defer r.Unlock()

	if r.periodStart.IsZero() || metric.Time().Before(r.periodStart) || metric.Time().After(r.periodEnd) {
		r.MetricsDropped.Incr(1)
		return false
	}

	r.Aggregator.Add(metric)
	return r.Config.DropOriginal
}

func (r *RunningAggregator) Push(acc telegraf.Accumulator) {
	r.Lock()
	defer r.Unlock()

	r.periodStart = r.periodEnd
	r.periodEnd = r.periodStart.Add(r.Config.Period).Add(r.Config.Delay)
	r.push(acc)
	r.Aggregator.Reset()
}

func (r *RunningAggregator) push(acc telegraf.Accumulator) {
	start := time.Now()
	r.Aggregator.Push(acc)
	elapsed := time.Since(start)
	r.PushTime.Incr(elapsed.Nanoseconds())
}
