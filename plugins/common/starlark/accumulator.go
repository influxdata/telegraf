package starlark //nolint

import (
	"errors"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	starlarktime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
)

type Accumulator struct {
	accumulator telegraf.Accumulator
}

// Wrap updates the starlark.Accumulator to wrap a new telegraf.Accumulator.
func (a *Accumulator) Wrap(accumulator telegraf.Accumulator) {
	a.accumulator = accumulator
}

// Unwrap removes the telegraf.Accumulator from the startlark.Accumulator.
func (a *Accumulator) Unwrap() telegraf.Accumulator {
	return a.accumulator
}

// String returns the starlark representation of the Accumulator.
func (a *Accumulator) String() string {
	return a.Type()
}

func (a *Accumulator) Type() string {
	return "Accumulator"
}

func (a *Accumulator) Freeze() {
}

func (a *Accumulator) Truth() starlark.Bool {
	return true
}

func (a *Accumulator) Hash() (uint32, error) {
	return 0, errors.New("not hashable")
}

const (
	FIELDS = iota
	GAUGE
	COUNTER
	SUMMARY
	HISTOGRAM
)

var accumulatorMethods = map[string]builtinMethod{
	"add_counter":   addCounter,
	"add_fields":    addFields,
	"add_gauge":     addGauge,
	"add_histogram": addHistogram,
	"add_metric":    addMetric,
	"add_summary":   addSummary,
	"set_precision": setPrecision,
}

func (a *Accumulator) Attr(name string) (starlark.Value, error) {
	return builtinAttr(a, name, accumulatorMethods)
}

func (a *Accumulator) AttrNames() []string {
	return builtinAttrNames(accumulatorMethods)
}

func add(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple, typeToAdd int) (starlark.Value, error) {
	var (
		measurement              string
		fields, tags, timestamps starlark.Value
	)

	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 3, &measurement, &fields, &tags, &timestamps); err != nil {
		return starlark.None, err
	}

	allFields, err := toFields(fields)
	if err != nil {
		return nil, err
	}
	allTags, err := toTags(tags)
	if err != nil {
		return nil, err
	}
	allTimestamps, err := toSlice(timestamps)
	if err != nil {
		return nil, err
	}
	acc := b.Receiver().(*Accumulator).Unwrap()
	switch typeToAdd {
	case FIELDS:
		acc.AddFields(measurement, allFields, allTags, allTimestamps...)
	case GAUGE:
		acc.AddGauge(measurement, allFields, allTags, allTimestamps...)
	case COUNTER:
		acc.AddCounter(measurement, allFields, allTags, allTimestamps...)
	case SUMMARY:
		acc.AddSummary(measurement, allFields, allTags, allTimestamps...)
	case HISTOGRAM:
		acc.AddHistogram(measurement, allFields, allTags, allTimestamps...)
	}
	return starlark.None, nil
}

func toSlice(timestamps starlark.Value) ([]time.Time, error) {
	if timestamps == nil {
		return nil, nil
	}
	switch x := timestamps.(type) {
	case starlarktime.Time:
		return []time.Time{time.Time(x)}, nil
	case starlark.Sequence:
		result := make([]time.Time, x.Len())
		iter := x.Iterate()
		defer iter.Done()
		var v starlark.Value
		i := 0
		for iter.Next(&v) {
			timestamp, ok := v.(starlarktime.Time)
			if !ok {
				return nil, fmt.Errorf("Only timestamps are expected, found %T", v)
			}
			result[i] = time.Time(timestamp)
			i++
		}
		return result, nil
	}
	return nil, fmt.Errorf("The timestamp argument should be one single timestamp or a sequence of timestamps, found %T", timestamps)
}

func addFields(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return add(b, args, kwargs, FIELDS)
}

func addGauge(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return add(b, args, kwargs, GAUGE)
}

func addCounter(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return add(b, args, kwargs, COUNTER)
}

func addSummary(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return add(b, args, kwargs, SUMMARY)
}

func addHistogram(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return add(b, args, kwargs, HISTOGRAM)
}

func addMetric(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		metric *Metric
	)

	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &metric); err != nil {
		return starlark.None, err
	}

	acc := b.Receiver().(*Accumulator).Unwrap()
	acc.AddMetric(metric.Unwrap())
	return starlark.None, nil
}

func setPrecision(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		precision starlarktime.Duration
	)

	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &precision); err != nil {
		return starlark.None, err
	}

	acc := b.Receiver().(*Accumulator).Unwrap()
	acc.SetPrecision(time.Duration(precision))
	return starlark.None, nil
}

func toFields(value starlark.Value) (map[string]interface{}, error) {
	items, err := items(value, "The type %T is unsupported as type of collection of fields")
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{}, len(items))
	for _, item := range items {
		key, err := toString(item[0], "The type %T is unsupported as type of key for fields")
		if err != nil {
			return nil, err
		}
		value, err := asGoValue(item[1])
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

func toTags(value starlark.Value) (map[string]string, error) {
	items, err := items(value, "The type %T is unsupported as type of collection of tags")
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(items))
	for _, item := range items {
		key, err := toString(item[0], "The type %T is unsupported as type of key for tags")
		if err != nil {
			return nil, err
		}
		value, err := toString(item[1], "The type %T is unsupported as type of value for tags")
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

func toString(value starlark.Value, errorMsg string) (string, error) {
	if value, ok := value.(starlark.String); ok {
		return string(value), nil
	}
	return "", fmt.Errorf(errorMsg, value)
}

func items(value starlark.Value, errorMsg string) ([]starlark.Tuple, error) {
	if iter, ok := value.(starlark.IterableMapping); ok {
		return iter.Items(), nil
	}
	return nil, fmt.Errorf(errorMsg, value)
}
