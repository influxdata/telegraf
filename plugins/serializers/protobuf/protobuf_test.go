package protobuf

import (
	"testing"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf/metric"

	"github.com/stretchr/testify/assert"
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

	s := ProtobufSerializer{}
	buf, _ := s.Serialize(m)

	mS := &Metric{}
	err = proto.Unmarshal(buf, mS)
	assert.NoError(t, err)

	assert.Equal(t, "cpu", mS.GetName())
	assert.Equal(t, "cpu0", mS.GetTags()["cpu"])
	assert.Equal(t, float64(91.5), mS.GetFields()["usage_idle"].GetFloatValue())
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

	s := ProtobufSerializer{}
	buf, _ := s.Serialize(m)

	mS := &Metric{}
	err = proto.Unmarshal(buf, mS)
	assert.NoError(t, err)

	assert.Equal(t, "cpu", mS.GetName())
	assert.Equal(t, "cpu0", mS.GetTags()["cpu"])
	assert.Equal(t, int64(90), mS.GetFields()["usage_idle"].GetIntValue())
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

	s := ProtobufSerializer{}
	buf, err := s.Serialize(m)
	assert.NoError(t, err)

	mS := &Metric{}
	err = proto.Unmarshal(buf, mS)
	assert.NoError(t, err)

	assert.Equal(t, "cpu", mS.GetName())
	assert.Equal(t, "cpu0", mS.GetTags()["cpu"])
	assert.Equal(t, "foobar", mS.GetFields()["usage_idle"].GetStringValue())
}

func TestSerializeMultiFields(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"int_field":    int64(90),
		"float_field":  float64(8559615),
		"string_field": "string_value",
		"bool_field":   true,
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := ProtobufSerializer{}
	buf, err := s.Serialize(m)
	assert.NoError(t, err)

	mS := &Metric{}
	err = proto.Unmarshal(buf, mS)
	assert.NoError(t, err)

	assert.Equal(t, "cpu", mS.GetName())
	assert.Equal(t, "cpu0", mS.GetTags()["cpu"])

	for k, v := range mS.GetFields() {
		switch v := v.GetValue().(type) {
		case *FieldValue_IntValue:
			assert.Equal(t, k, "int_field")
			assert.Equal(t, int64(90), v.IntValue)
		case *FieldValue_FloatValue:
			assert.Equal(t, k, "float_field")
			assert.Equal(t, float64(8559615), v.FloatValue)
		case *FieldValue_StringValue:
			assert.Equal(t, k, "string_field")
			assert.Equal(t, "string_value", v.StringValue)
		case *FieldValue_BoolValue:
			assert.Equal(t, k, "bool_field")
			assert.Equal(t, true, v.BoolValue)
		default:
			t.Fatalf("Impossible data type case!")
		}
	}
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

	s := ProtobufSerializer{}
	buf, err := s.Serialize(m)
	assert.NoError(t, err)

	mS := &Metric{}
	err = proto.Unmarshal(buf, mS)
	assert.NoError(t, err)

	assert.Equal(t, "My CPU", mS.GetName())
	assert.Equal(t, "cpu0", mS.GetTags()["cpu tag"])
	assert.Equal(t, int64(90), mS.GetFields()["U,age=Idle"].GetIntValue())
}
