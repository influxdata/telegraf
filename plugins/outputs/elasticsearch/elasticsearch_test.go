package elasticsearch

import (
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

const servicePort = "9200"

func TestConnect(t *testing.T) {
	tests := []struct {
		name              string
		index             string
		template          string
		manageTemplate    bool
		overwriteTemplate bool
	}{
		{
			name:              "template management",
			index:             "test-%Y.%m.%d",
			template:          "telegraf",
			manageTemplate:    true,
			overwriteTemplate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a mock Elasticsearch endpoint
			server := startMockServer(t)
			defer server.close()

			// Setup plugin
			plugin := &Elasticsearch{
				URLs:              []string{"http://" + server.addr()},
				IndexName:         tt.index,
				TemplateName:      tt.template,
				ManageTemplate:    tt.manageTemplate,
				OverwriteTemplate: tt.overwriteTemplate,
				Timeout:           config.Duration(time.Second * 5),
				Log:               testutil.Logger{},
			}

			require.NoError(t, plugin.Connect())
		})
	}
}

func TestConnectFail(t *testing.T) {
	tests := []struct {
		name           string
		index          string
		manageTemplate bool
		template       string
		expected       string
	}{
		{
			name:           "empty template",
			index:          "test-%Y.%m.%d",
			manageTemplate: true,
			expected:       "elasticsearch template_name configuration not defined",
		},
		{
			name:           "invalid index pattern",
			index:          "{{host}}-%Y.%m.%d",
			manageTemplate: true,
			template:       "telegraf",
			expected:       "template cannot be created for dynamic index names without an index prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a mock Elasticsearch endpoint
			server := startMockServer(t)
			defer server.close()

			// Setup plugin
			plugin := &Elasticsearch{
				URLs:           []string{"http://" + server.addr()},
				IndexName:      tt.index,
				ManageTemplate: tt.manageTemplate,
				TemplateName:   tt.template,
				Timeout:        config.Duration(time.Second * 5),
				Log:            testutil.Logger{},
			}

			require.ErrorContains(t, plugin.Connect(), tt.expected)
		})
	}
}

