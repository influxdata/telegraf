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
	assert.True(t, acc.HasPoint("m1_first", tags, "a", int64(1)))
	assert.True(t, acc.HasPoint("m1_last", tags, "a", int64(3)))
	assert.Equal(t, 2, int(acc.NMetrics()))
}

func TestFirstLastTwoTags(t *testing.T) {
	acc := testutil.Accumulator{}
	fl := NewFirstLast()

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
	assert.True(t, acc.HasPoint("m1_first", tags1, "a", int64(1)))
	assert.True(t, acc.HasPoint("m1_last", tags1, "a", int64(3)))
	assert.True(t, acc.HasPoint("m1_first", tags2, "a", int64(2)))
	assert.True(t, acc.HasPoint("m1_last", tags2, "a", int64(2)))
	assert.Equal(t, 4, int(acc.NMetrics()))
}

func TestFirstLastLongDifference(t *testing.T) {
	acc := testutil.Accumulator{}
	fl := NewFirstLast()

	tags := map[string]string{"foo": "bar"}

	// m1 and m2 are outside the timeout window. Should get a single _first and _last.
	m1, _ := metric.New("m",
		tags,
		map[string]interface{}{"a": int64(1)},
		time.Now().Add(time.Second*-300))
	m2, _ := metric.New("m",
		tags,
		map[string]interface{}{"a": int64(2)},
		time.Now().Add(time.Second*-290))
	// Within the timeout window but outside warmup. We should get a "_first"
	m3, _ := metric.New("m",
		tags,
		map[string]interface{}{"a": int64(3)},
		time.Now().Add(time.Second*-45))
	// Within the warmup window. Shouldn't do anything.
	m4, _ := metric.New("m",
		tags,
		map[string]interface{}{"a": int64(3)},
		time.Now().Add(time.Second*-15))
	fl.Add(m1)
	fl.Add(m2)
	fl.Push(&acc)
	fl.Add(m3)
	fl.Push(&acc)
	fl.Add(m4)

	assert.True(t, acc.HasPoint("m_first", tags, "a", int64(1)))
	assert.True(t, acc.HasPoint("m_last", tags, "a", int64(2)))
	assert.True(t, acc.HasPoint("m_first", tags, "a", int64(3)))
	assert.Equal(t, 3, int(acc.NMetrics()))
}
