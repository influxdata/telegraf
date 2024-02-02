package snmp_lookup

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/stretchr/testify/require"
)

func TestAddBacklog(t *testing.T) {
	var notifyCount atomic.Uint64
	s := newStore(0, 0, 0, 0)
	s.update = func(agent string) *tagMap { return nil }
	s.notify = func(agent string, tm *tagMap) { notifyCount.Add(1) }
	defer s.destroy()

	require.Empty(t, s.deferredUpdates)

	s.addBacklog("127.0.0.1", time.Now().Add(10*time.Millisecond))
	require.Contains(t, s.deferredUpdates, "127.0.0.1")
	require.Eventually(t, func() bool {
		return notifyCount.Load() == 1
	}, time.Second, time.Millisecond)
	require.Empty(t, s.deferredUpdates)
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
	s.update = func(agent string) *tagMap {
		return &tagMap{
			created: time.Now(),
			rows:    tmr,
		}
	}
	s.notify = func(agent string, tm *tagMap) { notifyCount.Add(1) }
	defer s.destroy()

	require.Equal(t, 0, s.cache.Len())

	// Initial lookup should cache entries
	s.lookup("127.0.0.1", "0")
	require.Eventually(t, func() bool {
		return s.cache.Contains("127.0.0.1")
	}, time.Second, time.Millisecond)
	require.EqualValues(t, 1, notifyCount.Load())

	entries, _ := s.cache.Get("127.0.0.1")
	require.Equal(t, tmr, entries.rows)

	// Second lookup should be deferred minUpdateInterval
	require.Empty(t, s.deferredUpdates)
	s.lookup("127.0.0.1", "999")
	require.EqualValues(t, 2, notifyCount.Load())
	require.Contains(t, s.deferredUpdates, "127.0.0.1")
	require.WithinDuration(t, time.Now(), s.deferredUpdates["127.0.0.1"], minUpdateInterval)

	// Wait until resolved
	require.Eventually(t, func() bool {
		return notifyCount.Load() == 3
	}, time.Second, time.Millisecond)
	require.Empty(t, s.deferredUpdates)
	time.Sleep(minUpdateInterval)

	// Third lookup should directly update
	s.lookup("127.0.0.1", "999")
	require.EqualValues(t, 4, notifyCount.Load())
	// TODO: How to test if s.enqueue was called without s.addBacklog?
	require.Eventually(t, func() bool {
		return notifyCount.Load() == 5
	}, time.Second, time.Millisecond)
}
