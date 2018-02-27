package json

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := JsonSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)
	expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":91.5},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
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

	s := JsonSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":90},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
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

	s := JsonSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":"foobar"},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
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

	s := JsonSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":{"usage_idle":90,"usage_total":8559615},"name":"cpu","tags":{"cpu":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
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

	s := JsonSerializer{}
	buf, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"fields":{"U,age=Idle":90},"name":"My CPU","tags":{"cpu tag":"cpu0"},"timestamp":%d}`, now.Unix()) + "\n")
	assert.Equal(t, string(expS), string(buf))
}
