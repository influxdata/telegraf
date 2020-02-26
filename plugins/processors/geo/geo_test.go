package geo

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
		TagKey:    "_ci",
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
				"_ci": "89e8ed4",
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
}
