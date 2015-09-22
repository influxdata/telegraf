package httpjson

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/koksan83/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validJSON = `
	{
		"parent": {
			"child": 3,
			"ignored_child": "hi"
		},
		"ignored_null": null,
		"integer": 4,
		"ignored_list": [3, 4],
		"ignored_parent": {
			"another_ignored_list": [4],
			"another_ignored_null": null,
			"ignored_string": "hello, world!"
		}
	}`

const invalidJSON = "I don't think this is JSON"

const empty = ""

type mockHTTPClient struct {
	responseBody string
	statusCode   int
}

// Mock implementation of MakeRequest. Usually returns an http.Response with
// hard-coded responseBody and statusCode. However, if the request uses a
// nonstandard method, it uses status code 405 (method not allowed)
func (c mockHTTPClient) MakeRequest(req *http.Request) (*http.Response, error) {
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

// Generates a pointer to an HttpJson object that uses a mock HTTP client.
// Parameters:
//     response  : Body of the response that the mock HTTP client should return
//     statusCode: HTTP status code the mock HTTP client should return
//
// Returns:
//     *HttpJson: Pointer to an HttpJson object that uses the generated mock HTTP client
func genMockHttpJson(response string, statusCode int) *HttpJson {
	return &HttpJson{
		client: mockHTTPClient{responseBody: response, statusCode: statusCode},
		Services: []Service{
			Service{
				Servers: []string{
					"http://server1.example.com/metrics/",
					"http://server2.example.com/metrics/",
				},
				Name:   "my_webapp",
				Method: "GET",
				Parameters: map[string]string{
					"httpParam1": "12",
					"httpParam2": "the second parameter",
				},
			},
			Service{
				Servers: []string{
					"http://server1.example.com/metrics/",
					"http://server2.example.com/metrics/",
				},
				Name:   "other_webapp",
				Method: "POST",
				Parameters: map[string]string{
					"httpParam1": "12",
					"httpParam2": "the second parameter",
				},
			},
		},
	}
}

// Test that the proper values are ignored or collected
func TestHttpJson200(t *testing.T) {
	httpjson := genMockHttpJson(validJSON, 200)

	var acc testutil.Accumulator
	err := httpjson.Gather(&acc)
	require.NoError(t, err)

	assert.Equal(t, 8, len(acc.Points))

	for _, service := range httpjson.Services {
		for _, srv := range service.Servers {
			require.NoError(t,
				acc.ValidateTaggedValue(
					fmt.Sprintf("%s_parent_child", service.Name),
					3.0,
					map[string]string{"server": srv},
				),
			)
			require.NoError(t,
				acc.ValidateTaggedValue(
					fmt.Sprintf("%s_integer", service.Name),
					4.0,
					map[string]string{"server": srv},
				),
			)
		}
	}
}

// Test response to HTTP 500
func TestHttpJson500(t *testing.T) {
	httpjson := genMockHttpJson(validJSON, 500)

	var acc testutil.Accumulator
	err := httpjson.Gather(&acc)

	assert.NotNil(t, err)
	// 4 error lines for (2 urls) * (2 services)
	assert.Equal(t, len(strings.Split(err.Error(), "\n")), 4)
	assert.Equal(t, 0, len(acc.Points))
}

// Test response to HTTP 405
func TestHttpJsonBadMethod(t *testing.T) {
	httpjson := genMockHttpJson(validJSON, 200)
	httpjson.Services[0].Method = "NOT_A_REAL_METHOD"

	var acc testutil.Accumulator
	err := httpjson.Gather(&acc)

	assert.NotNil(t, err)
	// 2 error lines for (2 urls) * (1 falied service)
	assert.Equal(t, len(strings.Split(err.Error(), "\n")), 2)

	// (2 measurements) * (2 servers) * (1 successful service)
	assert.Equal(t, 4, len(acc.Points))
}

// Test response to malformed JSON
func TestHttpJsonBadJson(t *testing.T) {
	httpjson := genMockHttpJson(invalidJSON, 200)

	var acc testutil.Accumulator
	err := httpjson.Gather(&acc)

	assert.NotNil(t, err)
	// 4 error lines for (2 urls) * (2 services)
	assert.Equal(t, len(strings.Split(err.Error(), "\n")), 4)
	assert.Equal(t, 0, len(acc.Points))
}

// Test response to empty string as response objectgT
func TestHttpJsonEmptyResponse(t *testing.T) {
	httpjson := genMockHttpJson(empty, 200)

	var acc testutil.Accumulator
	err := httpjson.Gather(&acc)

	assert.NotNil(t, err)
	// 4 error lines for (2 urls) * (2 services)
	assert.Equal(t, len(strings.Split(err.Error(), "\n")), 4)
	assert.Equal(t, 0, len(acc.Points))
}
