package smart

import (
	"errors"
	"sync"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherAttributes(t *testing.T) {
	s := &Smart{
		Path:       "smartctl",
		Attributes: true,
	}
	var acc testutil.Accumulator

	runCmd = func(sudo bool, command string, args ...string) ([]byte, error) {
		if len(args) > 0 {
			if args[0] == "--scan" {
				return []byte(mockScanData), nil
			} else if args[0] == "--info" {
				return []byte(mockInfoAttributeData), nil
			}
		}
		return nil, errors.New("command not found")
	}

	err := s.Gather(&acc)

	require.NoError(t, err)
	assert.Equal(t, 65, acc.NFields(), "Wrong number of fields gathered")

	var testsAda0Attributes = []struct {
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
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
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
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
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
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
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
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
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
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
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
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
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
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
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
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
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
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
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
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
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
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
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
				"serial_no": "S0X5NZBC422720",
				"wwn":       "5002538043584d30",
				"id":        "240",
				"name":      "Head_Flying_Hours",
				"flags":     "------",
				"fail":      "-",
			},
		},
	}

	for _, test := range testsAda0Attributes {
		acc.AssertContainsTaggedFields(t, "smart_attribute", test.fields, test.tags)
	}

	var testsAda0Device = []struct {
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

	for _, test := range testsAda0Device {
		acc.AssertContainsTaggedFields(t, "smart_device", test.fields, test.tags)
	}
}

