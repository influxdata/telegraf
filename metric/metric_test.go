package metric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
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
	m := New("cpu", tags, fields, now)

	require.Equal(t, "cpu", m.Name())
	require.Equal(t, tags, m.Tags())
	require.Equal(t, fields, m.Fields())
	require.Len(t, m.FieldList(), 2)
	require.Equal(t, now, m.Time())
}

// cpu value=1
func baseMetric() telegraf.Metric {
	tags := map[string]string{}
	fields := map[string]interface{}{
		"value": float64(1),
	}
	now := time.Now()

	m := New("cpu", tags, fields, now)
	return m
}

func TestHasTag(t *testing.T) {
	m := baseMetric()

	require.False(t, m.HasTag("host"))
	m.AddTag("host", "localhost")
	require.True(t, m.HasTag("host"))
	m.RemoveTag("host")
	require.False(t, m.HasTag("host"))
}

func TestAddTagOverwrites(t *testing.T) {
	m := baseMetric()

	m.AddTag("host", "localhost")
	m.AddTag("host", "example.org")

	value, ok := m.GetTag("host")
	require.True(t, ok)
	require.Equal(t, "example.org", value)
	require.Len(t, m.TagList(), 1)
}

func TestRemoveTagNoEffectOnMissingTags(t *testing.T) {
	m := baseMetric()

	m.RemoveTag("foo")
	m.AddTag("a", "x")
	m.RemoveTag("foo")
	m.RemoveTag("bar")
	value, ok := m.GetTag("a")
	require.True(t, ok)
	require.Equal(t, "x", value)
}

func TestGetTag(t *testing.T) {
	m := baseMetric()

	_, ok := m.GetTag("host")
	require.False(t, ok)

	m.AddTag("host", "localhost")

	value, ok := m.GetTag("host")
	require.True(t, ok)
	require.Equal(t, "localhost", value)

	m.RemoveTag("host")
	_, ok = m.GetTag("host")
	require.False(t, ok)
}

func TestHasField(t *testing.T) {
	m := baseMetric()

	require.False(t, m.HasField("x"))
	m.AddField("x", 42.0)
	require.True(t, m.HasField("x"))
	m.RemoveTag("x")
	require.False(t, m.HasTag("x"))
}

func TestAddFieldOverwrites(t *testing.T) {
	m := baseMetric()

	m.AddField("value", 1.0)
	m.AddField("value", 42.0)

	require.Len(t, m.FieldList(), 1)

	value, ok := m.GetField("value")
	require.True(t, ok)
	require.Equal(t, 42.0, value)
}

func TestAddFieldChangesType(t *testing.T) {
	m := baseMetric()

	m.AddField("value", 1.0)
	m.AddField("value", "xyzzy")

	require.Len(t, m.FieldList(), 1)

	value, ok := m.GetField("value")
	require.True(t, ok)
	require.Equal(t, "xyzzy", value)
}

func TestRemoveFieldNoEffectOnMissingFields(t *testing.T) {
	m := baseMetric()

	m.RemoveField("foo")
	m.AddField("a", "x")
	m.RemoveField("foo")
	m.RemoveField("bar")
	value, ok := m.GetField("a")
	require.True(t, ok)
	require.Equal(t, "x", value)
}

func TestGetField(t *testing.T) {
	m := baseMetric()

	_, ok := m.GetField("foo")
	require.False(t, ok)

	m.AddField("foo", "bar")

	value, ok := m.GetField("foo")
	require.True(t, ok)
	require.Equal(t, "bar", value)

	m.RemoveTag("foo")
	_, ok = m.GetTag("foo")
	require.False(t, ok)
}

func TestTagList_Sorted(t *testing.T) {
	m := baseMetric()

	m.AddTag("b", "y")
	m.AddTag("c", "z")
	m.AddTag("a", "x")

	taglist := m.TagList()
	require.Equal(t, "a", taglist[0].Key)
	require.Equal(t, "b", taglist[1].Key)
	require.Equal(t, "c", taglist[2].Key)
}

func TestEquals(t *testing.T) {
	now := time.Now()
	m1 := New("cpu",
		map[string]string{
			"host": "localhost",
		},
		map[string]interface{}{
			"value": 42.0,
		},
		now,
	)

	m2 := New("cpu",
		map[string]string{
			"host": "localhost",
		},
		map[string]interface{}{
			"value": 42.0,
		},
		now,
	)

	lhs := m1.(*metric)
	require.Equal(t, lhs, m2)

	m3 := m2.Copy()
	require.Equal(t, lhs, m3)
	m3.AddTag("a", "x")
	require.NotEqual(t, lhs, m3)
}

func TestHashID(t *testing.T) {
	m := New(
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
	require.Equal(t, hash, m.HashID())

	// removing a non-existent tag doesn't change the hash:
	m.RemoveTag("no-op")
	require.Equal(t, hash, m.HashID())

	// adding a tag does change it:
	m.AddTag("foo", "bar")
	require.NotEqual(t, hash, m.HashID())
	hash = m.HashID()

	// removing a tag also changes it:
	m.RemoveTag("mytag")
	require.NotEqual(t, hash, m.HashID())
}

func TestHashID_Consistency(t *testing.T) {
	m := New(
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

	m2 := New(
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
	require.Equal(t, hash, m2.HashID())

	m3 := m.Copy()
	require.Equal(t, m2.HashID(), m3.HashID())
}

func TestHashID_Delimiting(t *testing.T) {
	m1 := New(
		"cpu",
		map[string]string{
			"a": "x",
			"b": "y",
			"c": "z",
		},
		map[string]interface{}{
			"value": float64(1),
		},
		time.Now(),
	)
	m2 := New(
		"cpu",
		map[string]string{
			"a": "xbycz",
		},
		map[string]interface{}{
			"value": float64(1),
		},
		time.Now(),
	)
	require.NotEqual(t, m1.HashID(), m2.HashID())
}

func TestSetName(t *testing.T) {
	m := baseMetric()
	m.SetName("foo")
	require.Equal(t, "foo", m.Name())
}

func TestAddPrefix(t *testing.T) {
	m := baseMetric()
	m.AddPrefix("foo_")
	require.Equal(t, "foo_cpu", m.Name())
	m.AddPrefix("foo_")
	require.Equal(t, "foo_foo_cpu", m.Name())
}

func TestAddSuffix(t *testing.T) {
	m := baseMetric()
	m.AddSuffix("_foo")
	require.Equal(t, "cpu_foo", m.Name())
	m.AddSuffix("_foo")
	require.Equal(t, "cpu_foo_foo", m.Name())
}

func TestValueType(t *testing.T) {
	now := time.Now()

	tags := map[string]string{}
	fields := map[string]interface{}{
		"value": float64(42),
	}
	m := New("cpu", tags, fields, now, telegraf.Gauge)

	require.Equal(t, telegraf.Gauge, m.Type())
}
