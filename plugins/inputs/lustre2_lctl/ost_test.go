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

func TestGatherOST(t *testing.T) {
	collect := []string{
		"obdfilter.*.stats",
		"obdfilter.*.job_stats",
		"obdfilter.*.recovery_status",
		"obdfilter.*.kbytesfree",
		"obdfilter.*.kbytesavail",
		"obdfilter.*.kbytestotal",
	}

	expected := []telegraf.Metric{
		// recovery
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
			},
			map[string]interface{}{
				"recovery_status": 1,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),

		// jobstats
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
				"jobid":  "1306853",
			},
			map[string]interface{}{
				"jobstats_read_bytes_samples": uint64(169256),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
				"unit":   "bytes",
				"jobid":  "1306853",
			},
			map[string]interface{}{
				"jobstats_read_bytes_min":   uint64(4096),
				"jobstats_read_bytes_max":   uint64(4194304),
				"jobstats_read_bytes_sum":   uint64(62372188160),
				"jobstats_read_bytes_sumsq": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
				"jobid":  "1306853",
			},
			map[string]interface{}{
				"jobstats_destroy_samples": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
				"unit":   "reqs",
				"jobid":  "1306853",
			},
			map[string]interface{}{
				"jobstats_destroy_min":   uint64(0),
				"jobstats_destroy_max":   uint64(0),
				"jobstats_destroy_sum":   uint64(0),
				"jobstats_destroy_sumsq": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
				"jobid":  "kworker/13:2.0",
			},
			map[string]interface{}{
				"jobstats_read_bytes_samples": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
				"unit":   "bytes",
				"jobid":  "kworker/13:2.0",
			},
			map[string]interface{}{
				"jobstats_read_bytes_min":   uint64(0),
				"jobstats_read_bytes_max":   uint64(0),
				"jobstats_read_bytes_sum":   uint64(0),
				"jobstats_read_bytes_sumsq": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
				"jobid":  "kworker/13:2.0",
			},
			map[string]interface{}{
				"jobstats_create_samples": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
				"unit":   "reqs",
				"jobid":  "kworker/13:2.0",
			},
			map[string]interface{}{
				"jobstats_create_min":   uint64(0),
				"jobstats_create_max":   uint64(0),
				"jobstats_create_sum":   uint64(0),
				"jobstats_create_sumsq": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		// stats
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
			},
			map[string]interface{}{
				"stats_read_bytes_samples": uint64(1487077410),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
				"unit":   "bytes",
			},
			map[string]interface{}{
				"stats_read_bytes_min":   uint64(4096),
				"stats_read_bytes_max":   uint64(4194304),
				"stats_read_bytes_sum":   uint64(606617630789632),
				"stats_read_bytes_sumsq": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
			},
			map[string]interface{}{
				"stats_setattr_samples": uint64(21402423),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
				"unit":   "reqs",
			},
			map[string]interface{}{
				"stats_setattr_min":   uint64(0),
				"stats_setattr_max":   uint64(0),
				"stats_setattr_sum":   uint64(0),
				"stats_setattr_sumsq": uint64(0),
			},
			time.Unix(0, 2),
			telegraf.Gauge,
		),
		// capacity
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
			},
			map[string]interface{}{
				"capacity_kbytestotal": 46488188776,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
			},
			map[string]interface{}{
				"capacity_kbytesavail": 24598218684,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "OST0000",
			},
			map[string]interface{}{
				"capacity_kbytesfree": 26942292504,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
	}

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	var acc testutil.Accumulator
	gatherOST(collect, "lustre2", &acc)
	actual := acc.GetTelegrafMetrics()

	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}
