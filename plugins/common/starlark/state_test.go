package starlark

import (
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	starlark_time "go.starlark.net/lib/time"
	"go.starlark.net/starlark"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestMarshalInvalidState(t *testing.T) {
	_, err := marshal(nil)
	require.ErrorContains(t, err, "invalid state")
}

func TestUnmarshalInvalidState(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		expected string
	}{
		{
			name:     "empty state",
			state:    "",
			expected: `invalid state ""`,
		},
		{
			name:     "no JSON",
			state:    "not json",
			expected: "unmarshalling state failed",
		},
		{
			name:     "no dictionary",
			state:    `["L","i:1"]`,
			expected: "unexpected state type *starlark.List",
		},
		{
			name:     "scalar state",
			state:    `"i:1"`,
			expected: "unexpected state type starlark.Int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := unmarshal(tt.state)
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestSerializationFormat(t *testing.T) {
	timestamp := time.Date(2026, time.October, 7, 12, 34, 56, 789, time.UTC)

	tests := []struct {
		name     string
		state    *starlark.Dict
		expected string
	}{
		{
			name:     "none",
			state:    newState(t, starlark.String("v"), starlark.None),
			expected: `["D","s:v","n:"]`,
		},
		{
			name:     "boolean",
			state:    newState(t, starlark.String("v"), starlark.Bool(true)),
			expected: `["D","s:v","b:true"]`,
		},
		{
			name:     "integer",
			state:    newState(t, starlark.String("v"), starlark.MakeInt(-42)),
			expected: `["D","s:v","i:-42"]`,
		},
		{
			name:     "unsigned integer",
			state:    newState(t, starlark.String("v"), starlark.MakeUint64(math.MaxUint64)),
			expected: `["D","s:v","i:18446744073709551615"]`,
		},
		{
			name:     "float",
			state:    newState(t, starlark.String("v"), starlark.Float(0.5)),
			expected: `["D","s:v","f:0.5"]`,
		},
		{
			name:     "float without fraction",
			state:    newState(t, starlark.String("v"), starlark.Float(42)),
			expected: `["D","s:v","f:42"]`,
		},
		{
			name:     "infinity",
			state:    newState(t, starlark.String("v"), starlark.Float(math.Inf(-1))),
			expected: `["D","s:v","f:-Inf"]`,
		},
		{
			name:     "not a number",
			state:    newState(t, starlark.String("v"), starlark.Float(math.NaN())),
			expected: `["D","s:v","f:NaN"]`,
		},
		{
			name:     "string",
			state:    newState(t, starlark.String("v"), starlark.String("foo:bar")),
			expected: `["D","s:v","s:foo:bar"]`,
		},
		{
			name:     "empty string",
			state:    newState(t, starlark.String("v"), starlark.String("")),
			expected: `["D","s:v","s:"]`,
		},
		{
			name:     "non-UTF8 string",
			state:    newState(t, starlark.String("v"), starlark.String("\xff\xfe")),
			expected: `["D","s:v","r://4="]`,
		},
		{
			name:     "bytes",
			state:    newState(t, starlark.String("v"), starlark.Bytes([]byte{0x00, 0xff})),
			expected: `["D","s:v","y:AP8="]`,
		},
		{
			name:     "time",
			state:    newState(t, starlark.String("v"), starlark_time.Time(timestamp)),
			expected: `["D","s:v","t:2026-10-07T12:34:56.000000789Z"]`,
		},
		{
			name:     "duration",
			state:    newState(t, starlark.String("v"), starlark_time.Duration(42*time.Second)),
			expected: `["D","s:v","d:42000000000"]`,
		},
		{
			name:     "empty list",
			state:    newState(t, starlark.String("v"), starlark.NewList(nil)),
			expected: `["D","s:v",["L"]]`,
		},
		{
			name:     "empty tuple",
			state:    newState(t, starlark.String("v"), starlark.Tuple{}),
			expected: `["D","s:v",["T"]]`,
		},
		{
			name:     "empty set",
			state:    newState(t, starlark.String("v"), starlark.NewSet(0)),
			expected: `["D","s:v",["S"]]`,
		},
		{
			name:     "empty dictionary",
			state:    newState(t, starlark.String("v"), starlark.NewDict(0)),
			expected: `["D","s:v",["D"]]`,
		},
		{
			name:     "list",
			state:    newState(t, starlark.String("v"), starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.String("a")})),
			expected: `["D","s:v",["L","i:1","s:a"]]`,
		},
		{
			name:     "tuple",
			state:    newState(t, starlark.String("v"), starlark.Tuple{starlark.MakeInt(1), starlark.None}),
			expected: `["D","s:v",["T","i:1","n:"]]`,
		},
		{
			name:     "set",
			state:    newState(t, starlark.String("v"), newSet(t, starlark.String("a"), starlark.MakeInt(3))),
			expected: `["D","s:v",["S","s:a","i:3"]]`,
		},
		{
			name:     "nested dictionary",
			state:    newState(t, starlark.String("v"), newState(t, starlark.String("k"), starlark.MakeInt(1))),
			expected: `["D","s:v",["D","s:k","i:1"]]`,
		},
		{
			name:     "non-string key",
			state:    newState(t, starlark.Tuple{starlark.MakeInt(1), starlark.None}, starlark.String("x")),
			expected: `["D",["T","i:1","n:"],"s:x"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := marshal(tt.state)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)

			// The serialization must be reproducible after restoring the state
			restored, err := unmarshal(actual)
			require.NoError(t, err)
			again, err := marshal(restored)
			require.NoError(t, err)
			require.Equal(t, tt.expected, again)
		})
	}
}

func TestRoundtripTypes(t *testing.T) {
	bignum, ok := new(big.Int).SetString("123456789012345678901234567890", 10)
	require.True(t, ok)

	nested := newState(t,
		starlark.Tuple{starlark.MakeInt(1), starlark.String("x")}, starlark.Float(0.5),
		starlark.None, starlark.NewList([]starlark.Value{starlark.Float(1), starlark.MakeInt(1)}),
	)

	tests := []struct {
		name  string
		value starlark.Value
	}{
		{"none", starlark.None},
		{"boolean", starlark.Bool(false)},
		{"integer", starlark.MakeInt64(math.MinInt64)},
		{"unsigned integer", starlark.MakeUint64(math.MaxUint64)},
		{"big integer", starlark.MakeBigInt(bignum)},
		{"float", starlark.Float(3.14159265358979)},
		{"infinity", starlark.Float(math.Inf(1))},
		{"string", starlark.String(`a:b"c\d`)},
		{"non-UTF8 string", starlark.String("\xff\xfe")},
		{"bytes", starlark.Bytes([]byte{0x00, 0xff, 0xfe, 0x42})},
		{"time", starlark_time.Time(time.Date(2026, time.October, 7, 12, 0, 0, 1, time.UTC))},
		{"duration", starlark_time.Duration(23 * time.Minute)},
		{"list", starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.Float(1), starlark.None})},
		{"tuple", starlark.Tuple{starlark.Bool(true), starlark.String("x")}},
		{"set", newSet(t, starlark.String("a"), starlark.MakeInt(3), starlark.None)},
		{"dictionary", nested},
		{"nested containers", starlark.NewList([]starlark.Value{starlark.Tuple{newSet(t, starlark.MakeInt(1)), nested}})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newState(t, starlark.String("v"), tt.value)

			serialized, err := marshal(state)
			require.NoError(t, err)

			restored, err := unmarshal(serialized)
			require.NoError(t, err)

			actual, found, err := restored.Get(starlark.String("v"))
			require.NoError(t, err)
			require.True(t, found)
			require.Equal(t, tt.value.Type(), actual.Type())
			require.Equal(t, tt.value.String(), actual.String())
		})
	}
}

