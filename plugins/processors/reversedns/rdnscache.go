package reversedns

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"
)

type IPType uint8

var defaultMaxWorkers = 10

// ReverseDNSCache is safe to use across multiple goroutines.
// if multiple goroutines request the same IP at the same time, one of the
// requests will trigger the lookup and the rest will wait for its response.
type ReverseDNSCache struct {
	CacheHit    uint64
	CacheMiss   uint64
	CacheExpire uint64

	// settings
	ttl           time.Duration
	lookupTimeout time.Duration
	maxWorkers    int

	// internal
	rwLock sync.RWMutex
	sem    *semaphore.Weighted

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

func NewReverseDNSCache(ttl, lookupTimeout time.Duration, workerPoolSize int) *ReverseDNSCache {
	if workerPoolSize <= 0 {
		workerPoolSize = defaultMaxWorkers
	}
	d := &ReverseDNSCache{
		ttl:           ttl,
		lookupTimeout: lookupTimeout,
		cache:         map[string]*dnslookup{},
		expireList:    []*dnslookup{},
		maxWorkers:    workerPoolSize,
		sem:           semaphore.NewWeighted(int64(workerPoolSize)),
	}
	d.startCleanupWorker()
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

type callbackChannelType chan []string

// Lookup takes a string representing a parseable ipv4 or ipv6 IP, and blocks
// until it has resolved to 0-n results, or until its lookup timeout has elapsed.
// if the lookup timeout elapses, it returns an empty slice.
func (d *ReverseDNSCache) Lookup(ip string) []string {
	if len(ip) == 0 {
		return []string{}
	}
	// todo: fixed response for known localhost, etc.
	return d.lookup(ip)
}

func (d *ReverseDNSCache) lookup(ip string) []string {
	// check if the value is cached
	d.rwLock.RLock()
	result, found := d.unsafeGetFromCache(ip)
	if found && result.completed && result.expiresAt.After(time.Now()) {
		debug("found in cache", result)
		defer d.rwLock.RUnlock()
		atomic.AddUint64(&d.CacheHit, 1)
		// cache is valid
		debug("Found in cache: ", result.domains)
		return result.domains
	}
	d.rwLock.RUnlock()

	// if it's not cached, kick off a lookup job and subscribe to the result.
	debug("subscribe!")
	lookupChan := d.subscribeTo(ip)
	timer := time.NewTimer(d.lookupTimeout)
	defer timer.Stop()

	// timer is still necessary even if doLookup respects timeout due to worker
	// pool starvation.
	select {
	case domains := <-lookupChan:
		return domains
	case <-timer.C:
		fmt.Fprintf(os.Stderr, "reverse dns lookup timed out\n")
		return []string{}
	}
}

func (d *ReverseDNSCache) subscribeTo(ip string) callbackChannelType {
	callback := make(callbackChannelType, 1)

	d.rwLock.Lock()
	defer d.rwLock.Unlock()
	debug("d.rwLock.Lock()")
	defer debug("d.rwLock.Unlock()")

	// confirm it's still not in the cache. This needs to be done under an active lock.
	result, found := d.unsafeGetFromCache(ip)
	if found {
		debug("found")
		atomic.AddUint64(&d.CacheHit, 1)
		// has the request been answered since we last checked?
		if result.completed {
			debug("found completed")
			// we can return the answer with the channel.
			callback <- result.domains
			return callback
		}
		debug("found not completed")
		// there's a request but it hasn't been answered yet;
		// add yourself to the subscribers and return that.
		result.callbacks = append(result.callbacks, callback)
		d.unsafeSaveToCache(result)
		return callback
	}

	atomic.AddUint64(&d.CacheMiss, 1)

	debug("not found")

	// otherwise we need to register the request
	l := &dnslookup{
		ip:        ip,
		expiresAt: time.Now().Add(d.ttl),
		callbacks: []callbackChannelType{callback},
	}

	d.unsafeSaveToCache(l)
	go d.doLookup(*l)
	return callback
}

// unsafeGetFromCache fetches from the correct internal ip cache.
// you MUST first do a read or write lock before calling it, and keep locks around
// the dnslookup that is returned until you clone it.
func (d *ReverseDNSCache) unsafeGetFromCache(ip string) (lookup *dnslookup, found bool) {
	lookup, found = d.cache[ip]
	if found && lookup.expiresAt.Before(time.Now()) {
		return nil, false
	}
	return lookup, found
}

// unsafeSaveToCache stores a lookup in the correct internal ip cache.
// you MUST first do a read or write lock before calling it.
func (d *ReverseDNSCache) unsafeSaveToCache(lookup *dnslookup) {
	if lookup.expiresAt.Before(time.Now()) {
		return // don't cache.
	}
	d.cache[lookup.ip] = lookup
}

func (d *ReverseDNSCache) startCleanupWorker() {
	go func() {
		cleanupTick := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-cleanupTick.C:
				d.cleanup()
			}
		}
	}()
}

func (d *ReverseDNSCache) doLookup(l dnslookup) {
	d.sem.Acquire(context.TODO(), 1)
	defer d.sem.Release(1)

	ctx, cancel := context.WithTimeout(context.Background(), d.lookupTimeout)
	defer cancel()
	names, err := net.DefaultResolver.LookupAddr(ctx, l.ip)
	if err != nil {
		fmt.Printf("RDNS error: %s\n", err)
		return
	}

	d.rwLock.Lock()
	lookup, found := d.unsafeGetFromCache(l.ip)
	if !found {
		d.rwLock.Unlock()
		return
	}

	lookup.domains = names
	lookup.completed = true
	lookup.expiresAt = time.Now().Add(d.ttl) // extend the ttl now that we have a reply.
	callbacks := lookup.callbacks
	lookup.callbacks = nil

	d.unsafeSaveToCache(lookup)
	d.rwLock.Unlock()

	d.expireListLock.Lock()
	// add it to the expireList.
	// fmt.Println("added to expire list!")
	d.expireList = append(d.expireList, lookup)
	d.expireListLock.Unlock()

	for _, cb := range callbacks {
		cb <- names
	}
}

func (d *ReverseDNSCache) cleanup() {
	now := time.Now()
	d.expireListLock.Lock()
	if len(d.expireList) == 0 {
		d.expireListLock.Unlock()
		return
	}
	ipsToDelete := map[string]bool{}
	for i := 0; i < len(d.expireList); i++ {
		if d.expireList[i].expiresAt.After(now) {
			break // done. Nothing after this point is expired.
		}
		ipsToDelete[d.expireList[i].ip] = true
	}
	if len(ipsToDelete) == 0 { // maybe change to 1000
		d.expireListLock.Unlock()
		return
	}
	d.expireList = d.expireList[len(ipsToDelete):]
	d.expireListLock.Unlock()

	atomic.AddUint64(&d.CacheExpire, uint64(len(ipsToDelete)))

	d.rwLock.Lock()
	defer d.rwLock.Unlock()
	newMap := map[string]*dnslookup{}
	for k, v := range d.cache {
		if !ipsToDelete[k] {
			newMap[k] = v
		}
	}
	d.cache = newMap
}

// waitForWorkers is a test function that eats up all the worker pool space to
// make sure workers are done running.
func (d *ReverseDNSCache) waitForWorkers() {
	d.sem.Acquire(context.TODO(), int64(d.maxWorkers))
}

func debug(s ...interface{}) {
	// fmt.Println(s...)
}
