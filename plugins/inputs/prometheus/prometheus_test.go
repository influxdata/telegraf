package prometheus

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/fields"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

const sampleTextFormat = `# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0.00010425500000000001
go_gc_duration_seconds{quantile="0.25"} 0.000139108
go_gc_duration_seconds{quantile="0.5"} 0.00015749400000000002
go_gc_duration_seconds{quantile="0.75"} 0.000331463
go_gc_duration_seconds{quantile="1"} 0.000667154
go_gc_duration_seconds_sum 0.0018183950000000002
go_gc_duration_seconds_count 7
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 15
# HELP test_metric An untyped metric with a timestamp
# TYPE test_metric untyped
test_metric{label="value"} 1.0 1490802350000`

const sampleSummaryTextFormat = `# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0.00010425500000000001
go_gc_duration_seconds{quantile="0.25"} 0.000139108
go_gc_duration_seconds{quantile="0.5"} 0.00015749400000000002
go_gc_duration_seconds{quantile="0.75"} 0.000331463
go_gc_duration_seconds{quantile="1"} 0.000667154
go_gc_duration_seconds_sum 0.0018183950000000002
go_gc_duration_seconds_count 7`

const sampleGaugeTextFormat = `
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 15 1490802350000`

func TestPrometheusGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, sampleTextFormat)
		require.NoError(t, err)
	}))
	defer ts.Close()

	p := &Prometheus{
		Log:    testutil.Logger{},
		URLs:   []string{ts.URL},
		URLTag: "url",
	}
	err := p.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(p.Gather)
	require.NoError(t, err)

	require.True(t, acc.HasFloatField("go_gc_duration_seconds", "count"))
	require.True(t, acc.HasFloatField("go_goroutines", "gauge"))
	require.True(t, acc.HasFloatField("test_metric", "value"))
	require.True(t, acc.HasTimestamp("test_metric", time.Unix(1490802350, 0)))
	require.False(t, acc.HasTag("test_metric", "address"))
	require.True(t, acc.TagValue("test_metric", "url") == ts.URL+"/metrics")
}

func TestPrometheusCustomHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Header.Get("accept"))
		switch r.Header.Get("accept") {
		case "application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3":
			_, err := fmt.Fprintln(w, "proto 15 1490802540000")
			require.NoError(t, err)
		case "text/plain":
			_, err := fmt.Fprintln(w, "plain 42 1490802380000")
			require.NoError(t, err)
		default:
			_, err := fmt.Fprintln(w, "other 44 1490802420000")
			require.NoError(t, err)
		}
	}))
	defer ts.Close()

	tests := []struct {
		name                    string
		headers                 map[string]string
		expectedMeasurementName string
	}{
		{
			"default",
			map[string]string{},
			"proto",
		},
		{
			"plain text",
			map[string]string{
				"accept": "text/plain",
			},
			"plain",
		},
		{
			"other",
			map[string]string{
				"accept": "fakeACCEPTitem",
			},
			"other",
		},
	}

	for _, test := range tests {
		p := &Prometheus{
			Log:         testutil.Logger{},
			URLs:        []string{ts.URL},
			URLTag:      "url",
			HTTPHeaders: test.headers,
		}
		err := p.Init()
		require.NoError(t, err)

		var acc testutil.Accumulator
		require.NoError(t, acc.GatherError(p.Gather))
		require.Equal(t, test.expectedMeasurementName, acc.Metrics[0].Measurement)
	}
}

