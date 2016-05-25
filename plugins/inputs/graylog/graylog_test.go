package graylog

import (
//	"fmt"
	"io/ioutil"
	"net/http"
//	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validJSON = `
  {
    "total": 2,
    "metrics": [
      {
        "full_name": "jvm.cl.loaded",
        "metric": {
          "value": 18910
        },
        "name": "loaded",
        "type": "gauge"
      },
      {
        "full_name": "jvm.memory.pools.Metaspace.committed",
        "metric": {
          "value": 108040192
        },
        "name": "committed",
        "type": "gauge"
      }
    ]
  }`


var expectedFields = map[string]interface{}{
	"jvm.memory.pools.Metaspace.committed":         float64(108040192),
	"jvm.cl.loaded":         float64(18910),
}

const invalidJSON = "I don't think this is JSON"

const empty = ""

type mockHTTPClient struct {
	responseBody string
	statusCode   int
}

// Mock implementation of MakeRequest. Usually returns an http.Response with
// hard-coded responseBody and statusCode. However, if the request uses a
// nonstandard method, it uses status code 405 (method not allowed)
func (c *mockHTTPClient) MakeRequest(req *http.Request) (*http.Response, error) {
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

func (c *mockHTTPClient) SetHTTPClient(_ *http.Client) {
}

func (c *mockHTTPClient) HTTPClient() *http.Client {
	return nil
}

// Generates a pointer to an HttpJson object that uses a mock HTTP client.
// Parameters:
//     response  : Body of the response that the mock HTTP client should return
//     statusCode: HTTP status code the mock HTTP client should return
//
// Returns:
//     *HttpJson: Pointer to an HttpJson object that uses the generated mock HTTP client
func genMockGrayLog(response string, statusCode int) []*GrayLog {
	return []*GrayLog{
		&GrayLog{
			client: &mockHTTPClient{responseBody: response, statusCode: statusCode},
			Servers: []string{
				"http://localhost:12900/system/metrics/multiple",
			},
			Name:   "my_webapp",
			Metrics: []string{
         "jvm.cl.loaded",
			},
			Headers: map[string]string{
        "Content-Type" : "application/json",
        "Accept" : "application/json",
        "Authorization" : "Basic DESfdsfffoffo",
			},
		},
		&GrayLog{
			client: &mockHTTPClient{responseBody: response, statusCode: statusCode},
      Servers: []string{
				"http://server2:12900/system/metrics/multiple",
			},
			Name:   "other_webapp",
      Metrics: []string{
         "jvm.memory.pools.Metaspace.committed",
			},
			Headers: map[string]string{
        "Content-Type" : "application/json",
        "Accept" : "application/json",
        "Authorization" : "Basic DESfdsfffoffo",
			},
			TagKeys: []string{
				"role",
				"build",
			},
		},
	}
}

// Test that the proper values are ignored or collected
func TestNormalResponse(t *testing.T) {
	graylog := genMockGrayLog(validJSON, 200)

	for _, service := range graylog {
		var acc testutil.Accumulator
		err := service.Gather(&acc)
		require.NoError(t, err)
		assert.Equal(t, 3, acc.NFields())
		// Set responsetime
		for _, p := range acc.Metrics {
			p.Fields["response_time"] = 1.0
		}

		for _, srv := range service.Servers {
			tags := map[string]string{"server": srv}
			mname := "graylog_" + service.Name
			expectedFields["response_time"] = 1.0
			acc.AssertContainsTaggedFields(t, mname, expectedFields, tags)
		}
	}
}


// Test response to HTTP 500
func TestHttpJson500(t *testing.T) {
	graylog := genMockGrayLog(validJSON, 500)

	var acc testutil.Accumulator
	err := graylog[0].Gather(&acc)

	assert.NotNil(t, err)
	assert.Equal(t, 0, acc.NFields())
}

// Test response to malformed JSON
func TestHttpJsonBadJson(t *testing.T) {
	graylog := genMockGrayLog(invalidJSON, 200)

	var acc testutil.Accumulator
	err := graylog[0].Gather(&acc)

	assert.NotNil(t, err)
	assert.Equal(t, 0, acc.NFields())
}

// Test response to empty string as response objectgT
func TestHttpJsonEmptyResponse(t *testing.T) {
	graylog := genMockGrayLog(empty, 200)

	var acc testutil.Accumulator
	err := graylog[0].Gather(&acc)

	assert.NotNil(t, err)
	assert.Equal(t, 0, acc.NFields())
}
