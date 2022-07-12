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
	## If more than this amount of time passes between data points, the
	## previous value will be considered old and the value will be recalculated
	## as if it hadn't been seen before. A zero expiration means never expire.
	# expiration = "0s"

	## The operation that should be performed between two observed points.
	## It can be 'diff' or 'rate'
	# transform = "rate"

	## For the fields who's key/value pairs don't match, should the original
	## field be removed?
	# remove-original = true

[processors.t128_transform.fields]
	## Replace fields with their computed values, renaming them if indicated
	# "/rate/metric" = "/total/metric"
	# "/inline/replace" = "/inline/replace"
`

type transformer = func(t1, t2 time.Time, v1, v2 float64) (float64, error)

type T128Transform struct {
	Fields         map[string]string `toml:"fields"`
	Expiration     config.Duration   `toml:"expiration"`
	RemoveOriginal bool              `toml:"remove-original"`
	Transform      string            `toml:"transform"`

	Log telegraf.Logger `toml:"-"`

	transform    transformer
	targetFields map[string]target
	cache        map[uint64]map[string]observedValue
}

type target struct {
	key           string
	matchesSource bool
}

type observedValue struct {
	value     float64
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

			currentValue, converted := convert(field.Value)
			if !converted {
				r.Log.Warnf("Failed to convert field '%s' to float for transformation. This transformation will be skipped.")
				continue
			}

			cacheFields, metricIsCached := r.cache[seriesHash]
			if !metricIsCached {
				r.cache[seriesHash] = make(map[string]observedValue, 0)
			}

			itemAdded := false
			if observed, ok := cacheFields[field.Key]; ok {
				if point.Time().Before(observed.expires) {
					value, err := r.transform(observed.timestamp, point.Time(), observed.value, currentValue)
					if err != nil {
						r.Log.Warnf("excluding failed transform: %v", err)
					} else {
						point.AddField(target.key, value)
						itemAdded = true
					}
				}
			}

			if (target.matchesSource && !itemAdded) || (!target.matchesSource && r.RemoveOriginal) {
				removeFields = append(removeFields, field.Key)
			}

			r.cache[seriesHash][field.Key] = observedValue{
				value:     currentValue,
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
		r.transform = func(t1, t2 time.Time, v1, v2 float64) (float64, error) {
			return v2 - v1, nil
		}
	case "rate":
		r.transform = func(t1, t2 time.Time, v1, v2 float64) (float64, error) {
			if !t1.Before(t2) {
				return 0, fmt.Errorf(
					"asked to compute the rate between points with non-increasing timestamps: %v at %v and %v at %v",
					v1, t1, v2, t2,
				)
			}

			return (v2 - v1) / (t2.Sub(t1).Seconds()), nil
		}
	default:
		return fmt.Errorf("'transform' is required and must be 'diff' or 'rate'")
	}

	for dest, src := range r.Fields {
		if target, ok := r.targetFields[src]; ok {
			// For simple testing
			conflicting := []string{dest, target.key}
			sort.Strings(conflicting)

			return fmt.Errorf("both '%s' and '%s' are configured to be calculated from '%s'", conflicting[0], conflicting[1], src)
		}

		r.targetFields[src] = target{
			key:           dest,
			matchesSource: src == dest,
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
	return &T128Transform{
		Transform:    "rate",
		targetFields: make(map[string]target),
		cache:        make(map[uint64]map[string]observedValue),
	}
}

func init() {
	processors.Add("t128_transform", func() telegraf.Processor {
		return newTransform()
	})
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}
