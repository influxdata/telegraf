package elasticsearch

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	elasticsearch5 "github.com/elastic/go-elasticsearch/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

const servicePort = "9200"

func TestConnect(t *testing.T) {
	tests := []struct {
		version           string
		name              string
		index             string
		template          string
		manageTemplate    bool
		overwriteTemplate bool
		existingTemplate  bool
		expectedOverwrite bool
	}{
		{
			name:              "Elasticsearch v5 template management",
			version:           "5.6.16",
			index:             "test-%Y.%m.%d",
			template:          "telegraf",
			manageTemplate:    true,
			expectedOverwrite: true,
		},
		{
			name:              "Elasticsearch v7 template management",
			version:           "7.17.29",
			index:             "test-%Y.%m.%d",
			template:          "telegraf",
			manageTemplate:    true,
			expectedOverwrite: true,
		},
		{
			name:              "template management",
			index:             "test-%Y.%m.%d",
			template:          "telegraf",
			manageTemplate:    true,
			overwriteTemplate: true,
			expectedOverwrite: true,
		},
		{
			name:              "create non-existing template",
			index:             "test-%Y.%m.%d",
			template:          "telegraf",
			manageTemplate:    true,
			expectedOverwrite: true,
		},
		{
			name:              "overwrite existing template",
			index:             "test-%Y.%m.%d",
			template:          "telegraf",
			manageTemplate:    true,
			overwriteTemplate: true,
			existingTemplate:  true,
			expectedOverwrite: true,
		},
		{
			name:             "keep existing template",
			index:            "test-%Y.%m.%d",
			template:         "telegraf",
			manageTemplate:   true,
			existingTemplate: true,
		},
		{
			name:             "no template management",
			index:            "test-%Y.%m.%d",
			template:         "telegraf",
			existingTemplate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a mock Elasticsearch endpoint
			server := startMockServer(t)
			if tt.version != "" {
				server.setVersion(tt.version)
			}
			defer server.close()

			// Pre-populate the template if requested
			const existing = `{"existing": true}`
			if tt.existingTemplate {
				server.setTemplate(tt.template, existing)
			}

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

			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			// Check the template state
			tmpl, found := server.template(tt.template)
			if tt.expectedOverwrite {
				require.True(t, found, "template not created")
				require.NotEqual(t, existing, tmpl, "template not updated")
			} else {
				require.Equal(t, tt.existingTemplate, found)
				if tt.existingTemplate {
					require.Equal(t, existing, tmpl, "existing template modified")
				}
			}
		})
	}
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *Elasticsearch
		expected string
	}{
		{
			name:     "no URLs",
			plugin:   &Elasticsearch{IndexName: "test-%Y.%m.%d"},
			expected: "no 'urls' defined",
		},
		{
			name:     "no index name",
			plugin:   &Elasticsearch{URLs: []string{"http://localhost:9200"}},
			expected: "no 'index_name' defined",
		},
		{
			name: "invalid float handling",
			plugin: &Elasticsearch{
				URLs:          []string{"http://localhost:9200"},
				IndexName:     "test-%Y.%m.%d",
				FloatHandling: "invalid",
			},
			expected: `invalid 'float_handling' type "invalid"`,
		},
		{
			name: "empty template name",
			plugin: &Elasticsearch{
				URLs:           []string{"http://localhost:9200"},
				IndexName:      "test-%Y.%m.%d",
				ManageTemplate: true,
			},
			expected: "no 'template_name' defined",
		},
		{
			name: "index name with leading date",
			plugin: &Elasticsearch{
				URLs:           []string{"http://localhost:9200"},
				IndexName:      "%Y.%m.%d",
				TemplateName:   "telegraf",
				ManageTemplate: true,
			},
			expected: "dynamic 'index_name' must have a static prefix",
		},
		{
			name: "index name with leading tag",
			plugin: &Elasticsearch{
				URLs:           []string{"http://localhost:9200"},
				IndexName:      "{{host}}-%Y.%m.%d",
				TemplateName:   "telegraf",
				ManageTemplate: true,
			},
			expected: "dynamic 'index_name' must have a static prefix",
		},
		{
			name: "header with integer value",
			plugin: &Elasticsearch{
				URLs:      []string{"http://localhost:9200"},
				IndexName: "test-%Y.%m.%d",
				Headers:   map[string]any{"X-Invalid": 123},
			},
			expected: `header "X-Invalid" has invalid type int, expected string or []string`,
		},
		{
			name: "header with boolean value",
			plugin: &Elasticsearch{
				URLs:      []string{"http://localhost:9200"},
				IndexName: "test-%Y.%m.%d",
				Headers:   map[string]any{"X-Invalid": true},
			},
			expected: `header "X-Invalid" has invalid type bool, expected string or []string`,
		},
		{
			name: "header with float value",
			plugin: &Elasticsearch{
				URLs:      []string{"http://localhost:9200"},
				IndexName: "test-%Y.%m.%d",
				Headers:   map[string]any{"X-Invalid": 45.67},
			},
			expected: `header "X-Invalid" has invalid type float64, expected string or []string`,
		},
		{
			name: "header with nil value",
			plugin: &Elasticsearch{
				URLs:      []string{"http://localhost:9200"},
				IndexName: "test-%Y.%m.%d",
				Headers:   map[string]any{"X-Invalid": nil},
			},
			expected: `header "X-Invalid" has invalid type <nil>, expected string or []string`,
		},
		{
			name: "header with array with non-string value",
			plugin: &Elasticsearch{
				URLs:      []string{"http://localhost:9200"},
				IndexName: "test-%Y.%m.%d",
				Headers:   map[string]any{"X-Invalid": []any{"string-value", 123, "another-string"}},
			},
			expected: `header "X-Invalid" contains non-string value in array: 123 with type int`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = testutil.Logger{}
			require.ErrorContains(t, tt.plugin.Init(), tt.expected)
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
		version         string
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
			name:     "Elasticsearch v5",
			version:  "5.6.16",
			value:    1,
			expected: msgBulk,
		},
		{
			name:    "Elasticsearch v7",
			version: "7.17.29",
			value:   1,
			expected: []string{
				`{"index":{"_index":"test-2009.11.10"}}`,
				msgBulk[1],
			},
		},
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
			if tt.version != "" {
				server.setVersion(tt.version)
			}
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
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

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

func TestWriteAuthentication(t *testing.T) {
	tests := []struct {
		name        string
		userinfo    string
		username    string
		password    string
		bearerToken string
		expected    []string
	}{
		{
			name:     "rotating basic credentials",
			username: "@{test:credential}",
			password: "@{test:credential}",
			expected: []string{"Basic Zmlyc3Q6Zmlyc3Q=", "Basic c2Vjb25kOnNlY29uZA=="},
		},
		{
			name:        "rotating bearer token",
			bearerToken: "@{test:credential}",
			expected:    []string{"Bearer first", "Bearer second"},
		},
		{
			name:     "literal URL password",
			userinfo: "user:%40%7Ba%3Ab%7D",
			expected: []string{"Basic dXNlcjpAe2E6Yn0=", "Basic dXNlcjpAe2E6Yn0="},
		},
		{
			name:     "configured credentials override URL credentials",
			userinfo: "urluser:urlpassword",
			username: "@{test:credential}",
			password: "@{test:credential}",
			expected: []string{"Basic Zmlyc3Q6Zmlyc3Q=", "Basic c2Vjb25kOnNlY29uZA=="},
		},
	}
	for _, version := range []string{"5.6.16", "6.8.23", "7.17.29"} {
		t.Run(version, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					server := startMockServer(t)
					server.setVersion(version)
					defer server.close()

					address := "http://" + server.addr()
					if tt.userinfo != "" {
						address = "http://" + tt.userinfo + "@" + server.addr()
					}
					plugin := &Elasticsearch{
						URLs:            []string{address},
						IndexName:       "test-%Y.%m.%d",
						Username:        config.NewSecret([]byte(tt.username)),
						Password:        config.NewSecret([]byte(tt.password)),
						AuthBearerToken: config.NewSecret([]byte(tt.bearerToken)),
						Timeout:         config.Duration(5 * time.Second),
						Log:             testutil.Logger{},
					}
					credential := "first"
					resolvers := map[string]telegraf.ResolveFunc{
						"@{test:credential}": func() ([]byte, bool, error) {
							return []byte(credential), true, nil
						},
					}
					for _, secret := range []*config.Secret{&plugin.Username, &plugin.Password, &plugin.AuthBearerToken} {
						require.NoError(t, secret.Link(resolvers))
					}
					require.NoError(t, plugin.Init())
					require.NoError(t, plugin.Connect())
					defer plugin.Close()

					require.NoError(t, plugin.Write([]telegraf.Metric{testutil.TestMetric(1)}))
					credential = "second"
					require.NoError(t, plugin.Write([]telegraf.Metric{testutil.TestMetric(1)}))

					headers := server.headers()
					require.Len(t, headers, len(tt.expected))
					for i, expected := range tt.expected {
						require.Equal(t, []string{expected}, headers[i].Values("Authorization"))
					}
				})
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
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			require.ErrorContains(t, plugin.Write([]telegraf.Metric{m}), tt.expected)
		})
	}
}