func TestPrometheusGeneratesMetricsWithHostNameTag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, sampleTextFormat)
		require.NoError(t, err)
	}))
	defer ts.Close()

	p := &Prometheus{
		Log:                testutil.Logger{},
		KubernetesServices: []string{ts.URL},
		URLTag:             "url",
	}
	err := p.Init()
	require.NoError(t, err)

	u, _ := url.Parse(ts.URL)
	tsAddress := u.Hostname()

	var acc testutil.Accumulator

	err = acc.GatherError(p.Gather)
	require.NoError(t, err)

	require.True(t, acc.HasFloatField("go_gc_duration_seconds", "count"))
	require.True(t, acc.HasFloatField("go_goroutines", "gauge"))
	require.True(t, acc.HasFloatField("test_metric", "value"))
	require.True(t, acc.HasTimestamp("test_metric", time.Unix(1490802350, 0)))
	require.True(t, acc.TagValue("test_metric", "address") == tsAddress)
	require.True(t, acc.TagValue("test_metric", "url") == ts.URL)
}

func TestPrometheusWithTimestamp(t *testing.T) {
	prommetric := `# HELP test_counter A sample test counter.
# TYPE test_counter counter
test_counter{label="test"} 1 1685443805885`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, prommetric)
		require.NoError(t, err)
	}))
	defer ts.Close()

	p := &Prometheus{
		Log:                testutil.Logger{},
		KubernetesServices: []string{ts.URL},
	}
	require.NoError(t, p.Init())

	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	tsAddress := u.Hostname()

	expected := []telegraf.Metric{
		metric.New(
			"test_counter",
			map[string]string{"address": tsAddress, "label": "test"},
			map[string]interface{}{"counter": float64(1.0)},
			time.UnixMilli(1685443805885),
			telegraf.Counter,
		),
	}

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(p.Gather))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestPrometheusGeneratesMetricsAlthoughFirstDNSFailsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, sampleTextFormat)
		require.NoError(t, err)
	}))
	defer ts.Close()

	p := &Prometheus{
		Log:                testutil.Logger{},
		URLs:               []string{ts.URL},
		KubernetesServices: []string{"http://random.telegraf.local:88/metrics"},
	}
	err := p.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(p.Gather)
	require.NoError(t, err)

	require.True(t, acc.HasFloatField("go_gc_duration_seconds", "count"))
	require.True(t, acc.HasFloatField("go_goroutines", "gauge"))
	require.True(t, acc.HasFloatField("test_metric", "value"))
	require.True(t, acc.HasTimestamp("test_metric", time.Unix(1490802350, 0)))
}

func TestPrometheusGeneratesMetricsSlowEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(4 * time.Second)
		_, err := fmt.Fprintln(w, sampleTextFormat)
		require.NoError(t, err)
	}))
	defer ts.Close()

	p := &Prometheus{
		Log:             testutil.Logger{},
		URLs:            []string{ts.URL},
		URLTag:          "url",
		ResponseTimeout: config.Duration(time.Second * 5),
	}
	err := p.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(p.Gather)
	require.NoError(t, err)

	require.True(t, acc.HasFloatField("go_gc_duration_seconds", "count"))
	require.True(t, acc.HasFloatField("go_goroutines", "gauge"))
	require.True(t, acc.HasFloatField("test_metric", "value"))
	require.True(t, acc.HasTimestamp("test_metric", time.Unix(1490802350, 0)))
	require.False(t, acc.HasTag("test_metric", "address"))
	require.True(t, acc.TagValue("test_metric", "url") == ts.URL+"/metrics")
}

func TestPrometheusGeneratesMetricsSlowEndpointHitTheTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(6 * time.Second)
		_, err := fmt.Fprintln(w, sampleTextFormat)
		require.NoError(t, err)
	}))
	defer ts.Close()

	p := &Prometheus{
		Log:             testutil.Logger{},
		URLs:            []string{ts.URL},
		URLTag:          "url",
		ResponseTimeout: config.Duration(time.Second * 5),
	}
	err := p.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(p.Gather)
	errMessage := fmt.Sprintf("error making HTTP request to \"%s/metrics\": Get \"%s/metrics\": "+
		"context deadline exceeded (Client.Timeout exceeded while awaiting headers)", ts.URL, ts.URL)
	errExpected := errors.New(errMessage)
	require.Error(t, err)
	require.Equal(t, errExpected.Error(), err.Error())
}

