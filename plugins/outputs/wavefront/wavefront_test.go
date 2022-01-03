package wavefront

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// default config used by Tests
func defaultWavefront() *Wavefront {
	return &Wavefront{
		Host:            "localhost",
		Port:            2878,
		Prefix:          "testWF.",
		SimpleFields:    false,
		MetricSeparator: ".",
		ConvertPaths:    true,
		ConvertBool:     true,
		UseRegex:        false,
		Log:             testutil.Logger{},
	}
}

func TestBuildMetrics(t *testing.T) {
	w := defaultWavefront()
	w.Prefix = "testthis."

	pathReplacer = strings.NewReplacer("_", w.MetricSeparator)

	testMetric1 := metric.New(
		"test.simple.metric",
		map[string]string{"tag1": "value1", "host": "testHost"},
		map[string]interface{}{"value": 123},
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	var timestamp int64 = 1257894000

	var metricTests = []struct {
		metric       telegraf.Metric
		metricPoints []MetricPoint
	}{
		{
			testutil.TestMetric(float64(1), "testing_just*a%metric:float", "metric2"),
			[]MetricPoint{
				{Metric: w.Prefix + "testing.just-a-metric-float", Value: 1, Timestamp: timestamp, Tags: map[string]string{"tag1": "value1"}},
				{Metric: w.Prefix + "testing.metric2", Value: 1, Timestamp: timestamp, Tags: map[string]string{"tag1": "value1"}},
			},
		},
		{
			testutil.TestMetric(float64(1), "testing_just/another,metric:float", "metric2"),
			[]MetricPoint{
				{Metric: w.Prefix + "testing.just-another-metric-float", Value: 1, Timestamp: timestamp, Tags: map[string]string{"tag1": "value1"}},
				{Metric: w.Prefix + "testing.metric2", Value: 1, Timestamp: timestamp, Tags: map[string]string{"tag1": "value1"}},
			},
		},
		{
			testMetric1,
			[]MetricPoint{{Metric: w.Prefix + "test.simple.metric", Value: 123, Timestamp: timestamp, Source: "testHost", Tags: map[string]string{"tag1": "value1"}}},
		},
	}

	for _, mt := range metricTests {
		ml := w.buildMetrics(mt.metric)
		for i, line := range ml {
			if mt.metricPoints[i].Metric != line.Metric || mt.metricPoints[i].Value != line.Value {
				t.Errorf("\nexpected\t%+v %+v\nreceived\t%+v %+v\n", mt.metricPoints[i].Metric, mt.metricPoints[i].Value, line.Metric, line.Value)
			}
		}
	}
}

func TestBuildMetricsStrict(t *testing.T) {
	w := defaultWavefront()
	w.Prefix = "testthis."
	w.UseStrict = true

	pathReplacer = strings.NewReplacer("_", w.MetricSeparator)

	var timestamp int64 = 1257894000

	var metricTests = []struct {
		metric       telegraf.Metric
		metricPoints []MetricPoint
	}{
		{
			testutil.TestMetric(float64(1), "testing_just*a%metric:float", "metric2"),
			[]MetricPoint{
				{Metric: w.Prefix + "testing.just-a-metric-float", Value: 1, Timestamp: timestamp, Tags: map[string]string{"tag1": "value1"}},
				{Metric: w.Prefix + "testing.metric2", Value: 1, Timestamp: timestamp, Tags: map[string]string{"tag1": "value1"}},
			},
		},
		{
			testutil.TestMetric(float64(1), "testing_just/another,metric:float", "metric2"),
			[]MetricPoint{
				{Metric: w.Prefix + "testing.just/another,metric-float", Value: 1, Timestamp: timestamp, Tags: map[string]string{"tag/1": "value1", "tag,2": "value2"}},
				{Metric: w.Prefix + "testing.metric2", Value: 1, Timestamp: timestamp, Tags: map[string]string{"tag/1": "value1", "tag,2": "value2"}},
			},
		},
	}

	for _, mt := range metricTests {
		ml := w.buildMetrics(mt.metric)
		for i, line := range ml {
			if mt.metricPoints[i].Metric != line.Metric || mt.metricPoints[i].Value != line.Value {
				t.Errorf("\nexpected\t%+v %+v\nreceived\t%+v %+v\n", mt.metricPoints[i].Metric, mt.metricPoints[i].Value, line.Metric, line.Value)
			}
		}
	}
}

func TestBuildMetricsWithSimpleFields(t *testing.T) {
	w := defaultWavefront()
	w.Prefix = "testthis."
	w.SimpleFields = true

	pathReplacer = strings.NewReplacer("_", w.MetricSeparator)

	testMetric1 := metric.New(
		"test.simple.metric",
		map[string]string{"tag1": "value1"},
		map[string]interface{}{"value": 123},
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	var metricTests = []struct {
		metric      telegraf.Metric
		metricLines []MetricPoint
	}{
		{
			testutil.TestMetric(float64(1), "testing_just*a%metric:float"),
			[]MetricPoint{{Metric: w.Prefix + "testing.just-a-metric-float.value", Value: 1}},
		},
		{
			testMetric1,
			[]MetricPoint{{Metric: w.Prefix + "test.simple.metric.value", Value: 123}},
		},
	}

	for _, mt := range metricTests {
		ml := w.buildMetrics(mt.metric)
		for i, line := range ml {
			if mt.metricLines[i].Metric != line.Metric || mt.metricLines[i].Value != line.Value {
				t.Errorf("\nexpected\t%+v %+v\nreceived\t%+v %+v\n", mt.metricLines[i].Metric, mt.metricLines[i].Value, line.Metric, line.Value)
			}
		}
	}
}

func TestBuildTags(t *testing.T) {
	w := defaultWavefront()

	var tagtests = []struct {
		ptIn      map[string]string
		outSource string
		outTags   map[string]string
	}{
		{
			map[string]string{},
			"",
			map[string]string{},
		},
		{
			map[string]string{"one": "two", "three": "four", "host": "testHost"},
			"testHost",
			map[string]string{"one": "two", "three": "four"},
		},
		{
			map[string]string{"aaa": "bbb", "host": "testHost"},
			"testHost",
			map[string]string{"aaa": "bbb"},
		},
		{
			map[string]string{"bbb": "789", "aaa": "123", "host": "testHost"},
			"testHost",
			map[string]string{"aaa": "123", "bbb": "789"},
		},
		{
			map[string]string{"host": "aaa", "dc": "bbb"},
			"aaa",
			map[string]string{"dc": "bbb"},
		},
		{
			map[string]string{"host": "aaa", "dc": "a*$a\\abbb\"som/et|hing else", "bad#k%e/y that*sho\\uld work": "value1"},
			"aaa",
			map[string]string{"dc": "a-$a\\abbb\"som/et|hing else", "bad-k-e-y-that-sho-uld-work": "value1"},
		},
	}

	for _, tt := range tagtests {
		source, tags := w.buildTags(tt.ptIn)
		if source != tt.outSource {
			t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", tt.outSource, source)
		}
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", tt.outTags, tags)
		}
	}
}

func TestBuildTagsWithSource(t *testing.T) {
	w := defaultWavefront()
	w.SourceOverride = []string{"snmp_host", "hostagent"}

	var tagtests = []struct {
		ptIn      map[string]string
		outSource string
		outTags   map[string]string
	}{
		{
			map[string]string{"host": "realHost"},
			"realHost",
			map[string]string{},
		},
		{
			map[string]string{"tag1": "value1", "host": "realHost"},
			"realHost",
			map[string]string{"tag1": "value1"},
		},
		{
			map[string]string{"snmp_host": "realHost", "host": "origHost"},
			"realHost",
			map[string]string{"telegraf_host": "origHost"},
		},
		{
			map[string]string{"hostagent": "realHost", "host": "origHost"},
			"realHost",
			map[string]string{"telegraf_host": "origHost"},
		},
		{
			map[string]string{"hostagent": "abc", "snmp_host": "realHost", "host": "origHost"},
			"realHost",
			map[string]string{"hostagent": "abc", "telegraf_host": "origHost"},
		},
		{
			map[string]string{"something": "abc", "host": "r*@l\"Ho/st"},
			"r-@l\"Ho/st",
			map[string]string{"something": "abc"},
		},
	}

	for _, tt := range tagtests {
		source, tags := w.buildTags(tt.ptIn)
		if source != tt.outSource {
			t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", tt.outSource, source)
		}
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", tt.outTags, tags)
		}
	}
}

