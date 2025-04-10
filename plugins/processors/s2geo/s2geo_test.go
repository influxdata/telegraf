package s2geo

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestGeo(t *testing.T) {
	plugin := &Geo{
		LatField:  "lat",
		LonField:  "lon",
		TagKey:    "s2_cell_id",
		CellLevel: 11,
	}

	pluginMostlyDefault := &Geo{
		CellLevel: 11,
	}

	err := plugin.Init()
	require.NoError(t, err)

	m := testutil.MustMetric(
		"mta",
		map[string]string{},
		map[string]interface{}{
			"lat": 40.878738,
			"lon": -72.517572,
		},
		time.Unix(1578603600, 0),
	)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"mta",
			map[string]string{
				"s2_cell_id": "89e8ed4",
			},
			map[string]interface{}{
				"lat": 40.878738,
				"lon": -72.517572,
			},
			time.Unix(1578603600, 0),
		),
	}

	actual := plugin.Apply(m)
	testutil.RequireMetricsEqual(t, expected, actual)
	actual = pluginMostlyDefault.Apply(m)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestTracking(t *testing.T) {
	inputRaw := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"lat": 40.878738, "lon": 72.517572}, time.Unix(0, 0)),
		metric.New("bar", map[string]string{}, map[string]interface{}{"lat": 42.842451, "lon": 74.211361}, time.Unix(0, 0)),
		metric.New("baz", map[string]string{}, map[string]interface{}{"lat": 32.300963, "lon": 14.123442}, time.Unix(0, 0)),
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
			map[string]string{"s2_cell_id": "3"},
			map[string]interface{}{"lat": 40.878738, "lon": 72.517572},
			time.Unix(0, 0),
		),
		metric.New(
			"bar",
			map[string]string{"s2_cell_id": "3"},
			map[string]interface{}{"lat": 42.842451, "lon": 74.211361},
			time.Unix(0, 0),
		),
		metric.New(
			"baz",
			map[string]string{"s2_cell_id": "1"},
			map[string]interface{}{"lat": 32.300963, "lon": 14.123442},
			time.Unix(0, 0),
		),
	}

	plugin := &Geo{
		LatField: "lat",
		LonField: "lon",
		TagKey:   "s2_cell_id",
	}
	require.NoError(t, plugin.Init())

	// Process expected metrics and compare with resulting metrics
	actual := plugin.Apply(input...)
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
