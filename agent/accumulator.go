package agent

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type MetricMaker interface {
	LogName() string
	MakeMetric(metric telegraf.Metric) telegraf.Metric
	Log() telegraf.Logger
}

type accumulator struct {
	sync.Mutex
	sync.Cond
	maker     MetricMaker
	metrics   chan<- telegraf.Metric
	precision time.Duration
}

func NewAccumulator(
	maker MetricMaker,
	metrics chan<- telegraf.Metric,
) telegraf.Accumulator {
	acc := accumulator{
		maker:     maker,
		metrics:   metrics,
		precision: time.Nanosecond,
	}
	acc.Cond.L = &acc.Mutex

	return &acc
}

func (ac *accumulator) AddFields(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addFields(measurement, tags, fields, telegraf.Untyped, t...)
}

func (ac *accumulator) AddGauge(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addFields(measurement, tags, fields, telegraf.Gauge, t...)
}

func (ac *accumulator) AddCounter(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addFields(measurement, tags, fields, telegraf.Counter, t...)
}

func (ac *accumulator) AddSummary(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addFields(measurement, tags, fields, telegraf.Summary, t...)
}

func (ac *accumulator) AddHistogram(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addFields(measurement, tags, fields, telegraf.Histogram, t...)
}

func (ac *accumulator) AddMetric(m telegraf.Metric) {
	ac.Lock()
	defer ac.Unlock()
	m.SetTime(m.Time().Round(ac.precision))
	if m := ac.maker.MakeMetric(m); m != nil {
		if ac.metrics == nil {
			ac.Cond.Wait() // unlock and wait for metrics to be set.
		}
		ac.metrics <- m
	}
}

func (ac *accumulator) addFields(
	measurement string,
	tags map[string]string,
	fields map[string]interface{},
	tp telegraf.ValueType,
	t ...time.Time,
) {
	m, err := metric.New(measurement, tags, fields, ac.getTime(t), tp)
	if err != nil {
		return
	}
	ac.Lock()
	defer ac.Unlock()
	if m := ac.maker.MakeMetric(m); m != nil {
		if ac.metrics == nil {
			ac.Cond.Wait() // unlock and wait for metrics to be set.
		}
		ac.metrics <- m
	}
}

// setOutput changes the destination of the accumulator
func (ac *accumulator) setOutput(outCh chan<- telegraf.Metric) {
	ac.Lock()
	defer ac.Unlock()
	ac.metrics = outCh
	ac.Cond.Broadcast()
}

// AddError passes a runtime error to the accumulator.
// The error will be tagged with the plugin name and written to the log.
func (ac *accumulator) AddError(err error) {
	if err == nil {
		return
	}
	ac.maker.Log().Errorf("Error in plugin: %v", err)
}

func (ac *accumulator) SetPrecision(precision time.Duration) {
	ac.precision = precision
}

func (ac *accumulator) getTime(t []time.Time) time.Time {
	var timestamp time.Time
	if len(t) > 0 {
		timestamp = t[0]
	} else {
		timestamp = time.Now()
	}
	return timestamp.Round(ac.precision)
}

func (ac *accumulator) WithTracking(maxTracked int) telegraf.TrackingAccumulator {
	return &trackingAccumulator{
		Accumulator: ac,
		delivered:   make(chan telegraf.DeliveryInfo, maxTracked),
	}
}

func (ac *accumulator) WithNewMetricMaker(logName string, logger telegraf.Logger, f func(metric telegraf.Metric) telegraf.Metric) telegraf.Accumulator {
	return &metricMakerAccumulator{
		Accumulator: ac,
		makerFunc:   f,
		logName:     logName,
		logger:      logger,
	}
}

type metricMakerAccumulator struct {
	telegraf.Accumulator
	logName   string
	logger    telegraf.Logger
	makerFunc func(m telegraf.Metric) telegraf.Metric
}

func (ma *metricMakerAccumulator) LogName() string {
	return ma.logName
}
func (ma *metricMakerAccumulator) Log() telegraf.Logger {
	return ma.logger
}
func (ma *metricMakerAccumulator) MakeMetric(m telegraf.Metric) telegraf.Metric {
	return ma.makerFunc(m)
}

type trackingAccumulator struct {
	telegraf.Accumulator
	delivered chan telegraf.DeliveryInfo
}

func (a *trackingAccumulator) AddTrackingMetric(m telegraf.Metric) telegraf.TrackingID {
	dm, id := metric.WithTracking(m, a.onDelivery)
	a.AddMetric(dm)
	return id
}

func (a *trackingAccumulator) AddTrackingMetricGroup(group []telegraf.Metric) telegraf.TrackingID {
	db, id := metric.WithGroupTracking(group, a.onDelivery)
	for _, m := range db {
		a.AddMetric(m)
	}
	return id
}

func (a *trackingAccumulator) Delivered() <-chan telegraf.DeliveryInfo {
	return a.delivered
}

func (a *trackingAccumulator) onDelivery(info telegraf.DeliveryInfo) {
	select {
	case a.delivered <- info:
	default:
		// This is a programming error in the input.  More items were sent for
		// tracking than space requested.
		panic("channel is full")
	}
}
