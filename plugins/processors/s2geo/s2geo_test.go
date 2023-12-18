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

func TestGeoUserField(t *testing.T) {
	plugin := &Geo{
		LatField:       "lat",
		LonField:       "lon",
		TagKey:         "s2_cell_id",
		CellLevel:      11,
		CellLevelField: "level",
	}
	err := plugin.Init()
	require.NoError(t, err)

	tc := []struct {
		name   string
		level  int
		cellID string
	}{
		{name: "to small", level: -1, cellID: "89e8ed4"},
		{name: "too large", level: 31, cellID: "89e8ed4"},
		{name: "level 0", level: 0, cellID: "9"},
		{name: "level 1", level: 1, cellID: "8c"},
		{name: "level 2", level: 2, cellID: "89"},
		{name: "level 3", level: 3, cellID: "89c"},
		{name: "level 4", level: 4, cellID: "89f"},
		{name: "level 5", level: 5, cellID: "89ec"},
		{name: "level 6", level: 6, cellID: "89e9"},
		{name: "level 7", level: 7, cellID: "89e8c"},
		{name: "level 8", level: 8, cellID: "89e8f"},
		{name: "level 9", level: 9, cellID: "89e8ec"},
		{name: "level 10", level: 10, cellID: "89e8ed"},
		{name: "level 11", level: 11, cellID: "89e8ed4"},
		{name: "level 12", level: 12, cellID: "89e8ed5"},
		{name: "level 13", level: 13, cellID: "89e8ed5c"},
		{name: "level 14", level: 14, cellID: "89e8ed59"},
		{name: "level 15", level: 15, cellID: "89e8ed58c"},
		{name: "level 16", level: 16, cellID: "89e8ed58f"},
		{name: "level 17", level: 17, cellID: "89e8ed58fc"},
		{name: "level 18", level: 18, cellID: "89e8ed58ff"},
		{name: "level 19", level: 19, cellID: "89e8ed58fe4"},
		{name: "level 20", level: 20, cellID: "89e8ed58fe1"},
		{name: "level 21", level: 21, cellID: "89e8ed58fe0c"},
		{name: "level 22", level: 22, cellID: "89e8ed58fe09"},
		{name: "level 23", level: 23, cellID: "89e8ed58fe084"},
		{name: "level 24", level: 24, cellID: "89e8ed58fe087"},
		{name: "level 25", level: 25, cellID: "89e8ed58fe087c"},
		{name: "level 26", level: 26, cellID: "89e8ed58fe087f"},
		{name: "level 27", level: 27, cellID: "89e8ed58fe087e4"},
		{name: "level 28", level: 28, cellID: "89e8ed58fe087e5"},
		{name: "level 29", level: 29, cellID: "89e8ed58fe087e44"},
		{name: "level 30", level: 30, cellID: "89e8ed58fe087e45"},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			metric := testutil.MustMetric(
				"mta",
				map[string]string{},
				map[string]interface{}{
					"lat":   40.878738,
					"lon":   -72.517572,
					"level": c.level,
				},
				time.Unix(1578603600, 0),
			)
			expected := []telegraf.Metric{
				testutil.MustMetric(
					"mta",
					map[string]string{
						"s2_cell_id": c.cellID,
					},
					map[string]interface{}{
						"lat":   40.878738,
						"lon":   -72.517572,
						"level": c.level,
					},
					time.Unix(1578603600, 0),
				),
			}

			actual := plugin.Apply(metric)
			testutil.RequireMetricsEqual(t, expected, actual)
		})
	}
}
