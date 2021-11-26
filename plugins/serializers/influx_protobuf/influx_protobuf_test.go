package influx_protobuf

import (
	"fmt"
	"testing"
	"time"

	influx "github.com/influxdata/influxdb-pb-data-protocol/golang"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
)

func isNull(mask []byte, idx int) bool {
	offset, bit := (idx / 8), (idx % 8)

	if len(mask) < offset+1 {
		return false
	}

	return (mask[offset] & (1 << bit)) != 0
}

func getValue(values *influx.Column_Values, idx int) (interface{}, error) {
	length := map[string]int{
		"i64":    len(values.I64Values),
		"u64":    len(values.U64Values),
		"f64":    len(values.F64Values),
		"string": len(values.StringValues),
		"bool":   len(values.BoolValues),
		"byte":   len(values.BytesValues),
	}

	// Check if only one value
	var valid []string
	for k, l := range length {
		if l > 0 {
			valid = append(valid, k)
		}
	}
	if len(valid) == 0 {
		return nil, fmt.Errorf("no values")
	}
	if len(valid) > 1 {
		return nil, fmt.Errorf("mixture of value types %v", valid)
	}

	switch valid[0] {
	case "i64":
		v := values.I64Values
		if len(v) < idx {
			return v[len(v)-1], nil
		}
		return v[idx], nil
	case "u64":
		v := values.U64Values
		if len(v) < idx {
			return v[len(v)-1], nil
		}
		return v[idx], nil
	case "f64":
		v := values.F64Values
		if len(v) < idx {
			return v[len(v)-1], nil
		}
		return v[idx], nil
	case "string":
		v := values.StringValues
		if len(v) < idx {
			return v[len(v)-1], nil
		}
		return v[idx], nil
	case "bool":
		v := values.BoolValues
		if len(v) < idx {
			return v[len(v)-1], nil
		}
		return v[idx], nil
	case "byte":
		v := values.BytesValues
		if len(v) < idx {
			return v[len(v)-1], nil
		}
		return v[idx], nil
	}

	return nil, fmt.Errorf("unhandled value type %q", valid[0])
}

func getRow(columns []*influx.Column, idx int) (telegraf.Metric, error) {
	var tset bool
	var t time.Time

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	for _, col := range columns {
		name := col.ColumnName
		if isNull(col.NullMask, idx) {
			// This column has no valid data in this row
			continue
		}
		value, err := getValue(col.Values, idx)
		if err != nil {
			return nil, fmt.Errorf("get value of column %q failed: %v", name, err)
		}
		switch col.SemanticType {
		case influx.Column_SEMANTIC_TYPE_IOX, influx.Column_SEMANTIC_TYPE_FIELD:
			fields[name] = value
		case influx.Column_SEMANTIC_TYPE_TAG:
			v, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("tag in column %q is not a string but %T", name, value)
			}
			tags[name] = v
		case influx.Column_SEMANTIC_TYPE_TIME:
			if tset {
				return nil, fmt.Errorf("time previously set but set again in column %q", name)
			}
			v, ok := value.(int64)
			if !ok {
				return nil, fmt.Errorf("time in column %q is not int64 but %T", name, value)
			}
			t, tset = time.Unix(0, v).UTC(), true
		default:
			return nil, fmt.Errorf("unknown column type %v", col.SemanticType)
		}
	}

	return testutil.MustMetric("", tags, fields, t), nil
}

func toMetrics(tables []*influx.TableBatch) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric

	for _, table := range tables {
		name := table.TableName
		for i := 0; i < int(table.RowCount); i++ {
			m, err := getRow(table.Columns, i)
			if err != nil {
				return nil, fmt.Errorf("getting row %d in table %q failed: %v", i, name, err)
			}
			m.SetName(name)
			metrics = append(metrics, m)
		}
	}

	return metrics, nil
}

