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

func TestGatherClient(t *testing.T) {

	expected := []telegraf.Metric{
		metric.New(
			"lustre2_client",
			map[string]string{
				"volume": "THL9-MDT0000",
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
				"volume": "thfs1-MDT0000",
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
				"volume": "THL9-OST003d",
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
				"volume": "thfs3-OST0076",
			},
			map[string]interface{}{
				"osc_volume_active": 1,
			},
			time.Unix(0, 1),
			telegraf.Gauge,
		),
	}

	execCommand = fakeClientExecCommand
	defer func() { execCommand = exec.Command }()

	var acc testutil.Accumulator
	gatherClient(true, "lustre2", &acc)
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestHelperMDCActive(t *testing.T) {
	data := `mdc.THL9-MDT0000-mdc-ffff98795f16f000.active=1
	mdc.thfs1-MDT0000-mdc-ffff0781dc988800.active=1`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)
	os.Exit(0)
}

func TestHelperOSCActive(t *testing.T) {
	data := `osc.THL9-OST003d-osc-ffff98795f16f000.active=1
	osc.thfs3-OST0076-osc-ffff0181ddbfa000.active=1`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)
	os.Exit(0)
}

func fakeClientExecCommand(command string, args ...string) *exec.Cmd {

	tmp := make([]string, 0)
	tmp = append(tmp, command)
	tmp = append(tmp, args...)
	tmpc := strings.Join(tmp, " ")
	switch tmpc {
	case "lctl get_param mdc.*.active":
		cs := []string{"-test.run=TestHelperMDCActive", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param osc.*.active":
		cs := []string{"-test.run=TestHelperOSCActive", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n health_check":
		cs := []string{"-test.run=TestHelperHealthCheck", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	return nil
}