func TestWrite(t *testing.T) {
	// Bulk request expected for a metric with a single integer value of one
	msgBulk := []string{
		`
			{
				"index": {
					"_index": "test-2009.11.10",
					"_type": "metrics"
				}
			}
		`,
		`{
			"@timestamp": "2009-11-10T23:00:00Z",
			"measurement_name": "test1",
			"tag": {
				"tag1": "value1"
			},
			"test1": {
				"value": 1
			}
		}
		`,
	}

	tests := []struct {
		name            string
		value           any
		floatHandling   string
		floatReplace    float64
		useOpTypeCreate bool
		forceDocumentID bool
		enableGzip      bool
		authBearerToken string
		headers         map[string]any
		expected        []string
		expectedHeaders map[string][]string
	}{
		{
			name:          "float handling drop NaN",
			value:         math.NaN(),
			floatHandling: "drop",
			expected: []string{
				`
					{
						"index": {
							"_index": "test-2009.11.10",
							"_type": "metrics"
						}
					}
				`,
				`{
					"@timestamp": "2009-11-10T23:00:00Z",
					"measurement_name": "test1",
						"tag": {
							"tag1": "value1"
						},
						"test1": {
						}
					}
				`,
			},
		},
		{
			name:          "float handling drop +Inf",
			value:         math.Inf(1),
			floatHandling: "drop",
			expected: []string{
				`
					{
						"index": {
							"_index": "test-2009.11.10",
							"_type": "metrics"
						}
					}
				`,
				`{
					"@timestamp": "2009-11-10T23:00:00Z",
					"measurement_name": "test1",
						"tag": {
							"tag1": "value1"
						},
						"test1": {
						}
					}
				`,
			},
		},
		{
			name:          "float handling drop -Inf",
			value:         math.Inf(-1),
			floatHandling: "drop",
			expected: []string{
				`
					{
						"index": {
							"_index": "test-2009.11.10",
							"_type": "metrics"
						}
					}
				`,
				`{
					"@timestamp": "2009-11-10T23:00:00Z",
					"measurement_name": "test1",
						"tag": {
							"tag1": "value1"
						},
						"test1": {
						}
					}
				`,
			},
		},
		{
			name:          "float handling replace NaN",
			value:         math.NaN(),
			floatHandling: "replace",
			floatReplace:  42,
			expected: []string{
				`
					{
						"index": {
							"_index": "test-2009.11.10",
							"_type": "metrics"
						}
					}
				`,
				`{
					"@timestamp": "2009-11-10T23:00:00Z",
					"measurement_name": "test1",
						"tag": {
							"tag1": "value1"
						},
						"test1": {
							"value": 42
						}
					}
				`,
			},
		},
		{
			name:          "float handling replace +Inf",
			value:         math.Inf(1),
			floatHandling: "replace",
			floatReplace:  42,
			expected: []string{
				`
					{
						"index": {
							"_index": "test-2009.11.10",
							"_type": "metrics"
						}
					}
				`,
				`{
					"@timestamp": "2009-11-10T23:00:00Z",
					"measurement_name": "test1",
						"tag": {
							"tag1": "value1"
						},
						"test1": {
							"value": 42
						}
					}
				`,
			},
		},
		{
			name:          "float handling replace -Inf",
			value:         math.Inf(-1),
			floatHandling: "replace",
			floatReplace:  42,
			expected: []string{
				`
					{
						"index": {
							"_index": "test-2009.11.10",
							"_type": "metrics"
						}
					}
				`,
				`{
					"@timestamp": "2009-11-10T23:00:00Z",
					"measurement_name": "test1",
						"tag": {
							"tag1": "value1"
						},
						"test1": {
							"value":-42
						}
					}
				`,
			},
		},
		{
			name:     "gzip disabled",
			value:    1,
			expected: msgBulk,
			expectedHeaders: map[string][]string{
				"Content-Encoding": nil,
			},
		},
		{
			name:       "gzip enabled",
			value:      1,
			enableGzip: true,
			expected:   msgBulk,
			expectedHeaders: map[string][]string{
				"Content-Encoding": {"gzip"},
				"Accept-Encoding":  {"gzip"},
			},
		},
		{
			name:            "auth bearer token",
			value:           1,
			authBearerToken: "0123456789abcdef",
			expected:        msgBulk,
			expectedHeaders: map[string][]string{
				"Authorization": {"Bearer 0123456789abcdef"},
			},
		},
		{
			// An empty header map must not add any header
			name:     "empty headers map",
			value:    1,
			headers:  map[string]any{},
			expected: msgBulk,
		},
		{
			// Invalid types are rejected with error logging
			name:  "invalid header types",
			value: 1,
			headers: map[string]any{
				"X-Numeric": 123,
				"X-Boolean": true,
				"X-Float":   45.67,
				"X-Nil":     nil,
			},
			expected: msgBulk,
			expectedHeaders: map[string][]string{
				"X-Numeric": nil,
				"X-Boolean": nil,
				"X-Float":   nil,
				"X-Nil":     nil,
			},
		},
		{
			// Single string values are split on commas (deprecated behavior)
			name:  "header strings with commas (deprecated)",
			value: 1,
			headers: map[string]any{
				"Authorization":       "Bearer token123",
				"VL-Stream-Fields":    "tag.Source,tag.Channel,tag.EventID",
				"VL-Msg-Field":        "win_eventlog.Message",
				"CSV-Data":            "col1,col2,col3,col4",
				"Content-Disposition": `attachment; filename="file,with,commas.csv"`,
				"X-Special-Chars":     "value with spaces, commas, and \"quotes\"",
				"X-Unicode":           "测试值",
				"X-JSON-Like":         `{"key": "value", "array": [1,2,3]}`,
				"X-Empty-String":      "",
			},
			expected: msgBulk,
			expectedHeaders: map[string][]string{
				"Authorization":       {"Bearer token123"},
				"VL-Stream-Fields":    {"tag.Source", "tag.Channel", "tag.EventID"},
				"VL-Msg-Field":        {"win_eventlog.Message"},
				"CSV-Data":            {"col1", "col2", "col3", "col4"},
				"Content-Disposition": {`attachment; filename="file`, `with`, `commas.csv"`},
				"X-Special-Chars":     {"value with spaces", "commas", `and "quotes"`},
				"X-Unicode":           {"测试值"},
				"X-JSON-Like":         {`{"key": "value"`, `"array": [1`, `2`, `3]}`},
				"X-Empty-String":      {""},
			},
		},
		{
			// Arrays create multiple header values with whitespace trimmed,
			// empty arrays are ignored
			name:  "header arrays with whitespace",
			value: 1,
			headers: map[string]any{
				"Cache-Control": []any{"no-cache", "must-revalidate"},
				"X-Debug-Tags":  []any{"performance", "security", "monitoring"},
				"X-With-Spaces": []any{" application/json ", "  application/xml  ", "text/plain"},
				"X-Empty-Array": make([]any, 0),
			},
			expected: msgBulk,
			expectedHeaders: map[string][]string{
				"Cache-Control": {"no-cache", "must-revalidate"},
				"X-Debug-Tags":  {"performance", "security", "monitoring"},
				"X-With-Spaces": {"application/json", "application/xml", "text/plain"},
				"X-Empty-Array": nil,
			},
		},
		{
			// Arrays convert strings and log errors for non-string types
			name:  "header arrays with mixed types",
			value: 1,
			headers: map[string]any{
				"X-Forwarded-For":   []any{"192.168.1.1", "10.0.0.1", "172.16.0.1"},
				"X-Mixed-Types":     []any{"string-value", 123, true, "another-string"},
				"X-Empty-Interface": make([]any, 0),
			},
			expected: msgBulk,
			expectedHeaders: map[string][]string{
				"X-Forwarded-For":   {"192.168.1.1", "10.0.0.1", "172.16.0.1"},
				"X-Mixed-Types":     {"string-value", "another-string"},
				"X-Empty-Interface": nil,
			},
		},
		{
			// Headers set by the client itself are appended to, not replaced
			name:  "headers used by the client",
			value: 1,
			headers: map[string]any{
				"Content-Type": "application/json",
				"Accept":       []any{"application/json", "text/plain"},
			},
			expected: msgBulk,
			expectedHeaders: map[string][]string{
				"Content-Type": {"application/x-ndjson", "application/json"},
				"Accept":       {"application/json", "application/json", "text/plain"},
			},
		},
		{
			name:            "use_optype_create",
			value:           1,
			useOpTypeCreate: true,
			forceDocumentID: true,
			expected: []string{
				`
					{
						"create": {
							"_index": "test-2009.11.10",
							"_id": "%s",
							"_type": "metrics"
						}
					}
				`,
				`{
					"@timestamp": "2009-11-10T23:00:00Z",
					"measurement_name": "test1",
						"tag": {
							"tag1": "value1"
						},
						"test1": {
							"value": 1
						}
					}
				`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the input metric
			m := testutil.TestMetric(tt.value)

			// If a document ID is enforced we need to create it from the metric
			expected := slices.Clone(tt.expected)
			if tt.forceDocumentID {
				for i, l := range expected {
					if strings.Contains(l, `"_id": "%s"`) {
						expected[i] = fmt.Sprintf(l, getPointID(m))
					}
				}
			}

			// Setup a mock Elasticsearch endpoint
			server := startMockServer(t)
			defer server.close()

			// Setup plugin
			plugin := &Elasticsearch{
				URLs:             []string{"http://" + server.addr()},
				IndexName:        "test-%Y.%m.%d",
				Headers:          tt.headers,
				FloatHandling:    tt.floatHandling,
				FloatReplacement: tt.floatReplace,
				UseOpTypeCreate:  tt.useOpTypeCreate,
				ForceDocumentID:  tt.forceDocumentID,
				EnableGzip:       tt.enableGzip,
				Timeout:          config.Duration(time.Second * 5),
				Log:              testutil.Logger{},
			}
			if tt.authBearerToken != "" {
				plugin.AuthBearerToken = config.NewSecret([]byte(tt.authBearerToken))
			}
			require.NoError(t, plugin.Connect())

			require.NoError(t, plugin.Write([]telegraf.Metric{m}))

			// Check the action sent to the server. A bulk request consists of an
			// action line followed by the document itself.
			msgs := server.messages()
			lines := make([]string, 0, 2*len(msgs))
			for _, msg := range msgs {
				lines = append(lines, strings.Split(strings.TrimSpace(msg), "\n")...)
			}
			require.Len(t, lines, len(expected))
			for i, actual := range lines {
				require.JSONEqf(t, expected[i], actual, "mismatch in line %d", i)
			}

			// Check the request headers
			for i, hdr := range server.headers() {
				for k, v := range tt.expectedHeaders {
					require.Equalf(t, v, hdr.Values(k), "mismatch in header %q of request %d", k, i)
				}
			}
		})
	}
}

