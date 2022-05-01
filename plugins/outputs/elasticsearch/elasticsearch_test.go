package elasticsearch

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	e := &Elasticsearch{
		URLs:                urls,
		IndexName:           "test-%Y.%m.%d",
		Timeout:             config.Duration(time.Second * 5),
		EnableGzip:          true,
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		OverwriteTemplate:   false,
		HealthCheckInterval: config.Duration(time.Second * 10),
		HealthCheckTimeout:  config.Duration(time.Second * 1),
		Log:                 testutil.Logger{},
	}

	// Verify that we can connect to Elasticsearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to Elasticsearch
	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestConnectAndWriteMetricWithNaNValueEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	e := &Elasticsearch{
		URLs:                urls,
		IndexName:           "test-%Y.%m.%d",
		Timeout:             config.Duration(time.Second * 5),
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		OverwriteTemplate:   false,
		HealthCheckInterval: config.Duration(time.Second * 10),
		HealthCheckTimeout:  config.Duration(time.Second * 1),
		Log:                 testutil.Logger{},
	}

	metrics := []telegraf.Metric{
		testutil.TestMetric(math.NaN()),
		testutil.TestMetric(math.Inf(1)),
		testutil.TestMetric(math.Inf(-1)),
	}

	// Verify that we can connect to Elasticsearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can fail for metric with unhandled NaN/inf/-inf values
	for _, m := range metrics {
		err = e.Write([]telegraf.Metric{m})
		require.Error(t, err, "error sending bulk request to Elasticsearch: json: unsupported value: NaN")
	}
}

func TestConnectAndWriteMetricWithNaNValueNone(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	e := &Elasticsearch{
		URLs:                urls,
		IndexName:           "test-%Y.%m.%d",
		Timeout:             config.Duration(time.Second * 5),
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		OverwriteTemplate:   false,
		HealthCheckInterval: config.Duration(time.Second * 10),
		HealthCheckTimeout:  config.Duration(time.Second * 1),
		FloatHandling:       "none",
		Log:                 testutil.Logger{},
	}

	metrics := []telegraf.Metric{
		testutil.TestMetric(math.NaN()),
		testutil.TestMetric(math.Inf(1)),
		testutil.TestMetric(math.Inf(-1)),
	}

	// Verify that we can connect to Elasticsearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can fail for metric with unhandled NaN/inf/-inf values
	for _, m := range metrics {
		err = e.Write([]telegraf.Metric{m})
		require.Error(t, err, "error sending bulk request to Elasticsearch: json: unsupported value: NaN")
	}
}

func TestConnectAndWriteMetricWithNaNValueDrop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	e := &Elasticsearch{
		URLs:                urls,
		IndexName:           "test-%Y.%m.%d",
		Timeout:             config.Duration(time.Second * 5),
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		OverwriteTemplate:   false,
		HealthCheckInterval: config.Duration(time.Second * 10),
		HealthCheckTimeout:  config.Duration(time.Second * 1),
		FloatHandling:       "drop",
		Log:                 testutil.Logger{},
	}

	metrics := []telegraf.Metric{
		testutil.TestMetric(math.NaN()),
		testutil.TestMetric(math.Inf(1)),
		testutil.TestMetric(math.Inf(-1)),
	}

	// Verify that we can connect to Elasticsearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can fail for metric with unhandled NaN/inf/-inf values
	for _, m := range metrics {
		err = e.Write([]telegraf.Metric{m})
		require.NoError(t, err)
	}
}

func TestConnectAndWriteMetricWithNaNValueReplacement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	e := &Elasticsearch{
		URLs:                urls,
		IndexName:           "test-%Y.%m.%d",
		Timeout:             config.Duration(time.Second * 5),
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		OverwriteTemplate:   false,
		HealthCheckInterval: config.Duration(time.Second * 10),
		HealthCheckTimeout:  config.Duration(time.Second * 1),
		FloatHandling:       "3.1415",
		Log:                 testutil.Logger{},
	}

	metrics := []telegraf.Metric{
		testutil.TestMetric(math.NaN()),
		testutil.TestMetric(math.Inf(1)),
		testutil.TestMetric(math.Inf(-1)),
	}

	// Verify that we can connect to Elasticsearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can fail for metric with unhandled NaN/inf/-inf values
	for _, m := range metrics {
		err = e.Write([]telegraf.Metric{m})
		require.NoError(t, err)
	}
}

func TestTemplateManagementEmptyTemplateIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	ctx := context.Background()

	e := &Elasticsearch{
		URLs:              urls,
		IndexName:         "test-%Y.%m.%d",
		Timeout:           config.Duration(time.Second * 5),
		EnableGzip:        true,
		ManageTemplate:    true,
		TemplateName:      "",
		OverwriteTemplate: true,
		Log:               testutil.Logger{},
	}

	err := e.manageTemplate(ctx)
	require.Error(t, err)
}

func TestTemplateManagementIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	e := &Elasticsearch{
		URLs:              urls,
		IndexName:         "test-%Y.%m.%d",
		Timeout:           config.Duration(time.Second * 5),
		EnableGzip:        true,
		ManageTemplate:    true,
		TemplateName:      "telegraf",
		OverwriteTemplate: true,
		Log:               testutil.Logger{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.Timeout))
	defer cancel()

	err := e.Connect()
	require.NoError(t, err)

	err = e.manageTemplate(ctx)
	require.NoError(t, err)
}

func TestTemplateInvalidIndexPatternIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	e := &Elasticsearch{
		URLs:              urls,
		IndexName:         "{{host}}-%Y.%m.%d",
		Timeout:           config.Duration(time.Second * 5),
		EnableGzip:        true,
		ManageTemplate:    true,
		TemplateName:      "telegraf",
		OverwriteTemplate: true,
		Log:               testutil.Logger{},
	}

	err := e.Connect()
	require.Error(t, err)
}

func TestGetTagKeys(t *testing.T) {
	e := &Elasticsearch{
		DefaultTagValue: "none",
		Log:             testutil.Logger{},
	}

	tests := []struct {
		IndexName         string
		ExpectedIndexName string
		ExpectedTagKeys   []string
	}{
		{
			"indexname",
			"indexname",
			[]string{},
		}, {
			"indexname-%Y",
			"indexname-%Y",
			[]string{},
		}, {
			"indexname-%Y-%m",
			"indexname-%Y-%m",
			[]string{},
		}, {
			"indexname-%Y-%m-%d",
			"indexname-%Y-%m-%d",
			[]string{},
		}, {
			"indexname-%Y-%m-%d-%H",
			"indexname-%Y-%m-%d-%H",
			[]string{},
		}, {
			"indexname-%y-%m",
			"indexname-%y-%m",
			[]string{},
		}, {
			"indexname-{{tag1}}-%y-%m",
			"indexname-%s-%y-%m",
			[]string{"tag1"},
		}, {
			"indexname-{{tag1}}-{{tag2}}-%y-%m",
			"indexname-%s-%s-%y-%m",
			[]string{"tag1", "tag2"},
		}, {
			"indexname-{{tag1}}-{{tag2}}-{{tag3}}-%y-%m",
			"indexname-%s-%s-%s-%y-%m",
			[]string{"tag1", "tag2", "tag3"},
		},
	}
	for _, test := range tests {
		indexName, tagKeys := e.GetTagKeys(test.IndexName)
		if indexName != test.ExpectedIndexName {
			t.Errorf("Expected indexname %s, got %s\n", test.ExpectedIndexName, indexName)
		}
		if !reflect.DeepEqual(tagKeys, test.ExpectedTagKeys) {
			t.Errorf("Expected tagKeys %s, got %s\n", test.ExpectedTagKeys, tagKeys)
		}
	}
}

func TestGetIndexName(t *testing.T) {
	e := &Elasticsearch{
		DefaultTagValue: "none",
		Log:             testutil.Logger{},
	}

	tests := []struct {
		EventTime time.Time
		Tags      map[string]string
		TagKeys   []string
		IndexName string
		Expected  string
	}{
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"indexname",
			"indexname",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"indexname-%Y",
			"indexname-2014",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"indexname-%Y-%m",
			"indexname-2014-12",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"indexname-%Y-%m-%d",
			"indexname-2014-12-01",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"indexname-%Y-%m-%d-%H",
			"indexname-2014-12-01-23",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"indexname-%y-%m",
			"indexname-14-12",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"indexname-%Y-%V",
			"indexname-2014-49",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{"tag1"},
			"indexname-%s-%y-%m",
			"indexname-value1-14-12",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{"tag1", "tag2"},
			"indexname-%s-%s-%y-%m",
			"indexname-value1-value2-14-12",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{"tag1", "tag2", "tag3"},
			"indexname-%s-%s-%s-%y-%m",
			"indexname-value1-value2-none-14-12",
		},
	}
	for _, test := range tests {
		indexName := e.GetIndexName(test.IndexName, test.EventTime, test.TagKeys, test.Tags)
		if indexName != test.Expected {
			t.Errorf("Expected indexname %s, got %s\n", test.Expected, indexName)
		}
	}
}

func TestGetPipelineName(t *testing.T) {
	e := &Elasticsearch{
		UsePipeline:     "{{es-pipeline}}",
		DefaultPipeline: "myDefaultPipeline",
		Log:             testutil.Logger{},
	}
	e.pipelineName, e.pipelineTagKeys = e.GetTagKeys(e.UsePipeline)

	tests := []struct {
		EventTime       time.Time
		Tags            map[string]string
		PipelineTagKeys []string
		Expected        string
	}{
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"myDefaultPipeline",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"myDefaultPipeline",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "es-pipeline": "myOtherPipeline"},
			[]string{},
			"myOtherPipeline",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "es-pipeline": "pipeline2"},
			[]string{},
			"pipeline2",
		},
	}
	for _, test := range tests {
		pipelineName := e.getPipelineName(e.pipelineName, e.pipelineTagKeys, test.Tags)
		require.Equal(t, test.Expected, pipelineName)
	}

	// Setup testing for testing no pipeline set. All the tests in this case should return "".
	e = &Elasticsearch{
		Log: testutil.Logger{},
	}
	e.pipelineName, e.pipelineTagKeys = e.GetTagKeys(e.UsePipeline)

	for _, test := range tests {
		pipelineName := e.getPipelineName(e.pipelineName, e.pipelineTagKeys, test.Tags)
		require.Equal(t, "", pipelineName)
	}
}

