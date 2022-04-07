package extr

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func MustMetric(v telegraf.Metric, err error) telegraf.Metric {
	if err != nil {
		panic(err)
	}
	return v
}

func TestSerializeMetricFloat(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.Serialize(m)
	assert.NoError(t, err)
	expS := []byte(fmt.Sprintf(`{"fields":[{"usage_idle":91.5}],"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
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
			expected:       `{"fields":[{"value":42}],"name":"cpu","tags":{},"timestamp":1525478795}`,
		},
		{
			name:           "1ns",
			timestampUnits: 1 * time.Nanosecond,
			expected:       `{"fields":[{"value":42}],"name":"cpu","tags":{},"timestamp":1525478795123456789}`,
		},
		{
			name:           "1ms",
			timestampUnits: 1 * time.Millisecond,
			expected:       `{"fields":[{"value":42}],"name":"cpu","tags":{},"timestamp":1525478795123}`,
		},
		{
			name:           "10ms",
			timestampUnits: 10 * time.Millisecond,
			expected:       `{"fields":[{"value":42}],"name":"cpu","tags":{},"timestamp":152547879512}`,
		},
		{
			name:           "15ms is reduced to 10ms",
			timestampUnits: 15 * time.Millisecond,
			expected:       `{"fields":[{"value":42}],"name":"cpu","tags":{},"timestamp":152547879512}`,
		},
		{
			name:           "65ms is reduced to 10ms",
			timestampUnits: 65 * time.Millisecond,
			expected:       `{"fields":[{"value":42}],"name":"cpu","tags":{},"timestamp":152547879512}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42.0,
				},
				time.Unix(1525478795, 123456789),
			)
			s, _ := NewSerializer(tt.timestampUnits)
			actual, err := s.Serialize(m)
			require.NoError(t, err)
			require.Equal(t, tt.expected+"\n", string(actual))
		})
	}
}

func TestSerializeMetricInt(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":[{"usage_idle":90}],"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMetricString(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": "foobar",
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":[{"usage_idle":"foobar"}],"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMultiFields(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle":  int64(90),
		"usage_total": 8559615,
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":[{"usage_idle":90,"usage_total":8559615}],"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMetricWithEscapes(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu tag": "cpu0",
	}
	fields := map[string]interface{}{
		"U,age=Idle": int64(90),
	}
	m := metric.New("My CPU", tags, fields, now)

	s, _ := NewSerializer(0)
	buf, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":[{"U,age=Idle":90}],"name":"My CPU","tags":{"cpu tag":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatch(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)

	metrics := []telegraf.Metric{m, m}
	s, _ := NewSerializer(0)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`[{"fields":[{"value":42},{"value":42}],"name":"cpu","tags":{},"timestamp":0}]`), buf)
}

func TestSerializeBatchNameDiff(t *testing.T) {
	m1 := metric.New(
		"StatsCpu",
		map[string]string{},
		map[string]interface{}{
			"cpu": 0,
			"min": 20,
			"max": 30,
			"avg": 25,
		},
		time.Unix(0, 0),
	)

	m2 := metric.New(
		"StatsCpu",
		map[string]string{},
		map[string]interface{}{
			"cpu": 1,
			"min": 34,
			"max": 55,
			"avg": 40,
		},
		time.Unix(0, 0),
	)
	m3 := metric.New(
		"EventInterfaceStatus",
		map[string]string{
			"node": "NODE1",
		},
		map[string]interface{}{
			"ifIndex":     1001,
			"port":        "1:1",
			"adminStatus": 1,
			"operStatus":  1,
		},
		time.Unix(0, 0),
	)
	m4 := metric.New(
		"EventInterfaceStatus",
		map[string]string{
			"node": "NODE1",
		},
		map[string]interface{}{
			"ifIndex":     1002,
			"port":        "1:2",
			"adminStatus": 1,
			"operStatus":  0,
		},
		time.Unix(0, 0),
	)

	metrics := []telegraf.Metric{m1, m2, m3, m4}
	s, _ := NewSerializer(0)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`[{"fields":[{"avg":25,"cpu":0,"max":30,"min":20},{"avg":40,"cpu":1,"max":55,"min":34}],"name":"StatsCpu","tags":{},"timestamp":0},{"fields":[{"adminStatus":1,"ifIndex":1001,"operStatus":1,"port":"1:1"},{"adminStatus":1,"ifIndex":1002,"operStatus":0,"port":"1:2"}],"name":"EventInterfaceStatus","tags":{"node":"NODE1"},"timestamp":0}]`), buf)

}

