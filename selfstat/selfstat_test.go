package selfstat

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

// testCleanup resets the global registry for test cleanup & unlocks the test lock
func testCleanup() {
	registry = &Registry{
		stats: make(map[uint64]map[string]Stat),
	}
}

func BenchmarkStats(b *testing.B) {
	defer testCleanup()
	b1 := Register("benchmark1", "test_field1", map[string]string{"test": "foo"})
	for n := 0; n < b.N; n++ {
		b1.Incr(1)
		b1.Incr(3)
		b1.Get()
	}
}

func BenchmarkTimingStats(b *testing.B) {
	defer testCleanup()
	b2 := RegisterTiming("benchmark2", "test_field1", map[string]string{"test": "foo"})
	for n := 0; n < b.N; n++ {
		b2.Incr(1)
		b2.Incr(3)
		b2.Get()
	}
}

func TestRegisterAndIncrAndSet(t *testing.T) {
	defer testCleanup()
	s1 := Register("test", "test_field1", map[string]string{"test": "foo"})
	s2 := Register("test", "test_field2", map[string]string{"test": "foo"})
	require.Equal(t, int64(0), s1.Get())

	s1.Incr(10)
	s1.Incr(5)
	require.Equal(t, int64(15), s1.Get())

	s1.Set(12)
	require.Equal(t, int64(12), s1.Get())

	s1.Incr(-2)
	require.Equal(t, int64(10), s1.Get())

	s2.Set(101)
	require.Equal(t, int64(101), s2.Get())

	// make sure that the same field returns the same metric
	// this one should be the same as s2.
	foo := Register("test", "test_field2", map[string]string{"test": "foo"})
	require.Equal(t, int64(101), foo.Get())

	// check that tags are consistent
	require.Equal(t, map[string]string{"test": "foo"}, foo.Tags())
	require.Equal(t, "internal_test", foo.Name())
}

func TestRegisterTimingAndIncrAndSet(t *testing.T) {
	defer testCleanup()
	s1 := RegisterTiming("test", "test_field1_ns", map[string]string{"test": "foo"})
	s2 := RegisterTiming("test", "test_field2_ns", map[string]string{"test": "foo"})
	require.Equal(t, int64(0), s1.Get())

	s1.Incr(10)
	s1.Incr(5)
	require.Equal(t, int64(7), s1.Get())
	// previous value is used on subsequent calls to Get()
	require.Equal(t, int64(7), s1.Get())

	s1.Set(12)
	require.Equal(t, int64(12), s1.Get())

	s1.Incr(-2)
	require.Equal(t, int64(-2), s1.Get())

	s2.Set(101)
	require.Equal(t, int64(101), s2.Get())

	// make sure that the same field returns the same metric
	// this one should be the same as s2.
	foo := RegisterTiming("test", "test_field2_ns", map[string]string{"test": "foo"})
	require.Equal(t, int64(101), foo.Get())

	// check that tags are consistent
	require.Equal(t, map[string]string{"test": "foo"}, foo.Tags())
	require.Equal(t, "internal_test", foo.Name())
}

func TestStatKeyConsistency(t *testing.T) {
	lhs := key("internal_stats", map[string]string{
		"foo":   "bar",
		"bar":   "baz",
		"whose": "first",
	})
	rhs := key("internal_stats", map[string]string{
		"foo":   "bar",
		"bar":   "baz",
		"whose": "first",
	})
	require.Equal(t, lhs, rhs)
}

func TestRegisterMetricsAndVerify(t *testing.T) {
	defer testCleanup()

	// register two metrics with the same key
	s1 := RegisterTiming("test_timing", "test_field1_ns", map[string]string{"test": "foo"})
	s2 := RegisterTiming("test_timing", "test_field2_ns", map[string]string{"test": "foo"})
	s1.Incr(10)
	s2.Incr(15)
	require.Len(t, Metrics(), 1)

	// register two more metrics with different keys
	s3 := RegisterTiming("test_timing", "test_field1_ns", map[string]string{"test": "bar"})
	s4 := RegisterTiming("test_timing", "test_field2_ns", map[string]string{"test": "baz"})
	s3.Incr(10)
	s4.Incr(15)
	require.Len(t, Metrics(), 3)

	// register some non-timing metrics
	s5 := Register("test", "test_field1", map[string]string{"test": "bar"})
	s6 := Register("test", "test_field2", map[string]string{"test": "baz"})
	Register("test", "test_field3", map[string]string{"test": "baz"})
	s5.Incr(10)
	s5.Incr(18)
	s6.Incr(15)
	require.Len(t, Metrics(), 5)

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

func TestRegisterCopy(t *testing.T) {
	tags := map[string]string{"input": "mem", "alias": "mem1"}
	stat := Register("gather", "metrics_gathered", tags)
	tags["new"] = "value"
	require.NotEqual(t, tags, stat.Tags())
}