func TestRoundtripFloatSpecialValues(t *testing.T) {
	// NaN is not equal to itself, so check the restored value explicitly
	state := newState(t, starlark.String("v"), starlark.Float(math.NaN()))

	serialized, err := marshal(state)
	require.NoError(t, err)

	restored, err := unmarshal(serialized)
	require.NoError(t, err)

	value, found, err := restored.Get(starlark.String("v"))
	require.NoError(t, err)
	require.True(t, found)
	require.IsType(t, starlark.Float(0), value)
	require.True(t, math.IsNaN(float64(value.(starlark.Float))))
}

func TestRoundtripIntegerAndFloatNotMixedUp(t *testing.T) {
	state := newState(t,
		starlark.String("int"), starlark.MakeInt64(1<<62),
		starlark.String("float"), starlark.Float(1<<62),
	)

	serialized, err := marshal(state)
	require.NoError(t, err)

	restored, err := unmarshal(serialized)
	require.NoError(t, err)

	i, _, err := restored.Get(starlark.String("int"))
	require.NoError(t, err)
	require.IsType(t, starlark.Int{}, i)
	n, ok := i.(starlark.Int).Int64()
	require.True(t, ok)
	require.Equal(t, int64(1<<62), n)

	f, _, err := restored.Get(starlark.String("float"))
	require.NoError(t, err)
	require.IsType(t, starlark.Float(0), f)
}

