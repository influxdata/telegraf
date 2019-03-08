package vsphere

import (
	"log"
	"sync"
	"time"
)

// TSCache is a cache of timestamps used to determine the validity of datapoints
type TSCache struct {
	ttl   time.Duration
	table map[string]time.Time
	done  chan struct{}
	mux   sync.RWMutex
}

// NewTSCache creates a new TSCache with a specified time-to-live after which timestamps are discarded.
func NewTSCache(ttl time.Duration) *TSCache {
	return &TSCache{
		ttl:   ttl,
		table: make(map[string]time.Time),
		done:  make(chan struct{}),
	}
}

// Purge removes timestamps that are older than the time-to-live
func (t *TSCache) Purge() {
	t.mux.Lock()
	defer t.mux.Unlock()
	n := 0
	for k, v := range t.table {
		if time.Now().Sub(v) > t.ttl {
			delete(t.table, k)
			n++
		}
	}
	log.Printf("D! [inputs.vsphere] Purged timestamp cache. %d deleted with %d remaining", n, len(t.table))
}

// IsNew returns true if the supplied timestamp for the supplied key is more recent than the
// timestamp we have on record.
func (t *TSCache) IsNew(key string, tm time.Time) bool {
	t.mux.RLock()
	defer t.mux.RUnlock()
	v, ok := t.table[key]
	if !ok {
		return true // We've never seen this before, so consider everything a new sample
	}
	return !tm.Before(v)
}

// Get returns a timestamp (if present)
func (t *TSCache) Get(key string) (time.Time, bool) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	ts, ok := t.table[key]
	return ts, ok
}

// Put updates the latest timestamp for the supplied key.
func (t *TSCache) Put(key string, time time.Time) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.table[key] = time
}