func TestWriteFail(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{
			name:     "value NaN",
			value:    math.NaN(),
			expected: "error sending bulk request to Elasticsearch: json: unsupported value: NaN",
		},
		{
			name:     "value +Inf",
			value:    math.Inf(1),
			expected: "error sending bulk request to Elasticsearch: json: unsupported value: +Inf",
		},
		{
			name:     "value -Inf",
			value:    math.Inf(-1),
			expected: "error sending bulk request to Elasticsearch: json: unsupported value: -Inf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the input metric
			m := testutil.TestMetric(tt.value)

			// Setup a mock Elasticsearch endpoint
			server := startMockServer(t)
			defer server.close()

			// Setup plugin
			plugin := &Elasticsearch{
				URLs:      []string{"http://" + server.addr()},
				IndexName: "test-%Y.%m.%d",
				Timeout:   config.Duration(time.Second * 5),
				Log:       testutil.Logger{},
			}
			require.NoError(t, plugin.Connect())

			require.ErrorContains(t, plugin.Write([]telegraf.Metric{m}), tt.expected)
		})
	}
}

func TestGetTagKeys(t *testing.T) {
	tests := []struct {
		name              string
		indexName         string
		expectedIndexName string
		expectedTagKeys   []string
	}{
		{
			name:              "novars",
			indexName:         "indexname",
			expectedIndexName: "indexname",
			expectedTagKeys:   make([]string, 0),
		}, {
			name:              "year",
			indexName:         "indexname-%Y",
			expectedIndexName: "indexname-%Y",
			expectedTagKeys:   make([]string, 0),
		}, {
			name:              "year-month",
			indexName:         "indexname-%Y-%m",
			expectedIndexName: "indexname-%Y-%m",
			expectedTagKeys:   make([]string, 0),
		}, {
			name:              "year-month-day",
			indexName:         "indexname-%Y-%m-%d",
			expectedIndexName: "indexname-%Y-%m-%d",
			expectedTagKeys:   make([]string, 0),
		}, {
			name:              "year-month-day-hour",
			indexName:         "indexname-%Y-%m-%d-%H",
			expectedIndexName: "indexname-%Y-%m-%d-%H",
			expectedTagKeys:   make([]string, 0),
		}, {
			name:              "2-digit-year-month",
			indexName:         "indexname-%y-%m",
			expectedIndexName: "indexname-%y-%m",
			expectedTagKeys:   make([]string, 0),
		}, {
			name:              "tag",
			indexName:         "indexname-{{tag1}}-%y-%m",
			expectedIndexName: "indexname-%s-%y-%m",
			expectedTagKeys:   []string{"tag1"},
		}, {
			name:              "two tags",
			indexName:         "indexname-{{tag1}}-{{tag2}}-%y-%m",
			expectedIndexName: "indexname-%s-%s-%y-%m",
			expectedTagKeys:   []string{"tag1", "tag2"},
		}, {
			name:              "three tags",
			indexName:         "indexname-{{tag1}}-{{tag2}}-{{tag3}}-%y-%m",
			expectedIndexName: "indexname-%s-%s-%s-%y-%m",
			expectedTagKeys:   []string{"tag1", "tag2", "tag3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexName, tagKeys := getTagKeys(tt.indexName)
			require.Equal(t, tt.expectedIndexName, indexName)
			require.Equal(t, tt.expectedTagKeys, tagKeys)
		})
	}
}