func TestWriteBulkFail(t *testing.T) {
	tests := []struct {
		name        string
		metricCount int
		status      int
		response    string
		expected    string
	}{
		{
			name:        "HTTP failure",
			metricCount: 1,
			status:      http.StatusServiceUnavailable,
			response:    `{"error":{"type":"unavailable_shards_exception","reason":"no available shard"}}`,
			expected:    "no available shard",
		},
		{
			name:        "invalid bulk response",
			metricCount: 1,
			status:      http.StatusOK,
			response:    `{`,
			expected:    "decoding message body failed: unexpected EOF",
		},
		{
			name:        "partial bulk failure",
			metricCount: 2,
			status:      http.StatusOK,
			response: `{"errors":true,"items":[
				{"index":{"status":201}},
				{"index":{"status":429,"error":{"type":"es_rejected_execution_exception","reason":"queue full"}}}
			]}`,
			expected: "elasticsearch failed to index 1 metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := startMockServer(t)
			server.setBulkResponse(tt.status, tt.response)
			defer server.close()

			plugin := &Elasticsearch{
				URLs:      []string{"http://" + server.addr()},
				IndexName: "test-%Y.%m.%d",
				Timeout:   config.Duration(time.Second * 5),
				Log:       testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			metrics := make([]telegraf.Metric, tt.metricCount)
			for i := range metrics {
				metrics[i] = testutil.TestMetric(1)
			}
			require.ErrorContains(t, plugin.Write(metrics), tt.expected)
			require.Len(t, server.messages(), 1, "failed bulk requests must not be retried")
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
		version  int
		template map[string]any
		expected string
	}{
		{
			name:    "standard",
			version: 5,
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
			name:    "custom index",
			version: 5,
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
				URLs:               []string{"http://example.com"},
				ManageTemplate:     true,
				TemplateName:       "test",
				IndexName:          "test%Y.%m.%d",
				IndexTemplate:      tt.template,
				Log:                testutil.Logger{},
				majorReleaseNumber: tt.version,
			}
			require.NoError(t, plugin.Init())

			actual, err := createNewTemplate(plugin.IndexName, plugin.IndexTemplate, plugin.majorReleaseNumber)
			require.NoError(t, err)

			require.JSONEq(t, tt.expected, actual)
		})
	}
}

func TestWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name         string
		image        string
		env          map[string]string
		waitFor      wait.Strategy
		expectedType string
	}{
		{
			name:         "v5",
			image:        "elasticsearch:5.6.16",
			env:          map[string]string{"ES_JAVA_OPTS": "-Xms512m -Xmx512m"},
			waitFor:      wait.ForHTTP("/").WithPort(servicePort).WithStartupTimeout(5 * time.Minute),
			expectedType: "metrics",
		},
		{
			name:  "v6",
			image: "elasticsearch:6.8.23",
			env:   map[string]string{"discovery.type": "single-node"},
			waitFor: wait.ForAll(
				wait.ForLog("] mode [basic] - valid"),
				wait.ForListeningPort(servicePort),
			),
			expectedType: "metrics",
		},
		{
			name:  "v7",
			image: "docker.elastic.co/elasticsearch/elasticsearch:7.17.29",
			env: map[string]string{
				"discovery.type": "single-node",
				"ES_JAVA_OPTS":   "-Xms1g -Xmx1g",
			},
			waitFor:      wait.ForHTTP("/").WithPort(servicePort).WithStartupTimeout(120 * time.Second),
			expectedType: "_doc",
		},
		{
			name:  "OpenSearch v1 compatibility mode",
			image: "opensearchproject/opensearch:1.1.0",
			env: map[string]string{
				"discovery.type":                               "single-node",
				"DISABLE_INSTALL_DEMO_CONFIG":                  "true",
				"DISABLE_SECURITY_PLUGIN":                      "true",
				"DISABLE_PERFORMANCE_ANALYZER_AGENT_CLI":       "true",
				"compatibility.override_main_response_version": "true",
			},
			waitFor: wait.ForAll(
				wait.ForListeningPort(servicePort),
				wait.ForLog("Init AD version hash ring successfully"),
			),
			expectedType: "_doc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup container
			container := &testutil.Container{
				Image:        tt.image,
				ExposedPorts: []string{servicePort},
				Env:          tt.env,
				WaitingFor:   tt.waitFor,
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
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			// Verify that we can successfully write data to Elasticsearch
			require.NoError(t, plugin.Write(testutil.MockMetrics()))

			// Use a separate v5 client for the refresh/search APIs shared by all versions.
			client, err := elasticsearch5.NewClient(elasticsearch5.Config{
				Addresses: plugin.URLs,
			})
			require.NoError(t, err)

			// Refresh the index so the written documents are visible to searches.
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(plugin.Timeout))
			defer cancel()
			refresh, err := client.Indices.Refresh(
				client.Indices.Refresh.WithContext(ctx),
				client.Indices.Refresh.WithIndex("test-*"),
			)
			require.NoError(t, err)
			defer refresh.Body.Close()
			require.Equal(t, http.StatusOK, refresh.StatusCode)

			// Verify the stored document independently of the bulk request.
			result, err := client.Search(
				client.Search.WithContext(ctx),
				client.Search.WithIndex("test-*"),
				client.Search.WithFilterPath("hits.hits._index", "hits.hits._type", "hits.hits._source"),
			)
			require.NoError(t, err)
			defer result.Body.Close()
			require.Equal(t, http.StatusOK, result.StatusCode)

			actual, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			require.JSONEq(t, fmt.Sprintf(`{
				"hits": {
					"hits": [
						{
							"_index": "test-2009.11.10",
							"_type": %q,
							"_source": {
								"@timestamp": "2009-11-10T23:00:00Z",
								"measurement_name": "test1",
								"tag": {
									"tag1": "value1"
								},
								"test1": {
									"value": 1
								}
							}
						}
					]
				}
			}`, tt.expectedType), string(actual))
		})
	}
}

