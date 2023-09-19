package extr

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func MustMetric(v telegraf.Metric, err error) telegraf.Metric {
	if err != nil {
		panic(err)
	}
	return v
}

func TestSerializeBatchMetricFloat(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"core_key":  0,
		"usageIdle": float64(91.5),
	}
	field2 := map[string]interface{}{
		"core_key":  1,
		"usageIdle": float64(0.9999),
	}
	m1 := metric.New("CpuStats", tags, field1, now)
	m2 := metric.New("CpuStats", tags, field2, now)

	metrics := []telegraf.Metric{m1, m2}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)
	expS := []byte(fmt.Sprintf(`{"cpuStats":[{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"usageIdle":91.5},{"keys":{"core":1},"usageIdle":0.9999}],"name":"CpuStats","ts":%d}]}`, now.Unix()))
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatchMetricBool(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"core_key": 0,
		"mybool1":  true,
	}
	field2 := map[string]interface{}{
		"core_key": 1,
		"mybool1":  false,
	}
	m1 := metric.New("CpuStats", tags, field1, now)
	m2 := metric.New("CpuStats", tags, field2, now)

	metrics := []telegraf.Metric{m1, m2}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"cpuStats":[{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"mybool1":true},{"keys":{"core":1},"mybool1":false}],"name":"CpuStats","ts":%d}]}`, now.Unix()))
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatchMetricInt(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"core_key":  0,
		"usageIdle": int64(91),
	}
	field2 := map[string]interface{}{
		"core_key":  1,
		"usageIdle": int64(90),
	}
	m1 := metric.New("CpuStats", tags, field1, now)
	m2 := metric.New("CpuStats", tags, field2, now)

	metrics := []telegraf.Metric{m1, m2}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"cpuStats":[{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"usageIdle":91},{"keys":{"core":1},"usageIdle":90}],"name":"CpuStats","ts":%d}]}`, now.Unix()))
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatchMetricString(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"core_key":  0,
		"usageIdle": "foobar1",
	}
	field2 := map[string]interface{}{
		"core_key":  1,
		"usageIdle": "barfoo1",
	}
	m1 := metric.New("CpuStats", tags, field1, now)
	m2 := metric.New("CpuStats", tags, field2, now)

	metrics := []telegraf.Metric{m1, m2}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"cpuStats":[{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"usageIdle":"foobar1"},{"keys":{"core":1},"usageIdle":"barfoo1"}],"name":"CpuStats","ts":%d}]}`, now.Unix()))
	assert.Equal(t, string(expS), string(buf))
}

func TestSerialize_TimestampUnits(t *testing.T) {
	tests := []struct {
		name           string
		timestampUnits time.Duration
		expected       string
	}{
		{
			name:           "default of 1s",
			timestampUnits: 0,
			expected:       `{"cpuStats":[{"device":{},"items":[{"keys":{"core":1},"value":42}],"name":"CpuStats","ts":1525478795},{"device":{},"items":[{"keys":{"core":2},"value":43}],"name":"CpuStats","ts":1527778795}]}`,
		},
		{
			name:           "1ns",
			timestampUnits: 1 * time.Nanosecond,
			expected:       `{"cpuStats":[{"device":{},"items":[{"keys":{"core":1},"value":42}],"name":"CpuStats","ts":1525478795123456789},{"device":{},"items":[{"keys":{"core":2},"value":43}],"name":"CpuStats","ts":1527778795127756789}]}`,
		},
		{
			name:           "1ms",
			timestampUnits: 1 * time.Millisecond,
			expected:       `{"cpuStats":[{"device":{},"items":[{"keys":{"core":1},"value":42}],"name":"CpuStats","ts":1525478795123},{"device":{},"items":[{"keys":{"core":2},"value":43}],"name":"CpuStats","ts":1527778795127}]}`,
		},
		{
			name:           "10ms",
			timestampUnits: 10 * time.Millisecond,
			expected:       `{"cpuStats":[{"device":{},"items":[{"keys":{"core":1},"value":42}],"name":"CpuStats","ts":152547879512},{"device":{},"items":[{"keys":{"core":2},"value":43}],"name":"CpuStats","ts":152777879512}]}`,
		},
		{
			name:           "15ms is reduced to 10ms",
			timestampUnits: 15 * time.Millisecond,
			expected:       `{"cpuStats":[{"device":{},"items":[{"keys":{"core":1},"value":42}],"name":"CpuStats","ts":152547879512},{"device":{},"items":[{"keys":{"core":2},"value":43}],"name":"CpuStats","ts":152777879512}]}`,
		},
		{
			name:           "65ms is reduced to 10ms",
			timestampUnits: 65 * time.Millisecond,
			expected:       `{"cpuStats":[{"device":{},"items":[{"keys":{"core":1},"value":42}],"name":"CpuStats","ts":152547879512},{"device":{},"items":[{"keys":{"core":2},"value":43}],"name":"CpuStats","ts":152777879512}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m1 := metric.New(
				"CpuStats",
				map[string]string{},
				map[string]interface{}{
					"core_key": 1,
					"value":    42.0,
				},
				time.Unix(1525478795, 123456789),
			)
			m2 := metric.New(
				"CpuStats",
				map[string]string{},
				map[string]interface{}{
					"core_key": 2,
					"value":    43.0,
				},
				time.Unix(1527778795, 127756789),
			)
			s, _ := NewSerializer(tt.timestampUnits)
			metrics := []telegraf.Metric{m1, m2}
			actual, err := s.SerializeBatch(metrics)
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(actual))
		})
	}
}

