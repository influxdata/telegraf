package date

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestTagAndField(t *testing.T) {
	plugin := &Date{
		TagKey:   "month",
		FieldKey: "month",
	}
	require.Error(t, plugin.Init())
}

func TestNoOutputSpecified(t *testing.T) {
	plugin := &Date{}
	require.Error(t, plugin.Init())
}

func TestMonthTag(t *testing.T) {
	now := time.Now()
	month := now.Format("Jan")

	input := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42}, now),
	}

	expected := []telegraf.Metric{
		metric.New("foo", map[string]string{"month": month}, map[string]interface{}{"value": 42}, now),
		metric.New("bar", map[string]string{"month": month}, map[string]interface{}{"value": 42}, now),
		metric.New("baz", map[string]string{"month": month}, map[string]interface{}{"value": 42}, now),
	}

	plugin := &Date{
		TagKey:     "month",
		DateFormat: "Jan",
	}
	require.NoError(t, plugin.Init())

	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestMonthField(t *testing.T) {
	now := time.Now()
	month := now.Format("Jan")

	input := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42}, now),
	}

	expected := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42, "month": month}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42, "month": month}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42, "month": month}, now),
	}

	plugin := &Date{
		FieldKey:   "month",
		DateFormat: "Jan",
	}
	require.NoError(t, plugin.Init())

	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestOldDateTag(t *testing.T) {
	now := time.Date(1993, 05, 27, 0, 0, 0, 0, time.UTC)

	input := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42}, now),
	}

	expected := []telegraf.Metric{
		metric.New("foo", map[string]string{"year": "1993"}, map[string]interface{}{"value": 42}, now),
		metric.New("bar", map[string]string{"year": "1993"}, map[string]interface{}{"value": 42}, now),
		metric.New("baz", map[string]string{"year": "1993"}, map[string]interface{}{"value": 42}, now),
	}

	plugin := &Date{
		TagKey:     "year",
		DateFormat: "2006",
	}
	require.NoError(t, plugin.Init())

	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestFieldUnix(t *testing.T) {
	now := time.Now()
	ts := now.Unix()

	input := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42}, now),
	}

	expected := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42, "unix": ts}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42, "unix": ts}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42, "unix": ts}, now),
	}

	plugin := &Date{
		FieldKey:   "unix",
		DateFormat: "unix",
	}
	require.NoError(t, plugin.Init())

	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestFieldUnixNano(t *testing.T) {
	now := time.Now()
	ts := now.UnixNano()

	input := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42}, now),
	}

	expected := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42, "unix_ns": ts}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42, "unix_ns": ts}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42, "unix_ns": ts}, now),
	}

	plugin := &Date{
		FieldKey:   "unix_ns",
		DateFormat: "unix_ns",
	}
	require.NoError(t, plugin.Init())

	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestFieldUnixMillis(t *testing.T) {
	now := time.Now()
	ts := now.UnixMilli()

	input := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42}, now),
	}

	expected := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42, "unix_ms": ts}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42, "unix_ms": ts}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42, "unix_ms": ts}, now),
	}

	plugin := &Date{
		FieldKey:   "unix_ms",
		DateFormat: "unix_ms",
	}
	require.NoError(t, plugin.Init())

	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestFieldUnixMicros(t *testing.T) {
	now := time.Now()
	ts := now.UnixMicro()

	input := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42}, now),
	}

	expected := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42, "unix_us": ts}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42, "unix_us": ts}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42, "unix_us": ts}, now),
	}

	plugin := &Date{
		FieldKey:   "unix_us",
		DateFormat: "unix_us",
	}
	require.NoError(t, plugin.Init())

	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestDateOffset(t *testing.T) {
	plugin := &Date{
		TagKey:     "hour",
		DateFormat: "15",
		DateOffset: config.Duration(2 * time.Hour),
	}
	require.NoError(t, plugin.Init())

	input := testutil.MustMetric(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"time_idle": 42.0,
		},
		time.Unix(1578603600, 0),
	)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"hour": "23",
			},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(1578603600, 0),
		),
	}

	actual := plugin.Apply(input)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestTracking(t *testing.T) {
	now := time.Now()
	ts := now.UnixMicro()

	inputRaw := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 42}, now),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 42}, now),
	}

	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	input := make([]telegraf.Metric, 0, len(inputRaw))
	expected := make([]telegraf.Metric, 0, len(input))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)

		em := m.Copy()
		em.AddField("unix_us", ts)
		expected = append(expected, m)
	}

	plugin := &Date{
		FieldKey:   "unix_us",
		DateFormat: "unix_us",
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
