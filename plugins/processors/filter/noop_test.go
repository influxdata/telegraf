package filter

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestNoMetric(t *testing.T) {
	processor := Noop{}

	m := []telegraf.Metric{}
	actual := processor.Apply(m...)
	require.Empty(t, actual)
	testutil.RequireMetricsEqual(t, m, actual)
}

func TestSingleMetric(t *testing.T) {
	processor := Noop{}

	m := []telegraf.Metric{
		testutil.MustMetric(
			"test",
			map[string]string{
				"tag": "tag_value",
			},
			map[string]interface{}{
				"value": 42,
			},
			time.Now(),
			telegraf.Gauge,
		),
	}
	actual := processor.Apply(m...)
	require.Len(t, actual, 1)
	testutil.RequireMetricsEqual(t, m, actual)
}

func TestMultipleMetrics(t *testing.T) {
	processor := Noop{}

	m := []telegraf.Metric{
		testutil.MustMetric(
			"test",
			map[string]string{
				"tag": "tag_value",
			},
			map[string]interface{}{
				"value": 42,
			},
			time.Now(),
			telegraf.Gauge,
		),
		testutil.MustMetric(
			"test",
			map[string]string{
				"tag": "tag_value",
			},
			map[string]interface{}{
				"value": 42,
			},
			time.Now(),
			telegraf.Gauge,
		),
	}
	actual := processor.Apply(m...)
	require.Len(t, actual, 2)
	testutil.RequireMetricsEqual(t, m, actual)
}
