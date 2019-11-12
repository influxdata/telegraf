package customjson

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

func TestSerializeDefaultMetricTimestamp(t *testing.T) {
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

	inputJmespathExpression := ""
	inputTagsPrefix := "tags"

	s, _ := NewSerializer(inputJmespathExpression, inputTagsPrefix)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)
	expS := []byte(fmt.Sprintf(`{"metric_family":"cpu","metric_name":"usage_idle","metric_value":91.5,"tags_cpu":"cpu0","timestamp":1557233480819}`) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeCustomMetricTimestamp(t *testing.T) {
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

	inputJmespathExpression := "{timestamp:timestamp,event:'metric',family_name:metric_family,fields:{_value:metric_value,name:metric_name}}"
	inputTagsPrefix := "tags"

	s, _ := NewSerializer(inputJmespathExpression, inputTagsPrefix)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)
	expS := []byte(fmt.Sprintf(`{"event":"metric","family_name":"cpu","fields":{"_value":91.5,"name":"usage_idle"},"tags_cpu":"cpu0","timestamp":1557233480819}`) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeDefaultMetricInt(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	inputJmespathExpression := ""
	inputTagsPrefix := "tags"

	s, _ := NewSerializer(inputJmespathExpression, inputTagsPrefix)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"metric_family":"cpu","metric_name":"usage_idle","metric_value":90,"tags_cpu":"cpu0","timestamp":0}`) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeCustomMetricInt(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	inputJmespathExpression := "{timestamp:timestamp,event:'metric',family_name:metric_family,fields:{_value:metric_value,name:metric_name}}"
	inputTagsPrefix := "tags"

	s, _ := NewSerializer(inputJmespathExpression, inputTagsPrefix)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"event":"metric","family_name":"cpu","fields":{"_value":90,"name":"usage_idle"},"tags_cpu":"cpu0","timestamp":0}`) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeDefaultMetricBool(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"container_name": "telegraf-test",
	}
	fields := map[string]interface{}{
		"oomkiller": bool(true),
	}
	m, err := metric.New("docker", tags, fields, now)
	assert.NoError(t, err)

	inputJmespathExpression := ""
	inputTagsPrefix := "tags"

	s, _ := NewSerializer(inputJmespathExpression, inputTagsPrefix)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"metric_family":"docker","metric_name":"oomkiller","metric_value":1,"tags_container_name":"telegraf-test","timestamp":0}`) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeCustomMetricBool(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"container_name": "telegraf-test",
	}
	fields := map[string]interface{}{
		"oomkiller": bool(true),
	}
	m, err := metric.New("docker", tags, fields, now)
	assert.NoError(t, err)

	inputJmespathExpression := "{timestamp:timestamp,event:'metric',family_name:metric_family,fields:{_value:metric_value,name:metric_name}}"
	inputTagsPrefix := "tags"

	s, _ := NewSerializer(inputJmespathExpression, inputTagsPrefix)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"event":"metric","family_name":"docker","fields":{"_value":1,"name":"oomkiller"},"tags_container_name":"telegraf-test","timestamp":0}`) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeDefaultMetricString(t *testing.T) {
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

	inputJmespathExpression := ""
	inputTagsPrefix := "tags"

	s, _ := NewSerializer(inputJmespathExpression, inputTagsPrefix)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"metric_family":"cpu","metric_name":"usage_idle","metric_value":5,"tags_cpu":"cpu0","timestamp":0}`) + "\n")
	assert.Equal(t, string(expS), string(buf))
	assert.NoError(t, err)
}

func TestSerializeCustomMetricString(t *testing.T) {
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

	inputJmespathExpression := "{timestamp:timestamp,event:'metric',family_name:metric_family,fields:{_value:metric_value,name:metric_name}}"
	inputTagsPrefix := "tags"

	s, _ := NewSerializer(inputJmespathExpression, inputTagsPrefix)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"event":"metric","family_name":"cpu","fields":{"_value":5,"name":"usage_idle"},"tags_cpu":"cpu0","timestamp":0}`) + "\n")
	assert.Equal(t, string(expS), string(buf))
	assert.NoError(t, err)
}

func TestSerializeDefaultMetricBatch(t *testing.T) {
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

	inputJmespathExpression := ""
	inputTagsPrefix := "tags"

	metrics := []telegraf.Metric{m, n}
	s, _ := NewSerializer(inputJmespathExpression, inputTagsPrefix)
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"metric_family":"cpu","metric_name":"value","metric_value":42,"tags_cpu":"cpu0","timestamp":0}`) + "\n" + `{"metric_family":"cpu","metric_name":"value","metric_value":92,"timestamp":0}` + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeCustomMetricBatch(t *testing.T) {
	m := MustMetric(
		metric.New(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"usage_idle": 42.0,
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

	inputJmespathExpression := "{timestamp:timestamp,event:'metric',family_name:metric_family,fields:{_value:metric_value,name:metric_name}}"
	inputTagsPrefix := "tags"

	metrics := []telegraf.Metric{m, n}
	s, _ := NewSerializer(inputJmespathExpression, inputTagsPrefix)
	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	expS := []byte(fmt.Sprintf(`{"event":"metric","family_name":"cpu","fields":{"_value":42,"name":"usage_idle"},"tags_cpu":"cpu0","timestamp":0}`) + "\n" + `{"event":"metric","family_name":"cpu","fields":{"_value":92,"name":"value"},"timestamp":0}` + "\n")
	assert.Equal(t, string(expS), string(buf))
}
