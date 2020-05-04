package smc_ipmi

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
	s := &Smcipmi{
		Path:     "SMCIPMITool",
		Servers:  []string{"USERID:PASSW0RD@(192.168.1.1)"},
		TempUnit: "C",
		Timeout:  internal.Duration{Duration: time.Second * 5},
	}
	// overwriting exec commands with mock commands
	execCommand = fakeExecCommand
	execCommandPminfo = fakeExecCommandPminfo
	var acc testutil.Accumulator

	err := acc.GatherError(s.Gather)

	require.NoError(t, err)

	assert.Equal(t, 95, acc.NFields(), "non-numeric measurements should be ignored")

	conn := NewConnection(s.Servers[0])
	assert.Equal(t, "USERID", conn.Username)

	var tests = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(47),
			},
			map[string]string{
				"name":   "cpu1_temp",
				"server": "192.168.1.1",
				"unit":   "c",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(60),
			},
			map[string]string{
				"name":   "cpu2_temp",
				"server": "192.168.1.1",
				"unit":   "c",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(56),
			},
			map[string]string{
				"name":   "pch_temp",
				"server": "192.168.1.1",
				"unit":   "c",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(39),
			},
			map[string]string{
				"name":   "system_temp",
				"server": "192.168.1.1",
				"unit":   "c",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(61),
			},
			map[string]string{
				"name":   "peripheral_temp",
				"server": "192.168.1.1",
				"unit":   "c",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(44),
			},
			map[string]string{
				"name":   "vcpu1vrm_temp",
				"server": "192.168.1.1",
				"unit":   "c",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(37),
			},
			map[string]string{
				"name":   "vmemabvrm_temp",
				"server": "192.168.1.1",
				"unit":   "c",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(37),
			},
			map[string]string{
				"name":   "p1-dimma1_temp",
				"server": "192.168.1.1",
				"unit":   "c",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(3600),
			},
			map[string]string{
				"name":   "fan1",
				"server": "192.168.1.1",
				"unit":   "rpm",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(12),
			},
			map[string]string{
				"name":   "12v",
				"server": "192.168.1.1",
				"unit":   "v",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(3.38),
			},
			map[string]string{
				"name":   "3.3vcc",
				"server": "192.168.1.1",
				"unit":   "v",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(2.89),
			},
			map[string]string{
				"name":   "vbat",
				"server": "192.168.1.1",
				"unit":   "v",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
				"value":  float64(1.51),
			},
			map[string]string{
				"name":   "1.5v_pch",
				"server": "192.168.1.1",
				"unit":   "v",
			},
		},
		{
			map[string]interface{}{
				"status": int(1),
			},
			map[string]string{
				"name":   "pmbus_status",
				"server": "192.168.1.1",
			},
		},
		{
			map[string]interface{}{
				"value": float64(120.5),
			},
			map[string]string{
				"name":   "pmbus_input_voltage",
				"server": "192.168.1.1",
				"unit":   "v",
			},
		},
		{
			map[string]interface{}{
				"value": float64(0.98),
			},
			map[string]string{
				"name":   "pmbus_input_current",
				"server": "192.168.1.1",
				"unit":   "a",
			},
		},
		{
			map[string]interface{}{
				"value": float64(39),
			},
			map[string]string{
				"name":   "pmbus_temperature_1",
				"server": "192.168.1.1",
				"unit":   "c",
			},
		},
		{
			map[string]interface{}{
				"value": float64(6304),
			},
			map[string]string{
				"name":   "pmbus_fan_1",
				"server": "192.168.1.1",
				"unit":   "rpm",
			},
		},
	}

	for _, test := range tests {
		acc.AssertContainsTaggedFields(t, "smc_ipmi", test.fields, test.tags)
	}
}

