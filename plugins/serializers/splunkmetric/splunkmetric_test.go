package splunkmetric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
	// Test sub-second time
	now := time.Unix(1529875740, 819000000)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(false, false)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)
	expS := `{"_value":91.5,"config:hecRouting":false,"config:multiMetric":false,"cpu":"cpu0","metric_name":"cpu.usage_idle","time":1529875740.819}`
	assert.Equal(t, string(expS), string(buf))
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(true, false)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)
	expS := `{"time":1529875740.819,"event":"metric","fields":{"_value":91.5,"config:hecRouting":true,"config:multiMetric":false,"cpu":"cpu0","metric_name":"cpu.usage_idle"}}`
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMetricInt(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(false, false)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := `{"_value":90,"config:hecRouting":false,"config:multiMetric":false,"cpu":"cpu0","metric_name":"cpu.usage_idle","time":0}`
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMetricIntHec(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(true, false)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := `{"time":0,"event":"metric","fields":{"_value":90,"config:hecRouting":true,"config:multiMetric":false,"cpu":"cpu0","metric_name":"cpu.usage_idle"}}`
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMetricBool(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"container-name": "telegraf-test",
	}
	fields := map[string]interface{}{
		"oomkiller": bool(true),
	}
	m, err := metric.New("docker", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(false, false)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := `{"_value":1,"config:hecRouting":false,"config:multiMetric":false,"container-name":"telegraf-test","metric_name":"docker.oomkiller","time":0}`
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMetricBoolHec(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"container-name": "telegraf-test",
	}
	fields := map[string]interface{}{
		"oomkiller": bool(false),
	}
	m, err := metric.New("docker", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(true, false)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := `{"time":0,"event":"metric","fields":{"_value":0,"config:hecRouting":true,"config:multiMetric":false,"container-name":"telegraf-test","metric_name":"docker.oomkiller"}}`
	assert.Equal(t, string(expS), string(buf))
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(false, false)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := `{"_value":5,"config:hecRouting":false,"config:multiMetric":false,"cpu":"cpu0","metric_name":"cpu.usage_idle","time":0}`
	assert.Equal(t, string(expS), string(buf))
	assert.NoError(t, err)
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
	n := MustMetric(
		metric.New(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"value": 92.0,
			},
			time.Unix(0, 0),
		),
	)

	metrics := []telegraf.Metric{m, n}
	s, _ := NewSerializer(false, false)
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := `{"_value":42,"config:hecRouting":false,"config:multiMetric":false,"metric_name":"cpu.value","time":0}{"_value":92,"config:hecRouting":false,"config:multiMetric":false,"metric_name":"cpu.value","time":0}`
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMulti(t *testing.T) {
	m := MustMetric(
		metric.New(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"user":   42.0,
				"system": 8.0,
			},
			time.Unix(0, 0),
		),
	)

	metrics := []telegraf.Metric{m}
	s, _ := NewSerializer(false, true)
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := `{"config:hecRouting":false,"config:multiMetric":true,"metric_name:cpu.system":8,"metric_name:cpu.user":42,"time":0}`
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatchHec(t *testing.T) {
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
	n := MustMetric(
		metric.New(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"value": 92.0,
			},
			time.Unix(0, 0),
		),
	)

	metrics := []telegraf.Metric{m, n}
	s, _ := NewSerializer(true, false)
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := `{"time":0,"event":"metric","fields":{"_value":42,"config:hecRouting":true,"config:multiMetric":false,"metric_name":"cpu.value"}}{"time":0,"event":"metric","fields":{"_value":92,"config:hecRouting":true,"config:multiMetric":false,"metric_name":"cpu.value"}}`
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMultiHec(t *testing.T) {
	m := MustMetric(
		metric.New(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"usage":  42.0,
				"system": 8.0,
			},
			time.Unix(0, 0),
		),
	)

	metrics := []telegraf.Metric{m}
	s, _ := NewSerializer(true, true)
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := `{"time":0,"event":"metric","fields":{"config:hecRouting":true,"config:multiMetric":true,"metric_name:cpu.system":8,"metric_name:cpu.usage":42}}`
	assert.Equal(t, string(expS), string(buf))
}
