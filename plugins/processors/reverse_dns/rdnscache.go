package reverse_dns

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"
)

const defaultMaxWorkers = 10

var (
	errTimeout = errors.New("request timed out")
)

// AnyResolver is for the net.Resolver
type AnyResolver interface {
	// LookupAddr performs a reverse lookup for the given address, returning a list of names mapping to that address.
	LookupAddr(ctx context.Context, addr string) (names []string, err error)
}

// reverseDNSCache is safe to use across multiple goroutines.
// If multiple goroutines request the same IP at the same time, one of the requests will trigger the lookup and the rest will wait for its response.
type reverseDNSCache struct {
	resolver AnyResolver
	stats    rDNSCacheStats

	// settings
	ttl           time.Duration
	lookupTimeout time.Duration
	maxWorkers    int

	// internal
	rwLock              sync.RWMutex
	sem                 *semaphore.Weighted
	cancelCleanupWorker context.CancelFunc

	cache map[string]*dnslookup

	// keep an ordered list of what needs to be worked on and what is due to expire.
	// We can use this list for both with a job position marker, and by popping items
	// off the list as they expire. This avoids iterating over the whole map to find
	// things to do.
	// As a bonus, we only have to read the first item to know if anything in the
	// map has expired.
	// must lock to get access to this.
	expireList     []*dnslookup
	expireListLock sync.Mutex
}

type rDNSCacheStats struct {
	cacheHit          uint64
	cacheMiss         uint64
	cacheExpire       uint64
	requestsAbandoned uint64
	requestsFilled    uint64
}

func newReverseDNSCache(ttl, lookupTimeout time.Duration, workerPoolSize int) *reverseDNSCache {
	if workerPoolSize <= 0 {
		workerPoolSize = defaultMaxWorkers
	}
	ctx, cancel := context.WithCancel(context.Background())
	d := &reverseDNSCache{
		ttl:                 ttl,
		lookupTimeout:       lookupTimeout,
		cache:               make(map[string]*dnslookup),
		maxWorkers:          workerPoolSize,
		sem:                 semaphore.NewWeighted(int64(workerPoolSize)),
		cancelCleanupWorker: cancel,
		resolver:            net.DefaultResolver,
	}
	d.startCleanupWorker(ctx)
	return d
}

// dnslookup represents a lookup request/response. It may or may not be answered yet.
// interested parties register themselves with existing requests or create new ones
// to get their dns query answered. Answers will be pushed out to callbacks.
type dnslookup struct {
	ip        string // keep a copy for the expireList.
	domains   []string
	expiresAt time.Time
	completed bool
	callbacks []callbackChannelType
}

type lookupResult struct {
	domains []string
	err     error
}

type callbackChannelType chan lookupResult

// lookup takes a string representing a parseable ipv4 or ipv6 IP, and blocks
// until it has resolved to 0-n results, or until its lookup timeout has elapsed.
// if the lookup timeout elapses, it returns an empty slice.
func (d *reverseDNSCache) lookup(ip string) ([]string, error) {
	if len(ip) == 0 {
		return nil, nil
	}

	// check if the value is cached
	d.rwLock.RLock()
	result, found := d.lockedGetFromCache(ip)
	if found && result.completed && !result.expiresAt.Before(time.Now()) {
		defer d.rwLock.RUnlock()
		atomic.AddUint64(&d.stats.cacheHit, 1)
		// cache is valid
		return result.domains, nil
	}
	d.rwLock.RUnlock()

	// if it's not cached, kick off a lookup job and subscribe to the result.
	lookupChan := d.subscribeTo(ip)
	timer := time.NewTimer(d.lookupTimeout)
	defer timer.Stop()

	// timer is still necessary even if doLookup respects timeout due to worker
	// pool starvation.
	select {
	case result := <-lookupChan:
		return result.domains, result.err
	case <-timer.C:
		return nil, errTimeout
	}
}

func (d *reverseDNSCache) subscribeTo(ip string) callbackChannelType {
	callback := make(callbackChannelType, 1)

	d.rwLock.Lock()
	defer d.rwLock.Unlock()

	// confirm it's still not in the cache. This needs to be done under an active lock.
	result, found := d.lockedGetFromCache(ip)
	if found {
		atomic.AddUint64(&d.stats.cacheHit, 1)
		// has the request been answered since we last checked?
		if result.completed {
			// we can return the answer with the channel.
			callback <- lookupResult{domains: result.domains}
			return callback
		}
		// there's a request but it hasn't been answered yet;
		// add yourself to the subscribers and return that.
		result.callbacks = append(result.callbacks, callback)
		d.lockedSaveToCache(result)
		return callback
	}

	atomic.AddUint64(&d.stats.cacheMiss, 1)

	// otherwise we need to register the request
	l := &dnslookup{
		ip:        ip,
		expiresAt: time.Now().Add(d.ttl),
		callbacks: []callbackChannelType{callback},
	}

	d.lockedSaveToCache(l)
	go d.doLookup(l.ip)
	return callback
}

