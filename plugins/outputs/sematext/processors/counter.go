package processors

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/sematext/tags"
	"math"
	"sync"
	"time"
)

// HandleCounter keeps track of previous values of counter-type metrics and converts their value into delta between
// current and previous measurement.
type HandleCounter struct {
	lock          sync.Mutex
	countersCache map[string]*counterCacheEntry
	lastCleared   time.Time
}

type counterCacheEntry struct {
	lastValue    interface{}
	lastRecorded time.Time
}

// NewHandleCounter creates a new HandleCounter processor
func NewHandleCounter() MetricProcessor {
	return &HandleCounter{
		lastCleared:   time.Now(),
		countersCache: make(map[string]*counterCacheEntry),
	}
}

func (h *HandleCounter) Process(metric telegraf.Metric) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	// if metric is a counter, keep track of its last recording, change the current value to be delta
	if getSematextMetricType(metric.Type()) == Counter {
		for _, field := range metric.FieldList() {
			key := metric.Name() + "." + field.Key + "-" + tags.GetTagsKey(metric.Tags())
			prevValueEntry := h.countersCache[key]
			currValue := field.Value

			var delta interface{}
			if prevValueEntry == nil {
				delta = getZeroValue(currValue)

				h.countersCache[key] = &counterCacheEntry{
					lastValue:    currValue,
					lastRecorded: metric.Time(),
				}
			} else {
				delta = calculateDelta(prevValueEntry.lastValue, currValue)

				prevValueEntry.lastValue = currValue
				prevValueEntry.lastRecorded = metric.Time()
			}

			field.Value = delta
		}
	}

	h.clearCounterCache()

	return nil
}

// clearCounterCache once a day goes through all entries in the cache map and removes all that were not used for the
// past 24 hours
func (h *HandleCounter) clearCounterCache() {
	var now = time.Now()
	if hoursSince(now, h.lastCleared) >= 24 {
		newCountersCache := make(map[string]*counterCacheEntry)

		for k, v := range h.countersCache {
			if hoursSince(now, h.countersCache[k].lastRecorded) < 24 {
				newCountersCache[k] = v
			}
		}

		h.countersCache = newCountersCache
		h.lastCleared = time.Now()
	}
}

func hoursSince(end time.Time, start time.Time) float64 {
	return (end.Sub(start)).Hours()
}

func getZeroValue(value interface{}) interface{} {
	switch value.(type) {
	case string:
		return ""
	case bool:
		return false
	case float64:
		return 0.0
	case uint64:
		return 0
	case int64:
		return 0
	default:
		return 0
	}
}

func calculateDelta(prevValue interface{}, currValue interface{}) interface{} {
	switch v := prevValue.(type) {
	case string:
		return currValue
	case bool:
		return currValue
	case float64:
		if !math.IsNaN(v) && !math.IsInf(v, 0) && !math.IsNaN(currValue.(float64)) && !math.IsInf(currValue.(float64), 0) {
			delta := currValue.(float64) - prevValue.(float64)
			if delta < 0 {
				delta = 0
			}
			return delta
		}

		return 0.0
	case uint64:
		if currValue.(uint64) < prevValue.(uint64) {
			return 0
		}
		return currValue.(uint64) - prevValue.(uint64)
	case int64:
		delta := currValue.(int64) - prevValue.(int64)
		if delta < 0 {
			delta = 0
		}
		return delta
	default:
		return 0
	}
}

func (h *HandleCounter) Close() {
}
