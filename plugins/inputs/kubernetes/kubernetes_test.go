package kubernetes

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
`

const invalidData = "I don't think this is valid data"

const empty = ""

type mockHTTPClient struct {
	responseBody string
	statusCode   int
}

// Mock implementation of MakeRequest. Usually returns an http.Response with
// hard-coded responseBody and statusCode. However, if the request uses a
// nonstandard method, it uses status code 405 (method not allowed)
func (c mockHTTPClient) MakeRequest(req *http.Request, timeout float64) (*http.Response, error) {
	resp := http.Response{}
	resp.StatusCode = c.statusCode

	// basic error checking on request method
	allowedMethods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "CONNECT"}
	methodValid := false
	for _, method := range allowedMethods {
		if req.Method == method {
			methodValid = true
			break
		}
	}

	if !methodValid {
		resp.StatusCode = 405 // Method not allowed
	}

	resp.Body = ioutil.NopCloser(strings.NewReader(c.responseBody))
	return &resp, nil
}

// Generates a pointer to an Kubernetes object that uses a mock HTTP client.
// Parameters:
//     response  : Body of the response that the mock HTTP client should return
//     statusCode: HTTP status code the mock HTTP client should return
//
// Returns:
//     *Kubernetes: Pointer to an Kubernetes object that uses the generated mock HTTP client
func genMockKubernetes(response string, statusCode int) Kubernetes {
	return Kubernetes{
		client: mockHTTPClient{responseBody: response, statusCode: statusCode},
		Apiserver: []Apiserver{
			Apiserver{
				KubeService{
					Url:      "http://127.0.0.1:8080",
					Endpoint: "/metrics",
					Timeout:  1.0,
				},
			},
		},
		Scheduler: []Scheduler{
			Scheduler{
				KubeService{
					Url:      "http://127.0.0.1:10251",
					Excludes: []string{"http_request_duration_.*"},
				},
			},
		},
		Controllermanager: []Controllermanager{
			Controllermanager{
				KubeService{
					Url:      "http://127.0.0.1:10252",
					Endpoint: "metrics",
					Includes: []string{"http_request_duration_microseconds"},
				},
			},
		},
		Kubelet: []Kubelet{
			Kubelet{
				KubeService{
					Url:      "http://127.0.0.1:4194",
					Endpoint: "metrics",
					Includes: []string{"http_request_duration_microseconds"},
				},
			},
		},
	}
}

// Generates a pointer to an Kubernetes object that uses a mock HTTP client.
// Parameters:
//     response  : Body of the response that the mock HTTP client should return
//     statusCode: HTTP status code the mock HTTP client should return
//
// Returns:
//     *Kubernetes: Pointer to an Kubernetes object that uses the generated mock HTTP client
func genMockKubernetes2(response string, statusCode int) Kubernetes {
	return Kubernetes{
		client: mockHTTPClient{responseBody: response, statusCode: statusCode},
		Apiserver: []Apiserver{
			Apiserver{
				KubeService{
					Url:      "http://127.0.0.1:8080",
					Endpoint: "/metrics",
					Timeout:  1.0,
					Includes: []string{"http_request_duration_microseconds"},
				},
			},
		},
	}
}

// Test that the proper values are ignored or collected
func TestOK(t *testing.T) {
	kubernetes := genMockKubernetes(validData, 200)

	var acc testutil.Accumulator
	err := kubernetes.Gather(&acc)
	require.NoError(t, err)
	assert.Equal(t, 33, acc.NFields())

}

// Test that the proper values are ignored or collected
func TestOK2(t *testing.T) {
	kubernetes := genMockKubernetes2(validData, 200)

	var acc testutil.Accumulator
	err := kubernetes.Gather(&acc)
	require.NoError(t, err)
	assert.Equal(t, 5, acc.NFields())

	tags := map[string]string{
		"kubeservice": "apiserver",
		"serverURL":   "http://127.0.0.1:8080/metrics",
		"handler":     "prometheus",
	}
	mname := "http_request_duration_microseconds"
	expectedFields := map[string]interface{}{
		"0.5":   552048.506,
		"0.9":   5.876804288e+06,
		"0.99":  5.876804288e+06,
		"count": 0.0,
		"sum":   1.8909097205e+07}
	acc.AssertContainsTaggedFields(t, mname, expectedFields, tags)

}

// Test response to HTTP 500
func TestKubernetes500(t *testing.T) {
	kubernetes := genMockKubernetes(validData, 500)

	var acc testutil.Accumulator
	err := kubernetes.Gather(&acc)

	assert.NotNil(t, err)
	assert.Equal(t, 0, acc.NFields())
}

// Test response to malformed Data
func TestKubernetesBadData(t *testing.T) {
	kubernetes := genMockKubernetes(invalidData, 200)

	var acc testutil.Accumulator
	err := kubernetes.Gather(&acc)

	assert.NotNil(t, err)
	assert.Equal(t, 0, acc.NFields())
}

// Test response to empty string as response objectgT
func TestKubernetesEmptyResponse(t *testing.T) {
	kubernetes := Kubernetes{client: RealHTTPClient{client: &http.Client{}}}
	kubernetes.Apiserver = []Apiserver{
		Apiserver{
			KubeService{
				Url:      "http://127.0.0.1:59999",
				Endpoint: "/metrics",
				Timeout:  1.0,
			},
		},
	}

	var acc testutil.Accumulator
	err := kubernetes.Gather(&acc)

	require.NotNil(t, err)
	assert.Equal(t, 0, acc.NFields())
}
