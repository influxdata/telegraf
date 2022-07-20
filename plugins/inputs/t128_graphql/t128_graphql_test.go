package t128_graphql_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	plugin "github.com/influxdata/telegraf/plugins/inputs/t128_graphql"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type Endpoint struct {
	URL             string
	Code            int
	ExpectedRequest string
	Response        string
}

const (
	ValidExpectedRequest                  = `{"query":"query {\nallRouters(name:\"ComboEast\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ntest-field\ntest-tag}}}}}}}"}`
	ValidExpectedRequestNoTag             = `{"query":"query {\nallRouters(name:\"ComboEast\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ntest-field}}}}}}}"}`
	ValidExpectedRequestWithAbsPaths      = `{"query":"query {\nallRouters(name:\"ComboEast\"){\nnodes{\nname\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ntest-field\ntest-tag}}\nname}}}}}"}`
	ValidExpectedRequestWithMixedResponse = `{"query":"query {\nallRouters(name:\"ComboEast\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\nrouter{\npeers(names:\"peer-1\"){\nnodes{\npaths{\nstatus\nuptime}}}}}}}}}"}`
	InvalidRouterExpectedRequest          = `{"query":"query {\nallRouters(name:\"not-a-router\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ntest-field\ntest-tag}}}}}}}"}`
	InvalidFieldExpectedRequest           = `{"query":"query {\nallRouters(name:\"ComboEast\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ninvalid-field\ntest-tag}}}}}}}"}`
)

var (
	ValidQuery                  = ValidExpectedRequest[10 : len(ValidExpectedRequest)-2]
	ValidQueryNoTag             = ValidExpectedRequestNoTag[10 : len(ValidExpectedRequestNoTag)-2]
	ValidQueryWithAbsPaths      = ValidExpectedRequestWithAbsPaths[10 : len(ValidExpectedRequestWithAbsPaths)-2]
	ValidQueryWithMixedResponse = ValidExpectedRequestWithMixedResponse[10 : len(ValidExpectedRequestWithMixedResponse)-2]
	InvalidRouterQuery          = InvalidRouterExpectedRequest[10 : len(InvalidRouterExpectedRequest)-2]
	InvalidFieldQuery           = InvalidFieldExpectedRequest[10 : len(InvalidFieldExpectedRequest)-2]
)

