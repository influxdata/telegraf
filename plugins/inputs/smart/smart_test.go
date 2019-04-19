package smart

import (
	"errors"
	"fmt"
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

	// tags = map[string]string{}

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
	for i := range acc.Metrics {
		fmt.Printf("%+#v\n\n", acc.Metrics[i])
		/*
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"1", "name":"Raw_Read_Error_Rate", "flags":"PO-R--", "fail":"-", "device":"."}, Fields:map[string]interface {}{"threshold":62, "raw_value":0, "exit_status":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf26ac0863cded17, ext:6344670, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"id":"2", "name":"Throughput_Performance", "flags":"P-S---", "fail":"-", "device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"threshold":40, "raw_value":0, "exit_status":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf26ac0863ce337f, ext:6362693, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"fail":"-", "device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"3", "name":"Spin_Up_Time", "flags":"POS---"}, Fields:map[string]interface {}{"raw_value":1, "exit_status":0, "value":100, "worst":100, "threshold":33}, Time:time.Time{wall:0xbf26ac0863ce6113, ext:6374360, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"4", "name":"Start_Stop_Count", "flags":"-O--C-", "fail":"-"}, Fields:map[string]interface {}{"value":100, "worst":100, "threshold":0, "raw_value":4, "exit_status":0}, Time:time.Time{wall:0xbf26ac0863ce8b8e, ext:6385239, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"5", "name":"Reallocated_Sector_Ct", "flags":"PO--CK", "fail":"-"}, Fields:map[string]interface {}{"worst":100, "threshold":5, "raw_value":0, "exit_status":0, "value":100}, Time:time.Time{wall:0xbf26ac0863ceb1d5, ext:6395036, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"7", "name":"Seek_Error_Rate", "flags":"PO-R--", "fail":"-"}, Fields:map[string]interface {}{"threshold":67, "raw_value":0, "exit_status":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf26ac0863cee889, ext:6409040, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"fail":"-", "device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"8", "name":"Seek_Time_Performance", "flags":"P-S---"}, Fields:map[string]interface {}{"value":100, "worst":100, "threshold":40, "raw_value":0, "exit_status":0}, Time:time.Time{wall:0xbf26ac0863cf1198, ext:6419547, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"9", "name":"Power_On_Hours", "flags":"-O--C-", "fail":"-", "device":"."}, Fields:map[string]interface {}{"raw_value":743, "exit_status":0, "value":99, "worst":99, "threshold":0}, Time:time.Time{wall:0xbf26ac0863cf3ad7, ext:6430134, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"id":"10", "name":"Spin_Retry_Count", "flags":"PO--C-", "fail":"-", "device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b"}, Fields:map[string]interface {}{"exit_status":0, "value":100, "worst":100, "threshold":60, "raw_value":0}, Time:time.Time{wall:0xbf26ac0863cf77e8, ext:6445741, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"wwn":"5000cca90bc3a98b", "id":"12", "name":"Power_Cycle_Count", "flags":"-O--CK", "fail":"-", "device":".", "serial_no":"RCE50G20G81S9S"}, Fields:map[string]interface {}{"threshold":0, "raw_value":4, "exit_status":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf26ac0863cfb250, ext:6460695, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"name":"G-Sense_Error_Rate", "flags":"-O-R--", "fail":"-", "device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"191"}, Fields:map[string]interface {}{"value":100, "worst":100, "threshold":0, "raw_value":0, "exit_status":0}, Time:time.Time{wall:0xbf26ac0863cfdbd1, ext:6471314, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"192", "name":"Power-Off_Retract_Count", "flags":"-O--CK", "fail":"-"}, Fields:map[string]interface {}{"value":100, "worst":100, "threshold":0, "raw_value":2, "exit_status":0}, Time:time.Time{wall:0xbf26ac0863d003aa, ext:6481521, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"wwn":"5000cca90bc3a98b", "id":"193", "name":"Load_Cycle_Count", "flags":"-O--C-", "fail":"-", "device":".", "serial_no":"RCE50G20G81S9S"}, Fields:map[string]interface {}{"threshold":0, "raw_value":13, "exit_status":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf26ac0863d0471f, ext:6498782, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"name":"Temperature_Celsius", "flags":"-O----", "fail":"-", "device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"194"}, Fields:map[string]interface {}{"worst":250, "threshold":0, "raw_value":24, "exit_status":0, "value":250}, Time:time.Time{wall:0xbf26ac0863d0748b, ext:6510415, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"wwn":"5000cca90bc3a98b", "id":"196", "name":"Reallocated_Event_Count", "flags":"-O--CK", "fail":"-", "device":".", "serial_no":"RCE50G20G81S9S"}, Fields:map[string]interface {}{"value":100, "worst":100, "threshold":0, "raw_value":0, "exit_status":0}, Time:time.Time{wall:0xbf26ac0863d09d67, ext:6520877, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"197", "name":"Current_Pending_Sector", "flags":"-O---K", "fail":"-"}, Fields:map[string]interface {}{"threshold":0, "raw_value":0, "exit_status":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf26ac0863d0d730, ext:6535667, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"name":"Offline_Uncorrectable", "flags":"---R--", "fail":"-", "device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"198"}, Fields:map[string]interface {}{"exit_status":0, "value":100, "worst":100, "threshold":0, "raw_value":0}, Time:time.Time{wall:0xbf26ac0863d10020, ext:6546152, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"199", "name":"UDMA_CRC_Error_Count", "flags":"-O-R--", "fail":"-", "device":"."}, Fields:map[string]interface {}{"threshold":0, "raw_value":0, "exit_status":0, "value":200, "worst":200}, Time:time.Time{wall:0xbf26ac0863d1282d, ext:6556404, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "id":"223", "name":"Load_Retry_Count", "flags":"-O-R--", "fail":"-"}, Fields:map[string]interface {}{"raw_value":0, "exit_status":0, "value":100, "worst":100, "threshold":0}, Time:time.Time{wall:0xbf26ac0863d14f65, ext:6566441, loc:(*time.Location)(0x946040)}}
			&testutil.Metric{Measurement:"smart_device", Tags:map[string]string{"device":".", "model":"HGST HTE725050A7E630", "serial_no":"RCE50G20G81S9S", "wwn":"5000cca90bc3a98b", "capacity":"500107862016", "enabled":"Enabled"}, Fields:map[string]interface {}{"read_error_rate":0, "seek_error_rate":0, "temp_c":24, "udma_crc_errors":0, "exit_status":0, "health_ok":true}, Time:time.Time{wall:0xbf26ac0863d1dfa1, ext:6603367, loc:(*time.Location)(0x946040)}}
		*/
	}
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
	for i := range acc.Metrics {
		fmt.Printf("%+#v\n\n", acc.Metrics[i])
		/*
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"1", "name":"Raw_Read_Error_Rate", "flags":"PO-R--", "fail":"-"}, Fields:map[string]interface {}{"threshold":16, "raw_value":0, "exit_status":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf26ac45a96cddc1, ext:1907088, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"2", "name":"Throughput_Performance", "flags":"P-S---", "fail":"-", "device":"."}, Fields:map[string]interface {}{"value":135, "worst":135, "threshold":54, "raw_value":84, "exit_status":0}, Time:time.Time{wall:0xbf26ac45a96d102a, ext:1919985, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"id":"3", "name":"Spin_Up_Time", "flags":"POS---", "fail":"-", "device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c"}, Fields:map[string]interface {}{"worst":125, "threshold":24, "raw_value":621, "exit_status":0, "value":125}, Time:time.Time{wall:0xbf26ac45a96d2eb6, ext:1927804, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"flags":"-O--C-", "fail":"-", "device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"4", "name":"Start_Stop_Count"}, Fields:map[string]interface {}{"threshold":0, "raw_value":33, "exit_status":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf26ac45a96d4851, ext:1934357, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"fail":"-", "device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"5", "name":"Reallocated_Sector_Ct", "flags":"PO--CK"}, Fields:map[string]interface {}{"worst":100, "threshold":5, "raw_value":0, "exit_status":0, "value":100}, Time:time.Time{wall:0xbf26ac45a96d6055, ext:1940506, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"id":"7", "name":"Seek_Error_Rate", "flags":"PO-R--", "fail":"-", "device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c"}, Fields:map[string]interface {}{"raw_value":0, "exit_status":0, "value":100, "worst":100, "threshold":67}, Time:time.Time{wall:0xbf26ac45a96d8d78, ext:1952061, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"8", "name":"Seek_Time_Performance", "flags":"P-S---", "fail":"-", "device":"."}, Fields:map[string]interface {}{"exit_status":0, "value":119, "worst":119, "threshold":20, "raw_value":35}, Time:time.Time{wall:0xbf26ac45a96da608, ext:1958349, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"name":"Power_On_Hours", "flags":"-O--C-", "fail":"-", "device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"9"}, Fields:map[string]interface {}{"threshold":0, "raw_value":19371, "exit_status":0, "value":98, "worst":98}, Time:time.Time{wall:0xbf26ac45a96dbf4e, ext:1964821, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"wwn":"5000cca250ec3c9c", "id":"10", "name":"Spin_Retry_Count", "flags":"PO--C-", "fail":"-", "device":".", "serial_no":"PK1334PEK49SBS"}, Fields:map[string]interface {}{"exit_status":0, "value":100, "worst":100, "threshold":60, "raw_value":0}, Time:time.Time{wall:0xbf26ac45a96dfa0e, ext:1979862, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"wwn":"5000cca250ec3c9c", "id":"12", "name":"Power_Cycle_Count", "flags":"-O--CK", "fail":"-", "device":".", "serial_no":"PK1334PEK49SBS"}, Fields:map[string]interface {}{"exit_status":0, "value":100, "worst":100, "threshold":0, "raw_value":33}, Time:time.Time{wall:0xbf26ac45a96e22b2, ext:1990262, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"flags":"-O--CK", "fail":"-", "device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"192", "name":"Power-Off_Retract_Count"}, Fields:map[string]interface {}{"worst":100, "threshold":0, "raw_value":764, "exit_status":0, "value":100}, Time:time.Time{wall:0xbf26ac45a96e3be7, ext:1996718, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"193", "name":"Load_Cycle_Count", "flags":"-O--C-", "fail":"-"}, Fields:map[string]interface {}{"threshold":0, "raw_value":764, "exit_status":0, "value":100, "worst":100}, Time:time.Time{wall:0xbf26ac45a96e548e, ext:2003027, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"name":"Temperature_Celsius", "flags":"-O----", "fail":"-", "device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"194"}, Fields:map[string]interface {}{"worst":176, "threshold":0, "raw_value":34, "exit_status":0, "value":176}, Time:time.Time{wall:0xbf26ac45a96e854b, ext:2015506, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"196", "name":"Reallocated_Event_Count", "flags":"-O--CK", "fail":"-"}, Fields:map[string]interface {}{"raw_value":0, "exit_status":0, "value":100, "worst":100, "threshold":0}, Time:time.Time{wall:0xbf26ac45a96e9dc5, ext:2021770, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"197", "name":"Current_Pending_Sector", "flags":"-O---K", "fail":"-"}, Fields:map[string]interface {}{"raw_value":0, "exit_status":0, "value":100, "worst":100, "threshold":0}, Time:time.Time{wall:0xbf26ac45a96eb417, ext:2027485, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"198", "name":"Offline_Uncorrectable", "flags":"---R--", "fail":"-"}, Fields:map[string]interface {}{"raw_value":0, "exit_status":0, "value":100, "worst":100, "threshold":0}, Time:time.Time{wall:0xbf26ac45a96ee2ee, ext:2039477, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_attribute", Tags:map[string]string{"flags":"-O-R--", "fail":"-", "device":".", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c", "id":"199", "name":"UDMA_CRC_Error_Count"}, Fields:map[string]interface {}{"exit_status":0, "value":200, "worst":200, "threshold":0, "raw_value":0}, Time:time.Time{wall:0xbf26ac45a96efcb8, ext:2046077, loc:(*time.Location)(0x947060)}}
			&testutil.Metric{Measurement:"smart_device", Tags:map[string]string{"capacity":"4000787030016", "enabled":"Enabled", "device":".", "model":"HGST HDN724040ALE640", "serial_no":"PK1334PEK49SBS", "wwn":"5000cca250ec3c9c"}, Fields:map[string]interface {}{"temp_c":34, "udma_crc_errors":0, "exit_status":0, "health_ok":true, "read_error_rate":0, "seek_error_rate":0}, Time:time.Time{wall:0xbf26ac45a96f4ef8, ext:2067135, loc:(*time.Location)(0x947060)}}
		*/
	}
}

