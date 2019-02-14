package prometheus

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
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

	s, _ := NewSerializer()
	var buf []byte
	buf, err = s.Serialize(m)
	log.Printf("The result is %s", buf)
	assert.NoError(t, err)
	expS := []byte(fmt.Sprintf("# HELP cpu_usage_idle Telegraf collected metric\n# TYPE cpu_usage_idle untyped\ncpu_usage_idle{cpu=\"cpu0\"} 91.5\n"))
	assert.Equal(t, string(expS), string(buf))
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

	s, _ := NewSerializer()
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf("# HELP cpu_usage_idle Telegraf collected metric\n# TYPE cpu_usage_idle untyped\ncpu_usage_idle{cpu=\"cpu0\"} 90\n"))
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

	s, _ := NewSerializer()
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte("")
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

	s, _ := NewSerializer()
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf("# HELP cpu_usage_idle Telegraf collected metric\n# TYPE cpu_usage_idle untyped\ncpu_usage_idle{cpu=\"cpu0\"} 90\n# HELP cpu_usage_total Telegraf collected metric\n# TYPE cpu_usage_total untyped\ncpu_usage_total{cpu=\"cpu0\"} 8.559615e+06\n"))
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeWithSpaces(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu 0": "cpu 0",
	}
	fields := map[string]interface{}{
		"usage_idle 1": float64(91.5),
	}
	m, err := metric.New("cpu metric", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer()
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)
	expS := []byte(fmt.Sprintf("# HELP cpu_metric_usage_idle_1 Telegraf collected metric\n# TYPE cpu_metric_usage_idle_1 untyped\ncpu_metric_usage_idle_1{cpu_0=\"cpu 0\"} 91.5\n"))
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeBatch(t *testing.T) {
	m := MustMetric(
		metric.New(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(0, 0),
		),
	)

	metrics := []telegraf.Metric{m, m}
	s, _ := NewSerializer()
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)
	expS := []byte("# HELP cpu Telegraf collected metric\n# TYPE cpu untyped\ncpu 42\n# HELP cpu Telegraf collected metric\n# TYPE cpu untyped\ncpu 42\n")
	assert.Equal(t, string(expS), string(buf))
}