func TestRoundtripNonUTF8Strings(t *testing.T) {
	invalid := "\xff\xfe"

	state := newState(t,
		starlark.String("scalar"), starlark.String(invalid),
		starlark.String("nested"), starlark.NewList([]starlark.Value{starlark.String(invalid)}),
		starlark.String(invalid), starlark.String("key"),
	)

	serialized, err := marshal(state)
	require.NoError(t, err)

	restored, err := unmarshal(serialized)
	require.NoError(t, err)
	require.Equal(t, state.Len(), restored.Len())

	scalar, _, err := restored.Get(starlark.String("scalar"))
	require.NoError(t, err)
	require.Equal(t, invalid, string(scalar.(starlark.String)))

	nested, _, err := restored.Get(starlark.String("nested"))
	require.NoError(t, err)
	element := nested.(*starlark.List).Index(0)
	require.Equal(t, invalid, string(element.(starlark.String)))

	value, found, err := restored.Get(starlark.String(invalid))
	require.NoError(t, err)
	require.True(t, found, "key with invalid UTF-8 not restored")
	require.Equal(t, starlark.String("key"), value)
}

func TestRoundtripDictionaryInsertOrder(t *testing.T) {
	keys := []string{"zebra", "apple", "mango", "banana"}

	state := starlark.NewDict(len(keys))
	for i, k := range keys {
		require.NoError(t, state.SetKey(starlark.String(k), starlark.MakeInt(i)))
	}

	serialized, err := marshal(state)
	require.NoError(t, err)

	restored, err := unmarshal(serialized)
	require.NoError(t, err)
	require.Equal(t, state.Keys(), restored.Keys())
}

func TestRoundtripContainerTypes(t *testing.T) {
	elements := []starlark.Value{starlark.String("a"), starlark.MakeInt(3)}

	state := newState(t,
		starlark.String("list"), starlark.NewList(elements),
		starlark.String("tuple"), starlark.Tuple(elements),
		starlark.String("set"), newSet(t, elements...),
		starlark.String("set in list"), starlark.NewList([]starlark.Value{newSet(t, elements...)}),
		starlark.String("set in tuple"), starlark.Tuple{newSet(t, elements...)},
	)

	serialized, err := marshal(state)
	require.NoError(t, err)

	restored, err := unmarshal(serialized)
	require.NoError(t, err)

	for _, item := range state.Items() {
		key := item[0]
		actual, found, err := restored.Get(key)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, item[1].Type(), actual.Type(), "type of %s changed", key)

		equal, err := starlark.Equal(item[1], actual)
		require.NoError(t, err)
		require.True(t, equal, "%s: %s != %s", key, item[1], actual)
	}

	// The restored set must still behave like a set
	set, _, err := restored.Get(starlark.String("set"))
	require.NoError(t, err)
	require.NoError(t, set.(*starlark.Set).Insert(starlark.String("a")))
	require.Equal(t, 2, set.(*starlark.Set).Len())
}

func TestRoundtripMetric(t *testing.T) {
	timestamp := time.Date(3000, time.January, 1, 0, 0, 0, 789, time.UTC)
	expected := metric.New(
		"cpu",
		map[string]string{"host": "localhost", "cpu": "cpu0"},
		map[string]any{
			"idle":   float64(99.5),
			"count":  int64(-7),
			"big":    uint64(math.MaxUint64),
			"active": true,
			"label":  "some text",
		},
		timestamp,
		telegraf.Counter,
	)

	m := &Metric{}
	m.Wrap(expected.Copy())

	state := newState(t, starlark.String("last"), m)
	serialized, err := marshal(state)
	require.NoError(t, err)

	restored, err := unmarshal(serialized)
	require.NoError(t, err)

	value, found, err := restored.Get(starlark.String("last"))
	require.NoError(t, err)
	require.True(t, found)
	require.IsType(t, &Metric{}, value)

	actual := value.(*Metric).Unwrap()
	testutil.RequireMetricEqual(t, expected, actual)
	require.Equal(t, telegraf.Counter, actual.Type())
	require.True(t, timestamp.Equal(actual.Time()))
	require.Equal(t, expected.FieldList()[0].Key, actual.FieldList()[0].Key, "field order changed")
}

