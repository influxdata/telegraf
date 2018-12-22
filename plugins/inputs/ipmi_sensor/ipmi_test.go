package ipmi_sensor

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGather(t *testing.T) {
	i := &Ipmi{
		Servers:   []string{"USERID:PASSW0RD@lan(192.168.1.1)"},
		Path:      "ipmitool",
		Privilege: "USER",
		Timeout:   internal.Duration{Duration: time.Second * 5},
	}
	// overwriting exec commands with mock commands
	execCommand = fakeExecCommand
	var acc testutil.Accumulator

	err := acc.GatherError(i.Gather)

	require.NoError(t, err)

	assert.Equal(t, acc.NFields(), 262, "non-numeric measurements should be ignored")

	conn := NewConnection(i.Servers[0], i.Privilege)
	assert.Equal(t, "USERID", conn.Username)
	assert.Equal(t, "lan", conn.Interface)

	var testsWithServer = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"value":  float64(20),
				"status": int(1),
			},
			map[string]string{
				"name":   "ambient_temp",
				"server": "192.168.1.1",
				"unit":   "degrees_c",
			},
		},
		{
			map[string]interface{}{
				"value":  float64(80),
				"status": int(1),
			},
			map[string]string{
				"name":   "altitude",
				"server": "192.168.1.1",
				"unit":   "feet",
			},
		},
		{
			map[string]interface{}{
				"value":  float64(210),
				"status": int(1),
			},
			map[string]string{
				"name":   "avg_power",
				"server": "192.168.1.1",
				"unit":   "watts",
			},
		},
		{
			map[string]interface{}{
				"value":  float64(4.9),
				"status": int(1),
			},
			map[string]string{
				"name":   "planar_5v",
				"server": "192.168.1.1",
				"unit":   "volts",
			},
		},
		{
			map[string]interface{}{
				"value":  float64(3.05),
				"status": int(1),
			},
			map[string]string{
				"name":   "planar_vbat",
				"server": "192.168.1.1",
				"unit":   "volts",
			},
		},
		{
			map[string]interface{}{
				"value":  float64(2610),
				"status": int(1),
			},
			map[string]string{
				"name":   "fan_1a_tach",
				"server": "192.168.1.1",
				"unit":   "rpm",
			},
		},
		{
			map[string]interface{}{
				"value":  float64(1775),
				"status": int(1),
			},
			map[string]string{
				"name":   "fan_1b_tach",
				"server": "192.168.1.1",
				"unit":   "rpm",
			},
		},
	}

	for _, test := range testsWithServer {
		acc.AssertContainsTaggedFields(t, "ipmi_sensor", test.fields, test.tags)
	}

	i = &Ipmi{
		Path:    "ipmitool",
		Timeout: internal.Duration{Duration: time.Second * 5},
	}

	err = acc.GatherError(i.Gather)
	require.NoError(t, err)

	var testsWithoutServer = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"value":  float64(20),
				"status": int(1),
			},
			map[string]string{
				"name": "ambient_temp",
				"unit": "degrees_c",
			},
		},
		{
			map[string]interface{}{
				"value":  float64(80),
				"status": int(1),
			},
			map[string]string{
				"name": "altitude",
				"unit": "feet",
			},
		},
		{
			map[string]interface{}{
				"value":  float64(210),
				"status": int(1),
			},
			map[string]string{
				"name": "avg_power",
				"unit": "watts",
			},
		},
		{
			map[string]interface{}{
				"value":  float64(4.9),
				"status": int(1),
			},
			map[string]string{
				"name": "planar_5v",
				"unit": "volts",
			},
		},
		{
			map[string]interface{}{
				"value":  float64(3.05),
				"status": int(1),
			},
			map[string]string{
				"name": "planar_vbat",
				"unit": "volts",
			},
		},
		{
			map[string]interface{}{
				"value":  float64(2610),
				"status": int(1),
			},
			map[string]string{
				"name": "fan_1a_tach",
				"unit": "rpm",
			},
		},
		{
			map[string]interface{}{
				"value":  float64(1775),
				"status": int(1),
			},
			map[string]string{
				"name": "fan_1b_tach",
				"unit": "rpm",
			},
		},
	}

	for _, test := range testsWithoutServer {
		acc.AssertContainsTaggedFields(t, "ipmi_sensor", test.fields, test.tags)
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

	mockData := `Ambient Temp     | 20 degrees C      | ok
Altitude         | 80 feet           | ok
Avg Power        | 210 Watts         | ok
Planar 3.3V      | 3.29 Volts        | ok
Planar 5V        | 4.90 Volts        | ok
Planar 12V       | 12.04 Volts       | ok
Planar VBAT      | 3.05 Volts        | ok
Fan 1A Tach      | 2610 RPM          | ok
Fan 1B Tach      | 1775 RPM          | ok
Fan 2A Tach      | 2001 RPM          | ok
Fan 2B Tach      | 1275 RPM          | ok
Fan 3A Tach      | 2929 RPM          | ok
Fan 3B Tach      | 2125 RPM          | ok
Fan 1            | 0x00              | ok
Fan 2            | 0x00              | ok
Fan 3            | 0x00              | ok
Front Panel      | 0x00              | ok
Video USB        | 0x00              | ok
DASD Backplane 1 | 0x00              | ok
SAS Riser        | 0x00              | ok
PCI Riser 1      | 0x00              | ok
PCI Riser 2      | 0x00              | ok
CPU 1            | 0x00              | ok
CPU 2            | 0x00              | ok
All CPUs         | 0x00              | ok
One of The CPUs  | 0x00              | ok
IOH Temp Status  | 0x00              | ok
CPU 1 OverTemp   | 0x00              | ok
CPU 2 OverTemp   | 0x00              | ok
CPU Fault Reboot | 0x00              | ok
Aux Log          | 0x00              | ok
NMI State        | 0x00              | ok
ABR Status       | 0x00              | ok
Firmware Error   | 0x00              | ok
PCIs             | 0x00              | ok
CPUs             | 0x00              | ok
DIMMs            | 0x00              | ok
Sys Board Fault  | 0x00              | ok
Power Supply 1   | 0x00              | ok
Power Supply 2   | 0x00              | ok
PS 1 Fan Fault   | 0x00              | ok
PS 2 Fan Fault   | 0x00              | ok
VT Fault         | 0x00              | ok
Pwr Rail A Fault | 0x00              | ok
Pwr Rail B Fault | 0x00              | ok
Pwr Rail C Fault | 0x00              | ok
Pwr Rail D Fault | 0x00              | ok
Pwr Rail E Fault | 0x00              | ok
PS 1 Therm Fault | 0x00              | ok
PS 2 Therm Fault | 0x00              | ok
PS1 12V OV Fault | 0x00              | ok
PS2 12V OV Fault | 0x00              | ok
PS1 12V UV Fault | 0x00              | ok
PS2 12V UV Fault | 0x00              | ok
PS1 12V OC Fault | 0x00              | ok
PS2 12V OC Fault | 0x00              | ok
PS 1 VCO Fault   | 0x00              | ok
PS 2 VCO Fault   | 0x00              | ok
Power Unit       | 0x00              | ok
Cooling Zone 1   | 0x00              | ok
Cooling Zone 2   | 0x00              | ok
Cooling Zone 3   | 0x00              | ok
Drive 0          | 0x00              | ok
Drive 1          | 0x00              | ok
Drive 2          | 0x00              | ok
Drive 3          | 0x00              | ok
Drive 4          | 0x00              | ok
Drive 5          | 0x00              | ok
Drive 6          | 0x00              | ok
Drive 7          | 0x00              | ok
Drive 8          | 0x00              | ok
Drive 9          | 0x00              | ok
Drive 10         | 0x00              | ok
Drive 11         | 0x00              | ok
Drive 12         | 0x00              | ok
Drive 13         | 0x00              | ok
Drive 14         | 0x00              | ok
Drive 15         | 0x00              | ok
All DIMMS        | 0x00              | ok
One of the DIMMs | 0x00              | ok
DIMM 1           | 0x00              | ok
DIMM 2           | 0x00              | ok
DIMM 3           | 0x00              | ok
DIMM 4           | 0x00              | ok
DIMM 5           | 0x00              | ok
DIMM 6           | 0x00              | ok
DIMM 7           | 0x00              | ok
DIMM 8           | 0x00              | ok
DIMM 9           | 0x00              | ok
DIMM 10          | 0x00              | ok
DIMM 11          | 0x00              | ok
DIMM 12          | 0x00              | ok
DIMM 13          | 0x00              | ok
DIMM 14          | 0x00              | ok
DIMM 15          | 0x00              | ok
DIMM 16          | 0x00              | ok
DIMM 17          | 0x00              | ok
DIMM 18          | 0x00              | ok
DIMM 1 Temp      | 0x00              | ok
DIMM 2 Temp      | 0x00              | ok
DIMM 3 Temp      | 0x00              | ok
DIMM 4 Temp      | 0x00              | ok
DIMM 5 Temp      | 0x00              | ok
DIMM 6 Temp      | 0x00              | ok
DIMM 7 Temp      | 0x00              | ok
DIMM 8 Temp      | 0x00              | ok
DIMM 9 Temp      | 0x00              | ok
DIMM 10 Temp     | 0x00              | ok
DIMM 11 Temp     | 0x00              | ok
DIMM 12 Temp     | 0x00              | ok
DIMM 13 Temp     | 0x00              | ok
DIMM 14 Temp     | 0x00              | ok
DIMM 15 Temp     | 0x00              | ok
DIMM 16 Temp     | 0x00              | ok
DIMM 17 Temp     | 0x00              | ok
DIMM 18 Temp     | 0x00              | ok
PCI 1            | 0x00              | ok
PCI 2            | 0x00              | ok
PCI 3            | 0x00              | ok
PCI 4            | 0x00              | ok
All PCI Error    | 0x00              | ok
One of PCI Error | 0x00              | ok
IPMI Watchdog    | 0x00              | ok
Host Power       | 0x00              | ok
DASD Backplane 2 | 0x00              | ok
DASD Backplane 3 | Not Readable      | ns
DASD Backplane 4 | Not Readable      | ns
Backup Memory    | 0x00              | ok
Progress         | 0x00              | ok
Planar Fault     | 0x00              | ok
SEL Fullness     | 0x00              | ok
PCI 5            | 0x00              | ok
OS RealTime Mod  | 0x00              | ok
`

	args := os.Args

	// Previous arguments are tests stuff, that looks like :
	// /tmp/go-build970079519/…/_test/integration.test -test.run=TestHelperProcess --
	cmd, args := args[3], args[4:]

	if cmd == "ipmitool" {
		fmt.Fprint(os.Stdout, mockData)
	} else {
		fmt.Fprint(os.Stdout, "command not found")
		os.Exit(1)

	}
	os.Exit(0)
}

