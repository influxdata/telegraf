package metric

import (
	"fmt"
	"math"
	"regexp"
	"testing"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetric(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host":       "localhost",
		"datacenter": "us-east-1",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
		"usage_busy": float64(1),
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	assert.Equal(t, telegraf.Untyped, m.Type())
	assert.Equal(t, tags, m.Tags())
	assert.Equal(t, fields, m.Fields())
	assert.Equal(t, "cpu", m.Name())
	assert.Equal(t, now, m.Time())
	assert.Equal(t, now.UnixNano(), m.UnixNano())
}

func TestNewErrors(t *testing.T) {
	// creating a metric with an empty name produces an error:
	m, err := New(
		"",
		map[string]string{
			"datacenter": "us-east-1",
			"mytag":      "foo",
			"another":    "tag",
		},
		map[string]interface{}{
			"value": float64(1),
		},
		time.Now(),
	)
	assert.Error(t, err)
	assert.Nil(t, m)

	// creating a metric with empty fields produces an error:
	m, err = New(
		"foobar",
		map[string]string{
			"datacenter": "us-east-1",
			"mytag":      "foo",
			"another":    "tag",
		},
		map[string]interface{}{},
		time.Now(),
	)
	assert.Error(t, err)
	assert.Nil(t, m)
}

func TestNewMetric_Tags(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"datacenter": "us-east-1",
	}
	fields := map[string]interface{}{
		"value": float64(1),
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	assert.True(t, m.HasTag("host"))
	assert.True(t, m.HasTag("datacenter"))

	m.AddTag("newtag", "foo")
	assert.True(t, m.HasTag("newtag"))

	m.RemoveTag("host")
	assert.False(t, m.HasTag("host"))
	assert.True(t, m.HasTag("newtag"))
	assert.True(t, m.HasTag("datacenter"))

	m.RemoveTag("datacenter")
	assert.False(t, m.HasTag("datacenter"))
	assert.True(t, m.HasTag("newtag"))
	assert.Equal(t, map[string]string{"newtag": "foo"}, m.Tags())

	m.RemoveTag("newtag")
	assert.False(t, m.HasTag("newtag"))
	assert.Equal(t, map[string]string{}, m.Tags())

	assert.Equal(t, "cpu value=1 "+fmt.Sprint(now.UnixNano())+"\n", m.String())
}

func TestSerialize(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"datacenter": "us-east-1",
	}
	fields := map[string]interface{}{
		"value": float64(1),
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	assert.Equal(t,
		[]byte("cpu,datacenter=us-east-1 value=1 "+fmt.Sprint(now.UnixNano())+"\n"),
		m.Serialize())

	m.RemoveTag("datacenter")
	assert.Equal(t,
		[]byte("cpu value=1 "+fmt.Sprint(now.UnixNano())+"\n"),
		m.Serialize())
}

func TestHashID(t *testing.T) {
	m, _ := New(
		"cpu",
		map[string]string{
			"datacenter": "us-east-1",
			"mytag":      "foo",
			"another":    "tag",
		},
		map[string]interface{}{
			"value": float64(1),
		},
		time.Now(),
	)
	hash := m.HashID()

	// adding a field doesn't change the hash:
	m.AddField("foo", int64(100))
	assert.Equal(t, hash, m.HashID())

	// removing a non-existent tag doesn't change the hash:
	m.RemoveTag("no-op")
	assert.Equal(t, hash, m.HashID())

	// adding a tag does change it:
	m.AddTag("foo", "bar")
	assert.NotEqual(t, hash, m.HashID())
	hash = m.HashID()

	// removing a tag also changes it:
	m.RemoveTag("mytag")
	assert.NotEqual(t, hash, m.HashID())
}

func TestHashID_Consistency(t *testing.T) {
	m, _ := New(
		"cpu",
		map[string]string{
			"datacenter": "us-east-1",
			"mytag":      "foo",
			"another":    "tag",
		},
		map[string]interface{}{
			"value": float64(1),
		},
		time.Now(),
	)
	hash := m.HashID()

	for i := 0; i < 1000; i++ {
		m2, _ := New(
			"cpu",
			map[string]string{
				"datacenter": "us-east-1",
				"mytag":      "foo",
				"another":    "tag",
			},
			map[string]interface{}{
				"value": float64(1),
			},
			time.Now(),
		)
		assert.Equal(t, hash, m2.HashID())
	}
}

func TestNewMetric_NameModifiers(t *testing.T) {
	now := time.Now()
	tags := map[string]string{}
	fields := map[string]interface{}{
		"value": float64(1),
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	hash := m.HashID()
	suffix := fmt.Sprintf(" value=1 %d\n", now.UnixNano())
	assert.Equal(t, "cpu"+suffix, m.String())

	m.SetPrefix("pre_")
	assert.NotEqual(t, hash, m.HashID())
	hash = m.HashID()
	assert.Equal(t, "pre_cpu"+suffix, m.String())

	m.SetSuffix("_post")
	assert.NotEqual(t, hash, m.HashID())
	hash = m.HashID()
	assert.Equal(t, "pre_cpu_post"+suffix, m.String())

	m.SetName("mem")
	assert.NotEqual(t, hash, m.HashID())
	assert.Equal(t, "mem"+suffix, m.String())
}

func TestNewMetric_FieldModifiers(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"value": float64(1),
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	assert.True(t, m.HasField("value"))
	assert.False(t, m.HasField("foo"))

	m.AddField("newfield", "foo")
	assert.True(t, m.HasField("newfield"))

	assert.NoError(t, m.RemoveField("newfield"))
	assert.False(t, m.HasField("newfield"))

	// don't allow user to remove all fields:
	assert.Error(t, m.RemoveField("value"))

	m.AddField("value2", int64(101))
	assert.NoError(t, m.RemoveField("value"))
	assert.False(t, m.HasField("value"))
}

func TestNewMetric_Fields(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"float":                  float64(1),
		"int":                    int64(1),
		"bool":                   true,
		"false":                  false,
		"string":                 "test",
		"quote_string":           `x"y`,
		"backslash_quote_string": `x\"y`,
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	assert.Equal(t, fields, m.Fields())
}

func TestNewMetric_Time(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"float":  float64(1),
		"int":    int64(1),
		"bool":   true,
		"false":  false,
		"string": "test",
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)
	m = m.Copy()
	m2 := m.Copy()

	assert.Equal(t, now.UnixNano(), m.Time().UnixNano())
	assert.Equal(t, now.UnixNano(), m2.UnixNano())
}

func TestNewMetric_Copy(t *testing.T) {
	now := time.Now()
	tags := map[string]string{}
	fields := map[string]interface{}{
		"float": float64(1),
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)
	m2 := m.Copy()

	assert.Equal(t,
		fmt.Sprintf("cpu float=1 %d\n", now.UnixNano()),
		m.String())
	m.AddTag("host", "localhost")
	assert.Equal(t,
		fmt.Sprintf("cpu,host=localhost float=1 %d\n", now.UnixNano()),
		m.String())

	assert.Equal(t,
		fmt.Sprintf("cpu float=1 %d\n", now.UnixNano()),
		m2.String())
}

func TestNewMetric_AllTypes(t *testing.T) {
	now := time.Now()
	tags := map[string]string{}
	fields := map[string]interface{}{
		"float64":     float64(1),
		"float32":     float32(1),
		"int64":       int64(1),
		"int32":       int32(1),
		"int16":       int16(1),
		"int8":        int8(1),
		"int":         int(1),
		"uint64":      uint64(1),
		"uint32":      uint32(1),
		"uint16":      uint16(1),
		"uint8":       uint8(1),
		"uint":        uint(1),
		"bytes":       []byte("foo"),
		"nil":         nil,
		"maxuint64":   uint64(MaxInt) + 10,
		"maxuint":     uint(MaxInt) + 10,
		"unsupported": []int{1, 2},
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	assert.Contains(t, m.String(), "float64=1")
	assert.Contains(t, m.String(), "float32=1")
	assert.Contains(t, m.String(), "int64=1i")
	assert.Contains(t, m.String(), "int32=1i")
	assert.Contains(t, m.String(), "int16=1i")
	assert.Contains(t, m.String(), "int8=1i")
	assert.Contains(t, m.String(), "int=1i")
	assert.Contains(t, m.String(), "uint64=1i")
	assert.Contains(t, m.String(), "uint32=1i")
	assert.Contains(t, m.String(), "uint16=1i")
	assert.Contains(t, m.String(), "uint8=1i")
	assert.Contains(t, m.String(), "uint=1i")
	assert.NotContains(t, m.String(), "nil")
	assert.Contains(t, m.String(), fmt.Sprintf("maxuint64=%di", MaxInt))
	assert.Contains(t, m.String(), fmt.Sprintf("maxuint=%di", MaxInt))
}

func TestIndexUnescapedByte(t *testing.T) {
	tests := []struct {
		in       []byte
		b        byte
		expected int
	}{
		{
			in:       []byte(`foobar`),
			b:        'b',
			expected: 3,
		},
		{
			in:       []byte(`foo\bar`),
			b:        'b',
			expected: -1,
		},
		{
			in:       []byte(`foo\\bar`),
			b:        'b',
			expected: -1,
		},
		{
			in:       []byte(`foobar`),
			b:        'f',
			expected: 0,
		},
		{
			in:       []byte(`foobar`),
			b:        'r',
			expected: 5,
		},
		{
			in:       []byte(`\foobar`),
			b:        'f',
			expected: -1,
		},
	}

	for _, test := range tests {
		got := indexUnescapedByte(test.in, test.b)
		assert.Equal(t, test.expected, got)
	}
}

func TestNewGaugeMetric(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host":       "localhost",
		"datacenter": "us-east-1",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
		"usage_busy": float64(1),
	}
	m, err := New("cpu", tags, fields, now, telegraf.Gauge)
	assert.NoError(t, err)

	assert.Equal(t, telegraf.Gauge, m.Type())
	assert.Equal(t, tags, m.Tags())
	assert.Equal(t, fields, m.Fields())
	assert.Equal(t, "cpu", m.Name())
	assert.Equal(t, now, m.Time())
	assert.Equal(t, now.UnixNano(), m.UnixNano())
}

func TestNewCounterMetric(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host":       "localhost",
		"datacenter": "us-east-1",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
		"usage_busy": float64(1),
	}
	m, err := New("cpu", tags, fields, now, telegraf.Counter)
	assert.NoError(t, err)

	assert.Equal(t, telegraf.Counter, m.Type())
	assert.Equal(t, tags, m.Tags())
	assert.Equal(t, fields, m.Fields())
	assert.Equal(t, "cpu", m.Name())
	assert.Equal(t, now, m.Time())
	assert.Equal(t, now.UnixNano(), m.UnixNano())
}

