package msgpack

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func isIdentical(t *testing.T, m1 telegraf.Metric, m2 *Metric) bool {
	// Max precision in msgpack is nanoseconds
	// https://github.com/msgpack/msgpack/blob/master/spec.md
	if m1.Time().Truncate(time.Nanosecond) != m2.Time {
		t.Logf("expected: %v, actual: %v", m1.Time().Truncate(time.Nanosecond), m2.Time)
		return false
	}

	if m1.Name() != m2.Name {
		t.Logf("expected: %v, actual: %v", m1.Name(), m2.Name)
		return false
	}

	if len(m1.Tags()) != len(m2.Tags) {
		t.Logf("expected: %v, actual: %v", m1.Tags(), m2.Tags)
		return false
	}

	for k, v := range m1.Tags() {
		if m2.Tags[k] != v {
			t.Logf("expected: %v, actual: %v", m1.Tags(), m2.Tags)
			return false
		}
	}

	if len(m1.Fields()) != len(m2.Fields) {
		t.Logf("expected: %v, actual: %v", m1.Fields(), m2.Fields)
		return false
	}

	for k, v := range m1.Fields() {
		if m2.Fields[k] != v {
			t.Logf("expected: %v, actual: %v", m1.Fields(), m2.Fields)
			return false
		}
	}

	return true
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

	s := Serializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	assert.NoError(t, err)

	assert.Equal(t, len(left), 0)

	assert.Equal(t, true, isIdentical(t, m, m2))
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

	s := Serializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	assert.NoError(t, err)

	assert.Equal(t, len(left), 0)

	assert.Equal(t, true, isIdentical(t, m, m2))
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

	s := Serializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	assert.NoError(t, err)

	assert.Equal(t, len(left), 0)

	assert.Equal(t, true, isIdentical(t, m, m2))
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

	s := Serializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	assert.NoError(t, err)

	assert.Equal(t, len(left), 0)

	assert.Equal(t, true, isIdentical(t, m, m2))
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

	s := Serializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	assert.NoError(t, err)

	assert.Equal(t, len(left), 0)

	assert.Equal(t, true, isIdentical(t, m, m2))
}

func TestSerializeMultipleMetric(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu tag": "cpu0",
	}
	fields := map[string]interface{}{
		"U,age=Idle": int64(90),
	}
	m, err := metric.New("My CPU", tags, fields, now)
	assert.NoError(t, err)

	s := Serializer{}

	encoded, err := s.Serialize(m)
	assert.NoError(t, err)

	// Multiple metrics in continous bytes stream
	var buf []byte
	buf = append(buf, encoded...)
	buf = append(buf, encoded...)
	buf = append(buf, encoded...)
	buf = append(buf, encoded...)

	left := buf
	for len(left) > 0 {
		decodeM := &Metric{}
		left, err = decodeM.UnmarshalMsg(left)

		assert.NoError(t, err)
		assert.Equal(t, true, isIdentical(t, m, decodeM))
	}
}

func TestSerializeBatch(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu tag": "cpu0",
	}
	fields := map[string]interface{}{
		"U,age=Idle": int64(90),
	}
	m, err := metric.New("My CPU", tags, fields, now)
	assert.NoError(t, err)

	metrics := []telegraf.Metric{m, m, m, m}

	s := Serializer{}

	buf, err := s.SerializeBatch(metrics)
	assert.NoError(t, err)

	left := buf
	for len(left) > 0 {
		decodeM := &Metric{}
		left, err = decodeM.UnmarshalMsg(left)

		assert.NoError(t, err)
		assert.Equal(t, true, isIdentical(t, m, decodeM))
	}
}
