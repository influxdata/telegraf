package flattenjson

import (
	"fmt"
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

func TestSerializeMetricTimestamp(t *testing.T) {
	// Test sub-second time
	now := time.Unix(1557233480, 819000000)
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
	assert.NoError(t, err)
	//expS := `{"metric_family":"cpu","metric_name":"usage_idle","metric_value":91.5,"tags_cpu":"cpu0","timestamp":1557233480819}`
	expS := []byte(fmt.Sprintf(`{"metric_family":"cpu","metric_name":"usage_idle","metric_value":91.5,"tags_cpu":"cpu0","timestamp":1557233480819}`) + "\n")
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

	s, _ := NewSerializer()
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	//expS := `{"metric_family":"cpu","metric_name":"usage_idle","metric_value":90,"tags_cpu":"cpu0","timestamp":0}`
	expS := []byte(fmt.Sprintf(`{"metric_family":"cpu","metric_name":"usage_idle","metric_value":90,"tags_cpu":"cpu0","timestamp":0}`) + "\n")
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

	s, _ := NewSerializer()
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	//expS := `{"metric_family":"docker","metric_name":"oomkiller","metric_value":1,"tags_container-name":"telegraf-test","timestamp":0}`
	expS := []byte(fmt.Sprintf(`{"metric_family":"docker","metric_name":"oomkiller","metric_value":1,"tags_container-name":"telegraf-test","timestamp":0}`) + "\n")
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

	s, _ := NewSerializer()
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	//expS := `{"metric_family":"cpu","metric_name":"usage_idle","metric_value":5,"tags_cpu":"cpu0","timestamp":0}`
	expS := []byte(fmt.Sprintf(`{"metric_family":"cpu","metric_name":"usage_idle","metric_value":5,"tags_cpu":"cpu0","timestamp":0}`) + "\n")
	assert.Equal(t, string(expS), string(buf))
	assert.NoError(t, err)
}

func TestSerializeBatch(t *testing.T) {
	m := MustMetric(
		metric.New(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
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
	s, _ := NewSerializer()
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	//expS := `{"metric_family":"cpu","metric_name":"value","metric_value":42,"tags_cpu":"cpu0","timestamp":0}` + `{"metric_family":"cpu","metric_name":"value","metric_value":92,"timestamp":0}`
	expS := []byte(fmt.Sprintf(`{"metric_family":"cpu","metric_name":"value","metric_value":42,"tags_cpu":"cpu0","timestamp":0}`) + "\n" + `{"metric_family":"cpu","metric_name":"value","metric_value":92,"timestamp":0}` + "\n")
	assert.Equal(t, string(expS), string(buf))
}