// test splitting metric into various max lengths
func TestSplitMetric(t *testing.T) {
	now := time.Unix(0, 1480940990034083306)
	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"float":  float64(100001),
		"int":    int64(100001),
		"bool":   true,
		"false":  false,
		"string": "test",
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	split80 := m.Split(80)
	assert.Len(t, split80, 2)

	split70 := m.Split(70)
	assert.Len(t, split70, 3)

	split60 := m.Split(60)
	assert.Len(t, split60, 5)
}

// test splitting metric into various max lengths
// use a simple regex check to verify that the split metrics are valid
func TestSplitMetric_RegexVerify(t *testing.T) {
	now := time.Unix(0, 1480940990034083306)
	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"foo":     float64(98934259085),
		"bar":     float64(19385292),
		"number":  float64(19385292),
		"another": float64(19385292),
		"n":       float64(19385292),
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	// verification regex
	re := regexp.MustCompile(`cpu,host=localhost \w+=\d+(,\w+=\d+)* 1480940990034083306`)

	split90 := m.Split(90)
	assert.Len(t, split90, 2)
	for _, splitM := range split90 {
		assert.True(t, re.Match(splitM.Serialize()), splitM.String())
	}

	split70 := m.Split(70)
	assert.Len(t, split70, 3)
	for _, splitM := range split70 {
		assert.True(t, re.Match(splitM.Serialize()), splitM.String())
	}

	split20 := m.Split(20)
	assert.Len(t, split20, 5)
	for _, splitM := range split20 {
		assert.True(t, re.Match(splitM.Serialize()), splitM.String())
	}
}

