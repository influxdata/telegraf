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
	now := time.Unix(0, 0)
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
	expS := `{"time":0,"event":"metric","fields":{"_value":91.5,"cpu":"cpu0","metric_name":"cpu.usage_idle"}}` + "\n"
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

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := `{"time":0,"event":"metric","fields":{"_value":90,"cpu":"cpu0","metric_name":"cpu.usage_idle"}}` + "\n"
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

	expS := ""
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
	s, _ := NewSerializer(0)
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := `{"time":0,"event":"metric","fields":{"_value":42,"metric_name":"cpu.value"}}` + "\n" + `{"time":0,"event":"metric","fields":{"_value":92,"metric_name":"cpu.value"}}` + "\n"
	assert.Equal(t, string(expS), string(buf))
}
