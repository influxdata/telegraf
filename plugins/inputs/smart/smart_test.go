package smart

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherAttributes(t *testing.T) {
	s := NewSmart()
	s.Attributes = true

	assert.Equal(t, time.Second*30, s.Timeout.Duration)

	runCmd = func(timeout internal.Duration, sudo bool, command string, args ...string) ([]byte, error) {
		if len(args) > 0 {
			if args[0] == "--info" && args[7] == "/dev/ada0" {
				return []byte(mockInfoAttributeData), nil
			} else if args[0] == "--info" && args[7] == "/dev/nvme0" {
				return []byte(smartctlNvmeInfoData), nil
			} else if args[0] == "--scan" && len(args) == 1 {
				return []byte(mockScanData), nil
			} else if args[0] == "--scan" && len(args) >= 2 && args[1] == "--device=nvme" {
				return []byte(mockScanNvmeData), nil
			}
		}
		return nil, errors.New("command not found")
	}

	t.Run("Wrong path to smartctl", func(t *testing.T) {
		s.PathSmartctl = "this_path_to_smartctl_does_not_exist"
		err := s.Init()

		assert.Error(t, err)
	})

	t.Run("Smartctl presence", func(t *testing.T) {
		s.PathSmartctl = "smartctl"
		s.PathNVMe = ""

		t.Run("Only non nvme device", func(t *testing.T) {
			s.Devices = []string{"/dev/ada0"}
			var acc testutil.Accumulator

			err := s.Gather(&acc)

			require.NoError(t, err)
			assert.Equal(t, 65, acc.NFields(), "Wrong number of fields gathered")

			for _, test := range testsAda0Attributes {
				acc.AssertContainsTaggedFields(t, "smart_attribute", test.fields, test.tags)
			}

			for _, test := range testsAda0Device {
				acc.AssertContainsTaggedFields(t, "smart_device", test.fields, test.tags)
			}
		})
		t.Run("Only nvme device", func(t *testing.T) {
			s.Devices = []string{"/dev/nvme0"}
			var acc testutil.Accumulator

			err := s.Gather(&acc)

			require.NoError(t, err)
			assert.Equal(t, 32, acc.NFields(), "Wrong number of fields gathered")

			testutil.RequireMetricsEqual(t, testSmartctlNvmeAttributes, acc.GetTelegrafMetrics(),
				testutil.SortMetrics(), testutil.IgnoreTime())
		})
	})
}

func TestGatherNoAttributes(t *testing.T) {
	s := NewSmart()
	s.Attributes = false

	assert.Equal(t, time.Second*30, s.Timeout.Duration)

	runCmd = func(timeout internal.Duration, sudo bool, command string, args ...string) ([]byte, error) {
		if len(args) > 0 {
			if args[0] == "--scan" && len(args) == 1 {
				return []byte(mockScanData), nil
			} else if args[0] == "--info" && args[7] == "/dev/ada0" {
				return []byte(mockInfoAttributeData), nil
			} else if args[0] == "--info" && args[7] == "/dev/nvme0" {
				return []byte(smartctlNvmeInfoData), nil
			} else if args[0] == "--scan" && args[1] == "--device=nvme" {
				return []byte(mockScanNvmeData), nil
			}
		}
		return nil, errors.New("command not found")
	}

	t.Run("scan for devices", func(t *testing.T) {
		var acc testutil.Accumulator
		s.PathSmartctl = "smartctl"

		err := s.Gather(&acc)

		require.NoError(t, err)
		assert.Equal(t, 8, acc.NFields(), "Wrong number of fields gathered")
		acc.AssertDoesNotContainMeasurement(t, "smart_attribute")

		for _, test := range testsAda0Device {
			acc.AssertContainsTaggedFields(t, "smart_device", test.fields, test.tags)
		}
		for _, test := range testNvmeDevice {
			acc.AssertContainsTaggedFields(t, "smart_device", test.fields, test.tags)
		}
	})
}

func TestExcludedDev(t *testing.T) {
	assert.Equal(t, true, excludedDev([]string{"/dev/pass6"}, "/dev/pass6 -d atacam"), "Should be excluded.")
	assert.Equal(t, false, excludedDev([]string{}, "/dev/pass6 -d atacam"), "Shouldn't be excluded.")
	assert.Equal(t, false, excludedDev([]string{"/dev/pass6"}, "/dev/pass1 -d atacam"), "Shouldn't be excluded.")
}

func TestGatherSATAInfo(t *testing.T) {
	runCmd = func(timeout internal.Duration, sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(hgstSATAInfoData), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, internal.Duration{Duration: time.Second * 30}, true, true, "", "", "", wg)
	assert.Equal(t, 101, acc.NFields(), "Wrong number of fields gathered")
	assert.Equal(t, uint64(20), acc.NMetrics(), "Wrong number of metrics gathered")
}

func TestGatherSATAInfo65(t *testing.T) {
	runCmd = func(timeout internal.Duration, sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(hgstSATAInfoData65), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, internal.Duration{Duration: time.Second * 30}, true, true, "", "", "", wg)
	assert.Equal(t, 91, acc.NFields(), "Wrong number of fields gathered")
	assert.Equal(t, uint64(18), acc.NMetrics(), "Wrong number of metrics gathered")
}

func TestGatherHgstSAS(t *testing.T) {
	runCmd = func(timeout internal.Duration, sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(hgstSASInfoData), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, internal.Duration{Duration: time.Second * 30}, true, true, "", "", "", wg)
	assert.Equal(t, 6, acc.NFields(), "Wrong number of fields gathered")
	assert.Equal(t, uint64(4), acc.NMetrics(), "Wrong number of metrics gathered")
}