func TestSerializeBatchSingleMetric(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"core_key":       0,
		"usage_min":      int64(2),
		"usage_max":      100,
		"usage_avg":      52.1,
		"partNumber_tag": "1647G-00129 800751-00-01",
		"revision_tag":   "01",
		"mystring":       "Elon Musk was here",
		"operStatus_old": 1,
		"operStatus_new": 0,
	}
	m1 := metric.New("CpuStats", tags, field1, now)

	metrics := []telegraf.Metric{m1}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"cpuStats":[{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"mystring":"Elon Musk was here","operStatus":{"new":0,"old":1},"tags":{"partNumber":"1647G-00129 800751-00-01","revision":"01"},"usage":{"avg":52.1,"max":100,"min":2}}],"name":"CpuStats","ts":%d}]}`, now.Unix()))

	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatchSingleMetricWithEscapes(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"core_key":         0,
		"usage_min":        int64(2),
		"usage_max":        100,
		"usage_avg":        52.1,
		"field with space": 99,
		"field with,comma": 38,
		"mystring":         "Elon Musk was here",
	}
	m1 := metric.New("Cpu Stats", tags, field1, now)

	metrics := []telegraf.Metric{m1}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"cpu Stats":[{"device":{"serialnumber":"ABC-123"},"items":[{"field with space":99,"field with,comma":38,"keys":{"core":0},"mystring":"Elon Musk was here","usage":{"avg":52.1,"max":100,"min":2}}],"name":"Cpu Stats","ts":%d}]}`, now.Unix()))

	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatchMultiFields(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber":         "ABC-123",
		"reporterSerialnumber": "XYZ-456",
	}
	field1 := map[string]interface{}{
		"core_key":  0,
		"usage_min": int64(2),
		"usage_max": 100,
		"usage_avg": 52.1,
		"mystring":  "Elon Musk was here",
	}
	field2 := map[string]interface{}{
		"core_key":  1,
		"usage_min": int64(10),
		"usage_max": 98,
		"usage_avg": 49.9998,
		"mystring":  "Jeff Bezos was here",
	}
	field3 := map[string]interface{}{
		"ifIndex_key":       1001,
		"name_key":          "1:1",
		"ifAdminStatus_old": 0,
		"ifAdminStatus_new": 1,
		"ifOperStatus_old":  0,
		"ifOperStatus_new":  1,
	}
	field4 := map[string]interface{}{
		"ifIndex_key":       1002,
		"name_key":          "1:2",
		"ifAdminStatus_old": 1,
		"ifAdminStatus_new": 0,
		"ifOperStatus_old":  1,
		"ifOperStatus_new":  0,
	}
	field5 := map[string]interface{}{
		"routerId_key":                     10,
		"neighborAddress_key":              "10.100.2.1",
		"neighborAddressLessInterface_key": 0,
		"neighborRouterId_key":             10,
		"name_vrf_key":                     "vrf-1",
		"id_vrf_key":                       100,
		"state_old":                        "2Way",
		"state_new":                        "Full",
		"reason":                           "ExchangeComplete",
	}
	field6 := map[string]interface{}{
		"slot_key":     1,
		"fan_key":      1,
		"tray_key":     2,
		"rpm_min":      10,
		"rpm_max":      99,
		"rpm_avg":      75,
		"cpu1_pwm_min": 10,
		"cpu2_pwm_max": 99,
		"cpu3_pwm_avg": 75,
	}
	field7 := map[string]interface{}{
		"ifIndex_key":           4022,
		"name_key":              "4:22",
		"if1_ifAdminStatus_old": 1,
		"if1_ifAdminStatus_new": 0,
		"ifOperStatus_old":      1,
		"ifOperStatus_new":      0,
		"if2_ifAdminStatus_old": 1,
		"if2_ifAdminStatus_new": 0,
	}
	m1 := metric.New("CpuStats", tags, field1, now)
	m2 := metric.New("CpuStats", tags, field2, now)
	m3 := metric.New("InterfaceStateChanged", tags, field3, now)
	m4 := metric.New("InterfaceStateChanged", tags, field4, now)
	m5 := metric.New("OspfNeighborStateChanged", tags, field5, now)
	m6 := metric.New("FanTestTwoLayerStats", tags, field6, now)
	m7 := metric.New("FanTestTwoLayerStats", tags, field7, now)

	metrics := []telegraf.Metric{m1, m2, m3, m4, m5, m6, m7}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"cpuStats":[{"device":{"reporterSerialnumber":"XYZ-456","serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"mystring":"Elon Musk was here","usage":{"avg":52.1,"max":100,"min":2}},{"keys":{"core":1},"mystring":"Jeff Bezos was here","usage":{"avg":49.9998,"max":98,"min":10}}],"name":"CpuStats","ts":%d}],"fanTestTwoLayerStats":[{"device":{"reporterSerialnumber":"XYZ-456","serialnumber":"ABC-123"},"items":[{"keys":{"fan":1,"slot":1,"tray":2},"pwm":{"cpu1":{"min":10},"cpu2":{"max":99},"cpu3":{"avg":75}},"rpm":{"avg":75,"max":99,"min":10}},{"ifAdminStatus":{"if1":{"new":0,"old":1},"if2":{"new":0,"old":1}},"ifOperStatus":{"new":0,"old":1},"keys":{"ifIndex":4022,"name":"4:22"}}],"name":"FanTestTwoLayerStats","ts":%d}],"interfaceStateChanged":[{"device":{"reporterSerialnumber":"XYZ-456","serialnumber":"ABC-123"},"items":[{"ifAdminStatus":{"new":1,"old":0},"ifOperStatus":{"new":1,"old":0},"keys":{"ifIndex":1001,"name":"1:1"}},{"ifAdminStatus":{"new":0,"old":1},"ifOperStatus":{"new":0,"old":1},"keys":{"ifIndex":1002,"name":"1:2"}}],"name":"InterfaceStateChanged","ts":%d}],"ospfNeighborStateChanged":[{"device":{"reporterSerialnumber":"XYZ-456","serialnumber":"ABC-123"},"items":[{"keys":{"neighborAddress":"10.100.2.1","neighborAddressLessInterface":0,"neighborRouterId":10,"routerId":10,"vrf":{"id":100,"name":"vrf-1"}},"reason":"ExchangeComplete","state":{"new":"Full","old":"2Way"}}],"name":"OspfNeighborStateChanged","ts":%d}]}`, now.Unix(), now.Unix(), now.Unix(), now.Unix()))

	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatchMultiGroups(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"core_key":  0,
		"usage_min": int64(2),
		"usage_max": 100,
		"usage_avg": 52.1,
		"mystring":  "Elon Musk was here",
	}
	field2 := map[string]interface{}{
		"core_key":  1,
		"usage_min": int64(10),
		"usage_max": 98,
		"usage_avg": 49.9998,
		"mystring":  "Jeff Bezos was here",
	}
	m1 := metric.New("CpuStats", tags, field1, time.Unix(0, 0))
	m2 := metric.New("CpuStats", tags, field2, time.Unix(0, 0))
	m3 := metric.New("CpuStats", tags, field1, now)
	m4 := metric.New("CpuStats", tags, field2, now)

	metrics := []telegraf.Metric{m1, m2, m3, m4}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"cpuStats":[{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"mystring":"Elon Musk was here","usage":{"avg":52.1,"max":100,"min":2}},{"keys":{"core":1},"mystring":"Jeff Bezos was here","usage":{"avg":49.9998,"max":98,"min":10}}],"name":"CpuStats","ts":0},{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"mystring":"Elon Musk was here","usage":{"avg":52.1,"max":100,"min":2}},{"keys":{"core":1},"mystring":"Jeff Bezos was here","usage":{"avg":49.9998,"max":98,"min":10}}],"name":"CpuStats","ts":%d}]}`, now.Unix()))

	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatchMultiMetricTypesMultiGroups(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"core_key":  0,
		"usage_min": int64(2),
		"usage_max": 100,
		"usage_avg": 52.1,
	}
	field2 := map[string]interface{}{
		"core_key":       1,
		"usage_min":      int64(10),
		"usage_max":      98,
		"usage_avg":      49.9998,
		"partNumber_tag": "1647G-00129 800751-00-01",
		"revision_tag":   "01",
	}
	m1 := metric.New("CpuStats", tags, field1, time.Unix(0, 0))
	m2 := metric.New("CpuStats", tags, field2, time.Unix(0, 0))
	m3 := metric.New("CpuStats", tags, field1, time.Unix(100000000, 0))
	m4 := metric.New("CpuStats", tags, field2, time.Unix(100000000, 0))
	m5 := metric.New("MemoryStats", tags, field1, time.Unix(20000000, 0))
	m6 := metric.New("MemoryStats", tags, field2, time.Unix(20000000, 0))
	m7 := metric.New("CpuStats", tags, field1, time.Unix(550000000, 0))
	m8 := metric.New("CpuStats", tags, field2, time.Unix(550000000, 0))
	m9 := metric.New("MemoryStats", tags, field1, now)
	m10 := metric.New("CpuStats", tags, field2, now)

	metrics := []telegraf.Metric{m1, m2, m3, m4, m5, m6, m7, m8, m9, m10}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"cpuStats":[{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"usage":{"avg":52.1,"max":100,"min":2}},{"keys":{"core":1},"tags":{"partNumber":"1647G-00129 800751-00-01","revision":"01"},"usage":{"avg":49.9998,"max":98,"min":10}}],"name":"CpuStats","ts":0},{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"usage":{"avg":52.1,"max":100,"min":2}},{"keys":{"core":1},"tags":{"partNumber":"1647G-00129 800751-00-01","revision":"01"},"usage":{"avg":49.9998,"max":98,"min":10}}],"name":"CpuStats","ts":100000000},{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"usage":{"avg":52.1,"max":100,"min":2}},{"keys":{"core":1},"tags":{"partNumber":"1647G-00129 800751-00-01","revision":"01"},"usage":{"avg":49.9998,"max":98,"min":10}}],"name":"CpuStats","ts":550000000},{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":1},"tags":{"partNumber":"1647G-00129 800751-00-01","revision":"01"},"usage":{"avg":49.9998,"max":98,"min":10}}],"name":"CpuStats","ts":%d}],"memoryStats":[{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"usage":{"avg":52.1,"max":100,"min":2}},{"keys":{"core":1},"tags":{"partNumber":"1647G-00129 800751-00-01","revision":"01"},"usage":{"avg":49.9998,"max":98,"min":10}}],"name":"MemoryStats","ts":20000000},{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"usage":{"avg":52.1,"max":100,"min":2}}],"name":"MemoryStats","ts":%d}]}`, now.Unix(), now.Unix()))

	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatchMultiGroupsMultiLevel(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"core_key":               1,
		"cpu1_subcore_key":       2,
		"cpu2_subcore_key":       3,
		"cpu1_subcore_usage_old": "up",
		"cpu2_subcore_usage_new": "down",
		"usage_min":              22,
		"usage_max":              99,
		"usage_avg":              44,
		"xyz/_temp_min":          1,
		"xyz/_temp_avg":          22,
		"xyz/_temp_max":          100,
		"cpu1_subcore_usage_min": 1,
		"cpu1_subcore_usage_max": 100,
		"cpu1_subcore_usage_avg": 50,
		"cpu2_subcore_usage_min": 1,
		"cpu2_subcore_usage_max": 100,
		"cpu2_subcore_usage_avg": 50,
		"abc_name_subcore":       "foo",
		"xyz_name_subcore":       "bar",
	}
	m1 := metric.New("TestMultiLevelStats", tags, field1, now)

	metrics := []telegraf.Metric{m1}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"testMultiLevelStats":[{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":1,"subcore":{"cpu1":2,"cpu2":3}},"subcore":{"name":{"abc":"foo","xyz":"bar"}},"usage":{"avg":44,"max":99,"min":22,"subcore":{"cpu1":{"avg":50,"max":100,"min":1,"old":"up"},"cpu2":{"avg":50,"max":100,"min":1,"new":"down"}}},"xyz_temp":{"avg":22,"max":100,"min":1}}],"name":"TestMultiLevelStats","ts":%d}]}`, now.Unix()))

	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatchArrays1(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"@a_sysCapSupported_tag": "ROUTER",
	}

	m1 := metric.New("TestArrays1", tags, field1, now)

	metrics := []telegraf.Metric{m1}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"testArrays1":[{"device":{"serialnumber":"ABC-123"},"items":[{"tags":{"sysCapSupported":["ROUTER"]}}],"name":"TestArrays1","ts":%d}]}`, now.Unix()))

	assert.Equal(t, string(expS), string(buf))

}

func TestSerializeBatchArrays2(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"@a_sysCapSupported_tag": "ROUTER",
		"@b_sysCapSupported_tag": "BRIDGE",
	}

	m1 := metric.New("TestArrays2", tags, field1, now)

	metrics := []telegraf.Metric{m1}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	// For some reason, metric field order processing can vary, so depending
	// on order, the output can vary since array append will append
	// to slice in the order processed.  Need to account for differnt order
	expS1 := []byte(fmt.Sprintf(`{"testArrays2":[{"device":{"serialnumber":"ABC-123"},"items":[{"tags":{"sysCapSupported":["ROUTER","BRIDGE"]}}],"name":"TestArrays2","ts":%d}]}`, now.Unix()))
	
	expS2 := []byte(fmt.Sprintf(`{"testArrays2":[{"device":{"serialnumber":"ABC-123"},"items":[{"tags":{"sysCapSupported":["BRIDGE","ROUTER"]}}],"name":"TestArrays2","ts":%d}]}`, now.Unix()))

	if (string(expS1) != string(buf)) {
		if (string(expS2) != string(buf)) {

			fmt.Printf("--- NO MATCHES ---\n")
			fmt.Printf("ACTUAL:\n%v\n\n",string(buf))
			fmt.Printf("S1:\n%v\n\n",string(expS1))
			fmt.Printf("S2:\n%v\n\n",string(expS2))
			assert.Equal(t, string(expS1), string(buf))
		} else {
			fmt.Printf("--- MATCHES S2 ---\n")
			fmt.Printf("S2:\n%v\n\n",string(expS2))
		}
	} else {
		fmt.Printf("--- MATCHES S1 ---\n")
		fmt.Printf("S1:\n%v\n\n",string(expS1))
	}
}

func TestSerializeBatchArrays3(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"@1_sysCapSupported_tag": "ROUTER",
		"@2_sysCapSupported_tag": "BRIDGE",
		"@3_sysCapSupported_tag": "REPEATER",
	}

	m1 := metric.New("TestArrays3", tags, field1, now)

	metrics := []telegraf.Metric{m1}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	// For some reason, metric field order processing can vary, so depending
	// on order, the output can vary since array append will append
	// to slice in the order processed.  Need to account for differnt order

	expS1 := []byte(fmt.Sprintf(`{"testArrays3":[{"device":{"serialnumber":"ABC-123"},"items":[{"tags":{"sysCapSupported":["ROUTER","BRIDGE","REPEATER"]}}],"name":"TestArrays3","ts":%d}]}`, now.Unix()))
	
	expS2 := []byte(fmt.Sprintf(`{"testArrays3":[{"device":{"serialnumber":"ABC-123"},"items":[{"tags":{"sysCapSupported":["ROUTER","REPEATER","BRIDGE"]}}],"name":"TestArrays3","ts":%d}]}`, now.Unix()))

	expS3 := []byte(fmt.Sprintf(`{"testArrays3":[{"device":{"serialnumber":"ABC-123"},"items":[{"tags":{"sysCapSupported":["BRIDGE","ROUTER","REPEATER"]}}],"name":"TestArrays3","ts":%d}]}`, now.Unix()))

	expS4 := []byte(fmt.Sprintf(`{"testArrays3":[{"device":{"serialnumber":"ABC-123"},"items":[{"tags":{"sysCapSupported":["BRIDGE","REPEATER","ROUTER"]}}],"name":"TestArrays3","ts":%d}]}`, now.Unix()))

	expS5 := []byte(fmt.Sprintf(`{"testArrays3":[{"device":{"serialnumber":"ABC-123"},"items":[{"tags":{"sysCapSupported":["REPEATER","BRIDGE","ROUTER"]}}],"name":"TestArrays3","ts":%d}]}`, now.Unix()))

	expS6 := []byte(fmt.Sprintf(`{"testArrays3":[{"device":{"serialnumber":"ABC-123"},"items":[{"tags":{"sysCapSupported":["REPEATER","ROUTER","BRIDGE"]}}],"name":"TestArrays3","ts":%d}]}`, now.Unix()))


	if (string(expS1) != string(buf)) {
		if (string(expS2) != string(buf)) {
			if (string(expS3) != string(buf)) {
				if (string(expS4) != string(buf)) {
					if (string(expS5) != string(buf)) {
						if (string(expS6) != string(buf)) {

							fmt.Printf("--- NO MATCHES ---\n")
							fmt.Printf("ACTUAL:\n%v\n\n",string(buf))
							fmt.Printf("S1:\n%v\n\n",string(expS1))
							fmt.Printf("S2:\n%v\n\n",string(expS2))
							fmt.Printf("S3:\n%v\n\n",string(expS3))
							fmt.Printf("S4:\n%v\n\n",string(expS4))
							fmt.Printf("S5:\n%v\n\n",string(expS5))
							fmt.Printf("S6:\n%v\n\n",string(expS6))
			
							assert.Equal(t, string(expS1), string(buf))
						} else {
							fmt.Printf("--- MATCHES S6 ---\n")
							fmt.Printf("S6:\n%v\n\n",string(expS6))
						}
					} else {
						fmt.Printf("--- MATCHES S5 ---\n")
						fmt.Printf("S5:\n%v\n\n",string(expS5))
					}
				} else {
					fmt.Printf("--- MATCHES S4 ---\n")
					fmt.Printf("S4:\n%v\n\n",string(expS4))
				}
			} else {
				fmt.Printf("--- MATCHES S3---\n")
				fmt.Printf("S3:\n%v\n\n",string(expS3))
			}
		} else {
			fmt.Printf("--- MATCHES S2 ---\n")
			fmt.Printf("S2:\n%v\n\n",string(expS2))
		}
	} else {
		fmt.Printf("--- MATCHES S1 ---\n")
		fmt.Printf("S1:\n%v\n\n",string(expS1))
	}
}

func TestSerializeBatchArraysMiddle(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"type_@0_ipv6Addresses_ipv6Settings": "LinkLocalAddress",
	}

	m1 := metric.New("TestArraysMiddle", tags, field1, now)

	metrics := []telegraf.Metric{m1}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"testArraysMiddle":[{"device":{"serialnumber":"ABC-123"},"items":[{"ipv6Settings":{"ipv6Addresses":[{"type":"LinkLocalAddress"}]}}],"name":"TestArraysMiddle","ts":%d}]}`, now.Unix()))

	assert.Equal(t, string(expS), string(buf))

}


