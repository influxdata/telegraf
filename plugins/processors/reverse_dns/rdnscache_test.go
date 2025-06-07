package reverse_dns

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSimpleReverseDNSLookup(t *testing.T) {
	d := newReverseDNSCache(60*time.Second, 1*time.Second, -1)
	defer d.stop()

	d.resolver = &localResolver{}
	answer, err := d.lookup("127.0.0.1")
	require.NoError(t, err)
	require.Equal(t, []string{"localhost"}, answer)
	err = blockAllWorkers(t.Context(), d)
	require.NoError(t, err)

	// do another request with no workers available.
	// it should read from cache instantly.
	answer, err = d.lookup("127.0.0.1")
	require.NoError(t, err)
	require.Equal(t, []string{"localhost"}, answer)

	require.Len(t, d.cache, 1)
	require.Len(t, d.expireList, 1)
	d.cleanup()
	require.Len(t, d.expireList, 1) // ttl hasn't hit yet.

	stats := d.getStats()

	require.EqualValues(t, 0, stats.cacheExpire)
	require.EqualValues(t, 1, stats.cacheMiss)
	require.EqualValues(t, 1, stats.cacheHit)
	require.EqualValues(t, 1, stats.requestsFilled)
	require.EqualValues(t, 0, stats.requestsAbandoned)
}

func TestParallelReverseDNSLookup(t *testing.T) {
	d := newReverseDNSCache(1*time.Second, 1*time.Second, -1)
	defer d.stop()

	d.resolver = &localResolver{}
	var answer1, answer2 []string
	var err1, err2 error
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		answer1, err1 = d.lookup("127.0.0.1")
		wg.Done()
	}()
	go func() {
		answer2, err2 = d.lookup("127.0.0.1")
		wg.Done()
	}()

	wg.Wait()

	require.NoError(t, err1)
	require.NoError(t, err2)

	t.Log(answer1)
	t.Log(answer2)

	require.Equal(t, []string{"localhost"}, answer1)
	require.Equal(t, []string{"localhost"}, answer2)

	require.Len(t, d.cache, 1)

	stats := d.getStats()

	require.EqualValues(t, 1, stats.cacheMiss)
	require.EqualValues(t, 1, stats.cacheHit)
}

func TestUnavailableDNSServerRespectsTimeout(t *testing.T) {
	d := newReverseDNSCache(0, 1, -1)
	defer d.stop()

	d.resolver = &timeoutResolver{}

	result, err := d.lookup("192.153.33.3")
	require.Error(t, err)
	require.Equal(t, errTimeout, err)

	require.Nil(t, result)
}

func TestCleanupHappens(t *testing.T) {
	ttl := 100 * time.Millisecond
	d := newReverseDNSCache(ttl, 1*time.Second, -1)
	defer d.stop()

	d.resolver = &localResolver{}
	_, err := d.lookup("127.0.0.1")
	require.NoError(t, err)

	require.Len(t, d.cache, 1)

	time.Sleep(ttl) // wait for cache entry to expire.
	d.cleanup()
	require.Empty(t, d.expireList)

	stats := d.getStats()

	require.EqualValues(t, 1, stats.cacheExpire)
	require.EqualValues(t, 1, stats.cacheMiss)
	require.EqualValues(t, 0, stats.cacheHit)
}

func TestLookupTimeout(t *testing.T) {
	d := newReverseDNSCache(10*time.Second, 10*time.Second, -1)
	defer d.stop()

	d.resolver = &timeoutResolver{}
	_, err := d.lookup("127.0.0.1")
	require.Error(t, err)
	require.EqualValues(t, 1, d.getStats().requestsAbandoned)
}

type timeoutResolver struct{}

func (*timeoutResolver) LookupAddr(context.Context, string) (names []string, err error) {
	return nil, errors.New("timeout")
}

type localResolver struct{}

func (*localResolver) LookupAddr(context.Context, string) (names []string, err error) {
	return []string{"localhost"}, nil
}

// blockAllWorkers is a test function that eats up all the worker pool space to
// make sure workers are done running and there's no room to acquire a new worker.
func blockAllWorkers(testContext context.Context, d *reverseDNSCache) error {
	return d.sem.Acquire(testContext, int64(d.maxWorkers))
}