var CollectorTestCases = []struct {
	Name             string
	EntryPoint       string
	Fields           map[string]string
	Tags             map[string]string
	InitError        bool
	Query            string
	Endpoint         Endpoint
	ExpectedMetrics  []*testutil.Metric
	ExpectedErrors   []string
	RetryIfNotFound  bool
	ExpectedRequests []int
}{
	{
		Name:             "missing entry-point produces no request or metrics",
		EntryPoint:       "",
		Fields:           nil,
		Tags:             nil,
		InitError:        true,
		ExpectedRequests: []int{0},
	},
	{
		Name:             "missing extract-fields produces no request or metrics",
		EntryPoint:       "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:           nil,
		Tags:             nil,
		InitError:        true,
		ExpectedRequests: []int{0},
	},
	{
		Name:       "tag with graphQL argument in path produces no request or metrics",
		EntryPoint: "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:     map[string]string{"test-field": "test-field"},
		Tags: map[string]string{
			"test-tag":       "test-tag",
			"other-test-tag": "allRouters(name:'ComboEast')/nodes/name",
		},
		InitError:        true,
		ExpectedRequests: []int{0},
	},
	{
		Name:            "empty response produces error",
		EntryPoint:      "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:          map[string]string{"test-field": "test-field"},
		Tags:            map[string]string{"test-tag": "test-tag"},
		Query:           ValidQuery,
		Endpoint:        Endpoint{"/api/v1/graphql/", 200, ValidExpectedRequest, "{}"},
		ExpectedMetrics: nil,
		ExpectedErrors: []string{
			"no data found in response for collector test-collector",
		},
		ExpectedRequests: []int{1},
	},
	{
		Name:            "propogates not found error to accumulator",
		EntryPoint:      "allRouters(name:'not-a-router')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:          map[string]string{"test-field": "test-field"},
		Tags:            map[string]string{"test-tag": "test-tag"},
		Query:           InvalidRouterQuery,
		Endpoint:        Endpoint{"/api/v1/graphql/", 404, InvalidRouterExpectedRequest, `it's not right`},
		ExpectedMetrics: nil,
		ExpectedErrors: []string{
			"status code 404 not OK for collector test-collector: it's not right",
		},
		ExpectedRequests: []int{1},
	},
	{
		Name:             "propogates invalid json error to accumulator",
		EntryPoint:       "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:           map[string]string{"test-field": "test-field"},
		Tags:             map[string]string{"test-tag": "test-tag"},
		Query:            ValidQuery,
		Endpoint:         Endpoint{"/api/v1/graphql/", 200, ValidExpectedRequest, `{"test": }`},
		ExpectedMetrics:  nil,
		ExpectedErrors:   []string{"invalid json response for collector test-collector: invalid character '}' looking for beginning of value"},
		ExpectedRequests: []int{1},
	},
	{
		Name:       "propogates graphQL error to accumulator",
		EntryPoint: "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:     map[string]string{"invalid-field": "invalid-field"},
		Tags:       map[string]string{"test-tag": "test-tag"},
		Query:      InvalidFieldQuery,
		Endpoint: Endpoint{"/api/v1/graphql/", 200, InvalidFieldExpectedRequest, `
		{
			"errors": [{
				"name": "GraphQLError",
				"message": "Cannot query field \"invalid-field\" on type \"ArpEntryType\".",
				"locations": [{
					"line": 2,
					"column": 1
				}]
			}]
		  }`},
		ExpectedMetrics: nil,
		ExpectedErrors: []string{
			"found errors in response for collector test-collector: Cannot query field \"invalid-field\" on type \"ArpEntryType\".",
			"no data found in response for collector test-collector",
		},
		ExpectedRequests: []int{1},
	},
	{
		Name:       "retries if not found",
		EntryPoint: "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:     map[string]string{"test-field": "test-field"},
		Tags:       nil,
		Query:      ValidQueryNoTag,
		Endpoint: Endpoint{"/api/v1/graphql/", 200, ValidExpectedRequestNoTag, `
		{
			"errors": [{
				"name": "GraphQLError",
				"message": "highwayManager@CHSSDWCond01CHI.CHSSDWCondMD returned a 404"
			}]
		  }`},
		ExpectedMetrics: nil,
		ExpectedErrors: []string{
			"found errors in response for collector test-collector: highwayManager@CHSSDWCond01CHI.CHSSDWCondMD returned a 404",
			"no data found in response for collector test-collector",
			"found errors in response for collector test-collector: highwayManager@CHSSDWCond01CHI.CHSSDWCondMD returned a 404",
			"no data found in response for collector test-collector",
		},
		RetryIfNotFound:  true,
		ExpectedRequests: []int{1, 2},
	},
	{
		Name:       "doesn't retry if not found",
		EntryPoint: "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:     map[string]string{"test-field": "test-field"},
		Tags:       nil,
		Query:      ValidQueryNoTag,
		Endpoint: Endpoint{"/api/v1/graphql/", 200, ValidExpectedRequestNoTag, `
		{
			"errors": [{
				"name": "GraphQLError",
				"message": "highwayManager@CHSSDWCond01CHI.CHSSDWCondMD returned a 404"
			}]
		  }`},
		ExpectedMetrics: nil,
		ExpectedErrors: []string{
			"found errors in response for collector test-collector: highwayManager@CHSSDWCond01CHI.CHSSDWCondMD returned a 404",
			"no data found in response for collector test-collector",
			"collector configured to not retry when endpoint not found (404), stopping queries",
		},
		RetryIfNotFound:  false,
		ExpectedRequests: []int{1, 1},
	},
	{
		Name:       "missing extract-tags produces response",
		EntryPoint: "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:     map[string]string{"test-field": "test-field"},
		Tags:       nil,
		Query:      ValidQueryNoTag,
		Endpoint: Endpoint{"/api/v1/graphql/", 200, ValidExpectedRequestNoTag, `{
			"data": {
				"allRouters": {
				  	"nodes": [{
					  	"nodes": {
							"nodes": [{
								"arp": {
							  		"nodes": [{
								  		"test-field": 128
									}]
								}
						  	}]
					  	}
					}]
				}
			}
		}`},
		ExpectedMetrics: []*testutil.Metric{
			{
				Measurement: "test-collector",
				Tags:        map[string]string{},
				Fields:      map[string]interface{}{"test-field": 128.0},
			},
		},
		ExpectedErrors:   nil,
		ExpectedRequests: []int{1},
	},
	{
		Name:       "missing tags/fields with absolute path produces response", //complex processing tested separately
		EntryPoint: "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:     map[string]string{"test-field": "test-field"},
		Tags:       map[string]string{"test-tag": "test-tag"},
		Query:      ValidQuery,
		Endpoint: Endpoint{"/api/v1/graphql/", 200, ValidExpectedRequest, `{
			"data": {
				"allRouters": {
				  	"nodes": [{
					  	"nodes": {
							"nodes": [{
								"arp": {
							  		"nodes": [{
								  		"test-field": 128,
								  		"test-tag": "test-string-1"
									}]
								}
						  	}]
					  	}
					}]
				}
			}
		}`},
		ExpectedMetrics: []*testutil.Metric{
			{
				Measurement: "test-collector",
				Tags:        map[string]string{"test-tag": "test-string-1"},
				Fields:      map[string]interface{}{"test-field": 128.0},
			},
		},
		ExpectedErrors:   nil,
		ExpectedRequests: []int{1},
	},
	{
		Name:       "full config produces response", //complex processing tested separately
		EntryPoint: "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields: map[string]string{
			"test-field":       "test-field",
			"other-test-field": "allRouters/nodes/nodes/nodes/name",
		},
		Tags: map[string]string{
			"test-tag":       "test-tag",
			"other-test-tag": "allRouters/nodes/name",
		},
		Query: ValidQueryWithAbsPaths,
		Endpoint: Endpoint{"/api/v1/graphql/", 200, ValidExpectedRequestWithAbsPaths, `{
			"data": {
				"allRouters": {
				  	"nodes": [{
						"name": "ComboEast",
					  	"nodes": {
							"nodes": [{
								"name": "east-combo",
								"arp": {
							  		"nodes": [{
								  		"test-field": 128,
								  		"test-tag": "test-string-1"
									}]
								}
						  	}]
					  	}
					}]
				}
			}
		}`},
		ExpectedMetrics: []*testutil.Metric{
			{
				Measurement: "test-collector",
				Tags: map[string]string{
					"test-tag":       "test-string-1",
					"other-test-tag": "ComboEast",
				},
				Fields: map[string]interface{}{
					"test-field":       128.0,
					"other-test-field": "east-combo",
				},
			},
		},
		ExpectedErrors:   nil,
		ExpectedRequests: []int{1},
	},
	{
		Name:       "mixed produces errors and response",
		EntryPoint: "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/router/peers(names:'peer-1')/nodes",
		Fields: map[string]string{
			"status": "paths/status",
		},
		Tags: map[string]string{
			"uptime": "paths/uptime",
		},
		Query: ValidQueryWithMixedResponse,
		Endpoint: Endpoint{"/api/v1/graphql/", 200, ValidExpectedRequestWithMixedResponse, `{
			"errors": [
			  {
				"name": "TypeError",
				"message": "Int cannot represent non 32-bit signed integer value: 3066521082",
				"locations": [
				  {
					"line": 10,
					"column": 19
				  }
				],
				"path": [
				  "allRouters",
				  "nodes",
				  0,
				  "nodes",
				  "nodes",
				  0,
				  "router",
				  "peers",
				  "nodes",
				  0,
				  "paths",
				  0,
				  "uptime"
				]
			  }
			],
			"data": {
			  "allRouters": {
				"nodes": [
				  {
					"nodes": {
					  "nodes": [
						{
						  "router": {
							"peers": {
							  "nodes": [
								{
								  "paths": [
									{
									  "uptime": null,
									  "status": "UP"
									}
								  ]
								}
							  ]
							}
						  }
						}
					  ]
					}
				  }
				]
			  }
			}
		  }`},
		ExpectedMetrics: []*testutil.Metric{
			{
				Measurement: "test-collector",
				Tags:        map[string]string{},
				Fields: map[string]interface{}{
					"status": "UP",
				},
			},
		},
		ExpectedErrors: []string{
			"found errors in response for collector test-collector: Int cannot represent non 32-bit signed integer value: 3066521082",
		},
		ExpectedRequests: []int{1},
	},
}

