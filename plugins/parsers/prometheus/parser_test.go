package prometheus

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

const (
	validUniqueGauge = `# HELP cadvisor_version_info A metric with a constant '1' value labeled by kernel version, OS version, docker version, cadvisor version & cadvisor revision.
# TYPE cadvisor_version_info gauge
cadvisor_version_info{cadvisorRevision="",cadvisorVersion="",dockerVersion="1.8.2",kernelVersion="3.10.0-229.20.1.el7.x86_64",osVersion="CentOS Linux 7 (Core)"} 1
`
	validUniqueCounter = `# HELP get_token_fail_count Counter of failed Token() requests to the alternate token source
# TYPE get_token_fail_count counter
get_token_fail_count 0
`

	validUniqueSummary = `# HELP http_request_duration_microseconds The HTTP request latencies in microseconds.
# TYPE http_request_duration_microseconds summary
http_request_duration_microseconds{handler="prometheus",quantile="0.5"} 552048.506
http_request_duration_microseconds{handler="prometheus",quantile="0.9"} 5.876804288e+06
http_request_duration_microseconds{handler="prometheus",quantile="0.99"} 5.876804288e+06
http_request_duration_microseconds_sum{handler="prometheus"} 1.8909097205e+07
http_request_duration_microseconds_count{handler="prometheus"} 9
`

	validUniqueHistogram = `# HELP apiserver_request_latencies Response latency distribution in microseconds for each verb, resource and client.
# TYPE apiserver_request_latencies histogram
apiserver_request_latencies_bucket{resource="bindings",verb="POST",le="125000"} 1994
apiserver_request_latencies_bucket{resource="bindings",verb="POST",le="250000"} 1997
apiserver_request_latencies_bucket{resource="bindings",verb="POST",le="500000"} 2000
apiserver_request_latencies_bucket{resource="bindings",verb="POST",le="1e+06"} 2005
apiserver_request_latencies_bucket{resource="bindings",verb="POST",le="2e+06"} 2012
apiserver_request_latencies_bucket{resource="bindings",verb="POST",le="4e+06"} 2017
apiserver_request_latencies_bucket{resource="bindings",verb="POST",le="8e+06"} 2024
apiserver_request_latencies_bucket{resource="bindings",verb="POST",le="+Inf"} 2025
apiserver_request_latencies_sum{resource="bindings",verb="POST"} 1.02726334e+08
apiserver_request_latencies_count{resource="bindings",verb="POST"} 2025
`
)

