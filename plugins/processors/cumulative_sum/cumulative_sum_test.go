package cumulative_sum

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// TestCumulativeSum perform sum of two metrics
func TestCumulativeSum(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_sum": float64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_sum": float64(4)},
			time.Unix(0, 0),
		),
	}

	plugin := &CumulativeSum{}
	require.NoError(t, plugin.Init())

	actual := plugin.Apply(
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": 1},
			time.Unix(0, 0),
		), metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": 3},
			time.Unix(0, 0),
		))
	testutil.RequireMetricsEqual(t, expected, actual)
}

// TestCumulativeSum perform sum of two metrics and left original field
func TestCumulativeSumKeepOriginalFieldTrue(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(1), "value_sum": float64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(3), "value_sum": float64(4)},
			time.Unix(0, 0),
		),
	}

	plugin := &CumulativeSum{
		KeepOriginalField: true,
	}
	require.NoError(t, plugin.Init())

	actual := plugin.Apply(
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": 1},
			time.Unix(0, 0),
		), metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": 3},
			time.Unix(0, 0),
		))
	testutil.RequireMetricsEqual(t, expected, actual)
}

// TestCumulativeSum perform sum of two metrics and don't touch string field
func TestCumulativeSumStringField(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_name": "name", "value_sum": float64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_name": "name", "value_sum": float64(4)},
			time.Unix(0, 0),
		),
	}

	plugin := &CumulativeSum{}
	require.NoError(t, plugin.Init())

	actual := plugin.Apply(
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_name": "name", "value": float64(1)},
			time.Unix(0, 0),
		), metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_name": "name", "value": float64(3)},
			time.Unix(0, 0),
		))
	testutil.RequireMetricsEqual(t, expected, actual)
}

// TestCumulativeSum don't perform sum of two metrics with filtered out fields
func TestCumulativeFieldFilteredOut(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_name": "name", "value": float64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_name": "name", "value": float64(3)},
			time.Unix(0, 0),
		),
	}

	plugin := &CumulativeSum{}
	plugin.Fields = []string{"another_name"}
	require.NoError(t, plugin.Init())

	// same as expected
	actual := plugin.Apply(
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_name": "name", "value": float64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_name": "name", "value": float64(3)},
			time.Unix(0, 0),
		))
	testutil.RequireMetricsEqual(t, expected, actual)
}

// TestCumulativeSum perform sum of two metrics when field name match config
func TestCumulativeFieldMatch(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_name": "name", "value_sum": float64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_name": "name", "value_sum": float64(4)},
			time.Unix(0, 0),
		),
	}

	plugin := &CumulativeSum{}
	plugin.Fields = []string{"value"}
	require.NoError(t, plugin.Init())

	actual := plugin.Apply(
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_name": "name", "value_sum": float64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_name": "name", "value_sum": float64(4)},
			time.Unix(0, 0),
		))
	testutil.RequireMetricsEqual(t, expected, actual)
}

// TestCumulativeSum clean up internal interval for metric fields that wasn't updated too long
func TestCumulativeSumCleanedAccumulatorAfterCleanupInterval(t *testing.T) {
	currentTime := time.Unix(5, 0)

	timeNow = func() time.Time {
		return currentTime
	}
	t.Cleanup(func() {
		timeNow = time.Now
	})

	plugin := &CumulativeSum{}
	plugin.ResetInterval = config.Duration(60 * time.Second)
	require.NoError(t, plugin.Init())

	plugin.Apply(
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": 1},
			time.Unix(0, 0),
		), metric.New(
			"m2",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": 7},
			time.Unix(0, 0),
		))

	currentTime = time.Unix(30, 0)

	plugin.Apply(
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": 1},
			time.Unix(0, 0),
		))

	currentTime = time.Unix(70, 0)

	// force clean up
	plugin.Apply()

	expected := []telegraf.Metric{
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_sum": float64(3)},
			time.Unix(0, 0),
		),
		metric.New(
			"m2",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value_sum": float64(7)},
			time.Unix(0, 0),
		),
	}

	actual := plugin.Apply(
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": 1},
			time.Unix(0, 0),
		),
		metric.New(
			"m2",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": 7},
			time.Unix(0, 0),
		),
	)

	testutil.RequireMetricsEqual(t, expected, actual)
}