func TestT128GraphqlCollector(t *testing.T) {
	for _, testCase := range CollectorTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			fakeServer, requestCount := createTestServer(t, testCase.Endpoint)
			defer fakeServer.Close()

			plugin := &plugin.T128GraphQL{
				CollectorName:   "test-collector",
				BaseURL:         fakeServer.URL + "/api/v1/graphql",
				EntryPoint:      testCase.EntryPoint,
				Fields:          testCase.Fields,
				Tags:            testCase.Tags,
				RetryIfNotFound: testCase.RetryIfNotFound,
			}

			var acc testutil.Accumulator

			if testCase.InitError {
				require.Error(t, plugin.Init())
				return
			} else {
				require.NoError(t, plugin.Init())
			}

			plugin.Query = testCase.Query

			for _, expectedRequests := range testCase.ExpectedRequests {
				plugin.Gather(&acc)
				require.Equal(t, expectedRequests, *requestCount)

				// Timestamps aren't important, but need to match
				for _, m := range acc.Metrics {
					m.Time = time.Time{}
				}

				// Avoid specifying this unused type for each field
				for _, m := range testCase.ExpectedMetrics {
					m.Type = telegraf.Untyped
				}
			}

			var errorStrings []string = nil
			for _, err := range acc.Errors {
				errorStrings = append(errorStrings, err.Error())
			}

			require.ElementsMatch(t, testCase.ExpectedErrors, errorStrings)
			require.ElementsMatch(t, testCase.ExpectedMetrics, acc.Metrics)
		})
	}
}

