//go:build windows

package mem

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/testutil"
)

func TestMemStatsCollectExtended(t *testing.T) {
	plugin := &Mem{
		ps:              psutil.NewSystemPS(),
		CollectExtended: true,
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"mem",
			map[string]string{},
			map[string]interface{}{
				"total":             uint64(0),
				"available":         uint64(0),
				"used":              uint64(0),
				"used_percent":      float64(0),
				"available_percent": float64(0),
				"commit_limit":      uint64(0),
				"commit_total":      uint64(0),
				"virtual_total":     uint64(0),
				"virtual_avail":     uint64(0),
				"phys_total":        uint64(0),
				"phys_avail":        uint64(0),
				"page_file_total":   uint64(0),
				"page_file_avail":   uint64(0),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}
	testutil.RequireMetricsStructureEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
