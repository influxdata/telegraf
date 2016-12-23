package agent

import (
	"log"
	"time"

	"github.com/influxdata/telegraf/plugins"
	"github.com/influxdata/telegraf/selfstat"
)

var (
	NErrors = selfstat.Register("agent", "gather_errors", map[string]string{})
)

type MetricMaker interface {
	Name() string
	MakeMetric(
		measurement string,
		fields map[string]interface{},
		tags map[string]string,
		mType plugins.ValueType,
		t time.Time,
	) plugins.Metric
}

func NewAccumulator(
	maker MetricMaker,
	metrics chan plugins.Metric,
) *accumulator {
	acc := accumulator{
		maker:     maker,
		metrics:   metrics,
		precision: time.Nanosecond,
	}
	return &acc
}

type accumulator struct {
	metrics chan plugins.Metric

	maker MetricMaker

	precision time.Duration
}

func (ac *accumulator) AddFields(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	if m := ac.maker.MakeMetric(measurement, fields, tags, plugins.Untyped, ac.getTime(t)); m != nil {
		ac.metrics <- m
	}
}

func (ac *accumulator) AddGauge(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	if m := ac.maker.MakeMetric(measurement, fields, tags, plugins.Gauge, ac.getTime(t)); m != nil {
		ac.metrics <- m
	}
}

func (ac *accumulator) AddCounter(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	if m := ac.maker.MakeMetric(measurement, fields, tags, plugins.Counter, ac.getTime(t)); m != nil {
		ac.metrics <- m
	}
}

// AddError passes a runtime error to the accumulator.
// The error will be tagged with the plugin name and written to the log.
func (ac *accumulator) AddError(err error) {
	if err == nil {
		return
	}
	NErrors.Incr(1)
	//TODO suppress/throttle consecutive duplicate errors?
	log.Printf("E! Error in plugin [%s]: %s", ac.maker.Name(), err)
}

// SetPrecision takes two time.Duration objects. If the first is non-zero,
// it sets that as the precision. Otherwise, it takes the second argument
// as the order of time that the metrics should be rounded to, with the
// maximum being 1s.
func (ac *accumulator) SetPrecision(precision, interval time.Duration) {
	if precision > 0 {
		ac.precision = precision
		return
	}
	switch {
	case interval >= time.Second:
		ac.precision = time.Second
	case interval >= time.Millisecond:
		ac.precision = time.Millisecond
	case interval >= time.Microsecond:
		ac.precision = time.Microsecond
	default:
		ac.precision = time.Nanosecond
	}
}

func (ac accumulator) getTime(t []time.Time) time.Time {
	var timestamp time.Time
	if len(t) > 0 {
		timestamp = t[0]
	} else {
		timestamp = time.Now()
	}
	return timestamp.Round(ac.precision)
}
