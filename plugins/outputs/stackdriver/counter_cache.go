package stackdriver

import (
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"

	monpb "google.golang.org/genproto/googleapis/monitoring/v3"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

type CounterCache struct {
	sync.RWMutex
	cache map[string]*CounterCacheEntry
	log   telegraf.Logger
}

type CounterCacheEntry struct {
	LastValue *monpb.TypedValue
	StartTime *tspb.Timestamp
}

func (cce *CounterCacheEntry) Reset(ts *tspb.Timestamp) {
	// always backdate a reset by -1ms, otherwise stackdriver's API will hate us
	cce.StartTime = tspb.New(ts.AsTime().Add(time.Millisecond * -1))
}

func (cc *CounterCache) get(key string) (*CounterCacheEntry, bool) {
	cc.RLock()
	value, ok := cc.cache[key]
	cc.RUnlock()
	return value, ok
}

func (cc *CounterCache) set(key string, value *CounterCacheEntry) {
	cc.Lock()
	cc.cache[key] = value
	cc.Unlock()
}

func (cc *CounterCache) GetStartTime(key string, value *monpb.TypedValue, endTime *tspb.Timestamp) *tspb.Timestamp {
	lastObserved, ok := cc.get(key)

	// init: create a new key, backdate the state time to 1ms before the end time
	if !ok {
		newEntry := NewCounterCacheEntry(value, endTime)
		cc.set(key, newEntry)
		return newEntry.StartTime
	}

	// update of existing entry
	if value.GetDoubleValue() < lastObserved.LastValue.GetDoubleValue() || value.GetInt64Value() < lastObserved.LastValue.GetInt64Value() {
		// counter reset
		lastObserved.Reset(endTime)
	} else {
		// counter increment
		//
		// ...but...
		// start times cannot be over 25 hours old; reset after 1 day to be safe
		age := endTime.GetSeconds() - lastObserved.StartTime.GetSeconds()
		cc.log.Debugf("age: %d", age)
		if age > 86400 {
			lastObserved.Reset(endTime)
		}
	}
	// update last observed value
	lastObserved.LastValue = value
	return lastObserved.StartTime
}

func NewCounterCache(log telegraf.Logger) *CounterCache {
	return &CounterCache{
		cache: make(map[string]*CounterCacheEntry),
		log:   log}
}

func NewCounterCacheEntry(value *monpb.TypedValue, ts *tspb.Timestamp) *CounterCacheEntry {
	// Start times must be _before_ the end time, so backdate our original start time
	// to 1ms before the observed time.
	backDatedStart := ts.AsTime().Add(time.Millisecond * -1)
	return &CounterCacheEntry{LastValue: value, StartTime: tspb.New(backDatedStart)}
}

func GetCounterCacheKey(m telegraf.Metric, f *telegraf.Field) string {
	// normalize tag list to form a predictable key
	var tags []string
	for _, t := range m.TagList() {
		tags = append(tags, strings.Join([]string{t.Key, t.Value}, "="))
	}
	sort.Strings(tags)
	return path.Join(m.Name(), strings.Join(tags, "/"), f.Key)
}
