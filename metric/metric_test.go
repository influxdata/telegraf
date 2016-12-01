package metric

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/stretchr/testify/assert"
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
		"float":  float64(1),
		"int":    int64(1),
		"bool":   true,
		"false":  false,
		"string": "test",
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
			expected: 5,
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

func TestNewMetricPoint(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	p := m.Point()

	assert.Equal(t, fields, m.Fields())
	assert.Equal(t, fields, p.Fields())
	assert.Equal(t, "cpu", p.Name())
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
