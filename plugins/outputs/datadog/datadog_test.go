package datadog

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

var (
	fakeURL    = "http://test.datadog.com"
	fakeAPIKey = "123456"
)

func NewDatadog(url string) *Datadog {
	return &Datadog{
		URL: url,
		Log: testutil.Logger{},
	}
}

func fakeDatadog() *Datadog {
	d := NewDatadog(fakeURL)
	d.Apikey = fakeAPIKey
	return d
}

func TestUriOverride(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"status":"ok"}`) //nolint:errcheck // Ignore the returned error as the test will fail anyway
	}))
	defer ts.Close()

	d := NewDatadog(ts.URL)
	d.Apikey = "123456"
	err := d.Connect()
	require.NoError(t, err)
	err = d.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestCompressionOverride(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"status":"ok"}`) //nolint:errcheck // Ignore the returned error as the test will fail anyway
	}))
	defer ts.Close()

	d := NewDatadog(ts.URL)
	d.Apikey = "123456"
	d.Compression = "zlib"
	err := d.Connect()
	require.NoError(t, err)
	err = d.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestBadStatusCode(t *testing.T) {
	errorString := `{"errors": ["Something bad happened to the server.", "Your query made the server very sad."]}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, errorString)
	}))
	defer ts.Close()

	d := NewDatadog(ts.URL)
	d.Apikey = "123456"
	err := d.Connect()
	require.NoError(t, err)
	err = d.Write(testutil.MockMetrics())
	if err == nil {
		t.Errorf("error expected but none returned")
	} else {
		require.EqualError(t, err, fmt.Sprintf("received bad status code, %v: %s", http.StatusInternalServerError, errorString))
	}
}

func TestAuthenticatedUrl(t *testing.T) {
	d := fakeDatadog()

	authURL := d.authenticatedURL()
	require.EqualValues(t, fmt.Sprintf("%s?api_key=%s", fakeURL, fakeAPIKey), authURL)
}

func TestBuildTags(t *testing.T) {
	var tagtests = []struct {
		ptIn    []*telegraf.Tag
		outTags []string
	}{
		{
			[]*telegraf.Tag{
				{
					Key:   "one",
					Value: "two",
				},
				{
					Key:   "three",
					Value: "four",
				},
			},
			[]string{"one:two", "three:four"},
		},
		{
			[]*telegraf.Tag{
				{
					Key:   "aaa",
					Value: "bbb",
				},
			},
			[]string{"aaa:bbb"},
		},
		{
			[]*telegraf.Tag{},
			[]string{},
		},
	}
	for _, tt := range tagtests {
		tags := buildTags(tt.ptIn)
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected %+v\ngot %+v\n", tt.outTags, tags)
		}
	}
}

func TestBuildPoint(t *testing.T) {
	var tagtests = []struct {
		ptIn  telegraf.Metric
		outPt Point
		err   error
	}{
		{
			testutil.TestMetric(0.0, "test1"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				0.0,
			},
			nil,
		},
		{
			testutil.TestMetric(1.0, "test2"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				1.0,
			},
			nil,
		},
		{
			testutil.TestMetric(10, "test3"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				10.0,
			},
			nil,
		},
		{
			testutil.TestMetric(int32(112345), "test4"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				112345.0,
			},
			nil,
		},
		{
			testutil.TestMetric(int64(112345), "test5"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				112345.0,
			},
			nil,
		},
		{
			testutil.TestMetric(float32(11234.5), "test6"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				11234.5,
			},
			nil,
		},
		{
			testutil.TestMetric(true, "test7"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				1.0,
			},
			nil,
		},
		{
			testutil.TestMetric(false, "test8"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				0.0,
			},
			nil,
		},
		{
			testutil.TestMetric(int64(0), "test int64"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				0.0,
			},
			nil,
		},
		{
			testutil.TestMetric(uint64(0), "test uint64"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				0.0,
			},
			nil,
		},
		{
			testutil.TestMetric(true, "test bool"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				1.0,
			},
			nil,
		},
	}
	for _, tt := range tagtests {
		pt, err := buildMetrics(tt.ptIn)
		if err != nil && tt.err == nil {
			t.Errorf("%s: unexpected error, %+v\n", tt.ptIn.Name(), err)
		}
		if tt.err != nil && err == nil {
			t.Errorf("%s: expected an error (%s) but none returned", tt.ptIn.Name(), tt.err.Error())
		}
		if !reflect.DeepEqual(pt["value"], tt.outPt) && tt.err == nil {
			t.Errorf("%s: \nexpected %+v\ngot %+v\n",
				tt.ptIn.Name(), tt.outPt, pt["value"])
		}
	}
}

func TestVerifyValue(t *testing.T) {
	var tagtests = []struct {
		ptIn        telegraf.Metric
		validMetric bool
	}{
		{
			testutil.TestMetric(float32(11234.5), "test1"),
			true,
		},
		{
			testutil.TestMetric("11234.5", "test2"),
			false,
		},
	}
	for _, tt := range tagtests {
		ok := verifyValue(tt.ptIn.Fields()["value"])
		if tt.validMetric != ok {
			t.Errorf("%s: verification failed\n", tt.ptIn.Name())
		}
	}
}

func TestNaNIsSkipped(t *testing.T) {
	plugin := &Datadog{
		Apikey: "testing",
		URL:    "", // No request will be sent because all fields are skipped
	}

	err := plugin.Connect()
	require.NoError(t, err)

	err = plugin.Write([]telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"time_idle": math.NaN(),
			},
			time.Now()),
	})
	require.NoError(t, err)
}

func TestInfIsSkipped(t *testing.T) {
	plugin := &Datadog{
		Apikey: "testing",
		URL:    "", // No request will be sent because all fields are skipped
	}

	err := plugin.Connect()
	require.NoError(t, err)

	err = plugin.Write([]telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"time_idle": math.Inf(0),
			},
			time.Now()),
	})
	require.NoError(t, err)
}

func TestNonZeroRateIntervalConvertsRatesToCount(t *testing.T) {
	d := &Datadog{
		Apikey:       "123456",
		RateInterval: config.Duration(10 * time.Second),
	}

	var tests = []struct {
		name       string
		metricsIn  []telegraf.Metric
		metricsOut []*Metric
	}{
		{
			"convert counter metrics to rate",
			[]telegraf.Metric{
				testutil.MustMetric(
					"count_metric",
					map[string]string{
						"metric_type": "counter",
					},
					map[string]interface{}{
						"value": 100,
					},
					time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
					telegraf.Counter,
				),
			},
			[]*Metric{
				{
					Metric: "count_metric",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							10,
						},
					},
					Type: "rate",
					Tags: []string{
						"metric_type:counter",
					},
					Interval: 10,
				},
			},
		},
		{
			"convert count value in timing metrics to rate",
			[]telegraf.Metric{
				testutil.MustMetric(
					"timing_metric",
					map[string]string{
						"metric_type": "timing",
					},
					map[string]interface{}{
						"count":  1,
						"lower":  float64(10),
						"mean":   float64(10),
						"median": float64(10),
						"stddev": float64(0),
						"sum":    float64(10),
						"upper":  float64(10),
					},
					time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
					telegraf.Untyped,
				),
			},
			[]*Metric{
				{
					Metric: "timing_metric.count",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							0.1,
						},
					},
					Type: "rate",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 10,
				},
				{
					Metric: "timing_metric.lower",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
				{
					Metric: "timing_metric.mean",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
				{
					Metric: "timing_metric.median",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
				{
					Metric: "timing_metric.stddev",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(0),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
				{
					Metric: "timing_metric.sum",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
				{
					Metric: "timing_metric.upper",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
			},
		},
		{
			"convert count value in histogram metrics to rate",
			[]telegraf.Metric{
				testutil.MustMetric(
					"histogram_metric",
					map[string]string{
						"metric_type": "histogram",
					},
					map[string]interface{}{
						"count":  1,
						"lower":  float64(10),
						"mean":   float64(10),
						"median": float64(10),
						"stddev": float64(0),
						"sum":    float64(10),
						"upper":  float64(10),
					},
					time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
					telegraf.Untyped,
				),
			},
			[]*Metric{
				{
					Metric: "histogram_metric.count",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							0.1,
						},
					},
					Type: "rate",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 10,
				},
				{
					Metric: "histogram_metric.lower",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
				{
					Metric: "histogram_metric.mean",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
				{
					Metric: "histogram_metric.median",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
				{
					Metric: "histogram_metric.stddev",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(0),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
				{
					Metric: "histogram_metric.sum",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
				{
					Metric: "histogram_metric.upper",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualMetricsOut := d.convertToDatadogMetric(tt.metricsIn)
			require.ElementsMatch(t, tt.metricsOut, actualMetricsOut)
		})
	}
}

func TestZeroRateIntervalConvertsRatesToCount(t *testing.T) {
	d := &Datadog{
		Apikey: "123456",
	}

	var tests = []struct {
		name       string
		metricsIn  []telegraf.Metric
		metricsOut []*Metric
	}{
		{
			"does not convert counter metrics to rate",
			[]telegraf.Metric{
				testutil.MustMetric(
					"count_metric",
					map[string]string{
						"metric_type": "counter",
					},
					map[string]interface{}{
						"value": 100,
					},
					time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
					telegraf.Counter,
				),
			},
			[]*Metric{
				{
					Metric: "count_metric",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							100,
						},
					},
					Type: "count",
					Tags: []string{
						"metric_type:counter",
					},
					Interval: 1,
				},
			},
		},
		{
			"does not convert count value in timing metrics to rate",
			[]telegraf.Metric{
				testutil.MustMetric(
					"timing_metric",
					map[string]string{
						"metric_type": "timing",
					},
					map[string]interface{}{
						"count":  1,
						"lower":  float64(10),
						"mean":   float64(10),
						"median": float64(10),
						"stddev": float64(0),
						"sum":    float64(10),
						"upper":  float64(10),
					},
					time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
					telegraf.Untyped,
				),
			},
			[]*Metric{
				{
					Metric: "timing_metric.count",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							1,
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
				{
					Metric: "timing_metric.lower",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
				{
					Metric: "timing_metric.mean",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
				{
					Metric: "timing_metric.median",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
				{
					Metric: "timing_metric.stddev",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(0),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
				{
					Metric: "timing_metric.sum",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
				{
					Metric: "timing_metric.upper",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:timing",
					},
					Interval: 1,
				},
			},
		},
		{
			"does not convert count value in histogram metrics to rate",
			[]telegraf.Metric{
				testutil.MustMetric(
					"histogram_metric",
					map[string]string{
						"metric_type": "histogram",
					},
					map[string]interface{}{
						"count":  1,
						"lower":  float64(10),
						"mean":   float64(10),
						"median": float64(10),
						"stddev": float64(0),
						"sum":    float64(10),
						"upper":  float64(10),
					},
					time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
					telegraf.Untyped,
				),
			},
			[]*Metric{
				{
					Metric: "histogram_metric.count",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							1,
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
				{
					Metric: "histogram_metric.lower",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
				{
					Metric: "histogram_metric.mean",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
				{
					Metric: "histogram_metric.median",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
				{
					Metric: "histogram_metric.stddev",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(0),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
				{
					Metric: "histogram_metric.sum",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
				{
					Metric: "histogram_metric.upper",
					Points: [1]Point{
						{
							float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							float64(10),
						},
					},
					Type: "",
					Tags: []string{
						"metric_type:histogram",
					},
					Interval: 1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualMetricsOut := d.convertToDatadogMetric(tt.metricsIn)
			require.ElementsMatch(t, tt.metricsOut, actualMetricsOut)
		})
	}
}