func TestGatherHtSAS(t *testing.T) {
	runCmd = func(timeout internal.Duration, sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(htSASInfoData), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, internal.Duration{Duration: time.Second * 30}, true, true, "", "", "", wg)

	testutil.RequireMetricsEqual(t, testHtsasAtributtes, acc.GetTelegrafMetrics(), testutil.SortMetrics(), testutil.IgnoreTime())
}

func TestGatherSSD(t *testing.T) {
	runCmd = func(timeout internal.Duration, sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(ssdInfoData), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, internal.Duration{Duration: time.Second * 30}, true, true, "", "", "", wg)
	assert.Equal(t, 105, acc.NFields(), "Wrong number of fields gathered")
	assert.Equal(t, uint64(26), acc.NMetrics(), "Wrong number of metrics gathered")
}

func TestGatherSSDRaid(t *testing.T) {
	runCmd = func(timeout internal.Duration, sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(ssdRaidInfoData), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, internal.Duration{Duration: time.Second * 30}, true, true, "", "", "", wg)
	assert.Equal(t, 74, acc.NFields(), "Wrong number of fields gathered")
	assert.Equal(t, uint64(15), acc.NMetrics(), "Wrong number of metrics gathered")
}

func TestGatherNvme(t *testing.T) {
	runCmd = func(timeout internal.Duration, sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(smartctlNvmeInfoData), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, internal.Duration{Duration: time.Second * 30}, true, true, "", "", "nvme0", wg)

	testutil.RequireMetricsEqual(t, testSmartctlNvmeAttributes, acc.GetTelegrafMetrics(),
		testutil.SortMetrics(), testutil.IgnoreTime())
}

func TestGatherIntelNvme(t *testing.T) {
	runCmd = func(timeout internal.Duration, sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(nvmeIntelInfoData), nil
	}

	var (
		acc    = &testutil.Accumulator{}
		wg     = &sync.WaitGroup{}
		device = NVMeDevice{
			name:         "nvme0",
			model:        mockModel,
			serialNumber: mockSerial,
		}
	)

	wg.Add(1)
	gatherIntelNVMeDisk(acc, internal.Duration{Duration: time.Second * 30}, true, "", device, wg)

	result := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, testIntelInvmeAttributes, result,
		testutil.SortMetrics(), testutil.IgnoreTime())
}

func Test_findVIDFromNVMeOutput(t *testing.T) {
	vid, sn, mn, err := findNVMeDeviceInfo(nvmeIdentifyController)

	assert.Nil(t, err)
	assert.Equal(t, "0x8086", vid)
	assert.Equal(t, "CVFT5123456789ABCD", sn)
	assert.Equal(t, "INTEL SSDPEDABCDEFG", mn)
}

func Test_checkForNVMeDevices(t *testing.T) {
	devices := []string{"sda1", "nvme0", "sda2", "nvme2"}
	expectedNVMeDevices := []string{"nvme0", "nvme2"}
	resultNVMeDevices := distinguishNVMeDevices(devices, expectedNVMeDevices)
	assert.Equal(t, expectedNVMeDevices, resultNVMeDevices)
}

func Test_excludeWrongDeviceNames(t *testing.T) {
	devices := []string{"/dev/sda", "/dev/nvme -d nvme", "/dev/sda1 -d megaraid,1", "/dev/sda ; ./suspicious_script.sh"}
	validDevices := []string{"/dev/sda", "/dev/nvme -d nvme", "/dev/sda1 -d megaraid,1"}
	result := excludeWrongDeviceNames(devices)
	assert.Equal(t, validDevices, result)
}

func Test_contains(t *testing.T) {
	devices := []string{"/dev/sda", "/dev/nvme1"}
	device := "/dev/nvme1"
	deviceNotIncluded := "/dev/nvme5"
	assert.True(t, contains(devices, device))
	assert.False(t, contains(devices, deviceNotIncluded))
}

func Test_difference(t *testing.T) {
	devices := []string{"/dev/sda", "/dev/nvme1", "/dev/nvme2"}
	secondDevices := []string{"/dev/sda", "/dev/nvme1"}
	expected := []string{"/dev/nvme2"}
	result := difference(devices, secondDevices)
	assert.Equal(t, expected, result)
}

