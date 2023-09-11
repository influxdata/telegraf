//go:build linux

package lustre2_lctl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestGatherOSTRecoveryStatus(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "THL9-OST0004",
			},
			map[string]interface{}{
				"recovery_status": 1,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
	}

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	var acc testutil.Accumulator
	gatherOSTObdfilterRecoveryStatus(true, []string{"THL9-OST0004"}, "lustre2_ost", &acc)
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGatherOSTObdfilterJobstats(t *testing.T) {

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	var acc testutil.Accumulator

	expected := []telegraf.Metric{
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "THL9-OST0004",
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
				"volume": "THL9-OST0004",
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
				"volume": "THL9-OST0004",
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
				"volume": "THL9-OST0004",
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
				"volume": "THL9-OST0004",
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
				"volume": "THL9-OST0004",
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
				"volume": "THL9-OST0004",
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
				"volume": "THL9-OST0004",
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
	}

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	gatherOSTObdfilterJobstats(Stats{true, true}, []string{"THL9-OST0004"}, "lustre2_ost", &acc)
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGatherOSTObdfilterStats(t *testing.T) {

	/* 1. */
	stats := Stats{RW: false, OP: false}
	expected := []telegraf.Metric{}
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	var acc testutil.Accumulator
	gatherOSTObdfilterStats(stats, []string{"THL9-OST0004"}, "lustre2_ost", &acc)
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())

	/* 2. */
	stats = Stats{RW: true, OP: false}
	acc.ClearMetrics()
	expected = []telegraf.Metric{
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "THL9-OST0004",
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
				"volume": "THL9-OST0004",
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
	}

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	gatherOSTObdfilterStats(stats, []string{"THL9-OST0004"}, "lustre2_ost", &acc)
	actual = acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())

	/* 3. */
	stats = Stats{RW: false, OP: true}
	acc.ClearMetrics()
	expected = []telegraf.Metric{
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "THL9-OST0004",
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
				"volume": "THL9-OST0004",
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
	}

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	gatherOSTObdfilterStats(stats, []string{"THL9-OST0004"}, "lustre2_ost", &acc)
	actual = acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGatherOSTCapacity(t *testing.T) {

	expected := []telegraf.Metric{
		metric.New(
			"lustre2_ost",
			map[string]string{
				"volume": "THL9-OST0004",
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
				"volume": "THL9-OST0004",
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
				"volume": "THL9-OST0004",
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
	gatherOSTCapacity(true, []string{"THL9-OST0004"}, "lustre2_ost", &acc)
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestHelperRecoveryStatus(t *testing.T) {

	data := `status: COMPLETE
	recovery_start: 55
	recovery_duration: 0
	completed_clients: 1/1
	replayed_requests: 0
	last_transno: 17180113303
	VBR: DISABLED
	IR: DISABLED`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)
	os.Exit(0)
}

func TestHelperHealthCheck(t *testing.T) {

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, `healthy`)
	os.Exit(0)
}

func TestTestHelperJobstats(t *testing.T) {
	data := `job_stats:
	- job_id:          1306853
	  snapshot_time:   1693988320
	  read_bytes:      { samples:      169256, unit: bytes, min:    4096, max: 4194304, sum:     62372188160 }
	  destroy:         { samples:           0, unit:  reqs }
	- job_id:          kworker/13:2.0
	  snapshot_time:   1693988353
	  read_bytes:      { samples:           0, unit: bytes, min:       0, max:       0, sum:               0 }
	  create:          { samples:           0, unit:  reqs }
  `

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)
	os.Exit(0)
}

func TestHelperStats(t *testing.T) {
	data := `snapshot_time             1693990463.128002841 secs.nsecs
	read_bytes                1487077410 samples [bytes] 4096 4194304 606617630789632
	setattr                   21402423 samples [reqs]`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)
	os.Exit(0)
}

func TestHelperKbytestotal(t *testing.T) {
	data := `46488188776`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)
	os.Exit(0)
}

func TestHelperKbytesavail(t *testing.T) {
	data := `24598218684`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)
	os.Exit(0)
}

func TestHelperKbytesfree(t *testing.T) {
	data := `26942292504`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)
	os.Exit(0)
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {

	tmp := make([]string, 0)
	tmp = append(tmp, command)
	tmp = append(tmp, args...)
	tmpc := strings.Join(tmp, " ")

	if strings.Contains(tmpc, "health_check") {
		cs := []string{"-test.run=TestHelperHealthCheck", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	if strings.Contains(tmpc, "recovery_status") {
		cs := []string{"-test.run=TestHelperRecoveryStatus", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	if strings.Contains(tmpc, "job_stats") {
		cs := []string{"-test.run=TestHelperJobstats", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	if strings.Contains(tmpc, ".stats") {
		cs := []string{"-test.run=TestHelperStats", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	if strings.Contains(tmpc, "kbytestotal") {
		cs := []string{"-test.run=TestHelperKbytestotal", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	if strings.Contains(tmpc, "kbytesavail") {
		cs := []string{"-test.run=TestHelperKbytesavail", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	if strings.Contains(tmpc, "kbytesfree") {
		cs := []string{"-test.run=TestHelperKbytesfree", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	return nil
}