func TestGatherNoAttributes(t *testing.T) {
	s := &Smart{
		Path:       "smartctl",
		Attributes: false,
	}
	// overwriting exec commands with mock commands
	var acc testutil.Accumulator

	err := s.Gather(&acc)

	require.NoError(t, err)
	assert.Equal(t, 5, acc.NFields(), "Wrong number of fields gathered")
	acc.AssertDoesNotContainMeasurement(t, "smart_attribute")

	var testsAda0Device = []struct {
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

	for _, test := range testsAda0Device {
		acc.AssertContainsTaggedFields(t, "smart_device", test.fields, test.tags)
	}
}

func TestExcludedDev(t *testing.T) {
	assert.Equal(t, true, excludedDev([]string{"/dev/pass6"}, "/dev/pass6 -d atacam"), "Should be excluded.")
	assert.Equal(t, false, excludedDev([]string{}, "/dev/pass6 -d atacam"), "Shouldn't be excluded.")
	assert.Equal(t, false, excludedDev([]string{"/dev/pass6"}, "/dev/pass1 -d atacam"), "Shouldn't be excluded.")
}

func TestGatherSATAInfo(t *testing.T) {
	runCmd = func(sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(hgstSATAInfoData), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, true, true, "", "", "", wg)
	assert.Equal(t, 101, acc.NFields(), "Wrong number of fields gathered")
	assert.Equal(t, uint64(20), acc.NMetrics(), "Wrong number of metrics gathered")

	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"PO-R--", "id":"1", "name":"Raw_Read_Error_Rate", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "threshold":62, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948ce5386, ext:2331252, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"P-S---", "id":"2", "name":"Throughput_Performance", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "threshold":40, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948ce72b5, ext:2339232, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"POS---", "id":"3", "name":"Spin_Up_Time", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":1, "threshold":33, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948ce8b37, ext:2345506, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--C-", "id":"4", "name":"Start_Stop_Count", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":4, "threshold":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948cea1c7, ext:2351282, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"PO--CK", "id":"5", "name":"Reallocated_Sector_Ct", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "threshold":5, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948cec9a4, ext:2361487, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"PO-R--", "id":"7", "name":"Seek_Error_Rate", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "threshold":67, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948cee174, ext:2367583, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"P-S---", "id":"8", "name":"Seek_Time_Performance", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "threshold":40, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948cef6bb, ext:2373029, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--C-", "id":"9", "name":"Power_On_Hours", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":743, "threshold":0, "value":99, "worst":99}, Time:time.Time{wall:0xbf27f09948cf0c06, ext:2378480, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"PO--C-", "id":"10", "name":"Spin_Retry_Count", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "threshold":60, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948cf5f83, ext:2399856, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"12", "name":"Power_Cycle_Count", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":4, "threshold":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948cf76dd, ext:2405831, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O-R--", "id":"191", "name":"G-Sense_Error_Rate", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "threshold":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948cf95a3, ext:2413711, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"192", "name":"Power-Off_Retract_Count", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":2, "threshold":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948cfba9f, ext:2423179, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--C-", "id":"193", "name":"Load_Cycle_Count", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":13, "threshold":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948cfd1bf, ext:2429099, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O----", "id":"194", "name":"Temperature_Celsius", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":24, "threshold":0, "value":250, "worst":250}, Time:time.Time{wall:0xbf27f09948cfe922, ext:2435086, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"196", "name":"Reallocated_Event_Count", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "threshold":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948cffeff, ext:2440683, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O---K", "id":"197", "name":"Current_Pending_Sector", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "threshold":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948d023d4, ext:2450113, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"---R--", "id":"198", "name":"Offline_Uncorrectable", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "threshold":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948d039f1, ext:2455773, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O-R--", "id":"199", "name":"UDMA_CRC_Error_Count", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "threshold":0, "value":200, "worst":200}, Time:time.Time{wall:0xbf27f09948d05040, ext:2461485, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O-R--", "id":"223", "name":"Load_Retry_Count", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "threshold":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf27f09948d0755a, ext:2470982, loc:(*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement:"smart_device", Tags:map[string]string{"capacity":"500107862016", "device":".", "enabled":"Enabled", "model":"HGST HTE725050A7E630", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "health_ok":true, "read_error_rate":0, "seek_error_rate":0, "temp_c":24, "udma_crc_errors":0}, Time:time.Time{wall:0xbf27f09948d0c438, ext:2491172, loc:(*time.Location)(0x9cd280)}}
}

func TestGatherSATAInfo65(t *testing.T) {
	runCmd = func(sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(hgstSATAInfoData65), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, true, true, "", "", "", wg)
	assert.Equal(t, 91, acc.NFields(), "Wrong number of fields gathered")
	assert.Equal(t, uint64(18), acc.NMetrics(), "Wrong number of metrics gathered")

	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "PO-R--", "id": "1", "name": "Raw_Read_Error_Rate", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 0, "threshold": 16, "value": 100, "worst": 100}, Time: time.Time{wall: 0xbf27f09948da776f, ext: 3126876, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "P-S---", "id": "2", "name": "Throughput_Performance", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 84, "threshold": 54, "value": 135, "worst": 135}, Time: time.Time{wall: 0xbf27f09948da9177, ext: 3133540, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "POS---", "id": "3", "name": "Spin_Up_Time", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 621, "threshold": 24, "value": 125, "worst": 125}, Time: time.Time{wall: 0xbf27f09948dab157, ext: 3141701, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "-O--C-", "id": "4", "name": "Start_Stop_Count", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 33, "threshold": 0, "value": 100, "worst": 100}, Time: time.Time{wall: 0xbf27f09948dad35c, ext: 3150407, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "PO--CK", "id": "5", "name": "Reallocated_Sector_Ct", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 0, "threshold": 5, "value": 100, "worst": 100}, Time: time.Time{wall: 0xbf27f09948daf2e9, ext: 3158490, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "PO-R--", "id": "7", "name": "Seek_Error_Rate", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 0, "threshold": 67, "value": 100, "worst": 100}, Time: time.Time{wall: 0xbf27f09948db1b78, ext: 3168867, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "P-S---", "id": "8", "name": "Seek_Time_Performance", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 35, "threshold": 20, "value": 119, "worst": 119}, Time: time.Time{wall: 0xbf27f09948db333d, ext: 3174953, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "-O--C-", "id": "9", "name": "Power_On_Hours", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 19371, "threshold": 0, "value": 98, "worst": 98}, Time: time.Time{wall: 0xbf27f09948db4af8, ext: 3181029, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "PO--C-", "id": "10", "name": "Spin_Retry_Count", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 0, "threshold": 60, "value": 100, "worst": 100}, Time: time.Time{wall: 0xbf27f09948db619e, ext: 3186827, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "-O--CK", "id": "12", "name": "Power_Cycle_Count", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 33, "threshold": 0, "value": 100, "worst": 100}, Time: time.Time{wall: 0xbf27f09948db7f70, ext: 3194463, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "-O--CK", "id": "192", "name": "Power-Off_Retract_Count", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 764, "threshold": 0, "value": 100, "worst": 100}, Time: time.Time{wall: 0xbf27f09948db95c6, ext: 3200178, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "-O--C-", "id": "193", "name": "Load_Cycle_Count", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 764, "threshold": 0, "value": 100, "worst": 100}, Time: time.Time{wall: 0xbf27f09948dbbbac, ext: 3209883, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "-O----", "id": "194", "name": "Temperature_Celsius", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 34, "threshold": 0, "value": 176, "worst": 176}, Time: time.Time{wall: 0xbf27f09948dbf101, ext: 3223533, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "-O--CK", "id": "196", "name": "Reallocated_Event_Count", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 0, "threshold": 0, "value": 100, "worst": 100}, Time: time.Time{wall: 0xbf27f09948dc0916, ext: 3229697, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "-O---K", "id": "197", "name": "Current_Pending_Sector", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 0, "threshold": 0, "value": 100, "worst": 100}, Time: time.Time{wall: 0xbf27f09948dc1ee8, ext: 3235286, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "---R--", "id": "198", "name": "Offline_Uncorrectable", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 0, "threshold": 0, "value": 100, "worst": 100}, Time: time.Time{wall: 0xbf27f09948dc3b1e, ext: 3242507, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_attribute", Tags: map[string]string{"device": ".", "fail": "-", "flags": "-O-R--", "id": "199", "name": "UDMA_CRC_Error_Count", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "raw_value": 0, "threshold": 0, "value": 200, "worst": 200}, Time: time.Time{wall: 0xbf27f09948dc5843, ext: 3249968, loc: (*time.Location)(0x9cd280)}}
	// &testutil.Metric{Measurement: "smart_device", Tags: map[string]string{"capacity": "4000787030016", "device": ".", "enabled": "Enabled", "model": "HGST HDN724040ALE640", "serial_no": "PK1334PEK49SBS", "wwn": "5000cca250ec3c9c"}, Fields: map[string]interface{}{"exit_status": 0, "health_ok": true, "read_error_rate": 0, "seek_error_rate": 0, "temp_c": 34, "udma_crc_errors": 0}, Time: time.Time{wall: 0xbf27f09948dca7f7, ext: 3270372, loc: (*time.Location)(0x9cd280)}}
}

func TestGatherHgstSAS(t *testing.T) {
	runCmd = func(sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(hgstSASInfoData), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, true, true, "", "", "", wg)
	assert.Equal(t, 6, acc.NFields(), "Wrong number of fields gathered")
	assert.Equal(t, uint64(4), acc.NMetrics(), "Wrong number of metrics gathered")

	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"id":"194", "name":"Temperature_Celsius"}, Fields:map[string]interface {}{"raw_value":34}, Time:time.Time{wall:0xbf27f343d2756216, ext:1805738, loc:(*time.Location)(0x9cd2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"id":"4", "name":"Start_Stop_Count"}, Fields:map[string]interface {}{"raw_value":7}, Time:time.Time{wall:0xbf27f343d2759415, ext:1818537, loc:(*time.Location)(0x9cd2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"id":"193", "name":"Load_Cycle_Count"}, Fields:map[string]interface {}{"raw_value":39}, Time:time.Time{wall:0xbf27f343d275aa6f, ext:1824257, loc:(*time.Location)(0x9cd2a0)}}
	// &testutil.Metric{Measurement:"smart_device", Tags:map[string]string{"capacity":"12000138625024", "device":".", "enabled":"Enabled", "model":"HUH721212AL5204"}, Fields:map[string]interface {}{"exit_status":0, "health_ok":true, "temp_c":34}, Time:time.Time{wall:0xbf27f343d275c950, ext:1832162, loc:(*time.Location)(0x9cd2a0)}}
}

func TestGatherHtSAS(t *testing.T) {
	runCmd = func(sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(htSASInfoData), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, true, true, "", "", "", wg)
	assert.Equal(t, 5, acc.NFields(), "Wrong number of fields gathered")
	assert.Equal(t, uint64(3), acc.NMetrics(), "Wrong number of metrics gathered")

	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"id":"194", "name":"Temperature_Celsius"}, Fields:map[string]interface {}{"raw_value":36}, Time:time.Time{wall:0xbf27f32fc76b8ef0, ext:1401923, loc:(*time.Location)(0x9cd2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"id":"4", "name":"Start_Stop_Count"}, Fields:map[string]interface {}{"raw_value":47}, Time:time.Time{wall:0xbf27f32fc76bc2b3, ext:1415174, loc:(*time.Location)(0x9cd2a0)}}
	// &testutil.Metric{Measurement:"smart_device", Tags:map[string]string{"device":".", "enabled":"Enabled", "model":"HUC103030CSS600"}, Fields:map[string]interface {}{"exit_status":0, "health_ok":true, "temp_c":36}, Time:time.Time{wall:0xbf27f32fc76be571, ext:1424067, loc:(*time.Location)(0x9cd2a0)}}
}