func TestBuildValue(t *testing.T) {
	w := defaultWavefront()

	var valuetests = []struct {
		value interface{}
		name  string
		out   float64
		isErr bool
	}{
		{value: int64(123), out: 123},
		{value: uint64(456), out: 456},
		{value: float64(789), out: 789},
		{value: true, out: 1},
		{value: false, out: 0},
		{value: "bad", out: 0, isErr: true},
	}

	for _, vt := range valuetests {
		value, err := buildValue(vt.value, vt.name, w)
		if vt.isErr && err == nil {
			t.Errorf("\nexpected error with\t%+v\nreceived\t%+v\n", vt.out, value)
		} else if value != vt.out {
			t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", vt.out, value)
		}
	}
}

func TestBuildValueString(t *testing.T) {
	w := defaultWavefront()
	w.StringToNumber = map[string][]map[string]float64{
		"test1": {{"green": 1, "red": 10}},
		"test2": {{"active": 1, "hidden": 2}},
	}

	var valuetests = []struct {
		value interface{}
		name  string
		out   float64
		isErr bool
	}{
		{value: int64(123), name: "", out: 123},
		{value: "green", name: "test1", out: 1},
		{value: "red", name: "test1", out: 10},
		{value: "hidden", name: "test2", out: 2},
		{value: "bad", name: "test1", out: 0, isErr: true},
	}

	for _, vt := range valuetests {
		value, err := buildValue(vt.value, vt.name, w)
		if vt.isErr && err == nil {
			t.Errorf("\nexpected error with\t%+v\nreceived\t%+v\n", vt.out, value)
		} else if value != vt.out {
			t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", vt.out, value)
		}
	}
}

