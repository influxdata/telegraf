package prometheus

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
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

	validUniqueLine = `# HELP get_token_fail_count Counter of failed Token() requests to the alternate token source
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

	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestParsingValieCounter(t *testing.T) {
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

	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
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

	assert.NoError(t, err)
	assert.Len(t, metrics, 4)
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

	assert.NoError(t, err)
	assert.Len(t, metrics, 9)
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

	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
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
	sampleProtoBufData := []uint8{204, 1, 10, 22, 103, 111, 95, 103, 99, 95, 100, 117, 114, 97, 116, 105, 111, 110, 95, 115, 101, 99, 111, 110, 100, 115, 18, 61, 65, 32, 115, 117, 109, 109, 97, 114, 121, 32, 111, 102, 32, 116, 104, 101, 32, 112, 97, 117, 115, 101, 32, 100, 117, 114, 97, 116, 105, 111, 110, 32, 111, 102, 32, 103, 97, 114, 98, 97, 103, 101, 32, 99, 111, 108, 108, 101, 99, 116, 105, 111, 110, 32, 99, 121, 99, 108, 101, 115, 46, 24, 2, 34, 113, 34, 111, 8, 16, 17, 8, 212, 109, 25, 8, 179, 61, 63, 26, 18, 9, 0, 0, 0, 0, 0, 0, 0, 0, 17, 68, 183, 36, 40, 250, 83, 229, 62, 26, 18, 9, 0, 0, 0, 0, 0, 0, 208, 63, 17, 109, 43, 210, 209, 101, 194, 243, 62, 26, 18, 9, 0, 0, 0, 0, 0, 0, 224, 63, 17, 91, 16, 227, 152, 217, 165, 245, 62, 26, 18, 9, 0, 0, 0, 0, 0, 0, 232, 63, 17, 39, 18, 180, 24, 40, 104, 4, 63, 26, 18, 9, 0, 0, 0, 0, 0, 0, 240, 63, 17, 217, 170, 186, 205, 111, 38, 19, 63, 74, 10, 13, 103, 111, 95, 103, 111, 114, 111, 117, 116, 105, 110, 101, 115, 18, 42, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 103, 111, 114, 111, 117, 116, 105, 110, 101, 115, 32, 116, 104, 97, 116, 32, 99, 117, 114, 114, 101, 110, 116, 108, 121, 32, 101, 120, 105, 115, 116, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 0, 49, 64, 84, 10, 7, 103, 111, 95, 105, 110, 102, 111, 18, 37, 73, 110, 102, 111, 114, 109, 97, 116, 105, 111, 110, 32, 97, 98, 111, 117, 116, 32, 116, 104, 101, 32, 71, 111, 32, 101, 110, 118, 105, 114, 111, 110, 109, 101, 110, 116, 46, 24, 1, 34, 32, 10, 19, 10, 7, 118, 101, 114, 115, 105, 111, 110, 18, 8, 103, 111, 49, 46, 49, 52, 46, 53, 18, 9, 9, 0, 0, 0, 0, 0, 0, 240, 63, 85, 10, 23, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 97, 108, 108, 111, 99, 95, 98, 121, 116, 101, 115, 18, 43, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 98, 121, 116, 101, 115, 32, 97, 108, 108, 111, 99, 97, 116, 101, 100, 32, 97, 110, 100, 32, 115, 116, 105, 108, 108, 32, 105, 110, 32, 117, 115, 101, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 230, 140, 90, 65, 95, 10, 29, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 97, 108, 108, 111, 99, 95, 98, 121, 116, 101, 115, 95, 116, 111, 116, 97, 108, 18, 47, 84, 111, 116, 97, 108, 32, 110, 117, 109, 98, 101, 114, 32, 111, 102, 32, 98, 121, 116, 101, 115, 32, 97, 108, 108, 111, 99, 97, 116, 101, 100, 44, 32, 101, 118, 101, 110, 32, 105, 102, 32, 102, 114, 101, 101, 100, 46, 24, 0, 34, 11, 26, 9, 9, 0, 0, 0, 128, 172, 166, 125, 65, 106, 10, 31, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 98, 117, 99, 107, 95, 104, 97, 115, 104, 95, 115, 121, 115, 95, 98, 121, 116, 101, 115, 18, 56, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 98, 121, 116, 101, 115, 32, 117, 115, 101, 100, 32, 98, 121, 32, 116, 104, 101, 32, 112, 114, 111, 102, 105, 108, 105, 110, 103, 32, 98, 117, 99, 107, 101, 116, 32, 104, 97, 115, 104, 32, 116, 97, 98, 108, 101, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 167, 49, 54, 65, 64, 10, 23, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 102, 114, 101, 101, 115, 95, 116, 111, 116, 97, 108, 18, 22, 84, 111, 116, 97, 108, 32, 110, 117, 109, 98, 101, 114, 32, 111, 102, 32, 102, 114, 101, 101, 115, 46, 24, 0, 34, 11, 26, 9, 9, 0, 0, 0, 0, 144, 54, 10, 65, 137, 1, 10, 27, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 103, 99, 95, 99, 112, 117, 95, 102, 114, 97, 99, 116, 105, 111, 110, 18, 91, 84, 104, 101, 32, 102, 114, 97, 99, 116, 105, 111, 110, 32, 111, 102, 32, 116, 104, 105, 115, 32, 112, 114, 111, 103, 114, 97, 109, 39, 115, 32, 97, 118, 97, 105, 108, 97, 98, 108, 101, 32, 67, 80, 85, 32, 116, 105, 109, 101, 32, 117, 115, 101, 100, 32, 98, 121, 32, 116, 104, 101, 32, 71, 67, 32, 115, 105, 110, 99, 101, 32, 116, 104, 101, 32, 112, 114, 111, 103, 114, 97, 109, 32, 115, 116, 97, 114, 116, 101, 100, 46, 24, 1, 34, 11, 18, 9, 9, 25, 82, 195, 222, 99, 67, 223, 62, 103, 10, 24, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 103, 99, 95, 115, 121, 115, 95, 98, 121, 116, 101, 115, 18, 60, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 98, 121, 116, 101, 115, 32, 117, 115, 101, 100, 32, 102, 111, 114, 32, 103, 97, 114, 98, 97, 103, 101, 32, 99, 111, 108, 108, 101, 99, 116, 105, 111, 110, 32, 115, 121, 115, 116, 101, 109, 32, 109, 101, 116, 97, 100, 97, 116, 97, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 132, 84, 75, 65, 95, 10, 28, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 104, 101, 97, 112, 95, 97, 108, 108, 111, 99, 95, 98, 121, 116, 101, 115, 18, 48, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 104, 101, 97, 112, 32, 98, 121, 116, 101, 115, 32, 97, 108, 108, 111, 99, 97, 116, 101, 100, 32, 97, 110, 100, 32, 115, 116, 105, 108, 108, 32, 105, 110, 32, 117, 115, 101, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 230, 140, 90, 65, 86, 10, 27, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 104, 101, 97, 112, 95, 105, 100, 108, 101, 95, 98, 121, 116, 101, 115, 18, 40, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 104, 101, 97, 112, 32, 98, 121, 116, 101, 115, 32, 119, 97, 105, 116, 105, 110, 103, 32, 116, 111, 32, 98, 101, 32, 117, 115, 101, 100, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 224, 139, 65, 84, 10, 28, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 104, 101, 97, 112, 95, 105, 110, 117, 115, 101, 95, 98, 121, 116, 101, 115, 18, 37, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 104, 101, 97, 112, 32, 98, 121, 116, 101, 115, 32, 116, 104, 97, 116, 32, 97, 114, 101, 32, 105, 110, 32, 117, 115, 101, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 96, 95, 65, 71, 10, 24, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 104, 101, 97, 112, 95, 111, 98, 106, 101, 99, 116, 115, 18, 28, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 97, 108, 108, 111, 99, 97, 116, 101, 100, 32, 111, 98, 106, 101, 99, 116, 115, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 112, 216, 64, 86, 10, 31, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 104, 101, 97, 112, 95, 114, 101, 108, 101, 97, 115, 101, 100, 95, 98, 121, 116, 101, 115, 18, 36, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 104, 101, 97, 112, 32, 98, 121, 116, 101, 115, 32, 114, 101, 108, 101, 97, 115, 101, 100, 32, 116, 111, 32, 79, 83, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 140, 139, 65, 87, 10, 26, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 104, 101, 97, 112, 95, 115, 121, 115, 95, 98, 121, 116, 101, 115, 18, 42, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 104, 101, 97, 112, 32, 98, 121, 116, 101, 115, 32, 111, 98, 116, 97, 105, 110, 101, 100, 32, 102, 114, 111, 109, 32, 115, 121, 115, 116, 101, 109, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 204, 143, 65, 107, 10, 32, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 108, 97, 115, 116, 95, 103, 99, 95, 116, 105, 109, 101, 95, 115, 101, 99, 111, 110, 100, 115, 18, 56, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 115, 101, 99, 111, 110, 100, 115, 32, 115, 105, 110, 99, 101, 32, 49, 57, 55, 48, 32, 111, 102, 32, 108, 97, 115, 116, 32, 103, 97, 114, 98, 97, 103, 101, 32, 99, 111, 108, 108, 101, 99, 116, 105, 111, 110, 46, 24, 1, 34, 11, 18, 9, 9, 74, 50, 253, 75, 28, 247, 215, 65, 76, 10, 25, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 108, 111, 111, 107, 117, 112, 115, 95, 116, 111, 116, 97, 108, 18, 32, 84, 111, 116, 97, 108, 32, 110, 117, 109, 98, 101, 114, 32, 111, 102, 32, 112, 111, 105, 110, 116, 101, 114, 32, 108, 111, 111, 107, 117, 112, 115, 46, 24, 0, 34, 11, 26, 9, 9, 0, 0, 0, 0, 0, 0, 0, 0, 68, 10, 25, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 109, 97, 108, 108, 111, 99, 115, 95, 116, 111, 116, 97, 108, 18, 24, 84, 111, 116, 97, 108, 32, 110, 117, 109, 98, 101, 114, 32, 111, 102, 32, 109, 97, 108, 108, 111, 99, 115, 46, 24, 0, 34, 11, 26, 9, 9, 0, 0, 0, 0, 144, 68, 13, 65, 93, 10, 30, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 109, 99, 97, 99, 104, 101, 95, 105, 110, 117, 115, 101, 95, 98, 121, 116, 101, 115, 18, 44, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 98, 121, 116, 101, 115, 32, 105, 110, 32, 117, 115, 101, 32, 98, 121, 32, 109, 99, 97, 99, 104, 101, 32, 115, 116, 114, 117, 99, 116, 117, 114, 101, 115, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 32, 155, 64, 111, 10, 28, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 109, 99, 97, 99, 104, 101, 95, 115, 121, 115, 95, 98, 121, 116, 101, 115, 18, 64, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 98, 121, 116, 101, 115, 32, 117, 115, 101, 100, 32, 102, 111, 114, 32, 109, 99, 97, 99, 104, 101, 32, 115, 116, 114, 117, 99, 116, 117, 114, 101, 115, 32, 111, 98, 116, 97, 105, 110, 101, 100, 32, 102, 114, 111, 109, 32, 115, 121, 115, 116, 101, 109, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 0, 208, 64, 91, 10, 29, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 109, 115, 112, 97, 110, 95, 105, 110, 117, 115, 101, 95, 98, 121, 116, 101, 115, 18, 43, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 98, 121, 116, 101, 115, 32, 105, 110, 32, 117, 115, 101, 32, 98, 121, 32, 109, 115, 112, 97, 110, 32, 115, 116, 114, 117, 99, 116, 117, 114, 101, 115, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 17, 241, 64, 109, 10, 27, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 109, 115, 112, 97, 110, 95, 115, 121, 115, 95, 98, 121, 116, 101, 115, 18, 63, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 98, 121, 116, 101, 115, 32, 117, 115, 101, 100, 32, 102, 111, 114, 32, 109, 115, 112, 97, 110, 32, 115, 116, 114, 117, 99, 116, 117, 114, 101, 115, 32, 111, 98, 116, 97, 105, 110, 101, 100, 32, 102, 114, 111, 109, 32, 115, 121, 115, 116, 101, 109, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 0, 244, 64, 110, 10, 25, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 110, 101, 120, 116, 95, 103, 99, 95, 98, 121, 116, 101, 115, 18, 66, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 104, 101, 97, 112, 32, 98, 121, 116, 101, 115, 32, 119, 104, 101, 110, 32, 110, 101, 120, 116, 32, 103, 97, 114, 98, 97, 103, 101, 32, 99, 111, 108, 108, 101, 99, 116, 105, 111, 110, 32, 119, 105, 108, 108, 32, 116, 97, 107, 101, 32, 112, 108, 97, 99, 101, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 128, 249, 100, 65, 96, 10, 27, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 111, 116, 104, 101, 114, 95, 115, 121, 115, 95, 98, 121, 116, 101, 115, 18, 50, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 98, 121, 116, 101, 115, 32, 117, 115, 101, 100, 32, 102, 111, 114, 32, 111, 116, 104, 101, 114, 32, 115, 121, 115, 116, 101, 109, 32, 97, 108, 108, 111, 99, 97, 116, 105, 111, 110, 115, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 68, 153, 27, 65, 94, 10, 29, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 115, 116, 97, 99, 107, 95, 105, 110, 117, 115, 101, 95, 98, 121, 116, 101, 115, 18, 46, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 98, 121, 116, 101, 115, 32, 105, 110, 32, 117, 115, 101, 32, 98, 121, 32, 116, 104, 101, 32, 115, 116, 97, 99, 107, 32, 97, 108, 108, 111, 99, 97, 116, 111, 114, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 0, 26, 65, 103, 10, 27, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 115, 116, 97, 99, 107, 95, 115, 121, 115, 95, 98, 121, 116, 101, 115, 18, 57, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 98, 121, 116, 101, 115, 32, 111, 98, 116, 97, 105, 110, 101, 100, 32, 102, 114, 111, 109, 32, 115, 121, 115, 116, 101, 109, 32, 102, 111, 114, 32, 115, 116, 97, 99, 107, 32, 97, 108, 108, 111, 99, 97, 116, 111, 114, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 0, 26, 65, 77, 10, 21, 103, 111, 95, 109, 101, 109, 115, 116, 97, 116, 115, 95, 115, 121, 115, 95, 98, 121, 116, 101, 115, 18, 37, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 98, 121, 116, 101, 115, 32, 111, 98, 116, 97, 105, 110, 101, 100, 32, 102, 114, 111, 109, 32, 115, 121, 115, 116, 101, 109, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 4, 85, 145, 65, 58, 10, 10, 103, 111, 95, 116, 104, 114, 101, 97, 100, 115, 18, 29, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 79, 83, 32, 116, 104, 114, 101, 97, 100, 115, 32, 99, 114, 101, 97, 116, 101, 100, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 0, 28, 64, 92, 10, 25, 112, 114, 111, 99, 101, 115, 115, 95, 99, 112, 117, 95, 115, 101, 99, 111, 110, 100, 115, 95, 116, 111, 116, 97, 108, 18, 48, 84, 111, 116, 97, 108, 32, 117, 115, 101, 114, 32, 97, 110, 100, 32, 115, 121, 115, 116, 101, 109, 32, 67, 80, 85, 32, 116, 105, 109, 101, 32, 115, 112, 101, 110, 116, 32, 105, 110, 32, 115, 101, 99, 111, 110, 100, 115, 46, 24, 0, 34, 11, 26, 9, 9, 123, 20, 174, 71, 225, 122, 236, 63, 74, 10, 15, 112, 114, 111, 99, 101, 115, 115, 95, 109, 97, 120, 95, 102, 100, 115, 18, 40, 77, 97, 120, 105, 109, 117, 109, 32, 110, 117, 109, 98, 101, 114, 32, 111, 102, 32, 111, 112, 101, 110, 32, 102, 105, 108, 101, 32, 100, 101, 115, 99, 114, 105, 112, 116, 111, 114, 115, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 0, 48, 65, 67, 10, 16, 112, 114, 111, 99, 101, 115, 115, 95, 111, 112, 101, 110, 95, 102, 100, 115, 18, 32, 78, 117, 109, 98, 101, 114, 32, 111, 102, 32, 111, 112, 101, 110, 32, 102, 105, 108, 101, 32, 100, 101, 115, 99, 114, 105, 112, 116, 111, 114, 115, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 0, 34, 64, 78, 10, 29, 112, 114, 111, 99, 101, 115, 115, 95, 114, 101, 115, 105, 100, 101, 110, 116, 95, 109, 101, 109, 111, 114, 121, 95, 98, 121, 116, 101, 115, 18, 30, 82, 101, 115, 105, 100, 101, 110, 116, 32, 109, 101, 109, 111, 114, 121, 32, 115, 105, 122, 101, 32, 105, 110, 32, 98, 121, 116, 101, 115, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 110, 134, 65, 99, 10, 26, 112, 114, 111, 99, 101, 115, 115, 95, 115, 116, 97, 114, 116, 95, 116, 105, 109, 101, 95, 115, 101, 99, 111, 110, 100, 115, 18, 54, 83, 116, 97, 114, 116, 32, 116, 105, 109, 101, 32, 111, 102, 32, 116, 104, 101, 32, 112, 114, 111, 99, 101, 115, 115, 32, 115, 105, 110, 99, 101, 32, 117, 110, 105, 120, 32, 101, 112, 111, 99, 104, 32, 105, 110, 32, 115, 101, 99, 111, 110, 100, 115, 46, 24, 1, 34, 11, 18, 9, 9, 41, 92, 175, 167, 26, 247, 215, 65, 76, 10, 28, 112, 114, 111, 99, 101, 115, 115, 95, 118, 105, 114, 116, 117, 97, 108, 95, 109, 101, 109, 111, 114, 121, 95, 98, 121, 116, 101, 115, 18, 29, 86, 105, 114, 116, 117, 97, 108, 32, 109, 101, 109, 111, 114, 121, 32, 115, 105, 122, 101, 32, 105, 110, 32, 98, 121, 116, 101, 115, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 216, 137, 199, 65, 103, 10, 32, 112, 114, 111, 99, 101, 115, 115, 95, 118, 105, 114, 116, 117, 97, 108, 95, 109, 101, 109, 111, 114, 121, 95, 109, 97, 120, 95, 98, 121, 116, 101, 115, 18, 52, 77, 97, 120, 105, 109, 117, 109, 32, 97, 109, 111, 117, 110, 116, 32, 111, 102, 32, 118, 105, 114, 116, 117, 97, 108, 32, 109, 101, 109, 111, 114, 121, 32, 97, 118, 97, 105, 108, 97, 98, 108, 101, 32, 105, 110, 32, 98, 121, 116, 101, 115, 46, 24, 1, 34, 11, 18, 9, 9, 0, 0, 0, 0, 0, 0, 240, 191, 67, 10, 9, 115, 119, 97, 112, 95, 102, 114, 101, 101, 18, 25, 84, 101, 108, 101, 103, 114, 97, 102, 32, 99, 111, 108, 108, 101, 99, 116, 101, 100, 32, 109, 101, 116, 114, 105, 99, 24, 1, 34, 25, 10, 12, 10, 4, 104, 111, 115, 116, 18, 4, 111, 109, 115, 107, 18, 9, 9, 0, 0, 0, 0, 80, 67, 205, 65, 65, 10, 7, 115, 119, 97, 112, 95, 105, 110, 18, 25, 84, 101, 108, 101, 103, 114, 97, 102, 32, 99, 111, 108, 108, 101, 99, 116, 101, 100, 32, 109, 101, 116, 114, 105, 99, 24, 0, 34, 25, 10, 12, 10, 4, 104, 111, 115, 116, 18, 4, 111, 109, 115, 107, 26, 9, 9, 0, 0, 0, 0, 0, 144, 58, 65, 66, 10, 8, 115, 119, 97, 112, 95, 111, 117, 116, 18, 25, 84, 101, 108, 101, 103, 114, 97, 102, 32, 99, 111, 108, 108, 101, 99, 116, 101, 100, 32, 109, 101, 116, 114, 105, 99, 24, 0, 34, 25, 10, 12, 10, 4, 104, 111, 115, 116, 18, 4, 111, 109, 115, 107, 26, 9, 9, 0, 0, 0, 0, 0, 14, 101, 65, 68, 10, 10, 115, 119, 97, 112, 95, 116, 111, 116, 97, 108, 18, 25, 84, 101, 108, 101, 103, 114, 97, 102, 32, 99, 111, 108, 108, 101, 99, 116, 101, 100, 32, 109, 101, 116, 114, 105, 99, 24, 1, 34, 25, 10, 12, 10, 4, 104, 111, 115, 116, 18, 4, 111, 109, 115, 107, 18, 9, 9, 0, 0, 0, 0, 104, 153, 205, 65, 67, 10, 9, 115, 119, 97, 112, 95, 117, 115, 101, 100, 18, 25, 84, 101, 108, 101, 103, 114, 97, 102, 32, 99, 111, 108, 108, 101, 99, 116, 101, 100, 32, 109, 101, 116, 114, 105, 99, 24, 1, 34, 25, 10, 12, 10, 4, 104, 111, 115, 116, 18, 4, 111, 109, 115, 107, 18, 9, 9, 0, 0, 0, 0, 0, 134, 101, 65, 75, 10, 17, 115, 119, 97, 112, 95, 117, 115, 101, 100, 95, 112, 101, 114, 99, 101, 110, 116, 18, 25, 84, 101, 108, 101, 103, 114, 97, 102, 32, 99, 111, 108, 108, 101, 99, 116, 101, 100, 32, 109, 101, 116, 114, 105, 99, 24, 1, 34, 25, 10, 12, 10, 4, 104, 111, 115, 116, 18, 4, 111, 109, 115, 107, 18, 9, 9, 3, 176, 67, 208, 213, 45, 242, 63}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=delimited")
		w.Write(sampleProtoBufData)
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading body: %s", err)
	}
	parser := Parser{Header: resp.Header}
	_, err = parser.Parse(body)
	if err != nil {
		t.Fatalf("error reading metrics for %s: %s",
			ts.URL, err)
	}
}
