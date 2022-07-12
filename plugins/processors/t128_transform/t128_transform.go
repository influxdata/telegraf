package t128_transform

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
	## For 'rate' and 'diff', if more than this amount of time passes between
	## data points, the previous value will be considered old and the value will
	## be recalculated as if it hadn't been seen before. A zero expiration means
	## never expire.
	##
	## When using the 'state-change' transform, an update metric will be sent
	## upon expiration even if the value has not changed.
	# expiration = "0s"

	## The operation that should be performed between two observed points.
	## It can be 'diff', 'rate', or 'state-change'.
	# transform = "rate"

	## For the fields who's key/value pairs don't match, should the original
	## field be removed?
	# remove-original = true

	## Specify a field to be populated with the last produced value. If the
	## field name is an empty string or there is no prior value, the field will
	## be excluded.
	# previous_field = ""

[processors.t128_transform.fields]
	## Replace fields with their computed values, renaming them if indicated
	# "/rate/metric" = "/total/metric"
	# "/inline/replace" = "/inline/replace"

[processors.t128_transform.previous_fields]
	## Populate these fields with the previous transformed value. If there is no
	## prior value, the field will be excluded.
	# "/rate/metric/previous" = "/rate/metric"
`

type transformer = func(expired bool, t1, t2 time.Time, v1, v2 interface{}) (
	value interface{}, recordAsPrevious bool, err error,
)

type T128Transform struct {
	Fields         map[string]string `toml:"fields"`
	PreviousFields map[string]string `toml:"previous_fields"`
	Expiration     config.Duration   `toml:"expiration"`
	RemoveOriginal bool              `toml:"remove-original"`
	Transform      string            `toml:"transform"`

	Log telegraf.Logger `toml:"-"`

	transform    transformer
	targetFields map[string]*target
	cache        map[uint64]map[string]observedValue
}

type target struct {
	key           string
	previousKey   string
	matchesSource bool
}

type observedValue struct {
	value interface{}
	// previous produced (transformed) value, not previous observed
	// (the two would be the same in some cases)
	previous  interface{}
	expires   time.Time
	timestamp time.Time
}

func (r *T128Transform) SampleConfig() string {
	return sampleConfig
}

func (r *T128Transform) Description() string {
	return "Transform metrics based on the differences between two observed points."
}

func (r *T128Transform) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		seriesHash := point.HashID()

		removeFields := make([]string, 0)

		for _, field := range point.FieldList() {
			target, ok := r.targetFields[field.Key]
			if !ok {
				continue
			}

			cacheFields, metricIsCached := r.cache[seriesHash]
			if !metricIsCached {
				r.cache[seriesHash] = make(map[string]observedValue, 0)
			}

			observed, ok := cacheFields[field.Key]
			if !ok {
				observed = observedValue{
					value: nil,
				}
			}

			expired := !point.Time().Before(observed.expires)

			itemTransformed := false
			value, recordAsPrevious, err := r.transform(
				expired,
				observed.timestamp,
				point.Time(),
				observed.value,
				field.Value,
			)
			if err != nil {
				r.Log.Warnf("excluding failed transform: %v", err)
			} else if value != nil {
				itemTransformed = true
				point.AddField(target.key, value)
			}

			if (target.matchesSource && !itemTransformed) || (!target.matchesSource && r.RemoveOriginal) {
				removeFields = append(removeFields, field.Key)
			}

			if itemTransformed && target.previousKey != "" && observed.previous != nil {
				point.AddField(target.previousKey, observed.previous)
			}

			newPrevious := observed.previous
			if recordAsPrevious {
				newPrevious = value
			}

			r.cache[seriesHash][field.Key] = observedValue{
				value:     field.Value,
				previous:  newPrevious,
				expires:   point.Time().Add(time.Duration(r.Expiration)),
				timestamp: point.Time(),
			}
		}

		for _, fieldKey := range removeFields {
			point.RemoveField(fieldKey)
		}
	}

	return in
}

func (r *T128Transform) Init() error {
	if len(r.Fields) == 0 {
		return fmt.Errorf("at least one value must be specified in the 'fields' list")
	}

	switch r.Transform {
	case "diff":
		r.transform = func(expired bool, t1, t2 time.Time, v1, v2 interface{}) (interface{}, bool, error) {
			if expired || v1 == nil {
				return nil, true, nil
			}

			prev, current, err := convertToFloats(v1, v2)
			if err != nil {
				return 0, false, err
			}

			return current - prev, true, nil
		}
	case "rate":
		r.transform = func(expired bool, t1, t2 time.Time, v1, v2 interface{}) (interface{}, bool, error) {
			if expired || v1 == nil {
				return nil, true, nil
			}

			if !t1.Before(t2) {
				return 0, false, fmt.Errorf(
					"asked to compute the rate between points with non-increasing timestamps: %v at %v and %v at %v",
					v1, t1, v2, t2,
				)
			}

			prev, current, err := convertToFloats(v1, v2)
			if err != nil {
				return 0, false, err
			}

			return (current - prev) / (t2.Sub(t1).Seconds()), true, nil
		}
	case "state-change":
		r.transform = func(expired bool, t1, t2 time.Time, v1, v2 interface{}) (interface{}, bool, error) {
			if expired || v1 == nil {
				return v2, true, nil
			}

			if v1 != v2 {
				return v2, true, nil
			}

			return nil, false, nil
		}
	default:
		return fmt.Errorf("'transform' is required and must be 'diff', 'rate', or 'state-change'")
	}

	for dest, src := range r.Fields {
		if target, ok := r.targetFields[src]; ok {
			// For simple testing
			conflicting := []string{dest, target.key}
			sort.Strings(conflicting)

			return fmt.Errorf("both '%s' and '%s' are configured to be calculated from '%s'", conflicting[0], conflicting[1], src)
		}

		r.targetFields[src] = &target{
			key:           dest,
			matchesSource: src == dest,
		}
	}

	for previous, original := range r.PreviousFields {
		if src, ok := r.Fields[original]; ok {
			if target, ok := r.targetFields[src]; ok {
				target.previousKey = previous
			} else {
				return fmt.Errorf("failed to lookup the target for previous field '%v' which is based on '%v' (developer error)", previous, original)
			}
		} else {
			return fmt.Errorf("the previous field '%v' references a transformed field '%v' which does not exist", previous, original)
		}
	}

	if r.Expiration == 0 {
		// No expiration means never expire, so set to maximum duration
		r.Expiration = math.MaxInt64
	} else {
		// If the time difference matches, don't expire. Adjusting here makes
		// later math easier.
		r.Expiration++
	}

	return nil
}

func newTransform() *T128Transform {
	return newTransformType("rate")
}

func newTransformType(transformType string) *T128Transform {
	return &T128Transform{
		Transform:      transformType,
		PreviousFields: make(map[string]string),
		targetFields:   make(map[string]*target),
		cache:          make(map[uint64]map[string]observedValue),
	}
}

func init() {
	processors.Add("t128_transform", func() telegraf.Processor {
		return newTransform()
	})
}

func convertToFloats(a, b interface{}) (float64, float64, error) {
	v1, err := convertToFloat(a)
	if err != nil {
		return 0, 0, err
	}

	v2, err := convertToFloat(b)
	if err != nil {
		return 0, 0, err
	}

	return v1, v2, nil
}

func convertToFloat(in interface{}) (float64, error) {
	switch v := in.(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("Failed to convert field '%s' to float for transformation. This transformation will be skipped.", in)
	}
}