func TestGatherV2(t *testing.T) {
	i := &Ipmi{
		Servers:       []string{"USERID:PASSW0RD@lan(192.168.1.1)"},
		Path:          "ipmitool",
		Privilege:     "USER",
		Timeout:       internal.Duration{Duration: time.Second * 5},
		MetricVersion: 2,
	}
	// overwriting exec commands with mock commands
	execCommand = fakeExecCommandV2
	var acc testutil.Accumulator

	err := acc.GatherError(i.Gather)

	require.NoError(t, err)

	conn := NewConnection(i.Servers[0], i.Privilege)
	assert.Equal(t, "USERID", conn.Username)
	assert.Equal(t, "lan", conn.Interface)

	var testsWithServer = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		//SEL              | 72h | ns  |  7.1 | No Reading
		{
			map[string]interface{}{
				"value": float64(0),
			},
			map[string]string{
				"name":        "sel",
				"entity_id":   "7.1",
				"status_code": "ns",
				"status_desc": "no_reading",
				"server":      "192.168.1.1",
			},
		},
	}

	for _, test := range testsWithServer {
		acc.AssertContainsTaggedFields(t, "ipmi_sensor", test.fields, test.tags)
	}

	i = &Ipmi{
		Path:          "ipmitool",
		Timeout:       internal.Duration{Duration: time.Second * 5},
		MetricVersion: 2,
	}

	err = acc.GatherError(i.Gather)
	require.NoError(t, err)

	var testsWithoutServer = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		//SEL              | 72h | ns  |  7.1 | No Reading
		{
			map[string]interface{}{
				"value": float64(0),
			},
			map[string]string{
				"name":        "sel",
				"entity_id":   "7.1",
				"status_code": "ns",
				"status_desc": "no_reading",
			},
		},
		//Intrusion        | 73h | ok  |  7.1 |
		{
			map[string]interface{}{
				"value": float64(0),
			},
			map[string]string{
				"name":        "intrusion",
				"entity_id":   "7.1",
				"status_code": "ok",
				"status_desc": "ok",
			},
		},
		//Fan1             | 30h | ok  |  7.1 | 5040 RPM
		{
			map[string]interface{}{
				"value": float64(5040),
			},
			map[string]string{
				"name":        "fan1",
				"entity_id":   "7.1",
				"status_code": "ok",
				"unit":        "rpm",
			},
		},
		//Inlet Temp       | 04h | ok  |  7.1 | 25 degrees C
		{
			map[string]interface{}{
				"value": float64(25),
			},
			map[string]string{
				"name":        "inlet_temp",
				"entity_id":   "7.1",
				"status_code": "ok",
				"unit":        "degrees_c",
			},
		},
		//USB Cable Pres   | 50h | ok  |  7.1 | Connected
		{
			map[string]interface{}{
				"value": float64(0),
			},
			map[string]string{
				"name":        "usb_cable_pres",
				"entity_id":   "7.1",
				"status_code": "ok",
				"status_desc": "connected",
			},
		},
		//Current 1        | 6Ah | ok  | 10.1 | 7.20 Amps
		{
			map[string]interface{}{
				"value": float64(7.2),
			},
			map[string]string{
				"name":        "current_1",
				"entity_id":   "10.1",
				"status_code": "ok",
				"unit":        "amps",
			},
		},
		//Power Supply 1   | 03h | ok  | 10.1 | 110 Watts, Presence detected
		{
			map[string]interface{}{
				"value": float64(110),
			},
			map[string]string{
				"name":        "power_supply_1",
				"entity_id":   "10.1",
				"status_code": "ok",
				"unit":        "watts",
				"status_desc": "presence_detected",
			},
		},
	}

	for _, test := range testsWithoutServer {
		acc.AssertContainsTaggedFields(t, "ipmi_sensor", test.fields, test.tags)
	}
}

