package execd

import (
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/stretchr/testify/require"
)

func TestExternalProcessorWorks(t *testing.T) {
	influxParser, err := parsers.NewInfluxParser()
	require.NoError(t, err)

	influxSerializer, err := serializers.NewInfluxSerializer()
	require.NoError(t, err)

	e := &Execd{
		Command:      []string{shell(), fileShellScriptPath()},
		RestartDelay: config.Duration(5 * time.Second),
		parser:       influxParser,
		serializer:   influxSerializer,
	}

	out := make(chan telegraf.Metric, 10)
	acc := agent.NewMetricStreamAccumulator(out)

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

	m := readChanWithTimeout(t, out, 10*time.Second)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		require.NoError(t, e.Stop())
		wg.Done()
	}()

	require.Equal(t, "test", m.Name())

	city, ok := m.GetTag("city")
	require.True(t, ok)
	require.EqualValues(t, "Toronto", city)

	val, ok := m.GetField("population")
	require.True(t, ok)
	require.EqualValues(t, 6000000, val)

	val, ok = m.GetField("count")
	require.True(t, ok)
	require.EqualValues(t, 2, val)

	metricTime := m.Time().UnixNano()

	// read the other 9 and make sure they're ordered properly
	for i := 0; i < 9; i++ {
		m = readChanWithTimeout(t, out, 10*time.Second)
		require.EqualValues(t, metricTime+1, m.Time().UnixNano())
		metricTime = m.Time().UnixNano()
	}
	wg.Done()
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
