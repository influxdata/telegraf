package extr

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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
		"core_key":   0,
		"usage_idle": float64(91.5),
	}
	field2 := map[string]interface{}{
		"core_key":   1,
		"usage_idle": float64(0.9999),
	}
	m1 := metric.New("CpuStats", tags, field1, now)
	m2 := metric.New("CpuStats", tags, field2, now)

	metrics := []telegraf.Metric{m1, m2}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)
	expS := []byte(fmt.Sprintf(`{"cpuStats":[{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"usage_idle":91.5},{"keys":{"core":1},"usage_idle":0.9999}],"name":"CpuStats","ts":%d}]}`, now.Unix()))
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
		"core_key":   0,
		"usage_idle": int64(91),
	}
	field2 := map[string]interface{}{
		"core_key":   1,
		"usage_idle": int64(90),
	}
	m1 := metric.New("CpuStats", tags, field1, now)
	m2 := metric.New("CpuStats", tags, field2, now)

	metrics := []telegraf.Metric{m1, m2}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"cpuStats":[{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"usage_idle":91},{"keys":{"core":1},"usage_idle":90}],"name":"CpuStats","ts":%d}]}`, now.Unix()))
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatchMetricString(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"serialnumber": "ABC-123",
	}
	field1 := map[string]interface{}{
		"core_key":   0,
		"usage_idle": "foobar1",
	}
	field2 := map[string]interface{}{
		"core_key":   1,
		"usage_idle": "barfoo1",
	}
	m1 := metric.New("CpuStats", tags, field1, now)
	m2 := metric.New("CpuStats", tags, field2, now)

	metrics := []telegraf.Metric{m1, m2}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"cpuStats":[{"device":{"serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"usage_idle":"foobar1"},{"keys":{"core":1},"usage_idle":"barfoo1"}],"name":"CpuStats","ts":%d}]}`, now.Unix()))
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
	m1 := metric.New("CpuStats", tags, field1, now)
	m2 := metric.New("CpuStats", tags, field2, now)
	m3 := metric.New("InterfaceStateChanged", tags, field3, now)
	m4 := metric.New("InterfaceStateChanged", tags, field4, now)

	metrics := []telegraf.Metric{m1, m2, m3, m4}

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"cpuStats":[{"device":{"reporterSerialnumber":"XYZ-456","serialnumber":"ABC-123"},"items":[{"keys":{"core":0},"mystring":"Elon Musk was here","usage":{"avg":52.1,"max":100,"min":2}},{"keys":{"core":1},"mystring":"Jeff Bezos was here","usage":{"avg":49.9998,"max":98,"min":10}}],"name":"CpuStats","ts":%d}],"interfaceStateChanged":[{"device":{"reporterSerialnumber":"XYZ-456","serialnumber":"ABC-123"},"items":[{"ifAdminStatus":{"new":1,"old":0},"ifOperStatus":{"new":1,"old":0},"keys":{"ifIndex":1001,"name":"1:1"}},{"ifAdminStatus":{"new":0,"old":1},"ifOperStatus":{"new":0,"old":1},"keys":{"ifIndex":1002,"name":"1:2"}}],"name":"InterfaceStateChanged","ts":%d}]}`, now.Unix(), now.Unix()))

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
