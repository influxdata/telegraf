package selfstat

import (
	"sync"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

var (
	// only allow one test at a time
	// this is because we are dealing with a global registry
	testLock sync.Mutex
	a        int64
)

// testCleanup resets the global registry for test cleanup & unlocks the test lock
func testCleanup() {
	registry = &rgstry{
		stats: make(map[uint64]map[string]Stat),
	}
	testLock.Unlock()
}

func BenchmarkStats(b *testing.B) {
	testLock.Lock()
	defer testCleanup()
	b1 := Register("benchmark1", "test_field1", map[string]string{"test": "foo"})
	for n := 0; n < b.N; n++ {
		b1.Incr(1)
		b1.Incr(3)
		a = b1.Get()
	}
}

func BenchmarkTimingStats(b *testing.B) {
	testLock.Lock()
	defer testCleanup()
	b2 := RegisterTiming("benchmark2", "test_field1", map[string]string{"test": "foo"})
	for n := 0; n < b.N; n++ {
		b2.Incr(1)
		b2.Incr(3)
		a = b2.Get()
	}
}

func TestRegisterAndIncrAndSet(t *testing.T) {
	testLock.Lock()
	defer testCleanup()
	s1 := Register("test", "test_field1", map[string]string{"test": "foo"})
	s2 := Register("test", "test_field2", map[string]string{"test": "foo"})
	assert.Equal(t, int64(0), s1.Get())

	s1.Incr(10)
	s1.Incr(5)
	assert.Equal(t, int64(15), s1.Get())

	s1.Set(12)
	assert.Equal(t, int64(12), s1.Get())

	s1.Incr(-2)
	assert.Equal(t, int64(10), s1.Get())

	s2.Set(101)
	assert.Equal(t, int64(101), s2.Get())

	// make sure that the same field returns the same metric
	// this one should be the same as s2.
	foo := Register("test", "test_field2", map[string]string{"test": "foo"})
	assert.Equal(t, int64(101), foo.Get())

	// check that tags are consistent
	assert.Equal(t, map[string]string{"test": "foo"}, foo.Tags())
	assert.Equal(t, "internal_test", foo.Name())
}

func TestRegisterTimingAndIncrAndSet(t *testing.T) {
	testLock.Lock()
	defer testCleanup()
	s1 := RegisterTiming("test", "test_field1_ns", map[string]string{"test": "foo"})
	s2 := RegisterTiming("test", "test_field2_ns", map[string]string{"test": "foo"})
	assert.Equal(t, int64(0), s1.Get())

	s1.Incr(10)
	s1.Incr(5)
	assert.Equal(t, int64(7), s1.Get())
	// previous value is used on subsequent calls to Get()
	assert.Equal(t, int64(7), s1.Get())

	s1.Set(12)
	assert.Equal(t, int64(12), s1.Get())

	s1.Incr(-2)
	assert.Equal(t, int64(-2), s1.Get())

	s2.Set(101)
	assert.Equal(t, int64(101), s2.Get())

	// make sure that the same field returns the same metric
	// this one should be the same as s2.
	foo := RegisterTiming("test", "test_field2_ns", map[string]string{"test": "foo"})
	assert.Equal(t, int64(101), foo.Get())

	// check that tags are consistent
	assert.Equal(t, map[string]string{"test": "foo"}, foo.Tags())
	assert.Equal(t, "internal_test", foo.Name())
}

func TestStatKeyConsistency(t *testing.T) {
	s := &stat{
		measurement: "internal_stat",
		field:       "myfield",
		tags: map[string]string{
			"foo":   "bar",
			"bar":   "baz",
			"whose": "first",
		},
	}
	k := s.Key()
	for i := 0; i < 5000; i++ {
		// assert that the Key() func doesn't change anything.
		assert.Equal(t, k, s.Key())

		// assert that two identical measurements always produce the same key.
		tmp := &stat{
			measurement: "internal_stat",
			field:       "myfield",
			tags: map[string]string{
				"foo":   "bar",
				"bar":   "baz",
				"whose": "first",
			},
		}
		assert.Equal(t, k, tmp.Key())
	}
}

func TestRegisterMetricsAndVerify(t *testing.T) {
	testLock.Lock()
	defer testCleanup()

	// register two metrics with the same key
	s1 := RegisterTiming("test_timing", "test_field1_ns", map[string]string{"test": "foo"})
	s2 := RegisterTiming("test_timing", "test_field2_ns", map[string]string{"test": "foo"})
	s1.Incr(10)
	s2.Incr(15)
	assert.Len(t, Metrics(), 1)

	// register two more metrics with different keys
	s3 := RegisterTiming("test_timing", "test_field1_ns", map[string]string{"test": "bar"})
	s4 := RegisterTiming("test_timing", "test_field2_ns", map[string]string{"test": "baz"})
	s3.Incr(10)
	s4.Incr(15)
	assert.Len(t, Metrics(), 3)

	// register some non-timing metrics
	s5 := Register("test", "test_field1", map[string]string{"test": "bar"})
	s6 := Register("test", "test_field2", map[string]string{"test": "baz"})
	Register("test", "test_field3", map[string]string{"test": "baz"})
	s5.Incr(10)
	s5.Incr(18)
	s6.Incr(15)
	assert.Len(t, Metrics(), 5)

	acc := testutil.Accumulator{}
	acc.AddMetrics(Metrics())

	// verify s1 & s2
	acc.AssertContainsTaggedFields(t, "internal_test_timing",
		map[string]interface{}{
			"test_field1_ns": int64(10),
			"test_field2_ns": int64(15),
		},
		map[string]string{
			"test": "foo",
		},
	)

	// verify s3
	acc.AssertContainsTaggedFields(t, "internal_test_timing",
		map[string]interface{}{
			"test_field1_ns": int64(10),
		},
		map[string]string{
			"test": "bar",
		},
	)

	// verify s4
	acc.AssertContainsTaggedFields(t, "internal_test_timing",
		map[string]interface{}{
			"test_field2_ns": int64(15),
		},
		map[string]string{
			"test": "baz",
		},
	)

	// verify s5
	acc.AssertContainsTaggedFields(t, "internal_test",
		map[string]interface{}{
			"test_field1": int64(28),
		},
		map[string]string{
			"test": "bar",
		},
	)

	// verify s6 & s7
	acc.AssertContainsTaggedFields(t, "internal_test",
		map[string]interface{}{
			"test_field2": int64(15),
			"test_field3": int64(0),
		},
		map[string]string{
			"test": "baz",
		},
	)
}