func TestSerializeMetricSingleNonIOx(t *testing.T) {
	m := testutil.MustMetric(
		"protobuf_test",
		map[string]string{
			"foo":  "42",
			"bar":  "abc",
			"pi":   "3.14",
			"test": "true",
		},
		map[string]interface{}{
			"speed":      float64(42.23),
			"rpm":        uint64(3500),
			"gear":       int64(3),
			"status":     "ok",
			"autonomous": true,
		},
		time.Now().UTC(),
	)

	s := Serializer{DatabaseName: "wunderbar"}
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	var db influx.DatabaseBatch
	require.NoError(t, proto.Unmarshal(buf, &db))
	require.Equal(t, db.DatabaseName, "wunderbar")

	actual, err := toMetrics(db.TableBatches)
	require.NoError(t, err)
	require.Len(t, actual, 1)
	testutil.RequireMetricEqual(t, m, actual[0])
}

func TestSerializeMetricSingleBatchNonIOx(t *testing.T) {
	m := []telegraf.Metric{
		testutil.MustMetric(
			"protobuf_test",
			map[string]string{
				"foo":  "42",
				"bar":  "abc",
				"pi":   "3.14",
				"test": "true",
			},
			map[string]interface{}{
				"speed":      float64(42.23),
				"rpm":        uint64(3500),
				"gear":       int64(3),
				"status":     "ok",
				"autonomous": true,
			},
			time.Now().UTC(),
		),
	}

	s := Serializer{DatabaseName: "wunderbar"}
	buf, err := s.SerializeBatch(m)
	require.NoError(t, err)

	var db influx.DatabaseBatch
	require.NoError(t, proto.Unmarshal(buf, &db))
	require.Equal(t, db.DatabaseName, "wunderbar")

	actual, err := toMetrics(db.TableBatches)
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, m, actual, testutil.SortMetrics())
}

func TestSerializeMultipleNonIOx(t *testing.T) {
	input := []telegraf.Metric{
		testutil.TestMetric(float64(42.1), "a"),
		testutil.TestMetric(float32(42.2), "b"),
		testutil.TestMetric(uint64(3500), "c"),
		testutil.TestMetric(uint32(3500), "d"),
		testutil.TestMetric(uint16(35), "e"),
		testutil.TestMetric(uint8(35), "f"),
		testutil.TestMetric(uint(35), "g"),
		testutil.TestMetric(int64(23), "h"),
		testutil.TestMetric(int32(23), "i"),
		testutil.TestMetric(int16(23), "j"),
		testutil.TestMetric(int8(23), "k"),
		testutil.TestMetric(int(23), "l"),
		testutil.TestMetric("ok", "m"),
		testutil.TestMetric("", "n"),
		testutil.TestMetric(true, "o"),
		testutil.TestMetric(false, "p"),
	}
	expected := []telegraf.Metric{
		testutil.TestMetric(float64(42.1), "a"),
		testutil.TestMetric(float64(float32(42.2)), "b"),
		testutil.TestMetric(uint64(3500), "c"),
		testutil.TestMetric(uint64(3500), "d"),
		testutil.TestMetric(uint64(35), "e"),
		testutil.TestMetric(uint64(35), "f"),
		testutil.TestMetric(uint64(35), "g"),
		testutil.TestMetric(int64(23), "h"),
		testutil.TestMetric(int64(23), "i"),
		testutil.TestMetric(int64(23), "j"),
		testutil.TestMetric(int64(23), "k"),
		testutil.TestMetric(int64(23), "l"),
		testutil.TestMetric("ok", "m"),
		testutil.TestMetric("", "n"),
		testutil.TestMetric(true, "o"),
		testutil.TestMetric(false, "p"),
	}

	s := Serializer{DatabaseName: "wunderbar"}
	buf, err := s.SerializeBatch(input)
	require.NoError(t, err)

	var db influx.DatabaseBatch
	require.NoError(t, proto.Unmarshal(buf, &db))
	require.Equal(t, db.DatabaseName, "wunderbar")

	actual, err := toMetrics(db.TableBatches)
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
}

