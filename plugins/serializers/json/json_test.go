package json

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestSerializeMetricFloat(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(0, "")
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)
	expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":91.5},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	require.Equal(t, string(expS), string(buf))
}

func TestSerialize_TimestampUnits(t *testing.T) {
	tests := []struct {
		name            string
		timestampUnits  time.Duration
		timestampFormat string
		expected        string
	}{
		{
			name:           "default of 1s",
			timestampUnits: 0,
			expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":1525478795}`,
		},
		{
			name:           "1ns",
			timestampUnits: 1 * time.Nanosecond,
			expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":1525478795123456789}`,
		},
		{
			name:           "1ms",
			timestampUnits: 1 * time.Millisecond,
			expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":1525478795123}`,
		},
		{
			name:           "10ms",
			timestampUnits: 10 * time.Millisecond,
			expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":152547879512}`,
		},
		{
			name:           "15ms is reduced to 10ms",
			timestampUnits: 15 * time.Millisecond,
			expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":152547879512}`,
		},
		{
			name:           "65ms is reduced to 10ms",
			timestampUnits: 65 * time.Millisecond,
			expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":152547879512}`,
		},
		{
			name:            "timestamp format",
			timestampFormat: "2006-01-02T15:04:05Z07:00",
			expected:        `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":"2018-05-05T00:06:35Z"}`,
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
			s, _ := NewSerializer(tt.timestampUnits, tt.timestampFormat)
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

	s, _ := NewSerializer(0, "")
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":90},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	require.Equal(t, string(expS), string(buf))
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

	s, _ := NewSerializer(0, "")
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":"foobar"},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	require.Equal(t, string(expS), string(buf))
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

	s, _ := NewSerializer(0, "")
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":90,"usage_total":8559615},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	require.Equal(t, string(expS), string(buf))
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

	s, _ := NewSerializer(0, "")
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":{"U,age=Idle":90},"name":"My CPU","tags":{"cpu tag":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	require.Equal(t, string(expS), string(buf))
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
	s, _ := NewSerializer(0, "")
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`{"metrics":[{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":0},{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":0}]}`), buf)
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

	s, err := NewSerializer(0, "")
	require.NoError(t, err)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`{"metrics":[{"fields":{"time_idle":42},"name":"cpu","tags":{},"timestamp":0}]}`), buf)
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

	s, err := NewSerializer(0, "")
	require.NoError(t, err)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`{"metrics":[{"fields":{},"name":"cpu","tags":{},"timestamp":0}]}`), buf)
}
