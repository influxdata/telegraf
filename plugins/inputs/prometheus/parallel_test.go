package prometheus

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/selfstat"
	"github.com/influxdata/telegraf/testutil"
)

const parallelTestTargets = 8

func parallelTestURLs(base string) []string {
	urls := make([]string, 0, parallelTestTargets)
	for i := range parallelTestTargets {
		urls = append(urls, fmt.Sprintf("%s/metrics%d", base, i))
	}
	return urls
}

func TestMaxParallelScrapesInvalid(t *testing.T) {
	p := &Prometheus{MaxParallelScrapes: -1, Log: testutil.Logger{}}
	require.ErrorContains(t, p.Init(), "max_parallel_scrapes")
}

func TestMaxParallelScrapesLimit(t *testing.T) {
	for _, limit := range []int{1, 3} {
		t.Run(fmt.Sprintf("limit %d", limit), func(t *testing.T) {
			var inFlight, maxInFlight, requests atomic.Int64
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				current := inFlight.Add(1)
				defer inFlight.Add(-1)
				for {
					m := maxInFlight.Load()
					if current <= m || maxInFlight.CompareAndSwap(m, current) {
						break
					}
				}
				requests.Add(1)
				time.Sleep(20 * time.Millisecond)
				if _, err := w.Write([]byte("up 1\n")); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			defer ts.Close()

			p := &Prometheus{
				URLs:               parallelTestURLs(ts.URL),
				MaxParallelScrapes: limit,
				Statistics:         selfstat.NewCollector(nil),
				Log:                testutil.Logger{},
			}
			require.NoError(t, p.Init())

			var acc testutil.Accumulator
			require.NoError(t, p.Gather(&acc))
			require.Empty(t, acc.Errors)
			require.EqualValues(t, parallelTestTargets, requests.Load())
			require.EqualValues(t, parallelTestTargets, acc.NMetrics())
			require.LessOrEqual(t, maxInFlight.Load(), int64(limit))
		})
	}
}

func TestMaxParallelScrapesUnlimited(t *testing.T) {
	// Every request blocks until all targets are being scraped at the same
	// time, so the scrapes only succeed if they run in parallel.
	var arrived sync.WaitGroup
	arrived.Add(parallelTestTargets)
	all := make(chan struct{})
	go func() {
		arrived.Wait()
		close(all)
	}()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		arrived.Done()
		select {
		case <-all:
		case <-time.After(5 * time.Second):
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		if _, err := w.Write([]byte("up 1\n")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	p := &Prometheus{
		URLs:       parallelTestURLs(ts.URL),
		Statistics: selfstat.NewCollector(nil),
		Log:        testutil.Logger{},
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))
	require.Empty(t, acc.Errors)
	require.EqualValues(t, parallelTestTargets, acc.NMetrics())
}

func TestMaxParallelScrapesOrder(t *testing.T) {
	var mu sync.Mutex
	var order []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		order = append(order, r.URL.Path)
		mu.Unlock()
		if _, err := w.Write([]byte("up 1\n")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	urls := parallelTestURLs(ts.URL)
	slices.Reverse(urls)
	p := &Prometheus{
		URLs:               urls,
		MaxParallelScrapes: 1,
		Statistics:         selfstat.NewCollector(nil),
		Log:                testutil.Logger{},
	}
	require.NoError(t, p.Init())

	for range 2 {
		var acc testutil.Accumulator
		require.NoError(t, p.Gather(&acc))
		require.Empty(t, acc.Errors)
	}

	expected := make([]string, 0, parallelTestTargets)
	for i := range parallelTestTargets {
		expected = append(expected, fmt.Sprintf("/metrics%d", i))
	}
	require.Equal(t, slices.Concat(expected, expected), order)
}
