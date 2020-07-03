package prometheus

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var exptime = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

const validUniqueGauge = `# HELP cadvisor_version_info A metric with a constant '1' value labeled by kernel version, OS version, docker version, cadvisor version & cadvisor revision.
# TYPE cadvisor_version_info gauge
cadvisor_version_info{cadvisorRevision="",cadvisorVersion="",dockerVersion="1.8.2",kernelVersion="3.10.0-229.20.1.el7.x86_64",osVersion="CentOS Linux 7 (Core)"} 1
`

const validUniqueCounter = `# HELP get_token_fail_count Counter of failed Token() requests to the alternate token source
# TYPE get_token_fail_count counter
get_token_fail_count 0
`

const validUniqueLine = `# HELP get_token_fail_count Counter of failed Token() requests to the alternate token source
`

const validUniqueSummary = `# HELP http_request_duration_microseconds The HTTP request latencies in microseconds.
# TYPE http_request_duration_microseconds summary
http_request_duration_microseconds{handler="prometheus",quantile="0.5"} 552048.506
http_request_duration_microseconds{handler="prometheus",quantile="0.9"} 5.876804288e+06
http_request_duration_microseconds{handler="prometheus",quantile="0.99"} 5.876804288e+06
http_request_duration_microseconds_sum{handler="prometheus"} 1.8909097205e+07
http_request_duration_microseconds_count{handler="prometheus"} 9
`

const validUniqueHistogram = `# HELP apiserver_request_latencies Response latency distribution in microseconds for each verb, resource and client.
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

const validData = `# HELP cadvisor_version_info A metric with a constant '1' value labeled by kernel version, OS version, docker version, cadvisor version & cadvisor revision.
# TYPE cadvisor_version_info gauge
cadvisor_version_info{cadvisorRevision="",cadvisorVersion="",dockerVersion="1.8.2",kernelVersion="3.10.0-229.20.1.el7.x86_64",osVersion="CentOS Linux 7 (Core)"} 1
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0.013534896000000001
go_gc_duration_seconds{quantile="0.25"} 0.02469263
go_gc_duration_seconds{quantile="0.5"} 0.033727822000000005
go_gc_duration_seconds{quantile="0.75"} 0.03840335
go_gc_duration_seconds{quantile="1"} 0.049956604
go_gc_duration_seconds_sum 1970.341293002
go_gc_duration_seconds_count 65952
# HELP http_request_duration_microseconds The HTTP request latencies in microseconds.
# TYPE http_request_duration_microseconds summary
http_request_duration_microseconds{handler="prometheus",quantile="0.5"} 552048.506
http_request_duration_microseconds{handler="prometheus",quantile="0.9"} 5.876804288e+06
http_request_duration_microseconds{handler="prometheus",quantile="0.99"} 5.876804288e+06
http_request_duration_microseconds_sum{handler="prometheus"} 1.8909097205e+07
http_request_duration_microseconds_count{handler="prometheus"} 9
# HELP get_token_fail_count Counter of failed Token() requests to the alternate token source
# TYPE get_token_fail_count counter
get_token_fail_count 0
# HELP apiserver_request_latencies Response latency distribution in microseconds for each verb, resource and client.
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

const prometheusMulti = `
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
`

const prometheusMultiSomeInvalid = `
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu3, host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu4 , usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
`

func TestParseValidPrometheus(t *testing.T) {
	// Gauge value
	metrics, err := Parse([]byte(validUniqueGauge), http.Header{})
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

	// Counter value
	metrics, err = Parse([]byte(validUniqueCounter), http.Header{})
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "get_token_fail_count", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"counter": float64(0),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())

	// Summary data
	//SetDefaultTags(map[string]string{})
	metrics, err = Parse([]byte(validUniqueSummary), http.Header{})
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

	// histogram data
	metrics, err = Parse([]byte(validUniqueHistogram), http.Header{})
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
