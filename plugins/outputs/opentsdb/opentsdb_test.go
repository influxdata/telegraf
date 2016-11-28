package opentsdb

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	//"github.com/stretchr/testify/require"
)

func TestCleanTags(t *testing.T) {
	var tagtests = []struct {
		ptIn    map[string]string
		outTags map[string]string
	}{
		{
			map[string]string{"one": "two", "three": "four"},
			map[string]string{"one": "two", "three": "four"},
		},
		{
			map[string]string{"aaa": "bbb"},
			map[string]string{"aaa": "bbb"},
		},
		{
			map[string]string{"Sp%ci@l Chars": "g$t repl#ced"},
			map[string]string{"Sp-ci-l_Chars": "g-t_repl-ced"},
		},
		{
			map[string]string{},
			map[string]string{},
		},
	}
	for _, tt := range tagtests {
		tags := cleanTags(tt.ptIn)
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected %+v\ngot %+v\n", tt.outTags, tags)
		}
	}
}

func TestBuildTagsTelnet(t *testing.T) {
	var tagtests = []struct {
		ptIn    map[string]string
		outTags string
	}{
		{
			map[string]string{"one": "two", "three": "four"},
			"one=two three=four",
		},
		{
			map[string]string{"aaa": "bbb"},
			"aaa=bbb",
		},
		{
			map[string]string{"one": "two", "aaa": "bbb"},
			"aaa=bbb one=two",
		},
		{
			map[string]string{},
			"",
		},
	}
	for _, tt := range tagtests {
		tags := ToLineFormat(tt.ptIn)
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected %+v\ngot %+v\n", tt.outTags, tags)
		}
	}
}

func BenchmarkHttpSend(b *testing.B) {
	const BatchSize = 50
	const MetricsCount = 4 * BatchSize
	metrics := make([]telegraf.Metric, MetricsCount)
	for i := 0; i < MetricsCount; i++ {
		metrics[i] = testutil.TestMetric(1.0)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, "{}")
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		panic(err)
	}

	_, p, _ := net.SplitHostPort(u.Host)

	port, err := strconv.Atoi(p)
	if err != nil {
		panic(err)
	}

	o := &OpenTSDB{
		Host:          ts.URL,
		Port:          port,
		Prefix:        "",
		HttpBatchSize: BatchSize,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o.Write(metrics)
	}
}

// func TestWrite(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration test in short mode")
// 	}

// 	o := &OpenTSDB{
// 		Host:   testutil.GetLocalHost(),
// 		Port:   4242,
// 		Prefix: "prefix.test.",
// 	}

// 	// Verify that we can connect to the OpenTSDB instance
// 	err := o.Connect()
// 	require.NoError(t, err)

// 	// Verify that we can successfully write data to OpenTSDB
// 	err = o.Write(testutil.MockMetrics())
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

// 	err = o.Write(metrics)
// 	require.NoError(t, err)
// }
