package snmp_lookup

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/stretchr/testify/require"
)

func TestAddBacklog(t *testing.T) {
	s := newStore(0, 0, 0, 0)
	s.update = func(agent string) *tagMap { return nil }
	s.notify = func(agent string, tm *tagMap) {}
	defer s.destroy()

	require.Empty(t, s.deferredUpdates)

	s.addBacklog("127.0.0.1", time.Now().Add(10*time.Millisecond))
	require.Contains(t, s.deferredUpdates, "127.0.0.1")
	require.Eventually(t, func() bool {
		return len(s.deferredUpdates) == 0
	}, time.Second, time.Millisecond)
}

func TestLookup(t *testing.T) {
	tmr := tagMapRows{
		"0": {"ifName": "eth0"},
		"1": {"ifName": "eth1"},
	}
	minUpdateInterval := 50 * time.Millisecond
	cacheTTL := config.Duration(2 * minUpdateInterval)
	s := newStore(defaultCacheSize, cacheTTL, defaultParallelLookups, config.Duration(minUpdateInterval))
	s.update = func(agent string) *tagMap {
		return &tagMap{
			created: time.Now(),
			rows:    tmr,
		}
	}
	s.notify = func(agent string, tm *tagMap) {}
	defer s.destroy()

	require.Equal(t, 0, s.cache.Len())

	// Initial lookup should cache entries
	s.lookup("127.0.0.1", "0")
	require.Eventually(t, func() bool {
		return s.cache.Contains("127.0.0.1")
	}, time.Second, time.Millisecond)

	entries, _ := s.cache.Get("127.0.0.1")
	require.Equal(t, tmr, entries.rows)

	// Second lookup should be deferred minUpdateInterval
	require.Empty(t, s.deferredUpdates)
	s.lookup("127.0.0.1", "999")
	require.Contains(t, s.deferredUpdates, "127.0.0.1")
	require.WithinDuration(t, time.Now(), s.deferredUpdates["127.0.0.1"], minUpdateInterval)

	// Wait until resolved
	require.Eventually(t, func() bool {
		return len(s.deferredUpdates) == 0
	}, time.Second, time.Millisecond)
	time.Sleep(minUpdateInterval)

	// Third lookup should directly update
	s.lookup("127.0.0.1", "999")
	// TODO: How to test this?
}
