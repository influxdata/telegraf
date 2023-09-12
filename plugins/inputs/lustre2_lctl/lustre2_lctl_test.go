//go:build linux

package lustre2_lctl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testdataDir = getTestdataDir()
)

func TestHelperMDCActive(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdataDir, "mdc_active"))
	require.NoError(t, err)

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, string(data))

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperOSCActive(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdataDir, "osc_active"))
	require.NoError(t, err)

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, string(data))

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperMDTVolumes(_ *testing.T) {
	data := `mdt.MDT0000`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperRecoveryStatus(_ *testing.T) {
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

func TestHelperMDTJobStats(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdataDir, "mdt_jobstats"))
	require.NoError(t, err)

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, string(data))

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperMDTStats(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdataDir, "mdt_stats"))
	require.NoError(t, err)

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, string(data))

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperOSTKbytesfree(_ *testing.T) {
	data := `26942292504`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperOSTKbytestotal(_ *testing.T) {
	data := `46488188776`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperOSTKbytesavail(_ *testing.T) {
	data := `24598218684`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperOSTVolumes(_ *testing.T) {
	data := `obdfilter.OST0000`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperOSTStats(_ *testing.T) {
	data := `snapshot_time             1693990463.128002841 secs.nsecs
	read_bytes                1487077410 samples [bytes] 4096 4194304 606617630789632
	setattr                   21402423 samples [reqs]`

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, data)

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func TestHelperOSTJobstats(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdataDir, "ost_jobstats"))
	require.NoError(t, err)

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, string(data))

	//nolint:revive // os.Exit called intentionally
	os.Exit(0)
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	tmp := make([]string, 0)
	tmp = append(tmp, command)
	tmp = append(tmp, args...)
	tmpc := strings.Join(tmp, " ")
	switch tmpc {
	case "lctl get_param -n health_check":
		cs := []string{"-test.run=TestHelperHealthCheck", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	// Client
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
	// MDT
	case "lctl get_param -N mdt.*":
		cs := []string{"-test.run=TestHelperMDTVolumes", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n mdt.MDT0000.recovery_status":
		cs := []string{"-test.run=TestHelperRecoveryStatus", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n mdt.MDT0000.job_stats":
		cs := []string{"-test.run=TestHelperMDTJobStats", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n mdt.MDT0000.md_stats":
		cs := []string{"-test.run=TestHelperMDTStats", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	// OST
	case "lctl get_param -N obdfilter.*":
		cs := []string{"-test.run=TestHelperOSTVolumes", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n obdfilter.OST0000.recovery_status":
		cs := []string{"-test.run=TestHelperRecoveryStatus", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n obdfilter.OST0000.job_stats":
		cs := []string{"-test.run=TestHelperOSTJobstats", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n obdfilter.OST0000.stats":
		cs := []string{"-test.run=TestHelperOSTStats", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n obdfilter.OST0000.kbytesfree":
		cs := []string{"-test.run=TestHelperOSTKbytesfree", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n obdfilter.OST0000.kbytesavail":
		cs := []string{"-test.run=TestHelperOSTKbytesavail", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	case "lctl get_param -n obdfilter.OST0000.kbytestotal":
		cs := []string{"-test.run=TestHelperOSTKbytestotal", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	return nil
}

func getTestdataDir() string {
	dir, err := os.Getwd()
	if err != nil {
		// if we cannot even establish the test directory, further progress is meaningless
		panic(err)
	}

	return filepath.Join(dir, "testdata")
}