func TestTimoutUsedForRequests(t *testing.T) {
	done := make(chan struct{}, 1)

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
		}

		w.Write([]byte("[]"))
	}))

	plugin := &plugin.T128GraphQL{
		CollectorName: "test-collector",
		BaseURL:       fakeServer.URL + "/api/v1/graphql",
		EntryPoint:    "fake/entry/point",
		Fields:        map[string]string{"test-field": "test-field"},
		Tags:          map[string]string{},
		Timeout:       config.Duration(1 * time.Millisecond),
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())

	require.NoError(t, plugin.Gather(&acc))
	done <- struct{}{}

	require.Len(t, acc.Errors, 1)
	require.EqualError(
		t,
		acc.Errors[0],
		fmt.Sprintf("failed to make graphQL request for collector test-collector: Post \"%s/api/v1/graphql/\": context deadline exceeded (Client.Timeout exceeded while awaiting headers)", fakeServer.URL))

	fakeServer.Close()
}

func createTestServer(t *testing.T, endpoint Endpoint) (*httptest.Server, *int) {
	requestCount := 0
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		requestCount += 1

		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NotEqual(t, r.Header.Get("deadline"), "")
		require.Equal(t, "POST", r.Method)

		if endpoint.URL != r.URL.Path {
			fmt.Printf("There isn't an endpoint for: %v\n", r.URL.Path)
			w.WriteHeader(404)
			return
		}

		if endpoint.ExpectedRequest != "" {
			contents, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				return
			}

			require.JSONEq(t, endpoint.ExpectedRequest, string(contents), "Unexpected request body for endpoint %s", endpoint.URL)
		}

		w.WriteHeader(endpoint.Code)
		w.Write([]byte(endpoint.Response))
	}))

	return fakeServer, &requestCount
}
