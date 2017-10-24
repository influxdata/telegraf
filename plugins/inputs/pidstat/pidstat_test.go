// +build linux

package pidstat

import (
	"fmt"
	"os"
	"os/exec"
	//"path"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

var p = Pidstat{
	interval: 1,

	per_pid:     true,
	per_command: true,

	programs: []string{"kworker*"},
}

func TestGather(t *testing.T) {
	// overwriting exec commands with mock commands
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	var acc testutil.Accumulator

	err := acc.GatherError(p.Gather)
	if err != nil {
		t.Fatal("Error: %s", err)
	}

	tests := []struct {
		measurement string
		fields      map[string]interface{}
		tags        map[string]string
	}{
		{
			"pidstat_pid",
			map[string]interface{}{
				"pct_MEM":       0.03,
				"pct_system":    0.00,
				"pct_CPU":       0.00,
				"CPU":           1.0,
				"nvcswch_per_s": 0.00,
				"RSS":           2468.0,
				"pct_guest":     0.00,
				"kB_wr_per_s":   -1.00,
				"kB_ccwr_per_s": -1.00,
				"majflt_per_s":  0.00,
				"pct_usr":       0.00,
				"cswch_per_s":   0.00,
				"iodelay":       2.0,
				"minflt_per_s":  0.00,
				"VSZ":           102464.0,
				"kB_rd_per_s":   -1.00,
			},
			map[string]string{
				"sys_name": "(tyler-GL753VD)",
				"arch":     "_x86_64_",
				"os":       "Linux",
				"os_ver":   "4.10.0-37-generic",
				"cores":    "(8",
				"PID":      "1005",
				"UID":      "100",
				"Command":  "systemd-timesyn",
			},
		},

		{
			"pidstat_pid",
			map[string]interface{}{
				"RSS":           2616.0,
				"VSZ":           20416.0,
				"pct_MEM":       0.03,
				"pct_guest":     0.00,
				"CPU":           6.0,
				"minflt_per_s":  0.00,
				"majflt_per_s":  0.00,
				"nvcswch_per_s": 0.00,
				"pct_usr":       0.00,
				"pct_system":    0.00,
				"pct_CPU":       0.00,
				"cswch_per_s":   0.03,
			},
			map[string]string{
				"os_ver":   "4.10.0-37-generic",
				"PID":      "1012",
				"Command":  "systemd-logind",
				"cores":    "(8",
				"os":       "Linux",
				"sys_name": "(tyler-GL753VD)",
				"UID":      "0",
				"arch":     "_x86_64_",
			},
		},
	}
	for _, test := range tests {
		acc.AssertContainsTaggedFields(t, test.measurement, test.fields, test.tags)
	}
}

// Helper function that mock the exec.Command call (and call the test binary)
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	//fmt.Printf("execCommand running %s", os.Args[0])
	//for _, c := range cs{
	//fmt.Println(c)
	//}
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// TestHelperProcess isn't a real test. It's used to mock exec.Command
// For example, if you run:
// GO_WANT_HELPER_PROCESS=1 go test -test.run=TestHelperProcess -- sadf -p -- -p -C tmpFile
// it returns mockData["C"] output.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	mockData := map[string]string{

		"-d": `Linux 4.10.0-37-generic (tyler-GL753VD)         10/21/2017      _x86_64_        (8 CPU)

03:06:56 PM   UID       PID   kB_rd/s   kB_wr/s kB_ccwr/s iodelay  Command
03:06:56 PM     0         1     -1.00     -1.00     -1.00      34  systemd
03:06:56 PM     0       305     -1.00     -1.00     -1.00       3  systemd-journal
03:06:56 PM   100      1005     -1.00     -1.00     -1.00       2  systemd-timesyn
03:06:56 PM   104      1056     -1.00     -1.00     -1.00       6  rsyslogd
03:06:56 PM  1000      1784      0.00      0.00      0.00       0  systemd
03:06:56 PM  1000      2752      0.00      0.00      0.00       1  syndaemon
`,

		"-r": `Linux 4.10.0-37-generic (tyler-GL753VD)         10/21/2017      _x86_64_        (8 CPU),

03:06:56 PM   UID       PID  minflt/s  majflt/s     VSZ     RSS   %MEM  Command
03:06:56 PM     0         1      0.17      0.00  185440    5584   0.07  systemd
03:06:56 PM     0       305      0.01      0.00   35524    7024   0.09  systemd-journal
03:06:56 PM     0       330      0.83      0.00   45756    4264   0.05  systemd-udevd
03:06:56 PM   100      1005      0.00      0.00  102464    2468   0.03  systemd-timesyn
03:06:56 PM     0      1012      0.00      0.00   20416    2616   0.03  systemd-logind
03:06:56 PM   104      1056      0.00      0.00  262776    2924   0.04  rsyslogd
03:06:56 PM  1000      1784      0.00      0.00   37012    3512   0.04  systemd
03:06:56 PM  1000      2752      0.00      0.00   22636    1920   0.02  syndaemon"
`,

		"-v": `Linux 4.10.0-37-generic (tyler-GL753VD)         10/21/2017      _x86_64_        (8 CPU),

03:06:56 PM   UID       PID threads   fd-nr  Command
03:06:56 PM  1000      1784       1      15  systemd
03:06:56 PM  1000      2752       1       5  syndaemon
`,

		"-u": `Linux 4.10.0-37-generic (tyler-GL753VD)         10/21/2017      _x86_64_        (8 CPU)
03:06:56 PM   UID       PID    %usr %system  %guest    %CPU   CPU  Command
03:06:56 PM     0         1    0.00    0.00    0.00    0.00     6  systemd
03:06:56 PM     0       305    0.00    0.00    0.00    0.00     6  systemd-journal
03:06:56 PM     0       330    0.00    0.00    0.00    0.00     3  systemd-udevd
03:06:56 PM   100      1005    0.00    0.00    0.00    0.00     1  systemd-timesyn
03:06:56 PM     0      1012    0.00    0.00    0.00    0.00     6  systemd-logind
03:06:56 PM   104      1056    0.00    0.00    0.00    0.00     1  rsyslogd
03:06:56 PM  1000      1784    0.00    0.00    0.00    0.00     1  systemd
03:06:56 PM  1000      2752    0.00    0.00    0.00    0.00     3  syndaemon
`,

		"-w": `Linux 4.10.0-37-generic (tyler-GL753VD)         10/21/2017      _x86_64_        (8 CPU)

03:06:56 PM   UID       PID   cswch/s nvcswch/s  Command
03:06:56 PM     0         1      0.02      0.00  systemd
03:06:56 PM     0       305      0.05      0.00  systemd-journal
03:06:56 PM     0       330      0.05      0.01  systemd-udevd
03:06:56 PM   100      1005      0.00      0.00  systemd-timesyn
03:06:56 PM     0      1012      0.03      0.00  systemd-logind
03:06:56 PM   104      1056      0.00      0.00  rsyslogd
03:06:56 PM  1000      1784      0.00      0.00  systemd
03:06:56 PM  1000      2752      0.62      0.00  syndaemon
`,

		"-s": `Linux 4.10.0-37-generic (tyler-GL753VD)         10/21/2017      _x86_64_        (8 CPU)

03:06:56 PM   UID       PID StkSize  StkRef  Command
03:06:56 PM  1000      1784     132      16  systemd
03:06:56 PM  1000      2752     132       8  syndaemon
`,
	}

	args := os.Args
	fmt.Fprint(os.Stdout, mockData[args[4]])

	os.Exit(0)
}
