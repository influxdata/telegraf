package azure_monitor

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name   string
		plugin *AzureMonitor
		metric telegraf.Metric
		check  func(t *testing.T, plugin *AzureMonitor)
	}{
		{
			name: "metric outside window is dropped",
			plugin: &AzureMonitor{
				Region:     "test",
				ResourceID: "test",
			},
			metric: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
			check: func(t *testing.T, plugin *AzureMonitor) {
				require.Equal(t, int64(1), plugin.MetricOutsideWindow.Get())
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
