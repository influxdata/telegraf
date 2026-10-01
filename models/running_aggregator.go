package models

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	logging "github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/selfstat"
)

type RunningAggregator struct {
	sync.Mutex
	Aggregator  telegraf.Aggregator
	Config      *AggregatorConfig
	periodStart time.Time
	periodEnd   time.Time
	log         telegraf.Logger

	MetricsPushed   selfstat.Stat
	MetricsFiltered selfstat.Stat
	MetricsDropped  selfstat.Stat
	PushTime        selfstat.Stat
}

func NewRunningAggregator(aggregator telegraf.Aggregator, config *AggregatorConfig) *RunningAggregator {
	tags := map[string]string{
		"_id":        config.ID,
		"aggregator": config.Name,
	}
	if config.Alias != "" {
		tags["alias"] = config.Alias
	}

	aggErrorsRegister := selfstat.Register("aggregate", "errors", tags)
	logger := logging.New("aggregators", config.Name, config.Alias)
	logger.RegisterErrorCallback(func() {
		aggErrorsRegister.Incr(1)
	})
	if err := logger.SetLogLevel(config.LogLevel); err != nil {
		logger.Error(err)
	}
	SetLoggerOnPlugin(aggregator, logger)
	SetStatisticsOnPlugin(aggregator, logger, tags)

	return &RunningAggregator{
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
}

// AggregatorConfig is the common config for all aggregators.
type AggregatorConfig struct {
	Name         string
	Source       string
	Alias        string
	ID           string
	DropOriginal bool
	Period       time.Duration
	Delay        time.Duration
	Grace        time.Duration
	LogLevel     string

	NameOverride      string
	MeasurementPrefix string
	MeasurementSuffix string
	Tags              map[string]string
	Filter            Filter
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

func (r *RunningAggregator) ID() string {
	if p, ok := r.Aggregator.(telegraf.PluginWithID); ok {
		return p.ID()
	}
	return r.Config.ID
}

func (r *RunningAggregator) Period() time.Duration {
	return r.Config.Period
}

func (r *RunningAggregator) EndPeriod() time.Time {
	r.Lock()
	defer r.Unlock()
	return r.periodEnd
}

func (r *RunningAggregator) UpdateWindow(start, until time.Time) {
	r.periodStart = start
	r.periodEnd = until
	r.log.Debugf("Updated aggregation range [%s, %s]", start, until)
}

func (r *RunningAggregator) MakeMetric(telegrafMetric telegraf.Metric) telegraf.Metric {
	m := makeMetric(
		telegrafMetric,
		r.Config.NameOverride,
		r.Config.MeasurementPrefix,
		r.Config.MeasurementSuffix,
		r.Config.Tags,
		nil)

	r.MetricsPushed.Incr(1)

	return m
}

// Add a metric to the aggregator and return true if the original metric
// should be dropped.
func (r *RunningAggregator) Add(m telegraf.Metric) bool {
	ok, err := r.Config.Filter.Select(m)
	if err != nil {
		r.log.Errorf("filtering failed: %v", err)
	} else if !ok {
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

	if m.Time().Before(r.periodStart.Add(-r.Config.Grace)) || m.Time().After(r.periodEnd.Add(r.Config.Delay)) {
		r.log.Debugf("Metric is outside aggregation window; discarding. %s: m: %s e: %s g: %s",
			m.Time(), r.periodStart, r.periodEnd, r.Config.Grace)
		r.MetricsDropped.Incr(1)
		return r.Config.DropOriginal
	}

	r.Aggregator.Add(m)
	return r.Config.DropOriginal
}

func (r *RunningAggregator) Push(acc telegraf.Accumulator) {
	collector := &pushCollector{log: r.log, precision: time.Nanosecond}

	r.Lock()

	since := r.periodEnd
	until := r.periodEnd.Add(r.Config.Period)

	// Check if the next aggregation window will contain "now". This might
	// not be the case if the machine's clock was adjusted or the machine
	// hibernated as in those cases the clock might be advanced before or
	// after the initial aggregation window.
	nowWall := time.Now().Truncate(-1)
	if nowWall.Before(since.Truncate(-1)) || nowWall.After(until.Truncate(-1)) {
		since = nowWall.Truncate(r.Config.Period)
		until = since.Add(r.Config.Period)
	}

	r.UpdateWindow(since, until)

	start := time.Now()
	// Push into an in-memory collector rather than the real accumulator.
	// The real accumulator ultimately blocks on a bounded channel to the
	// next stage; if that channel is momentarily full (e.g. a slow output,
	// or a large/high-cardinality aggregation taking a while to drain),
	// doing that send here would keep the lock held for the whole time,
	// starving concurrent Add() calls -- and, transitively, any input
	// plugin blocked delivering a metric via Add(). Collecting first lets
	// us release the lock before delivery.
	r.Aggregator.Push(collector)
	elapsed := time.Since(start)
	r.PushTime.Incr(elapsed.Nanoseconds())
	r.Aggregator.Reset()

	r.Unlock()

	// Deliver the collected metrics without holding the lock. MakeMetric
	// (name/tag prefixing, MetricsPushed accounting) and precision rounding
	// still happen exactly once, inside acc.AddMetric, same as before.
	for _, m := range collector.metrics {
		acc.AddMetric(m)
	}
}

// pushCollector is a telegraf.Accumulator that buffers metrics produced by
// an aggregator's Push() call in memory instead of forwarding them
// immediately. See RunningAggregator.Push for why this decoupling matters.
type pushCollector struct {
	log       telegraf.Logger
	precision time.Duration
	metrics   []telegraf.Metric
}

func (c *pushCollector) AddFields(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {
	c.append(metric.New(measurement, tags, fields, c.getTime(t), telegraf.Untyped))
}

func (c *pushCollector) AddGauge(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {
	c.append(metric.New(measurement, tags, fields, c.getTime(t), telegraf.Gauge))
}

func (c *pushCollector) AddCounter(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {
	c.append(metric.New(measurement, tags, fields, c.getTime(t), telegraf.Counter))
}

func (c *pushCollector) AddSummary(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {
	c.append(metric.New(measurement, tags, fields, c.getTime(t), telegraf.Summary))
}

func (c *pushCollector) AddHistogram(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {
	c.append(metric.New(measurement, tags, fields, c.getTime(t), telegraf.Histogram))
}

func (c *pushCollector) AddMetric(m telegraf.Metric) {
	m.SetTime(m.Time().Round(c.precision))
	c.append(m)
}

func (c *pushCollector) append(m telegraf.Metric) {
	if m != nil {
		c.metrics = append(c.metrics, m)
	}
}

func (c *pushCollector) SetPrecision(precision time.Duration) {
	c.precision = precision
}

func (c *pushCollector) AddError(err error) {
	if err == nil {
		return
	}
	if c.log != nil {
		c.log.Errorf("Error in plugin: %v", err)
	}
}

func (c *pushCollector) getTime(t []time.Time) time.Time {
	var timestamp time.Time
	if len(t) > 0 {
		timestamp = t[0]
	} else {
		timestamp = time.Now()
	}
	return timestamp.Round(c.precision)
}

func (c *pushCollector) WithTracking(maxTracked int) telegraf.TrackingAccumulator {
	return &pushTrackingCollector{
		pushCollector: c,
		delivered:     make(chan telegraf.DeliveryInfo, maxTracked),
	}
}

// pushTrackingCollector supports the rare case of an aggregator requesting a
// TrackingAccumulator from within Push(). Tracked metrics are still buffered
// like any other metric; delivery information is reported immediately since
// aggregated metrics are not tied to the original inputs' delivery guarantees.
type pushTrackingCollector struct {
	*pushCollector
	delivered chan telegraf.DeliveryInfo
}

func (c *pushTrackingCollector) AddTrackingMetric(m telegraf.Metric) telegraf.TrackingID {
	dm, id := metric.WithTracking(m, c.onDelivery)
	c.AddMetric(dm)
	return id
}

func (c *pushTrackingCollector) AddTrackingMetricGroup(group []telegraf.Metric) telegraf.TrackingID {
	db, id := metric.WithGroupTracking(group, c.onDelivery)
	for _, m := range db {
		c.AddMetric(m)
	}
	return id
}

func (c *pushTrackingCollector) Delivered() <-chan telegraf.DeliveryInfo {
	return c.delivered
}

func (c *pushTrackingCollector) onDelivery(info telegraf.DeliveryInfo) {
	select {
	case c.delivered <- info:
	default:
	}
}

func (r *RunningAggregator) Log() telegraf.Logger {
	return r.log
}
