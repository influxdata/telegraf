package defaults

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestDefaults(t *testing.T) {
	scenarios := []struct {
		name     string
		defaults *Defaults
		input    telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "Test that no values are changed since they are not nil or empty",
			defaults: &Defaults{
				DefaultFieldsSets: map[string]interface{}{
					"usage":     30,
					"wind_feel": "very chill",
					"is_dead":   true,
				},
			},
			input: testutil.MustMetric(
				"CPU metrics",
				map[string]string{},
				map[string]interface{}{
					"usage":     45,
					"wind_feel": "a dragon's breath",
					"is_dead":   false,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"CPU metrics",
					map[string]string{},
					map[string]interface{}{
						"usage":     45,
						"wind_feel": "a dragon's breath",
						"is_dead":   false,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "Tests that the missing fields are set on the metric",
			defaults: &Defaults{
				DefaultFieldsSets: map[string]interface{}{
					"max_clock_gz":  6,
					"wind_feel":     "Unknown",
					"boost_enabled": false,
					"variance":      1.2,
				},
			},
			input: testutil.MustMetric(
				"CPU metrics",
				map[string]string{},
				map[string]interface{}{
					"usage":       45,
					"temperature": 64,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"CPU metrics",
					map[string]string{},
					map[string]interface{}{
						"usage":         45,
						"temperature":   64,
						"max_clock_gz":  6,
						"wind_feel":     "Unknown",
						"boost_enabled": false,
						"variance":      1.2,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "Tests that set but empty fields are replaced by specified defaults",
			defaults: &Defaults{
				DefaultFieldsSets: map[string]interface{}{
					"max_clock_gz":  6,
					"wind_feel":     "Unknown",
					"fan_loudness":  "Inaudible",
					"boost_enabled": false,
				},
			},
			input: testutil.MustMetric(
				"CPU metrics",
				map[string]string{},
				map[string]interface{}{
					"max_clock_gz": "",
					"wind_feel":    " ",
					"fan_loudness": "         ",
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"CPU metrics",
					map[string]string{},
					map[string]interface{}{
						"max_clock_gz":  6,
						"wind_feel":     "Unknown",
						"fan_loudness":  "Inaudible",
						"boost_enabled": false,
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			defaults := scenario.defaults

			resultMetrics := defaults.Apply(scenario.input)
			require.Len(t, resultMetrics, 1)
			testutil.RequireMetricsEqual(t, scenario.expected, resultMetrics)
		})
	}
}

func TestTracking(t *testing.T) {
	inputRaw := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42, "topic": "telegraf"}, time.Unix(0, 0)),
		metric.New("bar", map[string]string{}, map[string]interface{}{"hours": 23}, time.Unix(0, 0)),
		metric.New("baz", map[string]string{}, map[string]interface{}{"status": "fixed"}, time.Unix(0, 0)),
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
			map[string]string{},
			map[string]interface{}{"value": 42, "status": "unknown", "topic": "telegraf"},
			time.Unix(0, 0),
		),
		metric.New(
			"bar",
			map[string]string{},
			map[string]interface{}{"value": 6, "status": "unknown", "hours": 23},
			time.Unix(0, 0),
		),
		metric.New(
			"baz",
			map[string]string{},
			map[string]interface{}{"value": 6, "status": "fixed"},
			time.Unix(0, 0),
		),
	}

	plugin := &Defaults{
		DefaultFieldsSets: map[string]interface{}{
			"value":  6,
			"status": "unknown",
		},
	}

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
