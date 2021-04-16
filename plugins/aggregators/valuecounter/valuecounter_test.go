package valuecounter

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// Create a valuecounter with config
func NewTestValueCounter(fields []string) telegraf.Aggregator {
	vc := &ValueCounter{
		Fields: fields,
	}
	vc.Reset()

	return vc
}

var m1 = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"status": 200,
		"foobar": "bar",
	},
	time.Now(),
)

var m2 = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"status":    "OK",
		"ignoreme":  "string",
		"andme":     true,
		"boolfield": false,
	},
	time.Now(),
)

func BenchmarkApply(b *testing.B) {
	vc := NewTestValueCounter([]string{"status"})

	for n := 0; n < b.N; n++ {
		vc.Add(m1)
		vc.Add(m2)
	}
}

// Test basic functionality
func TestBasic(t *testing.T) {
	vc := NewTestValueCounter([]string{"status"})
	acc := testutil.Accumulator{}

	vc.Add(m1)
	vc.Add(m2)
	vc.Add(m1)
	vc.Push(&acc)

	expectedFields := map[string]interface{}{
		"status_200": 2,
		"status_OK":  1,
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}

// Test with multiple fields to count
func TestMultipleFields(t *testing.T) {
	vc := NewTestValueCounter([]string{"status", "somefield", "boolfield"})
	acc := testutil.Accumulator{}

	vc.Add(m1)
	vc.Add(m2)
	vc.Add(m2)
	vc.Add(m1)
	vc.Push(&acc)

	expectedFields := map[string]interface{}{
		"status_200":      2,
		"status_OK":       2,
		"boolfield_false": 2,
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}

// Test with a reset between two runs
func TestWithReset(t *testing.T) {
	vc := NewTestValueCounter([]string{"status"})
	acc := testutil.Accumulator{}

	vc.Add(m1)
	vc.Add(m1)
	vc.Add(m2)
	vc.Push(&acc)

	expectedFields := map[string]interface{}{
		"status_200": 2,
		"status_OK":  1,
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)

	acc.ClearMetrics()
	vc.Reset()

	vc.Add(m2)
	vc.Add(m2)
	vc.Add(m1)
	vc.Push(&acc)

	expectedFields = map[string]interface{}{
		"status_200": 1,
		"status_OK":  2,
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}