func TestPrometheusGeneratesMetricsSlowEndpointNewConfigParameter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(4 * time.Second)
		_, err := fmt.Fprintln(w, sampleTextFormat)
		require.NoError(t, err)
	}))
	defer ts.Close()

	p := &Prometheus{
		Log:    testutil.Logger{},
		URLs:   []string{ts.URL},
		URLTag: "url",
	}
	err := p.Init()
	require.NoError(t, err)
	p.client.Timeout = time.Second * 5

	var acc testutil.Accumulator

	err = acc.GatherError(p.Gather)
	require.NoError(t, err)

	require.True(t, acc.HasFloatField("go_gc_duration_seconds", "count"))
	require.True(t, acc.HasFloatField("go_goroutines", "gauge"))
	require.True(t, acc.HasFloatField("test_metric", "value"))
	require.True(t, acc.HasTimestamp("test_metric", time.Unix(1490802350, 0)))
	require.False(t, acc.HasTag("test_metric", "address"))
	require.True(t, acc.TagValue("test_metric", "url") == ts.URL+"/metrics")
}

func TestPrometheusGeneratesMetricsSlowEndpointHitTheTimeoutNewConfigParameter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(6 * time.Second)
		_, err := fmt.Fprintln(w, sampleTextFormat)
		require.NoError(t, err)
	}))
	defer ts.Close()

	p := &Prometheus{
		Log:    testutil.Logger{},
		URLs:   []string{ts.URL},
		URLTag: "url",
	}
	err := p.Init()
	require.NoError(t, err)
	p.client.Timeout = time.Second * 5

	var acc testutil.Accumulator

	err = acc.GatherError(p.Gather)
	require.ErrorContains(t, err, "error making HTTP request to \""+ts.URL+"/metrics\"")
}

func TestPrometheusGeneratesSummaryMetricsV2(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, sampleSummaryTextFormat)
		require.NoError(t, err)
	}))
	defer ts.Close()

	p := &Prometheus{
		Log:           &testutil.Logger{},
		URLs:          []string{ts.URL},
		URLTag:        "url",
		MetricVersion: 2,
	}
	err := p.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(p.Gather)
	require.NoError(t, err)

	require.True(t, acc.TagSetValue("prometheus", "quantile") == "0")
	require.True(t, acc.HasFloatField("prometheus", "go_gc_duration_seconds_sum"))
	require.True(t, acc.HasFloatField("prometheus", "go_gc_duration_seconds_count"))
	require.True(t, acc.TagValue("prometheus", "url") == ts.URL+"/metrics")
}

func TestSummaryMayContainNaN(t *testing.T) {
	const data = `# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} NaN
go_gc_duration_seconds{quantile="1"} NaN
go_gc_duration_seconds_sum 42.0
go_gc_duration_seconds_count 42`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, data)
		require.NoError(t, err)
	}))
	defer ts.Close()

	p := &Prometheus{
		Log:           &testutil.Logger{},
		URLs:          []string{ts.URL},
		URLTag:        "",
		MetricVersion: 2,
	}
	err := p.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = p.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"quantile": "0",
			},
			map[string]interface{}{
				"go_gc_duration_seconds": math.NaN(),
			},
			time.Unix(0, 0),
			telegraf.Summary,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"quantile": "1",
			},
			map[string]interface{}{
				"go_gc_duration_seconds": math.NaN(),
			},
			time.Unix(0, 0),
			telegraf.Summary,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{},
			map[string]interface{}{
				"go_gc_duration_seconds_sum":   42.0,
				"go_gc_duration_seconds_count": 42.0,
			},
			time.Unix(0, 0),
			telegraf.Summary,
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(),
		testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestPrometheusGeneratesGaugeMetricsV2(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, sampleGaugeTextFormat)
		require.NoError(t, err)
	}))
	defer ts.Close()

	p := &Prometheus{
		Log:           &testutil.Logger{},
		URLs:          []string{ts.URL},
		URLTag:        "url",
		MetricVersion: 2,
	}
	err := p.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(p.Gather)
	require.NoError(t, err)

	require.True(t, acc.HasFloatField("prometheus", "go_goroutines"))
	require.True(t, acc.TagValue("prometheus", "url") == ts.URL+"/metrics")
	require.True(t, acc.HasTimestamp("prometheus", time.Unix(1490802350, 0)))
}

