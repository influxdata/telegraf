package execd

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestExternalProcessorWorks(t *testing.T) {

	e := New()
	e.Command = []string{shell(), fileShellScriptPath()}
	e.RestartDelay = config.Duration(5 * time.Second)

	acc := &testutil.Accumulator{}

	require.NoError(t, e.Start(acc))

	now := time.Now()
	metrics := []telegraf.Metric{}
	for i := 0; i < 10; i++ {
		m, err := metric.New("test",
			map[string]string{
				"city": "Toronto",
			},
			map[string]interface{}{
				"population": 6000000,
				"count":      1,
			},
			now)
		require.NoError(t, err)
		metrics = append(metrics, m)
		now = now.Add(1)

		e.Add(m)
	}

	acc.Wait(1)
	m := acc.Metrics[0]

	require.NoError(t, e.Stop())

	require.Equal(t, "test", m.Measurement)

	city, ok := m.Tags["city"]
	require.True(t, ok)
	require.EqualValues(t, "Toronto", city)

	val, ok := m.Fields["population"]
	require.True(t, ok)
	require.EqualValues(t, 6000000, val)

	val, ok = m.Fields["count"]
	require.True(t, ok)
	require.EqualValues(t, 2, val)

	metricTime := m.Time.UnixNano()

	// read the other 9 and make sure they're ordered properly
	acc.Wait(9)
	for i := 0; i < 9; i++ {
		m = acc.Metrics[i+1]
		require.EqualValues(t, metricTime+1, m.Time.UnixNano())
		metricTime = m.Time.UnixNano()
	}
}

func readChanWithTimeout(t *testing.T, metrics chan telegraf.Metric, timeout time.Duration) telegraf.Metric {
	to := time.NewTimer(timeout)
	defer to.Stop()
	select {
	case m := <-metrics:
		return m
	case <-to.C:
		require.FailNow(t, "timeout waiting for metric")
	}
	return nil
}

func fileShellScriptPath() string {
	return "./examples/multiplier_line_protocol/multiplier_line_protocol.rb"
}

func shell() string {
	return "ruby"
}
