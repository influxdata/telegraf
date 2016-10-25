package wavefront

import (
	"reflect"
	"testing"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/telegraf"
	"strings"
)

func defaultWavefront() *Wavefront {
	return &Wavefront{
		Host: "localhost",
		Port: 2878,
		Prefix: "testWF.",
		MetricSeparator: ".",
		ConvertPaths: true,
		UseRegex: false,
		Debug: true,
	}
}

func TestBuildMetrics(t *testing.T) {
	w := defaultWavefront()
	w.UseRegex = false
	w.Prefix = "testthis."
	pathReplacer = strings.NewReplacer("_", w.MetricSeparator)

	var metricTests = []struct {
		metric  telegraf.Metric
		metricLines []MetricLine
	} {
		{
			testutil.TestMetric(float64(1.0), "testing_just*a%metric:float"),
			[]MetricLine{{Metric: w.Prefix + "testing.just-a-metric-float.value"}},
		},
	}

	for _, mt := range metricTests {
		ml := buildMetrics(mt.metric, w)
		for i, line := range ml {
			if mt.metricLines[i].Metric != line.Metric {
				t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", mt.metricLines[i].Metric, line.Metric)
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
			map[string]string{"one": "two", "three": "four"},
			[]string{"one=\"two\"", "three=\"four\""},
		},
		{
			map[string]string{"aaa": "bbb"},
			[]string{"aaa=\"bbb\""},
		},
		{
			map[string]string{"one": "two", "aaa": "bbb"},
			[]string{"aaa=\"bbb\"", "one=\"two\""},
		},
		{
			map[string]string{"Sp%ci@l Chars": "g$t repl#ced"},
			[]string{"Sp-ci-l-Chars=\"g-t-repl-ced\""},
		},
		{
			map[string]string{},
			[]string{},
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
