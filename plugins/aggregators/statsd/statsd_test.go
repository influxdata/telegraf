package statsd

import (
	parser "github.com/influxdata/telegraf/plugins/parsers/statsd"
	"github.com/influxdata/telegraf/testutil"
	"strings"
	"testing"
)

type testMetric struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
}

func TestGauges(t *testing.T) {
	// Test that gauge +- values work
	input := []string{
		"plus.minus:100|g",
		"plus.minus:-10|g",
		"plus.minus:+30|g",
		"plus.plus:100|g",
		"plus.plus:+100|g",
		"plus.plus:+100|g",
		"minus.minus:100|g",
		"minus.minus:-100|g",
		"minus.minus:-100|g",
		"lone.plus:+100|g",
		"lone.minus:-100|g",
		"overwrite:100|g",
		"overwrite:300|g",
		"scientific.notation:4.696E+5|g",
		"scientific.notation.minus:4.7E-5|g",
	}
	validations := []struct {
		name  string
		value float64
	}{
		{
			"scientific_notation",
			469600,
		},
		{
			"scientific_notation_minus",
			0.000047,
		},
		{
			"plus_minus",
			120,
		},
		{
			"plus_plus",
			300,
		},
		{
			"minus_minus",
			-100,
		},
		{
			"lone_plus",
			100,
		},
		{
			"lone_minus",
			-100,
		},
		{
			"overwrite",
			300,
		},
	}

	testMetrics := make([]testMetric, len(validations))
	for i, validation := range validations {
		testMetrics[i] = testMetric{
			name: validation.name,
			fields: map[string]interface{}{
				"value": float64(validation.value),
			},
			tags: map[string]string{
				"statsd_type": "g",
			},
		}
	}

	assertAggreate(t, []byte(strings.Join(input, "\n")), testMetrics)
}

func TestSets(t *testing.T) {
	// Test that sets work
	input := []string{
		"unique.user.ids:100|s",
		"unique.user.ids:100|s",
		"unique.user.ids:100|s",
		"unique.user.ids:100|s",
		"unique.user.ids:100|s",
		"unique.user.ids:101|s",
		"unique.user.ids:102|s",
		"unique.user.ids:102|s",
		"unique.user.ids:123456789|s",
		"oneuser.id:100|s",
		"oneuser.id:100|s",
		"scientific.notation.sets:4.696E+5|s",
		"scientific.notation.sets:4.696E+5|s",
		"scientific.notation.sets:4.697E+5|s",
		"string.sets:foobar|s",
		"string.sets:foobar|s",
		"string.sets:bar|s",
	}

	validations := []struct {
		name  string
		value int64
	}{
		{
			"scientific_notation_sets",
			2,
		},
		{
			"unique_user_ids",
			4,
		},
		{
			"oneuser_id",
			1,
		},
		{
			"string_sets",
			2,
		},
	}

	testMetrics := make([]testMetric, len(validations))
	for i, validation := range validations {
		testMetrics[i] = testMetric{
			name: validation.name,
			fields: map[string]interface{}{
				"value": validation.value,
			},
			tags: map[string]string{
				"statsd_type": "s",
			},
		}
	}

	assertAggreate(t, []byte(strings.Join(input, "\n")), testMetrics)
}

func TestCounters(t *testing.T) {
	// Test that counters work
	input := []string{
		"small.inc:1|c",
		"big.inc:100|c",
		"big.inc:1|c",
		"big.inc:100000|c",
		"big.inc:1000000|c",
		"small.inc:1|c",
		"zero.init:0|c",
		"sample.rate:1|c|@0.1",
		"sample.rate:1|c",
		"scientific.notation:4.696E+5|c",
		"negative.test:100|c",
		"negative.test:-5|c",
	}

	validations := []struct {
		name  string
		value int64
	}{
		{
			"scientific_notation",
			469600,
		},
		{
			"small_inc",
			2,
		},
		{
			"big_inc",
			1100101,
		},
		{
			"zero_init",
			0,
		},
		{
			"sample_rate",
			11,
		},
		{
			"negative_test",
			95,
		},
	}

	testMetrics := make([]testMetric, len(validations))
	for i, validation := range validations {
		testMetrics[i] = testMetric{
			name: validation.name,
			fields: map[string]interface{}{
				"value": validation.value,
			},
			tags: map[string]string{
				"statsd_type": "c",
			},
		}
	}

	assertAggreate(t, []byte(strings.Join(input, "\n")), testMetrics)
}

func TestTiming(t *testing.T) {
	// Test that counters work
	input := []string{
		"test.timing:1|ms",
		"test.timing:11|ms",
		"test.timing:1|ms",
		"test.timing:1|ms",
		"test.timing:1|ms",
	}

	metric := testMetric{
		name: "test_timing",
		fields: map[string]interface{}{
			"90_percentile": float64(11),
			"count":         int64(5),
			"lower":         float64(1),
			"mean":          float64(3),
			"stddev":        float64(4),
			"sum":           float64(15),
			"upper":         float64(11),
		},
		tags: map[string]string{
			"statsd_type": "ms",
		},
	}

	assertAggreate(t, []byte(strings.Join(input, "\n")), []testMetric{metric})
}

func assertAggreate(t *testing.T, inputs []byte, testMetrics []testMetric) {
	testParser, _ := parser.NewParser("_", []string{}, nil)
	metrics, _ := testParser.Parse(inputs)
	agg := newTestAggregator()
	for _, metric := range metrics {
		agg.Add(metric)
	}
	acc := testutil.Accumulator{}
	agg.Push(&acc)
	for _, metric := range testMetrics {
		acc.AssertContainsTaggedFields(t, metric.name, metric.fields, metric.tags)
	}
}

func newTestAggregator() *Statsd {
	s := NewStatsd()
	s.PercentileLimit = 1000
	s.Percentiles = []int{90}

	s.DeleteTimings = true
	s.DeleteCounters = true
	s.DeleteGauges = true
	s.DeleteSets = true

	return s
}