func TestClientDiscovery(t *testing.T) {
	for _, tt := range []struct {
		name    string
		version string
	}{
		{name: "v6", version: "6.8.23"},
		{name: "v7", version: "7.17.29"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var discoveredWrites, seedWrites, discoveries atomic.Int32
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/_bulk" {
					discoveredWrites.Add(1)
				}
				if _, err := io.WriteString(w, "{}"); err != nil {
					t.Error(err)
				}
			}))
			defer target.Close()
			seed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/_nodes/http":
					discoveries.Add(1)
					_, err := fmt.Fprintf(w,
						`{"nodes":{"node":{"roles":["data","ingest"],"http":{"publish_address":%q}}}}`,
						strings.TrimPrefix(target.URL, "http://"),
					)
					if err != nil {
						t.Error(err)
					}
				case "/_bulk":
					seedWrites.Add(1)
					if _, err := io.WriteString(w, "{}"); err != nil {
						t.Error(err)
					}
				default:
					if _, err := fmt.Fprintf(w, `{"version":{"number":%q}}`, tt.version); err != nil {
						t.Error(err)
					}
				}
			}))
			defer seed.Close()
			plugin := &Elasticsearch{
				URLs:                []string{seed.URL},
				IndexName:           "test-%Y.%m.%d",
				EnableSniffer:       true,
				HealthCheckInterval: config.Duration(time.Hour),
				Timeout:             config.Duration(time.Second),
				Log:                 testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			require.Eventually(t, func() bool {
				return plugin.Write(testutil.MockMetrics()) == nil && discoveredWrites.Load() > 0
			}, 5*time.Second, 10*time.Millisecond)
			before := seedWrites.Load()
			require.NoError(t, plugin.Write(testutil.MockMetrics()))
			require.EqualValues(t, 1, discoveries.Load())
			require.EqualValues(t, 2, discoveredWrites.Load())
			require.Equal(t, before, seedWrites.Load())
		})
	}
}

