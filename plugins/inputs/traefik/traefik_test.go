package traefik

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pki = testutil.NewPKI("../../../testutil/pki")

func TestInvalidAddress(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	traefik := &Traefik{Address: "http://not_valid_aaaaa_for____anything_sksjdkaqiouer:9999"}
	var acc testutil.Accumulator
	require.Error(t, traefik.Gather(&acc))
	acc.AssertDoesNotContainMeasurement(t, "traefik")
}

func TestGatherPrimaryHealthCheck(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	traefik := &Traefik{
		Address: server.URL,
		IncludeStatusCodeMeasurement: false,
	}
	var acc testutil.Accumulator
	require.NoError(t, traefik.Gather(&acc))

	expectedFields := make(map[string]interface{})
	for k, v := range standardTraefikExpectedFields {
		expectedFields[k] = v
	}
	expectedFields["health_response_time_sec"] = traefik.lastRequestTiming

	acc.AssertContainsFields(t, "traefik", expectedFields)
	acc.AssertDoesNotContainMeasurement(t, "traefik_status_codes")
}

func TestGatherStatusCodes(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	traefik := &Traefik{
		Address: server.URL,
		IncludeStatusCodeMeasurement: true,
	}
	var acc testutil.Accumulator
	require.NoError(t, traefik.Gather(&acc))

	expectedFields := copyFields(standardTraefikExpectedFields)
	expected200Fields := copyFields(statusCode200ExpectedFields)
	expected404Fields := copyFields(statusCode404ExpectedFields)
	expectedFields["health_response_time_sec"] = traefik.lastRequestTiming
	expected200Fields["health_response_time_sec"] = traefik.lastRequestTiming
	expected404Fields["health_response_time_sec"] = traefik.lastRequestTiming

	acc.AssertContainsFields(t, "traefik", expectedFields)
	assert.False(t, acc.HasField("traefik_status_codes", "status_code_200"), "should not have status_code_* fields")
	assert.False(t, acc.HasField("traefik_status_codes", "status_code_400"), "should not have status_code_* fields")
	acc.AssertContainsTaggedFields(t, "traefik_status_codes",
		expected200Fields,
		map[string]string{"status_code": "200", "server": server.URL})
	acc.AssertContainsTaggedFields(t, "traefik_status_codes",
		expected404Fields,
		map[string]string{"status_code": "404", "server": server.URL})
}

func TestHTTPSGatherHealthCheckIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	address := fmt.Sprintf("https://%v:%v", testutil.GetLocalHost(), 8443)
	traefik := &Traefik{
		Address:      address,
		ClientConfig: *pki.TLSClientConfig(),
	}
	traefik.ClientConfig.InsecureSkipVerify = true

	var acc testutil.Accumulator
	require.NoError(t, traefik.Gather(&acc))

	assert.True(t, acc.HasMeasurement("traefik"), "expecting measurement traefik to be present")

	assert.True(t, acc.HasField("traefik", "total_response_time_sec"), "expecting field: total_response_time_sec")
	assert.True(t, acc.HasField("traefik", "total_count"), "expecting field: total_count")
	assert.True(t, acc.HasField("traefik", "average_response_time_sec"), "expecting field: average_response_time_sec")
	assert.True(t, acc.HasField("traefik", "unixtime"), "expecting field: unixtime")
	assert.True(t, acc.HasField("traefik", "uptime_sec"), "expecting field: uptime_sec")
	assert.True(t, acc.HasField("traefik", "health_response_time_sec"), "expecting field: health_response_time_sec")
}

func TestGatherHealthCheckIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	address := fmt.Sprintf("http://%v:%v", testutil.GetLocalHost(), 8080)
	traefik := &Traefik{Address: address}
	var acc testutil.Accumulator
	require.NoError(t, traefik.Gather(&acc))

	assert.True(t, acc.HasMeasurement("traefik"), "expecting measurement traefik to be present")

	assert.True(t, acc.HasField("traefik", "total_response_time_sec"), "expecting field: total_response_time_sec")
	assert.True(t, acc.HasField("traefik", "total_count"), "expecting field: total_count")
	assert.True(t, acc.HasField("traefik", "average_response_time_sec"), "expecting field: average_response_time_sec")
	assert.True(t, acc.HasField("traefik", "unixtime"), "expecting field: unixtime")
	assert.True(t, acc.HasField("traefik", "uptime_sec"), "expecting field: uptime_sec")
	assert.True(t, acc.HasField("traefik", "health_response_time_sec"), "expecting field: health_response_time_sec")
}

func createMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/health") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, standardTraefikHttpHealthResponse)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "nope")
		}
	}))
}

var statusCode200ExpectedFields = map[string]interface{}{
	"total_count": int(13),
	"uptime_sec":  float64(113.450952875),
	"unixtime":    int64(1492162320),
	"count":       int(7),
}
var statusCode404ExpectedFields = map[string]interface{}{
	"total_count": int(13),
	"uptime_sec":  float64(113.450952875),
	"unixtime":    int64(1492162320),
	"count":       int(6),
}
var standardTraefikExpectedFields = map[string]interface{}{
	"total_response_time_sec":   float64(0.015202713),
	"average_response_time_sec": float64(0.001169439),
	"total_count":               int(13),
	"status_code_200":           int(7),
	"status_code_404":           int(6),
	"uptime_sec":                float64(113.450952875),
	"unixtime":                  int64(1492162320),
}

const standardTraefikHttpHealthResponse = `{
    "pid": 1,
    "uptime": "1m53.450952875s",
    "uptime_sec": 113.450952875,
    "time": "2017-04-14 09:32:00.350042707 +0000 UTC",
    "unixtime": 1492162320,
    "status_code_count": {},
    "total_status_code_count": {
      "200": 7,
      "404": 6
    },
    "count": 0,
    "total_count": 13,
    "total_response_time": "15.202713ms",
    "total_response_time_sec": 0.015202713,
    "average_response_time": "1.169439ms",
    "average_response_time_sec": 0.001169439
  }`