// fakeExecCommand is a helper function that mocks
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

	mockData := `Getting SDR records ...
  Status | (#)Sensor                |      Reading | Low Limit | High Limit |
  ------ | ---------                |      ------- | --------- | ---------- |
  OK     | (4) CPU1 Temp            |     47C/117F |    0C/32F |   90C/194F |
  OK     | (71) CPU2 Temp           |     60C/140F |    0C/32F |   90C/194F |
  OK     | (138) PCH Temp           |     56C/133F |   -8C/18F |   95C/203F |
  OK     | (205) System Temp        |     39C/102F |   -7C/19F |   85C/185F |
  OK     | (272) Peripheral Temp    |     61C/142F |   -7C/19F |   85C/185F |
  OK     | (339) Vcpu1VRM Temp      |     44C/111F |   -7C/19F |  100C/212F |
  OK     | (406) Vcpu2VRM Temp      |     60C/140F |   -7C/19F |  100C/212F |
  OK     | (473) VmemABVRM Temp     |      37C/99F |   -7C/19F |  100C/212F |
  OK     | (540) VmemCDVRM Temp     |     44C/111F |   -7C/19F |  100C/212F |
  OK     | (607) VmemEFVRM Temp     |     58C/136F |   -7C/19F |  100C/212F |
  OK     | (674) VmemGHVRM Temp     |     51C/124F |   -7C/19F |  100C/212F |
  OK     | (741) P1-DIMMA1 Temp     |      37C/99F |    2C/36F |   85C/185F |
  OK     | (808) P1-DIMMB1 Temp     |      37C/99F |    2C/36F |   85C/185F |
  OK     | (875) P1-DIMMC1 Temp     |     44C/111F |    2C/36F |   85C/185F |
  OK     | (942) P1-DIMMD1 Temp     |     41C/106F |    2C/36F |   85C/185F |
  OK     | (1009) P2-DIMME1 Temp    |     48C/118F |    2C/36F |   85C/185F |
  OK     | (1076) P2-DIMMF1 Temp    |     48C/118F |    2C/36F |   85C/185F |
  OK     | (1143) P2-DIMMG1 Temp    |     43C/109F |    2C/36F |   85C/185F |
  OK     | (1210) P2-DIMMH1 Temp    |     44C/111F |    2C/36F |   85C/185F |
  OK     | (1277) FAN1              |     3600 RPM |   500 RPM |  25400 RPM |
  OK     | (1344) FAN2              |     3600 RPM |   500 RPM |  25400 RPM |
  OK     | (1411) FAN3              |     3700 RPM |   500 RPM |  25400 RPM |
  OK     | (1478) FAN4              |     3700 RPM |   500 RPM |  25400 RPM |
         | (1545) FAN5              |          N/A |   500 RPM |  25400 RPM |
         | (1612) FAN6              |          N/A |   500 RPM |  25400 RPM |
         | (1679) FANA              |          N/A |   500 RPM |  25400 RPM |
         | (1746) FANB              |          N/A |   500 RPM |  25400 RPM |
  OK     | (1813) 12V               |       12.0 V |   10.29 V |    13.26 V |
  OK     | (1880) 5VCC              |        5.0 V |    4.29 V |     5.54 V |
  OK     | (1947) 3.3VCC            |       3.38 V |    2.82 V |     3.65 V |
  OK     | (2014) VBAT              |       2.89 V |    2.43 V |     3.78 V |
  OK     | (2081) Vcpu1             |        1.8 V |    1.26 V |     2.08 V |
  OK     | (2148) Vcpu2             |        1.8 V |    1.26 V |     2.08 V |
  OK     | (2215) VDIMMAB           |        1.2 V |    0.97 V |     1.42 V |
  OK     | (2282) VDIMMCD           |        1.2 V |    0.97 V |     1.42 V |
  OK     | (2349) VDIMMEF           |       1.21 V |    0.97 V |     1.42 V |
  OK     | (2416) VDIMMGH           |        1.2 V |    0.97 V |     1.42 V |
  OK     | (2483) 5VSB              |       4.94 V |    4.29 V |     5.54 V |
  OK     | (2550) 3.3VSB            |       3.24 V |    2.82 V |     3.65 V |
  OK     | (2617) 1.5V PCH          |       1.51 V |    1.34 V |     1.67 V |
  OK     | (2684) 1.2V BMC          |        1.2 V |    1.04 V |     1.37 V |
  OK     | (2751) 1.05V PCH         |       1.05 V |    0.89 V |     1.22 V |
  OK     | (2818) Chassis Intru     |                  OK                   |
  OK     | (3354) PS1 Status        |           Presence detected           |
`

	args := os.Args

	// Previous arguments are tests stuff, that looks like :
	// /tmp/go-build970079519/…/_test/integration.test -test.run=TestHelperProcess --
	cmd, args := args[3], args[4:]

	if cmd == "SMCIPMITool" {
		fmt.Fprint(os.Stdout, mockData)
	} else {
		fmt.Fprint(os.Stdout, "command not found")
		os.Exit(1)

	}
	os.Exit(0)
}

// fakeExecCommandPminfo is a helper function that mocks
// the exec.Command call (and call the test binary)
func fakeExecCommandPminfo(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcessPminfo", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// TestHelperProcessPminfo isn't a real test. It's used to mock exec.Command
// For example, if you run:
// GO_WANT_HELPER_PROCESS=1 go test -test.run=TestHelperProcessPminfo -- chrony tracking
// it returns below mockData.
func TestHelperProcessPminfo(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	mockData := ` [SlaveAddress = 78h] [Module 1]
 Item                           |                          Value
 ----                           |                          -----
 Status                         |               [STATUS OK](00h)
 Input Voltage                  |                        120.5 V
 Input Current                  |                         0.98 A
 Main Output Voltage            |                        12.09 V
 Main Output Current            |                         8.37 A
 Temperature 1                  |                       39C/102F
 Temperature 2                  |                       43C/109F
 Fan 1                          |                       6304 RPM
 Fan 2                          |                          0 RPM
 Main Output Power              |                          101 W
 Input Power                    |                          118 W
 PMBus Revision                 |                           0x22
 PWS Serial Number              |                P000XXX00XX0000
 PWS Module Number              |                    PWS-406P-1R
 PWS Revision                   |                         REV1.3
 `

	args := os.Args

	// Previous arguments are tests stuff, that looks like :
	// /tmp/go-build970079519/…/_test/integration.test -test.run=TestHelperProcess --
	cmd, args := args[3], args[4:]

	if cmd == "SMCIPMITool" {
		fmt.Fprint(os.Stdout, mockData)
	} else {
		fmt.Fprint(os.Stdout, "command not found")
		os.Exit(1)

	}
	os.Exit(0)
}

func TestToTemp(t *testing.T) {
	value := "47C/117F"
	temp := toTemp(value, "F")
	if temp != float64(117) {
		t.Errorf("Temp incorrect, got: %f, wanted :%f", temp, float64(117))
	}
}
