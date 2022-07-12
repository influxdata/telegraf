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
	ValidExpectedRequestSingleTag = `{"query":"query {\nallRouters(name:\"ComboEast\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ntest-field\ntest-tag}}}}}}}"}`
	ValidQuerySingleTag           = "query {\nallRouters(name:\"ComboEast\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ntest-field\ntest-tag}}}}}}}"
	InvalidRouterExpectedRequest  = `{"query":"query {\nallRouters(name:\"not-a-router\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ntest-field\ntest-tag}}}}}}}"}`
	InvalidRouterQuery            = "query {\nallRouters(name:\"not-a-router\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ntest-field\ntest-tag}}}}}}}"
	InvalidFieldExpectedRequest   = `{"query":"query {\nallRouters(name:\"ComboEast\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ninvalid-field\ntest-tag}}}}}}}"}`
	InvalidFieldQuery             = "query {\nallRouters(name:\"ComboEast\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ninvalid-field\ntest-tag}}}}}}}"
)

var CollectorTestCases = []struct {
	Name            string
	EntryPoint      string
	Fields          map[string]string
	Tags            map[string]string
	Query           string
	Endpoint        Endpoint
	ExpectedMetrics []*testutil.Metric
	ExpectedErrors  []string
}{
	{
		Name:            "empty query produces no request or metrics",
		EntryPoint:      "",
		Fields:          nil,
		Tags:            nil,
		Query:           "",
		Endpoint:        Endpoint{},
		ExpectedMetrics: nil,
		ExpectedErrors:  nil,
	},
	{
		Name:            "empty response produces error",
		EntryPoint:      "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:          map[string]string{"test-field": "test-field"},
		Tags:            map[string]string{"test-tag": "test-tag"},
		Query:           ValidQuerySingleTag,
		Endpoint:        Endpoint{"/api/v1/graphql/", 200, ValidExpectedRequestSingleTag, "{}"},
		ExpectedMetrics: nil,
		ExpectedErrors: []string{
			"empty response for collector test-collector: {}",
		},
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
	},
	{
		Name:            "propogates invalid json error to accumulator",
		EntryPoint:      "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:          map[string]string{"test-field": "test-field"},
		Tags:            map[string]string{"test-tag": "test-tag"},
		Query:           ValidQuerySingleTag,
		Endpoint:        Endpoint{"/api/v1/graphql/", 200, ValidExpectedRequestSingleTag, `{"test": }`},
		ExpectedMetrics: nil,
		ExpectedErrors:  []string{"invalid json response for collector test-collector: invalid character '}' looking for beginning of value"},
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
		ExpectedErrors:  []string{"unexpected response for collector test-collector: Cannot query field \"invalid-field\" on type \"ArpEntryType\"."},
	},
	{
		Name:       "processes simple response", //more cases units are tested separately
		EntryPoint: "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:     map[string]string{"test-field": "test-field"},
		Tags:       map[string]string{"test-tag": "test-tag"},
		Query:      ValidQuerySingleTag,
		Endpoint: Endpoint{"/api/v1/graphql/", 200, ValidExpectedRequestSingleTag, `{
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
		ExpectedErrors: nil,
	},
}

func TestT128GraphqlCollector(t *testing.T) {
	for _, testCase := range CollectorTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			fakeServer := createTestServer(t, testCase.Endpoint)
			defer fakeServer.Close()

			plugin := &plugin.T128GraphQL{
				CollectorName: "test-collector",
				BaseURL:       fakeServer.URL + "/api/v1/graphql",
				EntryPoint:    testCase.EntryPoint,
				Fields:        testCase.Fields,
				Tags:          testCase.Tags,
			}

			var acc testutil.Accumulator

			if plugin.EntryPoint == "" {
				require.Error(t, plugin.Init())
				return
			} else {
				require.NoError(t, plugin.Init())
			}

			plugin.Query = testCase.Query
			plugin.Gather(&acc)

			var errorStrings []string = nil
			for _, err := range acc.Errors {
				errorStrings = append(errorStrings, err.Error())
			}

			require.ElementsMatch(t, testCase.ExpectedErrors, errorStrings)

			// Timestamps aren't important, but need to match
			for _, m := range acc.Metrics {
				m.Time = time.Time{}
			}

			// Avoid specifying this unused type for each field
			for _, m := range testCase.ExpectedMetrics {
				m.Type = telegraf.Untyped
			}

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

func createTestServer(t *testing.T, endpoint Endpoint) *httptest.Server {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
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

	return fakeServer
}
