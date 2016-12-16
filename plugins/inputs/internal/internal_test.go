package internal

import (
	"testing"

	"github.com/influxdata/telegraf/selfstat"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

func TestSelfPlugin(t *testing.T) {
	s := NewSelf()
	acc := &testutil.Accumulator{}

	s.Gather(acc)
	assert.True(t, acc.HasMeasurement("internal_memstats"))

	// test that a registered stat is incremented
	stat := selfstat.Register("mytest", "test", map[string]string{"test": "foo"})
	stat.Incr(1)
	stat.Incr(2)
	s.Gather(acc)
	acc.AssertContainsTaggedFields(t, "internal_mytest",
		map[string]interface{}{
			"test": int64(3),
		},
		map[string]string{
			"test": "foo",
		},
	)
	acc.ClearMetrics()

	// test that a registered stat is set properly
	stat.Set(101)
	s.Gather(acc)
	acc.AssertContainsTaggedFields(t, "internal_mytest",
		map[string]interface{}{
			"test": int64(101),
		},
		map[string]string{
			"test": "foo",
		},
	)
	acc.ClearMetrics()

	// test that regular and timing stats can share the same measurement, and
	// that timings are set properly.
	timing := selfstat.RegisterTiming("mytest", "test_ns", map[string]string{"test": "foo"})
	timing.Incr(100)
	timing.Incr(200)
	s.Gather(acc)
	acc.AssertContainsTaggedFields(t, "internal_mytest",
		map[string]interface{}{
			"test":    int64(101),
			"test_ns": int64(150),
		},
		map[string]string{
			"test": "foo",
		},
	)
}
