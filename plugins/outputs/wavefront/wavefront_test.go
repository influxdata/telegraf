package wavefront

import (
	"reflect"
	"testing"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/telegraf"
	"strings"
	"time"
)

func defaultWavefront() *Wavefront {
	return &Wavefront{
		Host: "localhost",
		Port: 2878,
		Prefix: "testWF.",
		SimpleFields: false,
		MetricSeparator: ".",
		ConvertPaths: true,
		UseRegex: false,
	}
}

func TestSourceTags(t *testing.T) {
	w := defaultWavefront()
	w.SourceOverride = []string{"snmp_host", "hostagent"}

	var tagtests = []struct {
		ptIn    map[string]string
		outTags []string
	}{
		{
			map[string]string{"snmp_host": "realHost", "host": "origHost"},
			[]string{"source=\"realHost\"", "telegraf_host=\"origHost\""},
		},
		{
			map[string]string{"hostagent": "realHost", "host": "origHost"},
			[]string{"source=\"realHost\"", "telegraf_host=\"origHost\""},
		},
		{
			map[string]string{"hostagent": "abc", "snmp_host": "realHost", "host": "origHost"},
			[]string{"hostagent=\"abc\"", "source=\"realHost\"", "telegraf_host=\"origHost\""},
		},
		{
			map[string]string{"something": "abc", "host": "realHost"},
			[]string{"something=\"abc\"", "source=\"realHost\"", "telegraf_host=\"realHost\""},
		},
	}
	for _, tt := range tagtests {
		tags := buildTags(tt.ptIn, w)
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", tt.outTags, tags)
		}
	}
}

func TestBuildMetricsNoSimpleFields(t *testing.T) {
	w := defaultWavefront()
	w.UseRegex = false
	w.Prefix = "testthis."
	w.SimpleFields = false

	pathReplacer = strings.NewReplacer("_", w.MetricSeparator)

	testMetric1, _ := telegraf.NewMetric(
		"test.simple.metric",
		map[string]string{"tag1": "value1"},
		map[string]interface{}{"value": 123},
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	var metricTests = []struct {
		metric      telegraf.Metric
		metricLines []MetricLine
	}{
		{
			testutil.TestMetric(float64(1.0), "testing_just*a%metric:float"),
			[]MetricLine{{Metric: w.Prefix + "testing.just-a-metric-float", Value: "1.000000"}},
		},
		{
			testMetric1,
			[]MetricLine{{Metric: w.Prefix + "test.simple.metric", Value: "123"}},
		},
	}

	for _, mt := range metricTests {
		ml := buildMetrics(mt.metric, w)
		for i, line := range ml {
			if mt.metricLines[i].Metric != line.Metric || mt.metricLines[i].Value != line.Value {
				t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", mt.metricLines[i].Metric + " " + mt.metricLines[i].Value, line.Metric + " " + line.Value)
			}
		}
	}

}

func TestBuildMetricsWithSimpleFields(t *testing.T) {
	w := defaultWavefront()
	w.UseRegex = false
	w.Prefix = "testthis."
	w.SimpleFields = true

	pathReplacer = strings.NewReplacer("_", w.MetricSeparator)

	testMetric1, _ := telegraf.NewMetric(
		"test.simple.metric",
		map[string]string{"tag1": "value1"},
		map[string]interface{}{"value": 123},
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	var metricTests = []struct {
		metric      telegraf.Metric
		metricLines []MetricLine
	}{
		{
			testutil.TestMetric(float64(1.0), "testing_just*a%metric:float"),
			[]MetricLine{{Metric: w.Prefix + "testing.just-a-metric-float.value", Value: "1.000000"}},
		},
		{
			testMetric1,
			[]MetricLine{{Metric: w.Prefix + "test.simple.metric.value", Value: "123"}},
		},
	}

	for _, mt := range metricTests {
		ml := buildMetrics(mt.metric, w)
		for i, line := range ml {
			if mt.metricLines[i].Metric != line.Metric || mt.metricLines[i].Value != line.Value {
				t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", mt.metricLines[i].Metric + " " + mt.metricLines[i].Value, line.Metric + " " + line.Value)
			}
		}
	}

}

func TestBuildTags(t *testing.T) {

	w := defaultWavefront()

	var tagtests = []struct {
		ptIn    map[string]string
		outTags []string
	}{
		{
			map[string]string{"one": "two", "three": "four", "host": "testHost"},
			[]string{"one=\"two\"", "source=\"testHost\"", "telegraf_host=\"testHost\"", "three=\"four\""},
		},
		{
			map[string]string{"aaa": "bbb", "host": "testHost"},
			[]string{"aaa=\"bbb\"", "source=\"testHost\"", "telegraf_host=\"testHost\""},
		},
		{
			map[string]string{"bbb": "789", "aaa": "123", "host": "testHost"},
			[]string{"aaa=\"123\"", "bbb=\"789\"", "source=\"testHost\"", "telegraf_host=\"testHost\""},
		},
		{
			map[string]string{"host": "aaa", "dc": "bbb"},
			[]string{"dc=\"bbb\"", "source=\"aaa\"", "telegraf_host=\"aaa\""},
		},
		{
			map[string]string{"Sp%ci@l Chars": "\"g*t repl#ced", "host": "testHost"},
			[]string{"Sp-ci-l-Chars=\"\\\"g-t repl#ced\"", "source=\"testHost\"", "telegraf_host=\"testHost\""},
		},
	}
	for _, tt := range tagtests {
		tags := buildTags(tt.ptIn, w)
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", tt.outTags, tags)
		}
	}
}

// func TestWrite(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration test in short mode")
// 	}

// 	w := &Wavefront{
// 		Host:   testutil.GetLocalHost(),
// 		Port:   2878,
// 		Prefix: "prefix.test.",
// 	}

// 	// Verify that we can connect to the Wavefront instance
// 	err := w.Connect()
// 	require.NoError(t, err)

// 	// Verify that we can successfully write data to Wavefront
// 	err = w.Write(testutil.MockMetrics())
// 	require.NoError(t, err)

// 	// Verify postive and negative test cases of writing data
// 	metrics := testutil.MockMetrics()
// 	metrics = append(metrics, testutil.TestMetric(float64(1.0),
// 		"justametric.float"))
// 	metrics = append(metrics, testutil.TestMetric(int64(123456789),
// 		"justametric.int"))
// 	metrics = append(metrics, testutil.TestMetric(uint64(123456789012345),
// 		"justametric.uint"))
// 	metrics = append(metrics, testutil.TestMetric("Lorem Ipsum",
// 		"justametric.string"))
// 	metrics = append(metrics, testutil.TestMetric(float64(42.0),
// 		"justametric.anotherfloat"))
// 	metrics = append(metrics, testutil.TestMetric(float64(42.0),
// 		"metric w/ specialchars"))

// 	err = w.Write(metrics)
// 	require.NoError(t, err)
// }

