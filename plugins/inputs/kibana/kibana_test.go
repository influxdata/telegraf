package kibana

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func defaulttags63() map[string]string {
	return map[string]string{
		"name":    "my-kibana",
		"source":  "example.com:5601",
		"version": "6.3.2",
		"status":  "green",
	}
}

func defaulttags65() map[string]string {
	return map[string]string{
		"name":    "my-kibana",
		"source":  "example.com:5601",
		"version": "6.5.4",
		"status":  "green",
	}
}

func defaulttags815() map[string]string {
	return map[string]string{
		"name":    "my-kibana",
		"source":  "example.com:5601",
		"version": "8.15.2",
		"status":  "green", // available maps to green
	}
}

func defaulttags815Degraded() map[string]string {
	return map[string]string{
		"name":    "my-kibana",
		"source":  "example.com:5601",
		"version": "8.15.2",
		"status":  "yellow", // degraded maps to yellow
	}
}

func defaulttags815Unavailable() map[string]string {
	return map[string]string{
		"name":    "my-kibana",
		"source":  "example.com:5601",
		"version": "8.15.2",
		"status":  "red", // unavailable maps to red
	}
}

type transportMock struct {
	statusCode int
	body       string
}

func newTransportMock(statusCode int, body string) http.RoundTripper {
	return &transportMock{
		statusCode: statusCode,
		body:       body,
	}
}

func (t *transportMock) RoundTrip(r *http.Request) (*http.Response, error) {
	res := &http.Response{
		Header:     make(http.Header),
		Request:    r,
		StatusCode: t.statusCode,
	}
	res.Header.Set("Content-Type", "application/json")
	res.Body = io.NopCloser(strings.NewReader(t.body))
	return res, nil
}

func checkKibanaStatusResult(version, statusLevel string, t *testing.T, acc *testutil.Accumulator) {
	switch version {
	case "6.3.2":
		tags := defaulttags63()
		acc.AssertContainsTaggedFields(t, "kibana", kibanastatusexpected63, tags)
	case "6.5.4":
		tags := defaulttags65()
		acc.AssertContainsTaggedFields(t, "kibana", kibanastatusexpected65, tags)
	case "8.15.2":
		switch statusLevel {
		case "available":
			tags := defaulttags815()
			acc.AssertContainsTaggedFields(t, "kibana", kibanastatusexpected815, tags)
		case "degraded":
			tags := defaulttags815Degraded()
			acc.AssertContainsTaggedFields(t, "kibana", kibanastatusexpected815Degraded, tags)
		case "unavailable":
			tags := defaulttags815Unavailable()
			acc.AssertContainsTaggedFields(t, "kibana", kibanastatusexpected815Unavailable, tags)
		}
	}
}

func TestGather(t *testing.T) {
	ks := newKibanahWithClient()
	ks.Servers = []string{"http://example.com:5601"}

	// Unit test for Kibana version < 6.4
	ks.client.Transport = newTransportMock(http.StatusOK, kibanastatusresponse63)
	var acc1 testutil.Accumulator
	if err := acc1.GatherError(ks.Gather); err != nil {
		t.Fatal(err)
	}
	checkKibanaStatusResult(defaulttags63()["version"], "", t, &acc1)

	// Unit test for Kibana version >= 6.4
	ks.client.Transport = newTransportMock(http.StatusOK, kibanastatusresponse65)
	var acc2 testutil.Accumulator
	if err := acc2.GatherError(ks.Gather); err != nil {
		t.Fatal(err)
	}
	checkKibanaStatusResult(defaulttags65()["version"], "", t, &acc2)

	// Unit test for Kibana 8.x with "available" status
	ks.client.Transport = newTransportMock(http.StatusOK, kibanastatusresponse815)
	var acc3 testutil.Accumulator
	if err := acc3.GatherError(ks.Gather); err != nil {
		t.Fatal(err)
	}
	checkKibanaStatusResult(defaulttags815()["version"], "available", t, &acc3)

	// Unit test for Kibana 8.x with "degraded" status
	ks.client.Transport = newTransportMock(http.StatusOK, kibanastatusresponse815Degraded)
	var acc4 testutil.Accumulator
	if err := acc4.GatherError(ks.Gather); err != nil {
		t.Fatal(err)
	}
	checkKibanaStatusResult(defaulttags815()["version"], "degraded", t, &acc4)

	// Unit test for Kibana 8.x with "unavailable" status
	ks.client.Transport = newTransportMock(http.StatusOK, kibanastatusresponse815Unavailable)
	var acc5 testutil.Accumulator
	if err := acc5.GatherError(ks.Gather); err != nil {
		t.Fatal(err)
	}
	checkKibanaStatusResult(defaulttags815()["version"], "unavailable", t, &acc5)
}

