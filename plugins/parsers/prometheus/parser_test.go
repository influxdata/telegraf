package prometheus

import (
	"testing"

	"github.com/influxdata/telegraf"

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

func ParseValidGauge(t *testing.T) {
	metrics, err := parse([]byte(validUniqueGauge))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "cadvisor_version_info", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"gauge": float64(1),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"osVersion":        "CentOS Linux 7 (Core)",
		"cadvisorRevision": "",
		"cadvisorVersion":  "",
		"dockerVersion":    "1.8.2",
		"kernelVersion":    "3.10.0-229.20.1.el7.x86_64",
	}, metrics[0].Tags())
}

func ParseValieCounter(t *testing.T) {
	// Counter value
	metrics, err := parse([]byte(validUniqueCounter))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "get_token_fail_count", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"counter": float64(0),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}

func ParseValidSummary(t *testing.T) {
	// Summary data
	//SetDefaultTags(map[string]string{})
	metrics, err := parse([]byte(validUniqueSummary))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "http_request_duration_microseconds", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"0.5":   552048.506,
		"0.9":   5.876804288e+06,
		"0.99":  5.876804288e+06,
		"count": 9.0,
		"sum":   1.8909097205e+07,
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"handler": "prometheus"}, metrics[0].Tags())
}

func ParseValidHistogram(t *testing.T) {
	// histogram data
	metrics, err := parse([]byte(validUniqueHistogram))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "apiserver_request_latencies", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"500000": 2000.0,
		"count":  2025.0,
		"sum":    1.02726334e+08,
		"250000": 1997.0,
		"2e+06":  2012.0,
		"4e+06":  2017.0,
		"8e+06":  2024.0,
		"+Inf":   2025.0,
		"125000": 1994.0,
		"1e+06":  2005.0,
	}, metrics[0].Fields())
	assert.Equal(t,
		map[string]string{"verb": "POST", "resource": "bindings"},
		metrics[0].Tags())
}

func AddsDefautTags(t *testing.T) {
	parser := Parser{
		DefaultTags: map[string]string{
			"defaultTag": "defaultTagValue",
		},
	}
	metrics, _ := parser.Parse([]byte(validUniqueGauge))
	assert.Equal(t, map[string]string{
		"osVersion":        "CentOS Linux 7 (Core)",
		"cadvisorRevision": "",
		"cadvisorVersion":  "",
		"dockerVersion":    "1.8.2",
		"kernelVersion":    "3.10.0-229.20.1.el7.x86_64",
		"defaultTag":       "defaultTagValue",
	}, metrics[0].Tags())
}

func parse(buf []byte) ([]telegraf.Metric, error) {
	parser := Parser{}
	return parser.Parse(buf)
}
