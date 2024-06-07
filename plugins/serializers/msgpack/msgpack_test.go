package msgpack

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

func toTelegrafMetric(m Metric) telegraf.Metric {
	tm := metric.New(m.Name, m.Tags, m.Fields, m.Time.time)
	return tm
}

func TestSerializeMetricInt(t *testing.T) {
	m := testutil.TestMetric(int64(90))

	s := Serializer{}
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	require.NoError(t, err)

	require.Empty(t, left)

	testutil.RequireMetricEqual(t, m, toTelegrafMetric(*m2))
}

func TestSerializeMetricString(t *testing.T) {
	m := testutil.TestMetric("foobar")

	s := Serializer{}
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	require.NoError(t, err)

	require.Empty(t, left)

	testutil.RequireMetricEqual(t, m, toTelegrafMetric(*m2))
}

func TestSerializeMultiFields(t *testing.T) {
	m := testutil.TestMetric(90)
	m.AddField("value2", 8559615)

	s := Serializer{}
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	require.NoError(t, err)

	require.Empty(t, left)

	testutil.RequireMetricEqual(t, m, toTelegrafMetric(*m2))
}

func TestSerializeMetricWithEscapes(t *testing.T) {
	m := testutil.TestMetric(90)
	m.AddField("U,age=Idle", int64(90))
	m.AddTag("cpu tag", "cpu0")

	s := Serializer{}
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	m2 := &Metric{}
	left, err := m2.UnmarshalMsg(buf)
	require.NoError(t, err)

	require.Empty(t, left)

	testutil.RequireMetricEqual(t, m, toTelegrafMetric(*m2))
}

func TestSerializeMultipleMetric(t *testing.T) {
	m := testutil.TestMetric(90)

	s := Serializer{}

	encoded, err := s.Serialize(m)
	require.NoError(t, err)

	// Multiple metrics in continuous bytes stream
	var buf []byte
	buf = append(buf, encoded...)
	buf = append(buf, encoded...)
	buf = append(buf, encoded...)
	buf = append(buf, encoded...)

	left := buf
	for len(left) > 0 {
		decodeM := &Metric{}
		left, err = decodeM.UnmarshalMsg(left)

		require.NoError(t, err)
		testutil.RequireMetricEqual(t, m, toTelegrafMetric(*decodeM))
	}
}

func TestSerializeBatch(t *testing.T) {
	m := testutil.TestMetric(90)

	metrics := []telegraf.Metric{m, m, m, m}

	s := Serializer{}

	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)

	left := buf
	for len(left) > 0 {
		decodeM := &Metric{}
		left, err = decodeM.UnmarshalMsg(left)

		require.NoError(t, err)
		testutil.RequireMetricEqual(t, m, toTelegrafMetric(*decodeM))
	}
}

func BenchmarkSerialize(b *testing.B) {
	s := &Serializer{}
	metrics := serializers.BenchmarkMetrics(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.Serialize(metrics[i%len(metrics)])
		require.NoError(b, err)
	}
}

func BenchmarkSerializeBatch(b *testing.B) {
	s := &Serializer{}
	m := serializers.BenchmarkMetrics(b)
	metrics := m[:]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.SerializeBatch(metrics)
		require.NoError(b, err)
	}
}