// fackeExecCommandV2 is a helper function that mock
// the exec.Command call (and call the test binary)
func fakeExecCommandV2(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcessV2", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// TestHelperProcessV2 isn't a real test. It's used to mock exec.Command
// For example, if you run:
// GO_WANT_HELPER_PROCESS=1 go test -test.run=TestHelperProcessV2 -- chrony tracking
// it returns below mockData.
func TestHelperProcessV2(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// Curated list of use cases instead of full dumps
	mockData := `SEL              | 72h | ns  |  7.1 | No Reading
Intrusion        | 73h | ok  |  7.1 |
Fan1             | 30h | ok  |  7.1 | 5040 RPM
Inlet Temp       | 04h | ok  |  7.1 | 25 degrees C
USB Cable Pres   | 50h | ok  |  7.1 | Connected
Current 1        | 6Ah | ok  | 10.1 | 7.20 Amps
Power Supply 1   | 03h | ok  | 10.1 | 110 Watts, Presence detected
`

	args := os.Args

	// Previous arguments are tests stuff, that looks like :
	// /tmp/go-build970079519/…/_test/integration.test -test.run=TestHelperProcess --
	cmd, args := args[3], args[4:]

	if cmd == "ipmitool" {
		fmt.Fprint(os.Stdout, mockData)
	} else {
		fmt.Fprint(os.Stdout, "command not found")
		os.Exit(1)

	}
	os.Exit(0)
}

func TestExtractFields(t *testing.T) {
	v1Data := `Ambient Temp     | 20 degrees C      | ok
Altitude         | 80 feet           | ok
Avg Power        | 210 Watts         | ok
Planar 3.3V      | 3.29 Volts        | ok
Planar 5V        | 4.90 Volts        | ok
Planar 12V       | 12.04 Volts       | ok
B                | 0x00              | ok
Unable to send command: Invalid argument
ECC Corr Err     | Not Readable      | ns
Unable to send command: Invalid argument
ECC Uncorr Err   | Not Readable      | ns
Unable to send command: Invalid argument
`

	v2Data := `SEL              | 72h | ns  |  7.1 | No Reading
Intrusion        | 73h | ok  |  7.1 |
Fan1             | 30h | ok  |  7.1 | 5040 RPM
Inlet Temp       | 04h | ok  |  7.1 | 25 degrees C
USB Cable Pres   | 50h | ok  |  7.1 | Connected
Unable to send command: Invalid argument
Current 1        | 6Ah | ok  | 10.1 | 7.20 Amps
Unable to send command: Invalid argument
Power Supply 1   | 03h | ok  | 10.1 | 110 Watts, Presence detected
`

	tests := []string{
		v1Data,
		v2Data,
	}

	for i := range tests {
		t.Logf("Checking v%d data...", i+1)
		extractFieldsFromRegex(re_v1_parse_line, tests[i])
		extractFieldsFromRegex(re_v2_parse_line, tests[i])
	}
}