func TestGetIndexName(t *testing.T) {
	tests := []struct {
		name      string
		indexName string
		expected  string
		tagKeys   []string
	}{
		{
			name:      "novars",
			indexName: "indexname",
			expected:  "indexname",
		},
		{
			name:      "year",
			indexName: "indexname-%Y",
			expected:  "indexname-2014",
		},
		{
			name:      "year-month",
			indexName: "indexname-%Y-%m",
			expected:  "indexname-2014-12",
		},
		{
			name:      "year-month-day",
			indexName: "indexname-%Y-%m-%d",
			expected:  "indexname-2014-12-01",
		},
		{
			name:      "year-month-day-hour",
			indexName: "indexname-%Y-%m-%d-%H",
			expected:  "indexname-2014-12-01-23",
		},
		{
			name:      "2-digit-year-month",
			indexName: "indexname-%y-%m",
			expected:  "indexname-14-12",
		},
		{
			name:      "year-week",
			indexName: "indexname-%Y-%V",
			expected:  "indexname-2014-49",
		},
		{
			name:      "tag",
			indexName: "indexname-%s-%y-%m",
			expected:  "indexname-value1-14-12",
			tagKeys:   []string{"tag1"},
		},
		{
			name:      "two tags",
			indexName: "indexname-%s-%s-%y-%m",
			expected:  "indexname-value1-value2-14-12",
			tagKeys:   []string{"tag1", "tag2"},
		},
		{
			name:      "three tags",
			indexName: "indexname-%s-%s-%s-%y-%m",
			expected:  "indexname-value1-value2-none-14-12",
			tagKeys:   []string{"tag1", "tag2", "tag3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventTime := time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC)
			tags := map[string]string{"tag1": "value1", "tag2": "value2"}

			// Setup plugin
			plugin := &Elasticsearch{
				DefaultTagValue: "none",
				Log:             testutil.Logger{},
			}

			indexName := plugin.getIndexName(tt.indexName, eventTime, tt.tagKeys, tags)
			require.Equal(t, tt.expected, indexName)
		})
	}
}