func TestSerializeBatchTagDiff(t *testing.T) {
	m1 := metric.New(
		"StatsCpu",
		map[string]string{
			"node": "NODE1",
		},
		map[string]interface{}{
			"cpu": 0,
			"min": 20,
			"max": 30,
			"avg": 25,
		},
		time.Unix(0, 0),
	)

	m2 := metric.New(
		"StatsCpu",
		map[string]string{
			"node": "NODE1",
		},
		map[string]interface{}{
			"cpu": 1,
			"min": 34,
			"max": 55,
			"avg": 40,
		},
		time.Unix(0, 0),
	)
	m3 := metric.New(
		"StatsCpu",
		map[string]string{
			"node": "NODE2",
		},
		map[string]interface{}{
			"cpu": 0,
			"min": 31,
			"max": 99,
			"avg": 59,
		},
		time.Unix(0, 0),
	)
	m4 := metric.New(
		"EventInterfaceStatus",
		map[string]string{
			"node": "NODE1",
		},
		map[string]interface{}{
			"ifIndex":     1001,
			"port":        "1:1",
			"adminStatus": 1,
			"operStatus":  1,
		},
		time.Unix(0, 0),
	)
	m5 := metric.New(
		"EventInterfaceStatus",
		map[string]string{
			"node": "NODE1",
		},
		map[string]interface{}{
			"ifIndex":     1002,
			"port":        "1:2",
			"adminStatus": 1,
			"operStatus":  0,
		},
		time.Unix(0, 0),
	)

	metrics := []telegraf.Metric{m1, m2, m3, m4, m5}
	s, _ := NewSerializer(0)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`[{"fields":[{"avg":25,"cpu":0,"max":30,"min":20},{"avg":40,"cpu":1,"max":55,"min":34}],"name":"StatsCpu","tags":{"node":"NODE1"},"timestamp":0},{"fields":[{"avg":59,"cpu":0,"max":99,"min":31}],"name":"StatsCpu","tags":{"node":"NODE2"},"timestamp":0},{"fields":[{"adminStatus":1,"ifIndex":1001,"operStatus":1,"port":"1:1"},{"adminStatus":1,"ifIndex":1002,"operStatus":0,"port":"1:2"}],"name":"EventInterfaceStatus","tags":{"node":"NODE1"},"timestamp":0}]`), buf)

}

func TestSerializeBatchTimestampDiff(t *testing.T) {
	m1 := metric.New(
		"StatsCpu",
		map[string]string{
			"node": "NODE1",
		},
		map[string]interface{}{
			"cpu": 0,
			"min": 20,
			"max": 30,
			"avg": 25,
		},
		time.Unix(1525478795, 123456789),
	)

	m2 := metric.New(
		"StatsCpu",
		map[string]string{
			"node": "NODE1",
		},
		map[string]interface{}{
			"cpu": 1,
			"min": 34,
			"max": 55,
			"avg": 40,
		},
		time.Unix(2525478795, 123456789),
	)
	m3 := metric.New(
		"StatsCpu",
		map[string]string{
			"node": "NODE1",
		},
		map[string]interface{}{
			"cpu": 2,
			"min": 31,
			"max": 99,
			"avg": 59,
		},
		time.Unix(2525478795, 123456789),
	)
	m4 := metric.New(
		"EventInterfaceStatus",
		map[string]string{
			"node": "NODE1",
		},
		map[string]interface{}{
			"ifIndex":     1001,
			"port":        "1:1",
			"adminStatus": 1,
			"operStatus":  1,
		},
		time.Unix(3525478795, 123456789),
	)
	m5 := metric.New(
		"EventInterfaceStatus",
		map[string]string{
			"node": "NODE1",
		},
		map[string]interface{}{
			"ifIndex":     1002,
			"port":        "1:2",
			"adminStatus": 1,
			"operStatus":  0,
		},
		time.Unix(4525478795, 123456789),
	)

	metrics := []telegraf.Metric{m1, m2, m3, m4, m5}
	s, _ := NewSerializer(0)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`[{"fields":[{"avg":25,"cpu":0,"max":30,"min":20}],"name":"StatsCpu","tags":{"node":"NODE1"},"timestamp":1525478795},{"fields":[{"avg":40,"cpu":1,"max":55,"min":34},{"avg":59,"cpu":2,"max":99,"min":31}],"name":"StatsCpu","tags":{"node":"NODE1"},"timestamp":2525478795},{"fields":[{"adminStatus":1,"ifIndex":1001,"operStatus":1,"port":"1:1"}],"name":"EventInterfaceStatus","tags":{"node":"NODE1"},"timestamp":3525478795},{"fields":[{"adminStatus":1,"ifIndex":1002,"operStatus":0,"port":"1:2"}],"name":"EventInterfaceStatus","tags":{"node":"NODE1"},"timestamp":4525478795}]`), buf)

}

func TestSerializeBatchSkipInf(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"inf":       math.Inf(1),
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	}

	s, err := NewSerializer(0)
	require.NoError(t, err)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`[{"fields":[{"time_idle":42}],"name":"cpu","tags":{},"timestamp":0}]`), buf)
}

func TestSerializeBatchSkipInfAllFields(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"inf": math.Inf(1),
			},
			time.Unix(0, 0),
		),
	}

	s, err := NewSerializer(0)
	require.NoError(t, err)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`[{"fields":[{}],"name":"cpu","tags":{},"timestamp":0}]`), buf)
}
