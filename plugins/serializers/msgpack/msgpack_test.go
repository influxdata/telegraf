package msgpack

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func isIdentical(m1 telegraf.Metric, m2 *Metric) bool {
	if m1.Time() != m2.Time {
		return false
	}

	if m1.Name() != m2.Name {
		return false
	}

	if len(m1.Tags()) != len(m2.Tags) {
		return false
	}

	for k, v := range m1.Tags() {
		if m2.Tags[k] != v {
			return false
		}
	}

	if len(m1.Fields()) != len(m2.Fields) {
		return false
	}

	for k, v := range m1.Fields() {
		if m2.Fields[k] != v {
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

	s := MsgpackSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	assert.NoError(t, err)

	assert.Equal(t, len(left), 0)

	assert.Equal(t, true, isIdentical(m, m2))
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

	s := MsgpackSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	assert.NoError(t, err)

	assert.Equal(t, len(left), 0)

	assert.Equal(t, true, isIdentical(m, m2))
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

	s := MsgpackSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	assert.NoError(t, err)

	assert.Equal(t, len(left), 0)

	assert.Equal(t, true, isIdentical(m, m2))
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

	s := MsgpackSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	assert.NoError(t, err)

	assert.Equal(t, len(left), 0)

	assert.Equal(t, true, isIdentical(m, m2))
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

	s := MsgpackSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	assert.NoError(t, err)

	assert.Equal(t, len(left), 0)

	assert.Equal(t, true, isIdentical(m, m2))
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

	s := MsgpackSerializer{}

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
		assert.Equal(t, true, isIdentical(m, decodeM))
	}
}
