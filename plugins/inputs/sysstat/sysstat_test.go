// +build linux

package sysstat

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

var s = Sysstat{
	interval:   10,
	Sadc:       "/usr/lib/sa/sadc",
	Sadf:       "/usr/bin/sadf",
	Group:      false,
	Activities: []string{"DISK", "SNMP"},
	Options: map[string]string{
		"C": "cpu",
		"d": "disk",
	},
	DeviceTags: map[string][]map[string]string{
		"sda": {
			{
				"vg": "rootvg",
			},
		},
	},
}

func TestGather(t *testing.T) {
	// overwriting exec commands with mock commands
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	var acc testutil.Accumulator

	err := acc.GatherError(s.Gather)
	if err != nil {
		t.Fatal(err)
	}

	cpuTags := map[string]string{"device": "all"}
	diskTags := map[string]string{"device": "sda", "vg": "rootvg"}
	tests := []struct {
		measurement string
		fields      map[string]interface{}
		tags        map[string]string
	}{
		{
			"cpu_pct_user",
			map[string]interface{}{
				"value": 0.65,
			},
			cpuTags,
		},
		{
			"cpu_pct_nice",
			map[string]interface{}{
				"value": 0.0,
			},
			cpuTags,
		},
		{
			"cpu_pct_system",
			map[string]interface{}{
				"value": 0.10,
			},
			cpuTags,
		},
		{
			"cpu_pct_iowait",
			map[string]interface{}{
				"value": 0.15,
			},
			cpuTags,
		},
		{
			"cpu_pct_steal",
			map[string]interface{}{
				"value": 0.0,
			},
			cpuTags,
		},
		{
			"cpu_pct_idle",
			map[string]interface{}{
				"value": 99.1,
			},
			cpuTags,
		},
		{
			"disk_tps",
			map[string]interface{}{
				"value": 0.00,
			},
			diskTags,
		},
		{
			"disk_rd_sec_per_s",
			map[string]interface{}{
				"value": 0.00,
			},
			diskTags,
		},
		{
			"disk_wr_sec_per_s",
			map[string]interface{}{
				"value": 0.00,
			},
			diskTags,
		},
		{
			"disk_avgrq-sz",
			map[string]interface{}{
				"value": 0.00,
			},
			diskTags,
		},
		{
			"disk_avgqu-sz",
			map[string]interface{}{
				"value": 0.00,
			},
			diskTags,
		},
		{
			"disk_await",
			map[string]interface{}{
				"value": 0.00,
			},
			diskTags,
		},
		{
			"disk_svctm",
			map[string]interface{}{
				"value": 0.00,
			},
			diskTags,
		},
		{
			"disk_pct_util",
			map[string]interface{}{
				"value": 0.00,
			},
			diskTags,
		},
	}
	for _, test := range tests {
		acc.AssertContainsTaggedFields(t, test.measurement, test.fields, test.tags)
	}
}

func TestGatherGrouped(t *testing.T) {
	s.Group = true
	// overwriting exec commands with mock commands
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	var acc testutil.Accumulator

	err := acc.GatherError(s.Gather)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		measurement string
		fields      map[string]interface{}
		tags        map[string]string
	}{
		{
			"cpu",
			map[string]interface{}{
				"pct_user":   0.65,
				"pct_nice":   0.0,
				"pct_system": 0.10,
				"pct_iowait": 0.15,
				"pct_steal":  0.0,
				"pct_idle":   99.1,
			},
			map[string]string{"device": "all"},
		},
		{
			"disk",
			map[string]interface{}{
				"tps":          0.00,
				"rd_sec_per_s": 0.00,
				"wr_sec_per_s": 0.00,
				"avgrq-sz":     0.00,
				"avgqu-sz":     0.00,
				"await":        0.00,
				"svctm":        0.00,
				"pct_util":     0.00,
			},
			map[string]string{"device": "sda", "vg": "rootvg"},
		},
		{
			"disk",
			map[string]interface{}{
				"tps":          2.01,
				"rd_sec_per_s": 1.0,
				"wr_sec_per_s": 0.00,
				"avgrq-sz":     0.30,
				"avgqu-sz":     0.60,
				"await":        0.70,
				"svctm":        0.20,
				"pct_util":     0.30,
			},
			map[string]string{"device": "sdb"},
		},
	}
	for _, test := range tests {
		acc.AssertContainsTaggedFields(t, test.measurement, test.fields, test.tags)
	}
}

func TestEscape(t *testing.T) {
	var tests = []struct {
		input   string
		escaped string
	}{
		{
			"%util",
			"pct_util",
		},
		{
			"bread/s",
			"bread_per_s",
		},
		{
			"%nice",
			"pct_nice",
		},
	}

	for _, test := range tests {
		if test.escaped != escape(test.input) {
			t.Errorf("wrong escape, got %s, wanted %s", escape(test.input), test.escaped)
		}
	}
}

// Helper function that mock the exec.Command call (and call the test binary)
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
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

		"C": `dell-xps	5	2016-03-25 16:18:10 UTC	all	%user	0.65
dell-xps	5	2016-03-25 16:18:10 UTC	all	%nice	0.00
dell-xps	5	2016-03-25 16:18:10 UTC	all	%system	0.10
dell-xps	5	2016-03-25 16:18:10 UTC	all	%iowait	0.15
dell-xps	5	2016-03-25 16:18:10 UTC	all	%steal	0.00
dell-xps	5	2016-03-25 16:18:10 UTC	all	%idle	99.10
`,

		"d": `dell-xps	5	2016-03-25 16:18:10 UTC	sda	tps	0.00
dell-xps	5	2016-03-25 16:18:10 UTC	sda	rd_sec/s	0.00
dell-xps	5	2016-03-25 16:18:10 UTC	sda	wr_sec/s	0.00
dell-xps	5	2016-03-25 16:18:10 UTC	sda	avgrq-sz	0.00
dell-xps	5	2016-03-25 16:18:10 UTC	sda	avgqu-sz	0.00
dell-xps	5	2016-03-25 16:18:10 UTC	sda	await	0.00
dell-xps	5	2016-03-25 16:18:10 UTC	sda	svctm	0.00
dell-xps	5	2016-03-25 16:18:10 UTC	sda	%util	0.00
dell-xps	5	2016-03-25 16:18:10 UTC	sdb	tps	2.01
dell-xps	5	2016-03-25 16:18:10 UTC	sdb	rd_sec/s	1.00
dell-xps	5	2016-03-25 16:18:10 UTC	sdb	wr_sec/s	0.00
dell-xps	5	2016-03-25 16:18:10 UTC	sdb	avgrq-sz	0.30
dell-xps	5	2016-03-25 16:18:10 UTC	sdb	avgqu-sz	0.60
dell-xps	5	2016-03-25 16:18:10 UTC	sdb	await	0.70
dell-xps	5	2016-03-25 16:18:10 UTC	sdb	svctm	0.20
dell-xps	5	2016-03-25 16:18:10 UTC	sdb	%util	0.30
`,
	}

	args := os.Args

	// Previous arguments are tests stuff, that looks like :
	// /tmp/go-build970079519/…/_test/integration.test -test.run=TestHelperProcess --
	cmd, args := args[3], args[4:]
	// Handle the case where args[0] is dir:...

	switch path.Base(cmd) {
	case "sadf":
		fmt.Fprint(os.Stdout, mockData[args[3]])
	default:
	}
	// some code here to check arguments perhaps?
	os.Exit(0)
}
