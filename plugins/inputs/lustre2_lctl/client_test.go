//go:build linux

package lustre2_lctl

import (
	"os/exec"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestGatherClient(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"lustre2_client",
			map[string]string{
				"volume": "MDT0000",
			},
			map[string]interface{}{
				"mdc_volume_active": 1,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_client",
			map[string]string{
				"volume": "MDT0001",
			},
			map[string]interface{}{
				"mdc_volume_active": 1,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_client",
			map[string]string{
				"volume": "OST003d",
			},
			map[string]interface{}{
				"osc_volume_active": 1,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_client",
			map[string]string{
				"volume": "OST0076",
			},
			map[string]interface{}{
				"osc_volume_active": 1,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
	}

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	var acc testutil.Accumulator
	gatherClient([]string{"mdc.*.active", "osc.*.active"}, "lustre2", &acc)
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}
