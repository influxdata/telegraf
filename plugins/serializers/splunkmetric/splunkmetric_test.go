package splunkmetric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func TestSerializeMetricFloat(t *testing.T) {
	// Test sub-second time
	now := time.Unix(1529875740, 819000000)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(false, false)
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)
	expS := `{"_value":91.5,"cpu":"cpu0","metric_name":"cpu.usage_idle","time":1529875740.819}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMetricFloatHec(t *testing.T) {
	// Test sub-second time
	now := time.Unix(1529875740, 819000000)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(true, false)
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)
	expS := `{"time":1529875740.819,"fields":{"_value":91.5,"cpu":"cpu0","metric_name":"cpu.usage_idle"}}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMetricInt(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(false, false)
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := `{"_value":90,"cpu":"cpu0","metric_name":"cpu.usage_idle","time":0}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMetricIntHec(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(true, false)
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := `{"time":0,"fields":{"_value":90,"cpu":"cpu0","metric_name":"cpu.usage_idle"}}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMetricBool(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"container-name": "telegraf-test",
	}
	fields := map[string]interface{}{
		"oomkiller": true,
	}
	m := metric.New("docker", tags, fields, now)

	s, _ := NewSerializer(false, false)
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := `{"_value":1,"container-name":"telegraf-test","metric_name":"docker.oomkiller","time":0}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMetricBoolHec(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"container-name": "telegraf-test",
	}
	fields := map[string]interface{}{
		"oomkiller": false,
	}
	m := metric.New("docker", tags, fields, now)

	s, _ := NewSerializer(true, false)
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := `{"time":0,"fields":{"_value":0,"container-name":"telegraf-test","metric_name":"docker.oomkiller"}}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMetricString(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"processorType": "ARMv7 Processor rev 4 (v7l)",
		"usage_idle":    int64(5),
	}
	m := metric.New("cpu", tags, fields, now)

	s, _ := NewSerializer(false, false)
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := `{"_value":5,"cpu":"cpu0","metric_name":"cpu.usage_idle","time":0}`
	require.Equal(t, expS, string(buf))
	require.NoError(t, err)
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

	n := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 92.0,
		},
		time.Unix(0, 0),
	)

	metrics := []telegraf.Metric{m, n}
	s, _ := NewSerializer(false, false)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)

	expS := `{"_value":42,"metric_name":"cpu.value","time":0}{"_value":92,"metric_name":"cpu.value","time":0}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMulti(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"user":   42.0,
			"system": 8.0,
		},
		time.Unix(0, 0),
	)

	metrics := []telegraf.Metric{m}
	s, _ := NewSerializer(false, true)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)

	expS := `{"metric_name:cpu.system":8,"metric_name:cpu.user":42,"time":0}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeBatchHec(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	n := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 92.0,
		},
		time.Unix(0, 0),
	)
	metrics := []telegraf.Metric{m, n}
	s, _ := NewSerializer(true, false)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)

	expS := `{"time":0,"fields":{"_value":42,"metric_name":"cpu.value"}}{"time":0,"fields":{"_value":92,"metric_name":"cpu.value"}}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMultiHec(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"usage":  42.0,
			"system": 8.0,
		},
		time.Unix(0, 0),
	)

	metrics := []telegraf.Metric{m}
	s, _ := NewSerializer(true, true)
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)

	expS := `{"time":0,"fields":{"metric_name:cpu.system":8,"metric_name:cpu.usage":42}}`
	require.Equal(t, expS, string(buf))
}