// Test error handling with different HTTP status codes
func TestGatherErrors(t *testing.T) {
	ks := newKibanahWithClient()
	ks.Servers = []string{"http://example.com:5601"}

	// Test 500 Internal Server Error
	ks.client.Transport = newTransportMock(http.StatusInternalServerError, `{"error": "internal server error"}`)
	var acc1 testutil.Accumulator
	err := acc1.GatherError(ks.Gather)
	if err == nil {
		t.Fatal("Expected error for 500 status code")
	}

	// Test 404 Not Found
	ks.client.Transport = newTransportMock(http.StatusNotFound, `{"error": "not found"}`)
	var acc2 testutil.Accumulator
	err = acc2.GatherError(ks.Gather)
	if err == nil {
		t.Fatal("Expected error for 404 status code")
	}

	// Test 401 Unauthorized
	ks.client.Transport = newTransportMock(http.StatusUnauthorized, `{"error": "unauthorized"}`)
	var acc3 testutil.Accumulator
	err = acc3.GatherError(ks.Gather)
	if err == nil {
		t.Fatal("Expected error for 401 status code")
	}

	// Test 503 Service Unavailable
	ks.client.Transport = newTransportMock(http.StatusServiceUnavailable, `{"error": "service unavailable"}`)
	var acc4 testutil.Accumulator
	err = acc4.GatherError(ks.Gather)
	if err == nil {
		t.Fatal("Expected error for 503 status code")
	}
}

// Test invalid JSON response
func TestGatherInvalidJSON(t *testing.T) {
	ks := newKibanahWithClient()
	ks.Servers = []string{"http://example.com:5601"}

	// Test invalid JSON
	ks.client.Transport = newTransportMock(http.StatusOK, `{invalid json}`)
	var acc testutil.Accumulator
	err := acc.GatherError(ks.Gather)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}

// Test status mapping functions specifically
func TestStatusMapping(t *testing.T) {
	// Test legacy status mapping (Kibana 7.x and earlier)
	testCases := []struct {
		input    string
		expected int
	}{
		{"green", 1},
		{"yellow", 2},
		{"red", 3},
		{"unknown", 0},
		{"", 0},
	}

	for _, tc := range testCases {
		result := mapHealthStatusToCode(tc.input)
		if result != tc.expected {
			t.Errorf("mapHealthStatusToCode(%q) = %d, expected %d", tc.input, result, tc.expected)
		}
	}
}

// Test Kibana 8.x status level mapping
func TestKibana8xStatusMapping(t *testing.T) {
	testCases := []struct {
		level    string
		expected string
	}{
		{"available", "green"},
		{"degraded", "yellow"},
		{"unavailable", "red"},
		{"critical", "red"},
		{"unknown", "unknown"},
		{"", "unknown"},
	}

	for _, tc := range testCases {
		result := mapKibana8xStatus(tc.level)
		if result != tc.expected {
			t.Errorf("mapKibana8xStatus(%q) = %q, expected %q", tc.level, result, tc.expected)
		}
	}
}

// Test getStatusValue function with both legacy and new formats
func TestGetStatusValue(t *testing.T) {
	testCases := []struct {
		name     string
		overall  overallStatus
		expected string
	}{
		{
			name:     "Legacy green state",
			overall:  overallStatus{State: "green"},
			expected: "green",
		},
		{
			name:     "Legacy yellow state",
			overall:  overallStatus{State: "yellow"},
			expected: "yellow",
		},
		{
			name:     "Kibana 8.x available level",
			overall:  overallStatus{Level: "available"},
			expected: "green",
		},
		{
			name:     "Kibana 8.x degraded level",
			overall:  overallStatus{Level: "degraded"},
			expected: "yellow",
		},
		{
			name:     "Kibana 8.x unavailable level",
			overall:  overallStatus{Level: "unavailable"},
			expected: "red",
		},
		{
			name:     "Kibana 8.x critical level",
			overall:  overallStatus{Level: "critical"},
			expected: "red",
		},
		{
			name:     "Both fields present, level takes precedence",
			overall:  overallStatus{State: "green", Level: "degraded"},
			expected: "yellow",
		},
		{
			name:     "No fields present",
			overall:  overallStatus{},
			expected: "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getStatusValue(tc.overall)
			if result != tc.expected {
				t.Errorf("getStatusValue(%+v) = %q, expected %q", tc.overall, result, tc.expected)
			}
		})
	}
}

func newKibanahWithClient() *Kibana {
	ks := newKibana()
	ks.client = &http.Client{}
	return ks
}
