package t128_graphql_test

import (
	"fmt"
	"testing"

	"github.com/Jeffail/gabs"
	plugin "github.com/influxdata/telegraf/plugins/inputs/t128_graphql"
	"github.com/stretchr/testify/require"
)

const (
	complexResponseBasePath = ".data.allRouters.nodes.peers.nodes"
)

var ResponseProcessingTestCases = []struct {
	Name           string
	Fields         map[string]string
	Tags           map[string]string
	JsonInput      *gabs.Container
	ExpectedOutput []*plugin.ProcessedResponse
	ExpectedError  error
}{
	{
		Name:   "no data produces error",
		Fields: map[string]string{"/data/test-field": "test-field"},
		Tags:   map[string]string{"/data/test-tag": "test-tag"},
		JsonInput: generateJsonTestData([]byte(`{"data": {
				"test-field": null,
				"test-tag": null
			}}`)),
		ExpectedOutput: nil,
		ExpectedError:  fmt.Errorf("no data collected for collector test-collector"),
	},
	{
		Name:   "none value is dropped",
		Fields: map[string]string{".data.test-field": "test-field"},
		Tags:   map[string]string{".data.test-tag": "test-tag"},
		JsonInput: generateJsonTestData([]byte(`{"data": {
				"test-field": null,
				"test-tag": "test-string"
			}}`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{},
				Tags:   map[string]string{"test-tag": "test-string"},
			},
		},
		ExpectedError: nil,
	},
	{
		Name:   "process response with number tag",
		Fields: map[string]string{".data.test-field": "test-field"},
		Tags:   map[string]string{".data.test-tag": "test-tag"},
		JsonInput: generateJsonTestData([]byte(`{"data": {
				"test-field": 128,
				"test-tag": 128
			}}`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{"test-field": 128.0},
				Tags:   map[string]string{"test-tag": "128"},
			},
		},
		ExpectedError: nil,
	},
	{
		Name:   "process response with multiple fields",
		Fields: map[string]string{".data.test-field-1": "test-field-1", ".data.test-field-2": "test-field-2"},
		Tags:   map[string]string{".data.test-tag": "test-tag"},
		JsonInput: generateJsonTestData([]byte(`{"data": {
				"test-field-1": 128,
				"test-field-2": 95,
				"test-tag": "test-string"
	  		}}`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{"test-field-1": 128.0, "test-field-2": 95.0},
				Tags:   map[string]string{"test-tag": "test-string"},
			},
		},
		ExpectedError: nil,
	},
	{
		Name:   "process response with multiple tags",
		Fields: map[string]string{".data.test-field": "test-field"},
		Tags:   map[string]string{".data.test-tag-1": "test-tag-1", ".data.test-tag-2": "test-tag-2"},
		JsonInput: generateJsonTestData([]byte(`{"data": {
				"test-field": 128,
		  		"test-tag-1": "test-string-1",
		  		"test-tag-2": "test-string-2"
	  		}}`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{"test-field": 128.0},
				Tags:   map[string]string{"test-tag-1": "test-string-1", "test-tag-2": "test-string-2"},
			},
		},
		ExpectedError: nil,
	},
	{
		Name:   "process response with multiple tags some none value",
		Fields: map[string]string{".data.test-field": "test-field"},
		Tags:   map[string]string{".data.test-tag-1": "test-tag-1", ".data.test-tag-2": "test-tag-2"},
		JsonInput: generateJsonTestData([]byte(`{"data": {
				"test-field": 128,
		  		"test-tag-1": "test-string-1",
		  		"test-tag-2": null
	  		}}`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{"test-field": 128.0},
				Tags:   map[string]string{"test-tag-1": "test-string-1"},
			},
		},
		ExpectedError: nil,
	},
	{
		Name:   "renames tags and fields",
		Fields: map[string]string{".data.test-field": "test-field-renamed"},
		Tags:   map[string]string{".data.test-tag": "test-tag-renamed"},
		JsonInput: generateJsonTestData([]byte(`{"data": {
				"test-field": 128,
				"test-tag": 128
			}}`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{"test-field-renamed": 128.0},
				Tags:   map[string]string{"test-tag-renamed": "128"},
			},
		},
		ExpectedError: nil,
	},
	{
		Name:   "process response with multiple nodes",
		Fields: map[string]string{".data.test-field": "test-field"},
		Tags:   map[string]string{".data.test-tag": "test-tag"},
		JsonInput: generateJsonTestData([]byte(`{"data": [
			{
				"test-field": 128,
				"test-tag": "test-string-1"
			},
			{
				"test-field": 95,
				"test-tag": "test-string-2"
			}]}`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{"test-field": 128.0},
				Tags:   map[string]string{"test-tag": "test-string-1"},
			},
			{
				Fields: map[string]interface{}{"test-field": 95.0},
				Tags:   map[string]string{"test-tag": "test-string-2"},
			},
		},
		ExpectedError: nil,
	},
	{
		Name:   "process response with nested tags",
		Fields: map[string]string{".data.test-field": "test-field"},
		Tags:   map[string]string{".data.test-tag-1": "test-tag-1", ".data.state.test-tag-2": "test-tag-2"},
		JsonInput: generateJsonTestData([]byte(`{"data": [
			{
				"test-field": 128,
				"test-tag-1": "test-string-1",
			  	"state": {
				  "test-tag-2": "test-string-2"
			  	}
			},
			{
				"test-field": 95,
				"test-tag-1": "test-string-3",
				"state": {
					"test-tag-2": "test-string-4"
				}
			}
		]}`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{"test-field": 128.0},
				Tags:   map[string]string{"test-tag-1": "test-string-1", "test-tag-2": "test-string-2"},
			},
			{
				Fields: map[string]interface{}{"test-field": 95.0},
				Tags:   map[string]string{"test-tag-1": "test-string-3", "test-tag-2": "test-string-4"},
			},
		},
		ExpectedError: nil,
	},
	{
		Name:   "process response with multi-level nested tags",
		Fields: map[string]string{".data.test-field": "test-field"},
		Tags:   map[string]string{".data.test-tag-1": "test-tag-1", ".data.state1.state2.state3.test-tag-2": "test-tag-2"},
		JsonInput: generateJsonTestData([]byte(`{"data": {
				"test-field": 128,
				"test-tag-1": "test-string-1",
			  	"state1": {
					"state2": {
						"state3": {
							"test-tag-2": "test-string-2"
						}
					}
			  	}
		  	}}`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{"test-field": 128.0},
				Tags:   map[string]string{"test-tag-1": "test-string-1", "test-tag-2": "test-string-2"},
			},
		},
		ExpectedError: nil,
	},
	{
		Name:   "process response with multi-level nested fields",
		Fields: map[string]string{".data.state1.state2.state3.test-field-1": "test-field-1", ".data.test-field-2": "test-field-2"},
		Tags:   map[string]string{".data.test-tag": "test-tag"},
		JsonInput: generateJsonTestData([]byte(`{"data": {
				"test-field-2": 128,
				"test-tag": "test-string-1",
			  	"state1": {
					"state2": {
						"state3": {
							"test-field-1": 95
						}
					}
			  	}
		  	}}`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{"test-field-1": 95.0, "test-field-2": 128.0},
				Tags:   map[string]string{"test-tag": "test-string-1"},
			},
		},
		ExpectedError: nil,
	},
	{
		Name:   "process response with mixed nesting",
		Fields: map[string]string{".data.state1.state2.test-field": "test-field"},
		Tags:   map[string]string{".data.test-tag-1": "test-tag-1", ".data.state1.state2.state3.test-tag-2": "test-tag-2"},
		JsonInput: generateJsonTestData([]byte(`{"data": {
				"test-tag-1": "test-string-1",
			  	"state1": {
					"state2": {
						"test-field": 128,
						"state3": {
							"test-tag-2": "test-string-2"
						}
					}
			  	}
		  	}}`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{"test-field": 128.0},
				Tags:   map[string]string{"test-tag-1": "test-string-1", "test-tag-2": "test-string-2"},
			},
		},
		ExpectedError: nil,
	},
	{
		Name: "process complex response",
		Fields: map[string]string{
			complexResponseBasePath + ".paths.adjacentAddress":  "adjacent-address",
			complexResponseBasePath + ".paths.adjacentHostname": "adjacent-hostname",
			complexResponseBasePath + ".paths.isActive":         "is-active",
			complexResponseBasePath + ".paths.uptime":           "uptime",
		},
		Tags: map[string]string{
			complexResponseBasePath + ".name":                  "peer-name",
			complexResponseBasePath + ".paths.deviceInterface": "device-interface",
			complexResponseBasePath + ".paths.vlan":            "vlan",
		},
		JsonInput: generateJsonTestData([]byte(`{"data": {
			  "allRouters": {
				"nodes": [
				  {
					"peers": {
					  "nodes": [
						{
						  "name": "AZDCBBP1",
						  "paths": [
							{
							  "vlan": 0,
							  "uptime": 188333176,
							  "adjacentAddress": "12.51.52.30",
							  "adjacentHostname": null,
							  "deviceInterface": "StoreLTE",
							  "isActive": true
							},
							{
							  "vlan": 0,
							  "uptime": 82247253,
							  "adjacentAddress": "12.51.52.30",
							  "adjacentHostname": null,
							  "deviceInterface": "StoreWAN",
							  "isActive": true
							}
						  ]
						},
						{
						  "name": "AZDCLTEP1",
						  "paths": [
							{
							  "vlan": 0,
							  "uptime": 162241794,
							  "adjacentAddress": "12.51.52.22",
							  "adjacentHostname": null,
							  "deviceInterface": "StoreLTE",
							  "isActive": true
							},
							{
							  "vlan": 0,
							  "uptime": 82247352,
							  "adjacentAddress": "12.51.52.22",
							  "adjacentHostname": null,
							  "deviceInterface": "StoreWAN",
							  "isActive": true
							}
						  ]
						}
					  ]
					}
				  }
				]
			  }
		}}`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{"adjacent-address": "12.51.52.30", "is-active": true, "uptime": 188333176.0},
				Tags:   map[string]string{"peer-name": "AZDCBBP1", "device-interface": "StoreLTE", "vlan": "0"},
			},
			{
				Fields: map[string]interface{}{"adjacent-address": "12.51.52.30", "is-active": true, "uptime": 82247253.0},
				Tags:   map[string]string{"peer-name": "AZDCBBP1", "device-interface": "StoreWAN", "vlan": "0"},
			},
			{
				Fields: map[string]interface{}{"adjacent-address": "12.51.52.22", "is-active": true, "uptime": 162241794.0},
				Tags:   map[string]string{"peer-name": "AZDCLTEP1", "device-interface": "StoreLTE", "vlan": "0"},
			},
			{
				Fields: map[string]interface{}{"adjacent-address": "12.51.52.22", "is-active": true, "uptime": 82247352.0},
				Tags:   map[string]string{"peer-name": "AZDCLTEP1", "device-interface": "StoreWAN", "vlan": "0"},
			},
		},
		ExpectedError: nil,
	},
	{
		Name: "process response for tags and fields with absolute path",
		Fields: map[string]string{
			".data.allServices.nodes.timeSeriesAnalytic.value":     "value",
			".data.allServices.nodes.timeSeriesAnalytic.timestamp": "timestamp",
			".data.allServices.nodes.other":                        "other-field",
		},
		Tags: map[string]string{
			".data.allServices.nodes.timeSeriesAnalytic.test-tag": "test-tag",
			".data.allServices.nodes.name":                        "name",
		},
		JsonInput: generateJsonTestData([]byte(`{
			"data": {
			  "allServices": {
				"nodes": [
				  {
					"name": "east",
					"other": "moo",
					"timeSeriesAnalytic": [
					  {
						"timestamp": "2021-06-14T21:10:00Z",
						"value": "0",
						"test-tag": "foo"
					  }
					]
				  },
				  {
					"name": "west",
					"other": "cow",
					"timeSeriesAnalytic": [
					  {
						"timestamp": "2021-06-14T21:10:00Z",
						"value": "128",
						"test-tag": "bar"
					  }
					]
				  }
				]
			  }
			}
		  }`)),
		ExpectedOutput: []*plugin.ProcessedResponse{
			{
				Fields: map[string]interface{}{
					"value":       "0",
					"timestamp":   "2021-06-14T21:10:00Z",
					"other-field": "moo",
				},
				Tags: map[string]string{
					"test-tag": "foo",
					"name":     "east",
				},
			},
			{
				Fields: map[string]interface{}{
					"value":       "128",
					"timestamp":   "2021-06-14T21:10:00Z",
					"other-field": "cow",
				},
				Tags: map[string]string{
					"test-tag": "bar",
					"name":     "west",
				},
			},
		},
		ExpectedError: nil,
	},
}

func generateJsonTestData(data []byte) *gabs.Container {
	gabsData, err := gabs.ParseJSON(data)
	if err != nil {
		panic(err)
	}
	return gabsData
}

func TestT128GraphqlResponseProcessing(t *testing.T) {
	for _, testCase := range ResponseProcessingTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			processedResponse, err := plugin.ProcessResponse(
				testCase.JsonInput,
				"test-collector",
				testCase.Fields,
				testCase.Tags,
			)

			require.Equal(t, testCase.ExpectedError, err)
			require.ElementsMatch(t, testCase.ExpectedOutput, processedResponse)
		})
	}
}