func TestRoundtripMetricDropsTracking(t *testing.T) {
	var delivered int
	tracked, _ := metric.WithTracking(
		metric.New("cpu", nil, map[string]any{"value": 42}, time.Unix(0, 0)),
		func(telegraf.DeliveryInfo) { delivered++ },
	)

	m := &Metric{}
	m.Wrap(tracked)
	require.NotZero(t, m.ID, "expected a tracking metric")

	serialized, err := marshal(newState(t, starlark.String("last"), m))
	require.NoError(t, err)

	restored, err := unmarshal(serialized)
	require.NoError(t, err)

	value, _, err := restored.Get(starlark.String("last"))
	require.NoError(t, err)
	require.NotImplements(t, (*telegraf.TrackingMetric)(nil), value.(*Metric).Unwrap())
	require.Zero(t, value.(*Metric).ID)
	require.Zero(t, delivered)
}

func TestMaxNestingDepthWithinLimit(t *testing.T) {
	state := newState(t, starlark.String("v"), nest(starlark.MakeInt(1), maxNestingDepth-1))

	serialized, err := marshal(state)
	require.NoError(t, err)

	restored, err := unmarshal(serialized)
	require.NoError(t, err)
	again, err := marshal(restored)
	require.NoError(t, err)
	require.Equal(t, serialized, again)
}

func TestMaxNestingDepthExceeded(t *testing.T) {
	state := newState(t, starlark.String("v"), nest(starlark.MakeInt(1), maxNestingDepth+1))

	_, err := marshal(state)
	require.ErrorContains(t, err, "exceeding maximum nesting depth")
}

func TestMaxNestingDepthExceededSelfReference(t *testing.T) {
	list := starlark.NewList(nil)
	require.NoError(t, list.Append(list))

	_, err := marshal(newState(t, starlark.String("v"), list))
	require.ErrorContains(t, err, "exceeding maximum nesting depth")
}

func TestMaxNestingDepthExceededUnmarshal(t *testing.T) {
	serialized := `"i:1"`
	for range maxNestingDepth + 1 {
		serialized = `["L",` + serialized + `]`
	}

	_, err := unmarshal(`["D","s:v",` + serialized + `]`)
	require.ErrorContains(t, err, "exceeding maximum nesting depth")
}

