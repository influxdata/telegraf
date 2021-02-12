package models

import (
	"context"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/selfstat"
)

type RunningAggregator struct {
	sync.Mutex

	ID          uint64
	Aggregator  telegraf.Aggregator
	Config      *AggregatorConfig
	periodStart time.Time
	periodEnd   time.Time
	log         telegraf.Logger

	MetricsPushed   selfstat.Stat
	MetricsFiltered selfstat.Stat
	MetricsDropped  selfstat.Stat
	PushTime        selfstat.Stat
	State

	RoundInterval bool // comes from config

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewRunningAggregator(aggregator telegraf.Aggregator, config *AggregatorConfig) *RunningAggregator {
	tags := map[string]string{"aggregator": config.Name}
	if config.Alias != "" {
		tags["alias"] = config.Alias
	}

	aggErrorsRegister := selfstat.Register("aggregate", "errors", tags)
	logger := NewLogger("aggregators", config.Name, config.Alias)
	logger.OnErr(func() {
		aggErrorsRegister.Incr(1)
	})

	SetLoggerOnPlugin(aggregator, logger)

	a := &RunningAggregator{
		ID:         nextPluginID(),
		Aggregator: aggregator,
		Config:     config,
		MetricsPushed: selfstat.Register(
			"aggregate",
			"metrics_pushed",
			tags,
		),
		MetricsFiltered: selfstat.Register(
			"aggregate",
			"metrics_filtered",
			tags,
		),
		MetricsDropped: selfstat.Register(
			"aggregate",
			"metrics_dropped",
			tags,
		),
		PushTime: selfstat.Register(
			"aggregate",
			"push_time_ns",
			tags,
		),
		log: logger,
	}
	a.setState(PluginStateCreated)
	return a
}

// AggregatorConfig is the common config for all aggregators.
type AggregatorConfig struct {
	Name         string        `toml:"name"`
	Alias        string        `toml:"alias"`
	DropOriginal bool          `toml:"drop_original"`
	Period       time.Duration `toml:"period"`
	Delay        time.Duration `toml:"delay"`
	Grace        time.Duration `toml:"grace"`

	NameOverride      string            `toml:"name_override"`
	MeasurementPrefix string            `toml:"measurement_prefix"`
	MeasurementSuffix string            `toml:"measurement_suffix"`
	Tags              map[string]string `toml:"tags"`
	Filter            Filter            `toml:"filter"`
	Order             int64             `toml:"order"`
}

func (r *RunningAggregator) Order() int64 {
	return r.Config.Order
}

func (r *RunningAggregator) LogName() string {
	return logName("aggregators", r.Config.Name, r.Config.Alias)
}

func (r *RunningAggregator) Init() error {
	if p, ok := r.Aggregator.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

// func (r *RunningAggregator) Period() time.Duration {
// 	return r.Config.Period
// }

// func (r *RunningAggregator) EndPeriod() time.Time {
// 	return r.periodEnd
// }

func (r *RunningAggregator) updateWindow(start, until time.Time) {
	r.periodStart = start
	r.periodEnd = until
	r.log.Debugf("Updated aggregation range [%s, %s]", start, until)
}

func (r *RunningAggregator) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	return metric
}

// Add a metric to the aggregator and return true if the original metric
// should be dropped.
func (r *RunningAggregator) Add(m telegraf.Metric, acc telegraf.Accumulator) error {
	defer func() {
		if !r.Config.DropOriginal && len(m.FieldList()) > 0 {
			// m.Drop()
			acc.AddMetric(m)
		}
	}()

	if ok := r.Config.Filter.Select(m); !ok {
		return nil
	}

	// Make a copy of the metric but don't retain tracking.  We do not fail a
	// delivery due to the aggregation not being sent because we can't create
	// aggregations of historical data.  Additionally, waiting for the
	// aggregation to be pushed would introduce a hefty latency to delivery.
	m = metric.FromMetric(m)

	r.Config.Filter.Modify(m)
	if len(m.FieldList()) == 0 {
		r.MetricsFiltered.Incr(1)
		m.Drop()
		return nil
	}

	r.Lock()
	defer r.Unlock()

	// check if outside agg window
	if m.Time().Before(r.periodStart.Add(-r.Config.Grace)) || m.Time().After(r.periodEnd.Add(r.Config.Delay)) {
		r.log.Debugf("Metric is outside aggregation window; discarding. %s: m: %s e: %s g: %s",
			m.Time(), r.periodStart, r.periodEnd, r.Config.Grace)
		r.MetricsDropped.Incr(1)
		return nil
	}

	r.Aggregator.Add(m)
	return nil
}

func (r *RunningAggregator) Push(acc telegraf.Accumulator) {
	r.Lock()
	defer r.Unlock()

	since := r.periodEnd
	until := r.periodEnd.Add(r.Config.Period)
	r.updateWindow(since, until)
	r.push(acc.WithNewMetricMaker(r.LogName(), r.Log(), r.pushMetricMaker))
	r.Aggregator.Reset()
}

// not passed on to agg at this time
func (r *RunningAggregator) Start(acc telegraf.Accumulator) error {
	r.setState(PluginStateRunning)

	since, until := r.calculateUpdateWindow(time.Now())
	r.updateWindow(since, until)

	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.wg.Add(1)
	go r.pushLoop(acc)

	return nil
}

// not passed on to agg at this time
func (r *RunningAggregator) Stop() {
	r.setState(PluginStateStopping)
	r.cancel()
	r.wg.Wait()
	r.setState(PluginStateDead)
}

func (r *RunningAggregator) push(acc telegraf.Accumulator) {
	start := time.Now()
	r.Aggregator.Push(acc)
	elapsed := time.Since(start)
	r.PushTime.Incr(elapsed.Nanoseconds())
}

func (r *RunningAggregator) Log() telegraf.Logger {
	return r.log
}

// Before calling Add, initialize the aggregation window.  This ensures
// that any metric created after start time will be aggregated.
func (r *RunningAggregator) calculateUpdateWindow(start time.Time) (since time.Time, until time.Time) {
	if r.RoundInterval {
		until = internal.AlignTime(start, r.Config.Period)
		if until == start {
			until = internal.AlignTime(start.Add(time.Nanosecond), r.Config.Period)
		}
	} else {
		until = start.Add(r.Config.Period)
	}

	since = until.Add(-r.Config.Period)

	return since, until
}

func (r *RunningAggregator) pushLoop(acc telegraf.Accumulator) {
	for {
		// Ensures that Push will be called for each period, even if it has
		// already elapsed before this function is called.  This is guaranteed
		// because so long as only Push updates the EndPeriod.  This method
		// also avoids drift by not using a ticker.
		until := time.Until(r.periodEnd)

		select {
		case <-time.After(until):
			r.Push(acc)
			break
		case <-r.ctx.Done():
			r.Push(acc)
			r.wg.Done()
			return
		}
	}
}

func (r *RunningAggregator) pushMetricMaker(metric telegraf.Metric) telegraf.Metric {
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

func (r *RunningAggregator) GetID() uint64 {
	return r.ID
}