func TestGatherHgstSAS(t *testing.T) {
	runCmd = func(sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(htSASInfoData), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, true, true, "", "", "", wg)
	for i := range acc.Metrics {
		fmt.Printf("%+#v\n", acc.Metrics[i])
		// &testutil.Metric{Measurement:"smart_device", Tags:map[string]string{"device":".", "enabled":"Enabled"}, Fields:map[string]interface {}{"exit_status":0}, Time:time.Time{wall:0xbf26ac62798dff44, ext:2175076, loc:(*time.Location)(0x9480a0)}}
	}
}

func TestGatherHtSAS(t *testing.T) {
	runCmd = func(sudo bool, command string, args ...string) ([]byte, error) {
		return []byte(hgstSASInfoData), nil
	}

	var (
		acc = &testutil.Accumulator{}
		wg  = &sync.WaitGroup{}
	)

	wg.Add(1)
	gatherDisk(acc, true, true, "", "", "", wg)
	for i := range acc.Metrics {
		fmt.Printf("%+#v\n", acc.Metrics[i])
		// &testutil.Metric{Measurement:"smart_device", Tags:map[string]string{"device":".", "capacity":"12000138625024", "enabled":"Enabled"}, Fields:map[string]interface {}{"exit_status":0}, Time:time.Time{wall:0xbf26ac5cd1522730, ext:2296199, loc:(*time.Location)(0x9480a0)}}
	}
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
)
