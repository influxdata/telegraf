// +build linux

package sensors

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func TestGather(t *testing.T) {
	s := Sensors{
		path: "sensors",
	}
	// overwriting exec commands with mock commands
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	var acc testutil.Accumulator

	err := s.Gather(&acc)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		tags   map[string]string
		fields map[string]interface{}
	}{
		{
			map[string]string{
				"chip":    "coretemp-isa-0000",
				"feature": "Core 0",
			},
			map[string]interface{}{
				"reading": 45.00,
				"unit":    "C",
				"high":    99.00,
			},
		},
		{
			map[string]string{
				"chip":    "coretemp-isa-0001",
				"feature": "Core 1",
			},
			map[string]interface{}{
				"reading": 44.00,
				"unit":    "C",
				"high":    99.00,
			},
		},
		{
			map[string]string{
				"chip":    "coretemp-isa-0002",
				"feature": "Core 2",
			},
			map[string]interface{}{
				"reading": 45.00,
				"unit":    "C",
				"high":    99.00,
			},
		},
		{
			map[string]string{
				"chip":    "coretemp-isa-0003",
				"feature": "Core 3",
			},
			map[string]interface{}{
				"reading": 43.00,
				"unit":    "C",
				"high":    99.00,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "VCore",
			},
			map[string]interface{}{
				"reading": 0.90,
				"unit":    "V",
				"min":     0.60,
				"max":     1.49,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "in1",
			},
			map[string]interface{}{
				"reading": 11.88,
				"unit":    "V",
				"min":     10.72,
				"max":     13.15,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "AVCC",
			},
			map[string]interface{}{
				"reading": 3.30,
				"unit":    "V",
				"min":     2.96,
				"max":     3.63,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "3VCC",
			},
			map[string]interface{}{
				"reading": 3.30,
				"unit":    "V",
				"min":     2.96,
				"max":     3.63,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "in4",
			},
			map[string]interface{}{
				"reading": 1.54,
				"unit":    "V",
				"min":     1.35,
				"max":     1.65,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "in5",
			},
			map[string]interface{}{
				"reading": 1.26,
				"unit":    "V",
				"min":     1.13,
				"max":     1.38,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "in6",
			},
			map[string]interface{}{
				"reading": 4.66,
				"unit":    "V",
				"min":     4.53,
				"max":     4.86,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "VSB",
			},
			map[string]interface{}{
				"reading": 3.30,
				"unit":    "V",
				"min":     2.96,
				"max":     3.63,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "VBAT",
			},
			map[string]interface{}{
				"reading": 3.20,
				"unit":    "V",
				"min":     2.96,
				"max":     3.63,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "Case Fan",
			},
			map[string]interface{}{
				"reading": 0.00,
				"unit":    "RPM",
				"min":     753.00,
				"div":     128.00,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "CPU Fan",
			},
			map[string]interface{}{
				"reading": 3835.00,
				"unit":    "RPM",
				"min":     712.00,
				"div":     8.00,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "Aux Fan",
			},
			map[string]interface{}{
				"reading": 0.00,
				"unit":    "RPM",
				"min":     753.00,
				"div":     128.00,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "fan4",
			},
			map[string]interface{}{
				"reading": 0.00,
				"unit":    "RPM",
				"min":     753.00,
				"div":     128.00,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "fan5",
			},
			map[string]interface{}{
				"reading": 0.00,
				"unit":    "RPM",
				"min":     753.00,
				"div":     128.00,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "Sys Temp",
			},
			map[string]interface{}{
				"reading": 48.00,
				"unit":    "C",
				"high":    60.00,
				"hyst":    55.00,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "CPU Temp",
			},
			map[string]interface{}{
				"reading": 46.00,
				"unit":    "C",
				"high":    95.00,
				"hyst":    92.00,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "AUX Temp",
			},
			map[string]interface{}{
				"reading": 46.00,
				"unit":    "C",
				"high":    80.00,
				"hyst":    75.00,
			},
		},
		{
			map[string]string{
				"chip":    "w83627dhg-isa-0a10",
				"feature": "vid",
			},
			map[string]interface{}{
				"reading": 1.300,
				"unit":    "V",
			},
		},
	}

	for _, test := range tests {
		acc.AssertContainsTaggedFields(t, "sensors", test.fields, test.tags)
	}
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

	mockData := `
coretemp-isa-0000
Core 0:      +45°C  (high =   +99°C)

coretemp-isa-0001
Core 1:      +44°C  (high =   +99°C)

coretemp-isa-0002
Core 2:      +45°C  (high =   +99°C)

coretemp-isa-0003
Core 3:      +43°C  (high =   +99°C)

w83627dhg-isa-0a10
VCore:     +0.90 V  (min =  +0.60 V, max =  +1.49 V)
in1:      +11.88 V  (min = +10.72 V, max = +13.15 V)
AVCC:      +3.30 V  (min =  +2.96 V, max =  +3.63 V)
3VCC:      +3.30 V  (min =  +2.96 V, max =  +3.63 V)
in4:       +1.54 V  (min =  +1.35 V, max =  +1.65 V)
in5:       +1.26 V  (min =  +1.13 V, max =  +1.38 V)
in6:       +4.66 V  (min =  +4.53 V, max =  +4.86 V)
VSB:       +3.30 V  (min =  +2.96 V, max =  +3.63 V)
VBAT:      +3.20 V  (min =  +2.96 V, max =  +3.63 V)
Case Fan:    0 RPM  (min =  753 RPM, div = 128) ALARM
CPU Fan:  3835 RPM  (min =  712 RPM, div = 8)
Aux Fan:     0 RPM  (min =  753 RPM, div = 128) ALARM
fan4:        0 RPM  (min =  753 RPM, div = 128) ALARM
fan5:        0 RPM  (min =  753 RPM, div = 128) ALARM
Sys Temp:    +48°C  (high =   +60°C, hyst =   +55°C)  [thermistor]
CPU Temp:  +46.0°C  (high = +95.0°C, hyst = +92.0°C)  [CPU diode ]
AUX Temp:  +46.0°C  (high = +80.0°C, hyst = +75.0°C)  [CPU diode ]
vid:      +1.300 V
`

	args := os.Args

	// Previous arguments are tests stuff, that looks like :
	// /tmp/go-build970079519/…/_test/integration.test -test.run=TestHelperProcess --
	cmd, args := args[3], args[4:]

	if cmd == "sensors" {
		fmt.Fprint(os.Stdout, mockData)
	} else {
		fmt.Fprint(os.Stdout, "command not found")
		os.Exit(1)

	}
	os.Exit(0)
}
