package valuecounter

import (
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
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

func TestTracking(t *testing.T) {
	inputRaw := []telegraf.Metric{m1, m1, m2}
	expected := []telegraf.Metric{
		metric.New("m1", map[string]string{"foo": "bar"}, map[string]interface{}{"status_200": 2, "status_OK": 1}, time.Now()),
	}

	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}
	input := make([]telegraf.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
	}

	// Process expected metrics and compare with resulting metrics
	acc := &testutil.Accumulator{}
	plugin := NewTestValueCounter([]string{"status"})
	for _, m := range input {
		plugin.Add(m)
	}
	plugin.Push(acc)
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())

	// Simulate output acknowledging delivery
	for _, m := range actual {
		m.Accept()
	}

	// Check delivery
	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(input) == len(delivered)
	}, time.Second, 100*time.Millisecond, "%d delivered but %d expected", len(delivered), len(expected))
}
