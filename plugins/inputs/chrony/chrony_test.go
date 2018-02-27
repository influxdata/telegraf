package chrony

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func TestGather(t *testing.T) {
	c := Chrony{
		path: "chronyc",
	}
	// overwriting exec commands with mock commands
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	var acc testutil.Accumulator

	err := c.Gather(&acc)
	if err != nil {
		t.Fatal(err)
	}

	tags := map[string]string{
		"reference_id": "192.168.1.22",
		"leap_status":  "not synchronized",
		"stratum":      "3",
	}
	fields := map[string]interface{}{
		"system_time":     0.000020390,
		"last_offset":     0.000012651,
		"rms_offset":      0.000025577,
		"frequency":       -16.001,
		"residual_freq":   0.0,
		"skew":            0.006,
		"root_delay":      0.001655,
		"root_dispersion": 0.003307,
		"update_interval": 507.2,
	}

	acc.AssertContainsTaggedFields(t, "chrony", fields, tags)

	// test with dns lookup
	c.DNSLookup = true
	err = c.Gather(&acc)
	if err != nil {
		t.Fatal(err)
	}
	acc.AssertContainsTaggedFields(t, "chrony", fields, tags)

}

// fackeExecCommand is a helper function that mock
// the exec.Command call (and call the test binary)
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// TestHelperProcess isn't a real test. It's used to mock exec.Command
// For example, if you run:
// GO_WANT_HELPER_PROCESS=1 go test -test.run=TestHelperProcess -- chrony tracking
// it returns below mockData.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	lookup := "Reference ID    : 192.168.1.22 (ntp.example.com)\n"
	noLookup := "Reference ID    : 192.168.1.22 (192.168.1.22)\n"
	mockData := `Stratum         : 3
Ref time (UTC)  : Thu May 12 14:27:07 2016
System time     : 0.000020390 seconds fast of NTP time
Last offset     : +0.000012651 seconds
RMS offset      : 0.000025577 seconds
Frequency       : 16.001 ppm slow
Residual freq   : -0.000 ppm
Skew            : 0.006 ppm
Root delay      : 0.001655 seconds
Root dispersion : 0.003307 seconds
Update interval : 507.2 seconds
Leap status     : Not synchronized
`

	args := os.Args

	// Previous arguments are tests stuff, that looks like :
	// /tmp/go-build970079519/…/_test/integration.test -test.run=TestHelperProcess --
	cmd, args := args[3], args[4:]

	if cmd == "chronyc" {
		if args[0] == "tracking" {
			fmt.Fprint(os.Stdout, lookup+mockData)
		} else {
			fmt.Fprint(os.Stdout, noLookup+mockData)
		}
	} else {
		fmt.Fprint(os.Stdout, "command not found")
		os.Exit(1)

	}
	os.Exit(0)
}