// lockedGetFromCache fetches from the correct internal ip cache.
// you MUST first do a read or write lock before calling it, and keep locks around
// the dnslookup that is returned until you clone it.
func (d *reverseDNSCache) lockedGetFromCache(ip string) (lookup *dnslookup, found bool) {
	lookup, found = d.cache[ip]
	if found && !lookup.expiresAt.After(time.Now()) {
		return nil, false
	}
	return lookup, found
}

// lockedSaveToCache stores a lookup in the correct internal ip cache.
// you MUST first do a write lock before calling it.
func (d *reverseDNSCache) lockedSaveToCache(lookup *dnslookup) {
	if !lookup.expiresAt.After(time.Now()) {
		return // don't cache.
	}
	d.cache[lookup.ip] = lookup
}

func (d *reverseDNSCache) startCleanupWorker(ctx context.Context) {
	go func() {
		cleanupTick := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-cleanupTick.C:
				d.cleanup()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (d *reverseDNSCache) doLookup(ip string) {
	ctx, cancel := context.WithTimeout(context.Background(), d.lookupTimeout)
	defer cancel()
	if err := d.sem.Acquire(ctx, 1); err != nil {
		// lookup timeout
		d.abandonLookup(ip, errTimeout)
		return
	}
	defer d.sem.Release(1)

	names, err := d.resolver.LookupAddr(ctx, ip)
	if err != nil {
		d.abandonLookup(ip, err)
		return
	}

	d.rwLock.Lock()
	lookup, found := d.lockedGetFromCache(ip)
	if !found {
		d.rwLock.Unlock()
		return
	}

	lookup.domains = names
	lookup.completed = true
	lookup.expiresAt = time.Now().Add(d.ttl) // extend the ttl now that we have a reply.
	callbacks := lookup.callbacks
	lookup.callbacks = nil

	d.lockedSaveToCache(lookup)
	d.rwLock.Unlock()

	d.expireListLock.Lock()
	// add it to the expireList.
	d.expireList = append(d.expireList, lookup)
	d.expireListLock.Unlock()

	atomic.AddUint64(&d.stats.requestsFilled, uint64(len(callbacks)))
	for _, cb := range callbacks {
		cb <- lookupResult{domains: names}
		close(cb)
	}
}

func (d *reverseDNSCache) abandonLookup(ip string, err error) {
	d.rwLock.Lock()
	lookup, found := d.lockedGetFromCache(ip)
	if !found {
		d.rwLock.Unlock()
		return
	}

	callbacks := lookup.callbacks
	delete(d.cache, lookup.ip)
	d.rwLock.Unlock()
	// resolve the remaining callbacks to free the resources.
	atomic.AddUint64(&d.stats.requestsAbandoned, uint64(len(callbacks)))
	for _, cb := range callbacks {
		cb <- lookupResult{err: err}
		close(cb)
	}
}

func (d *reverseDNSCache) cleanup() {
	now := time.Now()
	d.expireListLock.Lock()
	if len(d.expireList) == 0 {
		d.expireListLock.Unlock()
		return
	}
	ipsToDelete := make([]string, 0, len(d.expireList))
	for i := 0; i < len(d.expireList); i++ {
		if !d.expireList[i].expiresAt.Before(now) {
			break // done. Nothing after this point is expired.
		}
		ipsToDelete = append(ipsToDelete, d.expireList[i].ip)
	}
	if len(ipsToDelete) == 0 {
		d.expireListLock.Unlock()
		return
	}
	d.expireList = d.expireList[len(ipsToDelete):]
	d.expireListLock.Unlock()

	atomic.AddUint64(&d.stats.cacheExpire, uint64(len(ipsToDelete)))

	d.rwLock.Lock()
	defer d.rwLock.Unlock()
	for _, ip := range ipsToDelete {
		delete(d.cache, ip)
	}
}

func (d *reverseDNSCache) getStats() rDNSCacheStats {
	stats := rDNSCacheStats{}
	stats.cacheHit = atomic.LoadUint64(&d.stats.cacheHit)
	stats.cacheMiss = atomic.LoadUint64(&d.stats.cacheMiss)
	stats.cacheExpire = atomic.LoadUint64(&d.stats.cacheExpire)
	stats.requestsAbandoned = atomic.LoadUint64(&d.stats.requestsAbandoned)
	stats.requestsFilled = atomic.LoadUint64(&d.stats.requestsFilled)
	return stats
}

func (d *reverseDNSCache) stop() {
	d.cancelCleanupWorker()
}