func TestGetPipelineName(t *testing.T) {
	tests := []struct {
		name            string
		pipeline        string
		defaultPipeline string
		tags            map[string]string
		expected        string
	}{
		{
			name: "default without pipeline or default",
			tags: map[string]string{"tag1": "value1", "tag2": "value2"},
		},
		{
			name:            "default only",
			defaultPipeline: "myDefaultPipeline",
			tags:            map[string]string{"tag1": "value1", "tag2": "value2"},
		},
		{
			name:            "default only with pipeline tag",
			defaultPipeline: "myDefaultPipeline",
			tags:            map[string]string{"tag1": "value1", "es-pipeline": "pipeline2"},
		},
		{
			name: "defined without pipeline",
			tags: map[string]string{"tag1": "value1", "es-pipeline": "myOtherPipeline"},
		},
		{
			name:     "missing tag without default",
			pipeline: "{{es-pipeline}}",
			tags:     map[string]string{"tag1": "value1"},
		},
		{
			name:            "default with pipeline",
			pipeline:        "{{es-pipeline}}",
			defaultPipeline: "myDefaultPipeline",
			tags:            map[string]string{"tag1": "value1", "tag2": "value2"},
			expected:        "myDefaultPipeline",
		},
		{
			name:            "defined with pipeline",
			pipeline:        "{{es-pipeline}}",
			defaultPipeline: "myDefaultPipeline",
			tags:            map[string]string{"tag1": "value1", "es-pipeline": "myOtherPipeline"},
			expected:        "myOtherPipeline",
		},
		{
			name:     "static pipeline",
			pipeline: "myDefaultPipeline",
			tags:     map[string]string{"tag1": "value1", "es-pipeline": "myOtherPipeline"},
			expected: "myDefaultPipeline",
		},
		{
			name:     "tag pipeline without default",
			pipeline: "{{es-pipeline}}",
			tags:     map[string]string{"tag1": "value1", "es-pipeline": "pipeline2"},
			expected: "pipeline2",
		},
		{
			name:     "multiple tags in pipeline",
			pipeline: "{{tag1}}-{{es-pipeline}}",
			tags:     map[string]string{"tag1": "value1", "es-pipeline": "pipeline2"},
			expected: "value1-pipeline2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pn, ptags := getTagKeys(tt.pipeline)

			// Setup plugin
			plugin := &Elasticsearch{
				UsePipeline:     tt.pipeline,
				DefaultPipeline: tt.defaultPipeline,
				Log:             testutil.Logger{},
				pipelineName:    pn,
				pipelineTagKeys: ptags,
			}

			pipelineName := plugin.getPipelineName(plugin.pipelineName, plugin.pipelineTagKeys, tt.tags)
			require.Equal(t, tt.expected, pipelineName)
		})
	}
}

