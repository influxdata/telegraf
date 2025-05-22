package cumulative_sum

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestApply(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		fields   []string
		input    []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "all fields keep original fields",
			input: []telegraf.Metric{
				metric.New(
					"foo",
					map[string]string{"tag": "some tag"},
					map[string]interface{}{
						"healthy":       false,
						"value":         float64(1.1),
						"error_counter": int64(10),
						"error":         "machine broken",
					},
					now,
				),
				metric.New(
					"bar",
					map[string]string{"tag": "another tag"},
					map[string]interface{}{
						"healthy": true,
						"value":   float64(4.4),
					},
					now,
				),
			},
			expected: []telegraf.Metric{
				metric.New(
					"foo",
					map[string]string{"tag": "some tag"},
					map[string]interface{}{
						"healthy":           false,
						"healthy_sum":       float64(0),
						"value":             float64(1.1),
						"value_sum":         float64(1.1),
						"error_counter":     int64(10),
						"error_counter_sum": float64(10),
						"error":             "machine broken",
					},
					now,
				),
				metric.New(
					"bar",
					map[string]string{"tag": "another tag"},
					map[string]interface{}{
						"healthy":     true,
						"healthy_sum": float64(1),
						"value":       float64(4.4),
						"value_sum":   float64(4.4),
					},
					now,
				),
			},
		},
		{
			name:   "filter value remove original",
			fields: []string{"value"},
			input: []telegraf.Metric{
				metric.New(
					"foo",
					map[string]string{"tag": "some tag"},
					map[string]interface{}{
						"healthy":       false,
						"value":         float64(1.1),
						"error_counter": int64(10),
						"error":         "machine broken",
					},
					now,
				),
				metric.New(
					"bar",
					map[string]string{"tag": "another tag"},
					map[string]interface{}{
						"healthy": true,
						"value":   float64(4.4),
					},
					now,
				),
			},
			expected: []telegraf.Metric{
				metric.New(
					"foo",
					map[string]string{"tag": "some tag"},
					map[string]interface{}{
						"healthy":       false,
						"value":         float64(1.1),
						"value_sum":     float64(1.1),
						"error_counter": int64(10),
						"error":         "machine broken",
					},
					now,
				),
				metric.New(
					"bar",
					map[string]string{"tag": "another tag"},
					map[string]interface{}{
						"healthy":   true,
						"value":     float64(4.4),
						"value_sum": float64(4.4),
					},
					now,
				),
			},
		},
		{
			name:   "multiple metrics",
			fields: []string{"value"},
			input: []telegraf.Metric{
				metric.New(
					"foo",
					map[string]string{"tag": "some tag"},
					map[string]interface{}{"value": float64(1.1)},
					now,
				),
				metric.New(
					"foo",
					map[string]string{"tag": "another tag"},
					map[string]interface{}{"value": float64(4.4)},
					now,
				),
				metric.New(
					"foo",
					map[string]string{"tag": "some tag"},
					map[string]interface{}{"value": float64(1.1)},
					now.Add(time.Second),
				),
				metric.New(
					"foo",
					map[string]string{"tag": "some tag"},
					map[string]interface{}{"value": float64(0.8)},
					now.Add(2*time.Second),
				),
			},
			expected: []telegraf.Metric{
				metric.New(
					"foo",
					map[string]string{"tag": "some tag"},
					map[string]interface{}{
						"value":     float64(1.1),
						"value_sum": float64(1.1),
					},
					now,
				),
				metric.New(
					"foo",
					map[string]string{"tag": "another tag"},
					map[string]interface{}{
						"value":     float64(4.4),
						"value_sum": float64(4.4),
					},
					now,
				),
				metric.New(
					"foo",
					map[string]string{"tag": "some tag"},
					map[string]interface{}{
						"value":     float64(1.1),
						"value_sum": float64(2.2),
					},
					now.Add(time.Second),
				),
				metric.New(
					"foo",
					map[string]string{"tag": "some tag"},
					map[string]interface{}{
						"value":     float64(0.8),
						"value_sum": float64(3.0),
					},
					now.Add(2*time.Second),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin
			plugin := &CumulativeSum{
				Fields: tt.fields,
				Log:    &testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			// Check the results
			actual := plugin.Apply(tt.input...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

func TestCacheExpiry(t *testing.T) {
	now := time.Now()
	// Define the input metrics for the first and second apply call
	input := []telegraf.Metric{
		metric.New(
			"foo",
			map[string]string{"tag": "some tag"},
			map[string]interface{}{"value": float64(1.1)},
			now,
		),
		metric.New(
			"foo",
			map[string]string{"tag": "another tag"},
			map[string]interface{}{"value": float64(4.4)},
			now,
		),
	}
	// Setup the plugin
	plugin := &CumulativeSum{
		ExpiryInterval: config.Duration(10 * time.Second),
	}
	require.NoError(t, plugin.Init())

	// Populate the cache with a value for all input metrics
	for _, m := range input {
		id := m.HashID()
		plugin.cache[id] = &entry{
			sums: map[string]float64{"value": 1},
			seen: now.Add(-time.Second),
		}
	}

	// Apply the processor for the first time. We expect the cache to be valid and not expired
	actual := plugin.Apply(input...)
	expected1 := []telegraf.Metric{
		metric.New(
			"foo",
			map[string]string{"tag": "some tag"},
			map[string]interface{}{
				"value":     float64(1.1),
				"value_sum": float64(2.1), // init 1 + 1.1 from metric
			},
			now,
		),
		metric.New(
			"foo",
			map[string]string{"tag": "another tag"},
			map[string]interface{}{
				"value":     float64(4.4),
				"value_sum": float64(5.4), // init 1 + 4.4 from metric
			},
			now,
		),
	}
	testutil.RequireMetricsEqual(t, expected1, actual)
	require.Len(t, plugin.cache, 2, "wrong number of cache entries")

	// Artificially age the cache for the second input metric simulating that this metric
	// was not seen for a longer period expiry interval
	id1 := input[1].HashID()
	plugin.cache[id1].seen = now.Add(-11 * time.Second)

	// Apply the processor a second time. This time we expect the second metric to be removed
	// from the cache.
	actual = plugin.Apply(input[0])
	expected2 := []telegraf.Metric{
		metric.New(
			"foo",
			map[string]string{"tag": "some tag"},
			map[string]interface{}{
				"value":     float64(1.1),
				"value_sum": float64(3.2), // init 1 + 1.1 + 1.1
			},
			now,
		),
	}
	testutil.RequireMetricsEqual(t, expected2, actual)
	require.Len(t, plugin.cache, 1, "wrong number of cache entries")
	require.NotContains(t, plugin.cache, id1)

	// Finally apply the processor a third time including the second input metric which should
	// now show a sum without expired value.
	actual = plugin.Apply(input...)
	expected3 := []telegraf.Metric{
		metric.New(
			"foo",
			map[string]string{"tag": "some tag"},
			map[string]interface{}{
				"value":     float64(1.1),
				"value_sum": float64(4.3),
			},
			now,
		),
		metric.New(
			"foo",
			map[string]string{"tag": "another tag"},
			map[string]interface{}{
				"value":     float64(4.4),
				"value_sum": float64(4.4),
			},
			now,
		),
	}
	testutil.RequireMetricsEqual(t, expected3, actual, cmpopts.EquateApprox(0.0, 1e-9))
	require.Len(t, plugin.cache, 2, "wrong number of cache entries")
}
