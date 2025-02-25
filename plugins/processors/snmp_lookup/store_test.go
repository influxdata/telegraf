package snmp_lookup

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
)

func TestAddBacklog(t *testing.T) {
	var notifyCount atomic.Uint64
	s := newStore(0, 0, 0, 0)
	s.update = func(string) *tagMap { return nil }
	s.notify = func(string, *tagMap) { notifyCount.Add(1) }
	defer s.destroy()

	s.Lock()
	require.Empty(t, s.deferredUpdates)
	s.Unlock()

	s.addBacklog("127.0.0.1", time.Now().Add(1*time.Second))

	s.Lock()
	require.Contains(t, s.deferredUpdates, "127.0.0.1")
	s.Unlock()
	require.Eventually(t, func() bool {
		return notifyCount.Load() == 1
	}, 3*time.Second, 100*time.Millisecond)
	s.Lock()
	require.Empty(t, s.deferredUpdates)
	s.Unlock()
}

func TestLookup(t *testing.T) {
	tmr := tagMapRows{
		"0": {"ifName": "eth0"},
		"1": {"ifName": "eth1"},
	}
	minUpdateInterval := 50 * time.Millisecond
	cacheTTL := config.Duration(2 * minUpdateInterval)
	var notifyCount atomic.Uint64
	s := newStore(defaultCacheSize, cacheTTL, defaultParallelLookups, config.Duration(minUpdateInterval))
	s.update = func(string) *tagMap {
		return &tagMap{
			created: time.Now(),
			rows:    tmr,
		}
	}
	s.notify = func(string, *tagMap) { notifyCount.Add(1) }
	defer s.destroy()

	require.Equal(t, 0, s.cache.Len())

	// Initial lookup should cache entries
	s.lookup("127.0.0.1", "999")
	require.Eventually(t, func() bool {
		return s.cache.Contains("127.0.0.1")
	}, time.Second, time.Millisecond)
	require.EqualValues(t, 1, notifyCount.Load())

	entries, _ := s.cache.Get("127.0.0.1")
	require.Equal(t, tmr, entries.rows)

	// Second lookup should be deferred minUpdateInterval
	s.Lock()
	require.Empty(t, s.deferredUpdates)
	s.Unlock()

	s.lookup("127.0.0.1", "999")

	require.EqualValues(t, 2, notifyCount.Load())

	s.Lock()
	require.Contains(t, s.deferredUpdates, "127.0.0.1")
	require.WithinDuration(t, time.Now(), s.deferredUpdates["127.0.0.1"], minUpdateInterval)
	s.Unlock()

	// Wait until resolved
	require.Eventually(t, func() bool {
		return notifyCount.Load() == 3
	}, time.Second, time.Millisecond)

	s.Lock()
	require.Empty(t, s.deferredUpdates)
	s.Unlock()

	time.Sleep(minUpdateInterval)

	// Third lookup should directly update
	s.lookup("127.0.0.1", "999")
	_, inflight := s.inflight.Load("127.0.0.1")
	require.True(t, inflight)
	require.Eventually(t, func() bool {
		return notifyCount.Load() == 4
	}, time.Second, time.Millisecond)
}
