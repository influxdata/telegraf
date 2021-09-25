// Package selfstat is a package for tracking and collecting internal statistics
// about telegraf. Metrics can be registered using this package, and then
// incremented or set within your code. If the inputs.internal plugin is enabled,
// then all registered stats will be collected as they would by any other input
// plugin.
package selfstat

import (
	"hash/fnv"
	"sort"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var (
	registry *Registry
)

// Stat is an interface for dealing with telegraf statistics collected
// on itself.
type Stat interface {
	// Name is the name of the measurement
	Name() string

	// FieldName is the name of the measurement field
	FieldName() string

	// Tags is a tag map. Each time this is called a new map is allocated.
	Tags() map[string]string

	// Incr increments a regular stat by 'v'.
	// in the case of a timing stat, increment adds the timing to the cache.
	Incr(v int64)

	// Set sets a regular stat to 'v'.
	// in the case of a timing stat, set adds the timing to the cache.
	Set(v int64)

	// Get gets the value of the stat. In the case of timings, this returns
	// an average value of all timings received since the last call to Get().
	// If no timings were received, it returns the previous value.
	Get() int64
}

// Register registers the given measurement, field, and tags in the selfstat
// registry. If given an identical measurement, it will return the stat that's
// already been registered.
//
// The returned Stat can be incremented by the consumer of Register(), and it's
// value will be returned as a telegraf metric when Metrics() is called.
func Register(measurement, field string, tags map[string]string) Stat {
	return registry.register("internal_"+measurement, field, tags)
}

// RegisterTiming registers the given measurement, field, and tags in the selfstat
// registry. If given an identical measurement, it will return the stat that's
// already been registered.
//
// Timing stats differ from regular stats in that they accumulate multiple
// "timings" added to them, and will return the average when Get() is called.
// After Get() is called, the average is cleared and the next timing returned
// from Get() will only reflect timings added since the previous call to Get().
// If Get() is called without receiving any new timings, then the previous value
// is used.
//
// In other words, timings are an averaged metric that get cleared on each call
// to Get().
//
// The returned Stat can be incremented by the consumer of Register(), and it's
// value will be returned as a telegraf metric when Metrics() is called.
func RegisterTiming(measurement, field string, tags map[string]string) Stat {
	return registry.registerTiming("internal_"+measurement, field, tags)
}

// Metrics returns all registered stats as telegraf metrics.
func Metrics() []telegraf.Metric {
	registry.mu.Lock()
	now := time.Now()
	metrics := make([]telegraf.Metric, len(registry.stats))
	i := 0
	for _, stats := range registry.stats {
		if len(stats) > 0 {
			var tags map[string]string
			var name string
			fields := map[string]interface{}{}
			j := 0
			for fieldname, stat := range stats {
				if j == 0 {
					tags = stat.Tags()
					name = stat.Name()
				}
				fields[fieldname] = stat.Get()
				j++
			}
			m := metric.New(name, tags, fields, now)
			metrics[i] = m
			i++
		}
	}
	registry.mu.Unlock()
	return metrics
}

type Registry struct {
	stats map[uint64]map[string]Stat
	mu    sync.Mutex
}

func (r *Registry) register(measurement, field string, tags map[string]string) Stat {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := key(measurement, tags)
	if stat, ok := registry.get(key, field); ok {
		return stat
	}

	t := make(map[string]string, len(tags))
	for k, v := range tags {
		t[k] = v
	}

	s := &stat{
		measurement: measurement,
		field:       field,
		tags:        t,
	}
	registry.set(key, s)
	return s
}

func (r *Registry) registerTiming(measurement, field string, tags map[string]string) Stat {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := key(measurement, tags)
	if stat, ok := registry.get(key, field); ok {
		return stat
	}

	t := make(map[string]string, len(tags))
	for k, v := range tags {
		t[k] = v
	}

	s := &timingStat{
		measurement: measurement,
		field:       field,
		tags:        t,
	}
	registry.set(key, s)
	return s
}

func (r *Registry) get(key uint64, field string) (Stat, bool) {
	if _, ok := r.stats[key]; !ok {
		return nil, false
	}

	if stat, ok := r.stats[key][field]; ok {
		return stat, true
	}

	return nil, false
}

func (r *Registry) set(key uint64, s Stat) {
	if _, ok := r.stats[key]; !ok {
		r.stats[key] = make(map[string]Stat)
	}

	r.stats[key][s.FieldName()] = s
}

func key(measurement string, tags map[string]string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(measurement))

	tmp := make([]string, len(tags))
	i := 0
	for k, v := range tags {
		tmp[i] = k + v
		i++
	}
	sort.Strings(tmp)

	for _, s := range tmp {
		h.Write([]byte(s))
	}

	return h.Sum64()
}

func init() {
	registry = &Registry{
		stats: make(map[uint64]map[string]Stat),
	}
}
