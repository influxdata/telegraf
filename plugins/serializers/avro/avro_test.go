package avro

import (
	"testing"
	"time"

	"github.com/linkedin/goavro/v2"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

func TestSerialize(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		config   Serializer
		metric   telegraf.Metric
		expected map[string]interface{}
	}{
		{
			name: "fields and tags",
			schema: `{
				"type": "record",
				"name": "cpu",
				"fields": [
					{"name": "host", "type": "string"},
					{"name": "value", "type": "double"},
					{"name": "count", "type": "long"}
				]
			}`,
			metric: metric.New(
				"cpu",
				map[string]string{"host": "server01"},
				map[string]interface{}{"value": 42.5, "count": int64(3)},
				time.Unix(1600000000, 0),
			),
			expected: map[string]interface{}{
				"host":  "server01",
				"value": 42.5,
				"count": int64(3),
			},
		},
		{
			name: "numeric coercion int field into double",
			schema: `{
				"type": "record",
				"name": "cpu",
				"fields": [{"name": "value", "type": "double"}]
			}`,
			metric: metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{"value": int64(7)},
				time.Unix(0, 0),
			),
			expected: map[string]interface{}{"value": float64(7)},
		},
		{
			name: "nullable union with a value",
			schema: `{
				"type": "record",
				"name": "cpu",
				"fields": [{"name": "value", "type": ["null", "long"]}]
			}`,
			metric: metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{"value": int64(5)},
				time.Unix(0, 0),
			),
			expected: map[string]interface{}{
				"value": map[string]interface{}{"long": int64(5)},
			},
		},
		{
			name: "measurement and timestamp fields",
			schema: `{
				"type": "record",
				"name": "cpu",
				"fields": [
					{"name": "measurement", "type": "string"},
					{"name": "time", "type": "long"},
					{"name": "value", "type": "double"}
				]
			}`,
			config: Serializer{
				MeasurementField: "measurement",
				Timestamp:        "time",
				TimestampFormat:  "unix_ms",
			},
			metric: metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{"value": 1.0},
				time.Unix(1600000000, 0),
			),
			expected: map[string]interface{}{
				"measurement": "cpu",
				"time":        int64(1600000000000),
				"value":       1.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.config
			s.Schema = tt.schema
			s.Log = testutil.Logger{}
			require.NoError(t, s.Init())

			buf, err := s.Serialize(tt.metric)
			require.NoError(t, err)

			codec, err := goavro.NewCodec(tt.schema)
			require.NoError(t, err)
			native, remaining, err := codec.NativeFromBinary(buf)
			require.NoError(t, err)
			require.Empty(t, remaining)
			require.Equal(t, tt.expected, native)
		})
	}
}

func TestSerializeJSONFormat(t *testing.T) {
	schema := `{
		"type": "record",
		"name": "cpu",
		"fields": [{"name": "value", "type": "long"}]
	}`
	s := Serializer{Schema: schema, Format: "json"}
	require.NoError(t, s.Init())

	buf, err := s.Serialize(metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{"value": int64(9)},
		time.Unix(0, 0),
	))
	require.NoError(t, err)
	require.JSONEq(t, `{"value": 9}`, string(buf))
}

func TestSerializeBatch(t *testing.T) {
	schema := `{
		"type": "record",
		"name": "cpu",
		"fields": [{"name": "value", "type": "long"}]
	}`
	s := Serializer{Schema: schema}
	require.NoError(t, s.Init())

	metrics := []telegraf.Metric{
		metric.New("cpu", map[string]string{}, map[string]interface{}{"value": int64(1)}, time.Unix(0, 0)),
		metric.New("cpu", map[string]string{}, map[string]interface{}{"value": int64(2)}, time.Unix(0, 0)),
	}

	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)

	codec, err := goavro.NewCodec(schema)
	require.NoError(t, err)

	var got []int64
	remaining := buf
	for len(remaining) > 0 {
		var native interface{}
		native, remaining, err = codec.NativeFromBinary(remaining)
		require.NoError(t, err)
		record := native.(map[string]interface{})
		got = append(got, record["value"].(int64))
	}
	require.Equal(t, []int64{1, 2}, got)
}

func TestInitErrors(t *testing.T) {
	tests := []struct {
		name string
		s    Serializer
	}{
		{name: "missing schema", s: Serializer{}},
		{name: "bad schema", s: Serializer{Schema: "{not valid avro}"}},
		{name: "bad format", s: Serializer{Schema: `{"type":"record","name":"x","fields":[]}`, Format: "yaml"}},
		{name: "bad timestamp format", s: Serializer{Schema: `{"type":"record","name":"x","fields":[]}`, TimestampFormat: "rfc3339"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.s.Init())
		})
	}
}

func TestRegistered(t *testing.T) {
	creator, ok := serializers.Serializers["avro"]
	require.True(t, ok)
	require.NotNil(t, creator())
}
