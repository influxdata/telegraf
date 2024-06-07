package ecs

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

const cgroupID = "c69461b2c836cc3f0e3e5deb07b1f16e25f6009da2a48bb0adc7dd580befaf55"

func TestParseCgroupV2Stats(t *testing.T) {
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	expected, err := testutil.ParseMetricsFromFile("testdata/cgroupv2/stats.out", parser)
	require.NoError(t, err)

	stats, err := os.Open("testdata/cgroupv2/stats.json")
	require.NoError(t, err)
	parsedStats, err := unmarshalStats(stats)
	require.NoError(t, err)

	tags := map[string]string{
		"test_tag": "test",
	}

	var acc testutil.Accumulator
	memstats(cgroupID, parsedStats[cgroupID], &acc, tags, time.Now())
	cpustats(cgroupID, parsedStats[cgroupID], &acc, tags, time.Now())
	netstats(cgroupID, parsedStats[cgroupID], &acc, tags, time.Now())
	blkstats(cgroupID, parsedStats[cgroupID], &acc, tags, time.Now())

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
}

func TestParseCgroupV2Meta(t *testing.T) {
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	expected, err := testutil.ParseMetricsFromFile("testdata/cgroupv2/meta.out", parser)
	require.NoError(t, err)

	meta, err := os.Open("testdata/cgroupv2/meta.json")
	require.NoError(t, err)
	validMeta, err := unmarshalTask(meta)
	require.NoError(t, err)

	tags := map[string]string{
		"test_tag": "test",
	}

	var acc testutil.Accumulator
	metastats(cgroupID, &validMeta.Containers[0], &acc, tags, time.Now())

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
}
