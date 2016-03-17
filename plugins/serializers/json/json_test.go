package json

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/influxdata/telegraf"
)

func TestSerializeMetricFloat(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := JsonSerializer{}
	mS, err := s.Serialize(m)
	assert.NoError(t, err)
	expS := []string{fmt.Sprintf("{\"fields\":{\"usage_idle\":91.5},\"name\":\"cpu\",\"tags\":{\"cpu\":\"cpu0\"},\"timestamp\":%d}", now.Unix())}
	assert.Equal(t, expS, mS)
}

func TestSerializeMetricInt(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := JsonSerializer{}
	mS, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []string{fmt.Sprintf("{\"fields\":{\"usage_idle\":90},\"name\":\"cpu\",\"tags\":{\"cpu\":\"cpu0\"},\"timestamp\":%d}", now.Unix())}
	assert.Equal(t, expS, mS)
}

func TestSerializeMetricString(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": "foobar",
	}
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := JsonSerializer{}
	mS, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []string{fmt.Sprintf("{\"fields\":{\"usage_idle\":\"foobar\"},\"name\":\"cpu\",\"tags\":{\"cpu\":\"cpu0\"},\"timestamp\":%d}", now.Unix())}
	assert.Equal(t, expS, mS)
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
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := JsonSerializer{}
	mS, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []string{fmt.Sprintf("{\"fields\":{\"usage_idle\":90,\"usage_total\":8559615},\"name\":\"cpu\",\"tags\":{\"cpu\":\"cpu0\"},\"timestamp\":%d}", now.Unix())}
	assert.Equal(t, expS, mS)
}
