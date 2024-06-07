package reverse_dns

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestSimpleReverseLookupIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	now := time.Now()
	m := metric.New("name", map[string]string{
		"dest_ip": "1.1.1.1",
	}, map[string]interface{}{
		"source_ip": "127.0.0.1",
	}, now)

	dns := newReverseDNS()
	dns.Log = &testutil.Logger{}
	dns.Lookups = []lookupEntry{
		{
			Field: "source_ip",
			Dest:  "source_name",
		},
		{
			Tag:  "dest_ip",
			Dest: "dest_name",
		},
	}
	acc := &testutil.Accumulator{}
	err := dns.Start(acc)
	require.NoError(t, err)
	err = dns.Add(m, acc)
	require.NoError(t, err)
	dns.Stop()
	// should be processed now.

	require.Len(t, acc.GetTelegrafMetrics(), 1)
	processedMetric := acc.GetTelegrafMetrics()[0]
	_, ok := processedMetric.GetField("source_name")
	require.True(t, ok)
	tag, ok := processedMetric.GetTag("dest_name")
	require.True(t, ok)
	require.EqualValues(t, "one.one.one.one.", tag)
}

func TestTracking(t *testing.T) {
	inputRaw := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"ip": "1.1.1.1"}, time.Unix(0, 0)),
		metric.New("bar", map[string]string{}, map[string]interface{}{"ip": "1.1.1.1"}, time.Unix(0, 0)),
		metric.New("baz", map[string]string{}, map[string]interface{}{"ip": "1.1.1.1"}, time.Unix(0, 0)),
	}

	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	input := make([]telegraf.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
	}

	expected := []telegraf.Metric{
		metric.New(
			"foo",
			map[string]string{},
			map[string]interface{}{"ip": "1.1.1.1", "name": "one.one.one.one."},
			time.Unix(0, 0),
		),
		metric.New(
			"bar",
			map[string]string{},
			map[string]interface{}{"ip": "1.1.1.1", "name": "one.one.one.one."},
			time.Unix(0, 0),
		),
		metric.New(
			"baz",
			map[string]string{},
			map[string]interface{}{"ip": "1.1.1.1", "name": "one.one.one.one."},
			time.Unix(0, 0),
		),
	}

	plugin := &ReverseDNS{
		CacheTTL:           config.Duration(24 * time.Hour),
		LookupTimeout:      config.Duration(1 * time.Minute),
		MaxParallelLookups: 10,
		Log:                &testutil.Logger{},
		Lookups: []lookupEntry{
			{
				Field: "ip",
				Dest:  "name",
			},
		},
	}

	// Process expected metrics and compare with resulting metrics
	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	for _, m := range input {
		require.NoError(t, plugin.Add(m, acc))
	}
	plugin.Stop()
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())

	// Simulate output acknowledging delivery
	for _, m := range actual {
		m.Accept()
	}

	// Check delivery
	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(input) == len(delivered)
	}, time.Second, 100*time.Millisecond, "%d delivered but %d expected", len(delivered), len(expected))
}