func TestGatherSSD(t *testing.T) {
	runCmd = func(sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(ssdInfoData), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, true, true, "", "", "", wg)
	assert.Equal(t, 105, acc.NFields(), "Wrong number of fields gathered")
	assert.Equal(t, uint64(26), acc.NMetrics(), "Wrong number of metrics gathered")

	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"5", "name":"Reallocated_Sector_Ct", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac24fed36, ext:1372338, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"9", "name":"Power_On_Hours", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":6383, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac25016d5, ext:1383002, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"12", "name":"Power_Cycle_Count", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":19, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac2503634, ext:1391024, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"165", "name":"Unknown_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":59310806, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac2504e6c, ext:1397226, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"166", "name":"Unknown_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":1, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac25064da, ext:1402967, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"167", "name":"Unknown_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":57, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac2508efa, ext:1413751, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"168", "name":"Unknown_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":43, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac250a69c, ext:1419803, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"169", "name":"Unknown_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":221, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac250bca3, ext:1425441, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"170", "name":"Unknown_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac250d927, ext:1432740, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"171", "name":"Unknown_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac250f012, ext:1438605, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"172", "name":"Unknown_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac2510bde, ext:1445722, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"173", "name":"Unknown_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":13, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac2512368, ext:1451748, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"174", "name":"Unknown_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":4, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac251689f, ext:1469468, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"184", "name":"End-to-End_Error", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac25180f1, ext:1475695, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"187", "name":"Reported_Uncorrect", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac251969e, ext:1481243, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"188", "name":"Command_Timeout", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac251b24f, ext:1488330, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O---K", "id":"194", "name":"Temperature_Celsius", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":34, "value":66, "worst":65}, Time:time.Time{wall:0xbf2840bac251cab5, ext:1494577, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"199", "name":"UDMA_CRC_Error_Count", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac251e1b9, ext:1500470, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"230", "name":"Unknown_SSD_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":2229110374919, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac251f98f, ext:1506573, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"PO--CK", "id":"232", "name":"Available_Reservd_Space", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":100, "threshold":4, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac25220cd, ext:1516619, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"233", "name":"Media_Wearout_Indicator", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":3129, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac252381d, ext:1522585, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"234", "name":"Unknown_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":7444, "value":100, "worst":100}, Time:time.Time{wall:0xbf2840bac2525a30, ext:1531316, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"----CK", "id":"241", "name":"Total_LBAs_Written", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":4812, "value":253, "worst":253}, Time:time.Time{wall:0xbf2840bac2528231, ext:1541550, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"----CK", "id":"242", "name":"Total_LBAs_Read", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":671, "value":253, "worst":253}, Time:time.Time{wall:0xbf2840bac25298eb, ext:1547367, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "fail":"-", "flags":"-O--CK", "id":"244", "name":"Unknown_Attribute", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "raw_value":0, "value":0, "worst":100}, Time:time.Time{wall:0xbf2840bac252ae29, ext:1552807, loc:(*time.Location)(0x9ce2a0)}}
	// &testutil.Metric{Measurement:"smart_device", Tags:map[string]string{"device":".", "enabled":"Enabled", "model":"SanDisk Ultra II 240GB", "serial_no":"XXXXXXXX", "wwn":"XXXXXXXX"}, Fields:map[string]interface {}{"exit_status":0, "health_ok":true, "temp_c":34, "udma_crc_errors":0}, Time:time.Time{wall:0xbf2840bac252fe0e, ext:1573258, loc:(*time.Location)(0x9ce2a0)}}
}

// smartctl output
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
Temp$rature Warning:  Disabled or Not Supported

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
)
