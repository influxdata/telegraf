package agent

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestMetricStreamPassesMetrics(t *testing.T) {
	now := time.Now()
	in := make(chan telegraf.Metric, 1)
	out := make(chan telegraf.Metric, 1)

	acc := NewMetricStream(in, out)

	require.False(t, acc.IsMetricAvailable())
	require.False(t, acc.IsStreamClosed())

	m, err := metric.New("test", map[string]string{}, map[string]interface{}{"value": 1}, now)
	require.NoError(t, err)
	in <- m

	require.False(t, acc.IsStreamClosed())
	require.True(t, acc.IsMetricAvailable())

	// check again because we know internally the message was read
	require.False(t, acc.IsStreamClosed())

	close(in)
	require.False(t, acc.IsStreamClosed()) // false because it hasn't been read yet.

	metric := acc.GetNextMetric()
	require.NotNil(t, metric)
	testutil.RequireMetricEqual(t, m, metric)

	require.False(t, acc.IsStreamClosed()) // false until we try to read a metric
	require.False(t, acc.IsMetricAvailable())
	require.True(t, acc.IsStreamClosed()) // no more values + stream is closed.

	acc.PassMetric(metric)

	metric = <-out

	require.NotNil(t, metric)
	testutil.RequireMetricEqual(t, m, metric)
}

func TestGetNextMetricBlocksUntilMetricAvailable(t *testing.T) {

}