func Test_integerOverflow(t *testing.T) {
	runCmd = func(timeout internal.Duration, sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(smartctlNvmeInfoDataWithOverflow), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	t.Run("If data raw_value is out of int64 range, there should be no metrics for that attribute", func(t *testing.T) {
		wg.Add(1)
		gatherDisk(acc, internal.Duration{Duration: time.Second * 30}, true, true, "", "", "nvme0", wg)

		result := acc.GetTelegrafMetrics()
		testutil.RequireMetricsEqual(t, testOverflowAttributes, result,
			testutil.SortMetrics(), testutil.IgnoreTime())
	})
}

var (
	testOverflowAttributes = []telegraf.Metric{
		testutil.MustMetric(
			"smart_attribute",
			map[string]string{
				"device": "nvme0",
				"name":   "Temperature_Sensor_3",
			},
			map[string]interface{}{
				"raw_value": int64(9223372036854775807),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"smart_attribute",
			map[string]string{
				"device": "nvme0",
				"name":   "Temperature_Sensor_4",
			},
			map[string]interface{}{
				"raw_value": int64(-9223372036854775808),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"smart_device",
			map[string]string{
				"device": "nvme0",
			},
			map[string]interface{}{
				"exit_status": 0,
			},
			time.Unix(0, 0),
		),
	}

	testHtsasAtributtes = []telegraf.Metric{
		testutil.MustMetric(
			"smart_attribute",
			map[string]string{
				"device":    ".",
				"serial_no": "PDWAR9GE",
				"enabled":   "Enabled",
				"id":        "194",
				"model":     "HUC103030CSS600",
				"name":      "Temperature_Celsius",
			},
			map[string]interface{}{
				"raw_value": 36,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"smart_attribute",
			map[string]string{
				"device":    ".",
				"serial_no": "PDWAR9GE",
				"enabled":   "Enabled",
				"id":        "4",
				"model":     "HUC103030CSS600",
				"name":      "Start_Stop_Count",
			},
			map[string]interface{}{
				"raw_value": 47,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"smart_device",
			map[string]string{
				"device":    ".",
				"serial_no": "PDWAR9GE",
				"enabled":   "Enabled",
				"model":     "HUC103030CSS600",
			},
			map[string]interface{}{
				"exit_status": 0,
				"health_ok":   true,
				"temp_c":      36,
			},
			time.Unix(0, 0),
		),
	}

	testsAda0Attributes = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"value":       int64(200),
				"worst":       int64(200),
				"threshold":   int64(0),
				"raw_value":   int64(0),
				"exit_status": int(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
				"id":        "1",
				"name":      "Raw_Read_Error_Rate",
				"flags":     "-O-RC-",
				"fail":      "-",
			},
		},
		{
			map[string]interface{}{
				"value":       int64(100),
				"worst":       int64(100),
				"threshold":   int64(0),
				"raw_value":   int64(0),
				"exit_status": int(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
				"id":        "5",
				"name":      "Reallocated_Sector_Ct",
				"flags":     "PO--CK",
				"fail":      "-",
			},
		},
		{
			map[string]interface{}{
				"value":       int64(99),
				"worst":       int64(99),
				"threshold":   int64(0),
				"raw_value":   int64(2988),
				"exit_status": int(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
				"id":        "9",
				"name":      "Power_On_Hours",
				"flags":     "-O--CK",
				"fail":      "-",
			},
		},
		{
			map[string]interface{}{
				"value":       int64(85),
				"worst":       int64(85),
				"threshold":   int64(0),
				"raw_value":   int64(14879),
				"exit_status": int(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
				"id":        "12",
				"name":      "Power_Cycle_Count",
				"flags":     "-O--CK",
				"fail":      "-",
			},
		},
		{
			map[string]interface{}{
				"value":       int64(253),
				"worst":       int64(253),
				"threshold":   int64(10),
				"raw_value":   int64(2044932921600),
				"exit_status": int(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
				"id":        "169",
				"name":      "Unknown_Attribute",
				"flags":     "PO--C-",
				"fail":      "-",
			},
		},
		{
			map[string]interface{}{
				"value":       int64(185),
				"worst":       int64(185),
				"threshold":   int64(100),
				"raw_value":   int64(957808640337),
				"exit_status": int(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
				"id":        "173",
				"name":      "Wear_Leveling_Count",
				"flags":     "-O--CK",
				"fail":      "-",
			},
		},
		{
			map[string]interface{}{
				"value":       int64(55),
				"worst":       int64(40),
				"threshold":   int64(45),
				"raw_value":   int64(45),
				"exit_status": int(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
				"id":        "190",
				"name":      "Airflow_Temperature_Cel",
				"flags":     "-O---K",
				"fail":      "Past",
			},
		},
		{
			map[string]interface{}{
				"value":       int64(97),
				"worst":       int64(97),
				"threshold":   int64(0),
				"raw_value":   int64(14716),
				"exit_status": int(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
				"id":        "192",
				"name":      "Power-Off_Retract_Count",
				"flags":     "-O--C-",
				"fail":      "-",
			},
		},
		{
			map[string]interface{}{
				"value":       int64(66),
				"worst":       int64(21),
				"threshold":   int64(0),
				"raw_value":   int64(34),
				"exit_status": int(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
				"id":        "194",
				"name":      "Temperature_Celsius",
				"flags":     "-O---K",
				"fail":      "-",
			},
		},
		{
			map[string]interface{}{
				"value":       int64(100),
				"worst":       int64(100),
				"threshold":   int64(0),
				"raw_value":   int64(0),
				"exit_status": int(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
				"id":        "197",
				"name":      "Current_Pending_Sector",
				"flags":     "-O---K",
				"fail":      "-",
			},
		},
		{
			map[string]interface{}{
				"value":       int64(200),
				"worst":       int64(200),
				"threshold":   int64(0),
				"raw_value":   int64(0),
				"exit_status": int(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
				"id":        "199",
				"name":      "UDMA_CRC_Error_Count",
				"flags":     "-O-RC-",
				"fail":      "-",
			},
		},
		{
			map[string]interface{}{
				"value":       int64(100),
				"worst":       int64(253),
				"threshold":   int64(0),
				"raw_value":   int64(23709323),
				"exit_status": int(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
				"id":        "240",
				"name":      "Head_Flying_Hours",
				"flags":     "------",
				"fail":      "-",
			},
		},
	}

	mockModel  = "INTEL SSDPEDABCDEFG"
	mockSerial = "CVFT5123456789ABCD"

	testSmartctlNvmeAttributes = []telegraf.Metric{
		testutil.MustMetric("smart_device",
			map[string]string{
				"device":    "nvme0",
				"model":     "TS128GMTE850",
				"serial_no": "D704940282?",
			},
			map[string]interface{}{
				"exit_status": 0,
				"health_ok":   true,
				"temp_c":      38,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"id":        "9",
				"name":      "Power_On_Hours",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": 6038,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"id":        "12",
				"name":      "Power_Cycle_Count",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": 472,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Media_and_Data_Integrity_Errors",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": 0,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Error_Information_Log_Entries",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": 119699,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Available_Spare",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": 100,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Available_Spare_Threshold",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": 10,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"id":        "194",
				"name":      "Temperature_Celsius",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": 38,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Critical_Warning",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": int64(9),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Percentage_Used",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": int64(16),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Data_Units_Read",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": int64(11836935),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Data_Units_Written",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": int64(62288091),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Host_Read_Commands",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": int64(135924188),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Host_Write_Commands",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": int64(7715573429),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Controller_Busy_Time",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": int64(4042),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Unsafe_Shutdowns",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": int64(355),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Warning_Temperature_Time",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": int64(11),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"name":      "Critical_Temperature_Time",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
			},
			map[string]interface{}{
				"raw_value": int64(7),
			},
			time.Now(),
		), testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
				"name":      "Temperature_Sensor_1",
			},
			map[string]interface{}{
				"raw_value": int64(57),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
				"name":      "Temperature_Sensor_2",
			},
			map[string]interface{}{
				"raw_value": int64(50),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
				"name":      "Temperature_Sensor_3",
			},
			map[string]interface{}{
				"raw_value": int64(44),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
				"name":      "Temperature_Sensor_4",
			},
			map[string]interface{}{
				"raw_value": int64(43),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
				"name":      "Temperature_Sensor_5",
			},
			map[string]interface{}{
				"raw_value": int64(57),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
				"name":      "Temperature_Sensor_6",
			},
			map[string]interface{}{
				"raw_value": int64(50),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
				"name":      "Temperature_Sensor_7",
			},
			map[string]interface{}{
				"raw_value": int64(44),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
				"name":      "Temperature_Sensor_8",
			},
			map[string]interface{}{
				"raw_value": int64(43),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
				"name":      "Thermal_Management_T1_Trans_Count",
			},
			map[string]interface{}{
				"raw_value": 0,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
				"name":      "Thermal_Management_T2_Trans_Count",
			},
			map[string]interface{}{
				"raw_value": 0,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
				"name":      "Thermal_Management_T1_Total_Time",
			},
			map[string]interface{}{
				"raw_value": 0,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": "D704940282?",
				"model":     "TS128GMTE850",
				"name":      "Thermal_Management_T2_Total_Time",
			},
			map[string]interface{}{
				"raw_value": 0,
			},
			time.Now(),
		),
	}

	testsAda0Device = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"exit_status":     int(0),
				"health_ok":       bool(true),
				"read_error_rate": int64(0),
				"temp_c":          int64(34),
				"udma_crc_errors": int64(0),
			},
			map[string]string{
				"device":    "ada0",
				"model":     "APPLE SSD SM256E",
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"enabled":   "Enabled",
				"capacity":  "251000193024",
			},
		},
	}

	testNvmeDevice = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"exit_status": int(0),
				"temp_c":      int64(38),
				"health_ok":   true,
			},
			map[string]string{
				"device":    "nvme0",
				"model":     "TS128GMTE850",
				"serial_no": "D704940282?",
			},
		},
	}

	testIntelInvmeAttributes = []telegraf.Metric{
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Program_Fail_Count",
			},
			map[string]interface{}{
				"raw_value": 0,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Erase_Fail_Count",
			},
			map[string]interface{}{
				"raw_value": 0,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "End_To_End_Error_Detection_Count",
			},
			map[string]interface{}{
				"raw_value": 0,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Crc_Error_Count",
			},
			map[string]interface{}{
				"raw_value": 13,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Retry_Buffer_Overflow_Count",
			},
			map[string]interface{}{
				"raw_value": 0,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Wear_Leveling_Min",
			},
			map[string]interface{}{
				"raw_value": 39,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Wear_Leveling_Max",
			},
			map[string]interface{}{
				"raw_value": 40,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Wear_Leveling_Avg",
			},
			map[string]interface{}{
				"raw_value": 39,
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Timed_Workload_Media_Wear",
			},
			map[string]interface{}{
				"raw_value": float64(0.13),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Timed_Workload_Host_Reads",
			},
			map[string]interface{}{
				"raw_value": float64(71),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Timed_Workload_Timer",
			},
			map[string]interface{}{
				"raw_value": int64(1612952),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Thermal_Throttle_Status_Prc",
			},
			map[string]interface{}{
				"raw_value": float64(0),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Thermal_Throttle_Status_Cnt",
			},
			map[string]interface{}{
				"raw_value": int64(0),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Pll_Lock_Loss_Count",
			},
			map[string]interface{}{
				"raw_value": int64(0),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Nand_Bytes_Written",
			},
			map[string]interface{}{
				"raw_value": int64(0),
			},
			time.Now(),
		),
		testutil.MustMetric("smart_attribute",
			map[string]string{
				"device":    "nvme0",
				"serial_no": mockSerial,
				"model":     mockModel,
				"name":      "Host_Bytes_Written",
			},
			map[string]interface{}{
				"raw_value": int64(0),
			},
			time.Now(),
		),
	}
	// smartctl --scan
	mockScanData = `/dev/ada0 -d atacam # /dev/ada0, ATA device`

	// smartctl --scan -d nvme
	mockScanNvmeData = `/dev/nvme0 -d nvme # /dev/nvme0, NVMe device`

	// smartctl --info --health --attributes --tolerance=verypermissive -n standby --format=brief [DEVICE]
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
SMART overall-health self-assessment test result: PASSED

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

	htSASInfoData = `smartctl 6.6 2016-05-31 r4324 [x86_64-linux-4.15.18-12-pve] (local build)
Copyright (C) 2002-16, Bruce Allen, Christian Franke, www.smar$montools.org

=== START OF INFORMATION SECTION ===
Vendor:               HITACHI
Product:              HUC103030CSS600
Revision:             J350
Compliance:           SPC-4
User Capacity:        300,$00,000,000 bytes [300 GB]
Logical block size:   512 bytes
Rotation Rate:        10020 rpm
Form Factor:          2.5 inches
Logical Unit id:      0x5000cca00a4bdbc8
Serial number:        PDWAR9GE
Devicetype:          disk
Transport protocol:   SAS (SPL-3)
Local Time is:        Wed Apr 17 15:01:28 2019 PDT
SMART support is:     Available - device has SMART capability.
SMART support is:     Enabled
Temperature Warning:  Disabled or Not Supported

=== START OF READ SMART DATA SECTION ===
SMART Health Status: OK

Current Drive Temperature:     36 C
Drive Trip Temperature:        85 C

Manufactured in $eek 52 of year 2009
Specified cycle count over device lifetime:  50000
Accumulated start-stop cycles:  47
Elements in grown defect list: 0

Vendor (Seagate) cache information
	Blocks sent to initiator= 7270983270400000
`

	hgstSASInfoData = `smartctl 6.6 2016-05-31 r4324 [x86_64-linux-4.15.0-46-generic] (local build)
Copyright (C) 2002-16, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Vendor:               HGST
Product:              HUH721212AL5204
Revision:             C3Q1
Compliance:           SPC-4
User Capacity:        12,000,138,625,024 bytes [12.0 TB]
Logical block size:   512 bytes
Physical block size:  4096 bytes
LU is fully provisioned
Rotation Rate:        7200 rpm
Form Factor:          3.5 inches
Logical Unit id:      0x5000cca27076bfe8
Serial number:        8HJ39K3H
Device type:          disk
Transport protocol:   SAS (SPL-3)
Local Time is:        Thu Apr 18 13:25:03 2019 MSK
SMART support is:     Available - device has SMART capability.
SMART support is:     Enabled
Temperature Warning:  Enabled

=== START OF READ SMART DATA SECTION ===
SMART Health Status: OK

Current Drive Temperature:     34 C
Drive Trip Temperature:        85 C

Manufactured in week 35 of year 2018
Specified cycle count over device lifetime:  50000
Accumulated start-stop cycles:  7
Specified load-unload count over device lifetime:  600000
Accumulated load-unload cycles:  39
Elements in grown defect list: 0

Vendor (Seagate) cache information
  Blocks sent to initiator = 544135446528
`

	hgstSATAInfoData = `smartctl 6.6 2016-05-31 r4324 [x86_64-linux-4.15.0-46-generic] (local build)
Copyright (C) 2002-16, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Model Family:     Hitachi/HGST Travelstar Z7K500
Device Model:     HGST HTE725050A7E630
Serial Number:    RCE50G20G81S9S
LU WWN Device Id: 5 000cca 90bc3a98b
Firmware Version: GS2OA3E0
User Capacity:    500,107,862,016 bytes [500 GB]
Sector Sizes:     512 bytes logical, 4096 bytes physical
Rotation Rate:    7200 rpm
Form Factor:      2.5 inches
Device is:        In smartctl database [for details use: -P show]
ATA Version is:   ATA8-ACS T13/1699-D revision 6
SATA Version is:  SATA 2.6, 6.0 Gb/s (current: 6.0 Gb/s)
Local Time is:    Thu Apr 18 13:27:51 2019 MSK
SMART support is: Available - device has SMART capability.
SMART support is: Enabled
Power mode is:    ACTIVE or IDLE

=== START OF READ SMART DATA SECTION ===
SMART overall-health self-assessment test result: PASSED

SMART Attributes Data Structure revision number: 16
Vendor Specific SMART Attributes with Thresholds:
ID# ATTRIBUTE_NAME          FLAGS    VALUE WORST THRESH FAIL RAW_VALUE
  1 Raw_Read_Error_Rate     PO-R--   100   100   062    -    0
  2 Throughput_Performance  P-S---   100   100   040    -    0
  3 Spin_Up_Time            POS---   100   100   033    -    1
  4 Start_Stop_Count        -O--C-   100   100   000    -    4
  5 Reallocated_Sector_Ct   PO--CK   100   100   005    -    0
  7 Seek_Error_Rate         PO-R--   100   100   067    -    0
  8 Seek_Time_Performance   P-S---   100   100   040    -    0
  9 Power_On_Hours          -O--C-   099   099   000    -    743
 10 Spin_Retry_Count        PO--C-   100   100   060    -    0
 12 Power_Cycle_Count       -O--CK   100   100   000    -    4
191 G-Sense_Error_Rate      -O-R--   100   100   000    -    0
192 Power-Off_Retract_Count -O--CK   100   100   000    -    2
193 Load_Cycle_Count        -O--C-   100   100   000    -    13
194 Temperature_Celsius     -O----   250   250   000    -    24 (Min/Max 15/29)
196 Reallocated_Event_Count -O--CK   100   100   000    -    0
197 Current_Pending_Sector  -O---K   100   100   000    -    0
198 Offline_Uncorrectable   ---R--   100   100   000    -    0
199 UDMA_CRC_Error_Count    -O-R--   200   200   000    -    0
223 Load_Retry_Count        -O-R--   100   100   000    -    0
                            ||||||_ K auto-keep
                            |||||__ C event count
                            ||||___ R error rate
                            |||____ S speed/performance
                            ||_____ O updated online
                            |______ P prefailure warning
`

	hgstSATAInfoData65 = `smartctl 6.5 2016-01-24 r4214 [x86_64-linux-4.4.0-145-generic] (local build)
Copyright (C) 2002-16, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Model Family:     HGST Deskstar NAS
Device Model:     HGST HDN724040ALE640
Serial Number:    PK1334PEK49SBS
LU WWN Device Id: 5 000cca 250ec3c9c
Firmware Version: MJAOA5E0
User Capacity:    4,000,787,030,016 bytes [4.00 TB]
Sector Sizes:     512 bytes logical, 4096 bytes physical
Rotation Rate:    7200 rpm
Form Factor:      3.5 inches
Device is:        In smartctl database [for details use: -P show]
ATA Version is:   ATA8-ACS T13/1699-D revision 4
SATA Version is:  SATA 3.0, 6.0 Gb/s (current: 6.0 Gb/s)
Local Time is:    Wed Apr 17 15:14:27 2019 PDT
SMART support is: Available - device has SMART capability.
SMART support is: Enabled
Power mode is:    ACTIVE or IDLE

=== START OF READ SMART DATA SECTION ===
SMART overall-health self-assessment test result: PASSED

SMART Attributes Data Structure revision number: 16
Vendor Specific SMART Attributes with Thresholds:
ID# ATTRIBUTE_NAME          FLAGS    VALUE WORST THRESH FAIL RAW_VALUE
  1 Raw_Read_Error_Rate     PO-R--   100   100   016    -    0
  2 Throughput_Performance  P-S---   135   135   054    -    84
  3 Spin_Up_Time            POS---   125   125   024    -    621 (Average 619)
  4 Start_Stop_Count        -O--C-   100   100   000    -    33
  5 Reallocated_Sector_Ct   PO--CK   100   100   005    -    0
  7 Seek_Error_Rate         PO-R--   100   100   067    -    0
  8 Seek_Time_Performance   P-S---   119   119   020    -    35
  9 Power_On_Hours          -O--C-   098   098   000    -    19371
 10 Spin_Retry_Count        PO--C-   100   100   060    -    0
 12 Power_Cycle_Count       -O--CK   100   100   000    -    33
192 Power-Off_Retract_Count -O--CK   100   100   000    -    764
193 Load_Cycle_Count        -O--C-   100   100   000    -    764
194 Temperature_Celsius     -O----   176   176   000    -    34 (Min/Max 21/53)
196 Reallocated_Event_Count -O--CK   100   100   000    -    0
197 Current_Pending_Sector  -O---K   100   100   000    -    0
198 Offline_Uncorrectable   ---R--   100   100   000    -    0
199 UDMA_CRC_Error_Count    -O-R--   200   200   000    -    0
                            ||||||_ K auto-keep
                            |||||__ C event count
                            ||||___ R error rate
                            |||____ S speed/performance
                            ||_____ O updated online
                            |______ P prefailure warning
`

	ssdInfoData = `smartctl 6.6 2016-05-31 r4324 [x86_64-linux-4.15.0-33-generic] (local build)
Copyright (C) 2002-16, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Device Model:     SanDisk Ultra II 240GB
Serial Number:    XXXXXXXX
LU WWN Device Id: XXXXXXXX
Firmware Version: XXXXXXX
User Capacity:    240.057.409.536 bytes [240 GB]
Sector Size:      512 bytes logical/physical
Rotation Rate:    Solid State Device
Form Factor:      2.5 inches
Device is:        Not in smartctl database [for details use: -P showall]
ATA Version is:   ACS-2 T13/2015-D revision 3
SATA Version is:  SATA 3.2, 6.0 Gb/s (current: 6.0 Gb/s)
Local Time is:    Mon Sep 17 13:22:19 2018 CEST
SMART support is: Available - device has SMART capability.
SMART support is: Enabled
Power mode is:    ACTIVE or IDLE

=== START OF READ SMART DATA SECTION ===
SMART overall-health self-assessment test result: PASSED

SMART Attributes Data Structure revision number: 4
Vendor Specific SMART Attributes with Thresholds:
ID# ATTRIBUTE_NAME          FLAGS    VALUE WORST THRESH FAIL RAW_VALUE
  5 Reallocated_Sector_Ct   -O--CK   100   100   ---    -    0
  9 Power_On_Hours          -O--CK   100   100   ---    -    6383
 12 Power_Cycle_Count       -O--CK   100   100   ---    -    19
165 Unknown_Attribute       -O--CK   100   100   ---    -    59310806
166 Unknown_Attribute       -O--CK   100   100   ---    -    1
167 Unknown_Attribute       -O--CK   100   100   ---    -    57
168 Unknown_Attribute       -O--CK   100   100   ---    -    43
169 Unknown_Attribute       -O--CK   100   100   ---    -    221
170 Unknown_Attribute       -O--CK   100   100   ---    -    0
171 Unknown_Attribute       -O--CK   100   100   ---    -    0
172 Unknown_Attribute       -O--CK   100   100   ---    -    0
173 Unknown_Attribute       -O--CK   100   100   ---    -    13
174 Unknown_Attribute       -O--CK   100   100   ---    -    4
184 End-to-End_Error        -O--CK   100   100   ---    -    0
187 Reported_Uncorrect      -O--CK   100   100   ---    -    0
188 Command_Timeout         -O--CK   100   100   ---    -    0
194 Temperature_Celsius     -O---K   066   065   ---    -    34 (Min/Max 19/65)
199 UDMA_CRC_Error_Count    -O--CK   100   100   ---    -    0
230 Unknown_SSD_Attribute   -O--CK   100   100   ---    -    2229110374919
232 Available_Reservd_Space PO--CK   100   100   004    -    100
233 Media_Wearout_Indicator -O--CK   100   100   ---    -    3129
234 Unknown_Attribute       -O--CK   100   100   ---    -    7444
241 Total_LBAs_Written      ----CK   253   253   ---    -    4812
242 Total_LBAs_Read         ----CK   253   253   ---    -    671
244 Unknown_Attribute       -O--CK   000   100   ---    -    0
                            ||||||_ K auto-keep
                            |||||__ C event count
                            ||||___ R error rate
                            |||____ S speed/performance
                            ||_____ O updated online
														|______ P prefailure warning
`
	ssdRaidInfoData = `smartctl 6.6 2017-11-05 r4594 [FreeBSD 11.1-RELEASE-p13 amd64] (local build)
Copyright (C) 2002-17, Bruce Allen, Christian Franke, www.smartmontools.org

CHECK POWER MODE: incomplete response, ATA output registers missing
CHECK POWER MODE not implemented, ignoring -n option
=== START OF INFORMATION SECTION ===
Model Family:     Samsung based SSDs
Device Model:     Samsung SSD 850 PRO 256GB
Serial Number:    S251NX0H869353L
LU WWN Device Id: 5 002538 84027f72f
Firmware Version: EXM02B6Q
User Capacity:    256 060 514 304 bytes [256 GB]
Sector Size:      512 bytes logical/physical
Rotation Rate:    Solid State Device
Device is:        In smartctl database [for details use: -P show]
ATA Version is:   ACS-2, ATA8-ACS T13/1699-D revision 4c
SATA Version is:  SATA 3.1, 6.0 Gb/s (current: 6.0 Gb/s)
Local Time is:    Fri Sep 21 17:49:16 2018 CEST
SMART support is: Available - device has SMART capability.
SMART support is: Enabled

=== START OF READ SMART DATA SECTION ===
SMART Status not supported: Incomplete response, ATA output registers missing
SMART overall-health self-assessment test result: PASSED
Warning: This result is based on an Attribute check.

General SMART Values:
Offline data collection status:  (0x00)	Offline data collection activity
					was never started.
					Auto Offline Data Collection: Disabled.
Self-test execution status:      (   0)	The previous self-test routine completed
					without error or no self-test has ever
					been run.
Total time to complete Offline
data collection: 		(    0) seconds.
Offline data collection
capabilities: 			 (0x53) SMART execute Offline immediate.
					Auto Offline data collection on/off support.
					Suspend Offline collection upon new
					command.
					No Offline surface scan supported.
					Self-test supported.
					No Conveyance Self-test supported.
					Selective Self-test supported.
SMART capabilities:            (0x0003)	Saves SMART data before entering
					power-saving mode.
					Supports SMART auto save timer.
Error logging capability:        (0x01)	Error logging supported.
					General Purpose Logging supported.
Short self-test routine
recommended polling time: 	 (   2) minutes.
Extended self-test routine
recommended polling time: 	 ( 136) minutes.
SCT capabilities: 	       (0x003d)	SCT Status supported.
					SCT Error Recovery Control supported.
					SCT Feature Control supported.
					SCT Data Table supported.

SMART Attributes Data Structure revision number: 1
Vendor Specific SMART Attributes with Thresholds:
ID# ATTRIBUTE_NAME          FLAGS    VALUE WORST THRESH FAIL RAW_VALUE
	5 Reallocated_Sector_Ct   PO--CK   099   099   010    -    1
	9 Power_On_Hours          -O--CK   094   094   000    -    26732
	12 Power_Cycle_Count       -O--CK   099   099   000    -    51
177 Wear_Leveling_Count     PO--C-   001   001   000    -    7282
179 Used_Rsvd_Blk_Cnt_Tot   PO--C-   099   099   010    -    1
181 Program_Fail_Cnt_Total  -O--CK   100   100   010    -    0
182 Erase_Fail_Count_Total  -O--CK   099   099   010    -    1
183 Runtime_Bad_Block       PO--C-   099   099   010    -    1
187 Uncorrectable_Error_Cnt -O--CK   100   100   000    -    0
190 Airflow_Temperature_Cel -O--CK   081   069   000    -    19
195 ECC_Error_Rate          -O-RC-   200   200   000    -    0
199 CRC_Error_Count         -OSRCK   100   100   000    -    0
235 POR_Recovery_Count      -O--C-   099   099   000    -    50
241 Total_LBAs_Written      -O--CK   099   099   000    -    61956393677
														||||||_ K auto-keep
														|||||__ C event count
														||||___ R error rate
														|||____ S speed/performance
														||_____ O updated online
														|______ P prefailure warning

SMART Error Log Version: 1
No Errors Logged

SMART Self-test log structure revision number 1
Num  Test_Description    Status                  Remaining  LifeTime(hours)  LBA_of_first_error
# 1  Short offline       Completed without error       00%     26717         -
# 2  Short offline       Completed without error       00%     26693         -
# 3  Short offline       Completed without error       00%     26669         -
# 4  Short offline       Completed without error       00%     26645         -
# 5  Short offline       Completed without error       00%     26621         -
# 6  Short offline       Completed without error       00%     26596         -
# 7  Extended offline    Completed without error       00%     26574         -
# 8  Short offline       Completed without error       00%     26572         -
# 9  Short offline       Completed without error       00%     26548         -
#10  Short offline       Completed without error       00%     26524         -
#11  Short offline       Completed without error       00%     26500         -
#12  Short offline       Completed without error       00%     26476         -
#13  Short offline       Completed without error       00%     26452         -
#14  Short offline       Completed without error       00%     26428         -
#15  Extended offline    Completed without error       00%     26406         -
#16  Short offline       Completed without error       00%     26404         -
#17  Short offline       Completed without error       00%     26380         -
#18  Short offline       Completed without error       00%     26356         -
#19  Short offline       Completed without error       00%     26332         -
#20  Short offline       Completed without error       00%     26308         -

SMART Selective self-test log data structure revision number 1
	SPAN  MIN_LBA  MAX_LBA  CURRENT_TEST_STATUS
		1        0        0  Not_testing
		2        0        0  Not_testing
		3        0        0  Not_testing
		4        0        0  Not_testing
		5        0        0  Not_testing
Selective self-test flags (0x0):
	After scanning selected spans, do NOT read-scan remainder of disk.
If Selective self-test is pending on power-up, resume after 0 minute delay.
`
	smartctlNvmeInfoData = `smartctl 6.5 2016-05-07 r4318 [x86_64-linux-4.1.27-gvt-yocto-standard] (local build)
Copyright (C) 2002-16, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Model Number: TS128GMTE850
Serial Number: D704940282?
Firmware Version: C2.3.13
PCI Vendor/Subsystem ID: 0x126f
IEEE OUI Identifier: 0x000000
Controller ID: 1
Number of Namespaces: 1
Namespace 1 Size/Capacity: 128,035,676,160 [128 GB]
Namespace 1 Formatted LBA Size: 512
Local Time is: Fri Jun 15 11:41:35 2018 UTC

=== START OF SMART DATA SECTION ===
SMART overall-health self-assessment test result: PASSED

SMART/Health Information (NVMe Log 0x02, NSID 0xffffffff)
Critical Warning: 0x09
Temperature: 38 Celsius
Available Spare: 100%
Available Spare Threshold: 10%
Percentage Used: 16%
Data Units Read: 11,836,935 [6.06 TB]
Data Units Written: 62,288,091 [31.8 TB]
Host Read Commands: 135,924,188
Host Write Commands: 7,715,573,429
Controller Busy Time: 4,042
Power Cycles: 472
Power On Hours: 6,038
Unsafe Shutdowns: 355
Media and Data Integrity Errors: 0
Error Information Log Entries: 119,699
Warning  Comp. Temperature Time: 11
Critical Comp. Temperature Time: 7
Thermal Temp. 1 Transition Count: 0
Thermal Temp. 2 Transition Count: 0
Thermal Temp. 1 Total Time: 0
Thermal Temp. 2 Total Time: 0
Temperature Sensor 1: 57 C
Temperature Sensor 2: 50 C
Temperature Sensor 3: 44 C
Temperature Sensor 4: 43 C
Temperature Sensor 5: 57 C
Temperature Sensor 6: 50 C
Temperature Sensor 7: 44 C
Temperature Sensor 8: 43 C
`

	smartctlNvmeInfoDataWithOverflow = `
Temperature Sensor 1: 9223372036854775808 C
Temperature Sensor 2: -9223372036854775809 C
Temperature Sensor 3: 9223372036854775807 C
Temperature Sensor 4: -9223372036854775808 C
`

	nvmeIntelInfoData = `Additional Smart Log for NVME device:nvme0 namespace-id:ffffffff
key                               normalized raw
program_fail_count              : 100%       0
erase_fail_count                : 100%       0
wear_leveling                   : 100%       min: 39, max: 40, avg: 39
end_to_end_error_detection_count: 100%       0
crc_error_count                 : 100%       13
timed_workload_media_wear       : 100%       0.130%
timed_workload_host_reads       : 100%       71%
timed_workload_timer            : 100%       1612952 min
thermal_throttle_status         : 100%       0%, cnt: 0
retry_buffer_overflow_count     : 100%       0
pll_lock_loss_count             : 100%       0
nand_bytes_written              :   0%       sectors: 0
host_bytes_written              :   0%       sectors: 0
`

	nvmeIdentifyController = `NVME Identify Controller:
vid     : 0x8086
ssvid   : 0x8086
sn      : CVFT5123456789ABCD
mn      : INTEL SSDPEDABCDEFG
fr      : 8DV10131
rab     : 0
ieee    : 5cd2e4
cmic    : 0
mdts    : 5
cntlid  : 0
ver     : 0
rtd3r   : 0
rtd3e   : 0
<<<<<<< HEAD
oaes    : 0
ctratt  : 0
oacs    : 0x6
acl     : 3
aerl    : 3
frmw    : 0x2
lpa     : 0
elpe    : 63
npss    : 0
avscc   : 0
apsta   : 0
wctemp  : 0
cctemp  : 0
mtfa    : 0
hmpre   : 0
hmmin   : 0
tnvmcap : 0
unvmcap : 0
rpmbs   : 0
edstt   : 0
dsto    : 0
fwug    : 0
kas     : 0
hctma   : 0
mntmt   : 0
mxtmt   : 0
sanicap : 0
hmminds : 0
hmmaxd  : 0
sqes    : 0x66
cqes    : 0x44
maxcmd  : 0
nn      : 1
oncs    : 0x6
fuses   : 0
fna     : 0x7
vwc     : 0
awun    : 0
awupf   : 0
nvscc   : 0
acwu    : 0
sgls    : 0
subnqn  :
ioccsz  : 0
iorcsz  : 0
icdoff  : 0
ctrattr : 0
msdbd   : 0
ps    0 : mp:25.00W operational enlat:0 exlat:0 rrt:0 rrl:0
          rwt:0 rwl:0 idle_power:- active_power:-
`
)