func TestPipelineConfigs(t *testing.T) {
	tests := []struct {
		EventTime       time.Time
		Tags            map[string]string
		PipelineTagKeys []string
		Expected        string
		Elastic         *Elasticsearch
	}{
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"",
			&Elasticsearch{
				Log: testutil.Logger{},
			},
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"",
			&Elasticsearch{
				DefaultPipeline: "myDefaultPipeline",
				Log:             testutil.Logger{},
			},
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "es-pipeline": "myOtherPipeline"},
			[]string{},
			"myDefaultPipeline",
			&Elasticsearch{
				UsePipeline: "myDefaultPipeline",
				Log:         testutil.Logger{},
			},
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "es-pipeline": "pipeline2"},
			[]string{},
			"",
			&Elasticsearch{
				DefaultPipeline: "myDefaultPipeline",
				Log:             testutil.Logger{},
			},
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "es-pipeline": "pipeline2"},
			[]string{},
			"pipeline2",
			&Elasticsearch{
				UsePipeline: "{{es-pipeline}}",
				Log:         testutil.Logger{},
			},
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "es-pipeline": "pipeline2"},
			[]string{},
			"value1-pipeline2",
			&Elasticsearch{
				UsePipeline: "{{tag1}}-{{es-pipeline}}",
				Log:         testutil.Logger{},
			},
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1"},
			[]string{},
			"",
			&Elasticsearch{
				UsePipeline: "{{es-pipeline}}",
				Log:         testutil.Logger{},
			},
		},
	}

	for _, test := range tests {
		e := test.Elastic
		e.pipelineName, e.pipelineTagKeys = e.GetTagKeys(e.UsePipeline)
		pipelineName := e.getPipelineName(e.pipelineName, e.pipelineTagKeys, test.Tags)
		require.Equal(t, test.Expected, pipelineName)
	}
}

func TestRequestHeaderWhenGzipIsEnabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_bulk":
			require.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
			require.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
			_, err := w.Write([]byte("{}"))
			require.NoError(t, err)
			return
		default:
			_, err := w.Write([]byte(`{"version": {"number": "7.8"}}`))
			require.NoError(t, err)
			return
		}
	}))
	defer ts.Close()

	urls := []string{"http://" + ts.Listener.Addr().String()}

	e := &Elasticsearch{
		URLs:           urls,
		IndexName:      "{{host}}-%Y.%m.%d",
		Timeout:        config.Duration(time.Second * 5),
		EnableGzip:     true,
		ManageTemplate: false,
		Log:            testutil.Logger{},
	}

	err := e.Connect()
	require.NoError(t, err)

	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestRequestHeaderWhenGzipIsDisabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_bulk":
			require.NotEqual(t, "gzip", r.Header.Get("Content-Encoding"))
			_, err := w.Write([]byte("{}"))
			require.NoError(t, err)
			return
		default:
			_, err := w.Write([]byte(`{"version": {"number": "7.8"}}`))
			require.NoError(t, err)
			return
		}
	}))
	defer ts.Close()

	urls := []string{"http://" + ts.Listener.Addr().String()}

	e := &Elasticsearch{
		URLs:           urls,
		IndexName:      "{{host}}-%Y.%m.%d",
		Timeout:        config.Duration(time.Second * 5),
		EnableGzip:     false,
		ManageTemplate: false,
		Log:            testutil.Logger{},
	}

	err := e.Connect()
	require.NoError(t, err)

	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestAuthorizationHeaderWhenBearerTokenIsPresent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_bulk":
			require.Equal(t, "Bearer 0123456789abcdef", r.Header.Get("Authorization"))
			_, err := w.Write([]byte("{}"))
			require.NoError(t, err)
			return
		default:
			_, err := w.Write([]byte(`{"version": {"number": "7.8"}}`))
			require.NoError(t, err)
			return
		}
	}))
	defer ts.Close()

	urls := []string{"http://" + ts.Listener.Addr().String()}

	e := &Elasticsearch{
		URLs:            urls,
		IndexName:       "{{host}}-%Y.%m.%d",
		Timeout:         config.Duration(time.Second * 5),
		EnableGzip:      false,
		ManageTemplate:  false,
		Log:             testutil.Logger{},
		AuthBearerToken: "0123456789abcdef",
	}

	err := e.Connect()
	require.NoError(t, err)

	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
