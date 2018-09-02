package azure_monitor

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestAggregate(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *AzureMonitor
		metrics  []telegraf.Metric
		addTime  time.Time
		pushTime time.Time
		check    func(t *testing.T, plugin *AzureMonitor, metrics []telegraf.Metric)
	}{
		{
			name: "add metric outside window is dropped",
			plugin: &AzureMonitor{
				Region:     "test",
				ResourceID: "test",
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			addTime:  time.Unix(3600, 0),
			pushTime: time.Unix(3600, 0),
			check: func(t *testing.T, plugin *AzureMonitor, metrics []telegraf.Metric) {
				require.Equal(t, int64(1), plugin.MetricOutsideWindow.Get())
				require.Len(t, metrics, 0)
			},
		},
		{
			name: "metric not sent until period expires",
			plugin: &AzureMonitor{
				Region:     "test",
				ResourceID: "test",
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			addTime:  time.Unix(0, 0),
			pushTime: time.Unix(0, 0),
			check: func(t *testing.T, plugin *AzureMonitor, metrics []telegraf.Metric) {
				require.Len(t, metrics, 0)
			},
		},
		{
			name: "add strings as dimensions",
			plugin: &AzureMonitor{
				Region:              "test",
				ResourceID:          "test",
				StringsAsDimensions: true,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host": "localhost",
					},
					map[string]interface{}{
						"value":   42,
						"message": "howdy",
					},
					time.Unix(0, 0),
				),
			},
			addTime:  time.Unix(0, 0),
			pushTime: time.Unix(3600, 0),
			check: func(t *testing.T, plugin *AzureMonitor, metrics []telegraf.Metric) {
				expected := []telegraf.Metric{
					testutil.MustMetric(
						"cpu-value",
						map[string]string{
							"host":    "localhost",
							"message": "howdy",
						},
						map[string]interface{}{
							"min":   42.0,
							"max":   42.0,
							"sum":   42.0,
							"count": 1,
						},
						time.Unix(0, 0),
					),
				}
				testutil.RequireMetricsEqual(t, expected, metrics)
			},
		},
		{
			name: "add metric to cache and push",
			plugin: &AzureMonitor{
				Region:     "test",
				ResourceID: "test",
				cache:      make(map[time.Time]map[uint64]*aggregate, 36),
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			addTime:  time.Unix(0, 0),
			pushTime: time.Unix(3600, 0),
			check: func(t *testing.T, plugin *AzureMonitor, metrics []telegraf.Metric) {
				expected := []telegraf.Metric{
					testutil.MustMetric(
						"cpu-value",
						map[string]string{},
						map[string]interface{}{
							"min":   42.0,
							"max":   42.0,
							"sum":   42.0,
							"count": 1,
						},
						time.Unix(0, 0),
					),
				}

				testutil.RequireMetricsEqual(t, expected, metrics)
			},
		},
		{
			name: "added metric are aggregated",
			plugin: &AzureMonitor{
				Region:     "test",
				ResourceID: "test",
				cache:      make(map[time.Time]map[uint64]*aggregate, 36),
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 84,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 2,
					},
					time.Unix(0, 0),
				),
			},
			addTime:  time.Unix(0, 0),
			pushTime: time.Unix(3600, 0),
			check: func(t *testing.T, plugin *AzureMonitor, metrics []telegraf.Metric) {
				expected := []telegraf.Metric{
					testutil.MustMetric(
						"cpu-value",
						map[string]string{},
						map[string]interface{}{
							"min":   2.0,
							"max":   84.0,
							"sum":   128.0,
							"count": 3,
						},
						time.Unix(0, 0),
					),
				}

				testutil.RequireMetricsEqual(t, expected, metrics)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Connect()
			require.NoError(t, err)

			// Reset globals
			tt.plugin.MetricOutsideWindow.Set(0)

			tt.plugin.timeFunc = func() time.Time { return tt.addTime }
			for _, m := range tt.metrics {
				tt.plugin.Add(m)
			}

			tt.plugin.timeFunc = func() time.Time { return tt.pushTime }
			metrics := tt.plugin.Push()
			tt.plugin.Reset()

			tt.check(t, tt.plugin, metrics)
		})
	}
}

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Connect()
			require.NoError(t, err)
			tt.plugin.Add(tt.metric)
			tt.check(t, tt.plugin)
		})
	}
}
