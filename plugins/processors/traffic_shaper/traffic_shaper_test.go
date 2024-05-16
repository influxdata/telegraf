package traffic_shaper

import (
	"math"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

var metrics []telegraf.Metric

func init() {
	now := time.Now()
	metrics = make([]telegraf.Metric, 5)
	for i := range metrics {
		metrics[i] = metric.New("cpu", map[string]string{
			"instanceId": strconv.Itoa(i),
		}, map[string]interface{}{
			"usage": strconv.Itoa(i * 10),
		}, now)
	}
}

// 1 metric is emitted per 100 ms
func TestTrafficShaper_Add(t *testing.T) {
	ts := &TrafficShaper{
		Samples:                1,
		Rate:                   config.Duration(time.Millisecond * 100),
		BufferSize:             5,
		WaitForDrainBeforeExit: true,
	}

	ts.Log = &testutil.Logger{}
	acc := &testutil.Accumulator{}
	err := ts.Start(acc)
	require.NoError(t, err)
	for _, m := range metrics {
		err := ts.Add(m, acc)
		require.NoError(t, err)
	}
	n := 1
	for n < 5 {
		time.Sleep(time.Millisecond * 100)
		//due to concurrency +/- of one extra metric is possible
		acc.Mutex.Lock()
		require.LessOrEqual(t, math.Abs(float64(n-len(acc.Metrics))), 1.0)
		acc.Mutex.Unlock()
		n++
	}
	ts.Stop()
	require.Len(t, acc.Metrics, n)
}

func TestTrafficShaper_Drop(t *testing.T) {
	ts := &TrafficShaper{
		Samples:                1,
		Rate:                   config.Duration(time.Millisecond * 100),
		BufferSize:             4,
		WaitForDrainBeforeExit: true,
	}

	ts.Log = &testutil.Logger{}
	acc := &testutil.Accumulator{}
	err := ts.Start(acc)
	require.NoError(t, err)
	for _, m := range metrics {
		err := ts.Add(m, acc)
		require.NoError(t, err)
	}
	ts.Stop()
	require.Len(t, acc.Metrics, 4)
}

func TestTrafficShaper_Stop(t *testing.T) {
	ts := &TrafficShaper{
		Samples:                1,
		Rate:                   config.Duration(time.Millisecond * 100),
		BufferSize:             5,
		WaitForDrainBeforeExit: true,
	}

	ts.Log = &testutil.Logger{}
	acc := &testutil.Accumulator{}
	err := ts.Start(acc)
	require.NoError(t, err)
	for _, m := range metrics {
		err := ts.Add(m, acc)
		require.NoError(t, err)
	}
	ts.Stop()
	require.Len(t, acc.Metrics, 5)
}

func TestTrafficShaper_StopImmediately(t *testing.T) {
	ts := &TrafficShaper{
		Samples:                1,
		Rate:                   config.Duration(time.Minute),
		BufferSize:             5,
		WaitForDrainBeforeExit: false,
	}

	ts.Log = &testutil.Logger{}
	acc := &testutil.Accumulator{}
	err := ts.Start(acc)
	require.NoError(t, err)
	for _, m := range metrics {
		err := ts.Add(m, acc)
		require.NoError(t, err)
	}
	ts.Stop()
	require.LessOrEqual(t, len(acc.Metrics), 1)
}

func TestTracking(t *testing.T) {
	// Setup raw input and expected output
	inputRaw := metrics

	expected := metrics

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
	plugin := &TrafficShaper{
		Samples:                1,
		Rate:                   config.Duration(time.Millisecond * 100),
		BufferSize:             5,
		WaitForDrainBeforeExit: true,
	}
	plugin.Log = &testutil.Logger{}

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
