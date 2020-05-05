// +build !windows

package execd

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/parsers"

	"github.com/influxdata/telegraf"
)

func TestExternalInputWorks(t *testing.T) {
	jsonParser, err := parsers.NewInfluxParser()
	require.NoError(t, err)

	e := &Execd{
		Command:      []string{shell(), fileShellScriptPath()},
		RestartDelay: config.Duration(5 * time.Second),
		parser:       jsonParser,
		Signal:       "STDIN",
	}

	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	acc := agent.NewAccumulator(&TestMetricMaker{}, metrics)

	require.NoError(t, e.Start(acc))
	require.NoError(t, e.Gather(acc))

	// grab a metric and make sure it's a thing
	m := readChanWithTimeout(t, metrics, 10*time.Second)

	e.Stop()

	require.Equal(t, "counter_bash", m.Name())
	val, ok := m.GetField("count")
	require.True(t, ok)
	require.Equal(t, float64(0), val)
	// test that a later gather will not panic
	e.Gather(acc)
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
	return "./examples/count.sh"
}

func shell() string {
	return "sh"
}

type TestMetricMaker struct{}

func (tm *TestMetricMaker) Name() string {
	return "TestPlugin"
}

func (tm *TestMetricMaker) LogName() string {
	return tm.Name()
}

func (tm *TestMetricMaker) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	return metric
}

func (tm *TestMetricMaker) Log() telegraf.Logger {
	return models.NewLogger("TestPlugin", "test", "")
}
