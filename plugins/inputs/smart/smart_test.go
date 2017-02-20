package smart

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mockScanData = `/dev/ada0 -d atacam # /dev/ada0, ATA device
`
	mockInfoAttributeData = `smartctl 6.5 2016-05-07 r4318 [Darwin 16.4.0 x86_64] (local build)
Copyright (C) 2002-16, Bruce Allen, Christian Franke, www.smartmontools.org

CHECK POWER MODE not implemented, ignoring -n option
=== START OF INFORMATION SECTION ===
Model Family:     Apple SD/SM/TS...E/F SSDs
Device Model:     APPLE SSD SM256E
Serial Number:    S0X5NZBC422720
LU WWN Device Id: 5 002538 043584d30
Firmware Version: CXM09A1Q
User Capacity:    251,000,193,024 bytes [251 GB]
Sector Sizes:     512 bytes logical, 4096 bytes physical
Rotation Rate:    Solid State Device
Device is:        In smartctl database [for details use: -P show]
ATA Version is:   ATA8-ACS T13/1699-D revision 4c
SATA Version is:  SATA 3.0, 6.0 Gb/s (current: 6.0 Gb/s)
Local Time is:    Thu Feb  9 16:48:45 2017 CET
SMART support is: Available - device has SMART capability.
SMART support is: Enabled

=== START OF READ SMART DATA SECTION ===
SMART Attributes Data Structure revision number: 1
Vendor Specific SMART Attributes with Thresholds:
ID# ATTRIBUTE_NAME          FLAGS    VALUE WORST THRESH FAIL RAW_VALUE
  1 Raw_Read_Error_Rate     -O-RC-   200   200   000    -    0
  5 Reallocated_Sector_Ct   PO--CK   100   100   000    -    0
  9 Power_On_Hours          -O--CK   099   099   000    -    2988
 12 Power_Cycle_Count       -O--CK   085   085   000    -    14879
169 Unknown_Attribute       PO--C-   253   253   010    -    2044932921600
173 Wear_Leveling_Count     -O--CK   185   185   100    -    957808640337
190 Airflow_Temperature_Cel -O---K   055   040   045    Past 45 (Min/Max 43/57 #2689)
192 Power-Off_Retract_Count -O--C-   097   097   000    -    14716
194 Temperature_Celsius     -O---K   066   021   000    -    34 (Min/Max 14/79)
197 Current_Pending_Sector  -O---K   100   100   000    -    0
199 UDMA_CRC_Error_Count    -O-RC-   200   200   000    -    0
240 Head_Flying_Hours       ------   100   253   000    -    6585h+55m+23.234s
                            ||||||_ K auto-keep
                            |||||__ C event count
                            ||||___ R error rate
                            |||____ S speed/performance
                            ||_____ O updated online
                            |______ P prefailure warning
`
)

