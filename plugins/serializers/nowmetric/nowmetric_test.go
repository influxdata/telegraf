package nowmetric

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)
	//expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":91.5},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	expS := []byte(fmt.Sprintf(`[ { "metric_type": "usage_idle", "resource": "", "node": "", "value": 91.5, "timestamp": %d, "ci2metric_id": { "node": "" }, "source": "Telegraf" } ]`, now.Unix()*1000) + "\n")
	assert.Equal(t, string(expS), string(buf))
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
			//expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":1525478795123}`,
			expected: `[ { "metric_type": "value", "resource": "", "node": "", "value": 42, "timestamp": 1525478795123000, "ci2metric_id": { "node": "" }, "source": "Telegraf" } ]`,
		},
		{
			name:           "10ms",
			timestampUnits: 10 * time.Millisecond,
			//expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":152547879512}`,
			expected: `[ { "metric_type": "value", "resource": "", "node": "", "value": 42, "timestamp": 152547879512000, "ci2metric_id": { "node": "" }, "source": "Telegraf" } ]`,
		},
		{
			name:           "15ms is reduced to 10ms",
			timestampUnits: 15 * time.Millisecond,
			//expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":152547879512}`,
			expected: `[ { "metric_type": "value", "resource": "", "node": "", "value": 42, "timestamp": 152547879512000, "ci2metric_id": { "node": "" }, "source": "Telegraf" } ]`,
		},
		{
			name:           "65ms is reduced to 10ms",
			timestampUnits: 65 * time.Millisecond,
			//expected:       `{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":152547879512}`,
			expected: `[ { "metric_type": "value", "resource": "", "node": "", "value": 42, "timestamp": 152547879512000, "ci2metric_id": { "node": "" }, "source": "Telegraf" } ]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MustMetric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(1525478795, 123456789),
				),
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	//expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":90},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	expS := []byte(fmt.Sprintf(`[ { "metric_type": "usage_idle", "resource": "", "node": "", "value": 90, "timestamp": %d, "ci2metric_id": { "node": "" }, "source": "Telegraf" } ]`, now.Unix()*1000) + "\n")
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	//expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":"foobar"},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	expS := []byte(fmt.Sprintf(`[  ]`) + "\n")
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	//expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":90,"usage_total":8559615},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	expS := []byte(fmt.Sprintf(`[ { "metric_type": "usage_idle", "resource": "", "node": "", "value": 90, "timestamp": %d, "ci2metric_id": { "node": "" }, "source": "Telegraf" },`+"\n"+`{ "metric_type": "usage_total", "resource": "", "node": "", "value": 8559615, "timestamp": %d, "ci2metric_id": { "node": "" }, "source": "Telegraf" } ]`, now.Unix()*1000, now.Unix()*1000) + "\n")
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
	m, err := metric.New("My CPU", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(0)
	buf, err := s.Serialize(m)
	assert.NoError(t, err)

	//expS := []byte(fmt.Sprintf(`{"fields":{"U,age=Idle":90},"name":"My CPU","tags":{"cpu tag":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	expS := []byte(fmt.Sprintf(`[ { "metric_type": "U,age=Idle", "resource": "", "node": "", "value": 90, "timestamp": %d, "ci2metric_id": { "node": "" }, "source": "Telegraf" } ]`, now.Unix()*1000) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatch(t *testing.T) {
	m := MustMetric(
		metric.New(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"value": 42.0,
			},
			time.Unix(0, 0),
		),
	)

	metrics := []telegraf.Metric{m, m}
	s, _ := NewSerializer(0)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	//require.Equal(t, []byte(`{"metrics":[{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":0},{"fields":{"value":42},"name":"cpu","tags":{},"timestamp":0}]}`), buf)
	require.Equal(t, []byte(`[ { "metric_type": "value", "resource": "", "node": "", "value": 42, "timestamp": 0, "ci2metric_id": { "node": "" }, "source": "Telegraf" } ]`+"\n"+`[ { "metric_type": "value", "resource": "", "node": "", "value": 42, "timestamp": 0, "ci2metric_id": { "node": "" }, "source": "Telegraf" } ]`+"\n"), buf)
}
