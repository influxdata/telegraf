package vsphere

import (
	"log"
	"sync"
	"time"
)

type TSCache struct {
	ttl   time.Duration
	table map[string]time.Time
	done  chan struct{}
	mux   sync.RWMutex
}

func NewTSCache(ttl time.Duration) *TSCache {
	t := &TSCache{
		ttl:   ttl,
		table: make(map[string]time.Time),
		done:  make(chan struct{}),
	}
	go func(t *TSCache) {
		tick := time.NewTicker(time.Minute)
		defer tick.Stop()
		for {
			select {
			case <-t.done:
				return
			case <-tick.C:
				t.purge()
			}
		}
	}(t)
	return t
}

func (t *TSCache) purge() {
	t.mux.Lock()
	defer t.mux.Unlock()
	n := 0
	for k, v := range t.table {
		if time.Now().Sub(v) > t.ttl {
			delete(t.table, k)
			n++
		}
	}
	log.Printf("D! [input.vsphere] Purged timestamp cache. %d deleted with %d remaining", n, len(t.table))
}

func (t *TSCache) IsNew(key string, tm time.Time) bool {
	t.mux.RLock()
	defer t.mux.RUnlock()
	v, ok := t.table[key]
	if !ok {
		return true // We've never seen this before, so consider everything a new sample
	}
	return !tm.Before(v)
}

func (t *TSCache) Put(key string, time time.Time) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.table[key] = time
}

func (t *TSCache) Destroy() {
	close(t.done)
}
