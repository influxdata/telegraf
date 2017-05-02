// selfstat is a package for tracking and collecting internal statistics
// about telegraf. Metrics can be registered using this package, and then
// incremented or set within your code. If the inputs.internal plugin is enabled,
// then all registered stats will be collected as they would by any other input
// plugin.
package selfstat

import (
	"hash/fnv"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var (
	registry *rgstry
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

	// Key is the unique measurement+tags key of the stat.
	Key() uint64

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
	return registry.register(&stat{
		measurement: "internal_" + measurement,
		field:       field,
		tags:        tags,
	})
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
	return registry.register(&timingStat{
		measurement: "internal_" + measurement,
		field:       field,
		tags:        tags,
	})
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
			metric, err := metric.New(name, tags, fields, now)
			if err != nil {
				log.Printf("E! Error creating selfstat metric: %s", err)
				continue
			}
			metrics[i] = metric
			i++
		}
	}
	registry.mu.Unlock()
	return metrics
}

type rgstry struct {
	stats map[uint64]map[string]Stat
	mu    sync.Mutex
}

func (r *rgstry) register(s Stat) Stat {
	r.mu.Lock()
	defer r.mu.Unlock()
	if stats, ok := r.stats[s.Key()]; ok {
		// measurement exists
		if stat, ok := stats[s.FieldName()]; ok {
			// field already exists, so don't create a new one
			return stat
		}
		r.stats[s.Key()][s.FieldName()] = s
		return s
	} else {
		// creating a new unique metric
		r.stats[s.Key()] = map[string]Stat{s.FieldName(): s}
		return s
	}
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
	registry = &rgstry{
		stats: make(map[uint64]map[string]Stat),
	}
}