func TestPrometheusGeneratesMetricsWithIgnoreTimestamp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, sampleTextFormat)
		require.NoError(t, err)
	}))
	defer ts.Close()

	p := &Prometheus{
		Log:             testutil.Logger{},
		URLs:            []string{ts.URL},
		URLTag:          "url",
		IgnoreTimestamp: true,
	}
	err := p.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(p.Gather)
	require.NoError(t, err)

	m, _ := acc.Get("test_metric")
	require.WithinDuration(t, time.Now(), m.Time, 5*time.Second)
}

func TestUnsupportedFieldSelector(t *testing.T) {
	fieldSelectorString := "spec.containerName=container"
	prom := &Prometheus{Log: testutil.Logger{}, KubernetesFieldSelector: fieldSelectorString}

	fieldSelector, _ := fields.ParseSelector(prom.KubernetesFieldSelector)
	isValid, invalidSelector := fieldSelectorIsSupported(fieldSelector)
	require.Equal(t, false, isValid)
	require.Equal(t, "spec.containerName", invalidSelector)
}

func TestInitConfigErrors(t *testing.T) {
	p := &Prometheus{
		MetricVersion:     2,
		Log:               testutil.Logger{},
		URLs:              nil,
		URLTag:            "url",
		MonitorPods:       true,
		PodScrapeScope:    "node",
		PodScrapeInterval: 60,
	}

	// Both invalid IP addresses
	t.Run("Both invalid IP addresses", func(t *testing.T) {
		p.NodeIP = "10.240.0.0.0"
		t.Setenv("NODE_IP", "10.000.0.0.0")
		err := p.Init()
		require.Error(t, err)
		expectedMessage := "the node_ip config and the environment variable NODE_IP are not set or invalid; " +
			"cannot get pod list for monitor_kubernetes_pods using node scrape scope"
		require.Equal(t, expectedMessage, err.Error())
	})

	t.Run("Valid IP address", func(t *testing.T) {
		t.Setenv("NODE_IP", "10.000.0.0")

		p.KubernetesLabelSelector = "label0==label0, label0 in (=)"
		err := p.Init()
		expectedMessage := "error parsing the specified label selector(s): unable to parse requirement: found '=', expected: ',', ')' or identifier"
		require.Error(t, err, expectedMessage)
		p.KubernetesLabelSelector = "label0==label"

		p.KubernetesFieldSelector = "field,"
		err = p.Init()
		expectedMessage = "error parsing the specified field selector(s): invalid selector: 'field,'; can't understand 'field'"
		require.Error(t, err, expectedMessage)

		p.KubernetesFieldSelector = "spec.containerNames=containerNames"
		err = p.Init()
		expectedMessage = "the field selector spec.containerNames is not supported for pods"
		require.Error(t, err, expectedMessage)
	})
}

func TestInitConfigSelectors(t *testing.T) {
	p := &Prometheus{
		MetricVersion:               2,
		Log:                         testutil.Logger{},
		URLs:                        nil,
		URLTag:                      "url",
		MonitorPods:                 true,
		MonitorKubernetesPodsMethod: MonitorMethodSettings,
		PodScrapeInterval:           60,
		KubernetesLabelSelector:     "app=test",
		KubernetesFieldSelector:     "spec.nodeName=node-0",
	}
	err := p.Init()
	require.NoError(t, err)

	require.NotNil(t, p.podLabelSelector)
	require.NotNil(t, p.podFieldSelector)
}