func TestIndexSettings(t *testing.T) {
	tests := []struct {
		name     string
		template map[string]any
		expected string
	}{
		{
			name: "standard",
			expected: `
				{
					"template": "test*",
					"settings": {
						"index": {
							"refresh_interval": "10s",
							"mapping.total_fields.limit": 5000,
							"auto_expand_replicas": "0-1",
							"codec": "best_compression"
						}
					},
					"mappings": {
						"metrics": {
							"_all": {
								"enabled": false
							},
							"properties": {
								"@timestamp": {
									"type": "date"
								},
								"measurement_name": {
									"type": "keyword"
								}
							},
							"dynamic_templates": [
								{
									"tags": {
										"match_mapping_type": "string",
										"path_match": "tag.*",
										"mapping": {
											"ignore_above": 512,
											"type": "keyword"
										}
									}
								},
								{
									"metrics_long": {
										"match_mapping_type": "long",
										"mapping": {
											"type": "float",
											"index": false
										}
									}
								},
								{
									"metrics_double": {
										"match_mapping_type": "double",
										"mapping": {
											"type": "float",
											"index": false
										}
									}
								},
								{
									"text_fields": {
										"match": "*",
										"mapping": {
											"norms": false
										}
									}
								}
							]
						}
					}
				}
			`,
		},
		{
			name: "custom index",
			template: map[string]any{
				"refresh_interval":           "20s",
				"mapping.total_fields.limit": 1000,
				"codec":                      "best_compression",
			},
			expected: `
				{
					"template": "test*",
					"settings": {
						"index": {
							"codec": "best_compression",
							"mapping.total_fields.limit": 1000,
							"refresh_interval": "20s"
						}
					},
					"mappings": {
						"metrics": {
							"_all": {
								"enabled": false
							},
							"properties": {
								"@timestamp": {
									"type": "date"
								},
								"measurement_name": {
									"type": "keyword"
								}
							},
							"dynamic_templates": [
								{
									"tags": {
										"match_mapping_type": "string",
										"path_match": "tag.*",
										"mapping": {
											"ignore_above": 512,
											"type": "keyword"
										}
									}
								},
								{
									"metrics_long": {
										"match_mapping_type": "long",
										"mapping": {
											"type": "float",
											"index": false
										}
									}
								},
								{
									"metrics_double": {
										"match_mapping_type": "double",
										"mapping": {
											"type": "float",
											"index": false
										}
									}
								},
								{
									"text_fields": {
										"match": "*",
										"mapping": {
											"norms": false
										}
									}
								}
							]
						}
					}
				}
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin
			plugin := &Elasticsearch{
				TemplateName:  "test",
				IndexName:     "telegraf-%Y.%m.%d",
				IndexTemplate: tt.template,
				Log:           testutil.Logger{},
			}

			buf, err := plugin.createNewTemplate("test")
			require.NoError(t, err)
			require.JSONEq(t, tt.expected, buf.String())
		})
	}
}

func TestWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup container
	container := &testutil.Container{
		Image:        "elasticsearch:6.8.23",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"discovery.type": "single-node",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("] mode [basic] - valid"),
			wait.ForListeningPort(servicePort),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Setup plugin
	plugin := &Elasticsearch{
		URLs:                []string{"http://" + container.Address + ":" + container.Ports[servicePort]},
		IndexName:           "test-%Y.%m.%d",
		TemplateName:        "telegraf",
		ManageTemplate:      true,
		EnableGzip:          true,
		Timeout:             config.Duration(time.Second * 5),
		HealthCheckInterval: config.Duration(time.Second * 10),
		HealthCheckTimeout:  config.Duration(time.Second * 1),
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())

	// Verify that we can successfully write data to Elasticsearch
	require.NoError(t, plugin.Write(testutil.MockMetrics()))
}

type mockServer struct {
	server *httptest.Server
	msgs   []string
	hdrs   []http.Header
	sync.Mutex
}

func startMockServer(t *testing.T) *mockServer {
	s := &mockServer{
		msgs: make([]string, 0),
		hdrs: make([]http.Header, 0),
	}

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_bulk":
			// The request body is compressed if the plugin is configured to use
			// gzip, so we need to decompress it before recording the message
			reader := r.Body
			if r.Header.Get("Content-Encoding") == "gzip" {
				gzipReader, err := gzip.NewReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Error(err)
					return
				}
				defer gzipReader.Close()
				reader = gzipReader
			}

			body, err := io.ReadAll(reader)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
				return
			}
			s.Lock()
			s.msgs = append(s.msgs, string(body))
			s.hdrs = append(s.hdrs, r.Header.Clone())
			s.Unlock()

			if _, err := w.Write([]byte("{}")); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
			}
		default:
			if _, err := w.Write([]byte(`{"version": {"number": "6.8.23"}}`)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
			}
		}
	}))

	return s
}

func (s *mockServer) close() {
	s.server.Close()
}

func (s *mockServer) addr() string {
	return s.server.Listener.Addr().String()
}

func (s *mockServer) messages() []string {
	s.Lock()
	defer s.Unlock()
	return slices.Clone(s.msgs)
}

func (s *mockServer) headers() []http.Header {
	s.Lock()
	defer s.Unlock()
	return slices.Clone(s.hdrs)
}