func TestSerializeMetricSingleIOx(t *testing.T) {
	input := testutil.MustMetric(
		"protobuf_test",
		map[string]string{
			"foo":  "42",
			"bar":  "abc",
			"pi":   "3.14",
			"test": "true",
		},
		map[string]interface{}{
			"speed":      float64(42.23),
			"rpm":        uint64(3500),
			"gear":       int64(3),
			"status":     "ok",
			"autonomous": true,
		},
		time.Now().UTC(),
	)

	expected := testutil.MustMetric(
		"protobuf_test",
		map[string]string{},
		map[string]interface{}{
			"speed":      float64(42.23),
			"rpm":        uint64(3500),
			"gear":       int64(3),
			"status":     "ok",
			"autonomous": true,
			"foo":        "42",
			"bar":        "abc",
			"pi":         "3.14",
			"test":       "true",
		},
		input.Time(),
	)

	s := Serializer{
		DatabaseName: "wunderbar",
		IsIox:        true,
	}
	buf, err := s.Serialize(input)
	require.NoError(t, err)

	var db influx.DatabaseBatch
	require.NoError(t, proto.Unmarshal(buf, &db))
	require.Equal(t, db.DatabaseName, "wunderbar")

	actual, err := toMetrics(db.TableBatches)
	require.NoError(t, err)
	require.Len(t, actual, 1)
	testutil.RequireMetricEqual(t, expected, actual[0])
}

func TestSerializeMetricSingleBatchIOx(t *testing.T) {
	input := []telegraf.Metric{
		testutil.MustMetric(
			"protobuf_test",
			map[string]string{
				"foo":  "42",
				"bar":  "abc",
				"pi":   "3.14",
				"test": "true",
			},
			map[string]interface{}{
				"speed":      float64(42.23),
				"rpm":        uint64(3500),
				"gear":       int64(3),
				"status":     "ok",
				"autonomous": true,
			},
			time.Now().UTC(),
		),
	}

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"protobuf_test",
			map[string]string{},
			map[string]interface{}{
				"speed":      float64(42.23),
				"rpm":        uint64(3500),
				"gear":       int64(3),
				"status":     "ok",
				"autonomous": true,
				"foo":        "42",
				"bar":        "abc",
				"pi":         "3.14",
				"test":       "true",
			},
			input[0].Time(),
		),
	}
	s := Serializer{
		DatabaseName: "wunderbar",
		IsIox:        true,
	}
	buf, err := s.SerializeBatch(input)
	require.NoError(t, err)

	var db influx.DatabaseBatch
	require.NoError(t, proto.Unmarshal(buf, &db))
	require.Equal(t, db.DatabaseName, "wunderbar")

	actual, err := toMetrics(db.TableBatches)
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
}

func TestSerializeMultipleIOx(t *testing.T) {
	input := []telegraf.Metric{
		testutil.TestMetric(float64(42.1), "a"),
		testutil.TestMetric(float32(42.2), "b"),
		testutil.TestMetric(uint64(3500), "c"),
		testutil.TestMetric(uint32(3500), "d"),
		testutil.TestMetric(uint16(35), "e"),
		testutil.TestMetric(uint8(35), "f"),
		testutil.TestMetric(uint(35), "g"),
		testutil.TestMetric(int64(23), "h"),
		testutil.TestMetric(int32(23), "i"),
		testutil.TestMetric(int16(23), "j"),
		testutil.TestMetric(int8(23), "k"),
		testutil.TestMetric(int(23), "l"),
		testutil.TestMetric("ok", "m"),
		testutil.TestMetric("", "n"),
		testutil.TestMetric(true, "o"),
		testutil.TestMetric(false, "p"),
	}

	expected := make([]telegraf.Metric, 0, len(input))
	for _, m := range input {
		e := m.Copy()
		for key, value := range e.Fields() {
			switch x := value.(type) {
			case int, int8, int16, int32:
				v, err := internal.ToInt64(value)
				require.NoError(t, err)
				e.AddField(key, v)
			case uint, uint8, uint16, uint32:
				v, err := internal.ToUint64(value)
				require.NoError(t, err)
				e.AddField(key, v)
			case float32:
				e.AddField(key, float64(x))
			}
		}
		for key, value := range e.Tags() {
			e.AddField(key, value)
			e.RemoveTag(key)
		}
		expected = append(expected, e)
	}

	s := Serializer{
		DatabaseName: "wunderbar",
		IsIox:        true,
	}
	buf, err := s.SerializeBatch(input)
	require.NoError(t, err)

	var db influx.DatabaseBatch
	require.NoError(t, proto.Unmarshal(buf, &db))
	require.Equal(t, db.DatabaseName, "wunderbar")

	actual, err := toMetrics(db.TableBatches)
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
}
