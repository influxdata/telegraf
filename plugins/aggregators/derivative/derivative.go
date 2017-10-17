package derivative

import (
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type Derivative struct {
	Variable    string
	Infix       string
	MaxRollOver uint
	cache       map[uint64]aggregate
}

type aggregate struct {
	first    *event
	last     *event
	name     string
	tags     map[string]string
	rollOver uint
}

type event struct {
	fields map[string]float64
	time   time.Time
}

func NewDerivative() telegraf.Aggregator {
	derivative := &Derivative{Infix: "_by_", MaxRollOver: 10}
	derivative.cache = make(map[uint64]aggregate)
	derivative.Reset()
	return derivative
}

var sampleConfig = `
  ## This Aggregator will estimate a derivative for each field, which is
	## contained in both the first and last metric of the aggregation interval.
  ## Without further configuration the derivative will be calculated with
	## respect to the time difference between these two measurements in seconds.
	## The formula applied is for every field:
	##
	##               value_last - value_first
	## derivative = --------------------------
  ##              time_difference_in_seconds
  ##
	## The resulting derivative will be named *fieldname_by_seconds*. The infix
	## "_by_" can be configured by the *infix* parameter.
	# infix = "_wrt_"
	##
	## As an abstraction the derivative can be calculated not only by the time
	## difference but by the difference of a field, which is contained in the
	## measurement. This field is assumed to be monotonously increasing. This
	## feature is used by specifying a *variable*.
	# variable = "parameter"
	##
	## When using a field as the derivation parameter the name of that field will
	## be used for the resulting derivative, e.g. *fieldname_by_parameter*.
	##
	## As a genereal parameter the period needs to be given. It determines the
	## period for which a derivative will be calculated
	period = "30s"
	##
	## Note, that the calculation is based on the actual timestamp of the
	## measurements. When there is only one measurement during that period, the
	## measurement will be rolled over to the next period. The maximum number of
	## such roll-overs can be configured with a default of 10.
	# max_roll_over = 10
	##
`

func (d *Derivative) SampleConfig() string {
	return sampleConfig
}

func (d *Derivative) Description() string {
	return "Calculates a derivative for every field."
}

func (d *Derivative) Add(in telegraf.Metric) {
	id := in.HashID()
	if current, ok := d.cache[id]; !ok {
		// hit an uncached metric, create caches for first time:
		d.cache[id] = newAggregate(in)
	} else {
		if current.first.time.After(in.Time()) {
			current.first = newEvent(in)
			current.rollOver = 0
		}
		if current.first.time.Equal(in.Time()) {
			current.first.fields = upsertConvertedFields(in.Fields(), current.first.fields)
			current.rollOver = 0
		}
		if current.last.time.Before(in.Time()) {
			current.last = newEvent(in)
			current.rollOver = 0
		}
		if current.last.time.Equal(in.Time()) {
			current.last.fields = upsertConvertedFields(in.Fields(), current.last.fields)
			current.rollOver = 0
		}

		d.cache[id] = current
	}
}

func newAggregate(in telegraf.Metric) aggregate {
	return aggregate{
		name:     in.Name(),
		tags:     in.Tags(),
		first:    newEvent(in),
		last:     newEvent(in),
		rollOver: 0,
	}
}

func newEvent(in telegraf.Metric) *event {
	return &event{
		fields: extractConvertedFields(in),
		time:   in.Time(),
	}
}

func extractConvertedFields(in telegraf.Metric) map[string]float64 {
	fields := make(map[string]float64, len(in.Fields()))
	return upsertConvertedFields(in.Fields(), fields)
}

func upsertConvertedFields(source map[string]interface{}, target map[string]float64) map[string]float64 {
	for k, v := range source {
		if value, ok := convert(v); ok {
			target[k] = value
		}
	}
	return target
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func (d *Derivative) variableFieldName() string {
	return strings.TrimSpace(d.Variable)
}

func (d *Derivative) Push(acc telegraf.Accumulator) {
	for _, aggregate := range d.cache {
		if aggregate.first != aggregate.last {
			if denominator := d.calculateDenominator(aggregate); denominator != 0 {
				derivatives := make(map[string]interface{})
				for key, start := range aggregate.first.fields {
					if end, ok := aggregate.last.fields[key]; key != d.variableFieldName() && ok {
						derivatives[d.derivativeFieldName(key)] = (end - start) / denominator
					}
				}
				acc.AddFields(aggregate.name, derivatives, aggregate.tags)
			}
		}
	}
}

func (d *Derivative) calculateDenominator(aggregate aggregate) float64 {
	if variable, present := d.derivationVariableName(); present {
		return aggregate.last.fields[variable] - aggregate.first.fields[variable]
	}
	return aggregate.last.time.Sub(aggregate.first.time).Seconds()
}

func (d *Derivative) derivationVariableName() (string, bool) {
	return d.variableFieldName(), len(d.variableFieldName()) != 0
}

func (d *Derivative) derivativeFieldName(field string) string {
	if param := d.variableFieldName(); len(param) != 0 {
		return field + d.trimmedInfix() + param
	}
	return field + d.trimmedInfix() + "seconds"
}

func (d *Derivative) trimmedInfix() string {
	return strings.TrimSpace(d.Infix)
}

func (d *Derivative) Reset() {
	for id, aggregate := range d.cache {
		if aggregate.rollOver < d.MaxRollOver {
			aggregate.first = aggregate.last
			aggregate.rollOver = aggregate.rollOver + 1
			d.cache[id] = aggregate
		} else {
			delete(d.cache, id)
		}
	}
}

func init() {
	aggregators.Add("derivative", func() telegraf.Aggregator {
		return NewDerivative()
	})
}
