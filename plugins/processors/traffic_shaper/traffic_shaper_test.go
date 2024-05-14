package traffic_shaper

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
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
		Samples:    1,
		TimeUnit:   time.Millisecond * 100,
		BufferSize: 5,
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
		Samples:    1,
		TimeUnit:   time.Millisecond * 100,
		BufferSize: 4,
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
		Samples:    1,
		TimeUnit:   time.Millisecond * 100,
		BufferSize: 5,
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