func TestUnsupportedTypes(t *testing.T) {
	m := &Metric{}
	m.Wrap(metric.New("cpu", map[string]string{"host": "h"}, map[string]any{"value": 42}, time.Unix(0, 0)))

	tests := []struct {
		name     string
		value    starlark.Value
		expected string
	}{
		{
			name:     "builtin function",
			value:    starlark.NewBuiltin("f", nil),
			expected: `unsupported starlark type *starlark.Builtin`,
		},
		{
			name:     "metric fields",
			value:    m.Fields(),
			expected: `unsupported starlark type starlark.FieldDict`,
		},
		{
			name:     "metric tags",
			value:    m.Tags(),
			expected: `unsupported starlark type starlark.TagDict`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := marshal(newState(t, starlark.String("v"), tt.value))
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestEncodingErrorPath(t *testing.T) {
	broken := starlark.NewBuiltin("f", nil)

	tests := []struct {
		name     string
		state    *starlark.Dict
		expected string
	}{
		{
			name:     "value of the state",
			state:    newState(t, starlark.String("broken"), broken),
			expected: `in "state.val:broken": unsupported starlark type *starlark.Builtin`,
		},
		{
			name: "element of a list",
			state: newState(t, starlark.String("items"),
				starlark.NewList([]starlark.Value{starlark.MakeInt(1), broken}),
			),
			expected: `in "state.val:items.1": unsupported starlark type *starlark.Builtin`,
		},
		{
			name: "nested containers",
			state: newState(t, starlark.String("nested"),
				newState(t, starlark.String("counters"),
					starlark.NewList([]starlark.Value{starlark.Tuple{starlark.String("a"), broken}}),
				),
			),
			expected: `in "state.val:nested.val:counters.0.1": unsupported starlark type *starlark.Builtin`,
		},
		{
			name:     "element of a tuple",
			state:    newState(t, starlark.String("pair"), starlark.Tuple{starlark.String("a"), broken}),
			expected: `in "state.val:pair.1": unsupported starlark type *starlark.Builtin`,
		},
		{
			name:     "element of a set",
			state:    newState(t, starlark.String("seen"), newSet(t, starlark.Tuple{broken})),
			expected: `in "state.val:seen.0.0": unsupported starlark type *starlark.Builtin`,
		},
		{
			name:     "key of a dictionary",
			state:    newState(t, starlark.String("lookup"), newState(t, starlark.Tuple{broken}, starlark.MakeInt(1))),
			expected: `in "state.val:lookup.key:(<built-in function f>,).0": unsupported starlark type *starlark.Builtin`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := marshal(tt.state)
			require.ErrorContains(t, err, tt.expected)
			require.ErrorContains(t, err, `in "state`, "error does not contain a path")
		})
	}
}

func TestDecodingErrors(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		expected string
	}{
		{
			name:     "null value",
			state:    `["D","s:v",null]`,
			expected: `unsupported serialized type <nil>`,
		},
		{
			name:     "number value",
			state:    `["D","s:v",42]`,
			expected: `unsupported serialized type float64`,
		},
		{
			name:     "object value",
			state:    `["D","s:v",{"a":"b"}]`,
			expected: `unsupported serialized type map[string]interface {}`,
		},
		{
			name:     "missing type prefix",
			state:    `["D","s:v","nocolon"]`,
			expected: `malformed serialized value "nocolon"`,
		},
		{
			name:     "unknown type prefix",
			state:    `["D","s:v","x:1"]`,
			expected: `unsupported type prefix "x"`,
		},
		{
			name:     "container without prefix",
			state:    `["D","s:v",[]]`,
			expected: `container without type prefix`,
		},
		{
			name:     "invalid container prefix",
			state:    `["D","s:v",[42]]`,
			expected: `invalid type float64 for container prefix`,
		},
		{
			name:     "unknown container prefix",
			state:    `["D","s:v",["Z","i:1"]]`,
			expected: `unsupported container prefix "Z"`,
		},
		{
			name:     "incomplete key-value pair",
			state:    `["D","s:v",["D","s:k"]]`,
			expected: `dictionary contains an incomplete key-value pair`,
		},
		{
			name:     "value for none",
			state:    `["D","s:v","n:1"]`,
			expected: `unexpected value "1" for none type`,
		},
		{
			name:     "invalid boolean",
			state:    `["D","s:v","b:maybe"]`,
			expected: `invalid boolean "maybe"`,
		},
		{
			name:     "invalid integer",
			state:    `["D","s:v","i:one"]`,
			expected: `invalid integer "one"`,
		},
		{
			name:     "invalid float",
			state:    `["D","s:v","f:half"]`,
			expected: `invalid float "half"`,
		},
		{
			name:     "invalid raw string",
			state:    `["D","s:v","r:not base64"]`,
			expected: `invalid raw string "not base64"`,
		},
		{
			name:     "invalid bytes",
			state:    `["D","s:v","y:not base64"]`,
			expected: `invalid bytes "not base64"`,
		},
		{
			name:     "invalid time",
			state:    `["D","s:v","t:yesterday"]`,
			expected: `invalid time "yesterday"`,
		},
		{
			name:     "time out of range",
			state:    `["D","s:v","t:12026-01-01T00:00:00Z"]`,
			expected: `invalid time "12026-01-01T00:00:00Z"`,
		},
		{
			name:     "invalid duration",
			state:    `["D","s:v","d:forever"]`,
			expected: `invalid duration "forever"`,
		},
		{
			name:     "invalid metric encoding",
			state:    `["D","s:v","M:not base64"]`,
			expected: `unmarshalling metric "not base64" failed`,
		},
		{
			name:     "invalid metric",
			state:    `["D","s:v","M:AAAA"]`,
			expected: `decoding metric "AAAA" failed`,
		},
		{
			name:     "unhashable set element",
			state:    `["D","s:v",["S",["L","i:1"]]]`,
			expected: `inserting set element failed: unhashable type: list`,
		},
		{
			name:     "unhashable dictionary key",
			state:    `["D",["L","i:1"],"i:1"]`,
			expected: `setting dictionary key failed: unhashable type: list`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := unmarshal(tt.state)
			require.ErrorContains(t, err, tt.expected)
			require.ErrorContains(t, err, `in "state`, "error does not contain a path")
		})
	}
}

