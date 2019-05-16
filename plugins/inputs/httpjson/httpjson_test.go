package httpjson

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

const validJSON2 = `{
  "user":{
    "hash_rate":0,
    "expected_24h_rewards":0,
    "total_rewards":0.000595109232,
    "paid_rewards":0,
    "unpaid_rewards":0.000595109232,
    "past_24h_rewards":0,
    "total_work":"5172625408",
    "blocks_found":0
  },
  "workers":{
    "brminer.1":{
      "hash_rate":0,
      "hash_rate_24h":0,
      "valid_shares":"6176",
      "stale_shares":"0",
      "invalid_shares":"0",
      "rewards":4.5506464e-5,
      "rewards_24h":0,
      "reset_time":1455409950
    },
    "brminer.2":{
      "hash_rate":0,
      "hash_rate_24h":0,
      "valid_shares":"0",
      "stale_shares":"0",
      "invalid_shares":"0",
      "rewards":0,
      "rewards_24h":0,
      "reset_time":1455936726
    },
    "brminer.3":{
      "hash_rate":0,
      "hash_rate_24h":0,
      "valid_shares":"0",
      "stale_shares":"0",
      "invalid_shares":"0",
      "rewards":0,
      "rewards_24h":0,
      "reset_time":1455936733
    }
  },
  "pool":{
    "hash_rate":114100000,
    "active_users":843,
    "total_work":"5015346808842682368",
    "pps_ratio":1.04,
    "pps_rate":7.655e-9
  },
  "network":{
    "hash_rate":1426117703,
    "block_number":944895,
    "time_per_block":156,
    "difficulty":51825.72835216,
    "next_difficulty":51916.15249019,
    "retarget_time":95053
  },
  "market":{
    "ltc_btc":0.00798,
    "ltc_usd":3.37801,
    "ltc_eur":3.113,
    "ltc_gbp":2.32807,
    "ltc_rub":241.796,
    "ltc_cny":21.3883,
    "btc_usd":422.852
  }
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
func genMockHttpJson(response string, statusCode int) []*HttpJson {
	return []*HttpJson{
		{
			client: &mockHTTPClient{responseBody: response, statusCode: statusCode},
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
		{
			client: &mockHTTPClient{responseBody: response, statusCode: statusCode},
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
		err := acc.GatherError(service.Gather)
		require.NoError(t, err)
		assert.Equal(t, 12, acc.NFields())
		// Set responsetime
		for _, p := range acc.Metrics {
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

// Test that GET Parameters from the url string are applied properly
func TestHttpJsonGET_URL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.FormValue("api_key")
		assert.Equal(t, "mykey", key)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, validJSON2)
	}))
	defer ts.Close()

	a := HttpJson{
		Servers: []string{ts.URL + "?api_key=mykey"},
		Name:    "",
		Method:  "GET",
		client:  &RealHTTPClient{client: &http.Client{}},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.NoError(t, err)

	// remove response_time from gathered fields because it's non-deterministic
	delete(acc.Metrics[0].Fields, "response_time")

	fields := map[string]interface{}{
		"market_btc_usd":                  float64(422.852),
		"market_ltc_btc":                  float64(0.00798),
		"market_ltc_cny":                  float64(21.3883),
		"market_ltc_eur":                  float64(3.113),
		"market_ltc_gbp":                  float64(2.32807),
		"market_ltc_rub":                  float64(241.796),
		"market_ltc_usd":                  float64(3.37801),
		"network_block_number":            float64(944895),
		"network_difficulty":              float64(51825.72835216),
		"network_hash_rate":               float64(1.426117703e+09),
		"network_next_difficulty":         float64(51916.15249019),
		"network_retarget_time":           float64(95053),
		"network_time_per_block":          float64(156),
		"pool_active_users":               float64(843),
		"pool_hash_rate":                  float64(1.141e+08),
		"pool_pps_rate":                   float64(7.655e-09),
		"pool_pps_ratio":                  float64(1.04),
		"user_blocks_found":               float64(0),
		"user_expected_24h_rewards":       float64(0),
		"user_hash_rate":                  float64(0),
		"user_paid_rewards":               float64(0),
		"user_past_24h_rewards":           float64(0),
		"user_total_rewards":              float64(0.000595109232),
		"user_unpaid_rewards":             float64(0.000595109232),
		"workers_brminer.1_hash_rate":     float64(0),
		"workers_brminer.1_hash_rate_24h": float64(0),
		"workers_brminer.1_reset_time":    float64(1.45540995e+09),
		"workers_brminer.1_rewards":       float64(4.5506464e-05),
		"workers_brminer.1_rewards_24h":   float64(0),
		"workers_brminer.2_hash_rate":     float64(0),
		"workers_brminer.2_hash_rate_24h": float64(0),
		"workers_brminer.2_reset_time":    float64(1.455936726e+09),
		"workers_brminer.2_rewards":       float64(0),
		"workers_brminer.2_rewards_24h":   float64(0),
		"workers_brminer.3_hash_rate":     float64(0),
		"workers_brminer.3_hash_rate_24h": float64(0),
		"workers_brminer.3_reset_time":    float64(1.455936733e+09),
		"workers_brminer.3_rewards":       float64(0),
		"workers_brminer.3_rewards_24h":   float64(0),
	}

	acc.AssertContainsFields(t, "httpjson", fields)
}

// Test that GET Parameters are applied properly
func TestHttpJsonGET(t *testing.T) {
	params := map[string]string{
		"api_key": "mykey",
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.FormValue("api_key")
		assert.Equal(t, "mykey", key)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, validJSON2)
	}))
	defer ts.Close()

	a := HttpJson{
		Servers:    []string{ts.URL},
		Name:       "",
		Method:     "GET",
		Parameters: params,
		client:     &RealHTTPClient{client: &http.Client{}},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.NoError(t, err)

	// remove response_time from gathered fields because it's non-deterministic
	delete(acc.Metrics[0].Fields, "response_time")

	fields := map[string]interface{}{
		"market_btc_usd":                  float64(422.852),
		"market_ltc_btc":                  float64(0.00798),
		"market_ltc_cny":                  float64(21.3883),
		"market_ltc_eur":                  float64(3.113),
		"market_ltc_gbp":                  float64(2.32807),
		"market_ltc_rub":                  float64(241.796),
		"market_ltc_usd":                  float64(3.37801),
		"network_block_number":            float64(944895),
		"network_difficulty":              float64(51825.72835216),
		"network_hash_rate":               float64(1.426117703e+09),
		"network_next_difficulty":         float64(51916.15249019),
		"network_retarget_time":           float64(95053),
		"network_time_per_block":          float64(156),
		"pool_active_users":               float64(843),
		"pool_hash_rate":                  float64(1.141e+08),
		"pool_pps_rate":                   float64(7.655e-09),
		"pool_pps_ratio":                  float64(1.04),
		"user_blocks_found":               float64(0),
		"user_expected_24h_rewards":       float64(0),
		"user_hash_rate":                  float64(0),
		"user_paid_rewards":               float64(0),
		"user_past_24h_rewards":           float64(0),
		"user_total_rewards":              float64(0.000595109232),
		"user_unpaid_rewards":             float64(0.000595109232),
		"workers_brminer.1_hash_rate":     float64(0),
		"workers_brminer.1_hash_rate_24h": float64(0),
		"workers_brminer.1_reset_time":    float64(1.45540995e+09),
		"workers_brminer.1_rewards":       float64(4.5506464e-05),
		"workers_brminer.1_rewards_24h":   float64(0),
		"workers_brminer.2_hash_rate":     float64(0),
		"workers_brminer.2_hash_rate_24h": float64(0),
		"workers_brminer.2_reset_time":    float64(1.455936726e+09),
		"workers_brminer.2_rewards":       float64(0),
		"workers_brminer.2_rewards_24h":   float64(0),
		"workers_brminer.3_hash_rate":     float64(0),
		"workers_brminer.3_hash_rate_24h": float64(0),
		"workers_brminer.3_reset_time":    float64(1.455936733e+09),
		"workers_brminer.3_rewards":       float64(0),
		"workers_brminer.3_rewards_24h":   float64(0),
	}

	acc.AssertContainsFields(t, "httpjson", fields)
}

// Test that POST Parameters are applied properly
func TestHttpJsonPOST(t *testing.T) {
	params := map[string]string{
		"api_key": "mykey",
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, "api_key=mykey", string(body))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, validJSON2)
	}))
	defer ts.Close()

	a := HttpJson{
		Servers:    []string{ts.URL},
		Name:       "",
		Method:     "POST",
		Parameters: params,
		client:     &RealHTTPClient{client: &http.Client{}},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.NoError(t, err)

	// remove response_time from gathered fields because it's non-deterministic
	delete(acc.Metrics[0].Fields, "response_time")

	fields := map[string]interface{}{
		"market_btc_usd":                  float64(422.852),
		"market_ltc_btc":                  float64(0.00798),
		"market_ltc_cny":                  float64(21.3883),
		"market_ltc_eur":                  float64(3.113),
		"market_ltc_gbp":                  float64(2.32807),
		"market_ltc_rub":                  float64(241.796),
		"market_ltc_usd":                  float64(3.37801),
		"network_block_number":            float64(944895),
		"network_difficulty":              float64(51825.72835216),
		"network_hash_rate":               float64(1.426117703e+09),
		"network_next_difficulty":         float64(51916.15249019),
		"network_retarget_time":           float64(95053),
		"network_time_per_block":          float64(156),
		"pool_active_users":               float64(843),
		"pool_hash_rate":                  float64(1.141e+08),
		"pool_pps_rate":                   float64(7.655e-09),
		"pool_pps_ratio":                  float64(1.04),
		"user_blocks_found":               float64(0),
		"user_expected_24h_rewards":       float64(0),
		"user_hash_rate":                  float64(0),
		"user_paid_rewards":               float64(0),
		"user_past_24h_rewards":           float64(0),
		"user_total_rewards":              float64(0.000595109232),
		"user_unpaid_rewards":             float64(0.000595109232),
		"workers_brminer.1_hash_rate":     float64(0),
		"workers_brminer.1_hash_rate_24h": float64(0),
		"workers_brminer.1_reset_time":    float64(1.45540995e+09),
		"workers_brminer.1_rewards":       float64(4.5506464e-05),
		"workers_brminer.1_rewards_24h":   float64(0),
		"workers_brminer.2_hash_rate":     float64(0),
		"workers_brminer.2_hash_rate_24h": float64(0),
		"workers_brminer.2_reset_time":    float64(1.455936726e+09),
		"workers_brminer.2_rewards":       float64(0),
		"workers_brminer.2_rewards_24h":   float64(0),
		"workers_brminer.3_hash_rate":     float64(0),
		"workers_brminer.3_hash_rate_24h": float64(0),
		"workers_brminer.3_reset_time":    float64(1.455936733e+09),
		"workers_brminer.3_rewards":       float64(0),
		"workers_brminer.3_rewards_24h":   float64(0),
	}

	acc.AssertContainsFields(t, "httpjson", fields)
}

// Test response to HTTP 500
func TestHttpJson500(t *testing.T) {
	httpjson := genMockHttpJson(validJSON, 500)

	var acc testutil.Accumulator
	err := acc.GatherError(httpjson[0].Gather)

	assert.Error(t, err)
	assert.Equal(t, 0, acc.NFields())
}

// Test response to HTTP 405
func TestHttpJsonBadMethod(t *testing.T) {
	httpjson := genMockHttpJson(validJSON, 200)
	httpjson[0].Method = "NOT_A_REAL_METHOD"

	var acc testutil.Accumulator
	err := acc.GatherError(httpjson[0].Gather)

	assert.Error(t, err)
	assert.Equal(t, 0, acc.NFields())
}

// Test response to malformed JSON
func TestHttpJsonBadJson(t *testing.T) {
	httpjson := genMockHttpJson(invalidJSON, 200)

	var acc testutil.Accumulator
	err := acc.GatherError(httpjson[0].Gather)

	assert.Error(t, err)
	assert.Equal(t, 0, acc.NFields())
}

// Test response to empty string as response object
func TestHttpJsonEmptyResponse(t *testing.T) {
	httpjson := genMockHttpJson(empty, 200)

	var acc testutil.Accumulator
	err := acc.GatherError(httpjson[0].Gather)
	assert.NoError(t, err)
}

// Test that the proper values are ignored or collected
func TestHttpJson200Tags(t *testing.T) {
	httpjson := genMockHttpJson(validJSONTags, 200)

	for _, service := range httpjson {
		if service.Name == "other_webapp" {
			var acc testutil.Accumulator
			err := acc.GatherError(service.Gather)
			// Set responsetime
			for _, p := range acc.Metrics {
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

const validJSONArrayTags = `
[
	{
		"value": 15,
		"role": "master",
		"build": "123"
	},
	{
		"value": 17,
		"role": "slave",
		"build": "456"
	}
]`

// Test that array data is collected correctly
func TestHttpJsonArray200Tags(t *testing.T) {
	httpjson := genMockHttpJson(validJSONArrayTags, 200)

	for _, service := range httpjson {
		if service.Name == "other_webapp" {
			var acc testutil.Accumulator
			err := acc.GatherError(service.Gather)
			// Set responsetime
			for _, p := range acc.Metrics {
				p.Fields["response_time"] = 1.0
			}
			require.NoError(t, err)
			assert.Equal(t, 8, acc.NFields())
			assert.Equal(t, uint64(4), acc.NMetrics())

			for _, m := range acc.Metrics {
				if m.Tags["role"] == "master" {
					assert.Equal(t, "123", m.Tags["build"])
					assert.Equal(t, float64(15), m.Fields["value"])
					assert.Equal(t, float64(1), m.Fields["response_time"])
					assert.Equal(t, "httpjson_"+service.Name, m.Measurement)
				} else if m.Tags["role"] == "slave" {
					assert.Equal(t, "456", m.Tags["build"])
					assert.Equal(t, float64(17), m.Fields["value"])
					assert.Equal(t, float64(1), m.Fields["response_time"])
					assert.Equal(t, "httpjson_"+service.Name, m.Measurement)
				} else {
					assert.FailNow(t, "unknown metric")
				}
			}
		}
	}
}

var jsonBOM = []byte("\xef\xbb\xbf[{\"value\":17}]")

// TestHttpJsonBOM tests that UTF-8 JSON with a BOM can be parsed
func TestHttpJsonBOM(t *testing.T) {
	httpjson := genMockHttpJson(string(jsonBOM), 200)

	for _, service := range httpjson {
		if service.Name == "other_webapp" {
			var acc testutil.Accumulator
			err := acc.GatherError(service.Gather)
			require.NoError(t, err)
		}
	}
}
