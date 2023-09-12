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

func TestGatherMDT(t *testing.T) {
	collect := []string{
		"mdt.*.recovery_status",
		"mdt.*.md_stats",
		"mdt.*.job_stats",
	}

	expected := []telegraf.Metric{
		// recovery
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
			},
			map[string]interface{}{
				"recovery_status": 1,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		// jobstats
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"jobid":  "rsync_env.sh.30026",
			},
			map[string]interface{}{
				"jobstats_read_samples": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"unit":   "usecs",
				"jobid":  "rsync_env.sh.30026",
			},
			map[string]interface{}{
				"jobstats_read_min":   uint64(0),
				"jobstats_read_max":   uint64(0),
				"jobstats_read_sum":   uint64(0),
				"jobstats_read_sumsq": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"jobid":  "rsync_env.sh.30026",
			},
			map[string]interface{}{
				"jobstats_write_samples": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"unit":   "usecs",
				"jobid":  "rsync_env.sh.30026",
			},
			map[string]interface{}{
				"jobstats_write_min":   uint64(0),
				"jobstats_write_max":   uint64(0),
				"jobstats_write_sum":   uint64(0),
				"jobstats_write_sumsq": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"jobid":  "rsync_env.sh.30026",
			},
			map[string]interface{}{
				"jobstats_read_bytes_samples": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"unit":   "bytes",
				"jobid":  "rsync_env.sh.30026",
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
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"jobid":  "rsync_env.sh.30026",
			},
			map[string]interface{}{
				"jobstats_write_bytes_samples": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"unit":   "bytes",
				"jobid":  "rsync_env.sh.30026",
			},
			map[string]interface{}{
				"jobstats_write_bytes_min":   uint64(0),
				"jobstats_write_bytes_max":   uint64(0),
				"jobstats_write_bytes_sum":   uint64(0),
				"jobstats_write_bytes_sumsq": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"jobid":  "rsync_env.sh.30026",
			},
			map[string]interface{}{
				"jobstats_close_samples": uint64(5051),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"unit":   "usecs",
				"jobid":  "rsync_env.sh.30026",
			},
			map[string]interface{}{
				"jobstats_close_min":   uint64(8),
				"jobstats_close_max":   uint64(1541),
				"jobstats_close_sum":   uint64(236426),
				"jobstats_close_sumsq": uint64(19935112),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"jobid":  "cadvisor.0",
			},
			map[string]interface{}{
				"jobstats_statfs_samples": uint64(11261065),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"unit":   "reqs",
				"jobid":  "cadvisor.0",
			},
			map[string]interface{}{
				"jobstats_statfs_min":   uint64(0),
				"jobstats_statfs_max":   uint64(0),
				"jobstats_statfs_sum":   uint64(0),
				"jobstats_statfs_sumsq": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"jobid":  "cadvisor.0",
			},
			map[string]interface{}{
				"jobstats_read_bytes_samples": uint64(11261065),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"unit":   "reqs",
				"jobid":  "cadvisor.0",
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
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"jobid":  "cadvisor.0",
			},
			map[string]interface{}{
				"jobstats_write_bytes_samples": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"unit":   "reqs",
				"jobid":  "cadvisor.0",
			},
			map[string]interface{}{
				"jobstats_write_bytes_min":   uint64(0),
				"jobstats_write_bytes_max":   uint64(0),
				"jobstats_write_bytes_sum":   uint64(0),
				"jobstats_write_bytes_sumsq": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		// stats
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
			},
			map[string]interface{}{
				"stats_open_samples": uint64(137391283844),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"unit":   "reqs",
			},
			map[string]interface{}{
				"stats_open_min":   uint64(0),
				"stats_open_max":   uint64(0),
				"stats_open_sum":   uint64(0),
				"stats_open_sumsq": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
			},
			map[string]interface{}{
				"stats_close_samples": uint64(97376107699),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"unit":   "reqs",
			},
			map[string]interface{}{
				"stats_close_min":   uint64(1),
				"stats_close_max":   uint64(1),
				"stats_close_sum":   uint64(97376107699),
				"stats_close_sumsq": uint64(0),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
			},
			map[string]interface{}{
				"stats_link_samples": uint64(293759658),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"unit":   "usecs",
			},
			map[string]interface{}{
				"stats_link_min":   uint64(5),
				"stats_link_max":   uint64(65525205),
				"stats_link_sum":   uint64(38223033629),
				"stats_link_sumsq": uint64(34619065558020419),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),

		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
			},
			map[string]interface{}{
				"stats_unlink_samples": uint64(675832392),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "MDT0000",
				"unit":   "usecs",
			},
			map[string]interface{}{
				"stats_unlink_min":   uint64(3),
				"stats_unlink_max":   uint64(905437),
				"stats_unlink_sum":   uint64(15359059495),
				"stats_unlink_sumsq": uint64(111766227630681),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
	}

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	var acc testutil.Accumulator
	gatherMDT(collect, "lustre2", &acc)
	actual := acc.GetTelegrafMetrics()

	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}