func TestTagLimits(t *testing.T) {
	w := defaultWavefront()
	w.TruncateTags = true

	// Should fail (all tags skipped)
	template := make(map[string]string)
	template[strings.Repeat("x", 255)] = "whatever"
	_, tags := w.buildTags(template)
	require.Empty(t, tags, "All tags should have been skipped")

	// Should truncate value
	template = make(map[string]string)
	longKey := strings.Repeat("x", 253)
	template[longKey] = "whatever"
	_, tags = w.buildTags(template)
	require.Contains(t, tags, longKey, "Should contain truncated long key")
	require.Equal(t, "w", tags[longKey])

	// Should not truncate
	template = make(map[string]string)
	longKey = strings.Repeat("x", 251)
	template[longKey] = "Hi!"
	_, tags = w.buildTags(template)
	require.Contains(t, tags, longKey, "Should contain non truncated long key")
	require.Equal(t, "Hi!", tags[longKey])

	// Turn off truncating and make sure it leaves the tags intact
	w.TruncateTags = false
	template = make(map[string]string)
	longKey = strings.Repeat("x", 255)
	template[longKey] = longKey
	_, tags = w.buildTags(template)
	require.Contains(t, tags, longKey, "Should contain non truncated long key")
	require.Equal(t, longKey, tags[longKey])
}

// Benchmarks to test performance of string replacement via Regex and Replacer
var testString = "this_is*my!test/string\\for=replacement"

func BenchmarkReplaceAllString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		sanitizedRegex.ReplaceAllString(testString, "-")
	}
}

func BenchmarkReplaceAllLiteralString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		sanitizedRegex.ReplaceAllLiteralString(testString, "-")
	}
}

func BenchmarkReplacer(b *testing.B) {
	for n := 0; n < b.N; n++ {
		sanitizedChars.Replace(testString)
	}
}
