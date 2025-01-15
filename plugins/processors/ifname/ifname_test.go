package ifname

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/snmp"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

type item struct {
	entry nameMap
	age   time.Duration
	err   error
}

func TestTableIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping test due to connect failures")

	d := IfName{}
	err := d.Init()
	require.NoError(t, err)
	tab, err := makeTable("1.3.6.1.2.1.2.2.1.2")
	require.NoError(t, err)

	gs, err := snmp.NewWrapper(*snmp.DefaultClientConfig())
	require.NoError(t, err)
	err = gs.SetAgent("127.0.0.1")
	require.NoError(t, err)

	err = gs.Connect()
	require.NoError(t, err)

	// Could use ifIndex but oid index is always the same
	m, err := buildMap(gs, tab)
	require.NoError(t, err)
	require.NotEmpty(t, m)
}

func TestIfNameIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping test due to connect failures")

	d := IfName{
		SourceTag:    "ifIndex",
		DestTag:      "ifName",
		AgentTag:     "agent",
		CacheSize:    1000,
		ClientConfig: *snmp.DefaultClientConfig(),
	}
	err := d.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	err = d.Start(&acc)

	require.NoError(t, err)

	m := testutil.MustMetric(
		"cpu",
		map[string]string{
			"ifIndex": "1",
			"agent":   "127.0.0.1",
		},
		map[string]interface{}{},
		time.Unix(0, 0),
	)

	expected := testutil.MustMetric(
		"cpu",
		map[string]string{
			"ifIndex": "1",
			"agent":   "127.0.0.1",
			"ifName":  "lo",
		},
		map[string]interface{}{},
		time.Unix(0, 0),
	)

	err = d.addTag(m)
	require.NoError(t, err)

	testutil.RequireMetricEqual(t, expected, m)
}

func TestGetMap(t *testing.T) {
	d := IfName{
		CacheSize: 1000,
		CacheTTL:  config.Duration(10 * time.Second),
	}

	require.NoError(t, d.Init())

	expected := nameMap{
		1: "ifname1",
		2: "ifname2",
	}

	var remoteCalls int32

	// Mock the snmp transaction
	d.getMapRemote = func(string) (nameMap, error) {
		atomic.AddInt32(&remoteCalls, 1)
		return expected, nil
	}
	m, age, err := d.getMap("agent")
	require.NoError(t, err)
	require.Zero(t, age) // Age is zero when map comes from getMapRemote
	require.Equal(t, expected, m)

	// Remote call should happen the first time getMap runs
	require.Equal(t, int32(1), remoteCalls)

	const thMax = 3
	ch := make(chan item, thMax)
	var wg sync.WaitGroup
	for th := 0; th < thMax; th++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m, age, err := d.getMap("agent")
			ch <- item{entry: m, age: age, err: err}
		}()
	}

	wg.Wait()
	close(ch)

	for entry := range ch {
		require.NoError(t, entry.err)
		require.NotZero(t, entry.age) // Age is nonzero when map comes from cache
		require.Equal(t, expected, entry.entry)
	}

	// Remote call should not happen subsequent times getMap runs
	require.Equal(t, int32(1), remoteCalls)
}

func TestTracking(t *testing.T) {
	// Setup raw input and expected output
	inputRaw := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"ifIndex": "1", "agent": "127.0.0.1"},
			map[string]interface{}{"value": 42},
			time.Unix(0, 0),
		),
	}

	expected := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{
				"ifIndex": "1",
				"agent":   "127.0.0.1",
				"ifName":  "lo",
			},
			map[string]interface{}{"value": 42},
			time.Unix(0, 0),
		),
	}

	// Create fake notification for testing
	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	// Convert raw input to tracking metric
	input := make([]telegraf.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
	}

	// Prepare and start the plugin
	plugin := &IfName{
		SourceTag:          "ifIndex",
		DestTag:            "ifName",
		AgentTag:           "agent",
		CacheSize:          1000,
		CacheTTL:           config.Duration(10 * time.Second),
		MaxParallelLookups: 100,
	}
	require.NoError(t, plugin.Init())
	plugin.cache.Put("127.0.0.1", nameMap{1: "lo"})

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Process expected metrics and compare with resulting metrics
	for _, in := range input {
		require.NoError(t, plugin.Add(in, &acc))
	}

	require.Eventually(t, func() bool {
		return int(acc.NMetrics()) >= len(expected)
	}, 3*time.Second, 100*time.Microsecond)

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual)

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
