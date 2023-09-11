//go:build linux

package lustre2_lctl

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestGetMDTVolumes(t *testing.T) {
	expected := []string{"THL9-MDT0000", "THL9-MDT0001", "THL9-MDT0002"}

	execCommand = fakeMDTExecuteCommand
	defer func() { execCommand = exec.Command }()

	volumes, err := getMDTVolumes()
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(expected)
	sort.Strings(volumes)

	if diff := cmp.Diff(expected, volumes, nil); diff != "" {
		t.Fatalf("[]string\n--- expected\n+++ actual\n%s", diff)
	}
}

func TestGatherMDTRecoveryStatus(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "THL9-MDT0000",
			},
			map[string]interface{}{
				"recovery_status": 1,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "thfs2-MDT0000",
			},
			map[string]interface{}{
				"recovery_status": 1,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
	}

	execCommand = fakeMDTExecuteCommand
	defer func() { execCommand = exec.Command }()

	var acc testutil.Accumulator
	gatherMDTRecoveryStatus(true, "lustre2_mdt", []string{"THL9-MDT0000", "thfs2-MDT0000"}, &acc)
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGatherMDTJobStats(t *testing.T) {
	// gather all.
	expected := []telegraf.Metric{
		// v2.15 job1
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "THL9-MDT0000",
				"jobid":  "thmc.0",
			},
			map[string]interface{}{
				"jobstats_statfs_samples": uint64(1119088626),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "THL9-MDT0000",
				"unit":   "reqs",
				"jobid":  "thmc.0",
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
				"volume": "THL9-MDT0000",
				"jobid":  "thmc.0",
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
				"volume": "THL9-MDT0000",
				"unit":   "reqs",
				"jobid":  "thmc.0",
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
				"volume": "THL9-MDT0000",
				"jobid":  "thmc.0",
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
				"volume": "THL9-MDT0000",
				"unit":   "reqs",
				"jobid":  "thmc.0",
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

		// v2.15 job2
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "THL9-MDT0000",
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
				"volume": "THL9-MDT0000",
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
				"volume": "THL9-MDT0000",
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
				"volume": "THL9-MDT0000",
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
				"volume": "THL9-MDT0000",
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
				"volume": "THL9-MDT0000",
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

		// v2.17
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
	}

	execCommand = fakeMDTExecuteCommand
	defer func() { execCommand = exec.Command }()

	var acc testutil.Accumulator
	gatherMDTJobstats(Stats{true, true}, "lustre2_mdt", []string{"THL9-MDT0000", "thfs2-MDT0000"}, &acc)
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())

	// gather OP
	acc.ClearMetrics()
	expectedOP := []telegraf.Metric{
		// v2.15 job1
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "THL9-MDT0000",
				"jobid":  "thmc.0",
			},
			map[string]interface{}{
				"jobstats_statfs_samples": uint64(1119088626),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "THL9-MDT0000",
				"unit":   "reqs",
				"jobid":  "thmc.0",
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

		// v2.15 job2
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "THL9-MDT0000",
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
				"volume": "THL9-MDT0000",
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

		// v2.17
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
	}
	gatherMDTJobstats(Stats{false, true}, "lustre2_mdt", []string{"THL9-MDT0000", "thfs2-MDT0000"}, &acc)
	actual = acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expectedOP, actual, testutil.IgnoreTime(), testutil.SortMetrics())

	// gather RW
	acc.ClearMetrics()
	expectedRW := []telegraf.Metric{
		// v2.15 job1
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "THL9-MDT0000",
				"jobid":  "thmc.0",
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
				"volume": "THL9-MDT0000",
				"unit":   "reqs",
				"jobid":  "thmc.0",
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
				"volume": "THL9-MDT0000",
				"jobid":  "thmc.0",
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
				"volume": "THL9-MDT0000",
				"unit":   "reqs",
				"jobid":  "thmc.0",
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

		// v2.15 job2
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "THL9-MDT0000",
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
				"volume": "THL9-MDT0000",
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
				"volume": "THL9-MDT0000",
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
				"volume": "THL9-MDT0000",
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

		// v2.17
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
				"volume": "thfs2-MDT0000",
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
	}
	gatherMDTJobstats(Stats{true, false}, "lustre2_mdt", []string{"THL9-MDT0000", "thfs2-MDT0000"}, &acc)
	actual = acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expectedRW, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGatherMDTStats(t *testing.T) {
	expected := []telegraf.Metric{
		// lustre@v2.15
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "THL9-MDT0000",
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
				"volume": "THL9-MDT0000",
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
				"volume": "THL9-MDT0000",
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
				"volume": "THL9-MDT0000",
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

		// lustre@v2.17
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "thfs2-MDT0000",
			},
			map[string]interface{}{
				"stats_open_samples": uint64(293759658),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "thfs2-MDT0000",
				"unit":   "usecs",
			},
			map[string]interface{}{
				"stats_open_min":   uint64(5),
				"stats_open_max":   uint64(65525205),
				"stats_open_sum":   uint64(38223033629),
				"stats_open_sumsq": uint64(34619065558020419),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),

		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "thfs2-MDT0000",
			},
			map[string]interface{}{
				"stats_close_samples": uint64(675832392),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
		metric.New(
			"lustre2_mdt",
			map[string]string{
				"volume": "thfs2-MDT0000",
				"unit":   "usecs",
			},
			map[string]interface{}{
				"stats_close_min":   uint64(3),
				"stats_close_max":   uint64(905437),
				"stats_close_sum":   uint64(15359059495),
				"stats_close_sumsq": uint64(111766227630681),
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
	}

	execCommand = fakeMDTExecuteCommand
	defer func() { execCommand = exec.Command }()

	var acc testutil.Accumulator
	// gatherMDTStats(Stats{true, true}, "lustre2_mdt", []string{"THL9-MDT0000"}, &acc)
	gatherMDTStats(Stats{true, true}, "lustre2_mdt", []string{"THL9-MDT0000", "thfs2-MDT0000"}, &acc)

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestHelperMDTStatsV215(_ *testing.T) {
	data := `snapshot_time             1694140455.278503266 secs.nsecs
	open                      137391283844 samples [reqs]
	close                     97376107699 samples [reqs] 1 1 97376107699`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperMDTStatsV217(_ *testing.T) {
	data := `snapshot_time             1964295.787757337 secs.nsecs
	start_time                0.000000000 secs.nsecs
	elapsed_time              1964295.787757337 secs.nsecs
	open                      293759658 samples [usecs] 5 65525205 38223033629 34619065558020419
	close                     675832392 samples [usecs] 3 905437 15359059495 111766227630681`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperMDTJobStatsV215(_ *testing.T) {
	data := `job_stats:
	- job_id:          thmc.0
	  snapshot_time:   1694072655
	  statfs:          { samples:  1119088626, unit:  reqs }
	  read_bytes:      { samples:           0, unit:  reqs, min:       0, max:       0, sum:               0 }
	  write_bytes:     { samples:           0, unit:  reqs, min:       0, max:       0, sum:               0 }
	- job_id:          cadvisor.0
	  snapshot_time:   1694072652
	  statfs:          { samples:    11261065, unit:  reqs }
	  read_bytes:      { samples:    11261065, unit:  reqs, min:       0, max:       0, sum:               0 }
	  write_bytes:     { samples:           0, unit:  reqs, min:       0, max:       0, sum:               0 }`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperMDTJobStatsV217(_ *testing.T) {
	data := `job_stats:
	- job_id:          rsync_env.sh.30026
	snapshot_time   : 1896247.780050045 secs.nsecs
	start_time      : 379440.150439687 secs.nsecs
	elapsed_time    : 1516807.629610358 secs.nsecs
	  close:           { samples:        5051, unit: usecs, min:        8, max:     1541, sum:           236426, sumsq:           19935112 }
	  read:            { samples:           0, unit: usecs, min:        0, max:        0, sum:                0, sumsq:                  0 }
	  write:           { samples:           0, unit: usecs, min:        0, max:        0, sum:                0, sumsq:                  0 }
	  read_bytes:      { samples:           0, unit: bytes, min:        0, max:        0, sum:                0, sumsq:                  0 }
	  write_bytes:     { samples:           0, unit: bytes, min:        0, max:        0, sum:                0, sumsq:                  0 }`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperMDTRecoveryStatusV215(_ *testing.T) {

	data := `status: COMPLETE
	recovery_start: 61
	recovery_duration: 44
	completed_clients: 2/2
	replayed_requests: 0
	last_transno: 8609762186
	VBR: DISABLED
	IR: DISABLED`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperMDTRecoveryStatusV217(_ *testing.T) {

	data := `status: COMPLETE
	recovery_start: 1692176311
	recovery_duration: 86
	completed_clients: 2738/2738
	replayed_requests: 5203
	last_transno: 52878551595
	VBR: DISABLED
	IR: DISABLED`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperMDTVolumes(_ *testing.T) {
	data := `mdt.THL9-MDT0000
	mdt.THL9-MDT0001
	mdt.THL9-MDT0002`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func fakeMDTExecuteCommand(command string, args ...string) *exec.Cmd {
	tmp := make([]string, 0)
	tmp = append(tmp, command)
	tmp = append(tmp, args...)
	tmpc := strings.Join(tmp, " ")
	// fmt.Println(tmpc)
	switch tmpc {
	case "lctl get_param -N mdt.*":
		cs := []string{"-test.run=TestHelperMDTVolumes", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n mdt.THL9-MDT0000.recovery_status":
		cs := []string{"-test.run=TestHelperMDTRecoveryStatusV215", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n mdt.thfs2-MDT0000.recovery_status":
		cs := []string{"-test.run=TestHelperMDTRecoveryStatusV217", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n mdt.THL9-MDT0000.job_stats":
		cs := []string{"-test.run=TestHelperMDTJobStatsV215", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n mdt.thfs2-MDT0000.job_stats":
		cs := []string{"-test.run=TestHelperMDTJobStatsV217", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n mdt.THL9-MDT0000.md_stats":
		cs := []string{"-test.run=TestHelperMDTStatsV215", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n mdt.thfs2-MDT0000.md_stats":
		cs := []string{"-test.run=TestHelperMDTStatsV217", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	return nil
}