// test splitting metric even when given length is shorter than
// shortest possible length
// Split should split metric as short as possible, ie, 1 field per metric
func TestSplitMetric_TooShort(t *testing.T) {
	now := time.Unix(0, 1480940990034083306)
	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"float":  float64(100001),
		"int":    int64(100001),
		"bool":   true,
		"false":  false,
		"string": "test",
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	split := m.Split(10)
	assert.Len(t, split, 5)
	strings := make([]string, 5)
	for i, splitM := range split {
		strings[i] = splitM.String()
	}

	assert.Contains(t, strings, "cpu,host=localhost float=100001 1480940990034083306\n")
	assert.Contains(t, strings, "cpu,host=localhost int=100001i 1480940990034083306\n")
	assert.Contains(t, strings, "cpu,host=localhost bool=true 1480940990034083306\n")
	assert.Contains(t, strings, "cpu,host=localhost false=false 1480940990034083306\n")
	assert.Contains(t, strings, "cpu,host=localhost string=\"test\" 1480940990034083306\n")
}

func TestSplitMetric_NoOp(t *testing.T) {
	now := time.Unix(0, 1480940990034083306)
	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"float":  float64(100001),
		"int":    int64(100001),
		"bool":   true,
		"false":  false,
		"string": "test",
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	split := m.Split(1000)
	assert.Len(t, split, 1)
	assert.Equal(t, m, split[0])
}