func TestGather(t *testing.T) {
	s := &Smart{
		path: "smartctl",
	}
	// overwriting exec commands with mock commands
	execCommand = fakeExecCommand
	var acc testutil.Accumulator

	err := s.Gather(&acc)

	require.NoError(t, err)
	assert.Equal(t, 48, acc.NFields(), "Wrong number of fields gathered")

	device_tags := map[string]string{
		"device":       "/dev/ada0",
		"device_model": "APPLE SSD SM256E",
		"serial_no":    "S0X5NZBC422720",
		"enabled":      "Enabled",
		"capacity":     "251000193024",
	}

	var testsAda0Device = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"value":     int(200),
				"worst":     int(200),
				"threshold": int(0),
				"raw_value": int(0),
			},
			map[string]string{
				"id":    "1",
				"name":  "Raw_Read_Error_Rate",
				"flags": "-O-RC-",
				"fail":  "-",
			},
		},
		{
			map[string]interface{}{
				"value":     int(100),
				"worst":     int(100),
				"threshold": int(0),
				"raw_value": int(0),
			},
			map[string]string{
				"id":    "5",
				"name":  "Reallocated_Sector_Ct",
				"flags": "PO--CK",
				"fail":  "-",
			},
		},
		{
			map[string]interface{}{
				"value":     int(99),
				"worst":     int(99),
				"threshold": int(0),
				"raw_value": int(2988),
			},
			map[string]string{
				"id":    "9",
				"name":  "Power_On_Hours",
				"flags": "-O--CK",
				"fail":  "-",
			},
		},
		{
			map[string]interface{}{
				"value":     int(85),
				"worst":     int(85),
				"threshold": int(0),
				"raw_value": int(14879),
			},
			map[string]string{
				"id":    "12",
				"name":  "Power_Cycle_Count",
				"flags": "-O--CK",
				"fail":  "-",
			},
		},
		{
			map[string]interface{}{
				"value":     int(253),
				"worst":     int(253),
				"threshold": int(10),
				"raw_value": int(2044932921600),
			},
			map[string]string{
				"id":    "169",
				"name":  "Unknown_Attribute",
				"flags": "PO--C-",
				"fail":  "-",
			},
		},
		{
			map[string]interface{}{
				"value":     int(185),
				"worst":     int(185),
				"threshold": int(100),
				"raw_value": int(957808640337),
			},
			map[string]string{
				"id":    "173",
				"name":  "Wear_Leveling_Count",
				"flags": "-O--CK",
				"fail":  "-",
			},
		},
		{
			map[string]interface{}{
				"value":     int(55),
				"worst":     int(40),
				"threshold": int(45),
				"raw_value": int(45),
			},
			map[string]string{
				"id":    "190",
				"name":  "Airflow_Temperature_Cel",
				"flags": "-O---K",
				"fail":  "Past",
			},
		},
		{
			map[string]interface{}{
				"value":     int(97),
				"worst":     int(97),
				"threshold": int(0),
				"raw_value": int(14716),
			},
			map[string]string{
				"id":    "192",
				"name":  "Power-Off_Retract_Count",
				"flags": "-O--C-",
				"fail":  "-",
			},
		},
		{
			map[string]interface{}{
				"value":     int(66),
				"worst":     int(21),
				"threshold": int(0),
				"raw_value": int(34),
			},
			map[string]string{
				"id":    "194",
				"name":  "Temperature_Celsius",
				"flags": "-O---K",
				"fail":  "-",
			},
		},
		{
			map[string]interface{}{
				"value":     int(100),
				"worst":     int(100),
				"threshold": int(0),
				"raw_value": int(0),
			},
			map[string]string{
				"id":    "197",
				"name":  "Current_Pending_Sector",
				"flags": "-O---K",
				"fail":  "-",
			},
		},
		{
			map[string]interface{}{
				"value":     int(200),
				"worst":     int(200),
				"threshold": int(0),
				"raw_value": int(0),
			},
			map[string]string{
				"id":    "199",
				"name":  "UDMA_CRC_Error_Count",
				"flags": "-O-RC-",
				"fail":  "-",
			},
		},
		{
			map[string]interface{}{
				"value":     int(100),
				"worst":     int(253),
				"threshold": int(0),
				"raw_value": int(23709323),
			},
			map[string]string{
				"id":    "240",
				"name":  "Head_Flying_Hours",
				"flags": "------",
				"fail":  "-",
			},
		},
	}

	for _, test := range testsAda0Device {
		for k, v := range device_tags {
			test.tags[k] = v
		}

		acc.AssertContainsTaggedFields(t, "smart", test.fields, test.tags)
	}
}

func TestExcludedDev(t *testing.T) {
	assert.Equal(t, true, excludedDev([]string{"/dev/pass6"}, "/dev/pass6"), "Should be excluded.")
	assert.Equal(t, false, excludedDev([]string{}, "/dev/pass6"), "Shouldn't be excluded.")
	assert.Equal(t, false, excludedDev([]string{"/dev/pass6"}, "/dev/pass1"), "Shouldn't be excluded.")

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
// GO_WANT_HELPER_PROCESS=1 go test -test.run=TestHelperProcess -- --scan
// it returns below mockScanData.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args

	// Previous arguments are tests stuff, that looks like :
	// /tmp/go-build970079519/…/_test/integration.test -test.run=TestHelperProcess --
	cmd, arg1, args := args[3], args[4], args[5:]

	if cmd == "smartctl" {
		if arg1 == "--scan" {
			fmt.Fprint(os.Stdout, mockScanData)
		}
		if arg1 == "--info" {
			fmt.Fprint(os.Stdout, mockInfoAttributeData)
		}
	} else {
		fmt.Fprint(os.Stdout, "command not found")
		os.Exit(1)
	}
	os.Exit(0)
}