func TestParsingValidGauge(t *testing.T) {
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"osVersion":        "CentOS Linux 7 (Core)",
				"cadvisorRevision": "",
				"cadvisorVersion":  "",
				"dockerVersion":    "1.8.2",
				"kernelVersion":    "3.10.0-229.20.1.el7.x86_64",
			},
			map[string]interface{}{
				"cadvisor_version_info": float64(1),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	metrics, err := parse([]byte(validUniqueGauge))

	require.NoError(t, err)
	require.Len(t, metrics, 1)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestParsingValidCounter(t *testing.T) {
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus",
			map[string]string{},
			map[string]interface{}{
				"get_token_fail_count": float64(0),
			},
			time.Unix(0, 0),
			telegraf.Counter,
		),
	}

	metrics, err := parse([]byte(validUniqueCounter))

	require.NoError(t, err)
	require.Len(t, metrics, 1)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestParsingValidSummary(t *testing.T) {
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"handler": "prometheus",
			},
			map[string]interface{}{
				"http_request_duration_microseconds_sum":   float64(1.8909097205e+07),
				"http_request_duration_microseconds_count": float64(9.0),
			},
			time.Unix(0, 0),
			telegraf.Summary,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"handler":  "prometheus",
				"quantile": "0.5",
			},
			map[string]interface{}{
				"http_request_duration_microseconds": float64(552048.506),
			},
			time.Unix(0, 0),
			telegraf.Summary,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"handler":  "prometheus",
				"quantile": "0.9",
			},
			map[string]interface{}{
				"http_request_duration_microseconds": float64(5.876804288e+06),
			},
			time.Unix(0, 0),
			telegraf.Summary,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"handler":  "prometheus",
				"quantile": "0.99",
			},
			map[string]interface{}{
				"http_request_duration_microseconds": float64(5.876804288e+6),
			},
			time.Unix(0, 0),
			telegraf.Summary,
		),
	}

	metrics, err := parse([]byte(validUniqueSummary))

	require.NoError(t, err)
	require.Len(t, metrics, 4)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestParsingValidHistogram(t *testing.T) {
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"verb":     "POST",
				"resource": "bindings",
			},
			map[string]interface{}{
				"apiserver_request_latencies_count": float64(2025.0),
				"apiserver_request_latencies_sum":   float64(1.02726334e+08),
			},
			time.Unix(0, 0),
			telegraf.Histogram,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"verb":     "POST",
				"resource": "bindings",
				"le":       "125000",
			},
			map[string]interface{}{
				"apiserver_request_latencies_bucket": float64(1994.0),
			},
			time.Unix(0, 0),
			telegraf.Histogram,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"verb":     "POST",
				"resource": "bindings",
				"le":       "250000",
			},
			map[string]interface{}{
				"apiserver_request_latencies_bucket": float64(1997.0),
			},
			time.Unix(0, 0),
			telegraf.Histogram,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"verb":     "POST",
				"resource": "bindings",
				"le":       "500000",
			},
			map[string]interface{}{
				"apiserver_request_latencies_bucket": float64(2000.0),
			},
			time.Unix(0, 0),
			telegraf.Histogram,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"verb":     "POST",
				"resource": "bindings",
				"le":       "1e+06",
			},
			map[string]interface{}{
				"apiserver_request_latencies_bucket": float64(2005.0),
			},
			time.Unix(0, 0),
			telegraf.Histogram,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"verb":     "POST",
				"resource": "bindings",
				"le":       "2e+06",
			},
			map[string]interface{}{
				"apiserver_request_latencies_bucket": float64(2012.0),
			},
			time.Unix(0, 0),
			telegraf.Histogram,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"verb":     "POST",
				"resource": "bindings",
				"le":       "4e+06",
			},
			map[string]interface{}{
				"apiserver_request_latencies_bucket": float64(2017.0),
			},
			time.Unix(0, 0),
			telegraf.Histogram,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"verb":     "POST",
				"resource": "bindings",
				"le":       "8e+06",
			},
			map[string]interface{}{
				"apiserver_request_latencies_bucket": float64(2024.0),
			},
			time.Unix(0, 0),
			telegraf.Histogram,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"verb":     "POST",
				"resource": "bindings",
				"le":       "+Inf",
			},
			map[string]interface{}{
				"apiserver_request_latencies_bucket": float64(2025.0),
			},
			time.Unix(0, 0),
			telegraf.Histogram,
		),
	}

	metrics, err := parse([]byte(validUniqueHistogram))

	require.NoError(t, err)
	require.Len(t, metrics, 9)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestDefautTags(t *testing.T) {
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"osVersion":        "CentOS Linux 7 (Core)",
				"cadvisorRevision": "",
				"cadvisorVersion":  "",
				"dockerVersion":    "1.8.2",
				"kernelVersion":    "3.10.0-229.20.1.el7.x86_64",
				"defaultTag":       "defaultTagValue",
			},
			map[string]interface{}{
				"cadvisor_version_info": float64(1),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	parser := Parser{
		DefaultTags: map[string]string{
			"defaultTag":    "defaultTagValue",
			"dockerVersion": "to_be_overriden",
		},
	}
	metrics, err := parser.Parse([]byte(validUniqueGauge))

	require.NoError(t, err)
	require.Len(t, metrics, 1)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestMetricsWithTimestamp(t *testing.T) {
	testTime := time.Date(2020, time.October, 4, 17, 0, 0, 0, time.UTC)
	testTimeUnix := testTime.UnixNano() / int64(time.Millisecond)
	metricsWithTimestamps := fmt.Sprintf(`
# TYPE test_counter counter
test_counter{label="test"} 1 %d
`, testTimeUnix)
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"label": "test",
			},
			map[string]interface{}{
				"test_counter": float64(1.0),
			},
			testTime,
			telegraf.Counter,
		),
	}

	metrics, _ := parse([]byte(metricsWithTimestamps))

	testutil.RequireMetricsEqual(t, expected, metrics, testutil.SortMetrics())
}

func TestMetricsWithoutIgnoreTimestamp(t *testing.T) {
	testTime := time.Date(2020, time.October, 4, 17, 0, 0, 0, time.UTC)
	testTimeUnix := testTime.UnixNano() / int64(time.Millisecond)
	metricsWithTimestamps := fmt.Sprintf(`
# TYPE test_counter counter
test_counter{label="test"} 1 %d
`, testTimeUnix)
	expected := testutil.MustMetric(
		"prometheus",
		map[string]string{
			"label": "test",
		},
		map[string]interface{}{
			"test_counter": float64(1.0),
		},
		testTime,
		telegraf.Counter,
	)

	parser := Parser{IgnoreTimestamp: true}
	metric, _ := parser.ParseLine(metricsWithTimestamps)

	testutil.RequireMetricEqual(t, expected, metric, testutil.IgnoreTime(), testutil.SortMetrics())
	require.WithinDuration(t, time.Now(), metric.Time(), 5*time.Second)
}

func parse(buf []byte) ([]telegraf.Metric, error) {
	parser := Parser{}
	return parser.Parse(buf)
}

