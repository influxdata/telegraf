package firstlast

import (
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFirstLastSimple(t *testing.T) {
	acc := testutil.Accumulator{}
	fl := NewFirstLast()
	fl.startTime = (time.Unix(1530939906, 0))

	tags := map[string]string{"foo": "bar"}
	m1, _ := metric.New("m1",
		tags,
		map[string]interface{}{"a": int64(1)},
		time.Unix(1530939936, 0))
	m2, _ := metric.New("m1",
		tags,
		map[string]interface{}{"a": int64(2)},
		time.Unix(1530939937, 0))
	m3, _ := metric.New("m1",
		tags,
		map[string]interface{}{"a": int64(3)},
		time.Unix(1530939938, 0))
	fl.Add(m1)
	fl.Add(m2)
	fl.Add(m3)
	fl.Push(&acc)
	assert.True(t, acc.HasPoint("m1", tags, "a_first", int64(1)))
	assert.True(t, acc.HasPoint("m1", tags, "a_last", int64(3)))
	assert.Equal(t, 2, int(acc.NMetrics()))
}

func TestFirstLastDisableLast(t *testing.T) {
	acc := testutil.Accumulator{}
	fl := NewFirstLast()
	fl.EnableFirst = false
	fl.startTime = (time.Unix(1530939906, 0))

	tags := map[string]string{"foo": "bar"}
	m1, _ := metric.New("m1",
		tags,
		map[string]interface{}{"a": int64(1)},
		time.Unix(1530939936, 0))
	m2, _ := metric.New("m1",
		tags,
		map[string]interface{}{"a": int64(2)},
		time.Unix(1530939937, 0))
	m3, _ := metric.New("m1",
		tags,
		map[string]interface{}{"a": int64(3)},
		time.Unix(1530939938, 0))
	fl.Add(m1)
	fl.Add(m2)
	fl.Add(m3)
	fl.Push(&acc)
	assert.False(t, acc.HasField("m1", "a_first"))
	assert.True(t, acc.HasPoint("m1", tags, "a_last", int64(3)))
	assert.Equal(t, 1, int(acc.NMetrics()))
}

func TestFirstLastDisableFirst(t *testing.T) {
	acc := testutil.Accumulator{}
	fl := NewFirstLast()
	fl.EnableLast = false
	fl.startTime = (time.Unix(1530939906, 0))

	tags := map[string]string{"foo": "bar"}
	m1, _ := metric.New("m1",
		tags,
		map[string]interface{}{"a": int64(1)},
		time.Unix(1530939936, 0))
	m2, _ := metric.New("m1",
		tags,
		map[string]interface{}{"a": int64(2)},
		time.Unix(1530939937, 0))
	m3, _ := metric.New("m1",
		tags,
		map[string]interface{}{"a": int64(3)},
		time.Unix(1530939938, 0))
	fl.Add(m1)
	fl.Add(m2)
	fl.Add(m3)
	fl.Push(&acc)
	assert.True(t, acc.HasPoint("m1", tags, "a_first", int64(1)))
	assert.False(t, acc.HasField("m1", "a_last"))
	assert.Equal(t, 1, int(acc.NMetrics()))
}

func TestFirstLastTwoTags(t *testing.T) {
	acc := testutil.Accumulator{}
	fl := NewFirstLast()

	fl.startTime = (time.Unix(1530939906, 0))

	tags1 := map[string]string{"foo": "bar"}
	tags2 := map[string]string{"foo": "baz"}

	m1, _ := metric.New("m1",
		tags1,
		map[string]interface{}{"a": int64(1)},
		time.Unix(1530939936, 0))
	m2, _ := metric.New("m1",
		tags2,
		map[string]interface{}{"a": int64(2)},
		time.Unix(1530939937, 0))
	m3, _ := metric.New("m1",
		tags1,
		map[string]interface{}{"a": int64(3)},
		time.Unix(1530939938, 0))
	fl.Add(m1)
	fl.Add(m2)
	fl.Add(m3)
	fl.Push(&acc)
	assert.True(t, acc.HasPoint("m1", tags1, "a_first", int64(1)))
	assert.True(t, acc.HasPoint("m1", tags1, "a_last", int64(3)))
	assert.True(t, acc.HasPoint("m1", tags2, "a_first", int64(2)))
	assert.True(t, acc.HasPoint("m1", tags2, "a_last", int64(2)))
	assert.Equal(t, 4, int(acc.NMetrics()))
}

func TestFirstLastLongDifference(t *testing.T) {
	acc := testutil.Accumulator{}
	fl := NewFirstLast()
	fl.startTime = time.Now().Add(time.Second * -500)
	tags := map[string]string{"foo": "bar"}

	m1, _ := metric.New("m",
		tags,
		map[string]interface{}{"a": int64(1)},
		time.Now().Add(time.Second*-290))
	m2, _ := metric.New("m",
		tags,
		map[string]interface{}{"a": int64(2)},
		time.Now().Add(time.Second*-275))
	m3, _ := metric.New("m",
		tags,
		map[string]interface{}{"a": int64(3)},
		time.Now().Add(time.Second*-100))
	m4, _ := metric.New("m",
		tags,
		map[string]interface{}{"a": int64(4)},
		time.Now().Add(time.Second*-20))
	fl.Add(m1)
	fl.Add(m2)
	fl.Push(&acc)
	fl.Add(m3)
	fl.Push(&acc)
	fl.Add(m4)
	fl.Push(&acc)

	assert.True(t, acc.HasPoint("m", tags, "a_first", int64(1)))
	assert.True(t, acc.HasPoint("m", tags, "a_last", int64(2)))
	assert.True(t, acc.HasPoint("m", tags, "a_first", int64(3)))
	assert.True(t, acc.HasPoint("m", tags, "a_last", int64(3)))
	assert.True(t, acc.HasPoint("m", tags, "a_first", int64(4)))
	assert.Equal(t, 5, int(acc.NMetrics()))
}
