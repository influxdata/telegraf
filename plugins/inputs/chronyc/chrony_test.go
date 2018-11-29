package chronyc_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/chronyc"
	"github.com/influxdata/telegraf/testutil"
)

func TestGather(t *testing.T) {

	c := chronyc.Chrony{
		ChronycPath:     "chronyc",
		ChronycCommands: []string{"tracking", "serverstats", "sources"},
	}
	// overwriting exec commands with mock commands
	chronyc.ExecCommand = makeExecCommand("TestHelperProcess")
	defer func() { chronyc.ExecCommand = exec.Command }()
	var acc testutil.Accumulator

	err := c.Gather(&acc)
	if err != nil {
		t.Fatal(err)
	}

	tags := map[string]string{
		"command": "tracking",
		"clockId": "chrony",
	}
	fields := map[string]interface{}{
		"refId":            "PPS",
		"refIdHex":         "50505300",
		"stratum":          int64(1),
		"refTime":          1541793798.895264285,
		"systemTimeOffset": -0.000001007,
		"lastOffset":       0.000000291,
		"rmsOffset":        0.000000239,
		"frequency":        -17.957,
		"freqResidual":     0.000,
		"freqSkew":         0.005,
		"rootDelay":        0.000001000,
		"rootDispersion":   0.000010123,
		"updateInterval":   16.0,
		"leapStatus":       "Normal",
	}

	acc.AssertContainsTaggedFields(t, "chronyc", fields, tags)

	fields = map[string]interface{}{
		"ntpPacketsReceived":      int64(191),
		"ntpPacketsDropped":       int64(222),
		"commandPacketsReceived":  int64(183),
		"commandPacketsDropped":   int64(111),
		"clientLogRecordsDropped": int64(231),
	}
	tags = map[string]string{
		"command": "serverstats",
		"clockId": "chrony",
	}
	acc.AssertContainsTaggedFields(t, "chronyc", fields, tags)

}

func _TestGatherEmptySources(t *testing.T) {

	c := chronyc.Chrony{
		ChronycPath:     "chronyc",
		ChronycCommands: []string{"tracking", "sources", "serverstats"},
	}

	// overwriting exec commands with mock commands
	chronyc.ExecCommand = makeExecCommand("TestHelperProcess")
	defer func() { chronyc.ExecCommand = exec.Command }()
	var acc testutil.Accumulator

	err := c.Gather(&acc)
	if err != nil {
		t.Fatal(err)
	}

	if acc.HasField("chronyc", "clockMode") {
		t.Fatal("Output shall not contain any sources")
	}

}

func makeExecCommand(helper string) func(string, ...string) *exec.Cmd {
	// fakeExecCommand is a helper function that mock
	// the exec.Command call (and call the test binary)
	return func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=" + helper, "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
}

// TestHelperProcess isn't a real test. It's used to mock exec.Command
// For example, if you run:
// GO_WANT_HELPER_PROCESS=1 go test -test.run=TestHelperProcess -- chrony tracking
// it returns below mockData.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args

	// Previous arguments are tests stuff, that looks like :
	// /tmp/go-build970079519/…/_test/integration.test -test.run=TestHelperProcess --

	cmd, args := args[3], args[4:]

	if cmd != "chronyc" {
		fmt.Fprintf(os.Stdout, "command not found\n")
		os.Exit(1)
	}
	if args[0] != "-c" {
		fmt.Println("First argument shall be -c")
		os.Exit(1)
	}

	trackingOut := "50505300,PPS,1,1541793798.895264285,-0.000001007,0.000000291,0.000000239,-17.957,0.000,0.005,0.000001000,0.000010123,16.0,Normal\n"
	serverstatsOut := "191,222,183,111,231"
	sourcesOut :=
		`#,?,NMEA,0,4,377,16,-0.001552465,-0.001552452,0.100367047
#,*,PPS,0,4,377,16,0.000000014,0.000000027,0.000000682
^,-,194.190.168.1,15,10,377,584,-0.000704591,-0.000704880,0.002281896
`

	for _, command := range args[1:] {
		//fmt.Fprintf(os.Stderr, "command = %s\n", command)
		switch command {
		case "-m":
		case "tracking":
			fmt.Fprint(os.Stdout, trackingOut)
		case "serverstats":
			fmt.Fprint(os.Stdout, serverstatsOut)
		case "sources":
			fmt.Fprint(os.Stdout, sourcesOut)
		default:
			fmt.Printf("Unknown chronyc command %q\n", command)
			os.Exit(1)
		}
	}
	os.Exit(0)
}

func TestHelperGather(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// chronyc -c -m tracking serverstats
	fmt.Print(
		`50505300,PPS,1,1541793798.895264285,-0.000001007,0.000000291,0.000000239,-17.957,0.000,0.005,0.000001000,0.000010123,16.0,Normal
191,222,183,111,231
`)
	os.Exit(0)
}

func TestHelperGatherSources(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// chronyc -c -m tracking sources serverstats
	fmt.Print(
		`50505300,PPS,1,1541793798.895264285,-0.000001007,0.000000291,0.000000239,-17.957,0.000,0.005,0.000001000,0.000010123,16.0,Normal
#,?,NMEA,0,4,377,16,-0.001552465,-0.001552452,0.100367047
#,*,PPS,0,4,377,16,0.000000014,0.000000027,0.000000682
^,-,194.190.168.1,15,10,377,584,-0.000704591,-0.000704880,0.002281896
191,222,183,111,231
`)
	os.Exit(0)
}
