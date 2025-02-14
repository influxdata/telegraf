package vsphere

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
)

// tsCache is a cache of timestamps used to determine the validity of datapoints
type tsCache struct {
	ttl   time.Duration
	table map[string]time.Time
	mux   sync.RWMutex
	log   telegraf.Logger
}

// newTSCache creates a new tsCache with a specified time-to-live after which timestamps are discarded.
func newTSCache(ttl time.Duration, log telegraf.Logger) *tsCache {
	return &tsCache{
		ttl:   ttl,
		table: make(map[string]time.Time),
		log:   log,
	}
}

// purge removes timestamps that are older than the time-to-live
func (t *tsCache) purge() {
	t.mux.Lock()
	defer t.mux.Unlock()
	n := 0
	for k, v := range t.table {
		if time.Since(v) > t.ttl {
			delete(t.table, k)
			n++
		}
	}
	t.log.Debugf("purged timestamp cache. %d deleted with %d remaining", n, len(t.table))
}

// get returns a timestamp (if present)
func (t *tsCache) get(key, metricName string) (time.Time, bool) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	ts, ok := t.table[makeKey(key, metricName)]
	return ts, ok
}

// put updates the latest timestamp for the supplied key.
func (t *tsCache) put(key, metricName string, timestamp time.Time) {
	t.mux.Lock()
	defer t.mux.Unlock()
	k := makeKey(key, metricName)
	if timestamp.After(t.table[k]) {
		t.table[k] = timestamp
	}
}

func makeKey(resource, metric string) string {
	return resource + "|" + metric
}
