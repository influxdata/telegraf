package elasticsearch

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

const servicePort = "9200"

func TestWriteFloatHandlingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		floatHandling string
		expected      string
	}{
		{
			expected: "error sending bulk request to Elasticsearch: json: unsupported value: ",
		},
		{
			floatHandling: "none",
			expected:      "error sending bulk request to Elasticsearch: json: unsupported value: ",
		},
		{
			floatHandling: "drop",
		},
		{
			floatHandling: "replace",
		},
	}

	for _, tt := range tests {
		name := tt.floatHandling
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
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
				ManageTemplate:      true,
				TemplateName:        "telegraf",
				Timeout:             config.Duration(time.Second * 5),
				HealthCheckInterval: config.Duration(time.Second * 10),
				HealthCheckTimeout:  config.Duration(time.Second * 1),
				FloatHandling:       tt.floatHandling,
				FloatReplacement:    0.0,
				Log:                 testutil.Logger{},
			}
			require.NoError(t, plugin.Connect())

			metrics := map[string]telegraf.Metric{
				"NaN":  testutil.TestMetric(math.NaN()),
				"+Inf": testutil.TestMetric(math.Inf(1)),
				"-Inf": testutil.TestMetric(math.Inf(-1)),
			}

			// Verify that we can fail for metric with unhandled NaN/inf/-inf values
			if tt.expected != "" {
				for k, m := range metrics {
					require.ErrorContains(t, plugin.Write([]telegraf.Metric{m}), tt.expected+k)
				}
			} else {
				for _, m := range metrics {
					require.NoError(t, plugin.Write([]telegraf.Metric{m}))
				}
			}
		})
	}
}

func TestTemplateManagementEmptyTemplateIntegration(t *testing.T) {
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
		URLs:              []string{"http://" + container.Address + ":" + container.Ports[servicePort]},
		IndexName:         "test-%Y.%m.%d",
		Timeout:           config.Duration(time.Second * 5),
		EnableGzip:        true,
		ManageTemplate:    true,
		TemplateName:      "",
		OverwriteTemplate: true,
		Log:               testutil.Logger{},
	}

	require.ErrorContains(t, plugin.manageTemplate(t.Context()), "elasticsearch template_name configuration not defined")
}

func TestUseOpTypeCreateIntegration(t *testing.T) {
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
		URLs:              []string{"http://" + container.Address + ":" + container.Ports[servicePort]},
		IndexName:         "test-%Y.%m.%d",
		Timeout:           config.Duration(time.Second * 5),
		EnableGzip:        true,
		ManageTemplate:    true,
		TemplateName:      "telegraf",
		OverwriteTemplate: true,
		UseOpTypeCreate:   true,
		ForceDocumentID:   true,
		Log:               testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())

	ctx, cancel := context.WithTimeout(t.Context(), time.Duration(plugin.Timeout))
	defer cancel()

	require.NoError(t, plugin.manageTemplate(ctx))
	require.NoError(t, plugin.Write([]telegraf.Metric{testutil.TestMetric(1)}))
}

func TestTemplateManagementIntegration(t *testing.T) {
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
		URLs:              []string{"http://" + container.Address + ":" + container.Ports[servicePort]},
		IndexName:         "test-%Y.%m.%d",
		Timeout:           config.Duration(time.Second * 5),
		EnableGzip:        true,
		ManageTemplate:    true,
		TemplateName:      "telegraf",
		OverwriteTemplate: true,
		Log:               testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())

	ctx, cancel := context.WithTimeout(t.Context(), time.Duration(plugin.Timeout))
	defer cancel()

	require.NoError(t, plugin.manageTemplate(ctx))
}

func TestTemplateInvalidIndexPatternIntegration(t *testing.T) {
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
		URLs:              []string{"http://" + container.Address + ":" + container.Ports[servicePort]},
		IndexName:         "{{host}}-%Y.%m.%d",
		Timeout:           config.Duration(time.Second * 5),
		EnableGzip:        true,
		ManageTemplate:    true,
		TemplateName:      "telegraf",
		OverwriteTemplate: true,
		Log:               testutil.Logger{},
	}

	require.ErrorContains(t, plugin.Connect(), "template cannot be created for dynamic index names without an index prefix")
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
			indexName, tagKeys := GetTagKeys(tt.indexName)
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

			indexName := plugin.GetIndexName(tt.indexName, eventTime, tt.tagKeys, tags)
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
			pn, ptags := GetTagKeys(tt.pipeline)

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

func TestRequestHeaderWhenGzipIsEnabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_bulk":
			if contentHeader := r.Header.Get("Content-Encoding"); contentHeader != "gzip" {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("Not equal, expected: %q, actual: %q", "gzip", contentHeader)
				return
			}
			if acceptHeader := r.Header.Get("Accept-Encoding"); acceptHeader != "gzip" {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("Not equal, expected: %q, actual: %q", "gzip", acceptHeader)
				return
			}

			if _, err := w.Write([]byte("{}")); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
			}
			return
		default:
			if _, err := w.Write([]byte(`{"version": {"number": "7.8"}}`)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
			}
			return
		}
	}))
	defer ts.Close()

	// Setup plugin
	plugin := &Elasticsearch{
		URLs:       []string{"http://" + ts.Listener.Addr().String()},
		IndexName:  "{{host}}-%Y.%m.%d",
		Timeout:    config.Duration(time.Second * 5),
		EnableGzip: true,
		Log:        testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())

	require.NoError(t, plugin.Write(testutil.MockMetrics()))
}

func TestRequestHeaderWhenGzipIsDisabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_bulk":
			if contentHeader := r.Header.Get("Content-Encoding"); contentHeader == "gzip" {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("Not equal, expected: %q, actual: %q", "gzip", contentHeader)
				return
			}
			if _, err := w.Write([]byte("{}")); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
			}
			return
		default:
			if _, err := w.Write([]byte(`{"version": {"number": "7.8"}}`)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
			}
			return
		}
	}))
	defer ts.Close()

	// Setup plugin
	plugin := &Elasticsearch{
		URLs:       []string{"http://" + ts.Listener.Addr().String()},
		IndexName:  "{{host}}-%Y.%m.%d",
		Timeout:    config.Duration(time.Second * 5),
		EnableGzip: false,
		Log:        testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())

	require.NoError(t, plugin.Write(testutil.MockMetrics()))
}

func TestAuthorizationHeaderWhenBearerTokenIsPresent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_bulk":
			if authHeader := r.Header.Get("Authorization"); authHeader != "Bearer 0123456789abcdef" {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("Not equal, expected: %q, actual: %q", "Bearer 0123456789abcdef", authHeader)
				return
			}
			if _, err := w.Write([]byte("{}")); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
			}
			return
		default:
			if _, err := w.Write([]byte(`{"version": {"number": "7.8"}}`)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
			}
			return
		}
	}))
	defer ts.Close()

	// Setup plugin
	plugin := &Elasticsearch{
		URLs:            []string{"http://" + ts.Listener.Addr().String()},
		IndexName:       "{{host}}-%Y.%m.%d",
		Timeout:         config.Duration(time.Second * 5),
		Log:             testutil.Logger{},
		AuthBearerToken: config.NewSecret([]byte("0123456789abcdef")),
	}
	require.NoError(t, plugin.Connect())

	require.NoError(t, plugin.Write(testutil.MockMetrics()))
}

func TestCustomHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]any
		expected map[string][]string
	}{
		{
			// If headers are not set http.Header should be empty
			name:     "no headers",
			expected: make(map[string][]string),
		},
		{
			// Empty headers map should return empty http.Header
			name:     "empty headers map",
			headers:  map[string]any{},
			expected: make(map[string][]string),
		},
		{
			// Invalid types should be rejected with error logging
			name: "invalid types",
			headers: map[string]any{
				"X-Numeric": 123,
				"X-Boolean": true,
				"X-Float":   45.67,
				"X-Nil":     nil,
			},
			expected: make(map[string][]string),
		},
		{
			// Single string values are split on commas (deprecated behavior with warnings)
			name: "strings with commas (deprecated)",
			headers: map[string]any{
				"Content-Type":        "application/json",
				"Authorization":       "Bearer token123",
				"VL-Stream-Fields":    "tag.Source,tag.Channel,tag.EventID",
				"VL-Msg-Field":        "win_eventlog.Message",
				"CSV-Data":            "col1,col2,col3,col4",
				"Content-Disposition": `attachment; filename="file,with,commas.csv"`,
				"X-Special-Chars":     "value with spaces, commas, and \"quotes\"",
				"X-Unicode":           "测试值",
				"X-JSON-Like":         `{"key": "value", "array": [1,2,3]}`,
			},
			expected: map[string][]string{
				"Content-Type":        {"application/json"},
				"Authorization":       {"Bearer token123"},
				"Vl-Stream-Fields":    {"tag.Source", "tag.Channel", "tag.EventID"}, // Split on commas
				"Vl-Msg-Field":        {"win_eventlog.Message"},
				"Csv-Data":            {"col1", "col2", "col3", "col4"},                      // Split on commas
				"Content-Disposition": {`attachment; filename="file`, `with`, `commas.csv"`}, // Split on commas
				"X-Special-Chars":     {"value with spaces", "commas", `and "quotes"`},       // Split on commas
				"X-Unicode":           {"测试值"},
				"X-Json-Like":         {`{"key": "value"`, `"array": [1`, `2`, `3]}`}, // Split on commas
			},
		},
		{
			// Interface arrays should create multiple header values with whitespace trimmed, empty arrays ignored
			// X-Empty-Array is not included - empty arrays don't create headers
			name: "arrays with whitespace",
			headers: map[string]any{
				"Accept":        []any{"application/json", "application/xml", "text/plain"},
				"Cache-Control": []any{"no-cache", "must-revalidate"},
				"X-Debug-Tags":  []any{"performance", "security", "monitoring"},
				"X-With-Spaces": []any{" application/json ", "  application/xml  ", "text/plain"},
				"X-Empty-Array": make([]any, 0),
			},
			expected: map[string][]string{
				"Accept":        {"application/json", "application/xml", "text/plain"},
				"Cache-Control": {"no-cache", "must-revalidate"},
				"X-Debug-Tags":  {"performance", "security", "monitoring"},
				"X-With-Spaces": {"application/json", "application/xml", "text/plain"}, // Trimmed
			},
		},
		{
			// Interface arrays should convert strings and log errors for non-string types, empty arrays ignored
			// X-Empty-Interface is not included - empty arrays don't create headers
			name: "interface arrays - TOML parsing and mixed types",
			headers: map[string]any{
				"X-Forwarded-For":   []any{"192.168.1.1", "10.0.0.1", "172.16.0.1"},
				"X-Mixed-Types":     []any{"string-value", 123, true, "another-string"},
				"X-Empty-Interface": make([]any, 0),
			},
			expected: map[string][]string{
				"X-Forwarded-For": {"192.168.1.1", "10.0.0.1", "172.16.0.1"},
				"X-Mixed-Types":   {"string-value", "another-string"}, // Only strings processed
			},
		},
		{
			// Mixed header types work correctly with comma-splitting for strings (deprecated)
			// X-Empty-Array is not included - empty arrays don't create headers
			name: "comprehensive mixed scenario",
			headers: map[string]any{
				"VL-Stream-Fields": "tag.Source,tag.Channel,tag.EventID",
				"VL-Time-Field":    "@timestamp",
				"Authorization":    "Bearer token123",
				"Accept":           []any{"application/json", "text/plain"},
				"X-Debug-Tags":     []any{"performance", "security"},
				"X-IPs":            []any{"1.1.1.1", "2.2.2.2"},
				"X-Empty-String":   "",
				"X-Empty-Array":    make([]any, 0),
			},
			expected: map[string][]string{
				"Vl-Stream-Fields": {"tag.Source", "tag.Channel", "tag.EventID"}, // Split on commas (deprecated)
				"Vl-Time-Field":    {"@timestamp"},
				"Authorization":    {"Bearer token123"},
				"Accept":           {"application/json", "text/plain"},
				"X-Debug-Tags":     {"performance", "security"},
				"X-Ips":            {"1.1.1.1", "2.2.2.2"},
				"X-Empty-String":   {""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin
			plugin := &Elasticsearch{
				Headers: tt.headers,
				Log:     testutil.Logger{},
			}

			result := plugin.processHeaders()
			require.EqualValues(t, tt.expected, result)
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
		Timeout:             config.Duration(time.Second * 5),
		EnableGzip:          true,
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		HealthCheckInterval: config.Duration(time.Second * 10),
		HealthCheckTimeout:  config.Duration(time.Second * 1),
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())

	// Verify that we can successfully write data to Elasticsearch
	require.NoError(t, plugin.Write(testutil.MockMetrics()))
}
