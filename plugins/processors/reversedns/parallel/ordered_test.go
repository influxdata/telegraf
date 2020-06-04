package parallel_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/processors/reversedns/parallel"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestJobsStayOrdered(t *testing.T) {
	acc := &testutil.Accumulator{}

	p := parallel.NewOrdered(acc, jobFunc, 10000, 10)
	now := time.Now()
	for i := 0; i < 20000; i++ {
		m, err := metric.New("test",
			map[string]string{},
			map[string]interface{}{
				"val": i,
			},
			now,
		)
		require.NoError(t, err)
		now = now.Add(1)
		p.Enqueue(m)
	}
	p.Stop()

	i := 0
	require.Len(t, acc.Metrics, 20000, fmt.Sprintf("expected 20k metrics but got %d", len(acc.GetTelegrafMetrics())))
	for _, m := range acc.GetTelegrafMetrics() {
		v, ok := m.GetField("val")
		require.True(t, ok)
		require.EqualValues(t, i, v)
		i++
	}
}

func BenchmarkJobs(b *testing.B) {
	acc := &testutil.Accumulator{}

	p := parallel.NewOrdered(acc, jobFunc, 10000, 10)

	m, _ := metric.New("test",
		map[string]string{},
		map[string]interface{}{
			"val": 1,
		},
		time.Now(),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Enqueue(m)
	}
	p.Stop()
}

func jobFunc(m telegraf.Metric) []telegraf.Metric {
	return []telegraf.Metric{m}
}

type testAccum struct {
	Metrics []telegraf.Metric
}

func (a *testAccum) PassMetric(m telegraf.Metric) {
	a.Metrics = append(a.Metrics, m)
}