func TestClientDiscoveryEmpty(t *testing.T) {
	for _, tt := range []struct {
		name     string
		version  string
		response string
	}{
		{name: "v6 no nodes", version: "6.8.23", response: `{"nodes":{}}`},
		{name: "v7 no nodes", version: "7.17.29", response: `{"nodes":{}}`},
		{
			name:    "v6 separate data and ingest nodes",
			version: "6.8.23",
			response: `{"nodes":{
				"data":{"roles":["data"],"http":{"publish_address":"127.0.0.1:9200"}},
				"ingest":{"roles":["ingest"],"http":{"publish_address":"127.0.0.1:9201"}}
			}}`,
		},
		{
			name:     "v7 master-only nodes",
			version:  "7.17.29",
			response: `{"nodes":{"master":{"roles":["master"],"http":{"publish_address":"127.0.0.1:9200"}}}}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var response string
				switch r.URL.Path {
				case "/_nodes/http":
					response = tt.response
				case "/_bulk":
					response = "{}"
				default:
					response = fmt.Sprintf(`{"version":{"number":%q}}`, tt.version)
				}
				if _, err := io.WriteString(w, response); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			plugin := &Elasticsearch{
				URLs:      []string{server.URL},
				IndexName: "test-%Y.%m.%d",
				Timeout:   config.Duration(time.Second),
				Log:       testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			var discover func() error
			switch connection := plugin.client.(type) {
			case *clientV6:
				discover = connection.client.DiscoverNodes
			case *clientV7:
				discover = connection.client.DiscoverNodes
			default:
				t.Fatalf("unexpected client type %T", plugin.client)
			}
			// Repeat discovery to verify the retained pool remains usable for discovery and writes.
			for range 2 {
				require.NoError(t, discover())
				require.NoError(t, plugin.Write(testutil.MockMetrics()))
			}
		})
	}
}

func TestClientDiscoveryDisabled(t *testing.T) {
	for _, tt := range []struct {
		name          string
		version       string
		enableSniffer bool
		interval      time.Duration
	}{
		{name: "v5 ignores discovery", version: "5.6.16", enableSniffer: true, interval: time.Hour},
		{name: "v6 sniffer disabled", version: "6.8.23", interval: time.Hour},
		{name: "v6 zero interval", version: "6.8.23", enableSniffer: true},
		{name: "v7 sniffer disabled", version: "7.17.29", interval: time.Hour},
		{name: "v7 zero interval", version: "7.17.29", enableSniffer: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var writes, discoveries atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/_nodes/http":
					discoveries.Add(1)
					if _, err := io.WriteString(w, "{}"); err != nil {
						t.Error(err)
					}
				case "/_bulk":
					writes.Add(1)
					if _, err := io.WriteString(w, "{}"); err != nil {
						t.Error(err)
					}
				default:
					if _, err := fmt.Fprintf(w, `{"version":{"number":%q}}`, tt.version); err != nil {
						t.Error(err)
					}
				}
			}))
			defer server.Close()
			plugin := &Elasticsearch{
				URLs:                []string{server.URL},
				IndexName:           "test-%Y.%m.%d",
				EnableSniffer:       tt.enableSniffer,
				HealthCheckInterval: config.Duration(tt.interval),
				Timeout:             config.Duration(time.Second),
				Log:                 testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			require.NoError(t, plugin.Write(testutil.MockMetrics()))
			require.NoError(t, plugin.Close())
			require.Zero(t, discoveries.Load())
			require.EqualValues(t, 1, writes.Load())
		})
	}
}

func TestConnectRetry(t *testing.T) {
	var rejectTemplate atomic.Bool
	rejectTemplate.Store(true)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/_template/telegraf":
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPut:
			if rejectTemplate.Load() {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if _, err := io.WriteString(w, `{"acknowledged":true}`); err != nil {
				t.Error(err)
			}
		default:
			if _, err := io.WriteString(w, `{"version":{"number":"6.8.23"}}`); err != nil {
				t.Error(err)
			}
		}
	}))
	defer server.Close()
	plugin := &Elasticsearch{
		URLs:                []string{server.URL},
		IndexName:           "test-%Y.%m.%d",
		TemplateName:        "telegraf",
		ManageTemplate:      true,
		Timeout:             config.Duration(time.Second),
		HealthCheckInterval: config.Duration(time.Second),
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	defer plugin.Close()
	require.ErrorContains(t, plugin.Connect(), "managing templates failed:")
	require.Nil(t, plugin.client)
	rejectTemplate.Store(false)
	require.NoError(t, plugin.Connect())
}

func TestClientClose(t *testing.T) {
	for _, tt := range []struct {
		name      string
		newClient func(clientConfig) (client, error)
	}{
		{name: "v6", newClient: newClientV6},
		{name: "v7", newClient: newClientV7},
	} {
		t.Run(tt.name, func(t *testing.T) {
			started := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				close(started)
				<-r.Context().Done()
			}))
			defer server.Close()
			connection, err := tt.newClient(clientConfig{
				urls:              []string{server.URL},
				enableSniffer:     true,
				discoveryInterval: time.Hour,
				httpClient:        &http.Client{Timeout: time.Second},
				log:               testutil.Logger{},
			})
			require.NoError(t, err)
			defer connection.close()
			select {
			case <-started:
			case <-time.After(5 * time.Second):
				t.Fatal("discovery did not start")
			}
			closed := make(chan struct{})
			go func() {
				connection.close()
				close(closed)
			}()
			select {
			case <-closed:
			case <-time.After(5 * time.Second):
				t.Fatal("client did not stop discovery after the HTTP timeout")
			}
		})
	}
}

type mockServer struct {
	version      string
	bulkStatus   int
	bulkResponse string
	server       *httptest.Server
	msgs         []string
	hdrs         []http.Header
	templates    map[string]string
	sync.Mutex
}

func startMockServer(t *testing.T) *mockServer {
	s := &mockServer{
		version:      "6.8.23",
		bulkStatus:   http.StatusOK,
		bulkResponse: "{}",
		msgs:         make([]string, 0),
		hdrs:         make([]http.Header, 0),
		templates:    make(map[string]string),
	}

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if name, found := strings.CutPrefix(r.URL.Path, "/_template/"); found {
			switch r.Method {
			case http.MethodHead:
				s.Lock()
				_, exists := s.templates[name]
				s.Unlock()
				if !exists {
					w.WriteHeader(http.StatusNotFound)
				}
			case http.MethodPut:
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Error(err)
					return
				}
				s.Lock()
				s.templates[name] = string(body)
				s.Unlock()
				if _, err := w.Write([]byte(`{"acknowledged": true}`)); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Error(err)
				}
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				t.Errorf("unexpected method %q for template %q", r.Method, name)
			}
			return
		}

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
			status, response := s.bulkStatus, s.bulkResponse
			s.Unlock()

			w.WriteHeader(status)
			if _, err := w.Write([]byte(response)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
			}
		default:
			s.Lock()
			version := s.version
			s.Unlock()
			if _, err := fmt.Fprintf(w, `{"version": {"number": %q}}`, version); err != nil {
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

func (s *mockServer) setTemplate(name, body string) {
	s.Lock()
	defer s.Unlock()
	s.templates[name] = body
}

func (s *mockServer) template(name string) (string, bool) {
	s.Lock()
	defer s.Unlock()
	body, found := s.templates[name]
	return body, found
}

func (s *mockServer) setVersion(version string) {
	s.Lock()
	defer s.Unlock()
	s.version = version
}

func (s *mockServer) setBulkResponse(status int, response string) {
	s.Lock()
	defer s.Unlock()
	s.bulkStatus = status
	s.bulkResponse = response
}
