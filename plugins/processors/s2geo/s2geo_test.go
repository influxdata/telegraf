package s2geo

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
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

	metric := testutil.MustMetric(
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

	actual := plugin.Apply(metric)
	testutil.RequireMetricsEqual(t, expected, actual)
	actual = pluginMostlyDefault.Apply(metric)
	testutil.RequireMetricsEqual(t, expected, actual)
}
