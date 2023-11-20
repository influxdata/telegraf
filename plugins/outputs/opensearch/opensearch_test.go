package opensearch

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

const servicePort = "9200"
const imageVersion1 = "1.1.0"
const imageVersion2 = "2.8.0"

func launchTestContainer(t *testing.T, imageVersion string) *testutil.Container {
	container := testutil.Container{
		Image:        "opensearchproject/opensearch:" + imageVersion,
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"discovery.type":                         "single-node",
			"DISABLE_INSTALL_DEMO_CONFIG":            "true",
			"DISABLE_SECURITY_PLUGIN":                "true",
			"DISABLE_PERFORMANCE_ANALYZER_AGENT_CLI": "true",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port(servicePort)),
			wait.ForLog("Init AD version hash ring successfully"),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")

	return &container
}

func TestGetIndexName(t *testing.T) {
	e := &Opensearch{
		Log: testutil.Logger{},
	}

	tests := []struct {
		EventTime time.Time
		Tags      map[string]string
		IndexName string
		Expected  string
	}{
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			"indexname",
			"indexname",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{},
			`indexname-{{.Time.Format "2006-01-02"}}`,
			"indexname-2014-12-01",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{},
			`indexname-{{.Tag "tag2"}}-{{.Time.Format "2006-01-02"}}`,
			"indexname--2014-12-01",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1"},
			`indexname-{{.Tag "tag1"}}-{{.Time.Format "2006-01-02"}}`,
			"indexname-value1-2014-12-01",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			`indexname-{{.Tag "tag1"}}-{{.Tag "tag2"}}-{{.Tag "tag3"}}-{{.Time.Format "2006-01-02"}}`,
			"indexname-value1-value2--2014-12-01",
		},
	}
	for _, test := range tests {
		mockMetric := testutil.MockMetrics()[0]
		mockMetric.SetTime(test.EventTime)
		for key, val := range test.Tags {
			mockMetric.AddTag(key, val)
		}
		e.indexTmpl, _ = template.New("index").Parse(test.IndexName)
		indexName, err := e.GetIndexName(mockMetric)
		require.NoError(t, err)
		if indexName != test.Expected {
			t.Errorf("Expected indexname %s, got %s\n", test.Expected, indexName)
		}
	}
}

func TestGetPipelineName(t *testing.T) {
	e := &Opensearch{
		DefaultPipeline: "myDefaultPipeline",
		Log:             testutil.Logger{},
	}

	tests := []struct {
		Tags            map[string]string
		PipelineTagKeys []string
		UsePipeline     string
		Expected        string
	}{
		{
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			`{{.Tag "es-pipeline"}}`,
			"myDefaultPipeline",
		},
		{
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			``,
			"",
		},
		{
			map[string]string{"tag1": "value1", "es-pipeline": "myOtherPipeline"},
			[]string{},
			`{{.Tag "es-pipeline"}}`,
			"myOtherPipeline",
		},
		{
			map[string]string{"tag1": "pipeline2", "es-pipeline": "myOtherPipeline"},
			[]string{},
			`{{.Tag "tag1"}}`,
			"pipeline2",
		},
	}
	for _, test := range tests {
		e.UsePipeline = test.UsePipeline
		e.pipelineTmpl, _ = template.New("index").Parse(test.UsePipeline)
		mockMetric := testutil.MockMetrics()[0]
		for key, val := range test.Tags {
			mockMetric.AddTag(key, val)
		}

		pipelineName, err := e.getPipelineName(mockMetric)
		require.NoError(t, err)
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

	e := &Opensearch{
		URLs:           urls,
		IndexName:      `test-{{.Tag "tag1"}}-{{.Time.Format "2006-01-02"}}`,
		Timeout:        config.Duration(time.Second * 5),
		EnableGzip:     true,
		ManageTemplate: false,
		Log:            testutil.Logger{},
	}
	e.indexTmpl, _ = template.New("index").Parse(e.IndexName)
	var indexName, err = e.GetIndexName(testutil.MockMetrics()[0])
	require.NoError(t, err)
	e.IndexName = indexName

	err = e.Connect()
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

	e := &Opensearch{
		URLs:           urls,
		IndexName:      `test-{{.Tag "tag1"}}-{{.Time.Format "2006-01-02"}}`,
		Timeout:        config.Duration(time.Second * 5),
		EnableGzip:     false,
		ManageTemplate: false,
		Log:            testutil.Logger{},
	}
	e.indexTmpl, _ = template.New("index").Parse(e.IndexName)
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

	e := &Opensearch{
		URLs:            urls,
		IndexName:       `{{.Tag "tag1"}}-{{.Time.Format "2006-01-02"}}`,
		Timeout:         config.Duration(time.Second * 5),
		EnableGzip:      false,
		ManageTemplate:  false,
		Log:             testutil.Logger{},
		AuthBearerToken: config.NewSecret([]byte("0123456789abcdef")),
	}
	e.indexTmpl, _ = template.New("index").Parse(e.IndexName)
	err := e.Connect()
	require.NoError(t, err)

	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