func TestParserProtobufHeader(t *testing.T) {
	var uClient = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"host": "omsk",
			},
			map[string]interface{}{
				"swap_free": 9.77911808e+08,
			},
			time.Unix(0, 0),
			2,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"host": "omsk",
			},
			map[string]interface{}{
				"swap_in": 2.031616e+06,
			},
			time.Unix(0, 0),
			1,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"host": "omsk",
			},
			map[string]interface{}{
				"swap_out": 1.579008e+07,
			},
			time.Unix(0, 0),
			1,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"host": "omsk",
			},
			map[string]interface{}{
				"swap_total": 9.93185792e+08,
			},
			time.Unix(0, 0),
			2,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"host": "omsk",
			},
			map[string]interface{}{
				"swap_used": 1.5273984e+07,
			},
			time.Unix(0, 0),
			2,
		),
		testutil.MustMetric(
			"prometheus",
			map[string]string{
				"host": "omsk",
			},
			map[string]interface{}{
				"swap_used_percent": 1.5378778193395661,
			},
			time.Unix(0, 0),
			2,
		),
	}
	sampleProtoBufData := []uint8{67, 10, 9, 115, 119, 97, 112, 95, 102, 114, 101, 101, 18, 25, 84, 101, 108, 101, 103, 114, 97, 102, 32, 99, 111, 108, 108, 101, 99, 116, 101, 100, 32, 109, 101, 116, 114, 105, 99, 24, 1, 34, 25, 10, 12, 10, 4, 104, 111, 115, 116, 18, 4, 111, 109, 115, 107, 18, 9, 9, 0, 0, 0, 0, 224, 36, 205, 65, 65, 10, 7, 115, 119, 97, 112, 95, 105, 110, 18, 25, 84, 101, 108, 101, 103, 114, 97, 102, 32, 99, 111, 108, 108, 101, 99, 116, 101, 100, 32, 109, 101, 116, 114, 105, 99, 24, 0, 34, 25, 10, 12, 10, 4, 104, 111, 115, 116, 18, 4, 111, 109, 115, 107, 26, 9, 9, 0, 0, 0, 0, 0, 0, 63, 65, 66, 10, 8, 115, 119, 97, 112, 95, 111, 117, 116, 18, 25, 84, 101, 108, 101, 103, 114, 97, 102, 32, 99, 111, 108, 108, 101, 99, 116, 101, 100, 32, 109, 101, 116, 114, 105, 99, 24, 0, 34, 25, 10, 12, 10, 4, 104, 111, 115, 116, 18, 4, 111, 109, 115, 107, 26, 9, 9, 0, 0, 0, 0, 0, 30, 110, 65, 68, 10, 10, 115, 119, 97, 112, 95, 116, 111, 116, 97, 108, 18, 25, 84, 101, 108, 101, 103, 114, 97, 102, 32, 99, 111, 108, 108, 101, 99, 116, 101, 100, 32, 109, 101, 116, 114, 105, 99, 24, 1, 34, 25, 10, 12, 10, 4, 104, 111, 115, 116, 18, 4, 111, 109, 115, 107, 18, 9, 9, 0, 0, 0, 0, 104, 153, 205, 65, 67, 10, 9, 115, 119, 97, 112, 95, 117, 115, 101, 100, 18, 25, 84, 101, 108, 101, 103, 114, 97, 102, 32, 99, 111, 108, 108, 101, 99, 116, 101, 100, 32, 109, 101, 116, 114, 105, 99, 24, 1, 34, 25, 10, 12, 10, 4, 104, 111, 115, 116, 18, 4, 111, 109, 115, 107, 18, 9, 9, 0, 0, 0, 0, 0, 34, 109, 65, 75, 10, 17, 115, 119, 97, 112, 95, 117, 115, 101, 100, 95, 112, 101, 114, 99, 101, 110, 116, 18, 25, 84, 101, 108, 101, 103, 114, 97, 102, 32, 99, 111, 108, 108, 101, 99, 116, 101, 100, 32, 109, 101, 116, 114, 105, 99, 24, 1, 34, 25, 10, 12, 10, 4, 104, 111, 115, 116, 18, 4, 111, 109, 115, 107, 18, 9, 9, 109, 234, 180, 197, 37, 155, 248, 63}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=delimited")
		_, err := w.Write(sampleProtoBufData)
		require.NoError(t, err)
	}))
	defer ts.Close()
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("unable to create new request '%s': %s", ts.URL, err)
	}
	var resp *http.Response
	resp, err = uClient.Do(req)
	if err != nil {
		t.Fatalf("error making HTTP request to %s: %s", ts.URL, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading body: %s", err)
	}
	parser := Parser{Header: resp.Header}
	metrics, err := parser.Parse(body)
	if err != nil {
		t.Fatalf("error reading metrics for %s: %s", ts.URL, err)
	}
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}
