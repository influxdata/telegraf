package models

import (
	"log"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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

func (r *RunningAggregator) EndPeriod() time.Time {
	return r.periodEnd
}

func (r *RunningAggregator) UpdateWindow(start, until time.Time) {
	r.periodStart = start
	r.periodEnd = until
	log.Printf("D! [%s] Updated aggregation range [%s, %s]", r.Name(), start, until)
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
func (r *RunningAggregator) Add(m telegraf.Metric) bool {
	if ok := r.Config.Filter.Select(m); !ok {
		return false
	}

	// Make a copy of the metric but don't retain tracking.  We do not fail a
	// delivery due to the aggregation not being sent because we can't create
	// aggregations of historical data.  Additionally, waiting for the
	// aggregation to be pushed would introduce a hefty latency to delivery.
	m = metric.FromMetric(m)

	r.Config.Filter.Modify(m)
	if len(m.FieldList()) == 0 {
		r.MetricsFiltered.Incr(1)
		return r.Config.DropOriginal
	}

	r.Lock()
	defer r.Unlock()

	if m.Time().Before(r.periodStart) || m.Time().After(r.periodEnd.Add(r.Config.Delay)) {
		log.Printf("D! [%s] metric is outside aggregation window; discarding. %s: m: %s e: %s",
			r.Name(), m.Time(), r.periodStart, r.periodEnd)
		r.MetricsDropped.Incr(1)
		return r.Config.DropOriginal
	}

	r.Aggregator.Add(m)
	return r.Config.DropOriginal
}

func (r *RunningAggregator) Push(acc telegraf.Accumulator) {
	r.Lock()
	defer r.Unlock()

	since := r.periodEnd
	until := r.periodEnd.Add(r.Config.Period)
	r.UpdateWindow(since, until)

	r.push(acc)
	r.Aggregator.Reset()
}

func (r *RunningAggregator) push(acc telegraf.Accumulator) {
	start := time.Now()
	r.Aggregator.Push(acc)
	elapsed := time.Since(start)
	r.PushTime.Incr(elapsed.Nanoseconds())
}
