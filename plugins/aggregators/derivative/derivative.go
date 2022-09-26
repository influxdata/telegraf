//go:generate ../../../tools/readme_config_includer/generator
package derivative

import (
	_ "embed"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

//go:embed sample.conf
var sampleConfig string

type Derivative struct {
	Variable    string          `toml:"variable"`
	Suffix      string          `toml:"suffix"`
	MaxRollOver uint            `toml:"max_roll_over"`
	Log         telegraf.Logger `toml:"-"`
	cache       map[uint64]*aggregate
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

const defaultSuffix = "_rate"

func NewDerivative() *Derivative {
	derivative := &Derivative{Suffix: defaultSuffix, MaxRollOver: 10}
	derivative.cache = make(map[uint64]*aggregate)
	derivative.Reset()
	return derivative
}

func (*Derivative) SampleConfig() string {
	return sampleConfig
}

func (d *Derivative) Add(in telegraf.Metric) {
	id := in.HashID()
	current, ok := d.cache[id]
	if !ok {
		// hit an uncached metric, create caches for first time:
		d.cache[id] = newAggregate(in)
		return
	}
	if current.first.time.After(in.Time()) {
		current.first = newEvent(in)
		current.rollOver = 0
	} else if current.first.time.Equal(in.Time()) {
		upsertConvertedFields(in.Fields(), current.first.fields)
		current.rollOver = 0
	}
	if current.last.time.Before(in.Time()) {
		current.last = newEvent(in)
		current.rollOver = 0
	} else if current.last.time.Equal(in.Time()) {
		upsertConvertedFields(in.Fields(), current.last.fields)
		current.rollOver = 0
	}
}

func newAggregate(in telegraf.Metric) *aggregate {
	event := newEvent(in)
	return &aggregate{
		name:     in.Name(),
		tags:     in.Tags(),
		first:    event,
		last:     event,
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
	upsertConvertedFields(in.Fields(), fields)
	return fields
}

func upsertConvertedFields(source map[string]interface{}, target map[string]float64) {
	for k, v := range source {
		if value, ok := convert(v); ok {
			target[k] = value
		}
	}
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case uint64:
		return float64(v), true
	}
	return 0, false
}

func (d *Derivative) Push(acc telegraf.Accumulator) {
	for _, aggregate := range d.cache {
		if aggregate.first == aggregate.last {
			d.Log.Debugf("Same first and last event for %q, skipping.", aggregate.name)
			continue
		}
		var denominator float64
		denominator = aggregate.last.time.Sub(aggregate.first.time).Seconds()
		if len(d.Variable) > 0 {
			var first float64
			var last float64
			var found bool
			if first, found = aggregate.first.fields[d.Variable]; !found {
				d.Log.Debugf("Did not find %q in first event for %q.", d.Variable, aggregate.name)
				continue
			}
			if last, found = aggregate.last.fields[d.Variable]; !found {
				d.Log.Debugf("Did not find %q in last event for %q.", d.Variable, aggregate.name)
				continue
			}
			denominator = last - first
		}
		if denominator == 0 {
			d.Log.Debugf("Got difference 0 in denominator for %q, skipping.", aggregate.name)
			continue
		}
		derivatives := make(map[string]interface{})
		for key, start := range aggregate.first.fields {
			if key == d.Variable {
				// Skip derivation variable
				continue
			}
			if end, ok := aggregate.last.fields[key]; ok {
				d.Log.Debugf("Adding derivative %q to %q.", key+d.Suffix, aggregate.name)
				derivatives[key+d.Suffix] = (end - start) / denominator
			}
		}
		acc.AddFields(aggregate.name, derivatives, aggregate.tags)
	}
}

func (d *Derivative) Reset() {
	for id, aggregate := range d.cache {
		if aggregate.rollOver < d.MaxRollOver {
			aggregate.first = aggregate.last
			aggregate.rollOver = aggregate.rollOver + 1
			d.cache[id] = aggregate
			d.Log.Debugf("Roll-Over %q for the %d time.", aggregate.name, aggregate.rollOver)
		} else {
			delete(d.cache, id)
			d.Log.Debugf("Removed %q from cache.", aggregate.name)
		}
	}
}

func (d *Derivative) Init() error {
	d.Suffix = strings.TrimSpace(d.Suffix)
	d.Variable = strings.TrimSpace(d.Variable)
	return nil
}

func init() {
	aggregators.Add("derivative", func() telegraf.Aggregator {
		return NewDerivative()
	})
}