func TestDecodingErrorPath(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		expected string
	}{
		{
			name:     "value of the state",
			state:    `["D","s:v","x:1"]`,
			expected: `in "state.val:v(1)": unsupported type prefix "x"`,
		},
		{
			name:     "element of a list",
			state:    `["D","s:v",["L","i:1","x:1"]]`,
			expected: `in "state.val:v(1).1": unsupported type prefix "x"`,
		},
		{
			name:     "element of a tuple",
			state:    `["D","s:v",["T","i:1","x:1"]]`,
			expected: `in "state.val:v(1).1": unsupported type prefix "x"`,
		},
		{
			name:     "element of a set",
			state:    `["D","s:v",["S","x:1"]]`,
			expected: `in "state.val:v(1).0": unsupported type prefix "x"`,
		},
		{
			name:     "key of a dictionary",
			state:    `["D","s:v",["D","x:1","i:1"]]`,
			expected: `in "state.val:v(1).key:0": unsupported type prefix "x"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := unmarshal(tt.state)
			require.EqualError(t, err, tt.expected)
		})
	}
}

func TestRoundtripState(t *testing.T) {
	// A state as a script might build it over time
	m := &Metric{}
	m.Wrap(metric.New("cpu", map[string]string{"host": "h"}, map[string]any{"value": 42}, time.Unix(1760000000, 0)))

	state := newState(t,
		starlark.String("count"), starlark.MakeInt(23),
		starlark.String("ratio"), starlark.Float(0.25),
		starlark.String("seen"), newSet(t, starlark.String("a"), starlark.String("b")),
		starlark.String("last"), m,
		starlark.String("window"), starlark.NewList([]starlark.Value{
			starlark.Tuple{starlark.MakeInt(1), starlark.Float(0.5)},
			starlark.Tuple{starlark.MakeInt(2), starlark.Float(1.5)},
		}),
		starlark.String("lookup"), newState(t,
			starlark.MakeInt(1), starlark.String("one"),
			starlark.None, starlark.Bool(true),
		),
	)

	serialized, err := marshal(state)
	require.NoError(t, err)

	restored, err := unmarshal(serialized)
	require.NoError(t, err)
	require.Equal(t, state.Keys(), restored.Keys())

	again, err := marshal(restored)
	require.NoError(t, err)
	require.Equal(t, serialized, again)

	for _, item := range state.Items() {
		actual, found, err := restored.Get(item[0])
		require.NoError(t, err)
		require.True(t, found, "missing key %s", item[0])
		require.Equal(t, item[1].Type(), actual.Type(), "type of %s changed", item[0])
		require.Equal(t, item[1].String(), actual.String(), "value of %s changed", item[0])
	}
}

// newState creates a state dictionary from the given key-value pairs
func newState(t *testing.T, pairs ...starlark.Value) *starlark.Dict {
	t.Helper()
	require.Zero(t, len(pairs)%2, "expected key-value pairs")

	state := starlark.NewDict(len(pairs) / 2)
	for i := 0; i < len(pairs); i += 2 {
		require.NoError(t, state.SetKey(pairs[i], pairs[i+1]))
	}

	return state
}

// newSet creates a set containing the given elements
func newSet(t *testing.T, elements ...starlark.Value) *starlark.Set {
	t.Helper()

	set := starlark.NewSet(len(elements))
	for _, element := range elements {
		require.NoError(t, set.Insert(element))
	}

	return set
}

// nest wraps the given value into the requested number of lists
func nest(value starlark.Value, depth int) starlark.Value {
	for range depth {
		value = starlark.NewList([]starlark.Value{value})
	}

	return value
}
