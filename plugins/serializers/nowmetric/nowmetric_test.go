package nowmetric

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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

	s, _ := NewSerializer()
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)
	expS := []byte(fmt.Sprintf(`[{"metric_type":"usage_idle","resource":"","node":"","value":91.5,"timestamp":%d,"ci2metric_id":null,"source":"Telegraf"}]`, now.UnixNano()/int64(time.Millisecond)))
	require.Equal(t, string(expS), string(buf))
}

func TestSerialize_TimestampUnits(t *testing.T) {
	tests := []struct {
		name           string
		timestampUnits time.Duration
		expected       string
	}{
		{
			name:           "1ms",
			timestampUnits: 1 * time.Millisecond,
			expected:       `[{"metric_type":"value","resource":"","node":"","value":42,"timestamp":1525478795123,"ci2metric_id":null,"source":"Telegraf"}]`,
		},
		{
			name:           "10ms",
			timestampUnits: 10 * time.Millisecond,
			expected:       `[{"metric_type":"value","resource":"","node":"","value":42,"timestamp":1525478795123,"ci2metric_id":null,"source":"Telegraf"}]`,
		},
		{
			name:           "15ms is reduced to 10ms",
			timestampUnits: 15 * time.Millisecond,
			expected:       `[{"metric_type":"value","resource":"","node":"","value":42,"timestamp":1525478795123,"ci2metric_id":null,"source":"Telegraf"}]`,
		},
		{
			name:           "65ms is reduced to 10ms",
			timestampUnits: 65 * time.Millisecond,
			expected:       `[{"metric_type":"value","resource":"","node":"","value":42,"timestamp":1525478795123,"ci2metric_id":null,"source":"Telegraf"}]`,
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
			s, _ := NewSerializer()
			actual, err := s.Serialize(m)
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(actual))
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

	s, _ := NewSerializer()
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := []byte(fmt.Sprintf(`[{"metric_type":"usage_idle","resource":"","node":"","value":90,"timestamp":%d,"ci2metric_id":null,"source":"Telegraf"}]`, now.UnixNano()/int64(time.Millisecond)))
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

	s, _ := NewSerializer()
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	require.Equal(t, "null", string(buf))
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

	// Sort for predictable field order
	sort.Slice(m.FieldList(), func(i, j int) bool {
		return m.FieldList()[i].Key < m.FieldList()[j].Key
	})

	s, _ := NewSerializer()
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := []byte(fmt.Sprintf(`[{"metric_type":"usage_idle","resource":"","node":"","value":90,"timestamp":%d,"ci2metric_id":null,"source":"Telegraf"},{"metric_type":"usage_total","resource":"","node":"","value":8559615,"timestamp":%d,"ci2metric_id":null,"source":"Telegraf"}]`, now.UnixNano()/int64(time.Millisecond), now.UnixNano()/int64(time.Millisecond)))
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

	s, _ := NewSerializer()
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := []byte(fmt.Sprintf(`[{"metric_type":"U,age=Idle","resource":"","node":"","value":90,"timestamp":%d,"ci2metric_id":null,"source":"Telegraf"}]`, now.UnixNano()/int64(time.Millisecond)))
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
	s, _ := NewSerializer()
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	require.Equal(t, []byte(`[{"metric_type":"value","resource":"","node":"","value":42,"timestamp":0,"ci2metric_id":null,"source":"Telegraf"},{"metric_type":"value","resource":"","node":"","value":42,"timestamp":0,"ci2metric_id":null,"source":"Telegraf"}]`), buf)
}