func TestSplitMetric_OneField(t *testing.T) {
	now := time.Unix(0, 1480940990034083306)
	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"float": float64(100001),
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	assert.Equal(t, "cpu,host=localhost float=100001 1480940990034083306\n", m.String())

	split := m.Split(1000)
	assert.Len(t, split, 1)
	assert.Equal(t, "cpu,host=localhost float=100001 1480940990034083306\n", split[0].String())

	split = m.Split(1)
	assert.Len(t, split, 1)
	assert.Equal(t, "cpu,host=localhost float=100001 1480940990034083306\n", split[0].String())

	split = m.Split(40)
	assert.Len(t, split, 1)
	assert.Equal(t, "cpu,host=localhost float=100001 1480940990034083306\n", split[0].String())
}

func TestSplitMetric_ExactSize(t *testing.T) {
	now := time.Unix(0, 1480940990034083306)
	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"float":  float64(100001),
		"int":    int64(100001),
		"bool":   true,
		"false":  false,
		"string": "test",
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)
	actual := m.Split(m.Len())
	// check that no copy was made
	require.Equal(t, &m, &actual[0])
}

func TestSplitMetric_NoRoomForNewline(t *testing.T) {
	now := time.Unix(0, 1480940990034083306)
	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"float": float64(100001),
		"int":   int64(100001),
		"bool":  true,
		"false": false,
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)
	actual := m.Split(m.Len() - 1)
	require.Equal(t, 2, len(actual))
}

func TestNewMetricAggregate(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	assert.False(t, m.IsAggregate())
	m.SetAggregate(true)
	assert.True(t, m.IsAggregate())
}

func TestNewMetricString(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	lineProto := fmt.Sprintf("cpu,host=localhost usage_idle=99 %d\n",
		now.UnixNano())
	assert.Equal(t, lineProto, m.String())
}

func TestNewMetricFailNaN(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"usage_idle": math.NaN(),
	}

	_, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)
}

func TestEmptyTagValueOrKey(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host":     "localhost",
		"emptytag": "",
		"":         "valuewithoutkey",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
	}
	m, err := New("cpu", tags, fields, now)

	assert.True(t, m.HasTag("host"))
	assert.False(t, m.HasTag("emptytag"))
	assert.Equal(t,
		fmt.Sprintf("cpu,host=localhost usage_idle=99 %d\n", now.UnixNano()),
		m.String())

	assert.NoError(t, err)

}

func TestNewMetric_TrailingSlash(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		tags   map[string]string
		fields map[string]interface{}
	}{
		{
			name: `cpu\`,
			fields: map[string]interface{}{
				"value": int64(42),
			},
		},
		{
			name: "cpu",
			fields: map[string]interface{}{
				`value\`: "x",
			},
		},
		{
			name: "cpu",
			fields: map[string]interface{}{
				"value": `x\`,
			},
		},
		{
			name: "cpu",
			tags: map[string]string{
				`host\`: "localhost",
			},
			fields: map[string]interface{}{
				"value": int64(42),
			},
		},
		{
			name: "cpu",
			tags: map[string]string{
				"host": `localhost\`,
			},
			fields: map[string]interface{}{
				"value": int64(42),
			},
		},
	}

	for _, tc := range tests {
		_, err := New(tc.name, tc.tags, tc.fields, now)
		assert.Error(t, err)
	}
}
