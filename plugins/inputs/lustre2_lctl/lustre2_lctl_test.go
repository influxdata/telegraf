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

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	c := strings.Join(args[3:], " ")

	switch c {
	case "lctl get_param -n health_check":
		fmt.Fprint(os.Stdout, `healthy`)
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	// Client
	case "lctl get_param mdc.*.active":
		data, err := os.ReadFile(filepath.Join(testdataDir, "mdc_active"))
		require.NoError(t, err)
		fmt.Fprint(os.Stdout, string(data))
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	case "lctl get_param osc.*.active":
		data, err := os.ReadFile(filepath.Join(testdataDir, "osc_active"))
		require.NoError(t, err)
		fmt.Fprint(os.Stdout, string(data))
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	// MDT
	case "lctl get_param -N mdt.*":
		fmt.Fprint(os.Stdout, `mdt.MDT0000`)
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	case "lctl get_param -n mdt.MDT0000.recovery_status":
		data := `status: COMPLETE
	recovery_start: 1692176311
	recovery_duration: 86
	completed_clients: 2738/2738
	replayed_requests: 5203
	last_transno: 52878551595
	VBR: DISABLED
	IR: DISABLED`
		fmt.Fprint(os.Stdout, data)
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	case "lctl get_param -n mdt.MDT0000.job_stats":
		data, err := os.ReadFile(filepath.Join(testdataDir, "mdt_jobstats"))
		require.NoError(t, err)
		fmt.Fprint(os.Stdout, string(data))
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	case "lctl get_param -n mdt.MDT0000.md_stats":
		data, err := os.ReadFile(filepath.Join(testdataDir, "mdt_stats"))
		require.NoError(t, err)
		fmt.Fprint(os.Stdout, string(data))
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	// OST
	case "lctl get_param -N obdfilter.*":
		fmt.Fprint(os.Stdout, `obdfilter.OST0000`)
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	case "lctl get_param -n obdfilter.OST0000.recovery_status":
		data := `status: COMPLETE
	recovery_start: 1692176311
	recovery_duration: 86
	completed_clients: 2738/2738
	replayed_requests: 5203
	last_transno: 52878551595
	VBR: DISABLED
	IR: DISABLED`
		fmt.Fprint(os.Stdout, data)
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	case "lctl get_param -n obdfilter.OST0000.job_stats":
		data, err := os.ReadFile(filepath.Join(testdataDir, "ost_jobstats"))
		require.NoError(t, err)
		fmt.Fprint(os.Stdout, string(data))
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	case "lctl get_param -n obdfilter.OST0000.stats":
		data := `snapshot_time             1693990463.128002841 secs.nsecs
	read_bytes                1487077410 samples [bytes] 4096 4194304 606617630789632
	setattr                   21402423 samples [reqs]`
		fmt.Fprint(os.Stdout, data)
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	case "lctl get_param -n obdfilter.OST0000.kbytesfree":
		data := `26942292504`
		fmt.Fprint(os.Stdout, data)
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	case "lctl get_param -n obdfilter.OST0000.kbytesavail":
		data := `24598218684`
		fmt.Fprint(os.Stdout, data)
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	case "lctl get_param -n obdfilter.OST0000.kbytestotal":
		data := `46488188776`
		fmt.Fprint(os.Stdout, data)
		//nolint:revive // os.Exit called intentionally
		os.Exit(0)
	default:
		fmt.Fprint(os.Stdout, "invalid argument")
		//nolint:revive // os.Exit called intentionally
		os.Exit(1)
	}

	//nolint:revive // error code is important for this "test"
	os.Exit(0)
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func getTestdataDir() string {
	dir, err := os.Getwd()
	if err != nil {
		// if we cannot even establish the test directory, further progress is meaningless
		panic(err)
	}

	return filepath.Join(dir, "testdata")
}

