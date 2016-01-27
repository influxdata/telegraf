package httpjson

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validJSON = `
	{
		"parent": {
			"child": 3.0,
			"ignored_child": "hi"
		},
		"ignored_null": null,
		"integer": 4,
		"list": [3, 4],
		"ignored_parent": {
			"another_ignored_null": null,
			"ignored_string": "hello, world!"
		},
		"another_list": [4]
	}`

const validJSONTags = `
	{
		"value": 15,
		"role": "master",
		"build": "123"
	}`

var expectedFields = map[string]interface{}{
	"parent_child":   float64(3),
	"list_0":         float64(3),
	"list_1":         float64(4),
	"another_list_0": float64(4),
	"integer":        float64(4),
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
func genMockHttpJson(response string, statusCode int) []*HttpJson {
	return []*HttpJson{
		&HttpJson{
			client: mockHTTPClient{responseBody: response, statusCode: statusCode},
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
			Headers: map[string]string{
				"X-Auth-Token": "the-first-parameter",
				"apiVersion":   "v1",
			},
		},
		&HttpJson{
			client: mockHTTPClient{responseBody: response, statusCode: statusCode},
			Servers: []string{
				"http://server3.example.com/metrics/",
				"http://server4.example.com/metrics/",
			},
			Name:   "other_webapp",
			Method: "POST",
			Parameters: map[string]string{
				"httpParam1": "12",
				"httpParam2": "the second parameter",
			},
			Headers: map[string]string{
				"X-Auth-Token": "the-first-parameter",
				"apiVersion":   "v1",
			},
			TagKeys: []string{
				"role",
				"build",
			},
		},
	}
}

// Test that the proper values are ignored or collected
func TestHttpJson200(t *testing.T) {
	httpjson := genMockHttpJson(validJSON, 200)

	for _, service := range httpjson {
		var acc testutil.Accumulator
		err := service.Gather(&acc)
		require.NoError(t, err)
		assert.Equal(t, 12, acc.NFields())
		// Set responsetime
		for _, p := range acc.Points {
			p.Fields["response_time"] = 1.0
		}

		for _, srv := range service.Servers {
			tags := map[string]string{"server": srv}
			mname := "httpjson_" + service.Name
			expectedFields["response_time"] = 1.0
			acc.AssertContainsTaggedFields(t, mname, expectedFields, tags)
		}
	}
}

// Test response to HTTP 500
func TestHttpJson500(t *testing.T) {
	httpjson := genMockHttpJson(validJSON, 500)

	var acc testutil.Accumulator
	err := httpjson[0].Gather(&acc)

	assert.NotNil(t, err)
	assert.Equal(t, 0, acc.NFields())
}

// Test response to HTTP 405
func TestHttpJsonBadMethod(t *testing.T) {
	httpjson := genMockHttpJson(validJSON, 200)
	httpjson[0].Method = "NOT_A_REAL_METHOD"

	var acc testutil.Accumulator
	err := httpjson[0].Gather(&acc)

	assert.NotNil(t, err)
	assert.Equal(t, 0, acc.NFields())
}

// Test response to malformed JSON
func TestHttpJsonBadJson(t *testing.T) {
	httpjson := genMockHttpJson(invalidJSON, 200)

	var acc testutil.Accumulator
	err := httpjson[0].Gather(&acc)

	assert.NotNil(t, err)
	assert.Equal(t, 0, acc.NFields())
}

// Test response to empty string as response objectgT
func TestHttpJsonEmptyResponse(t *testing.T) {
	httpjson := genMockHttpJson(empty, 200)

	var acc testutil.Accumulator
	err := httpjson[0].Gather(&acc)

	assert.NotNil(t, err)
	assert.Equal(t, 0, acc.NFields())
}

// Test that the proper values are ignored or collected
func TestHttpJson200Tags(t *testing.T) {
	httpjson := genMockHttpJson(validJSONTags, 200)

	for _, service := range httpjson {
		if service.Name == "other_webapp" {
			var acc testutil.Accumulator
			err := service.Gather(&acc)
			// Set responsetime
			for _, p := range acc.Points {
				p.Fields["response_time"] = 1.0
			}
			require.NoError(t, err)
			assert.Equal(t, 4, acc.NFields())
			for _, srv := range service.Servers {
				tags := map[string]string{"server": srv, "role": "master", "build": "123"}
				fields := map[string]interface{}{"value": float64(15), "response_time": float64(1)}
				mname := "httpjson_" + service.Name
				acc.AssertContainsTaggedFields(t, mname, fields, tags)
			}
		}
	}
}
